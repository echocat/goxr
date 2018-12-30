package main

import (
	"github.com/blaubaer/goxr/box/packed"
	"github.com/blaubaer/goxr/log"
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

func (instance *CatCommand) ExecuteFromCli(ctx *cli.Context) error {
	return instance.DoWithBox(func(box *packed.Box) error {
		l := log.WithField("box", instance.Filename)
		if entries, err := box.Entries.Filter(instance.EntryPredicate); err != nil {
			return err
		} else {
			l.
				WithField("name", box.Name).
				WithField("description", box.Description).
				WithField("version", box.Version).
				WithField("revision", box.Revision).
				WithField("built", box.Built).
				WithField("builtBy", box.BuiltBy).
				Infof("Displaying entries of %s...", instance.Filename)

			for path := range entries {
				l.Infof("  %s", path)
				if f, err := box.Open(path); err != nil {
					return err
				} else if _, err := io.Copy(os.Stdout, f); err != nil {
					return err
				}
			}
			return nil
		}
	})
}
