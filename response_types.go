package vitali

// 302
type found struct {
    uri string
}

func (c Ctx) Found(uri string) found {
    return found{uri}
}

// 303
type seeOther struct {
    uri string
}

func (c Ctx) SeeOther(uri string) seeOther {
    return seeOther{uri}
}

// 400
type badRequest struct {
    reason string
}

func (c Ctx) BadRequest(reason string) badRequest {
    return badRequest{reason}
}

// 404
type notFound struct {
}

func (c Ctx) NotFound() notFound {
    return notFound{}
}

// 405
type methodNotAllowed struct {
    allowed []string
}

func (c Ctx) MethodNotAllowed(allowed []string) methodNotAllowed {
    return methodNotAllowed{allowed}
}

// 501
type notImplemented struct {
}

func (c Ctx) NotImplemented() notImplemented {
    return notImplemented{}
}

// return this if client is disconnected
type clientGone struct {
}

func (c Ctx) ClientGone() clientGone {
    return clientGone{}
}

// return this if something bad leads to internal server error
type internalError struct {
    where string
    why string
    code uint32
}

func (c Ctx) InternalError(e error) internalError {
    return internalError {
        where: lineInfo(1),
        why: e.Error(),
        code: errorCode(e.Error()),
    }
}
