package common

import (
	"os"
	"time"
)

func PtimeTime(v time.Time) *time.Time {
	return &v
}

func PosFileMode(v os.FileMode) *os.FileMode {
	return &v
}
