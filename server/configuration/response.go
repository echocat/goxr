package configuration

import "github.com/blaubaer/goxr"

type Response struct {
	MimeTypes map[string]string   `yaml:"mimeTypes,omitempty"`
	Gzip      *bool               `yaml:"gzip,omitempty"`
	Headers   map[string][]string `yaml:"headers,omitempty"`
}

func (instance Response) GetMimeTypes() map[string]string {
	r := instance.MimeTypes
	if r == nil {
		return make(map[string]string)
	}
	return r
}

func (instance Response) GetGzip() bool {
	r := instance.Gzip
	if r == nil {
		return false
	}
	return *r
}

func (instance Response) GetHeaders() map[string][]string {
	r := instance.Headers
	if r == nil {
		return make(map[string][]string)
	}
	return r
}

func (instance *Response) Validate(using goxr.Box) (errors []error) {
	return
}
