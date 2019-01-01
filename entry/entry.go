package entry

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"github.com/echocat/goxr/common"
	"os"
	"time"
)

//noinspection GoStructTag
type Entry struct {
	_msgpack struct{}          `msgpack:",asArray"`
	BaseName string            // 0
	Offset   common.FileOffset // 1
	Length   int64             // 2
	FileMode os.FileMode       // 3
	Time     time.Time         // 4
	Checksum Sha256Checksum    // 5
	Meta     Meta              // 6
}

func (instance Entry) Name() string {
	return instance.BaseName
}

func (instance Entry) Size() int64 {
	return instance.Length
}

func (instance Entry) Mode() os.FileMode {
	return instance.FileMode
}

func (instance Entry) ModTime() time.Time {
	return instance.Time
}

func (instance Entry) IsDir() bool {
	return false
}

func (instance Entry) Sys() interface{} {
	return nil
}

func (instance Entry) String() string {
	return instance.BaseName
}

func (instance Entry) ChecksumString() string {
	buf := new(bytes.Buffer)
	encoder := base64.NewEncoder(base64.URLEncoding, buf)
	common.MustWrite(instance.Checksum[:], encoder)
	return buf.String()
}

type Sha256Checksum [sha256.Size]byte
type Meta map[string]interface{}

type Predicate func(path string, entry Entry) (bool, error)
