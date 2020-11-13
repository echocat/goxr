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
		address := instance.Configuration.Listen.GetHttpAddress()
		instance.Log().Debug(map[string]interface{}{
			"event":   "httpListenAndServe",
			"address": address,
		})
		return s.ListenAndServe(address)
	}
}

func (instance *Server) Validate() error {
	if instance == nil || instance.Box == nil {
		return os.ErrInvalid
	}
	return instance.Configuration.ValidateAndSummarize(instance.Box)
}

func (instance *Server) Handle(ctx *fasthttp.RequestCtx) {
	ctxToUse := ctx
	boxToUse := instance.Box
	if instance.Configuration.Logging.GetAccessLog() {
		start := time.Now()
		defer func(start time.Time) {
			d := time.Now().Sub(start)
			entry := map[string]interface{}{
				"event":      "accessLog",
				"duration":   d,
				"host":       string(ctx.Host()),
				"method":     string(ctx.Method()),
				"requestUri": string(ctx.RequestURI()),
				"remote":     ctx.RemoteAddr().String(),
				"local":      ctx.LocalAddr().String(),
				"status":     ctx.Response.StatusCode(),
				"userAgent":  string(ctx.Request.Header.UserAgent()),
			}
			if handled := instance.onAccessLog(boxToUse, ctxToUse, &entry); !handled {
				instance.Log().Info(entry)
			}
		}(start)
	}
	var handled bool
	handled, boxToUse, ctxToUse = instance.onBeforeHandle(instance.Box, ctx)
	defer instance.onAfterHandle(boxToUse, ctxToUse)

	if !handled {
		path := instance.ResolveInitialTargetPath(boxToUse, ctxToUse)
		instance.WriteGenericHeaders(ctxToUse)
		instance.ServeFile(boxToUse, path, ctxToUse, true, http.StatusOK)
	}
}

func (instance *Server) ResolveInitialTargetPath(box goxr.Box, ctx *fasthttp.RequestCtx) string {
	result := string(ctx.Path())
	if result == "/" || result == "" {
		index := instance.Configuration.Paths.GetIndex()
		if index != "" {
			result = index
		}
	}
	return instance.onTargetPathResolved(box, result, ctx)
}

func (instance *Server) ServeFile(box goxr.Box, path string, ctx *fasthttp.RequestCtx, interceptAllowed bool, statusCode int) {
	if statusCode <= 0 {
		statusCode = http.StatusOK
	}
	if f, err := box.Open(path); err != nil {
		instance.HandleError(box, err, interceptAllowed, ctx)
	} else {
		success := false
		defer func() {
			if !success {
				_ = f.Close()
			}
		}()

		if fi, err := f.GetFileInfo(); err != nil {
			instance.HandleError(box, err, interceptAllowed, ctx)
		} else if fi.IsDir() {
			instance.HandleError(box, os.ErrNotExist, interceptAllowed, ctx)
		} else if !instance.DoesETagMatched(box, fi, ctx) &&
			!instance.DoesModifiedMatched(box, fi, ctx) &&
			!(interceptAllowed && instance.ShouldHandleStatusCode(box, statusCode, ctx)) {
			instance.WriteFileHeadersFor(fi, ctx)
			ctx.Response.SetStatusCode(statusCode)
			ctx.Response.SetBodyStream(f, int(fi.Size()))
			success = true
		}
	}
}

