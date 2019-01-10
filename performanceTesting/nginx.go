package main

import (
	"github.com/echocat/goxr/common"
	"github.com/echocat/goxr/log"
	"os"
	"os/exec"
	"sync/atomic"
	"unsafe"
)

func init() {
	p := &nginxProcess{}
	processCandidates[p.name()] = p
	processes = append(processes, p)
}

type nginxProcess struct {
	cmd         unsafe.Pointer
	interrupted unsafe.Pointer
}

func (instance *nginxProcess) name() string {
	return "nginx"
}

func (instance *nginxProcess) String() string {
	return instance.name()
}

func (instance *nginxProcess) prepare() {}

func (instance *nginxProcess) start() {
	executable, err := exec.LookPath("nginx")
	must(err)

	cmd := &exec.Cmd{
		Path:        executable,
		Args:        []string{executable, "-p", rootDirectory, "-c", "nginx/nginx.conf"},
		Stdout:      os.Stdout,
		Stderr:      os.Stderr,
		SysProcAttr: createSysProcAttr(),
	}

	if !atomic.CompareAndSwapPointer(&instance.cmd, nil, unsafe.Pointer(cmd)) {
		panic("one instance of nginx is already running")
	}

	go func(cmd *exec.Cmd) {
		atomic.StorePointer(&instance.interrupted, unsafe.Pointer(common.Pfalse))
		instance.cmd = unsafe.Pointer(cmd)

		err := cmd.Run()
		defer atomic.CompareAndSwapPointer(&instance.cmd, unsafe.Pointer(cmd), nil)
		if !*(*bool)(atomic.LoadPointer(&instance.interrupted)) {
			must(err)
		}
	}(cmd)
}

func (instance *nginxProcess) shutdown() {
	cmd := (*exec.Cmd)(atomic.LoadPointer(&instance.cmd))
	if cmd == nil {
		return
	}
	if atomic.CompareAndSwapPointer(&instance.interrupted, unsafe.Pointer(common.Pfalse), unsafe.Pointer(common.Ptrue)) {
		if err := terminateRecursively(cmd.Process.Pid); err != nil {
			log.WithError(err).Warn("Could not terminate nginx process.")
		}
	}
	_, _ = cmd.Process.Wait()
}

func (instance *nginxProcess) createUriFor(t test) string {
	return "http://localhost:7180/" + t.getPath()
}
