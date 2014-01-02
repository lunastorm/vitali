package vitali

import (
    "os"
    "io"
    "log"
    "fmt"
    "strconv"
    "net/http"
    "net/http/httputil"
    "html/template"
    "time"
    "regexp"
    "strings"
    "reflect"
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

func checkPermission(perm reflect.StructTag, method Method, roles Roles) bool {
    requiredRole := perm.Get(string(method))
    if requiredRole == "" {
        requiredRole = perm.Get("*")
    }
    if requiredRole == "" {
        return true
    }
    for _, r := range(strings.Split(requiredRole, "|")) {
        _, exists := roles[r]
        if exists {
            return true
        }
    }
    return false
}

func checkMediaType(consumes reflect.StructTag, method Method, mediaType MediaType) bool {
    acceptedTypes := consumes.Get(string(method))
    if acceptedTypes == "" {
        return true
    }
    for _, acceptedType := range strings.Split(acceptedTypes, ",") {
        if mediaType == MediaType(acceptedType) {
            return true
        }
    }
    return false
}

type typeWithPriority struct {
    t string
    q float64
}

func chooseType(provided MediaTypes, acceptHeader string) MediaType {
    if acceptHeader == "" {
        acceptHeader = "*/*"
    }

    typeAndParams := strings.Split(acceptHeader, ",")
    typeWithPriorities := make([]typeWithPriority, len(typeAndParams))
    for i, tpstr := range(typeAndParams) {
        tppair := strings.Split(tpstr, ";")
        var q float64
        if len(tppair) == 1 {
            q = 1.0
        } else {
            q, _ = strconv.ParseFloat(strings.TrimSpace(tppair[1])[2:], 32)
        }
        j := 0
        for ; j<i ; j++ {
            if q > typeWithPriorities[j].q {
                break
            }
        }
        typeWithPriorities = append(typeWithPriorities[:j],
            append([]typeWithPriority{typeWithPriority{strings.TrimSpace(tppair[0]), q}},
                typeWithPriorities[j:]...)...)[:len(typeAndParams)]
    }

    for _, t := range(typeWithPriorities) {
        for _, p := range(provided) {
            matched, _ := regexp.MatchString(fmt.Sprintf("^%s$",
                strings.Replace(t.t, "*", "[^/]+", -1)), string(p))
            if matched {
                return p
            }
        }
    }
    return ""
}

func (c webApp) matchRules(w *wrappedWriter, r *http.Request) (result interface{}, chosenType MediaType) {
    for i, routeRule := range c.RouteRules {
        params := c.PatternMappings[i].Re.FindStringSubmatch(r.URL.Path)
        if params != nil {
            pathParams := make(map[string]string)
            if len(params) > 1 {
                for j, param := range params[1:] {
                    pathParams[c.PatternMappings[i].Names[j]] = param
                }
            }

            user, role := c.UserProvider.GetUserAndRole(r)
            ctx := Ctx {
                pathParams: pathParams,
                Username: user,
                Roles: make(Roles),
                Request: r,
                ResponseWriter: w,
            }
            ctx.Roles[role] = struct{}{}

            vResource := reflect.ValueOf(routeRule.Resource)
            tProvides, found := reflect.TypeOf(routeRule.Resource).FieldByName("Provides")
            if found {
                providedStr := tProvides.Tag.Get(r.Method)
                if providedStr != "" {
                    providedTmp := strings.Split(providedStr, ",")
                    provided := make(MediaTypes, len(providedTmp))
                    for i, v := range providedTmp {
                        provided[i] = MediaType(v)
                    }

                    ctx.ChosenType = MediaType(chooseType(provided, r.Header.Get("Accept")))
                    if ctx.ChosenType == "" {
                        return notAcceptable{provided}, ""
                    }
                    w.Header().Set("Content-Type", string(ctx.ChosenType))
                }
            }

            vNewResource := reflect.New(reflect.TypeOf(routeRule.Resource)).Elem()
            var PermTag reflect.StructTag
            for i := 0; i < vResource.NumField(); i++ {
                srcField := vResource.Field(i)
                newField := vNewResource.Field(i)

                switch reflect.TypeOf(srcField.Interface()).Name() {
                case "Ctx":
                    newField.Set(reflect.ValueOf(ctx))
                case "Perm":
                    PermTag = vResource.Type().Field(i).Tag
                case "Consumes":
                    if !checkMediaType(vResource.Type().Field(i).Tag, Method(r.Method),
                            MediaType(r.Header.Get("Content-Type"))) {
                        return unsupportedMediaType{}, ""
                    }
                default:
                    newField.Set(srcField)
                }
            }
            resource := vNewResource.Interface()
            h, ok := resource.(PreHooker)
            if ok {
                res := h.Pre()
                if res != nil {
                    return res, ctx.ChosenType
                }
            }
            if PermTag != "" {
               if !checkPermission(PermTag, Method(r.Method),
                       ctx.Roles) {
                   w.Header()["WWW-Authenticate"] = []string{c.UserProvider.AuthHeader(r)}
                   if c.Settings["401_PAGE"] != "" {
                       w.Header().Set("Content-Type", "text/html")
                       w.WriteHeader(http.StatusUnauthorized)
                       f, err := os.Open(c.Settings["401_PAGE"])
                       if err != nil {
                           panic(err)
                       }
                       io.Copy(w, f)
                   } else {
                       http.Error(w, "unauthorized", http.StatusUnauthorized)
                   }
                   return w, ""
               }
            }

            result := getResult(r.Method, resource)
            return result, ctx.ChosenType
        }
    }
    return notFound{}, ""
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
    result, chosenType := c.matchRules(ww, r)
    c.writeResponse(ww, r, result, chosenType)

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
type Perm struct{}
type Provides struct{}
type Consumes struct{}

type MediaType string
type MediaTypes []MediaType
type Roles map[string]struct{}

func (c Roles) Add(role string) {
    c[role] = struct{}{}
}
