package vitali

import (
    "fmt"
    "strings"
    "net/http"
    "encoding/xml"
    "encoding/json"
)

func marshalOutput(input interface{}, contentType MediaType) string {
    switch contentType {
    case "application/json":
        j, err := json.Marshal(input)
        if err != nil {
            panic(err)
        }
        return string(j)
    case "application/xml":
        x, err := xml.Marshal(input)
        if err != nil {
            panic(err)
        }
        return string(x)
    default:
        return fmt.Sprintf("%s", input)
    }
}

func (c webApp) writeResponse(w *wrappedWriter, r *http.Request, response interface{}, chosenType MediaType) {
    switch v := response.(type) {
    case noContent:
        w.WriteHeader(http.StatusNoContent)
    case view:
        w.Header().Set("Content-Type", "text/html")
        w.WriteHeader(http.StatusOK)
        v.template.Execute(w, v.model)
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
        http.Error(w, "unauthorized", http.StatusUnauthorized)
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
    case error:
        w.err = internalError{
            where: "",
            why: v.Error(),
            code: errorCode(v.Error()),
        }
        http.Error(w, fmt.Sprintf("%s: %d", http.StatusText(http.StatusInternalServerError),
            w.err.code), http.StatusInternalServerError)
    case http.ResponseWriter:
    case clientGone:
    default:
        w.WriteHeader(http.StatusOK)
        if chosenType != "" {
            fmt.Fprintf(w, "%s", marshalOutput(v, chosenType))
        } else {
            fmt.Fprintf(w, "%s", v)
        }
    }
}
