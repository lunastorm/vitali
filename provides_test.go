package vitali

import (
    "testing"
    "net/http"
    "net/url"
    "net/http/httptest"
)

type Provider struct {
    Ctx
    Provides `GET:"application/json,text/html"`
    Views `GET:"provider_test.html"`
}

type ProviderModel struct {
    Foo string `json:"foo"`
}

func (c *Provider) Get() interface{} {
    return ProviderModel{"bar"}
}

func TestProviderJson(t *testing.T) {
    r := &http.Request{
        Method: "GET",
        Host:   "lunastorm.tw",
        URL: &url.URL{
            Path: "/provider",
        },
        Header: make(http.Header),
    }
    r.Header.Set("Accept", "application/json")

    rr := httptest.NewRecorder()
    webapp := CreateWebApp([]RouteRule{
        {"/provider", Provider{}},
    })
    webapp.ServeHTTP(rr, r)

    if rr.Code != http.StatusOK {
        t.Errorf("response code is %d", rr.Code)
    }
    entity := rr.Body.String()
    if entity != `{"foo":"bar"}` {
        t.Errorf("entity is `%s`", entity)
    }
}

func TestProviderEmptyAccept(t *testing.T) {
    r := &http.Request{
        Method: "GET",
        Host:   "lunastorm.tw",
        URL: &url.URL{
            Path: "/provider",
        },
        Header: make(http.Header),
    }

    rr := httptest.NewRecorder()
    webapp := CreateWebApp([]RouteRule{
        {"/provider", Provider{}},
    })
    webapp.ServeHTTP(rr, r)

    if rr.Code != http.StatusOK {
        t.Errorf("response code is %d", rr.Code)
    }
    entity := rr.Body.String()
    if entity != `{"foo":"bar"}` {
        t.Errorf("entity is `%s`", entity)
    }
}

func TestProviderAcceptAll(t *testing.T) {
    r := &http.Request{
        Method: "GET",
        Host:   "lunastorm.tw",
        URL: &url.URL{
            Path: "/provider",
        },
        Header: make(http.Header),
    }
    r.Header.Set("Accept", "*/*")

    rr := httptest.NewRecorder()
    webapp := CreateWebApp([]RouteRule{
        {"/provider", Provider{}},
    })
    webapp.ServeHTTP(rr, r)

    if rr.Code != http.StatusOK {
        t.Errorf("response code is %d", rr.Code)
    }
    entity := rr.Body.String()
    if entity != `{"foo":"bar"}` {
        t.Errorf("entity is `%s`", entity)
    }
}

func TestProviderPartial(t *testing.T) {
    r := &http.Request{
        Method: "GET",
        Host:   "lunastorm.tw",
        URL: &url.URL{
            Path: "/provider",
        },
        Header: make(http.Header),
    }
    r.Header.Set("Accept", "text/*")

    rr := httptest.NewRecorder()
    webapp := CreateWebApp([]RouteRule{
        {"/provider", Provider{}},
    })
    webapp.ServeHTTP(rr, r)

    if rr.Code != http.StatusOK {
        t.Errorf("response code is %d", rr.Code)
    }
    entity := rr.Body.String()
    if entity != "foobar\n" {
        t.Errorf("entity is `%s`", entity)
    }
}

func TestProviderQ(t *testing.T) {
    r := &http.Request{
        Method: "GET",
        Host:   "lunastorm.tw",
        URL: &url.URL{
            Path: "/provider",
        },
        Header: make(http.Header),
    }
    r.Header.Set("Accept", "text/*; q=0.8, application/json; q=0.9")

    rr := httptest.NewRecorder()
    webapp := CreateWebApp([]RouteRule{
        {"/provider", Provider{}},
    })
    webapp.ServeHTTP(rr, r)

    if rr.Code != http.StatusOK {
        t.Errorf("response code is %d", rr.Code)
    }
    entity := rr.Body.String()
    if entity != `{"foo":"bar"}` {
        t.Errorf("entity is `%s`", entity)
    }
}

func TestNotAcceptable(t *testing.T) {
    r := &http.Request{
        Method: "GET",
        Host:   "lunastorm.tw",
        URL: &url.URL{
            Path: "/provider",
        },
        Header: make(http.Header),
    }
    r.Header.Set("Accept", "application/xml")

    rr := httptest.NewRecorder()
    webapp := CreateWebApp([]RouteRule{
        {"/provider", Provider{}},
    })
    webapp.ServeHTTP(rr, r)

    if rr.Code != http.StatusNotAcceptable {
        t.Errorf("response code is %d", rr.Code)
    }
}
