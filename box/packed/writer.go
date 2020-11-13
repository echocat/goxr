package packed

import (
	"crypto/sha256"
	"errors"
	"github.com/echocat/goxr/common"
	"github.com/echocat/goxr/entry"
	"github.com/vmihailenco/msgpack"
	_ "github.com/vmihailenco/msgpack"
	"hash"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

func NewWriter(filename string, om OpenMode, wm WriteMode) (writer *Writer, rErr error) {
	of := os.O_RDWR
	if om.IsCreate() {
		of |= os.O_CREATE
	}
	if !om.IsOpen() {
		of |= os.O_EXCL
	}
	if f, err := os.OpenFile(filename, of, 0644); err != nil {
		return nil, common.NewPathError("newWriter", filename, err)
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
			return nil, common.NewPathError("newWriter", filename, err)
		} else if header != nil {
			if !wm.IsReplace() {
				return nil, common.NewPathError("newWriter", filename, common.ErrDoesContainBox)
			} else if err := f.Truncate(int64(header.Offset)); err != nil {
				return nil, common.NewPathError("newWriter", filename, err)
			}
		} else if !wm.IsNew() {
			return nil, common.NewPathError("newWriter", filename, common.ErrDoesNotContainBox)
		}

		if fi, err := f.Stat(); err != nil {
			return nil, common.NewPathError("newWriter", filename, err)
		} else if err := common.Seek(common.FileOffset(fi.Size()), f); err != nil {
			return nil, common.NewPathError("newWriter", filename, err)
		} else if err := WriteHeader(Version(1), 0, f); err != nil {
			return nil, common.NewPathError("newWriter", filename, err)
		} else {
			writer = &Writer{
				f:            f,
				filename:     filename,
				headerOffset: common.FileOffset(fi.Size()),
				offset:       common.FileOffset(fi.Size()) + common.FileOffset(headerLength),
				box: Box{
					Built: time.Now(),
				},
			}
			success = true
			return writer, nil
		}
	}
}

type Writer struct {
	f            *os.File
	filename     string
	headerOffset common.FileOffset
	offset       common.FileOffset

	activeEntryWriter *entryWriter
	box               Box
	closed            bool
}

type TargetEntry struct {
	Filename string
	FileMode *os.FileMode
	Time     *time.Time
	Meta     entry.Meta
}

func (instance *Writer) NewWriter(te TargetEntry) (io.WriteCloser, error) {
	if instance.closed {
		return nil, common.NewPathError("newEntryWriter", te.Filename, io.ErrClosedPipe)
	}
	if instance.activeEntryWriter != nil {
		return nil, common.NewPathError("newEntryWriter", te.Filename, ErrActiveEntryWriter)
	}
	if te.Meta == nil {
		te.Meta = make(entry.Meta)
	}
	te.Filename = entry.CleanPath(te.Filename)

	entryPosition := instance.offset
	e := entry.Entry{
		Filename: te.Filename,
		Offset:   entryPosition,
		Time:     time.Now(),
		FileMode: os.FileMode(0644),
		Meta:     te.Meta,
	}

	if te.FileMode != nil {
		e.FileMode = *te.FileMode
	}
	if te.Time != nil {
		e.Time = *te.Time
	}

	if err := instance.box.Entries.Add(te.Filename, &e); err != nil {
		return nil, common.NewPathError("newEntryWriter", te.Filename, err)
	}

	instance.activeEntryWriter = &entryWriter{
		parent:      instance,
		targetEntry: te,
		hash:        sha256.New(),
	}
	return instance.activeEntryWriter, nil
}

func (instance *Writer) Box() *Box {
	return &instance.box
}

func (instance *Writer) Write(te TargetEntry, source io.Reader) (rErr error) {
	if writer, err := instance.NewWriter(te); err != nil {
		return err
	} else {
		defer func() {
			if dErr := writer.Close(); dErr != nil {
				rErr = dErr
			}
		}()
		if _, err := io.Copy(writer, source); err != nil {
			return common.NewPathError("writeEntry", te.Filename, err)
		}
		return nil
	}
}

func (instance *Writer) WriteFile(sourceFilename string, target TargetEntry) (rErr error) {
	if reader, err := os.OpenFile(sourceFilename, os.O_RDONLY, 0); err != nil {
		return common.NewPathError("writeFileToEntry", sourceFilename, err)
	} else if fi, err := reader.Stat(); err != nil {
		return common.NewPathError("writeFileToEntry", sourceFilename, err)
	} else {
		defer func() {
			if dErr := reader.Close(); dErr != nil {
				rErr = dErr
			}
		}()
		if target.Time == nil {
			target.Time = common.PtimeTime(fi.ModTime())
		}
		if target.FileMode == nil {
			target.FileMode = common.PosFileMode(fi.Mode())
		}
		return instance.Write(target, reader)
	}
}

