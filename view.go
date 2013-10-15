package vitali

import (
    "html/template"
)

type view struct {
    template *template.Template
    model interface{}
}

func (c Ctx) View(template *template.Template, model interface{}) view {
    return view{template, model}
}
