package vitali

import (
    "testing"
    "net/http"
    "net/url"
    "net/http/httptest"
)

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
