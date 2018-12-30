package entry

import (
	"io"
)

type Reader interface {
	io.Reader
	io.Seeker
}

type ReaderFactory func(entry Entry) (Reader, error)
