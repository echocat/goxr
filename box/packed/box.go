package packed

import (
	"fmt"
	"github.com/blaubaer/goxr/common"
	"github.com/blaubaer/goxr/entry"
	"os"
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

func (instance *Box) Open(pathname string) (common.File, error) {
	if e := instance.Entries.Find(pathname); e == nil {
		return nil, common.NewPathError("open", pathname, os.ErrNotExist)
	} else if instance.EntryToFileTransformer == nil {
		return nil, common.NewPathError("open", pathname, entry.ErrNoToFileTransformerProvided)
	} else {
		return instance.EntryToFileTransformer("open", pathname, *e)
	}
}

func (instance *Box) Info(pathname string) (os.FileInfo, error) {
	if e := instance.Entries.Find(pathname); e == nil {
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

type Meta map[string]interface{}
