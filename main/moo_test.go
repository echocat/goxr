package main

import (
	"github.com/echocat/goxr"
	"github.com/echocat/goxr/common"
	. "github.com/onsi/gomega"
	"math/rand"
	"os"
	"testing"
)

func DisabledTestMoo(t *testing.T) {
	g := NewGomegaWithT(t)

	box, err := goxr.OpenBox("../common", "../log")
	g.Expect(err).To(BeNil())
	g.Expect(box).NotTo(BeNil())

	f, err := os.Create("../resources/testBase2/moo2.bin")
	g.Expect(err).To(BeNil())
	g.Expect(f).NotTo(BeNil())
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
