package packed

import (
	"fmt"
	"github.com/echocat/goxr/common"
	"github.com/echocat/goxr/entry"
	"os"
	"strings"
	"time"
)

type Box struct {
	Name        string        `msgpack:"name"`
	Description string        `msgpack:"description"`
	Version     string        `msgpack:"version"`
	Revision    string        `msgpack:"revision"`
	Built       time.Time     `msgpack:"built"`
	BuiltBy     string        `msgpack:"builtBy"`
	Entries     entry.Entries `msgpack:"entries"`
	Meta        Meta          `msgpack:"meta"`

	EntryToFileTransformer ToFileTransformer `msgpack:"-"`
	OnClose                common.OnClose    `msgpack:"-"`
	Prefix                 string            `msgpack:"-"`
}

func (instance Box) String() string {
	return fmt.Sprintf(`%s
 Version:    %s
 Revision:   %s
 Built:      %v
 BuiltBy:    %v`,
		instance.Name, instance.Version, instance.Revision, instance.Built, instance.BuiltBy)
}

func (instance Box) ShortString() string {
	return fmt.Sprintf(`%s (version: %s, revision: %s)`,
		instance.Name, instance.Version, instance.Revision)
}

func (instance Box) LongVersion() string {
	return fmt.Sprintf(`%s (revision: %s)`,
		instance.Version, instance.Revision)
}

func (instance *Box) resolvePath(name string) (string, error) {
	candidate := entry.CleanPath(name)
	if instance.Prefix != "" {
		if !strings.HasPrefix(candidate, instance.Prefix) {
			return "", os.ErrNotExist
		}
		candidate = candidate[len(instance.Prefix):]
	}
	return candidate, nil
}

func (instance *Box) Open(pathname string) (common.File, error) {
	if candidate, err := instance.resolvePath(pathname); err != nil {
		return nil, common.NewPathError("open", pathname, err)
	} else if e := instance.Entries.Find(candidate); e == nil {
		return nil, common.NewPathError("open", pathname, os.ErrNotExist)
	} else if instance.EntryToFileTransformer == nil {
		return nil, common.NewPathError("open", pathname, entry.ErrNoToFileTransformerProvided)
	} else {
		return instance.EntryToFileTransformer("open", candidate, e)
	}
}

func (instance *Box) Info(pathname string) (common.FileInfo, error) {
	if candidate, err := instance.resolvePath(pathname); err != nil {
		return nil, common.NewPathError("open", pathname, err)
	} else if e := instance.Entries.Find(candidate); e == nil {
		return nil, common.NewPathError("info", pathname, os.ErrNotExist)
	} else {
		return *e, nil
	}
}

func (instance *Box) Close() error {
	onClose := instance.OnClose
	if onClose != nil {
		return onClose()
	}
	return nil
}

func (instance *Box) ForEach(predicate common.FilePredicate, callback func(common.FileInfo) error) error {
	for p, e := range instance.Entries {
		if predicate != nil {
			if ok, err := predicate(p); err != nil {
				return err
			} else if !ok {
				continue
			}
		}
		if err := callback(e); err != nil {
			return err
		}
	}
	return nil
}

type Meta map[string]interface{}
