package entry

import "errors"

var (
	ErrNoReaderFactoryProvided     = errors.New("no entry.ReaderFactory provided")
	ErrNoToFileTransformerProvided = errors.New("no entry.ToFileTransformer provided")
)
