package main

import (
    "log"
    "net/http"
    "html/template"
    "github.com/lunastorm/vitali"
    "github.com/lunastorm/vitali/example/util"
    "github.com/lunastorm/vitali/example/resources"
)

func main() {
    http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
    webapp := vitali.CreateWebApp([]vitali.RouteRule{
        {"/image", resources.Images{
        }},
        {"/image/{filename}", resources.Image{
        }},
        {"/progress/{slide}", resources.Progress{
            ChanMap: util.CreateChanMap(),
        }},
        {"/user/{user}/slide", resources.UserSlideList{
        }},
        {"/user/{user}/slide/{name}/{page}", resources.Slide{
        }},
    })
    webapp.UserProvider = &util.UserProvider{}
    webapp.LangProvider = &util.LangProvider{webapp.I18n}
    webapp.Settings["401_PAGE"] = "/login"
    webapp.ErrTemplate = template.Must(template.ParseFiles("views/base.html",
        "views/error.html"))
    http.Handle("/", webapp)
    log.Printf("starting server at port 8080...")
    http.ListenAndServe(":8080", nil)
}
