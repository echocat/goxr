package main

import (
	"github.com/echocat/goxr/box/packed"
	"github.com/echocat/goxr/common"
	"github.com/echocat/goxr/log"
	"github.com/urfave/cli"
	"io"
	"os"
	"regexp"
)

var CatCommandInstance = NewCatCommand()

type CatCommand struct {
	FilteringBoxCommand

	FilenamePatterns []*regexp.Regexp
}

func NewCatCommand() *CatCommand {
	r := &CatCommand{
		FilteringBoxCommand: NewFilteringBoxCommand(),
	}
	return r
}

func (instance *CatCommand) NewCliCommands() []cli.Command {
	return []cli.Command{{
		Name:      "cat",
		Usage:     "Prints the content box entries to stdout.",
		ArgsUsage: instance.ArgsUsage(),
		Before:    instance.BeforeCli,
		Flags:     instance.CliFlags(),
		Action:    instance.ExecuteFromCli,
		Description: `Prints the content of entries of the <box filename> to stdout.

   If [regexp file patterns] provided it will check if at least one of these patterns
   matches the name of the file candidate to be listed.`,
	}}
}

func (instance *CatCommand) ExecuteFromCli(_ *cli.Context) error {
	return instance.DoWithBox(func(box *packed.Box) error {
		l := log.WithField("box", instance.Filename)

		l.
			WithField("name", box.Name).
			WithField("description", box.Description).
			WithField("version", box.Version).
			WithField("revision", box.Revision).
			WithField("built", box.Built).
			WithField("builtBy", box.BuiltBy).
			Infof("Displaying of %s...", instance.Filename)

		return box.ForEach(instance.FilePredicate, func(info common.FileInfo) error {
			l.Infof("  %s", info.Path())
			return instance.displayEntry(info, box)
		})
	})
}

func (instance *CatCommand) displayEntry(info common.FileInfo, box *packed.Box) error {
	f, err := box.Open(info.Path())
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	if _, err := io.Copy(os.Stdout, f); err != nil {
		return err
	}

	return nil
}
