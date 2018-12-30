package server

import (
	"fmt"
	"github.com/blaubaer/goxr"
	"github.com/blaubaer/goxr/common"
	"net/http"
	"os"
	"time"
)

func BoxToHttpFileSystem(box goxr.Box) http.FileSystem {
	return &httpFS{box}
}

type httpFS struct {
	goxr.Box
}

func (instance *httpFS) Open(path string) (http.File, error) {
	return instance.Box.Open(path)
}

func (instance *HttpFileHandler) writeCacheHeadersFor(fi os.FileInfo, to http.ResponseWriter) {
	if efi, ok := fi.(common.ExtendedFileInfo); ok {
		to.Header().Set("Etag", fmt.Sprintf(`"%s"`, efi.ChecksumString()))
	}
	to.Header().Set("Date", fi.ModTime().Truncate(time.Second).Format(time.RFC1123))
}

func bodyAllowedForStatus(status int) bool {
	switch {
	case status >= 100 && status <= 199:
		return false
	case status == 204:
		return false
	case status == 304:
		return false
	}
	return true
}
