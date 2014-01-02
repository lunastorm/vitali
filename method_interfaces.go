package vitali

type Getter interface {
    Get() interface{}
}

type Poster interface {
    Post() interface{}
}

type Putter interface {
    Put() interface{}
}

type Deleter interface {
    Delete() interface{}
}

type PreHooker interface {
    Pre() interface{}
}
