package configuration

import (
	"github.com/echocat/goxr"
	"github.com/echocat/goxr/log"
)

type Logging struct {
	AccessLog *bool         `yaml:"accessLog,omitempty"`
	Level     LoggingLevel  `yaml:"level,omitempty"`
	Format    LoggingFormat `yaml:"format,omitempty"`
}

func (instance Logging) GetAccessLog() bool {
	r := instance.AccessLog
	if r == nil {
		return true
	}
	return *r
}

func (instance Logging) GetLevel() log.Level {
	r := instance.Level
	if r.v == nil {
		return log.GetLevel()
	}
	return r.v
}

func (instance Logging) GetFormat() log.Format {
	r := instance.Format
	if r.v == nil {
		return log.GetFormat()
	}
	return r.v
}

func (instance *Logging) Validate(using goxr.Box) (errors []error) {
	return
}

type LoggingLevel struct {
	v log.Level
}

func (instance *LoggingLevel) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var plain string
	if err := unmarshal(&plain); err != nil {
		return err
	}
	level := log.GetLevel()
	if err := level.Set(plain); err != nil {
		return err
	}
	instance.v = level
	return nil
}

func (instance *LoggingLevel) MarshalYAML() (interface{}, error) {
	level := instance.v
	if level == nil {
		return nil, nil
	}
	return level.String(), nil
}

type LoggingFormat struct {
	v log.Format
}

func (instance *LoggingFormat) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var plain string
	if err := unmarshal(&plain); err != nil {
		return err
	}
	format := log.GetFormat()
	if err := format.Set(plain); err != nil {
		return err
	}
	instance.v = format
	return nil
}

func (instance *LoggingFormat) MarshalYAML() (interface{}, error) {
	level := instance.v
	if level == nil {
		return nil, nil
	}
	return level.String(), nil
}
