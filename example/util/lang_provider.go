package util

import (
    "time"
    "strconv"
    "strings"
    "net/http"
    "html/template"
    "github.com/lunastorm/vitali"
)

type LangProvider struct {
    I18n map[string]map[string]template.HTML
}

func (c *LangProvider) Select(ctx *vitali.Ctx) (lang string) {
    lang = ctx.Cookie("lang")
    if lang != "" {
        return
    }
    lang = "en-us"

    acceptRaw := ctx.Header("Accept-Language")
    acceptLangs := strings.Split(strings.ToLower(acceptRaw), ",")
    currentQ := 0.0
    for _, l := range(acceptLangs) {
        q := 1.0
        tmp := strings.Split(l, ";")
        if len(tmp) == 2 {
            var err error
            q, err = strconv.ParseFloat(strings.Split(tmp[1], "=")[1], 32)
            if err != nil {
                continue
            }
        }
        if _, ok := c.I18n[tmp[0]]; ok {
            if q == 1.0 {
                lang = tmp[0]
                break
            } else if q > currentQ {
                lang = tmp[0]
                currentQ = q
            }
        }
    }

    cookie := &http.Cookie{
        Name: "lang",
        Value: lang,
        Path: "/",
        Expires: time.Now().Add(10*365*24*time.Hour),
        Secure: false,
        HttpOnly: false,
    }
    ctx.SetCookie(cookie)
    return
}
