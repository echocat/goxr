package configuration

import (
	"github.com/echocat/goxr"
	"github.com/urfave/cli"
)

type Listen struct {
	HttpAddress HttpAddress `yaml:"httpAddress,omitempty"`
}

func (instance Listen) GetHttpAddress() string {
	return instance.HttpAddress.String()
}

func (instance *Listen) Validate(using goxr.Box) (errors []error) {
	return
}

type HttpAddress string

func (instance *HttpAddress) Set(plain string) error {
	*instance = HttpAddress(plain)
	return nil
}

func (instance HttpAddress) String() string {
	if instance == "" {
		return ":8080"
	}
	return string(instance)
}

func (instance Listen) Merge(with Listen) Listen {
	result := instance
	if with.HttpAddress != "" {
		result.HttpAddress = with.HttpAddress
	}
	return result
}

func (instance *Listen) Flags() []cli.Flag {
	return []cli.Flag{
		cli.GenericFlag{
			Name:  "httpAddress",
			Usage: "Address where to listen to.",
			Value: &instance.HttpAddress,
		},
	}
}
