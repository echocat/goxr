package log

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"os"
	"reflect"
)

type LogrusLevel struct {
	logrus.Level
}

func (instance *LogrusLevel) Set(plain string) error {
	return instance.UnmarshalText([]byte(plain))
}

func (instance LogrusLevel) String() string {
	return instance.Level.String()
}

func (instance *LogrusLevel) Equals(o Level) bool {
	if lil, ok := o.(*LogrusLevel); ok {
		return lil.Level == instance.Level
	}
	return false
}

type LogrusFormat string

func (instance *LogrusFormat) Set(plain string) error {
	if plain != "json" && plain != "text" {
		return fmt.Errorf("unsupported log format: %s", plain)
	}
	*instance = LogrusFormat(plain)
	return nil
}

func (instance LogrusFormat) String() string {
	return string(instance)
}

type LogrusColorMode string

func (instance *LogrusColorMode) Set(plain string) error {
	if plain != "auto" && plain != "never" && plain != "always" {
		return fmt.Errorf("unsupported log color mode: %s", plain)
	}
	*instance = LogrusColorMode(plain)
	return nil
}

func (instance LogrusColorMode) String() string {
	return string(instance)
}

type LogrusLogger struct {
	Level              *LogrusLevel
	Format             LogrusFormat
	ColorMode          LogrusColorMode
	ReportCaller       bool
	Delegate           *logrus.Logger
	EntryLoggerFactory func(*logrus.Logger) Logger
}

func (instance *LogrusLogger) CreateEntryLogger() Logger {
	if instance.EntryLoggerFactory == nil {
		return &LogrusEntry{
			Root:     instance,
			Delegate: logrus.NewEntry(instance.Delegate),
		}
	}
	return instance.EntryLoggerFactory(instance.Delegate)
}

func (instance *LogrusLogger) WithField(key string, value interface{}) Logger {
	return instance.CreateEntryLogger().WithField(key, value)
}

func (instance *LogrusLogger) WithDeepField(key string, value interface{}) Logger {
	return instance.CreateEntryLogger().WithDeepField(key, value)
}

func (instance *LogrusLogger) WithDeepFieldOn(key string, value interface{}, on func() bool) Logger {
	return instance.CreateEntryLogger().WithDeepFieldOn(key, value, on)
}

func (instance *LogrusLogger) WithError(err error) Logger {
	return instance.CreateEntryLogger().WithError(err)
}

func (instance *LogrusLogger) Trace(what ...interface{}) {
	instance.CreateEntryLogger().Trace(what...)
}

func (instance *LogrusLogger) Debug(what ...interface{}) {
	instance.CreateEntryLogger().Debug(what...)
}

func (instance *LogrusLogger) Info(what ...interface{}) {
	instance.CreateEntryLogger().Info(what...)
}

func (instance *LogrusLogger) Warn(what ...interface{}) {
	instance.CreateEntryLogger().Warn(what...)
}

func (instance *LogrusLogger) Error(what ...interface{}) {
	instance.CreateEntryLogger().Error(what...)
}

func (instance *LogrusLogger) Fatal(what ...interface{}) {
	instance.CreateEntryLogger().Fatal(what...)
}
func (instance *LogrusLogger) Tracef(msg string, args ...interface{}) {
	instance.CreateEntryLogger().Tracef(msg, args...)
}

func (instance *LogrusLogger) Debugf(msg string, args ...interface{}) {
	instance.CreateEntryLogger().Debugf(msg, args...)
}

func (instance *LogrusLogger) Infof(msg string, args ...interface{}) {
	instance.CreateEntryLogger().Infof(msg, args...)
}

func (instance *LogrusLogger) Warnf(msg string, args ...interface{}) {
	instance.CreateEntryLogger().Warnf(msg, args...)
}

func (instance *LogrusLogger) Errorf(msg string, args ...interface{}) {
	instance.CreateEntryLogger().Errorf(msg, args...)
}

func (instance *LogrusLogger) Fatalf(msg string, args ...interface{}) {
	instance.CreateEntryLogger().Fatalf(msg, args...)
}

func (instance *LogrusLogger) IsTraceEnabled() bool {
	return instance.Delegate.Level >= logrus.TraceLevel
}

func (instance *LogrusLogger) IsDebugEnabled() bool {
	return instance.Delegate.Level >= logrus.DebugLevel
}

func (instance *LogrusLogger) IsInfoEnabled() bool {
	return instance.Delegate.Level >= logrus.InfoLevel
}

func (instance *LogrusLogger) IsWarnEnabled() bool {
	return instance.Delegate.Level >= logrus.WarnLevel
}

func (instance *LogrusLogger) IsErrorEnabled() bool {
	return instance.Delegate.Level >= logrus.ErrorLevel
}

func (instance *LogrusLogger) IsFatalEnabled() bool {
	return instance.Delegate.Level >= logrus.FatalLevel
}

func (instance *LogrusLogger) GetLevel() Level {
	return &LogrusLevel{
		Level: instance.Delegate.Level,
	}
}

func (instance *LogrusLogger) SetLevel(l Level) error {
	if lgl, ok := l.(*LogrusLevel); ok {
		instance.Level = lgl
		instance.Delegate.Level = lgl.Level
		return nil
	}
	return os.ErrInvalid
}

