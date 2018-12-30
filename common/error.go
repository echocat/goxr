package common

import (
	"errors"
	"os"
)

var (
	ErrAlreadyClosed     = errors.New("already closed")
	ErrDoesContainBox    = errors.New("does contain box")
	ErrDoesNotContainBox = errors.New("does not contain box")
)

func NewPathError(operation string, path string, detail error) *os.PathError {
	return &os.PathError{
		Op:   operation,
		Path: path,
		Err:  UnderlyingError(detail),
	}
}

func UnderlyingError(err error) error {
	for {
		switch candidate := err.(type) {
		case *os.PathError:
			err = candidate.Err
		case *os.LinkError:
			err = candidate.Err
		case *os.SyscallError:
			err = candidate.Err
		default:
			return err
		}
	}
}

func IsDoesContainBox(err error) bool {
	err = UnderlyingError(err)
	return err == ErrDoesContainBox
}

func IsDoesNotContainBox(err error) bool {
	err = UnderlyingError(err)
	return err == ErrDoesNotContainBox
}
