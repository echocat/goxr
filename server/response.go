package server

import (
	"bufio"
	"encoding/json"
	"github.com/blaubaer/goxr/log"
	"github.com/valyala/fasthttp"
	"net/http"
	"time"
)

type JsonResponse struct {
	Path    string   `json:"path,omitempty"`
	Code    int      `json:"code,omitempty"`
	Message string   `json:"message,omitempty"`
	Details string   `json:"details,omitempty"`
	Time    HttpTime `json:"time,omitempty"`
}

func (instance JsonResponse) Serve(ctx *fasthttp.RequestCtx, logger log.Logger) {
	t := instance
	if !bodyAllowedForStatus(t.Code) {
		t.Code = 200
	}
	t = t.Complete(ctx)

	ctx.Response.Header.Set("Content-Type", "application/json")
	ctx.Response.Header.Set("X-Content-Type-Options", "nosniff")
	ctx.Response.Header.SetStatusCode(t.Code)

	ctx.Response.SetBodyStreamWriter(func(w *bufio.Writer) {
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(t); err != nil {
			reportNotHandableProblem(err, ctx, logger)
		}
	})
}

func (instance JsonResponse) Complete(ctx *fasthttp.RequestCtx) JsonResponse {
	result := instance
	if result.Path == "" {
		result.Path = string(ctx.Path())
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
