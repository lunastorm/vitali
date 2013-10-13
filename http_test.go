package vitali

import (
    "testing"
    "net/http"
    "net/url"
    "net/http/httptest"
)

func TestNotFound(t *testing.T) {
    r := &http.Request{
        Method: "GET",
        Host:   "lunastorm.tw",
        URL: &url.URL{
            Path: "/",
        },
    }
    rr := httptest.NewRecorder()
    webapp := CreateWebApp([]RouteRule{})
    webapp.ServeHTTP(rr, r)

    if rr.Code != http.StatusNotFound {
        t.Errorf("response code is %d", rr.Code)
    }
}

type Root struct {
    Ctx
}

func (c Root) Get() interface{} {
    return "root"
}

func TestOK(t *testing.T) {
    r := &http.Request{
        Method: "GET",
        Host:   "lunastorm.tw",
        URL: &url.URL{
            Path: "/",
        },
    }
    rr := httptest.NewRecorder()
    webapp := CreateWebApp([]RouteRule{
        {"/", Root{}},
    })
    webapp.ServeHTTP(rr, r)

    if rr.Code != http.StatusOK {
        t.Errorf("response code is %d", rr.Code)
    }
    entity := rr.Body.String()
    if entity != "root" {
        t.Errorf("entity is `%s`", entity)
    }
}

func TestMethodNotAllowed(t *testing.T) {
    r := &http.Request{
        Method: "POST",
        Host:   "lunastorm.tw",
        URL: &url.URL{
            Path: "/",
        },
    }
    rr := httptest.NewRecorder()
    webapp := CreateWebApp([]RouteRule{
        {"/", Root{}},
    })
    webapp.ServeHTTP(rr, r)

    if rr.Code != http.StatusMethodNotAllowed {
        t.Errorf("response code is %d", rr.Code)
    }
    allowed := rr.HeaderMap.Get("Allow")
    if allowed != "GET, HEAD" {
        t.Errorf("allow header is %s", allowed)
    }
}

func TestNotImplemented(t *testing.T) {
    r := &http.Request{
        Method: "WTF",
        Host:   "lunastorm.tw",
        URL: &url.URL{
            Path: "/",
        },
    }
    rr := httptest.NewRecorder()
    webapp := CreateWebApp([]RouteRule{
        {"/", Root{}},
    })
    webapp.ServeHTTP(rr, r)

    if rr.Code != http.StatusNotImplemented {
        t.Errorf("response code is %d", rr.Code)
    }
}

type Something struct {
    Ctx
}

func (c Something) Get() interface{} {
    return c.PathParam("id1") + c.PathParam("id2")
}

func TestPathParam(t *testing.T) {
    r := &http.Request{
        Method: "GET",
        Host:   "lunastorm.tw",
        URL: &url.URL{
            Path: "/foo/123/bar/456/something",
        },
    }
    rr := httptest.NewRecorder()
    webapp := CreateWebApp([]RouteRule{
        {"/foo/{id1}/bar/{id2}/something", Something{}},
    })
    webapp.ServeHTTP(rr, r)

    if rr.Code != http.StatusOK {
        t.Errorf("response code is %d", rr.Code)
    }
    entity := rr.Body.String()
    if entity != "123456" {
        t.Errorf("entity is `%s`", entity)
    }
}

type NeedAuth struct {
    Ctx
    Perm
}

func (c NeedAuth) Get() interface{} {
    return c.Username
}

func (c NeedAuth) Post() interface{} {
    return "public post"
}

func (c NeedAuth) Delete() interface{} {
    return c.Username
}

func TestNeedAuth(t *testing.T) {
    r := &http.Request{
        Method: "GET",
        Host:   "lunastorm.tw",
        URL: &url.URL{
            Path: "/needauth",
        },
    }
    rr := httptest.NewRecorder()
    webapp := CreateWebApp([]RouteRule{
        {"/needauth", NeedAuth{
            Perm: Perm{"GET": AUTHENTICATED, "POST": PUBLIC, "*": AUTHENTICATED},
        }},
    })
    webapp.ServeHTTP(rr, r)
    if rr.Code != http.StatusUnauthorized {
        t.Errorf("response code is %d", rr.Code)
    }

    r.Method = "POST"
    rr = httptest.NewRecorder()
    webapp.ServeHTTP(rr, r)
    if rr.Code != http.StatusOK {
        t.Errorf("response code is %d", rr.Code)
    }
    entity := rr.Body.String()
    if entity != "public post" {
        t.Errorf("entity is `%s`", entity)
    }

    r.Method = "DELETE"
    rr = httptest.NewRecorder()
    webapp.ServeHTTP(rr, r)
    if rr.Code != http.StatusUnauthorized {
        t.Errorf("response code is %d", rr.Code)
    }
}

type Auther struct {
}

func (c Auther) User(r *http.Request) string {
    return "bob"
}

func (c Auther) AuthHeader(r *http.Request) string {
    return `Basic realm="test"`
}

func TestAuthed(t *testing.T) {
    r := &http.Request{
        Method: "GET",
        Host:   "lunastorm.tw",
        URL: &url.URL{
            Path: "/needauth",
        },
    }
    rr := httptest.NewRecorder()
    webapp := CreateWebApp([]RouteRule{
        {"/needauth", NeedAuth{
            Perm: Perm{"GET": AUTHENTICATED, "POST": PUBLIC, "*": AUTHENTICATED},
        }},
    })
    webapp.UserProvider = Auther{}
    webapp.ServeHTTP(rr, r)
    if rr.Code != http.StatusOK {
        t.Errorf("response code is %d", rr.Code)
    }
    entity := rr.Body.String()
    if entity != "bob" {
        t.Errorf("user is `%s`", entity)
    }
}

type Panikr struct {
    Ctx
}

func (c Panikr) Get() interface{} {
    panic("panic!!")
    return ""
}

func TestRecoverPanic(t *testing.T) {
    r := &http.Request{
        Method: "GET",
        Host:   "lunastorm.tw",
        URL: &url.URL{
            Path: "/panic",
        },
    }
    rr := httptest.NewRecorder()
    webapp := CreateWebApp([]RouteRule{
        {"/panic", Panikr{}},
    })
    webapp.ServeHTTP(rr, r)

    if rr.Code != http.StatusInternalServerError {
        t.Errorf("response code is %d", rr.Code)
    }
}
