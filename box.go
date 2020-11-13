package goxr

import (
	"errors"
	"github.com/echocat/goxr/box/fs"
	"github.com/echocat/goxr/box/packed"
	"github.com/echocat/goxr/common"
	"github.com/echocat/goxr/log"
	"github.com/echocat/goxr/runtime"
	"io"
	"path/filepath"
	sr "runtime"
	"strings"
)

type OnFallbackToFsBoxFunc func(packedBoxCandidateFilename string, bases []string, fsBox Box) error

var (
	// This variable could easily set while build time using:
	// go build -ldflags="-X github.com/echocat/goxr.AllowFallbackToFsBox=false" .
	// This is useful in case for behave differently for build versions of you application
	AllowFallbackToFsBox                       = true
	OnFallbackToFsBox    OnFallbackToFsBoxFunc = OnFallbackToFsBox_Default

	ErrBoxIterationNotSupported = errors.New("box iteration not supported")
)

type Box interface {
	io.Closer
	Open(name string) (common.File, error)
	Info(pathname string) (common.FileInfo, error)
}

type Iterable interface {
	ForEach(common.FilePredicate, func(common.FileInfo) error) error
}

func OpenBox(base ...string) (Box, error) {
	if executable, err := runtime.Executable(); err != nil {
		return nil, err
	} else {
		return OpenBoxBy(executable, base...)
	}
}

func OpenPackedBox() (Box, error) {
	if executable, err := runtime.Executable(); err != nil {
		return nil, err
	} else {
		return packed.OpenBox(executable)
	}
}

func OpenBoxBy(packedBoxCandidateFilename string, base ...string) (Box, error) {
	if packedBox, err := packed.OpenBox(packedBoxCandidateFilename); common.IsDoesNotContainBox(err) {
		if box, err := openFsBox(base); err != nil {
			return nil, err
		} else if OnFallbackToFsBox == nil {
			return box, nil
		} else if err := OnFallbackToFsBox(packedBoxCandidateFilename, base, box); err != nil {
			return nil, err
		} else {
			return box, nil
		}
	} else if err != nil {
		return nil, err
	} else {
		return packedBox, nil
	}
}

func openFsBox(bases []string) (Box, error) {
	callingDir := resolveCallingDir(2)

	boxes := make(CombinedBox, len(bases))
	for i, base := range bases {
		if !filepath.IsAbs(base) {
			var err error
			if base, err = filepath.Abs(filepath.Join(callingDir, base)); err != nil {
				return nil, common.NewPathError("openBox", base, err)
			}
		}
		if box, err := fs.OpenBox(base); err != nil {
			return nil, err
		} else {
			boxes[i] = box
		}
	}

	return boxes, nil
}

func resolveCallingDir(skipCallerFrames int) string {
	_, filename, _, _ := sr.Caller(skipCallerFrames + 1)
	result := filepath.Dir(filename)

	// this little hack courtesy of the `-cover` flag!!
	cov := filepath.Join("_test", "_obj_test")
	result = strings.Replace(result, string(filepath.Separator)+cov, "", 1)
	if result != "" && !filepath.IsAbs(result) {
		result = filepath.Join(runtime.GoPath(), "src", result)
	}

	return result
}

//noinspection GoSnakeCaseUsage
func OnFallbackToFsBox_Default(packedBoxCandidateFilename string, bases []string, fsBox Box) error {
	if AllowFallbackToFsBox {
		return OnFallbackToFsBox_Warn(packedBoxCandidateFilename, bases, fsBox)
	}
	return OnFallbackToFsBox_Fail(packedBoxCandidateFilename, bases, fsBox)
}

//noinspection GoSnakeCaseUsage
func OnFallbackToFsBox_Warn(packedBoxCandidateFilename string, bases []string, fsBox Box) error {
	log.Warnf("%s does not contain a packed box version. This could happen in development mode", packedBoxCandidateFilename)
	return nil
}

//noinspection GoSnakeCaseUsage
func OnFallbackToFsBox_Fail(packedBoxCandidateFilename string, bases []string, fsBox Box) error {
	return common.NewPathError("openBox", packedBoxCandidateFilename, common.ErrDoesNotContainBox)
}
