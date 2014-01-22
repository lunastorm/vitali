package vitali

//204
type  noContent struct {
}

func (c *Ctx) NoContent() noContent {
    return noContent{}
}

//301
type movedPermanently struct {
    uri string
}

func (c *Ctx) MovedPermanently(uri string) movedPermanently {
    return movedPermanently{uri}
}

// 302
type found struct {
    uri string
}

func (c *Ctx) Found(uri string) found {
    return found{uri}
}

// 303
type seeOther struct {
    uri string
}

func (c *Ctx) SeeOther(uri string) seeOther {
    return seeOther{uri}
}

// 400
type badRequest struct {
    reason string
}

func (c *Ctx) BadRequest(reason string) badRequest {
    return badRequest{reason}
}

// 401
type unauthorized struct {
    wwwAuthHeader string
}

func (c *Ctx) Unauthorized(wwwAuthHeader string) unauthorized {
    return unauthorized{wwwAuthHeader}
}

// 403
type forbidden struct {
}

func (c *Ctx) Forbidden() forbidden {
    return forbidden{}
}

// 404
type notFound struct {
}

func (c *Ctx) NotFound() notFound {
    return notFound{}
}

// 405
type methodNotAllowed struct {
    allowed []string
}

func (c *Ctx) MethodNotAllowed(allowed []string) methodNotAllowed {
    return methodNotAllowed{allowed}
}

//406
type notAcceptable struct {
    provided MediaTypes
}

func (c *Ctx) NotAcceptable(provided MediaTypes) notAcceptable {
    return notAcceptable{provided}
}

// 415
type unsupportedMediaType struct {
}

func (c *Ctx) UnsupportedMediaType() unsupportedMediaType {
    return unsupportedMediaType{}
}

// 501
type notImplemented struct {
}

func (c *Ctx) NotImplemented() notImplemented {
    return notImplemented{}
}

// 503
type serviceUnavailable struct {
    seconds int
}

func (c *Ctx) ServiceUnavailable(seconds int) serviceUnavailable {
    return serviceUnavailable{seconds}
}

// return this if client is disconnected
type clientGone struct {
}

func (c *Ctx) ClientGone() clientGone {
    return clientGone{}
}

// return this if something bad leads to internal server error
type internalError struct {
    where string
    why string
    code uint32
}

func (c *Ctx) InternalError(e error) internalError {
    return internalError {
        where: lineInfo(1),
        why: e.Error(),
        code: errorCode(e.Error()),
    }
}
