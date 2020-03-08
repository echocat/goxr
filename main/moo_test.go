package main

import (
	"github.com/echocat/goxr"
	"github.com/echocat/goxr/common"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"os"
	"testing"
)

func DisabledTestMoo(t *testing.T) {
	box, err := goxr.OpenBox("../common", "../log")
	assert.NoError(t, err)
	assert.NotNil(t, box)

	f, err := os.Create("../resources/testBase2/moo2.bin")
	assert.NoError(t, err)
	assert.NotNil(t, f)
	//noinspection GoUnhandledErrorResult
	defer f.Close()

	src := rand.NewSource(1)
	rng := rand.New(src)
	buf := make([]byte, 1024*50)
	for i := 0; i < 1024; i++ {
		common.MustRead(rng, buf)
		common.MustWrite(buf, f)
	}
}
