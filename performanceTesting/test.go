package main

import "github.com/c2h5oh/datasize"

type test interface {
	name() string
	prepare()
	getPath() string
	getSize() datasize.ByteSize
}
