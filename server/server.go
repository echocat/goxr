package server

import (
	"github.com/blaubaer/goxr"
	"github.com/blaubaer/goxr/log"
	"github.com/blaubaer/goxr/server/configuration"
	"mime"
	"net/http"
	"os"
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
		handler := instance.NewHandler()
		return http.ListenAndServe(instance.Configuration.Listen.GetHttpAddress(), handler)
	}
}

func (instance *Server) Validate() error {
	if instance == nil || instance.Box == nil {
		return os.ErrInvalid
	}
	return instance.Configuration.ValidateAndSummarize(instance.Box)
}

func (instance *Server) NewHandler() *BoxHandler {
	handler := NewBoxHandler(instance.Box, instance.Configuration)
	handler.Logger = instance.Log()
	return handler
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
