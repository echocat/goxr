package runtime

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

//noinspection SpellCheckingInspection
var (
	_GOPATH string
)

func init() {
	if v, ok := os.LookupEnv("GOPATH"); ok {
		_GOPATH = v
	} else if u, err := user.Current(); err != nil || u.HomeDir == "" {
		_GOPATH = ""
	} else {
		_GOPATH = filepath.Join(u.HomeDir, "go")
	}
}

func GoPath() string {
	return _GOPATH
}

func Executable() (string, error) {
	if executable, err := os.Executable(); err != nil {
		return "", fmt.Errorf("cannot resolve executable of the actual process: %v", err)
	} else {
		return executable, nil
	}
}