func (instance *Server) HandleError(box goxr.Box, err error, interceptAllowed bool, ctx *fasthttp.RequestCtx) {
	handled, newErr, newCtx := instance.onHandleError(box, err, interceptAllowed, ctx)
	if handled {
		return
	}
	code := instance.StatusCodeFor(newErr)

	if !interceptAllowed {
		ctx.Response.SetStatusCode(code)
		ReportNotHandableProblem(newErr, newCtx, instance.Log())
	}

	if code == http.StatusNotFound {
		if target := instance.Configuration.Paths.Catchall.GetTarget(); target != "" {
			if yes, eErr := instance.Configuration.Paths.Catchall.IsEligible(string(newCtx.Path())); eErr != nil {
				ReportNotHandableProblem(eErr, newCtx, instance.Log())
			} else if yes {
				instance.ServeFile(box, target, newCtx, false, http.StatusOK)
				return
			}
		}
	}
	if instance.ShouldHandleStatusCode(box, code, newCtx) {
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

func (instance *Server) NotModifiedFor(box goxr.Box, fi os.FileInfo, ctx *fasthttp.RequestCtx) {
	if instance.ShouldHandleStatusCode(box, http.StatusNotModified, ctx) {
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
	instance.onWriteHeadersFor(instance.Box, ctx, fi)
}

func (instance *Server) DoesETagMatched(box goxr.Box, fi os.FileInfo, ctx *fasthttp.RequestCtx) bool {
	if !instance.Configuration.Response.GetWithEtag() {
		return false
	}
	ifNonMatch := string(ctx.Request.Header.Peek("If-None-Match"))
	if efi, ok := fi.(common.ExtendedFileInfo); ok && ifNonMatch == fmt.Sprintf(`"%s"`, efi.ChecksumString()) {
		instance.NotModifiedFor(box, fi, ctx)
		return true
	}
	return false
}

func (instance *Server) DoesModifiedMatched(box goxr.Box, fi os.FileInfo, ctx *fasthttp.RequestCtx) bool {
	if !instance.Configuration.Response.GetWithLastModified() {
		return false
	}
	if ifModifiedSince := string(ctx.Request.Header.Peek("If-Modified-Since")); ifModifiedSince != "" {
		if tIfModifiedSince, err := time.Parse(time.RFC1123, ifModifiedSince); err != nil {
			// Ignored
		} else if tIfModifiedSince.Truncate(time.Second).Equal(fi.ModTime().Truncate(time.Second)) {
			instance.NotModifiedFor(box, fi, ctx)
			return true
		}
	}
	return false
}

func (instance *Server) ShouldHandleStatusCode(box goxr.Box, code int, ctx *fasthttp.RequestCtx) bool {
	if code <= 0 {
		return false
	}
	statusCodePath := instance.Configuration.Paths.FindStatusCode(code)
	if statusCodePath != "" {
		instance.ServeFile(box, statusCodePath, ctx, false, code)
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
	}
	if err := instance.configureMimeTypes(); err != nil {
		return err
	}
	if rl, ok := instance.Log().(log.RootLogger); ok {
		if err := rl.SetConfiguration(instance.Configuration.Logging.Configuration); err != nil {
			return err
		}
	}
	return nil
}

func (instance *Server) configureMimeTypes() error {
	for ext, typ := range instance.Configuration.Response.GetMimeTypes() {
		if err := mime.AddExtensionType(ext, typ); err != nil {
			return err
		}
	}
	return nil
}

func (instance *Server) onBeforeHandle(box goxr.Box, ctx *fasthttp.RequestCtx) (handled bool, newBox goxr.Box, newCtx *fasthttp.RequestCtx) {
	if i := instance.Interceptor; i != nil {
		return i.OnBeforeHandle(box, ctx)
	}
	return false, box, ctx
}

func (instance *Server) onAfterHandle(box goxr.Box, ctx *fasthttp.RequestCtx) {
	if i := instance.Interceptor; i != nil {
		i.OnAfterHandle(box, ctx)
	}
}

func (instance *Server) onTargetPathResolved(box goxr.Box, path string, ctx *fasthttp.RequestCtx) (newPath string) {
	if i := instance.Interceptor; i != nil {
		return i.OnTargetPathResolved(box, path, ctx)
	}
	return path
}

func (instance *Server) onHandleError(box goxr.Box, err error, interceptAllowed bool, ctx *fasthttp.RequestCtx) (handled bool, newErr error, newCtx *fasthttp.RequestCtx) {
	if i := instance.Interceptor; i != nil {
		return i.OnHandleError(box, err, interceptAllowed, ctx)
	}
	return false, err, ctx
}

func (instance *Server) onAccessLog(box goxr.Box, ctx *fasthttp.RequestCtx, event *map[string]interface{}) (handled bool) {
	if i := instance.Interceptor; i != nil {
		return i.OnAccessLog(box, ctx, event)
	}
	return false
}

func (instance *Server) onWriteHeadersFor(box goxr.Box, ctx *fasthttp.RequestCtx, info os.FileInfo) {
	if i := instance.Interceptor; i != nil {
		i.OnWriteHeadersFor(box, ctx, info)
	}
}