type WriteCandidate struct {
	Accept bool

	SourceFilename string
	SourceFileInfo os.FileInfo

	Target *TargetEntry
}

type WriteFilesInterceptor func(*WriteCandidate) error

func (instance *Writer) WriteFilesRecursive(root string, interceptor WriteFilesInterceptor) error {
	parts := strings.SplitN(root, "=", 2)
	prefix := ""
	if len(parts) > 1 {
		prefix = path.Clean(filepath.ToSlash(parts[0]))
		if prefix != "" {
			prefix += "/"
		}
		root = parts[1]
	}

	if absRoot, err := filepath.Abs(root); err != nil {
		return err
	} else if err := filepath.Walk(absRoot, func(sourceFilename string, sourceFileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		} else if sourceFileInfo.IsDir() {
			return nil
		} else if sourceFilename, err := filepath.Abs(sourceFilename); err != nil {
			return err
		} else if relativeSourceFilename, err := filepath.Rel(absRoot, sourceFilename); err != nil {
			return err
		} else {
			candidate := WriteCandidate{
				Accept:         true,
				SourceFilename: sourceFilename,
				SourceFileInfo: sourceFileInfo,
				Target: &TargetEntry{
					Filename: entry.CleanPath(path.Join(prefix, filepath.ToSlash(relativeSourceFilename))),
					FileMode: common.PosFileMode(sourceFileInfo.Mode()),
					Time:     common.PtimeTime(sourceFileInfo.ModTime()),
					Meta:     make(entry.Meta),
				},
			}
			if interceptor != nil {
				if err := interceptor(&candidate); err != nil {
					return err
				}
			}
			if !candidate.Accept {
				return nil
			} else {
				return instance.WriteFile(candidate.SourceFilename, *candidate.Target)
			}
		}
	}); err != nil {
		return common.NewPathError("writeFilesRecursive", root, err)
	}
	return nil
}

func (instance *Writer) writeBox() error {
	if err := msgpack.NewEncoder(instance.f).Encode(instance.box); err != nil {
		return common.NewPathError("writeBox", instance.filename, err)
	} else if err := common.Seek(instance.headerOffset, instance.f); err != nil {
		return common.NewPathError("writeBox", instance.filename, err)
	} else if err := WriteHeader(Version(1), instance.offset, instance.f); err != nil {
		return common.NewPathError("writeBox", instance.filename, err)
	} else {
		return nil
	}
}

func (instance *Writer) Close() (rErr error) {
	defer func() {
		if dErr := instance.f.Close(); dErr != nil {
			rErr = dErr
		}
	}()

	return instance.writeBox()
}

type entryWriter struct {
	parent      *Writer
	targetEntry TargetEntry
	hash        hash.Hash
	written     int64

	closed bool
}

func (instance *entryWriter) Write(p []byte) (n int, err error) {
	if instance.closed || instance.parent.closed {
		return 0, common.NewPathError("newEntry", instance.targetEntry.Filename, io.ErrClosedPipe)
	}

	if n, err := instance.hash.Write(p); err != nil {
		return 0, err
	} else if len(p) != n {
		return 0, io.ErrShortWrite
	} else if n, err := instance.parent.f.Write(p); err != nil {
		return 0, err
	} else if len(p) != n {
		return 0, io.ErrShortWrite
	} else {
		instance.written += int64(n)
		return n, nil
	}
}

func (instance *entryWriter) Close() error {
	if instance.closed {
		return nil
	}
	instance.closed = true

	hashArray := entry.Sha256Checksum{}
	copy(hashArray[:], instance.hash.Sum(nil))

	if e, err := instance.parent.box.Entries.Get(instance.targetEntry.Filename); err != nil {
		return common.NewPathError("close", instance.targetEntry.Filename, err)
	} else {
		e.Checksum = hashArray
		e.Length = instance.written
		if err := instance.parent.box.Entries.Replace(instance.targetEntry.Filename, e); err != nil {
			return common.NewPathError("close", instance.targetEntry.Filename, err)
		}
	}

	if instance.parent.activeEntryWriter != instance {
		return common.NewPathError("close", instance.targetEntry.Filename, errors.New("entryWriter is already de-attached from Writer"))
	}
	instance.parent.offset += common.FileOffset(instance.written)
	instance.parent.activeEntryWriter = nil
	return nil
}
