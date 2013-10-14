package vitali

import (
    "io/ioutil"
    "html/template"
)

func Template(name string, filename string) (t *template.Template) {
    content, err := ioutil.ReadFile(filename)
    if err != nil {
        panic(err)
    }
    return template.Must(template.New(name).Parse(string(content)))
}

type view struct {
    template *template.Template
    model interface{}
}

func View(template *template.Template, model interface{}) view {
    return view{template, model}
}
