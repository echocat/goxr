package server

import (
	"fmt"
	"github.com/echocat/goxr"
	"github.com/echocat/goxr/common"
	"github.com/echocat/goxr/log"
	"github.com/echocat/goxr/server/configuration"
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

	Logger      log.Logger
	Interceptor Interceptor
}

func (instance *Server) Run() error {
	if err := instance.configure(); err != nil {
		return err
	} else {
		s := &fasthttp.Server{
			Handler:               instance.Handle,
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

func (instance *Server) Handle(ctx *fasthttp.RequestCtx) {
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
	handled, newCtx := instance.onBeforeHandle(ctx)
	defer instance.onAfterHandle(newCtx)

	if !handled {
		path := instance.ResolveInitialTargetPath(newCtx)
		instance.WriteGenericHeaders(newCtx)
		instance.ServeFile(path, newCtx, true, http.StatusOK)
	}
}

func (instance *Server) ResolveInitialTargetPath(ctx *fasthttp.RequestCtx) string {
	result := string(ctx.Path())
	if result == "/" || result == "" {
		index := instance.Configuration.Paths.GetIndex()
		if index != "" {
			result = index
		}
	}
	return result
}

func (instance *Server) ServeFile(path string, ctx *fasthttp.RequestCtx, interceptAllowed bool, statusCode int) {
	if statusCode <= 0 {
		statusCode = http.StatusOK
	}
	if f, err := instance.Box.Open(path); err != nil {
		instance.HandleError(err, ctx)
	} else {
		success := false
		defer func() {
			if !success {
				_ = f.Close()
			}
		}()

		if fi, err := f.Stat(); err != nil {
			instance.HandleError(err, ctx)
		} else if fi.IsDir() {
			instance.HandleError(os.ErrNotExist, ctx)
		} else if !instance.DoesETagMatched(fi, ctx) &&
			!instance.DoesModifiedMatched(fi, ctx) &&
			!(interceptAllowed && instance.ShouldHandleStatusCode(statusCode, ctx)) {
			instance.WriteFileHeadersFor(fi, ctx)
			ctx.Response.SetStatusCode(statusCode)
			ctx.Response.SetBodyStream(f, int(fi.Size()))
			success = true
		}
	}
}

func (instance *Server) HandleError(err error, ctx *fasthttp.RequestCtx) {
	handled, newCtx := instance.onHandleError(err, ctx)
	if handled {
		return
	}
	code := instance.StatusCodeFor(err)
	if code == http.StatusNotFound {
		if target := instance.Configuration.Paths.Catchall.GetTarget(); target != "" {
			if yes, err := instance.Configuration.Paths.Catchall.IsEligible(string(newCtx.Path())); err != nil {
				ReportNotHandableProblem(err, newCtx, instance.Log())
			} else if yes {
				instance.ServeFile(target, newCtx, false, http.StatusOK)
				return
			}
		}
	}
	if instance.ShouldHandleStatusCode(code, newCtx) {
		return
	}
	(JsonResponse{
		Code: code,
		Path: string(newCtx.Path()),
	}).Serve(newCtx, instance.Log())
}

func (instance *Server) StatusCodeFor(err error) int {
	if os.IsNotExist(err) {
		return 404
	}
	if os.IsPermission(err) {
		return http.StatusForbidden
	}
	return http.StatusInternalServerError
}

func (instance *Server) NotModifiedFor(fi os.FileInfo, ctx *fasthttp.RequestCtx) {
	if instance.ShouldHandleStatusCode(http.StatusNotModified, ctx) {
		return
	}
	instance.WriteFileHeadersFor(fi, ctx)
	ctx.SetStatusCode(http.StatusNotModified)
}

func (instance *Server) WriteGenericHeaders(ctx *fasthttp.RequestCtx) {
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

func (instance *Server) WriteFileHeadersFor(fi os.FileInfo, ctx *fasthttp.RequestCtx) {
	if efi, ok := fi.(common.ExtendedFileInfo); ok && instance.Configuration.Response.GetWithEtag() {
		ctx.Response.Header.Set("Etag", fmt.Sprintf(`"%s"`, efi.ChecksumString()))
	}
	if instance.Configuration.Response.GetWithLastModified() {
		ctx.Response.Header.Set("Last-Modified", fi.ModTime().Truncate(time.Second).UTC().Format(time.RFC1123))
	}
	if typ := mime.TypeByExtension(sPath.Ext(fi.Name())); typ != "" && instance.Configuration.Response.GetWithContentType() {
		ctx.Response.Header.SetContentType(typ)
	}
}

func (instance *Server) DoesETagMatched(fi os.FileInfo, ctx *fasthttp.RequestCtx) bool {
	if !instance.Configuration.Response.GetWithEtag() {
		return false
	}
	ifNonMatch := string(ctx.Request.Header.Peek("If-None-Match"))
	if efi, ok := fi.(common.ExtendedFileInfo); ok && ifNonMatch == fmt.Sprintf(`"%s"`, efi.ChecksumString()) {
		instance.NotModifiedFor(fi, ctx)
		return true
	}
	return false
}

func (instance *Server) DoesModifiedMatched(fi os.FileInfo, ctx *fasthttp.RequestCtx) bool {
	if !instance.Configuration.Response.GetWithLastModified() {
		return false
	}
	if ifModifiedSince := string(ctx.Request.Header.Peek("If-Modified-Since")); ifModifiedSince != "" {
		if tIfModifiedSince, err := time.Parse(time.RFC1123, ifModifiedSince); err != nil {
			// Ignored
		} else if tIfModifiedSince.Truncate(time.Second).Equal(fi.ModTime().Truncate(time.Second)) {
			instance.NotModifiedFor(fi, ctx)
			return true
		}
	}
	return false
}

func (instance *Server) ShouldHandleStatusCode(code int, ctx *fasthttp.RequestCtx) bool {
	if code <= 0 {
		return false
	}
	statusCodePath := instance.Configuration.Paths.FindStatusCode(code)
	if statusCodePath != "" {
		instance.ServeFile(statusCodePath, ctx, false, code)
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

func (instance *Server) onBeforeHandle(ctx *fasthttp.RequestCtx) (handled bool, newCtx *fasthttp.RequestCtx) {
	if i := instance.Interceptor; i != nil {
		return i.OnBeforeHandle(ctx)
	}
	return false, ctx
}

func (instance *Server) onAfterHandle(ctx *fasthttp.RequestCtx) {
	if i := instance.Interceptor; i != nil {
		i.OnAfterHandle(ctx)
	}
}

func (instance *Server) onHandleError(err error, ctx *fasthttp.RequestCtx) (handled bool, newCtx *fasthttp.RequestCtx) {
	if i := instance.Interceptor; i != nil {
		return i.OnHandleError(err, ctx)
	}
	return false, ctx
}
