package server

import (
	"encoding/json"
	"github.com/blaubaer/goxr/log"
	"net/http"
	"time"
)

type HttpResponse struct {
	Path    string   `json:"path,omitempty"`
	Code    int      `json:"code,omitempty"`
	Message string   `json:"message,omitempty"`
	Details string   `json:"details,omitempty"`
	Time    HttpTime `json:"time,omitempty"`
}

func (instance HttpResponse) Serve(resp http.ResponseWriter, req *http.Request) error {
	t := instance
	if !bodyAllowedForStatus(t.Code) {
		t.Code = 200
	}
	t = t.Complete(req)

	resp.Header().Set("Content-Type", "application/json")
	resp.Header().Set("X-Content-Type-Options", "nosniff")
	resp.WriteHeader(t.Code)

	encoder := json.NewEncoder(resp)
	encoder.SetIndent("", "  ")
	return encoder.Encode(t)
}

func (instance HttpResponse) ServeOrWarn(resp http.ResponseWriter, req *http.Request, logger log.Logger) {
	if err := instance.Serve(resp, req); err != nil {
		var path string
		if req != nil && req.URL != nil {
			path = req.URL.Path
		}
		logger.
			WithError(err).
			WithField("path", path).
			WithField("code", instance.Code).
			Warnf("Could not write response to client.")
	}
}

func (instance HttpResponse) ServeOrWarnUsing(resp http.ResponseWriter, req *http.Request, hasLogger log.HasLogger) {
	instance.ServeOrWarn(resp, req, hasLogger.Log())
}

func (instance HttpResponse) Complete(req *http.Request) HttpResponse {
	result := instance
	if result.Path == "" && req != nil && req.URL != nil {
		result.Path = req.URL.Path
	}
	if result.Message == "" {
		result.Message = http.StatusText(result.Code)
	}
	if result.Time == noHttpTime {
		result.Time = HttpTime(time.Now().Truncate(time.Millisecond))
	}
	return result
}

type HttpTime time.Time

func (instance HttpTime) MarshalText() (text []byte, err error) {
	return []byte(time.Time(instance).Format(time.RFC3339)), nil
}

func (instance *HttpTime) UnmarshalText(text []byte) error {
	if t, err := time.Parse(time.RFC3339, string(text)); err != nil {
		return err
	} else {
		*instance = HttpTime(t)
		return nil
	}
}