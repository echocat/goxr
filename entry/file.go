package entry

import (
	"github.com/echocat/goxr/common"
	"io"
	"os"
)

type File struct {
	Entry         Entry
	Path          string
	ReaderFactory ReaderFactory

	closed      bool
	entryReader Reader
}

func (instance *File) Close() error {
	if instance.closed {
		return common.NewPathError("close", instance.Path, common.ErrAlreadyClosed)
	}

	instance.closed = true
	return nil
}

func (instance *File) ensureReader(operation string) (Reader, error) {
	if instance.closed {
		return nil, common.NewPathError(operation, instance.Path, common.ErrAlreadyClosed)
	}
	if instance.entryReader == nil {
		factory := instance.ReaderFactory
		if factory == nil {
			return nil, common.NewPathError(operation, instance.Path, ErrNoReaderFactoryProvided)
		}
		if r, err := factory(instance.Entry); err != nil {
			return nil, common.NewPathError(operation, instance.Path, err)
		} else {
			instance.entryReader = r
		}
	}
	return instance.entryReader, nil
}

func (instance *File) Read(p []byte) (n int, err error) {
	if r, err := instance.ensureReader("read"); err != nil {
		return 0, err
	} else if n, err := r.Read(p); err == io.EOF {
		return 0, io.EOF
	} else if err != nil {
		return 0, common.NewPathError("read", instance.Path, err)
	} else {
		return n, err
	}
}

func (instance *File) Seek(offset int64, whence int) (int64, error) {
	if r, err := instance.ensureReader("seek"); err != nil {
		return 0, err
	} else if n, err := r.Seek(offset, whence); err == io.EOF {
		return 0, io.EOF
	} else if err != nil {
		return 0, common.NewPathError("seek", instance.Path, err)
	} else {
		return n, err
	}
}

func (instance *File) Readdir(count int) ([]os.FileInfo, error) {
	return []os.FileInfo{}, nil
}

func (instance *File) Stat() (os.FileInfo, error) {
	return instance.Entry, nil
}

func (instance File) String() string {
	return instance.Path
}
