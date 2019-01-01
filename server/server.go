package server

import (
	"fmt"
	"github.com/blaubaer/goxr"
	"github.com/blaubaer/goxr/common"
	"github.com/blaubaer/goxr/log"
	"github.com/blaubaer/goxr/server/configuration"
	"github.com/valyala/fasthttp"
	"mime"
	"net/http"
	"os"
	sPath "path"
	"time"
)

type Server struct {
	Box           goxr.Box
	Configuration configuration.Configuration

	Logger log.Logger
}

func (instance *Server) Run() error {
	if err := instance.configure(); err != nil {
		return err
	} else {
		s := &fasthttp.Server{
			Handler:               instance.handle,
			NoDefaultServerHeader: true,
		}
		if instance.Configuration.Response.GetGzip() {
			s.Handler = fasthttp.CompressHandler(s.Handler)
		}
		return s.ListenAndServe(instance.Configuration.Listen.GetHttpAddress())
	}
}

func (instance *Server) Validate() error {
	if instance == nil || instance.Box == nil {
		return os.ErrInvalid
	}
	return instance.Configuration.ValidateAndSummarize(instance.Box)
}

func (instance *Server) handle(ctx *fasthttp.RequestCtx) {
	if instance.Configuration.Logging.GetAccessLog() {
		start := time.Now()
		defer func(start time.Time) {
			d := time.Now().Sub(start)
			instance.Log().Info(map[string]interface{}{
				"event":    "accessLog",
				"duration": d,
				"host":     string(ctx.Host()),
				"method":   string(ctx.Method()),
				"uri":      string(ctx.RequestURI()),
				"remote":   ctx.RemoteAddr().String(),
				"local":    ctx.LocalAddr().String(),
				"status":   ctx.Response.StatusCode(),
			})
		}(start)
	}
	path := instance.resolveInitialTargetPath(ctx)
	instance.writeGenericHeaders(ctx)
	instance.serveFile(path, ctx, true)
}

func (instance *Server) resolveInitialTargetPath(ctx *fasthttp.RequestCtx) string {
	result := string(ctx.Path())
	if result == "/" || result == "" {
		index := instance.Configuration.Paths.GetIndex()
		if index != "" {
			result = index
		}
	}
	return result
}

func (instance *Server) serveFile(path string, ctx *fasthttp.RequestCtx, interceptAllowed bool) {
	if f, err := instance.Box.Open(path); err != nil {
		instance.handleError(err, ctx)
	} else {
		success := false
		defer func() {
			if !success {
				_ = f.Close()
			}
		}()

		if fi, err := f.Stat(); err != nil {
			instance.handleError(err, ctx)
		} else if fi.IsDir() {
			instance.handleError(os.ErrNotExist, ctx)
		} else if !instance.doesETagMatched(fi, ctx) &&
			!instance.doesModifiedMatched(fi, ctx) &&
			!(interceptAllowed && instance.shouldHandleStatusCode(http.StatusOK, ctx)) {
			instance.writeFileHeadersFor(fi, ctx)
			ctx.Response.SetStatusCode(http.StatusOK)
			ctx.Response.SetBodyStream(f, int(fi.Size()))
			success = true
		}
	}
}

func (instance *Server) handleError(err error, ctx *fasthttp.RequestCtx) {
	code := instance.statusCodeFor(err)
	if instance.shouldHandleStatusCode(code, ctx) {
		return
	}
	if code == http.StatusNotFound {
		if target := instance.Configuration.Paths.Catchall.GetTarget(); target != "" {
			if yes, err := instance.Configuration.Paths.Catchall.IsEligible(string(ctx.Path())); err != nil {
				reportNotHandableProblem(err, ctx, instance.Log())
			} else if yes {
				instance.serveFile(target, ctx, false)
				return
			}
		}
	}
	(JsonResponse{
		Code: code,
		Path: string(ctx.Path()),
	}).Serve(ctx, instance.Log())
}

func (instance *Server) statusCodeFor(err error) int {
	if os.IsNotExist(err) {
		return 404
	}
	if os.IsPermission(err) {
		return http.StatusForbidden
	}
	return http.StatusInternalServerError
}

func (instance *Server) notModifiedFor(fi os.FileInfo, ctx *fasthttp.RequestCtx) {
	if instance.shouldHandleStatusCode(http.StatusNotModified, ctx) {
		return
	}
	instance.writeFileHeadersFor(fi, ctx)
	ctx.SetStatusCode(http.StatusNotModified)
}

func (instance *Server) writeGenericHeaders(ctx *fasthttp.RequestCtx) {
	for name, values := range instance.Configuration.Response.GetHeaders() {
		for i, value := range values {
			if i == 0 {
				ctx.Response.Header.Set(name, value)
			} else {
				ctx.Response.Header.Add(name, value)
			}
		}
	}
}

func (instance *Server) writeFileHeadersFor(fi os.FileInfo, ctx *fasthttp.RequestCtx) {
	if efi, ok := fi.(common.ExtendedFileInfo); ok {
		ctx.Response.Header.Set("Etag", fmt.Sprintf(`"%s"`, efi.ChecksumString()))
	}
	ctx.Response.Header.Set("Last-Modified", fi.ModTime().Truncate(time.Second).UTC().Format(time.RFC1123))
	typ := mime.TypeByExtension(sPath.Ext(fi.Name()))
	if typ != "" {
		ctx.Response.Header.SetContentType(typ)
	}
}

func (instance *Server) doesETagMatched(fi os.FileInfo, ctx *fasthttp.RequestCtx) bool {
	ifNonMatch := string(ctx.Request.Header.Peek("If-None-Match"))
	if efi, ok := fi.(common.ExtendedFileInfo); ok && ifNonMatch == fmt.Sprintf(`"%s"`, efi.ChecksumString()) {
		instance.notModifiedFor(fi, ctx)
		return true
	}
	return false
}

func (instance *Server) doesModifiedMatched(fi os.FileInfo, ctx *fasthttp.RequestCtx) bool {
	ifModifiedSince := string(ctx.Request.Header.Peek("If-Modified-Since"))
	if ifModifiedSince != "" {
		if tIfModifiedSince, err := time.Parse(time.RFC1123, ifModifiedSince); err != nil {
			// Ignored
		} else if tIfModifiedSince.Truncate(time.Second).Equal(fi.ModTime().Truncate(time.Second)) {
			instance.notModifiedFor(fi, ctx)
			return true
		}
	}
	return false
}

func (instance *Server) shouldHandleStatusCode(code int, ctx *fasthttp.RequestCtx) bool {
	if code <= 0 {
		return false
	}
	statusCodePath := instance.Configuration.Paths.FindStatusCode(code)
	if statusCodePath != "" {
		instance.serveFile(statusCodePath, ctx, false)
		return true
	}
	return false
}

func (instance *Server) Log() log.Logger {
	return log.OrDefault(instance.Logger)
}

func (instance *Server) configure() error {
	if err := instance.Validate(); err != nil {
		return err
	} else if err := instance.configureMimeTypes(); err != nil {
		return err
	} else {
		return nil
	}
}

func (instance *Server) configureMimeTypes() error {
	for ext, typ := range instance.Configuration.Response.GetMimeTypes() {
		if err := mime.AddExtensionType(ext, typ); err != nil {
			return err
		}
	}
	return nil
}
