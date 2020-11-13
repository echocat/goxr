package fs

import (
	"fmt"
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

func (instance *Box) clean(name string) (string, error) {
	candidate := entry.CleanPath(name)
	if instance.prefix != "" {
		if !strings.HasPrefix(candidate, instance.prefix) {
			return "", os.ErrNotExist
		}
		candidate = candidate[len(instance.prefix):]
	}
	return candidate, nil
}

func (instance *Box) resolvePath(name string) (string, error) {
	return filepath.Clean(filepath.Join(instance.base, filepath.FromSlash(name))), nil
}

func (instance *Box) Open(name string) (common.File, error) {
	if cleaned, err := instance.clean(name); err != nil {
		return nil, common.NewPathError("open", name, err)
	} else if candidate, err := instance.resolvePath(cleaned); err != nil {
		return nil, common.NewPathError("open", name, err)
	} else if candidate != instance.base && !strings.HasPrefix(candidate, instance.baseWithSeparator) {
		return nil, common.NewPathError("open", name, os.ErrNotExist)
	} else if f, err := os.Open(candidate); err != nil {
		return nil, common.NewPathError("open", name, err)
	} else {
		return &file{f, cleaned}, nil
	}
}

func (instance *Box) Info(name string) (common.FileInfo, error) {
	if cleaned, err := instance.clean(name); err != nil {
		return nil, common.NewPathError("info", name, err)
	} else if candidate, err := instance.resolvePath(cleaned); err != nil {
		return nil, common.NewPathError("info", name, err)
	} else if candidate != instance.base && !strings.HasPrefix(candidate, instance.baseWithSeparator) {
		return nil, common.NewPathError("info", name, os.ErrNotExist)
	} else if fi, err := os.Stat(candidate); err != nil {
		return nil, common.NewPathError("info", name, err)
	} else if fi.IsDir() {
		return nil, common.NewPathError("info", name, os.ErrNotExist)
	} else {
		return &fileInfo{fi, candidate}, nil
	}
}

func (instance *Box) ForEach(predicate common.FilePredicate, callback func(string, common.FileInfo) error) error {
	base, err := filepath.Abs(instance.base)
	if err != nil {
		return fmt.Errorf("cannot iterate over box %s: %v", instance.base, err)
	}
	base += fmt.Sprintf("%c", os.PathSeparator)

	if err := filepath.Walk(instance.base, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		fullPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		if !strings.HasPrefix(fullPath, base) {
			return fmt.Errorf("found inside of '%s' a file (%s) which is not in the same path", base, fullPath)
		}
		p := filepath.ToSlash(fullPath[len(base):])
		if predicate != nil {
			if ok, err := predicate(p); err != nil {
				return err
			} else if !ok {
				return nil
			}
		}
		return callback(p, &fileInfo{info, p})
	}); err != nil {
		return fmt.Errorf("cannot iterate over box %s: %v", instance.base, err)
	}

	return nil
}

func (instance *Box) Close() error {
	return nil
}

type file struct {
	*os.File
	path string
}

func (instance *file) Stat() (os.FileInfo, error) {
	return instance.GetFileInfo()
}

func (instance *file) GetFileInfo() (common.FileInfo, error) {
	fi, err := instance.File.Stat()
	if err != nil {
		return nil, err
	}
	return &fileInfo{fi, instance.path}, nil
}

type fileInfo struct {
	os.FileInfo
	path string
}

func (instance *fileInfo) Path() string {
	return instance.path
}
