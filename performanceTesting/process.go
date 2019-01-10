package main

import (
	"fmt"
	"github.com/echocat/goxr/log"
	"strings"
	"time"
)

var (
	processCandidates = map[string]process{}
	processes         = processSlice{}
)

type process interface {
	target
	prepare()
	start()
	shutdown()
}

type processSlice []process

func (instance processSlice) String() string {
	return strings.Join(instance.Strings(), ",")
}

func (instance processSlice) Get() interface{} {
	return &instance
}

func (instance processSlice) Strings() []string {
	result := make([]string, len(instance))
	for i, bs := range instance {
		result[i] = bs.name()
	}
	return result
}

func (instance *processSlice) Set(plain string) error {
	var result processSlice
	for _, str := range strings.Split(plain, ",") {
		str = strings.TrimSpace(str)
		if str != "" {
			if candidate, ok := processCandidates[str]; !ok {
				return fmt.Errorf("unknown process type: %s", str)
			} else {
				result = append(result, candidate)
			}
		}
	}
	*instance = result
	return nil
}

func (instance processSlice) prepare() {
	log.Info("Prepare processes...")
	start := time.Now()
	for _, candidate := range instance {
		candidate.prepare()
	}
	d := time.Now().Sub(start)
	log.
		WithField("duration", d).
		Info("Prepare processes... DONE!")
}

func availableProcessTypes() []string {
	result := make([]string, len(processCandidates))
	var i int
	for typ := range processCandidates {
		result[i] = typ
		i++
	}
	return result
}
