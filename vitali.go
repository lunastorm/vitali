package vitali

import (
    "log"
    "fmt"
    "strconv"
    "net/http"
    "encoding/json"
    "net/http/httputil"
    "io/ioutil"
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
    LangProvider LangProvider
    Settings map[string]string
    DumpRequest bool
    ErrTemplate *template.Template
    I18n map[string]map[string]template.HTML
    views map[string]*template.Template
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

func (c webApp) matchRules(w *wrappedWriter, r *http.Request) (result interface{}, ctx Ctx, viewName string) {
    for i, routeRule := range c.RouteRules {
        params := c.PatternMappings[i].Re.FindStringSubmatch(r.URL.Path)
        if params != nil {
            pathParams := make(map[string]string)
            if len(params) > 1 {
                for j, param := range params[1:] {
                    pathParams[c.PatternMappings[i].Names[j]] = param
                }
            }

            user, roles := c.UserProvider.GetUserAndRoles(r)
            ctx.pathParams = pathParams
            ctx.Username = user
            ctx.Roles = make(Roles)
            ctx.Request = r
            ctx.ResponseWriter = w
            for _, role := range roles {
                ctx.Roles[role] = struct{}{}
            }
            ctx.ChosenLang = c.LangProvider.Select(&ctx)

            contentType := r.Header.Get("Content-Type")
            if contentType != "" {
                ctx.ContentType = MediaType(strings.Split(contentType, ";")[0])
            }

            vResource := reflect.ValueOf(routeRule.Resource)
            tProvides, found := reflect.TypeOf(routeRule.Resource).FieldByName("Provides")
            if found {
                providedStr := tProvides.Tag.Get(r.Method)
                if providedStr != "" {
                    w.Header().Set("Vary", "Accept")
                    providedTmp := strings.Split(providedStr, ",")
                    provided := make(MediaTypes, len(providedTmp))
                    for i, v := range providedTmp {
                        provided[i] = MediaType(v)
                    }

                    ctx.ChosenType = MediaType(chooseType(provided, r.Header.Get("Accept")))
                    if ctx.ChosenType == "" {
                        result = notAcceptable{provided}
                        return
                    }
                    w.Header().Set("Content-Type", string(ctx.ChosenType))
                }
            }

            vNewResourcePtr := reflect.New(reflect.TypeOf(routeRule.Resource))
            vNewResource := vNewResourcePtr.Elem()
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
                            MediaType(ctx.ContentType)) {
                        result = unsupportedMediaType{}
                        return
                    }
                case "Views":
                    viewName = vResource.Type().Field(i).Tag.Get(r.Method)
                default:
                    newField.Set(srcField)
                }
            }

            vPreFunc := vNewResourcePtr.MethodByName("Pre")
            if vPreFunc.IsValid() {
                result = vPreFunc.Call([]reflect.Value{})[0].Interface()
                if result != nil {
                    return
                }
            }
            if PermTag != "" {
               if !checkPermission(PermTag, Method(r.Method),
                       ctx.Roles) {
                   if user == "" {
                       result = unauthorized{c.UserProvider.AuthHeader(r)}
                   } else {
                       result = forbidden{}
                   }
                   return
               }
            }

            result = getResult(r.Method, &vNewResourcePtr)
            return
        }
    }
    result = notFound{}
    return
}

func getAllowed(vResourcePtr *reflect.Value) (allowed []string) {
    for i:=0; i<vResourcePtr.NumMethod(); i++ {
        method := vResourcePtr.Type().Method(i)
        if method.PkgPath == "" && method.Type.NumIn()==1 && method.Type.NumOut()==1 &&
                method.Type.Out(0).Name() == "" && method.Name != "Pre"{
            if method.Name == "Get" {
                allowed = append(allowed, "HEAD")
                allowed = append(allowed, "GET")
            } else {
                allowed = append(allowed, strings.ToUpper(method.Name))
            }
        }
    }
    return
}

func getResult(method string, vResourcePtr *reflect.Value) (result interface{}) {
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

    methodName := strings.Title(strings.ToLower(method))
    if methodName == "Head"{
        methodName = "Get"
    }
    vMethod := vResourcePtr.MethodByName(methodName)
    if vMethod.IsValid() {
        result = vMethod.Call([]reflect.Value{})[0].Interface()
    }

    if result == nil {
        return methodNotAllowed{getAllowed(vResourcePtr)}
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
    result, ctx, templateName := c.matchRules(ww, r)
    c.writeResponse(ww, r, &result, &ctx, templateName)

    elapsedMs := float64(time.Now().UnixNano() - ww.inTime.UnixNano()) / 1000000
    c.logRequest(ww, r, elapsedMs, result)
}

func CreateWebApp(rules []RouteRule) webApp {
    patternMappings := make([]PatternMapping, len(rules))
    views := make(map[string]*template.Template)
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

        funcMap := template.FuncMap{
            "seq": Seq,
        }
        tViews, ok := reflect.TypeOf(v.Resource).FieldByName("Views")
        if ok {
            for _, kv := range(strings.Split(string(tViews.Tag), " ")) {
                vStr := strings.Split(kv, ":")[1]
                templatesName := vStr[1:len(vStr)-1]

                temp := template.New(templatesName).Funcs(funcMap)
                for _, t := range(strings.Split(templatesName, ",")) {
                    content := panicOnErr(ioutil.ReadFile(fmt.Sprintf("./views/%s", t))).([]uint8)
                    template.Must(temp.Parse(string(content)))
                }
                views[templatesName] = temp
            }
        }
    }
    i18n := make(map[string]map[string]template.HTML)
    content, err := ioutil.ReadFile("views/i18n.json")
    if err == nil {
        err = json.Unmarshal(content, &i18n)
    }

    return webApp{
        RouteRules: rules,
        PatternMappings: patternMappings,
        UserProvider: EmptyUserProvider{},
        LangProvider: &EmptyLangProvider{},
        Settings: make(map[string]string),
        I18n: i18n,
        views: views,
    }
}

type Method string
type Perm struct{}
type Provides struct{}
type Consumes struct{}
type Views struct{}

type MediaType string
type MediaTypes []MediaType
type Roles map[string]struct{}

func (c Roles) Add(role string) {
    c[role] = struct{}{}
}

func (c Roles) Exist(role string) bool {
    _, exist := c[role]
    return exist
}
