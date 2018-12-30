package packed

import (
	"fmt"
	"github.com/blaubaer/goxr/common"
	. "github.com/onsi/gomega"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"reflect"
	"testing"
)

var (
	rng = rand.New(rand.NewSource(666))
)

func Test_writeGarbageBytes(t *testing.T) {
	t.Run("works as expected", func(t *testing.T) {
		g := NewGomegaWithT(t)

		f, err := ioutil.TempFile("", "goxr.box.packed-writeGarbageBytes.*.bin")
		defer deletePathForT(f.Name(), t)
		defer closeForT(f, t)
		g.Expect(err).To(BeNil())
		g.Expect(f).ToNot(BeNil())
		g.Expect(f.Sync()).To(BeNil())
		fi, err := f.Stat()
		g.Expect(err).To(BeNil())
		g.Expect(fi.Size()).To(Equal(int64(0)))

		writeGarbageBytes(garbage(6666), f)
		g.Expect(f.Sync()).To(BeNil())
		fi, err = f.Stat()
		g.Expect(err).To(BeNil())
		g.Expect(fi.Size()).To(Equal(int64(6666)))
	})
}

type garbage int64

func tempFileWithBytesOf(args ...interface{}) string {
	if f, err := ioutil.TempFile("", "goxr-packed-test.*.box"); err != nil {
		panic(err)
	} else {
		defer func() {
			if err := f.Close(); err != nil {
				panic(err)
			}
		}()
		for _, arg := range args {
			switch v := arg.(type) {
			case []byte:
				common.MustWrite(v, f)
			case *[]byte:
				common.MustWrite(*v, f)
			case byte:
				common.MustWrite([]byte{v}, f)
			case *byte:
				common.MustWrite([]byte{*v}, f)
			case Version:
				common.MustWrite([]byte{byte(v)}, f)
			case *Version:
				common.MustWrite([]byte{byte(*v)}, f)
			case int:
				common.MustWrite([]byte{byte(v)}, f)
			case *int:
				common.MustWrite([]byte{byte(*v)}, f)
			case common.FileOffset:
				common.MustWrite(common.Uint64ToBytes(uint64(v)), f)
			case *common.FileOffset:
				common.MustWrite(common.Uint64ToBytes(uint64(*v)), f)
			case garbage:
				writeGarbageBytes(v, f)
			default:
				panic(fmt.Sprintf("Cannot handle type of %v", reflect.TypeOf(arg)))
			}
		}
		return f.Name()
	}
}

func concatBytes(args ...interface{}) []byte {
	var result []byte
	for _, arg := range args {
		switch v := arg.(type) {
		case []byte:
			result = append(result, v...)
		case *[]byte:
			result = append(result, *v...)
		case byte:
			result = append(result, v)
		case *byte:
			result = append(result, *v)
		case Version:
			result = append(result, byte(v))
		case *Version:
			result = append(result, byte(*v))
		case int:
			result = append(result, byte(v))
		case *int:
			result = append(result, byte(*v))
		case common.FileOffset:
			result = append(result, common.Uint64ToBytes(uint64(v))...)
		case *common.FileOffset:
			result = append(result, common.Uint64ToBytes(uint64(*v))...)
		case garbage:
			result = append(result, garbageBytes(v)...)
		default:
			panic(fmt.Sprintf("Cannot handle type of %v", reflect.TypeOf(arg)))
		}
	}
	return result
}

func garbageBytes(amount garbage) []byte {
	result := make([]byte, amount)
	common.MustRead(rng, result)
	return result
}

func writeGarbageBytes(amount garbage, to io.Writer) {
	buf := make([]byte, 4096)
	bufSize := garbage(len(buf))
	var written garbage
	for written < amount {
		target := amount - written
		if target > bufSize {
			target = bufSize
		}
		common.MustRead(rng, buf[:target])
		if n, err := to.Write(buf[:target]); err != nil {
			panic(err)
		} else {
			written += garbage(n)
		}
	}
}

func fileSizeForT(filename string, t *testing.T) int64 {
	if fi, err := os.Stat(filename); err != nil {
		t.Errorf("cannot determine size of '%s': %v", filename, err)
		return 0
	} else {
		return fi.Size()
	}
}

func deletePath(p string) error {
	return os.RemoveAll(p)
}

func deletePathForT(p string, t *testing.T) {
	if err := deletePath(p); err != nil {
		t.Errorf("cannot remove %s: %v", p, err)
	}
}

func closeForT(what io.Closer, t *testing.T) {
	if err := common.Close(what); err != nil {
		t.Errorf("cannot close %v: %v", what, err)
	}
}
