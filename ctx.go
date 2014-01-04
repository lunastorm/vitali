package vitali

import (
    "net/http"
)

type Ctx struct {
    Username    string
    Roles       Roles
    Request *http.Request
    ResponseWriter  http.ResponseWriter
    ChosenType  MediaType
    ContentType MediaType

    pathParams map[string]string
}

func (c Ctx) AddHeader(key string, value string) {
    c.ResponseWriter.Header().Add(key, value)
}

func (c Ctx) Cookie(name string) (value string) {
    cookie, _ := c.Request.Cookie(name)
    return cookie.Value
}

func (c Ctx) SetCookie(cookie *http.Cookie) {
    http.SetCookie(c.ResponseWriter, cookie)
}

func (c Ctx) Param(key string) string {
    return c.Request.Form.Get(key)
}

func (c Ctx) ParamArray(key string) []string {
    return c.Request.Form[key]
}

func (c Ctx) PathParam(key string) string {
    return c.pathParams[key]
}

func (c Ctx) Header(key string) string {
    return c.Request.Header.Get(key)
}
