package vitali

import (
    "net/http"
)

type LangProvider interface {
    Select(*http.Request) string
}

type EmptyLangProvider struct {
}

func (c EmptyLangProvider) Select(r *http.Request) string {
    return ""
}
