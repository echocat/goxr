package server

import (
	"github.com/echocat/goxr"
	"github.com/valyala/fasthttp"
)

type Interceptor interface {
	OnBeforeHandle(box goxr.Box, ctx *fasthttp.RequestCtx) (handled bool, newBox goxr.Box, newCtx *fasthttp.RequestCtx)
	OnAfterHandle(box goxr.Box, ctx *fasthttp.RequestCtx)
	OnTargetPathResolved(box goxr.Box, path string, ctx *fasthttp.RequestCtx) (newPath string)
	OnHandleError(box goxr.Box, err error, interceptAllowed bool, ctx *fasthttp.RequestCtx) (handled bool, newErr error, newCtx *fasthttp.RequestCtx)
}
