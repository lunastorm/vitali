package vitali

import (
    "net/http"
)

type UserProvider interface {
    AuthHeader(*http.Request) string
    User(*http.Request) string
}

type EmptyUserProvider struct {
}

func (c EmptyUserProvider) AuthHeader(r *http.Request) string {
    return ""
}

func (c EmptyUserProvider) User(r *http.Request) string {
    return ""
}
