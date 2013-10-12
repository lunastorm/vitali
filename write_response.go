package vitali

import (
    "fmt"
    "strings"
    "net/http"
)

func writeResponse(w *wrappedWriter, r *http.Request, response interface{}) {
    switch v := response.(type) {
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
    case notFound:
        http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
    case methodNotAllowed:
        w.Header().Set("Allow", strings.Join(v.allowed, ", "))
        http.Error(w, http.StatusText(http.StatusMethodNotAllowed),
            http.StatusMethodNotAllowed)
    case unsupportedMediaType:
        http.Error(w, http.StatusText(http.StatusUnsupportedMediaType),
            http.StatusUnsupportedMediaType)
    case internalError:
        w.err = v
        http.Error(w, fmt.Sprintf("%s: %d", http.StatusText(http.StatusInternalServerError),
            w.err.code), http.StatusInternalServerError)
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
        fmt.Fprintf(w, "%s", v)
    }
}
