package vitali

import (
    "io"
    "os"
    "log"
    "fmt"
    "strings"
    "net/http"
    "encoding/xml"
    "encoding/json"
)

func (c webApp) marshalOutput(w *wrappedWriter, input interface{}, contentType MediaType, templateName string) {
    switch contentType {
    case "application/json":
        fmt.Fprintf(w, "%s", string(panicOnErr(json.Marshal(input)).([]byte)))
    case "application/xml":
        fmt.Fprintf(w, "%s", string(panicOnErr(xml.Marshal(input)).([]byte)))
    case "text/html":
        c.views[templateName].Execute(w, input)
    default:
        fmt.Fprintf(w, "%s", input)
    }
}

func (c webApp) writeResponse(w *wrappedWriter, r *http.Request, response interface{}, chosenType MediaType, templateName string) {
    switch v := response.(type) {
    case noContent:
        w.WriteHeader(http.StatusNoContent)
    case view:
        w.Header().Set("Content-Type", "text/html")
        w.WriteHeader(http.StatusOK)
        v.template.Execute(w, v.model)
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
    case badRequest:
        http.Error(w, v.reason, http.StatusBadRequest)
    case unauthorized:
        w.Header()["WWW-Authenticate"] = []string{v.wwwAuthHeader}
        if c.Settings["401_PAGE"] != "" && chosenType == "text/html" {
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
            http.Error(w, "unauthorized", http.StatusUnauthorized)
        }
    case forbidden:
        http.Error(w, "Forbidden", http.StatusForbidden)
    case notFound:
        http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
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
        http.Error(w, http.StatusText(http.StatusUnsupportedMediaType),
            http.StatusUnsupportedMediaType)
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
        http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
    case serviceUnavailable:
        if v.seconds >= 0 {
            w.Header().Set("Retry-After", fmt.Sprintf("%d", v.seconds))
        }
        http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
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
        io.Copy(w, v)
    case http.ResponseWriter:
    case clientGone:
    default:
        w.WriteHeader(http.StatusOK)
        if chosenType != "" {
            c.marshalOutput(w, v, chosenType, templateName)
        } else {
            fmt.Fprintf(w, "%s", v)
        }
    }
}
