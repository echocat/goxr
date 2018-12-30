package common

import (
	"fmt"
	"io"
	"os"
)

type FileOffset int64
type FileSize int64

func Write(what []byte, to io.Writer) error {
	n, err := to.Write(what)
	if err != nil {
		return err
	}
	if n < len(what) {
		return io.ErrShortWrite
	}
	return nil
}

func MustWrite(what []byte, to io.Writer) {
	if err := Write(what, to); err != nil {
		panic(err)
	}
}

func Writef(to io.Writer, pattern string, args ...interface{}) error {
	return Write([]byte(fmt.Sprintf(pattern, args...)), to)
}

func MustWritef(to io.Writer, pattern string, args ...interface{}) {
	if err := Writef(to, pattern, args...); err != nil {
		panic(err)
	}
}

func Read(from io.Reader, to []byte) error {
	n, err := from.Read(to)
	if err != nil {
		return err
	}
	if n < len(to) {
		return io.EOF
	}
	return nil
}

func MustRead(from io.Reader, to []byte) {
	if err := Read(from, to); err != nil {
		panic(err)
	}
}

func ReadBytes(from io.Reader, amount int) ([]byte, error) {
	to := make([]byte, amount)
	if err := Read(from, to); err != nil {
		return nil, err
	}
	return to, nil
}

func MustReadBytes(from io.Reader, amount int) []byte {
	if b, err := ReadBytes(from, amount); err != nil {
		panic(err)
	} else {
		return b
	}
}

func Seek(to FileOffset, on io.Seeker) error {
	if n, err := on.Seek(int64(to), 0); err != nil {
		return err
	} else if n < int64(to) {
		return io.EOF
	} else {
		return nil
	}
}

func Close(what io.Closer) error {
	if err := what.Close(); UnderlyingError(err) == os.ErrClosed {
		return nil
	} else if err != nil {
		return err
	} else {
		return nil
	}
}

type OnClose func() error
