package configuration

import (
	"github.com/echocat/goxr"
	"github.com/urfave/cli"
)

type Logging struct {
	AccessLog *bool `yaml:"accessLog,omitempty"`
}

func (instance Logging) GetAccessLog() bool {
	r := instance.AccessLog
	if r == nil {
		return true
	}
	return *r
}

func (instance *Logging) Validate(goxr.Box) (errors []error) {
	return
}

func (instance Logging) Merge(with Logging) Logging {
	result := instance

	if with.AccessLog != nil {
		result.AccessLog = &(*with.AccessLog)
	}

	return result
}

func (instance *Logging) Flags() []cli.Flag {
	return []cli.Flag{}
}
