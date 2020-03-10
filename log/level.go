package log

import (
	"fmt"
	"strings"
)

var AllLevels = []Level{
	PanicLevel,
	FatalLevel,
	ErrorLevel,
	WarnLevel,
	InfoLevel,
	DebugLevel,
	TraceLevel,
}

const (
	DefaultLevel = Level("")
	PanicLevel   = Level("panic")
	FatalLevel   = Level("fatal")
	ErrorLevel   = Level("error")
	WarnLevel    = Level("warn")
	InfoLevel    = Level("info")
	DebugLevel   = Level("debug")
	TraceLevel   = Level("trace")
)

type Level string

func (instance *Level) Set(plain string) error {
	plain = strings.ToLower(plain)
	for _, candidate := range AllLevels {
		if candidate.String() == plain {
			*instance = candidate
			return nil
		}
	}

	switch plain {
	case "warning":
		*instance = WarnLevel
		return nil
	case "", "default":
		*instance = DefaultLevel
		return nil
	}

	allStr := make([]string, len(AllLevels))
	for i, v := range AllLevels {
		allStr[i] = v.String()
	}
	return fmt.Errorf("illegal log level '%s' (possible values are: %s)", plain, strings.Join(allStr, ","))
}

func (instance Level) String() string {
	return string(instance)
}
