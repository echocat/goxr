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
	GetFileInfo() (FileInfo, error)
}

type FileInfo interface {
	os.FileInfo
	Path() string
}

type ExtendedFileInfo interface {
	FileInfo
	ChecksumString() string
}

type FilePredicate func(name string) (bool, error)
