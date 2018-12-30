package main

import (
	"github.com/blaubaer/goxr/box/packed"
	"github.com/blaubaer/goxr/log"
	"github.com/urfave/cli"
	"regexp"
	"time"
)

var ListCommandInstance = NewListCommand()

type ListCommand struct {
	FilteringBoxCommand

	FilenamePatterns []*regexp.Regexp
}

func NewListCommand() *ListCommand {
	r := &ListCommand{
		FilteringBoxCommand: NewFilteringBoxCommand(),
	}
	return r
}

func (instance *ListCommand) NewCliCommands() []cli.Command {
	return []cli.Command{{
		Name:      "list",
		Usage:     "Lists the content of a box.",
		ArgsUsage: instance.ArgsUsage(),
		Before:    instance.BeforeCli,
		Flags:     instance.CliFlags(),
		Action:    instance.ExecuteFromCli,
		Description: `List the contents of the given <box filename>.

   If [regexp file patterns] provided it will check if at least one of these patterns
   matches the name of the file candidate to be listed.`,
	}}
}

func (instance *ListCommand) ExecuteFromCli(ctx *cli.Context) error {
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
				Infof("Entries of %s...", instance.Filename)

			for path, entry := range entries {
				l.Infof("  %-30s (size: %10d, modified: %v, mod: %v)", path, entry.Size(), entry.ModTime().Truncate(time.Second), entry.Mode())
			}
			return nil
		}
	})
}
