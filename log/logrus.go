package log

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"reflect"
)

type LogrusLogger struct {
	configuration Configuration

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

func (instance *LogrusLogger) SetConfiguration(configuration Configuration) error {
	lvl, err := logrus.ParseLevel(configuration.GetLevel(InfoLevel).String())
	if err != nil {
		return err
	}

	var formatter logrus.Formatter
	switch configuration.GetFormat(TextFormat) {
	case JsonFormat:
		formatter = &logrus.JSONFormatter{}
	default:
		formatter = &logrus.TextFormatter{
			FullTimestamp:    true,
			QuoteEmptyFields: true,
		}
		switch configuration.GetColorMode(AutoColorMode) {
		case AlwaysColorMode:
			formatter.(*logrus.TextFormatter).ForceColors = true
		case NeverColorMode:
			formatter.(*logrus.TextFormatter).DisableColors = true
		}
	}

	instance.Delegate.Level = lvl
	instance.Delegate.Formatter = formatter
	instance.Delegate.ReportCaller = configuration.ReportCaller
	return nil
}

func (instance LogrusLogger) GetConfiguration() Configuration {
	return instance.configuration
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
