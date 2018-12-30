package log

import (
	"bytes"
	"encoding/json"
	"flag"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var DefaultLogger RootLogger = &LogrusLogger{
	Level:     &LogrusLevel{logrus.InfoLevel},
	Format:    LogrusFormat("text"),
	ColorMode: LogrusColorMode("auto"),
	Delegate:  logrus.New(),
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

	Init() error
	Flags() []cli.Flag
	GetLevel() Level
	SetLevel(Level) error
}

func WithField(key string, value interface{}) Logger {
	return DefaultLogger.WithField(key, value)
}

func WithDeepField(key string, value interface{}) Logger {
	return DefaultLogger.WithDeepField(key, value)
}

func WithDeepFieldOn(key string, value interface{}, on func() bool) Logger {
	return DefaultLogger.WithDeepFieldOn(key, value, on)
}

func WithError(err error) Logger {
	return DefaultLogger.WithError(err)
}

func Tracef(msg string, args ...interface{}) {
	DefaultLogger.Tracef(msg, args...)
}

func Debugf(msg string, args ...interface{}) {
	DefaultLogger.Debugf(msg, args...)
}

func Infof(msg string, args ...interface{}) {
	DefaultLogger.Infof(msg, args...)
}

func Warnf(msg string, args ...interface{}) {
	DefaultLogger.Warnf(msg, args...)
}

func Errorf(msg string, args ...interface{}) {
	DefaultLogger.Errorf(msg, args...)
}

func Fatalf(msg string, args ...interface{}) {
	DefaultLogger.Fatalf(msg, args...)
}

func IsTraceEnabled() bool {
	return DefaultLogger.IsTraceEnabled()
}

func IsDebugEnabled() bool {
	return DefaultLogger.IsDebugEnabled()
}

func IsInfoEnabled() bool {
	return DefaultLogger.IsInfoEnabled()
}

func IsWarnEnabled() bool {
	return DefaultLogger.IsWarnEnabled()
}

func IsErrorEnabled() bool {
	return DefaultLogger.IsErrorEnabled()
}

func IsFatalEnabled() bool {
	return DefaultLogger.IsFatalEnabled()
}

func GetLevel() Level {
	return DefaultLogger.GetLevel()
}

func SetLevel(l Level) error {
	return DefaultLogger.SetLevel(l)
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
		return DefaultLogger
	}
	return in
}

type Level interface {
	flag.Value
	Equals(Level) bool
}
