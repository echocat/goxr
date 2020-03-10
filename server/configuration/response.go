package configuration

import (
	"github.com/echocat/goxr"
	"github.com/urfave/cli"
)

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

func (instance Response) cloneMimeTypes() map[string]string {
	r := map[string]string{}
	for k, v := range instance.GetMimeTypes() {
		r[k] = v
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

func (instance Response) cloneHeaders() map[string][]string {
	r := map[string][]string{}
	for k, vs := range instance.GetHeaders() {
		r[k] = make([]string, len(vs))
		for i, v := range vs {
			r[k][i] = v
		}
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

func (instance Response) Merge(with Response) Response {
	result := instance

	if with.MimeTypes != nil {
		result.MimeTypes = with.cloneMimeTypes()
	}
	if with.Gzip != nil {
		result.Gzip = &(*with.Gzip)
	}
	if with.Headers != nil {
		result.Headers = with.cloneHeaders()
	}
	if with.WithEtag != nil {
		result.WithEtag = &(*with.WithEtag)
	}
	if with.WithLastModified != nil {
		result.WithLastModified = &(*with.WithLastModified)
	}
	if with.WithContentType != nil {
		result.WithContentType = &(*with.WithContentType)
	}

	return result
}

func (instance *Response) Flags() []cli.Flag {
	return []cli.Flag{}
}
