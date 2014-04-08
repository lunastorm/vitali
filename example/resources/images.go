package resources

import (
    "io"
    "os"
    "fmt"
    "net/http"
    "image/gif"
    "image/png"
    "image/jpeg"
    "io/ioutil"
    "crypto/sha1"
    "github.com/lunastorm/vitali"
)

type Images struct {
    vitali.Ctx
    vitali.Perm `*:"AUTHED"`
    vitali.Provides `GET:"text/html" POST:"text/html"`
    vitali.Consumes `POST:"multipart/form-data,application/x-www-form-urlencoded"`
    vitali.Views `GET:"base.html,images.html" POST:"base.html,images.html"`
}

func (c *Images) Get() interface{} {
    return ""
}

func (c *Images) Post() interface{} {
    var f io.ReadCloser
    var err error
    if c.Param("url") != "" {
        res, err := http.Get(c.Param("url"))
        if err != nil {
            return c.BadRequest("cannot get image")
        }
        f = res.Body
    } else {
        f, _, err = c.Request.FormFile("image")
        if err != nil {panic(err)}
    }
    defer f.Close()

    tmpf, err := ioutil.TempFile("", "vitali-example-img-")
    if err != nil {panic(err)}
    defer tmpf.Close()
    defer os.Remove(tmpf.Name())

    h := sha1.New()
    mWriter := io.MultiWriter(tmpf, h)
    _, err = io.Copy(mWriter, f)
    if err != nil {panic(err)}

    sha1 := fmt.Sprintf("%x", h.Sum(nil))

    var imageType string
    tmpf.Seek(0, 0)
    _, err = jpeg.Decode(tmpf)
    if err == nil {
        imageType = "jpg"
    } else {
        tmpf.Seek(0, 0)
        _, err = png.Decode(tmpf)
        if err == nil {
            imageType = "png"
        } else {
            tmpf.Seek(0, 0)
            _, err = gif.Decode(tmpf)
            if err == nil {
                imageType = "gif"
            }
        }
    }
    if imageType == "" {
        return c.BadRequest("unsupported image type")
    }
    err = os.Rename(tmpf.Name(), fmt.Sprintf("images/%s.%s", sha1, imageType))
    if err != nil {panic(err)}

    return fmt.Sprintf("%s.%s", sha1, imageType)
}
