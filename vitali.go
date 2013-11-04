package vitali

import (
    "os"
    "io"
    "log"
    "fmt"
    "net/http"
    "net/http/httputil"
    "html/template"
    "time"
    "regexp"
    "strings"
    "reflect"
)

const (
    PUBLIC = iota
    AUTHENTICATED
)

type RouteRule struct {
    Pattern string
    Resource interface{}
}

type PatternMapping struct {
    Re *regexp.Regexp
    Names []string
}

type webApp struct {
    RouteRules []RouteRule
    PatternMappings []PatternMapping
    UserProvider UserProvider
    Settings map[string]string
    DumpRequest bool
    ErrTemplate *template.Template
}

func checkPermission(perm Perm, method Method, user string) bool {
    required, exist := perm[method]
    if !exist {
        required = perm["*"]
    }
    return !(required==AUTHENTICATED && user == "")
}

func checkMediaType(accept Accept, method Method, mediaType MediaType) bool {
    acceptedTypes, exist := accept[method]
    if !exist {
        return true
    }
    for _, acceptedType := range acceptedTypes {
        if mediaType == acceptedType {
            return true
        }
    }
    return false
}

func (c webApp) matchRules(w *wrappedWriter, r *http.Request) (result interface{}) {
    for i, routeRule := range c.RouteRules {
        params := c.PatternMappings[i].Re.FindStringSubmatch(r.URL.Path)
        if params != nil {
            pathParams := make(map[string]string)
            if len(params) > 1 {
                for j, param := range params[1:] {
                    pathParams[c.PatternMappings[i].Names[j]] = param
                }
            }

            ctx := Ctx {
                pathParams: pathParams,
                Username: c.UserProvider.User(r),
                Request: r,
                ResponseWriter: w,
            }

            vResource := reflect.ValueOf(routeRule.Resource)
            vNewResource := reflect.New(reflect.TypeOf(routeRule.Resource)).Elem()
            for i := 0; i < vResource.NumField(); i++ {
                srcField := vResource.Field(i)
                newField := vNewResource.Field(i)

                switch reflect.TypeOf(srcField.Interface()).Name() {
                case "Ctx":
                    newField.Set(reflect.ValueOf(ctx))
                case "Perm":
                    if !checkPermission(srcField.Interface().(Perm), Method(r.Method),
                            ctx.Username) {
                        if c.Settings["401_PAGE"] != "" {
                            w.Header().Set("Content-Type", "text/html")
                            w.Header()["WWW-Authenticate"] = []string{c.UserProvider.AuthHeader(r)}
                            w.WriteHeader(http.StatusUnauthorized)
                            f, err := os.Open(c.Settings["401_PAGE"])
                            if err != nil {
                                panic(err)
                            }
                            io.Copy(w, f)
                        } else {
                            http.Error(w, "unauthorized", http.StatusUnauthorized)
                        }
                        return w
                    }
                case "Accept":
                    if !checkMediaType(srcField.Interface().(Accept), Method(r.Method),
                            MediaType(r.Header.Get("Content-Type"))) {
                        return unsupportedMediaType{}
                    }
                default:
                    newField.Set(srcField)
                }
            }
            resource := vNewResource.Interface()

            result := getResult(r.Method, resource)
            return result
        }
    }
    return notFound{}
}

func getAllowed(resource interface{}) (allowed []string) {
    _, ok := resource.(Getter)
    if ok {
        allowed = append(allowed, "GET", "HEAD")
    }
    _, ok = resource.(Poster)
    if ok {
        allowed = append(allowed, "POST")
    }
    _, ok = resource.(Putter)
    if ok {
        allowed = append(allowed, "PUT")
    }
    _, ok = resource.(Deleter)
    if ok {
        allowed = append(allowed, "DELETE")
    }
    return
}

func getResult(method string, resource interface{}) (result interface{}) {
    defer func() {
        if r := recover(); r != nil {
            rstr := fmt.Sprintf("%s", r)
            result = internalError {
                where: lineInfo(3),
                why: rstr + fullTrace(5, "\n\t"),
                code: errorCode(rstr),
            }
        }
    }()

    switch method {
    case "HEAD", "GET":
        h, ok := resource.(Getter)
        if ok {
            result = h.Get()
        }
    case "POST":
        h, ok := resource.(Poster)
        if ok {
            result = h.Post()
        }
    case "PUT":
        h, ok := resource.(Putter)
        if ok {
            result = h.Put()
        }
    case "DELETE":
        h, ok := resource.(Deleter)
        if ok {
            result = h.Delete()
        }
    default:
        return notImplemented{}
    }

    if result == nil {
        return methodNotAllowed{getAllowed(resource)}
    }
    return
}

func (c webApp) logRequest(w *wrappedWriter, r *http.Request, elapsedMs float64,
        result interface{}) {
    if w.status == 0 {
        log.Printf("%s %s %s Client Disconnected (%.2f ms)", r.RemoteAddr, r.Method,
            r.URL.Path, elapsedMs)
    } else {
        errMsg := ""
        if w.err.why != "" {
            errMsg = fmt.Sprintf("%s #%d %s ", w.err.where, w.err.code, w.err.why)
        }
        switch result.(type) {
        case unsupportedMediaType:
            errMsg = fmt.Sprintf(": %s ", r.Header.Get("Content-Type"))
        }
        log.Printf("%s %s %s %s %s(%.2f ms, %d bytes)", r.RemoteAddr, r.Method, r.URL.Path,
            http.StatusText(w.status), errMsg, elapsedMs, w.written)

        if c.DumpRequest {
            dump, _ := httputil.DumpRequest(r, false)
            log.Printf("%s", dump)
        }
    }
}

func (c webApp) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    ww := &wrappedWriter{
        status: 0,
        writer: w,
        inTime: time.Now(),
    }
    r.ParseForm()
    result := c.matchRules(ww, r)
    c.writeResponse(ww, r, result)

    elapsedMs := float64(time.Now().UnixNano() - ww.inTime.UnixNano()) / 1000000
    c.logRequest(ww, r, elapsedMs, result)
}

func CreateWebApp(rules []RouteRule) webApp {
    patternMappings := make([]PatternMapping, len(rules))
    for i, v := range rules {
        re := regexp.MustCompile("/{[^}]*}")
        params := re.FindAllString(v.Pattern, -1)
        names := make([]string, len(params))

        transformedPattern := v.Pattern
        for j, param := range params {
            names[j] = param[2:len(param)-1]
            transformedPattern = strings.Replace(transformedPattern, param, "[/]{0,1}([^/]*)", -1)
        }
        patternMappings[i] = PatternMapping{regexp.MustCompile("^"+transformedPattern+"$"), names}
    }

    return webApp{
        RouteRules: rules,
        PatternMappings: patternMappings,
        UserProvider: EmptyUserProvider{},
        Settings: make(map[string]string),
    }
}

type Method string
type Perm map[Method]int

type MediaType string
type MediaTypes []MediaType
type Accept map[Method]MediaTypes
