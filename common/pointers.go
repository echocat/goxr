package common

import (
	"os"
	"time"
)

var (
	Ptrue  = Pbool(true)
	Pfalse = Pbool(false)
)

func Pstring(v string) *string {
	return &v
}
func PtimeTime(v time.Time) *time.Time {
	return &v
}

func PosFileMode(v os.FileMode) *os.FileMode {
	return &v
}

func Pbool(v bool) *bool {
	return &v
}
