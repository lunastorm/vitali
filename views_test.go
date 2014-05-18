package vitali

import (
    "testing"
    "net/http"
    "net/url"
    "net/http/httptest"
)

type I18nTest struct {
    Ctx
    Provides `GET:"text/html"`
    Views `GET:"i18n_test.html"`
}

func (c *I18nTest) Get() interface{} {
    return ""
}

type TestLangProvider struct {
}

func (c *TestLangProvider) Select(ctx *Ctx) string {
    return ctx.Header("Accept-Language")
}

func TestI18n(t *testing.T) {
    r := &http.Request{
        Method: "GET",
        Host:   "lunastorm.tw",
        URL: &url.URL{
            Path: "/i18n",
        },
        Header: make(http.Header),
    }
    r.Header.Set("Accept-Language", "zh-tw")

    rr := httptest.NewRecorder()
    webapp := CreateWebApp([]RouteRule{
        {"/i18n", I18nTest{}},
    })
    webapp.LangProvider = &TestLangProvider{}
    webapp.ServeHTTP(rr, r)

    if rr.Code != http.StatusOK {
        t.Errorf("response code is %d", rr.Code)
    }
    entity := rr.Body.String()
    if entity != "嗨\n" {
        t.Errorf("entity is `%s`", entity)
    }

    r.Header.Set("Accept-Language", "en-us")
    rr = httptest.NewRecorder()
    webapp.ServeHTTP(rr, r)
    if rr.Code != http.StatusOK {
        t.Errorf("response code is %d", rr.Code)
    }
    entity = rr.Body.String()
    if entity != "hi\n" {
        t.Errorf("entity is `%s`", entity)
    }
}

type View1 struct {
    Ctx
    Provides `GET:"text/html" POST:"text/html"`
    Views `GET:"base.html,foo_get.html" POST:"base.html,foo_post.html"`
}

func (c *View1) Get() interface{} {
    return "foo"
}

func (c *View1) Post() interface{} {
    return "bar"
}

func TestView1Get(t *testing.T) {
    r := &http.Request{
        Method: "GET",
        Host:   "lunastorm.tw",
        URL: &url.URL{
            Path: "/view1",
        },
        Header: make(http.Header),
    }

    rr := httptest.NewRecorder()
    webapp := CreateWebApp([]RouteRule{
        {"/view1", View1{}},
    })
    webapp.ServeHTTP(rr, r)

    if rr.Code != http.StatusOK {
        t.Errorf("response code is %d", rr.Code)
    }
    entity := rr.Body.String()
    if entity != "<html>get foo</html>\n" {
        t.Errorf("entity is `%s`", entity)
    }
}

func TestView1Post(t *testing.T) {
    r := &http.Request{
        Method: "POST",
        Host:   "lunastorm.tw",
        URL: &url.URL{
            Path: "/view1",
        },
        Header: make(http.Header),
    }

    rr := httptest.NewRecorder()
    webapp := CreateWebApp([]RouteRule{
        {"/view1", View1{}},
    })
    webapp.ServeHTTP(rr, r)

    if rr.Code != http.StatusOK {
        t.Errorf("response code is %d", rr.Code)
    }
    entity := rr.Body.String()
    if entity != "<html>post bar</html>\n" {
        t.Errorf("entity is `%s`", entity)
    }
}

type ViewCtx struct {
    Ctx
    Provides `GET:"text/html"`
    Views `GET:"ctx.html"`
}

func (c *ViewCtx) Get() interface{} {
    return ""
}

func TestViewCtx(t *testing.T) {
    r := &http.Request{
        Method: "GET",
        Host:   "lunastorm.tw",
        URL: &url.URL{
            Path: "/viewctx",
        },
        Header: make(http.Header),
    }
    r.Header.Set("Accept-Language", "en-us")

    rr := httptest.NewRecorder()
    webapp := CreateWebApp([]RouteRule{
        {"/viewctx", ViewCtx{}},
    })
    webapp.LangProvider = &TestLangProvider{}
    webapp.ServeHTTP(rr, r)

    if rr.Code != http.StatusOK {
        t.Errorf("response code is %d", rr.Code)
    }
    entity := rr.Body.String()
    if entity != "en-us\n" {
        t.Errorf("entity is `%s`", entity)
    }
}

type ViewWebapp struct {
    Ctx
    Provides `GET:"text/html"`
    Views `GET:"webapp.html"`
}

func (c *ViewWebapp) Get() interface{} {
    return ""
}

func TestViewWebapp(t *testing.T) {
    r := &http.Request{
        Method: "GET",
        Host:   "lunastorm.tw",
        URL: &url.URL{
            Path: "/viewwebapp",
        },
        Header: make(http.Header),
    }

    rr := httptest.NewRecorder()
    webapp := CreateWebApp([]RouteRule{
        {"/viewwebapp", ViewWebapp{}},
    })
    webapp.Settings["TEST_CONFIG"] = "test config"
    webapp.LangProvider = &TestLangProvider{}
    webapp.ServeHTTP(rr, r)

    if rr.Code != http.StatusOK {
        t.Errorf("response code is %d", rr.Code)
    }
    entity := rr.Body.String()
    if entity != "test config\n" {
        t.Errorf("entity is `%s`", entity)
    }
}
