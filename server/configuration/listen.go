package configuration

import "github.com/echocat/goxr"

type Listen struct {
	HttpAddress string `yaml:"httpAddress,omitempty"`
}

func (instance Listen) GetHttpAddress() string {
	r := instance.HttpAddress
	if r == "" {
		return ":8080"
	}
	return r
}

func (instance *Listen) Validate(using goxr.Box) (errors []error) {
	return
}
