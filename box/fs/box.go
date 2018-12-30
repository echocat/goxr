package fs

import (
	"github.com/blaubaer/goxr/common"
	"github.com/blaubaer/goxr/entry"
	"os"
	"path"
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

func (instance *Box) Open(name string) (common.File, error) {
	candidate := entry.CleanPath(name)
	if instance.prefix != "" {
		if !strings.HasPrefix(candidate, instance.prefix) {
			return nil, common.NewPathError("open", name, os.ErrNotExist)
		}
		candidate = candidate[len(instance.prefix):]
	}
	candidate = filepath.Clean(filepath.Join(instance.base, filepath.FromSlash(candidate)))

	if candidate != instance.base && !strings.HasPrefix(candidate, instance.baseWithSeparator) {
		return nil, common.NewPathError("open", name, os.ErrNotExist)
	} else if f, err := os.Open(candidate); err != nil {
		return nil, common.NewPathError("open", name, err)
	} else {
		return f, nil
	}
}

func (instance *Box) Info(name string) (os.FileInfo, error) {
	candidate := filepath.Clean(filepath.Join(instance.base, filepath.FromSlash(path.Clean(name))))
	if candidate != instance.base && !strings.HasPrefix(candidate, instance.baseWithSeparator) {
		return nil, common.NewPathError("info", name, os.ErrNotExist)
	} else if fi, err := os.Stat(candidate); err != nil {
		return nil, common.NewPathError("info", name, err)
	} else {
		return fi, nil
	}
}

func (instance *Box) Close() error {
	return nil
}
