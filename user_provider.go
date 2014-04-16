package vitali

import (
    "net/http"
)

type UserProvider interface {
    AuthHeader(*http.Request) string
    GetUserAndRoles(*http.Request) (user string, roles []string)
}

type EmptyUserProvider struct {
}

func (c EmptyUserProvider) AuthHeader(r *http.Request) string {
    return ""
}

func (c EmptyUserProvider) GetUserAndRoles(r *http.Request) (string, []string) {
    return "", []string{""}
}
