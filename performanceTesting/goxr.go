package main

import (
	"github.com/echocat/goxr/box/packed"
	"github.com/echocat/goxr/common"
	"github.com/echocat/goxr/server"
	"github.com/echocat/goxr/server/configuration"
	"io/ioutil"
	"os"
	"sync/atomic"
	"unsafe"
)

func init() {
	p := &goxrProcess{
		server: newGoxrServer(),
	}
	processCandidates[p.name()] = p
	processes = append(processes, p)
}

func newGoxrServer() *server.Server {
	return &server.Server{
		Configuration: configuration.Configuration{
			Listen: configuration.Listen{
				HttpAddress: "localhost:7180",
			},
			Paths: configuration.Paths{
				Catchall: configuration.Catchall{
					Target: common.Pstring(""),
				},
				Index: common.Pstring(""),
			},
			Logging: configuration.Logging{
				AccessLog: common.Pfalse,
			},
		},
	}
}

type goxrProcess struct {
	server *server.Server
	box    unsafe.Pointer
}

type goxrBoxReference struct {
	box  *packed.Box
	file string
}

func (instance goxrBoxReference) Close() error {
	var err error
	if cErr := instance.box.Close(); cErr != nil {
		err = cErr
	}
	_ = os.RemoveAll(instance.file)
	return err
}

func (instance *goxrProcess) name() string {
	return "goxr"
}

func (instance *goxrProcess) String() string {
	return instance.name()
}

func (instance *goxrProcess) prepare() {
	success := false
	f, err := ioutil.TempFile("", "goxr-performanceTest-*.box")
	must(err)
	close(f)
	fn := f.Name()

	defer func() {
		if !success {
			remove(fn)
		}
	}()

	writer, err := packed.NewWriter(fn, packed.OpenModeOpenOnly, packed.WriteModeNewOnly)
	must(err)
	defer close(writer)

	must(writer.WriteFilesRecursive(filesDirectory, nil))
	close(writer)

	box, err := packed.OpenBox(fn)
	must(err)

	ref := &goxrBoxReference{
		box:  box,
		file: fn,
	}

	for !atomic.CompareAndSwapPointer(&instance.box, nil, unsafe.Pointer(ref)) {
		ofn := (*goxrBoxReference)(atomic.LoadPointer(&instance.box))
		close(ofn)
		atomic.CompareAndSwapPointer(&instance.box, unsafe.Pointer(ofn), nil)
	}

	instance.server.Box = ref.box

	success = true
}

func (instance *goxrProcess) start() {
	go instance.run()
}

func (instance *goxrProcess) run() {
	must(instance.server.Run())
}

func (instance *goxrProcess) shutdown() {
	instance.shutdownServer()
	instance.cleanup()
}

func (instance *goxrProcess) shutdownServer() {
	must(instance.server.Shutdown())
}

func (instance *goxrProcess) cleanup() {
	ref := (*goxrBoxReference)(atomic.LoadPointer(&instance.box))
	if ref != nil {
		close(ref)
	}
	atomic.CompareAndSwapPointer(&instance.box, unsafe.Pointer(ref), nil)
}

func (instance *goxrProcess) createUriFor(t test) string {
	return "http://localhost:7180/" + t.getPath()
}
