package vitali

import (
    "net/http"
)

type UserProvider interface {
    AuthHeader(*http.Request) string
    GetUserAndRole(*http.Request) (string, string)
}

type EmptyUserProvider struct {
}

func (c EmptyUserProvider) AuthHeader(r *http.Request) string {
    return ""
}

func (c EmptyUserProvider) GetUserAndRole(r *http.Request) (string, string) {
    return "", ""
}
