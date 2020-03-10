package configuration

import (
	"github.com/echocat/goxr"
	"github.com/echocat/goxr/log"
	"github.com/urfave/cli"
)

type Logging struct {
	log.Configuration `yaml:",inline"`
	AccessLog         *bool `yaml:"accessLog,omitempty"`
}

func (instance Logging) GetAccessLog() bool {
	r := instance.AccessLog
	if r == nil {
		return true
	}
	return *r
}

func (instance *Logging) Validate(_ goxr.Box) (errors []error) {
	return
}

func (instance Logging) Merge(with Logging) Logging {
	result := instance

	result.Configuration = result.Configuration.Merge(with.Configuration)

	if with.AccessLog != nil {
		result.AccessLog = &(*with.AccessLog)
	}

	return result
}

func (instance *Logging) Flags() []cli.Flag {
	return instance.Configuration.Flags()
}
