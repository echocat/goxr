package log

import (
	"bytes"
	"encoding/json"
	"github.com/sirupsen/logrus"
)

var Default = NewDefault()

func NewDefault() RootLogger {
	instance := &LogrusLogger{
		Delegate: logrus.New(),
	}
	if err := instance.SetConfiguration(Configuration{}); err != nil {
		panic(err)
	}
	return instance
}

type Logger interface {
	WithField(key string, value interface{}) Logger
	WithDeepField(key string, value interface{}) Logger
	WithDeepFieldOn(key string, value interface{}, on func() bool) Logger
	WithError(err error) Logger

	Trace(what ...interface{})
	Debug(what ...interface{})
	Info(what ...interface{})
	Warn(what ...interface{})
	Error(what ...interface{})
	Fatal(what ...interface{})

	Tracef(msg string, args ...interface{})
	Debugf(msg string, args ...interface{})
	Infof(msg string, args ...interface{})
	Warnf(msg string, args ...interface{})
	Errorf(msg string, args ...interface{})
	Fatalf(msg string, args ...interface{})

	IsTraceEnabled() bool
	IsDebugEnabled() bool
	IsInfoEnabled() bool
	IsWarnEnabled() bool
	IsErrorEnabled() bool
	IsFatalEnabled() bool
}

type RootLogger interface {
	Logger

	SetConfiguration(Configuration) error
	GetConfiguration() Configuration
}

func WithField(key string, value interface{}) Logger {
	return Default.WithField(key, value)
}

func WithDeepField(key string, value interface{}) Logger {
	return Default.WithDeepField(key, value)
}

func WithDeepFieldOn(key string, value interface{}, on func() bool) Logger {
	return Default.WithDeepFieldOn(key, value, on)
}

func WithError(err error) Logger {
	return Default.WithError(err)
}

func Trace(what ...interface{}) {
	Default.Trace(what...)
}

func Debug(what ...interface{}) {
	Default.Debug(what...)
}

func Info(what ...interface{}) {
	Default.Info(what...)
}

func Warn(what ...interface{}) {
	Default.Warn(what...)
}

func Error(what ...interface{}) {
	Default.Error(what...)
}

func Fatal(what ...interface{}) {
	Default.Fatal(what...)
}

func Tracef(msg string, args ...interface{}) {
	Default.Tracef(msg, args...)
}

func Debugf(msg string, args ...interface{}) {
	Default.Debugf(msg, args...)
}

func Infof(msg string, args ...interface{}) {
	Default.Infof(msg, args...)
}

func Warnf(msg string, args ...interface{}) {
	Default.Warnf(msg, args...)
}

func Errorf(msg string, args ...interface{}) {
	Default.Errorf(msg, args...)
}

func Fatalf(msg string, args ...interface{}) {
	Default.Fatalf(msg, args...)
}

func IsTraceEnabled() bool {
	return Default.IsTraceEnabled()
}

func IsDebugEnabled() bool {
	return Default.IsDebugEnabled()
}

func IsInfoEnabled() bool {
	return Default.IsInfoEnabled()
}

func IsWarnEnabled() bool {
	return Default.IsWarnEnabled()
}

func IsErrorEnabled() bool {
	return Default.IsErrorEnabled()
}

func IsFatalEnabled() bool {
	return Default.IsFatalEnabled()
}

type JsonValue struct {
	Value       interface{}
	PrettyPrint bool
}

func (instance JsonValue) String() string {
	buf := new(bytes.Buffer)
	encoder := json.NewEncoder(buf)
	if instance.PrettyPrint {
		encoder.SetIndent("", "  ")
	}
	if err := encoder.Encode(instance.Value); err != nil {
		panic(err)
	}
	return buf.String()
}

type HasLogger interface {
	Log() Logger
}

func OrDefault(in Logger) Logger {
	if in == nil {
		return Default
	}
	return in
}
