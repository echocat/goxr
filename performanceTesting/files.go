package main

import (
	"fmt"
	"github.com/c2h5oh/datasize"
	"os"
	"path/filepath"
	"strings"
)

var (
	staticFileIndexHtml = staticFileBy("index.html")
	staticFiles         = files{
		staticFileIndexHtml,
	}
	baseFilenameToStaticFile = func(ins []file) map[string]staticFile {
		result := make(map[string]staticFile)
		for _, in := range ins {
			sf := in.(staticFile)
			result[sf.baseFileName] = sf
		}
		return result
	}(staticFiles)
)

type file interface {
	fmt.Stringer
	getPath() string
	isTemporary() bool
	getSize() datasize.ByteSize
	ensure()
}

func staticFileBy(baseFileName string) staticFile {
	p := filepath.Join(filesDirectory, baseFileName)
	return staticFile{
		baseFileName: baseFileName,
		path:         p,
		fileInfo:     fileInfo(p),
	}
}

type staticFile struct {
	baseFileName string
	path         string
	fileInfo     os.FileInfo
}

func (instance staticFile) String() string {
	return instance.baseFileName
}

func (instance staticFile) getPath() string {
	return instance.path
}

func (instance staticFile) isTemporary() bool {
	return false
}

func (instance staticFile) ensure() {}

func (instance staticFile) getSize() datasize.ByteSize {
	return datasize.ByteSize(instance.fileInfo.Size())
}

func temporaryFileBy(size datasize.ByteSize) temporaryFile {
	p := filepath.Join(filesDirectory, fmt.Sprintf("blob-%s.bin", size))

	return temporaryFile{
		path: p,
		size: size,
	}
}

type temporaryFile struct {
	path string
	size datasize.ByteSize
}

func (instance temporaryFile) String() string {
	return instance.size.String()
}

func (instance temporaryFile) getPath() string {
	return instance.path
}

func (instance temporaryFile) isTemporary() bool {
	return true
}

func (instance temporaryFile) getSize() datasize.ByteSize {
	return instance.size
}

func (instance temporaryFile) ensure() {
	if fi, err := os.Stat(instance.path); os.IsNotExist(err) {
		generateFile(instance.path, instance.size)
	} else if err != nil {
		panic(err)
	} else if fi.IsDir() {
		remove(instance.path)
		generateFile(instance.path, instance.size)
	} else if datasize.ByteSize(fi.Size()) != instance.size {
		generateFile(instance.path, instance.size)
	}
}

type files []file

func (instance files) String() string {
	return strings.Join(instance.Strings(), ",")
}

func (instance files) Get() interface{} {
	return &instance
}

func (instance files) MarshalText() ([]byte, error) {
	return []byte(instance.String()), nil
}

func (instance files) Strings() []string {
	result := make([]string, len(instance))
	for i, bs := range instance {
		result[i] = bs.String()
	}
	return result
}

func (instance *files) Set(plain string) error {
	var result files
	for _, str := range strings.Split(plain, ",") {
		str = strings.TrimSpace(str)
		if str != "" {
			size := new(datasize.ByteSize)
			if err := size.UnmarshalText([]byte(str)); err == nil {
				result = append(result, temporaryFileBy(*size))
			} else if sf, ok := baseFilenameToStaticFile[str]; ok {
				result = append(result, sf)
			} else {
				return fmt.Errorf("cannot handle file reference: %s", str)
			}
		}
	}
	*instance = result
	return nil
}

func (instance *files) UnmarshalText(t []byte) error {
	return instance.Set(string(t))
}

func (instance files) ensure() {
	for _, file := range instance {
		file.ensure()
	}
}
