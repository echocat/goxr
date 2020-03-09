package server

import "github.com/valyala/fasthttp"

type Interceptor interface {
	OnBeforeHandle(ctx *fasthttp.RequestCtx) (handled bool, newCtx *fasthttp.RequestCtx)
	OnAfterHandle(ctx *fasthttp.RequestCtx)
	OnTargetPathResolved(path string, ctx *fasthttp.RequestCtx) (newPath string)
	OnHandleError(err error, ctx *fasthttp.RequestCtx) (handled bool, newCtx *fasthttp.RequestCtx)
}
