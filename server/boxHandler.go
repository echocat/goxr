package server

import (
	"github.com/blaubaer/goxr"
	"github.com/blaubaer/goxr/log"
	"github.com/blaubaer/goxr/server/configuration"
	"io"
	"net/http"
	"os"
	"time"
)

func NewBoxHandler(box goxr.Box, config configuration.Configuration) *BoxHandler {
	return &BoxHandler{
		Delegate:      NewHttpFileHandler(box),
		Box:           box,
		Configuration: config,
		Logger:        log.DefaultLogger,
	}
}

type BoxHandler struct {
	Box           goxr.Box
	Delegate      Handler
	Configuration configuration.Configuration

	Logger log.Logger
}

func (instance *BoxHandler) TryServeHTTP(resp http.ResponseWriter, req *http.Request) error {
	instance.SetHeaders(resp)
	hw := &HandlerResponse{
		Box:           instance.Box,
		Delegate:      resp,
		Request:       req,
		Configuration: instance.Configuration,
		Logger:        instance.Log(),
	}

	if req.URL.Path == "/" || req.URL.Path == "" {
		index := instance.Configuration.Paths.GetIndex()
		if index != "" {
			req.URL.Path = index
		}
	}

	if err := instance.tryServeHttpWithPreChecks(hw, req); os.IsNotExist(err) && !hw.WasSomethingWritten() {
		if catchall := instance.Configuration.Paths.Catchall.GetTarget(); catchall != "" {
			if allowed, err := instance.Configuration.Paths.Catchall.IsEligible(req.URL.Path); err != nil {
				return err
			} else if !allowed {
				return err
			} else {
				req.URL.Path = catchall
				return instance.Delegate.TryServeHTTP(hw, req)
			}
		} else {
			return err
		}
	} else if err != nil {
		return err
	} else {
		return nil
	}
}

func (instance *BoxHandler) tryServeHttpWithPreChecks(resp http.ResponseWriter, req *http.Request) error {
	if req == nil || req.URL == nil {
		return os.ErrNotExist
	} else if allowed, err := instance.Configuration.Paths.PathAllowed(req.URL.Path); err != nil {
		return err
	} else if !allowed {
		return os.ErrNotExist
	} else {
		return instance.Delegate.TryServeHTTP(resp, req)
	}
}

func (instance *BoxHandler) SetHeaders(resp http.ResponseWriter) {
	for name, values := range instance.Configuration.Response.GetHeaders() {
		resp.Header()[http.CanonicalHeaderKey(name)] = values
	}
}

func (instance *BoxHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	if req == nil || req.URL == nil {
		return
	}

	rResp := &recordStatusCodeResponse{
		delegate: resp,
	}

	if instance.Configuration.Logging.GetAccessLog() {
		start := time.Now()
		defer func(start time.Time) {
			d := time.Now().Sub(start)
			instance.Log().Info(map[string]interface{}{
				"event":    "accessLog",
				"duration": d,
				"host":     req.Host,
				"method":   req.Method,
				"uri":      req.RequestURI,
				"remote":   req.RemoteAddr,
				"status":   rResp.statusCode,
			})
		}(start)
	}

	DoServeHTTP(rResp, req, instance)
}

func (instance *BoxHandler) Log() log.Logger {
	return log.OrDefault(instance.Logger)
}

type HandlerResponse struct {
	Box           goxr.Box
	Delegate      http.ResponseWriter
	Request       *http.Request
	Configuration configuration.Configuration

	Logger log.Logger

	somethingWasWritten   bool
	statusCodeWritten     *int
	ignoreWriteStatements bool
}

func (instance *HandlerResponse) Header() http.Header {
	return instance.Delegate.Header()
}

func (instance *HandlerResponse) Write(b []byte) (int, error) {
	if instance.statusCodeWritten == nil {
		instance.WriteHeader(200)
	}
	if instance.ignoreWriteStatements {
		// Pretend as we accept the bytes but we ignore them.
		return len(b), nil
	}
	instance.somethingWasWritten = true
	return instance.Delegate.Write(b)
}

func (instance *HandlerResponse) WasSomethingWritten() bool {
	return instance.somethingWasWritten
}

func (instance *HandlerResponse) WriteHeader(statusCode int) {
	instance.statusCodeWritten = &statusCode
	instance.somethingWasWritten = true

	if statusCodePath, err := instance.Configuration.Paths.FindStatusCode(statusCode); err != nil {
		instance.Log().
			WithField("statusCode", statusCode).
			WithField("path", instance.Request.URL.Path).
			WithField("err", instance.Request.URL.Path).
			Warnf("Cannot evaluate statusCodePage - ignoring and use default behavior.")
	} else if statusCodePath != "" {
		if f, err := instance.Box.Open(statusCodePath); err != nil {
			instance.Log().
				WithField("statusCode", statusCode).
				WithField("statusCodePath", statusCodePath).
				WithField("path", instance.Request.URL.Path).
				WithField("err", instance.Request.URL.Path).
				Warnf("StatusCodePage '%s' could not be opened - ignoring and use default behavior.", statusCodePath)
		} else {
			instance.ignoreWriteStatements = true
			//noinspection GoUnhandledErrorResult
			defer f.Close()
			if _, err := io.Copy(instance.Delegate, f); err != nil {
				instance.Log().
					WithField("statusCode", statusCode).
					WithField("statusCodePath", statusCodePath).
					WithField("path", instance.Request.URL.Path).
					WithError(err).
					Errorf("StatusCodePage '%s' could not be written to response.", statusCodePath)
			}
			return
		}
	}

	instance.Delegate.WriteHeader(statusCode)
}

func (instance *HandlerResponse) Log() log.Logger {
	return log.OrDefault(instance.Logger)
}

type recordStatusCodeResponse struct {
	delegate   http.ResponseWriter
	statusCode int
}

func (instance *recordStatusCodeResponse) Header() http.Header {
	return instance.delegate.Header()
}

func (instance *recordStatusCodeResponse) Write(b []byte) (int, error) {
	if instance.statusCode == 0 {
		instance.WriteHeader(200)
	}
	return instance.delegate.Write(b)
}

func (instance *recordStatusCodeResponse) WriteHeader(statusCode int) {
	instance.statusCode = statusCode
	instance.delegate.WriteHeader(statusCode)
}
