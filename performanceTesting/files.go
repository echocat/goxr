package main

import (
	"fmt"
	"github.com/c2h5oh/datasize"
	"github.com/echocat/goxr/log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	staticFileIndexHtml = staticFileBy("index.html")
	staticFiles         = fileSlice{
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
	files = append(staticFiles, temporaryFileBy(1*datasize.MB))
)

type file interface {
	fmt.Stringer
	test
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

func (instance staticFile) name() string {
	return instance.baseFileName
}

func (instance staticFile) String() string {
	return instance.name()
}

func (instance staticFile) getPath() string {
	return filepath.Base(instance.path)
}

func (instance staticFile) prepare() {}

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

func (instance temporaryFile) name() string {
	return instance.size.String()
}

func (instance temporaryFile) String() string {
	return instance.name()
}

func (instance temporaryFile) getPath() string {
	return filepath.Base(instance.path)
}

func (instance temporaryFile) getSize() datasize.ByteSize {
	return instance.size
}

func (instance temporaryFile) prepare() {
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

type fileSlice []file

func (instance fileSlice) String() string {
	return strings.Join(instance.Strings(), ",")
}

func (instance fileSlice) Get() interface{} {
	return &instance
}

func (instance fileSlice) Strings() []string {
	result := make([]string, len(instance))
	for i, bs := range instance {
		result[i] = bs.String()
	}
	return result
}

func (instance *fileSlice) Set(plain string) error {
	var result fileSlice
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

func (instance fileSlice) prepare() {
	log.Info("Prepare files...")
	start := time.Now()
	for _, file := range instance {
		file.prepare()
	}
	d := time.Now().Sub(start)
	log.
		WithField("duration", d).
		Info("Prepare files... DONE!")
}
