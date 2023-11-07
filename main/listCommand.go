package main

import (
	"github.com/echocat/goxr/box/packed"
	"github.com/echocat/goxr/common"
	"github.com/echocat/slf4g"
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

func (instance *ListCommand) ExecuteFromCli(*cli.Context) error {
	return instance.DoWithBox(func(box *packed.Box) error {
		l := log.With("box", instance.Filename)
		l.
			With("name", box.Name).
			With("description", box.Description).
			With("version", box.Version).
			With("revision", box.Revision).
			With("built", box.Built).
			With("builtBy", box.BuiltBy).
			Infof("Entries of %s...", instance.Filename)

		return box.ForEach(instance.FilePredicate, func(info common.FileInfo) error {
			l.Infof("  %-30s (size: %10d, modified: %v, mod: %v)", info.Path(), info.Size(), info.ModTime().Truncate(time.Second), info.Mode())
			return nil
		})
	})
}
