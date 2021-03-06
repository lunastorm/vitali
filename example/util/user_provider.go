package util

import (
    "strings"
    "net/http"
    "encoding/base64"
)

type UserProvider struct {
}

func (c *UserProvider) AuthHeader(r *http.Request) (WWWAuthenticate string) {
    return `Basic realm="vitali"`
}

func (c *UserProvider) GetUserAndRoles(r *http.Request) (user string, roles []string) {
    roles = make([]string, 0)
    authHeader := r.Header.Get("Authorization")
    if !strings.HasPrefix(authHeader, "Basic ") {
        return
    }

    data, err := base64.StdEncoding.DecodeString(strings.SplitN(authHeader, " ", 2)[1])
    if err != nil {return}
    tmp := strings.SplitN(string(data), ":", 2)
    user = tmp[0]
    password := tmp[1]

    if user == "foo" && password == "bar" {
        return user, []string{"AUTHED"}
    } else {
        user = ""
        return
    }
}
