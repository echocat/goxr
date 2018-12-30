package packed

import (
	"bufio"
	"github.com/blaubaer/goxr/common"
	"os"
)

func Truncate(filename string) (rErr error) {
	if f, err := os.OpenFile(filename, os.O_RDWR, 0); err != nil {
		return common.NewPathError("clean", filename, err)
	} else {
		defer func() {
			if err := f.Close(); err != nil {
				rErr = err
			}
		}()
		if header, err := FindHeader(bufio.NewReaderSize(f, HeaderBufferSize)); err != nil {
			return common.NewPathError("clean", filename, err)
		} else if header == nil {
			return nil
		} else if err := f.Truncate(int64(header.Offset)); err != nil {
			return common.NewPathError("clean", filename, err)
		} else {
			return nil
		}
	}
}
