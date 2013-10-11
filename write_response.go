package vitali

import (
    "fmt"
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
        http.Error(w, "not found", http.StatusNotFound)
    case internalError:
        w.err = v
        http.Error(w, fmt.Sprintf("internal server error: %d", w.err.code),
            http.StatusInternalServerError)
    case error:
        w.err = internalError{
            where: "",
            why: v.Error(),
            code: errorCode(v.Error()),
        }
        http.Error(w, fmt.Sprintf("internal server error: %d", w.err.code),
            http.StatusInternalServerError)
    case clientGone:
    default:
        w.WriteHeader(http.StatusOK)
        fmt.Fprintf(w, "%s", v)
    }
}
