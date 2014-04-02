package resources

import (
    "os"
    "time"
    "regexp"
    "github.com/lunastorm/vitali"
)

type Image struct {
    vitali.Ctx
}

func (c *Image) Get() interface{} {
    matched, _ := regexp.MatchString("[0-9a-f]{40}\\.(jpg|gif|png)", c.PathParam("filename"))
    if !matched {
        return c.NotFound()
    }
    f, err := os.Open("images/"+c.PathParam("filename"))
    if err != nil {
        return c.NotFound()
    }
    c.AddHeader("Cache-Control", "public, max-age=604800")
    c.AddHeader("Expires", time.Now().Add(604800*time.Second).Format("Mon, 2 Jan 2006 15:04:05 MST"))
    return f
}
