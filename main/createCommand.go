package main

import (
	"github.com/echocat/goxr/box/packed"
	"github.com/echocat/goxr/log"
	"github.com/urfave/cli"
)

var CreateCommandInstance = NewCreateCommand()

type CreateCommand struct {
	BaseCreateCommand

	OpenMode  packed.OpenMode
	WriteMode packed.WriteMode
}

func NewCreateCommand() *CreateCommand {
	r := &CreateCommand{
		BaseCreateCommand: NewBaseCreateCommand(),
		OpenMode:          packed.OpenModeOpenOnly,
		WriteMode:         packed.WriteModeNewOnly,
	}
	return r
}

func (instance *CreateCommand) NewCliCommands() []cli.Command {
	return []cli.Command{{
		Name:      "create",
		Usage:     "Creates a new box.",
		ArgsUsage: "<box filename> <name> <version> <description> [[<prefix=>]<path to add>] ...",
		Before:    instance.BeforeCli,
		Flags:     instance.CliFlags(),
		Action:    instance.ExecuteFromCli,
		Description: `Will create a box inside the given <box filename>.

   It will set the given <name>, <version> and <description> of the created box.
   
   Either there is at least one <path to add> specified to add:
     in this case everything under the specified path will be explicitly added to the box.
   OR there is no [paths to add] specified:
     in this case this command searches in the current working directory for every *.go file
     that contains a goxr.OpenBox(..) or goxr.OpenBoxBy(..) statement and will use its specified
     bases as paths to add to the target box.`,
	}}
}

func (instance *CreateCommand) CliFlags() []cli.Flag {
	return append(instance.BaseCreateCommand.CliFlags(),
		cli.GenericFlag{
			Name: "openMode, o",
			Usage: `Specifies how to open the <box file>.
     openOrCreate: Open an existing file to write to box to or will create a new one if required.
     openOnly:     Open an existing file or will fail if absent.
     createOnly:   Will create a new file or will fail if already one exists.`,
			Value: &instance.OpenMode,
		},
		cli.GenericFlag{
			Name: "writeMode, w",
			Usage: `Specifies how to write to box to the <box file>.
     newOrReplace: If the file does already contain a box it will be replaced or a new will be added to it.
     replaceOnly:  If the file does already contain a box it will be replaced or the command will fail.
     newOnly:      If the file does not already contain a box it will be added or the command will fail.`,
			Value: &instance.WriteMode,
		},
	)
}

func (instance *CreateCommand) ExecuteFromCli(ctx *cli.Context) error {
	return instance.DoWithWriter(func(writer *packed.Writer, bases []string) error {
		box := writer.Box()
		l := log.
			WithField("box", instance.Filename)
		l.
			WithField("name", box.Name).
			WithField("description", box.Description).
			WithField("version", box.Version).
			WithField("revision", box.Revision).
			WithField("built", box.Built).
			Infof("Creating box %s...", instance.Filename)

		for _, base := range bases {
			sl := l.WithField("base", base)
			sl.Infof("Adding files of %s...", base)
			if err := writer.WriteFilesRecursive(base, func(candidate *packed.WriteCandidate) error {
				sl.
					WithField("target", candidate.Target.Filename).
					WithField("source", candidate.SourceFilename).
					Infof("  %s", candidate.Target.Filename)
				return nil
			}); err != nil {
				return err
			}
		}
		return nil
	}, instance.OpenMode, instance.WriteMode)
}
