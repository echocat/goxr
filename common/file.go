package common

import (
	"io"
	"os"
)

type File interface {
	io.Closer
	io.Reader
	io.Seeker
	Readdir(count int) ([]os.FileInfo, error)
	Stat() (os.FileInfo, error)
}

type ExtendedFileInfo interface {
	os.FileInfo
	ChecksumString() string
}

type FilePredicate func(name string) (bool, error)
