vitali
======
## Install
Make sure you have configured the go runtime correctly first.

```
$ go get github.com/lunastorm/vitali
```

## Run example
```
$ cd $GOPATH/src/github.com/lunastorm/vitali/example
$ go run main.go
2014/04/11 01:50:22 starting server at port 8080...
```

Open http://foo:bar@127.0.0.1:8080/user/foo/slide in the browser, and you can create a new slide or edit the example slide.

## Basic webapp folder structure
You can place almost everything in the base folder. However, you should create the "views" subfolder which is where you put the template html files, and also the i18n.json dictionary.

## Create your first resource
resources/foo.go
```
package resources
import (
    "github.com/lunastorm/vitali"
)

type Foo struct {
    vitali.Ctx
}

func (c *Foo) Get() interface{} {
    return "hello world"
}
```
Every resource struct should embed the _vitali.Ctx_ struct. Then you implement the GET or other methods for it.

## Serve the webapp
main.go
```
package main

import (
    "log"
    "net/http"
    "github.com/lunastorm/vitali"
    "./resources"
)

func main() {
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
    webapp := vitali.CreateWebApp([]vitali.RouteRule{
        {"/foo", resources.Foo{
        }}, 
    })  
    http.Handle("/", webapp)
    log.Printf("starting server at port 8080...")
    http.ListenAndServe(":8080", nil)
}
```

## Routing
The first member of vitali.RouteRule is the path expression, and the second is the target resource's prototype instance. _{thisPresentsAPathParameter}_ You can access the path parameter through the PathParam method, here is an example:
```
webapp := vitali.CreateWebApp([]vitali.RouteRule{
    {"/image/{filename}", resources.Image{
    }},
})

type Image struct {
    vitali.Ctx
}

func (c *Image) Get() interface{} {
    f, _ := os.Open("images/"+c.PathParam("filename"))
    return f
}
```

## Method Dispatching
Implement the methods that returns anything (type interface{}) which corresponds to the HTTP methods.

Pre() is a special function that runs before the HTTP methods if implemented. You can do some initialization and checking in Pre(). Returning nil in Pre() means to continue to invoke the corresponding HTTP method.
```
type Foo struct {
    vitali.Ctx
}

func (c *Foo) Get() interface{}
func (c *Foo) Post() interface{}
func (c *Foo) Put() interface{}
func (c *Foo) Delete() interface{}
func (c *Foo) Pre() interface{}
```

## Predefined Response Types
Some typical HTTP responses are provided. See https://github.com/lunastorm/vitali/blob/master/response_types.go

For example,
```
func (c *Foo) Pre() interface{} {
    return c.BadRequest("bad!")
}

func (c *Foo) Get() interface{} {
    return c.NotFound()
}
```

## Authentication
You can provide your customized user and role provider when you implement vitali.UserProvider interface, and then setup the user provider as follows:
```
type UserProvider interface {
    AuthHeader(*http.Request) string
    GetUserAndRole(*http.Request) (string, string)
}

webapp.UserProvider = &YourUserProvider{...}
```

### Authentication Example: HTTP Basic Authentication
This example just verifies that the  username is _foo_ and the password is _bar_. And the role _AUTHED_ will be added when the user is authenticated.
```
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

func (c *UserProvider) GetUserAndRole(r *http.Request) (user string, role string) {
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
        return user, "AUTHED"
    } else {
        user = ""
        return
    }   
}
```

## vitali.Ctx
This is embedded in all your resource structs and wraps the original *http.Request and http.ResponseWriter.
```
type Ctx struct {
    Username    string
    Roles       Roles
    Request *http.Request
    ResponseWriter  http.ResponseWriter
    ChosenType  MediaType
    ChosenLang  string
    ContentType MediaType
}
```
Refer to https://github.com/lunastorm/vitali/blob/master/ctx.go for some convinient methods.

## vitali.Perm
An optional member to be embedded in the resource. 

For example,
```
type Image struct {
    vitali.Ctx
    vitali.Perm `GET:"AUTHED" DELETE:"ADMIN|OWNER"`
}
```
means that every GET request to Image should be authenticated, and every DELETE requests by roles other than _ADMIN_ or _OWNER_ are forbidden.

Actually you can use any role name you like. You can add roles to an authenticated user in your _UserProvider_ or the _Pre()_ function like this:
```
func (c *Image) Pre() interface{} {
    if /* check if c.Username is image's owner */ {
        c.Roles.Add("OWNER")
    }
}
```