func (instance *LogrusLogger) Flags() []cli.Flag {
	return []cli.Flag{
		cli.GenericFlag{
			Name:  "logLevel",
			Usage: "Specifies the minimum required log level.",
			Value: instance.Level,
		},
		cli.GenericFlag{
			Name:  "logFormat",
			Usage: "Specifies format output (text or json).",
			Value: &instance.Format,
		},
		cli.GenericFlag{
			Name:  "logColorMode",
			Usage: "Specifies if the output is in colors or not (auto, never or always).",
			Value: &instance.ColorMode,
		},
		cli.BoolFlag{
			Name:        "logCaller",
			Usage:       "If true the caller details will be logged too.",
			Destination: &instance.ReportCaller,
		},
	}
}

func (instance *LogrusLogger) Init() error {
	instance.Delegate.Level = instance.Level.Level
	instance.Delegate.ReportCaller = instance.ReportCaller

	textFormatter := &logrus.TextFormatter{
		FullTimestamp:    true,
		QuoteEmptyFields: true,
	}
	switch instance.ColorMode {
	case LogrusColorMode("always"):
		textFormatter.ForceColors = true
	case LogrusColorMode("never"):
		textFormatter.DisableColors = true
	}

	instance.Delegate.Formatter = textFormatter
	switch instance.Format {
	case LogrusFormat("json"):
		instance.Delegate.Formatter = &logrus.JSONFormatter{}
	}
	return nil
}

type LogrusEntry struct {
	Root     *LogrusLogger
	Delegate *logrus.Entry
}

func (instance *LogrusEntry) WithField(key string, value interface{}) Logger {
	return &LogrusEntry{
		Root:     instance.Root,
		Delegate: instance.Delegate.WithField(key, value),
	}
}

func (instance *LogrusEntry) WithDeepField(key string, value interface{}) Logger {
	return instance.WithField(key, JsonValue{
		Value: value,
	})
}

func (instance *LogrusEntry) WithDeepFieldOn(key string, value interface{}, on func() bool) Logger {
	if on() {
		return instance.WithDeepField(key, value)
	}
	return instance
}

func (instance *LogrusEntry) WithError(err error) Logger {
	return &LogrusEntry{
		Root:     instance.Root,
		Delegate: instance.Delegate.WithError(err),
	}
}

type skipArgType struct {
	val uint16
}

var skipArg = skipArgType{666}

func (instance *LogrusEntry) evalArg(what interface{}) interface{} {
	v := reflect.Indirect(reflect.ValueOf(what))
	t := v.Type()
	if t.Kind() == reflect.Map {
		for _, key := range v.MapKeys() {
			sKey := fmt.Sprint(key.Interface())
			instance.Delegate.Data[sKey] = v.MapIndex(key).Interface()
		}
		return skipArg
	}
	return what
}

func (instance *LogrusEntry) evalArgs(what []interface{}) []interface{} {
	var result []interface{}
	for _, arg := range what {
		arg = instance.evalArg(arg)
		if arg != skipArg {
			result = append(result, arg)
		}
	}
	return result
}

func (instance *LogrusEntry) Trace(what ...interface{}) {
	instance.Delegate.Trace(instance.evalArgs(what)...)
}

func (instance *LogrusEntry) Debug(what ...interface{}) {
	instance.Delegate.Debug(instance.evalArgs(what)...)
}

func (instance *LogrusEntry) Info(what ...interface{}) {
	instance.Delegate.Info(instance.evalArgs(what)...)
}

func (instance *LogrusEntry) Warn(what ...interface{}) {
	instance.Delegate.Warn(instance.evalArgs(what)...)
}

func (instance *LogrusEntry) Error(what ...interface{}) {
	instance.Delegate.Error(instance.evalArgs(what)...)
}

func (instance *LogrusEntry) Fatal(what ...interface{}) {
	instance.Delegate.Fatal(instance.evalArgs(what)...)
}
func (instance *LogrusEntry) Tracef(msg string, args ...interface{}) {
	instance.Delegate.Tracef(msg, args...)
}

func (instance *LogrusEntry) Debugf(msg string, args ...interface{}) {
	instance.Delegate.Debugf(msg, args...)
}

func (instance *LogrusEntry) Infof(msg string, args ...interface{}) {
	instance.Delegate.Infof(msg, args...)
}

func (instance *LogrusEntry) Warnf(msg string, args ...interface{}) {
	instance.Delegate.Warnf(msg, args...)
}

func (instance *LogrusEntry) Errorf(msg string, args ...interface{}) {
	instance.Delegate.Errorf(msg, args...)
}

func (instance *LogrusEntry) Fatalf(msg string, args ...interface{}) {
	instance.Delegate.Fatalf(msg, args...)
}

func (instance *LogrusEntry) IsTraceEnabled() bool {
	return instance.Root.IsTraceEnabled()
}

func (instance *LogrusEntry) IsDebugEnabled() bool {
	return instance.Root.IsDebugEnabled()
}

func (instance *LogrusEntry) IsInfoEnabled() bool {
	return instance.Root.IsInfoEnabled()
}

func (instance *LogrusEntry) IsWarnEnabled() bool {
	return instance.Root.IsWarnEnabled()
}

func (instance *LogrusEntry) IsErrorEnabled() bool {
	return instance.Root.IsErrorEnabled()
}

func (instance *LogrusEntry) IsFatalEnabled() bool {
	return instance.Root.IsFatalEnabled()
}
