package vitali

import (
    "io"
    "os"
    "log"
    "fmt"
    "strings"
    "net/http"
    "html/template"
    "encoding/xml"
    "encoding/json"
)

func (c *webApp) marshalOutput(w *wrappedWriter, model *interface{}, ctx *Ctx, templateName string) {
    switch ctx.ChosenType {
    case "application/json":
        fmt.Fprintf(w, "%s", string(panicOnErr(json.Marshal(model)).([]byte)))
    case "application/xml":
        fmt.Fprintf(w, "%s", string(panicOnErr(xml.Marshal(model)).([]byte)))
    case "text/html":
        m := struct{
            S map[string]template.HTML
            M *interface{}
            C *Ctx
            W *webApp
        }{
            c.I18n[ctx.ChosenLang],
            model,
            ctx,
            c,
        }
        c.views[templateName].Execute(w, m)
    default:
        fmt.Fprintf(w, "%s", *model)
    }
}

func (c *webApp) writeResponse(w *wrappedWriter, r *http.Request, response *interface{}, ctx *Ctx, templateName string) {
    switch v := (*response).(type) {
    case noContent:
        w.WriteHeader(http.StatusNoContent)
    case movedPermanently:
        w.Header().Set("Location", v.uri)
        w.WriteHeader(http.StatusMovedPermanently)
        fmt.Fprintf(w, "%s\n", v.uri)
    case found:
        w.Header().Set("Location", v.uri)
        w.WriteHeader(http.StatusFound)
        fmt.Fprintf(w, "%s\n", v.uri)
    case seeOther:
        w.Header().Set("Location", v.uri)
        w.WriteHeader(http.StatusSeeOther)
        fmt.Fprintf(w, "%s\n", v.uri)
    case tempRedirect:
        w.Header().Set("Location", v.uri)
        w.WriteHeader(http.StatusTemporaryRedirect)
    case badRequest:
        if v.body != nil {
            w.WriteHeader(http.StatusBadRequest)
            c.marshalOutput(w, &v.body, ctx, templateName)
        } else {
            http.Error(w, v.reason, http.StatusBadRequest)
        }
    case unauthorized:
        w.Header()["WWW-Authenticate"] = []string{v.wwwAuthHeader}
        if c.Settings["401_PAGE"] != "" && ctx.ChosenType == "text/html" {
            w.Header().Set("Content-Type", "text/html")
            w.WriteHeader(http.StatusUnauthorized)
            f, err := os.Open(c.Settings["401_PAGE"])
            if err != nil {
                log.Printf("open 401 page error: %s\n", err)
                return
            }
            defer f.Close()
            io.Copy(w, f)
        } else {
            if v.body != nil {
                w.WriteHeader(http.StatusUnauthorized)
                c.marshalOutput(w, &v.body, ctx, templateName)
            } else {
                http.Error(w, "unauthorized", http.StatusUnauthorized)
            }
        }
    case forbidden:
        if c.Settings["403_PAGE"] != "" && ctx.ChosenType == "text/html" {
            w.Header().Set("Content-Type", "text/html")
            w.WriteHeader(http.StatusForbidden)
            f, err := os.Open(c.Settings["403_PAGE"])
            if err != nil {
                log.Printf("open 403 page error: %s\n", err)
                return
            }
            defer f.Close()
            io.Copy(w, f)
        } else {
            if v.body != nil {
                w.WriteHeader(http.StatusForbidden)
                c.marshalOutput(w, &v.body, ctx, templateName)
            } else {
                http.Error(w, "Forbidden", http.StatusForbidden)
            }
        }
    case notFound:
        if v.body != nil {
            w.WriteHeader(http.StatusNotFound)
            c.marshalOutput(w, &v.body, ctx, templateName)
        } else {
            http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
        }
    case methodNotAllowed:
        w.Header().Set("Allow", strings.Join(v.allowed, ", "))
        http.Error(w, http.StatusText(http.StatusMethodNotAllowed),
            http.StatusMethodNotAllowed)
    case notAcceptable:
        w.Header().Set("Content-Type", "text/csv")
        types := make([]string, len(v.provided))
        for i, mediaType := range(v.provided){
            types[i] = string(mediaType)
        }
        http.Error(w, strings.Join(([]string)(types), ","),
            http.StatusNotAcceptable)
    case unsupportedMediaType:
        if v.body != nil {
            w.WriteHeader(http.StatusUnsupportedMediaType)
            c.marshalOutput(w, &v.body, ctx, templateName)
        } else {
            http.Error(w, http.StatusText(http.StatusUnsupportedMediaType),
                http.StatusUnsupportedMediaType)
        }
    case internalError:
        w.err = v
        if c.ErrTemplate != nil {
            w.WriteHeader(http.StatusInternalServerError)
            md := struct {Code uint32}{w.err.code}
            c.ErrTemplate.Execute(w, md)
        } else {
            http.Error(w, fmt.Sprintf("%s: %d", http.StatusText(http.StatusInternalServerError),
                w.err.code), http.StatusInternalServerError)
        }
    case notImplemented:
        if v.body != nil {
            w.WriteHeader(http.StatusNotImplemented)
            c.marshalOutput(w, &v.body, ctx, templateName)
        } else {
            http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
        }
    case serviceUnavailable:
        if v.seconds >= 0 {
            w.Header().Set("Retry-After", fmt.Sprintf("%d", v.seconds))
        }
        if v.body != nil {
            w.WriteHeader(http.StatusServiceUnavailable)
            c.marshalOutput(w, &v.body, ctx, templateName)
        } else {
            http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
        }
    case error:
        w.err = internalError{
            where: "",
            why: v.Error(),
            code: errorCode(v.Error()),
        }
        http.Error(w, fmt.Sprintf("%s: %d", http.StatusText(http.StatusInternalServerError),
            w.err.code), http.StatusInternalServerError)
    case io.ReadCloser:
        defer v.Close()
        if r.Header.Get("Range") != "" {
            w.WriteHeader(http.StatusPartialContent)
        } else {
            w.WriteHeader(http.StatusOK)
        }
        io.Copy(w, v)
    case http.ResponseWriter:
    case clientGone:
    default:
        if r.Header.Get("Range") != "" {
            w.WriteHeader(http.StatusPartialContent)
        } else {
            w.WriteHeader(http.StatusOK)
        }
        if ctx.ChosenType != "" {
            c.marshalOutput(w, &v, ctx, templateName)
        } else {
            fmt.Fprintf(w, "%s", v)
        }
    }
}
