package log

import (
	"fmt"
	"strings"
)

type Format string

const (
	DefaultFormat = Format("")
	TextFormat    = Format("text")
	JsonFormat    = Format("json")
)

var AllFormats = []Format{
	TextFormat,
	JsonFormat,
}

func (instance *Format) Set(plain string) error {
	for _, candidate := range AllFormats {
		if candidate.String() == plain {
			*instance = candidate
			return nil
		}
	}
	switch plain {
	case "default":
		*instance = DefaultFormat
		return nil
	}
	allStr := make([]string, len(AllFormats))
	for i, v := range AllFormats {
		allStr[i] = v.String()
	}
	return fmt.Errorf("illegal log format '%s' (possible values are: %s)", plain, strings.Join(allStr, ","))
}

func (instance Format) String() string {
	return string(instance)
}
