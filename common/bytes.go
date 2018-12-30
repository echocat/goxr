package common

import (
	"encoding/binary"
	"fmt"
	"reflect"
)

func Uint64ToBytes(in uint64) []byte {
	result := make([]byte, 8)
	binary.LittleEndian.PutUint64(result, in)
	return result
}

func Uint32ToBytes(in uint32) []byte {
	result := make([]byte, 4)
	binary.LittleEndian.PutUint32(result, in)
	return result
}

func Uint16ToBytes(in uint16) []byte {
	result := make([]byte, 2)
	binary.LittleEndian.PutUint16(result, in)
	return result
}

func BytesToUint64(in []byte) uint64 {
	return binary.LittleEndian.Uint64(in)
}

func BytesToUint32(in []byte) uint32 {
	return binary.LittleEndian.Uint32(in)
}

func BytesToUint16(in []byte) uint16 {
	return binary.LittleEndian.Uint16(in)
}

func ConcatBytes(args ...interface{}) []byte {
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
		case uint16:
			result = append(result, Uint16ToBytes(v)...)
		case *uint16:
			result = append(result, Uint16ToBytes(*v)...)
		case uint32:
			result = append(result, Uint32ToBytes(v)...)
		case *uint32:
			result = append(result, Uint32ToBytes(*v)...)
		case uint64:
			result = append(result, Uint64ToBytes(v)...)
		case *uint64:
			result = append(result, Uint64ToBytes(*v)...)
		default:
			panic(fmt.Sprintf("Cannot handle type of %v", reflect.TypeOf(arg)))
		}
	}
	return result
}
