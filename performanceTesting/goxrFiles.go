package main

import "github.com/echocat/goxr/box/fs"

func init() {
	p := &goxrFilesProcess{
		goxrProcess: &goxrProcess{
			server: newGoxrServer(),
		},
	}
	processCandidates[p.name()] = p
	processes = append(processes, p)
}

type goxrFilesProcess struct {
	*goxrProcess
}

func (instance *goxrFilesProcess) name() string {
	return "goxrFiles"
}

func (instance *goxrFilesProcess) prepare() {
	box, err := fs.OpenBox(filesDirectory)
	must(err)
	instance.server.Box = box
}

func (instance *goxrFilesProcess) start() {
	go instance.run()
}

func (instance *goxrFilesProcess) shutdown() {
	instance.shutdownServer()
	instance.cleanup()
}

func (instance *goxrFilesProcess) cleanup() {
	s := instance.server
	if s != nil {
		b := s.Box
		if b != nil {
			close(b)
		}
	}
}
