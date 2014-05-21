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

func (c *Root) post() interface{} {
    return ""
}

func (c *Root) Get() interface{} {
    return "root"
}

func (c *Root) Test() interface{} {
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

func TestPartial(t *testing.T) {
    r := &http.Request{
        Method: "GET",
        Host:   "lunastorm.tw",
        URL: &url.URL{
            Path: "/",
        },
        Header: make(http.Header),
    }
    r.Header.Set("Range", "bytes=0-")
    rr := httptest.NewRecorder()
    webapp := CreateWebApp([]RouteRule{
        {"/", Root{}},
    })
    webapp.ServeHTTP(rr, r)

    if rr.Code != http.StatusPartialContent {
        t.Errorf("response code is %d", rr.Code)
    }
    entity := rr.Body.String()
    if entity != "root" {
        t.Errorf("entity is `%s`", entity)
    }
}

type Nothing struct {
    Ctx
}

func (c Nothing) Get() interface{} {
    return c.NoContent()
}

func TestNoContent(t *testing.T) {
    r := &http.Request{
        Method: "GET",
        Host:   "lunastorm.tw",
        URL: &url.URL{
            Path: "/nothing",
        },
    }
    rr := httptest.NewRecorder()
    webapp := CreateWebApp([]RouteRule{
        {"/nothing", Nothing{}},
    })
    webapp.ServeHTTP(rr, r)

    if rr.Code != http.StatusNoContent {
        t.Errorf("response code is %d", rr.Code)
    }
}

type Info struct {
    Ctx
    Provided string
    Provided2 string
}

func (c Info) Get() interface{} {
    return c.Provided + c.Provided2
}

func TestProvided(t *testing.T) {
    r := &http.Request{
        Method: "GET",
        Host:   "lunastorm.tw",
        URL: &url.URL{
            Path: "/info",
        },
    }
    rr := httptest.NewRecorder()
    webapp := CreateWebApp([]RouteRule{
        {"/info", Info{
            Provided: "foo",
            Provided2: "bar",
            }},
    })
    webapp.ServeHTTP(rr, r)

    if rr.Code != http.StatusOK {
        t.Errorf("response code is %d", rr.Code)
    }
    entity := rr.Body.String()
    if entity != "foobar" {
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
    if allowed != "HEAD, GET, TEST" {
        t.Errorf("allow header is %s", allowed)
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

type PathOptional struct {
    Ctx
}

func (c PathOptional) Get() interface{} {
    return c.PathParam("id")
}

func TestOptionalPathParam(t *testing.T) {
    r := &http.Request{
        Method: "GET",
        Host:   "lunastorm.tw",
        URL: &url.URL{
            Path: "/foo/123",
        },
    }
    rr := httptest.NewRecorder()
    webapp := CreateWebApp([]RouteRule{
        {"/foo/{id}", PathOptional{}},
    })
    webapp.ServeHTTP(rr, r)

    if rr.Code != http.StatusOK {
        t.Errorf("response code is %d", rr.Code)
    }
    entity := rr.Body.String()
    if entity != "123" {
        t.Errorf("entity is `%s`", entity)
    }

    r.URL.Path = "/foo"
    rr = httptest.NewRecorder()
    webapp.ServeHTTP(rr, r)
    if rr.Code != http.StatusOK {
        t.Errorf("response code is %d", rr.Code)
    }
    entity = rr.Body.String()
    if entity != "" {
        t.Errorf("entity is `%s`", entity)
    }
}

type Something2 struct {
    Ctx
}

func (c Something2) Get() interface{} {
    if c.HasParam("foo") {
        return "foo"
    } else {
        return c.Param("id")
    }
}

func TestForm(t *testing.T) {
    r := &http.Request{
        Method: "GET",
        Host:   "lunastorm.tw",
        URL: &url.URL{
            Path: "/something2",
        },
        Form: make(url.Values),
    }
    r.Form.Add("id", "5566")
    rr := httptest.NewRecorder()
    webapp := CreateWebApp([]RouteRule{
        {"/something2", Something2{}},
    })
    webapp.ServeHTTP(rr, r)

    if rr.Code != http.StatusOK {
        t.Errorf("response code is %d", rr.Code)
    }
    entity := rr.Body.String()
    if entity != "5566" {
        t.Errorf("id is `%s`", entity)
    }

    rr = httptest.NewRecorder()
    r.Form.Add("foo", "bar")
    webapp.ServeHTTP(rr, r)
    if rr.Code != http.StatusOK {
        t.Errorf("response code is %d", rr.Code)
    }
    entity = rr.Body.String()
    if entity != "foo" {
        t.Errorf("id is `%s`", entity)
    }
}

type HeaderFoo struct {
    Ctx
}

func (c HeaderFoo) Get() interface{} {
    return c.Header("foo")
}

func TestHeader(t *testing.T) {
    r := &http.Request{
        Method: "GET",
        Host:   "lunastorm.tw",
        URL: &url.URL{
            Path: "/headerfoo",
        },
        Header: make(http.Header),
    }
    r.Header.Set("foo", "bar")
    rr := httptest.NewRecorder()
    webapp := CreateWebApp([]RouteRule{
        {"/headerfoo", HeaderFoo{}},
    })
    webapp.ServeHTTP(rr, r)

    if rr.Code != http.StatusOK {
        t.Errorf("response code is %d", rr.Code)
    }
    entity := rr.Body.String()
    if entity != "bar" {
        t.Errorf("entity is `%s`", entity)
    }
}

type NeedAuth struct {
    Ctx
    Perm `GET:"authed" DELETE:"authed"`
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
        {"/needauth", NeedAuth{}},
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

func (c Auther) GetUserAndRoles(r *http.Request) (string, []string) {
    return "bob", []string{"authed"}
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
        {"/needauth", NeedAuth{}},
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

type Bad struct {
    Ctx
}

func (c Bad) Get() interface{} {
    return c.BadRequest("reason")
}

func TestBadRequest(t *testing.T) {
    r := &http.Request{
        Method: "GET",
        Host:   "lunastorm.tw",
        URL: &url.URL{
            Path: "/bad",
        },
    }
    rr := httptest.NewRecorder()
    webapp := CreateWebApp([]RouteRule{
        {"/bad", Bad{}},
    })
    webapp.ServeHTTP(rr, r)

    if rr.Code != http.StatusBadRequest {
        t.Errorf("response code is %d", rr.Code)
    }
    entity := rr.Body.String()
    if entity != "reason\n" {
        t.Errorf("entity is `%s`", entity)
    }
}

type ConsumeSomething struct {
    Ctx
    Consumes `POST:"application/json,application/xml"`
}

func (c ConsumeSomething) Get() interface{} {
    return "get"
}

func (c ConsumeSomething) Post() interface{} {
    return "success"
}

func TestNoConsume(t *testing.T) {
    r := &http.Request{
        Method: "GET",
        Host:   "lunastorm.tw",
        URL: &url.URL{
            Path: "/consume",
        },
        Header: make(http.Header),
    }
    webapp := CreateWebApp([]RouteRule{
        {"/consume", ConsumeSomething{}},
    })

    rr := httptest.NewRecorder()
    webapp.ServeHTTP(rr, r)
    if rr.Code != http.StatusOK {
        t.Errorf("response code is %d", rr.Code)
    }
}

func TestConsume(t *testing.T) {
    r := &http.Request{
        Method: "POST",
        Host:   "lunastorm.tw",
        URL: &url.URL{
            Path: "/consume",
        },
        Header: make(http.Header),
    }
    webapp := CreateWebApp([]RouteRule{
        {"/consume", ConsumeSomething{}},
    })

    rr := httptest.NewRecorder()
    webapp.ServeHTTP(rr, r)
    if rr.Code != http.StatusUnsupportedMediaType {
        t.Errorf("response code is %d", rr.Code)
    }

    r.Header.Set("Content-Type", "application/json")
    rr = httptest.NewRecorder()
    webapp.ServeHTTP(rr, r)
    if rr.Code != http.StatusOK {
        t.Errorf("response code is %d", rr.Code)
    }

    r.Header.Set("Content-Type", "application/xml")
    rr = httptest.NewRecorder()
    webapp.ServeHTTP(rr, r)
    if rr.Code != http.StatusOK {
        t.Errorf("response code is %d", rr.Code)
    }

    r.Header.Set("Content-Type", "text/plain")
    rr = httptest.NewRecorder()
    webapp.ServeHTTP(rr, r)
    if rr.Code != http.StatusUnsupportedMediaType {
        t.Errorf("response code is %d", rr.Code)
    }
}

type Redirects struct {
    Ctx
}

func (c Redirects) Get() interface{} {
    switch c.Param("type") {
    case "moved":
        return c.MovedPermanently("/moved")
    case "found":
        return c.Found("/found")
    case "seeother":
        return c.SeeOther("/seeother")
    case "temp":
        return c.TempRedirect("/temp")
    }
    return ""
}

func TestRedirects(t *testing.T) {
    r := &http.Request{
        Method: "GET",
        Host:   "lunastorm.tw",
        URL: &url.URL{
            Path: "/redirects",
        },
        Form: make(url.Values),
    }
    webapp := CreateWebApp([]RouteRule{
        {"/redirects", Redirects{}},
    })

    r.Form.Set("type", "found")
    rr := httptest.NewRecorder()
    webapp.ServeHTTP(rr, r)
    if rr.Code != http.StatusFound {
        t.Errorf("response code is %d", rr.Code)
    }
    location := rr.Header().Get("Location")
    if location != "/found" {
        t.Errorf("location is `%s`", location)
    }

    r.Form.Set("type", "seeother")
    rr = httptest.NewRecorder()
    webapp.ServeHTTP(rr, r)
    if rr.Code != http.StatusSeeOther {
        t.Errorf("response code is %d", rr.Code)
    }
    location = rr.Header().Get("Location")
    if location != "/seeother" {
        t.Errorf("location is `%s`", location)
    }

    r.Form.Set("type", "moved")
    rr = httptest.NewRecorder()
    webapp.ServeHTTP(rr, r)
    if rr.Code != http.StatusMovedPermanently {
        t.Errorf("response code is %d", rr.Code)
    }
    location = rr.Header().Get("Location")
    if location != "/moved" {
        t.Errorf("location is `%s`", location)
    }

    r.Form.Set("type", "temp")
    rr = httptest.NewRecorder()
    webapp.ServeHTTP(rr, r)
    if rr.Code != http.StatusTemporaryRedirect {
        t.Errorf("response code is %d", rr.Code)
    }
    location = rr.Header().Get("Location")
    if location != "/temp" {
        t.Errorf("location is `%s`", location)
    }
}

type Forbid struct {
    Ctx
}

func (c Forbid) Get() interface{} {
    return c.Forbidden()
}

func TestForbidden(t *testing.T) {
    r := &http.Request{
        Method: "GET",
        Host:   "lunastorm.tw",
        URL: &url.URL{
            Path: "/forbid",
        },
    }
    rr := httptest.NewRecorder()
    webapp := CreateWebApp([]RouteRule{
        {"/forbid", Forbid{}},
    })
    webapp.ServeHTTP(rr, r)

    if rr.Code != http.StatusForbidden {
        t.Errorf("response code is %d", rr.Code)
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

type Unavailable struct {
    Ctx
    RetryIn int
}

func (c *Unavailable) Get() interface{} {
    return c.ServiceUnavailable(c.RetryIn)
}

func TestServiceUnavailableWithRetry(t *testing.T) {
    r := &http.Request{
        Method: "GET",
        Host:   "lunastorm.tw",
        URL: &url.URL{
            Path: "/unavail",
        },
    }
    rr := httptest.NewRecorder()
    webapp := CreateWebApp([]RouteRule{
        {"/unavail", Unavailable{RetryIn: 60}},
    })
    webapp.ServeHTTP(rr, r)

    if rr.Code != http.StatusServiceUnavailable {
        t.Errorf("response code is %d", rr.Code)
    }
    retryHeader := rr.HeaderMap.Get("Retry-After")
    if retryHeader != "60" {
        t.Errorf("retry header is %d", retryHeader)
    }
}

func TestServiceUnavailableWithoutRetry(t *testing.T) {
    r := &http.Request{
        Method: "GET",
        Host:   "lunastorm.tw",
        URL: &url.URL{
            Path: "/unavail",
        },
    }
    rr := httptest.NewRecorder()
    webapp := CreateWebApp([]RouteRule{
        {"/unavail", Unavailable{RetryIn: -1}},
    })
    webapp.ServeHTTP(rr, r)

    if rr.Code != http.StatusServiceUnavailable {
        t.Errorf("response code is %d", rr.Code)
    }
    retryHeader := rr.HeaderMap.Get("Retry-After")
    if retryHeader != "" {
        t.Errorf("retry header is %d", retryHeader)
    }
}

type HasPre struct {
    Ctx
}

func (c *HasPre) Pre() interface{} {
    if c.PathParam("id") == "1" {
        return nil
    } else {
        return c.NotFound()
    }
}

func (c *HasPre) Get() interface{} {
    return "haspre"
}

func TestHasPreOK(t *testing.T) {
    r := &http.Request{
        Method: "GET",
        Host:   "lunastorm.tw",
        URL: &url.URL{
            Path: "/haspre/1",
        },
    }
    rr := httptest.NewRecorder()
    webapp := CreateWebApp([]RouteRule{
        {"/haspre/{id}", HasPre{}},
    })
    webapp.ServeHTTP(rr, r)

    if rr.Code != http.StatusOK {
        t.Errorf("response code is %d", rr.Code)
    }
    entity := rr.Body.String()
    if entity != "haspre" {
        t.Errorf("wrong entity: %s", entity)
    }
}

func TestHasPreEarlyReturn(t *testing.T) {
    r := &http.Request{
        Method: "GET",
        Host:   "lunastorm.tw",
        URL: &url.URL{
            Path: "/haspre/2",
        },
    }
    rr := httptest.NewRecorder()
    webapp := CreateWebApp([]RouteRule{
        {"/haspre/{id}", HasPre{}},
    })
    webapp.ServeHTTP(rr, r)

    if rr.Code != http.StatusNotFound {
        t.Errorf("response code is %d", rr.Code)
    }
}
