package log

import (
	"fmt"
	"strings"
)

type ColorMode string

const (
	DefaultColorMode = ColorMode("")
	AutoColorMode    = ColorMode("auto")
	NeverColorMode   = ColorMode("never")
	AlwaysColorMode  = ColorMode("always")
)

var AllColorModes = []ColorMode{
	AutoColorMode,
	NeverColorMode,
	AlwaysColorMode,
}

func (instance *ColorMode) Set(plain string) error {
	switch ColorMode(plain) {
	case DefaultColorMode, AutoColorMode, NeverColorMode, AlwaysColorMode:
		*instance = ColorMode(plain)
	case "default":
		*instance = DefaultColorMode
		return nil
	}
	allStr := make([]string, len(AllColorModes))
	for i, v := range AllColorModes {
		allStr[i] = v.String()
	}
	return fmt.Errorf("illegal log color mode '%s' (possible values are: %s)", plain, strings.Join(allStr, ","))
}

func (instance ColorMode) String() string {
	return string(instance)
}
