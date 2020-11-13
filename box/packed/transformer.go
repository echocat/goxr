package packed

import (
	"github.com/echocat/goxr/common"
	"github.com/echocat/goxr/entry"
)

type ToFileTransformer func(operation string, fullPath string, entry *entry.Entry) (common.File, error)

func ToFileTransformerFor(rf entry.ReaderFactory) ToFileTransformer {
	return func(operation string, fullPath string, e *entry.Entry) (common.File, error) {
		return &entry.File{
			Entry:         e,
			Path:          fullPath,
			ReaderFactory: rf,
		}, nil
	}
}
