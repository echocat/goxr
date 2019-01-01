package configuration

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/echocat/goxr"
	"github.com/echocat/goxr/common"
	"gopkg.in/yaml.v2"
	"io"
	"os"
)

const LocationInBox = `goxr-server.yaml`

func OfBox(box goxr.Box) (c Configuration, rErr error) {
	if f, err := box.Open(LocationInBox); os.IsNotExist(err) {
		return Configuration{}, nil
	} else if err != nil {
		return Configuration{}, os.NewSyscallError(fmt.Sprintf(`broken box - read configration of "%s" failed`, LocationInBox), common.UnderlyingError(err))
	} else {
		defer func() {
			if err := f.Close(); err != nil {
				rErr = err
			}
		}()
		if c, err := Read(f); err != nil {
			return Configuration{}, os.NewSyscallError(fmt.Sprintf(`broken box - read configration of "%s" failed`, LocationInBox), common.UnderlyingError(err))
		} else {
			return c, nil
		}
	}
}

func Read(reader io.Reader) (c Configuration, err error) {
	decoder := yaml.NewDecoder(reader)
	err = decoder.Decode(&c)
	return
}

type Configuration struct {
	Listen   Listen   `yaml:"listen,omitempty"`
	Paths    Paths    `yaml:"paths,omitempty"`
	Response Response `yaml:"response,omitempty"`
	Logging  Logging  `yaml:"logging,omitempty"`
}

func (instance *Configuration) Validate(using goxr.Box) (errors []error) {
	errors = append(errors, instance.Listen.Validate(using)...)
	errors = append(errors, instance.Paths.Validate(using)...)
	errors = append(errors, instance.Response.Validate(using)...)
	errors = append(errors, instance.Logging.Validate(using)...)
	return
}

func (instance *Configuration) ValidateAndSummarize(using goxr.Box) error {
	errs := instance.Validate(using)
	if len(errs) <= 0 {
		return nil
	} else if len(errs) == 1 {
		return fmt.Errorf("configuration invalid: %v", errs[0])
	}
	buf := new(bytes.Buffer)
	common.MustWritef(buf, "configuration invalid:")
	for i, err := range errs {
		common.MustWritef(buf, "\n  %d. %v", i+1, err)
	}
	return errors.New(buf.String())
}
