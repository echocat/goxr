package server

import (
	"github.com/echocat/goxr"
	"github.com/echocat/slf4g"
	"github.com/valyala/fasthttp"
	"net/http"
)

func BoxToHttpFileSystem(box goxr.Box) http.FileSystem {
	return &httpFS{box}
}

type httpFS struct {
	goxr.Box
}

func (instance *httpFS) Open(path string) (http.File, error) {
	return instance.Box.Open(path)
}

func BodyAllowedForStatus(status int) bool {
	switch {
	case status >= 100 && status <= 199:
		return false
	case status == 204:
		return false
	case status == 304:
		return false
	}
	return true
}

func ReportNotHandableProblem(err error, ctx *fasthttp.RequestCtx, logger log.Logger) {
	logger.
		With("remote", ctx.RemoteAddr().String()).
		With("local", ctx.LocalAddr().String()).
		With("host", string(ctx.Host())).
		With("uri", string(ctx.RequestURI())).
		With("errorType", "notHandable").
		WithError(err).
		Warn()
}
