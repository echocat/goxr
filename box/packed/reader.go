package packed

import (
	"bytes"
	"github.com/echocat/goxr/common"
	"github.com/echocat/goxr/entry"
	"github.com/edsrzf/mmap-go"
	"github.com/vmihailenco/msgpack"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func OpenBox(filename string) (box *Box, rErr error) {
	parts := strings.SplitN(filename, "=", 2)
	prefix := ""
	if len(parts) > 1 {
		prefix = entry.CleanPath(filepath.ToSlash(parts[0]))
		if prefix != "" {
			prefix += "/"
		}
		filename = parts[1]
	}

	if fi, err := os.Stat(filename); err != nil {
		return nil, common.NewPathError("openBox", filename, err)
	} else if fi.IsDir() {
		return nil, common.NewPathError("openBox", filename, common.ErrDoesNotContainBox)
	} else if f, err := os.OpenFile(filename, os.O_RDONLY, 0); err != nil {
		return nil, common.NewPathError("openBox", filename, err)
	} else {
		success := false
		defer func() {
			if !success {
				if dErr := f.Close(); dErr != nil {
					rErr = dErr
				}
			}
		}()
		if header, err := FindHeader(f); err != nil {
			return nil, common.NewPathError("openBox", filename, err)
		} else if header == nil {
			return nil, common.NewPathError("openBox", filename, common.ErrDoesNotContainBox)
		} else if box, err := readBox(filename, f, header.TocOffset); err != nil {
			return nil, err
		} else if m, err := mmap.Map(f, mmap.RDONLY, 0); err != nil {
			return nil, common.NewPathError("openBox", filename, err)
		} else {
			defer func() {
				if !success {
					if dErr := m.Unmap(); dErr != nil {
						rErr = dErr
					}
				}
			}()
			reader := &reader{
				f:        f,
				filename: filename,
				box:      &box,
				mmap:     m,
			}
			box.OnClose = reader.close
			box.EntryToFileTransformer = ToFileTransformerFor(reader.newEntryReader)
			box.Prefix = prefix
			success = true
			return reader.box, nil
		}
	}
}

type reader struct {
	f        *os.File
	filename string
	box      *Box
	mmap     mmap.MMap
}

func (instance *reader) newEntryReader(e entry.Entry) (entry.Reader, error) {
	begin := int(e.Offset)
	end := begin + int(e.Length)
	reader := bytes.NewReader(instance.mmap[begin:end])
	return reader, nil
}

func (instance *reader) close() (rErr error) {
	if err := instance.mmap.Unmap(); err != nil && !strings.Contains(err.Error(), "FlushFileBuffers") {
		rErr = err
	}
	if err := instance.f.Close(); err != nil {
		rErr = err
	}
	return
}

func readBox(filename string, from io.ReadSeeker, tocOffset common.FileOffset) (Box, error) {
	if n, err := from.Seek(int64(tocOffset), 0); err == io.EOF {
		return Box{}, io.EOF
	} else if err != nil {
		return Box{}, common.NewPathError("readBox", filename, err)
	} else if n < int64(tocOffset) {
		return Box{}, common.NewPathError("readBox", filename, io.EOF)
	}
	result := Box{}
	decoder := msgpack.NewDecoder(from)
	if err := decoder.Decode(&result); err != nil {
		return Box{}, common.NewPathError("readBox", filename, err)
	}
	return result, nil
}
