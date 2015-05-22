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

// 307
type tempRedirect struct {
    uri string
}

func (c *Ctx) TempRedirect(uri string) tempRedirect {
    return tempRedirect{uri}
}

// 400
type badRequest struct {
    body interface{}
    reason string
}

func (c *Ctx) BadRequest(reason string, bodies ...interface{}) badRequest {
    return badRequest{extractBody(bodies), reason}
}

// 401
type unauthorized struct {
    body interface{}
    wwwAuthHeader string
}

func (c *Ctx) Unauthorized(wwwAuthHeader string, bodies ...interface{}) unauthorized {
    return unauthorized{extractBody(bodies), wwwAuthHeader}
}

// 403
type forbidden struct {
    body interface{}
}

func (c *Ctx) Forbidden(bodies ...interface{}) forbidden {
    return forbidden{extractBody(bodies)}
}

// 404
type notFound struct {
    body interface{}
}

func (c *Ctx) NotFound(bodies ...interface{}) notFound {
    return notFound{extractBody(bodies)}
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
    body interface{}
}

func (c *Ctx) UnsupportedMediaType(bodies ...interface{}) unsupportedMediaType {
    return unsupportedMediaType{extractBody(bodies)}
}

// 501
type notImplemented struct {
    body interface{}
}

func (c *Ctx) NotImplemented(bodies ...interface{}) notImplemented {
    return notImplemented{extractBody(bodies)}
}

// 503
type serviceUnavailable struct {
    body interface{}
    seconds int
}

func (c *Ctx) ServiceUnavailable(seconds int, bodies ...interface{}) serviceUnavailable {
    return serviceUnavailable{extractBody(bodies), seconds}
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

func extractBody(bodies []interface{}) interface{}{
    var body interface{}
    if len(bodies) > 0 {
        // Only uses first element as body
        body = bodies[0]
    }
    return body
}
