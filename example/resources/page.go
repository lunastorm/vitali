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
    for _, line := range(lines) {
        if line == "" {continue}
        switch line[0] {
        case '!':
            res += template.HTML(fmt.Sprintf("<h1>%s</h1>\n", line[1:]))
        case '[':
            res += template.HTML(fmt.Sprintf(`<img src="/image/%s"/>`, line[1:]))
        default:
            res += template.HTML(fmt.Sprintf("<h3>%s</h3>\n", line))
        }
    }
    return
}
