package resources

import (
    "fmt"
    "strings"
    "html/template" 
)

type PageModel struct {
    Raw string `json:"raw"`
}

func (c *PageModel) HTML() (res template.HTML) {
    lines := strings.Split(c.Raw, "\n")
    var code bool
    var lineCnt int
    for _, line := range(lines) {
        if code {
            lineCnt++
            res += template.HTML(fmt.Sprintf("%d.\t%s\n", lineCnt, line))
            continue
        }
        if line == "" {continue}
        switch line[0] {
        case '!':
            res += template.HTML(fmt.Sprintf("<h1>%s</h1>\n", line[1:]))
        case '[':
            res += template.HTML(fmt.Sprintf(`<img src="/image/%s"/>`, line[1:]))
        case '*':
            res += template.HTML(fmt.Sprintf("<li>%s</li>\n", line[1:]))
        case '{':
            code = true
            res += template.HTML("<pre>")
        case '@':
            res += template.HTML(fmt.Sprintf(`<a href="%s" target=_blank><h3>%s</h3></a>`, line[1:], line[1:]))
        default:
            res += template.HTML(fmt.Sprintf("<h3>%s</h3>\n", line))
        }
    }
    if code {
        res += template.HTML("</pre>")
    }
    return
}
