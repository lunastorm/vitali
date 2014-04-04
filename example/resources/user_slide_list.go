package resources

import (
    "os"
    "fmt"
    "encoding/json"
    "github.com/lunastorm/vitali"
)

type UserSlideList struct {
    vitali.Ctx
    vitali.Perm `POST:"OWNER" DELETE:"OWNER"`
    vitali.Provides `GET:"application/json,text/html"`
    vitali.Views `GET:"base.html,user_slide_list.html"`
}

func (c *UserSlideList) Pre() interface{} {
    if c.PathParam("user") == c.Username {
        c.Roles.Add("OWNER")
    }
    return nil
}

func (c *UserSlideList) Get() interface{} {
    dir, err := os.Open("files/" + c.PathParam("user"))
    if err != nil {panic(err)}
    defer dir.Close()

    slides, err := dir.Readdirnames(0)
    if err != nil {panic(err)}
    return slides
}

func (c *UserSlideList) Post() interface{} {
    f, err := os.Create(fmt.Sprintf("files/%s/%s", c.PathParam("user"), c.Param("slide_name")))
    if err != nil {panic(err)}
    defer f.Close()

    enc := json.NewEncoder(f)
    m := SlideModel{}
    m.InsertPage(0, "", "")
    enc.Encode(m)
    return c.SeeOther("slide/" + c.Param("slide_name"))
}
