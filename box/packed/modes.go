package packed

import (
	"os"
	"strings"
)

type OpenMode struct {
	name   string
	create bool
	open   bool
}

var (
	OpenModeOpenOrCreate = OpenMode{name: "openOrCreate", create: true, open: true}
	OpenModeOpenOnly     = OpenMode{name: "openOnly", open: true}
	OpenModeCreateOnly   = OpenMode{name: "createOnly", create: true}

	openModes        = []OpenMode{OpenModeOpenOrCreate, OpenModeOpenOnly, OpenModeCreateOnly}
	lowerToOpenModes = func(modes []OpenMode) map[string]OpenMode {
		result := make(map[string]OpenMode)
		for _, mode := range modes {
			result[strings.ToLower(mode.String())] = mode
		}
		return result
	}(openModes)
)

func OpenModes() []OpenMode {
	return openModes
}

func (instance *OpenMode) Set(in string) error {
	lIn := strings.ToLower(in)
	for req, candidate := range lowerToOpenModes {
		if req == lIn {
			*instance = candidate
			return nil
		}
	}
	return os.ErrInvalid
}

func (instance OpenMode) IsOpen() bool {
	return instance.open
}

func (instance OpenMode) IsCreate() bool {
	return instance.create
}

func (instance OpenMode) String() string {
	return instance.name
}

type WriteMode struct {
	name    string
	new     bool
	replace bool
}

var (
	WriteModeNewOrReplace = WriteMode{name: "newOrReplace", new: true, replace: true}
	WriteModeNewOnly      = WriteMode{name: "newOnly", new: true}
	WriteModeReplaceOnly  = WriteMode{name: "replaceOnly", replace: true}

	writeModes        = []WriteMode{WriteModeNewOrReplace, WriteModeNewOnly, WriteModeReplaceOnly}
	lowerToWriteModes = func(modes []WriteMode) map[string]WriteMode {
		result := make(map[string]WriteMode)
		for _, mode := range modes {
			result[strings.ToLower(mode.String())] = mode
		}
		return result
	}(writeModes)
)

func WriteModes() []WriteMode {
	return writeModes
}

func (instance *WriteMode) Set(in string) error {
	lIn := strings.ToLower(in)
	for req, candidate := range lowerToWriteModes {
		if req == lIn {
			*instance = candidate
			return nil
		}
	}
	return os.ErrInvalid
}

func (instance WriteMode) IsNew() bool {
	return instance.new
}

func (instance WriteMode) IsReplace() bool {
	return instance.replace
}

func (instance WriteMode) String() string {
	return instance.name
}
