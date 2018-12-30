package common

import (
	"hash/crc64"
)

var (
	crcTable = crc64.MakeTable(crc64.ECMA)
)

func Crc64Of(args ...interface{}) []byte {
	h := crc64.New(crcTable)
	MustWrite(ConcatBytes(args...), h)
	sum := h.Sum(nil)
	return sum
}
