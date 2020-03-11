package fs

import (
	"github.com/echocat/goxr/common"
	"github.com/echocat/goxr/entry"
	"os"
	"path/filepath"
	"strings"
)

func OpenBox(base string) (*Box, error) {
	result := &Box{
		base:   base,
		prefix: "",
	}
	parts := strings.SplitN(result.base, "=", 2)
	if len(parts) > 1 {
		result.prefix = entry.CleanPath(filepath.ToSlash(parts[0]))
		if result.prefix != "" {
			result.prefix += "/"
		}
		result.base = parts[1]
	}

	var err error
	if result.base, err = filepath.Abs(result.base); err != nil {
		return nil, common.NewPathError("openBox", result.base, err)
	}

	result.baseWithSeparator = string(append([]rune(result.base), filepath.Separator))
	return result, nil
}

type Box struct {
	base              string
	baseWithSeparator string
	prefix            string
}

func (instance *Box) resolvePath(name string) (string, error) {
	candidate := entry.CleanPath(name)
	if instance.prefix != "" {
		if !strings.HasPrefix(candidate, instance.prefix) {
			return "", os.ErrNotExist
		}
		candidate = candidate[len(instance.prefix):]
	}
	return filepath.Clean(filepath.Join(instance.base, filepath.FromSlash(candidate))), nil
}

func (instance *Box) Open(name string) (common.File, error) {
	if candidate, err := instance.resolvePath(name); err != nil {
		return nil, common.NewPathError("open", name, err)
	} else if candidate != instance.base && !strings.HasPrefix(candidate, instance.baseWithSeparator) {
		return nil, common.NewPathError("open", name, os.ErrNotExist)
	} else if f, err := os.Open(candidate); err != nil {
		return nil, common.NewPathError("open", name, err)
	} else {
		return f, nil
	}
}

func (instance *Box) Info(name string) (os.FileInfo, error) {
	if candidate, err := instance.resolvePath(name); err != nil {
		return nil, common.NewPathError("info", name, err)
	} else if candidate != instance.base && !strings.HasPrefix(candidate, instance.baseWithSeparator) {
		return nil, common.NewPathError("info", name, os.ErrNotExist)
	} else if fi, err := os.Stat(candidate); err != nil {
		return nil, common.NewPathError("info", name, err)
	} else if fi.IsDir() {
		return nil, common.NewPathError("info", name, os.ErrNotExist)
	} else {
		return fi, nil
	}
}

func (instance *Box) Close() error {
	return nil
}
