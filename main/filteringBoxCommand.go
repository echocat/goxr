package main

import (
	"fmt"
	"github.com/urfave/cli"
	"regexp"
)

type FilteringBoxCommand struct {
	BoxCommand

	FilenamePatterns []*regexp.Regexp
}

func NewFilteringBoxCommand() FilteringBoxCommand {
	return FilteringBoxCommand{
		BoxCommand: NewBoxCommand(),
	}
}

func (instance *FilteringBoxCommand) CliFlags() []cli.Flag {
	return instance.BoxCommand.CliFlags()
}

func (instance *FilteringBoxCommand) ArgsUsage() string {
	return "<box filename> [regexp file patterns]"
}

func (instance *FilteringBoxCommand) BeforeCli(cli *cli.Context) error {
	if err := instance.BoxCommand.BeforeCli(cli); err != nil {
		return err
	}
	instance.FilenamePatterns = make([]*regexp.Regexp, cli.NArg()-1)
	for i, plain := range cli.Args()[1:] {
		if r, err := regexp.Compile(plain); err != nil {
			return fmt.Errorf("illegal [regexp file patterns] %d# provided: %v", i, err)
		} else {
			instance.FilenamePatterns[i] = r
		}
	}
	return nil
}

func (instance *FilteringBoxCommand) FilePredicate(path string) (bool, error) {
	if len(instance.FilenamePatterns) == 0 {
		return true, nil
	}
	for _, pattern := range instance.FilenamePatterns {
		if pattern.FindStringIndex(path) != nil {
			return true, nil
		}
	}
	return false, nil
}
