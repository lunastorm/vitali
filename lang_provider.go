package vitali

type LangProvider interface {
    Select(*Ctx) string
}

type EmptyLangProvider struct {
}

func (c *EmptyLangProvider) Select(ctx *Ctx) string {
    return ""
}
