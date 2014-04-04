package resources

import (
    "os"
    "fmt"
    "time"
    "strconv"
    "net/http"
    "encoding/json"
    "github.com/lunastorm/vitali"
)

type SlideModel struct {
    Pages []PageModel `json:"pages"`
}

func (c *SlideModel) InsertPage(idx int, raw string) {
    c.Pages = append(c.Pages[:idx], append([]PageModel{PageModel{raw}}, c.Pages[idx:]...)...)
}

func (c *SlideModel) RemovePage(idx int) {
    c.Pages = append(c.Pages[:idx], c.Pages[idx+1:]...)
}

type Slide struct {
    vitali.Ctx
    vitali.Perm `GET:"AUTHED" *:"OWNER"`
    vitali.Provides `GET:"application/json,text/html"`
    vitali.Views `GET:"base.html,slide.html"`
    Page uint64
}

func (c *Slide) Pre() interface{} {
    var err error
    if c.Page, err = strconv.ParseUint(c.PathParam("page"), 10, 32); err != nil {
        page := "1"
        if c.Cookie("page") != "" {
            _, err := strconv.ParseUint(c.Cookie("page"), 10, 32)
            if err == nil {
                page = c.Cookie("page")
            }
        }
        return c.SeeOther(fmt.Sprintf("/user/%s/slide/%s/%s",
            c.PathParam("user"), c.PathParam("name"), page))
    }

    _, err = os.Open(fmt.Sprintf("files/%s/%s",
        c.PathParam("user"), c.PathParam("name")))
    if err != nil {
        return c.NotFound()
    }

    if c.PathParam("user") == c.Username {
        c.Roles.Add("OWNER")
    }
    return nil
}

func (c *Slide) getSlide() (m SlideModel) {
    f, err := os.Open(fmt.Sprintf("files/%s/%s",
        c.PathParam("user"), c.PathParam("name")))
    if err != nil {panic(err)}
    defer f.Close()

    dec := json.NewDecoder(f)
    err = dec.Decode(&m)
    if err != nil {panic(err)}
    return
}

func (c *Slide) Get() interface{} {
    slide := c.getSlide()
    if int(c.Page) > len(slide.Pages) {
        return c.NotFound()
    }

    c.SetCookie(&http.Cookie{
        Name: "page",
        Value: c.PathParam("page"),
        Path: fmt.Sprintf("/user/%s/slide/%s", c.PathParam("user"), c.PathParam("name")),
        Expires: time.Now().Add(30*24*time.Hour),
    })

    return struct{
        Page *PageModel
        TotalPages int
    }{
        &slide.Pages[c.Page-1],
        len(slide.Pages),
    }
}

func (c *Slide) saveSlide(s *SlideModel) {
    f, err := os.OpenFile(fmt.Sprintf("files/%s/%s",
        c.PathParam("user"), c.PathParam("name")), os.O_WRONLY, 0666)
    if err != nil {panic(err)}
    defer f.Close()

    b, err := json.Marshal(s)
    if err != nil {panic(err)}
    _, err = f.Write(b)
    if err != nil {panic(err)}
}

func (c *Slide) Post() interface{} {
    slide := c.getSlide()
    if c.Param("create") == "create" || c.Param("create") == "dup" {
        if c.Param("create") == "create" {
            slide.InsertPage(int(c.Page), "")
        } else {
            slide.InsertPage(int(c.Page), slide.Pages[int(c.Page-1)].Raw)
        }
        c.saveSlide(&slide)
        c.SetCookie(&http.Cookie{
            Name: "create",
            Value: "create",
            Path: fmt.Sprintf("/user/%s/slide/%s/%d", c.PathParam("user"), c.PathParam("name"), c.Page+1),
            Expires: time.Now().Add(30*24*time.Hour),
        })
        return c.SeeOther(fmt.Sprintf("%d", c.Page+1))
    }

    if slide.Pages[c.Page-1].Raw == c.Param("raw") {
        return c.SeeOther(fmt.Sprintf("%d", c.Page))
    }
    slide.Pages[c.Page-1].Raw = c.Param("raw")
    c.saveSlide(&slide)
    return c.SeeOther(fmt.Sprintf("%d", c.Page))
}

func (c *Slide) Delete() interface{} {
    slide := c.getSlide()
    slide.RemovePage(int(c.Page-1))
    c.saveSlide(&slide)
    return c.NoContent()
}
