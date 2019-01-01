package server

import (
	"github.com/echocat/goxr"
	"github.com/echocat/goxr/log"
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

func bodyAllowedForStatus(status int) bool {
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

func reportNotHandableProblem(err error, ctx *fasthttp.RequestCtx, logger log.Logger) {
	logger.
		WithField("remote", ctx.RemoteAddr().String()).
		WithField("local", ctx.LocalAddr().String()).
		WithField("host", string(ctx.Host())).
		WithField("uri", string(ctx.RequestURI())).
		WithField("errorType", "notHandable").
		WithError(err).
		Warn()
}
