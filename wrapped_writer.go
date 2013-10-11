package vitali

import (
    "time"
    "net/http"
)

type wrappedWriter struct {
    status int
    writer http.ResponseWriter
    inTime time.Time
    written int
    err internalError
}

func (c *wrappedWriter) Header() http.Header {
    return c.writer.Header()
}

func (c *wrappedWriter) Write(buf []byte) (int, error) {
    c.written += len(buf)
    return c.writer.Write(buf)
}

func (c *wrappedWriter) WriteHeader(status int) {
    c.status = status
    c.writer.WriteHeader(status)
}

func (c *wrappedWriter) CloseNotify() <-chan bool {
    return c.writer.(http.CloseNotifier).CloseNotify()
}
