package server

import (
	"github.com/blaubaer/goxr/log"
	"net/http"
)

type Handler interface {
	TryServeHTTP(resp http.ResponseWriter, req *http.Request) error
}

type DoServeHttpTarget interface {
	Handler
	log.HasLogger
}

func DoServeHTTP(resp http.ResponseWriter, req *http.Request, target DoServeHttpTarget) {
	if err := target.TryServeHTTP(resp, req); err != nil {
		ErrorToHttpResponse(err).ServeOrWarnUsing(resp, req, target)
	}
}
