package common

import (
	"errors"
	"net"
	"net/url"
	"os"
)

var (
	ErrDoesContainBox    = errors.New("does contain box")
	ErrDoesNotContainBox = errors.New("does not contain box")
	ErrAlreadyRunning    = errors.New("already running")
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
		case *url.Error:
			err = candidate.Err
		case *net.OpError:
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
