package configuration

import "github.com/echocat/goxr"

type Response struct {
	MimeTypes        map[string]string   `yaml:"mimeTypes,omitempty"`
	Gzip             *bool               `yaml:"gzip,omitempty"`
	Headers          map[string][]string `yaml:"headers,omitempty"`
	WithEtag         *bool               `yaml:"withEtag,omitempty"`
	WithLastModified *bool               `yaml:"withLastModified,omitempty"`
	WithContentType  *bool               `yaml:"withContentType,omitempty"`
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

func (instance Response) GetWithEtag() bool {
	r := instance.WithEtag
	if r == nil {
		return true
	}
	return *r
}

func (instance Response) GetWithLastModified() bool {
	r := instance.WithLastModified
	if r == nil {
		return true
	}
	return *r
}

func (instance Response) GetWithContentType() bool {
	r := instance.WithContentType
	if r == nil {
		return true
	}
	return *r
}

func (instance *Response) Validate(using goxr.Box) (errors []error) {
	return
}
