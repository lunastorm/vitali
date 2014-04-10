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
