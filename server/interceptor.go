package server

import (
	"github.com/echocat/goxr"
	"github.com/echocat/goxr/common"
	"github.com/valyala/fasthttp"
)

type Interceptor interface {
	OnBeforeHandle(box goxr.Box, ctx *fasthttp.RequestCtx) (handled bool, newBox goxr.Box, newCtx *fasthttp.RequestCtx)
	OnAfterHandle(box goxr.Box, ctx *fasthttp.RequestCtx)
	OnTargetPathResolved(box goxr.Box, path string, ctx *fasthttp.RequestCtx) (newPath string)
	OnHandleError(box goxr.Box, err error, interceptAllowed bool, ctx *fasthttp.RequestCtx) (handled bool, newErr error, newCtx *fasthttp.RequestCtx)
	OnAccessLog(box goxr.Box, ctx *fasthttp.RequestCtx, event *map[string]interface{}) (handled bool)
	OnWriteHeadersFor(box goxr.Box, ctx *fasthttp.RequestCtx, fi common.FileInfo)
}
