package server

import (
	"fmt"
	"github.com/blaubaer/goxr"
	"github.com/blaubaer/goxr/common"
	"github.com/blaubaer/goxr/log"
	"net/http"
	"os"
	"path"
	"time"
)

func NewHttpFileHandler(box goxr.Box) *HttpFileHandler {
	return &HttpFileHandler{
		Box: box,
	}
}

type HttpFileHandler struct {
	goxr.Box

	Logger log.Logger
}

func (instance *HttpFileHandler) TryServeHTTP(w http.ResponseWriter, r *http.Request) error {
	if r == nil || r.URL == nil {
		return os.ErrNotExist
	}
	if f, err := instance.Box.Open(r.URL.Path); err != nil {
		return err
	} else {
		//noinspection GoUnhandledErrorResult
		defer f.Close()
		if fi, err := f.Stat(); err != nil {
			return err
		} else if fi.IsDir() {
			return os.ErrNotExist
		} else {
			ifNonMatch := r.Header.Get("If-None-Match")
			if efi, ok := fi.(common.ExtendedFileInfo); ok && ifNonMatch == fmt.Sprintf(`"%s"`, efi.ChecksumString()) {
				instance.writeCacheHeadersFor(fi, w)
				w.WriteHeader(http.StatusNotModified)
				return nil
			}
			ifModifiedSince := r.Header.Get("If-Modified-Since")
			if ifModifiedSince != "" {
				if tIfModifiedSince, err := time.Parse(time.RFC1123, ifModifiedSince); err != nil {
					// Ignored
				} else if tIfModifiedSince.Truncate(time.Second).Equal(fi.ModTime().Truncate(time.Second)) {
					instance.writeCacheHeadersFor(fi, w)
					w.WriteHeader(http.StatusNotModified)
				}
				return nil
			}
			p := path.Base(r.URL.Path)
			instance.writeCacheHeadersFor(fi, w)
			http.ServeContent(w, r, p, fi.ModTime(), f)
			return nil
		}
	}
}

func (instance *HttpFileHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	DoServeHTTP(resp, req, instance)
}

func (instance *HttpFileHandler) Log() log.Logger {
	return log.OrDefault(instance.Logger)
}
