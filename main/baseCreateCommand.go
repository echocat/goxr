package main

import (
	"errors"
	"github.com/echocat/goxr/box/packed"
	"github.com/echocat/goxr/common"
	"github.com/echocat/goxr/runtime"
	"github.com/echocat/goxr/usagescanner"
	"github.com/urfave/cli"
	"os"
	"time"
)

type BaseCreateCommand struct {
	BoxCommand

	Name        string
	Version     string
	Description string
	Build       common.CliTime
	Revision    string
	SourceFiles []string
}

func NewBaseCreateCommand() BaseCreateCommand {
	return BaseCreateCommand{
		BoxCommand: NewBoxCommand(),
		Build:      common.CliTime{},
	}
}

func (instance *BaseCreateCommand) CliFlags() []cli.Flag {
	return append(instance.BoxCommand.CliFlags(),
		cli.GenericFlag{
			Name:  "build, b",
			Usage: "Defines the build timestamp of the created box. If not set the current time will be used.",
			Value: &instance.Build,
		},
		cli.StringFlag{
			Name:        "revision, r",
			Usage:       "Defines the revision of the created box. If not set it will be one created based on the build timestamp.",
			Destination: &instance.Revision,
		},
	)
}

func (instance *BaseCreateCommand) BeforeCli(cli *cli.Context) error {
	if err := instance.BoxCommand.BeforeCli(cli); err != nil {
		return err
	}
	if cli.NArg() < 2 {
		return errors.New("too few arguments provided - <name> missing")
	}
	if cli.NArg() < 3 {
		return errors.New("too few arguments provided - <version> missing")
	}
	if cli.NArg() < 4 {
		return errors.New("too few arguments provided - <description> missing")
	}
	instance.Name = cli.Args()[1]
	instance.Version = cli.Args()[2]
	instance.Description = cli.Args()[3]
	instance.SourceFiles = cli.Args()[4:]
	return nil
}

type DoWithWriterAndBasesFunc func(writer *packed.Writer, bases []string) error

func (instance *BaseCreateCommand) DoWithWriter(f DoWithWriterAndBasesFunc, om packed.OpenMode, wm packed.WriteMode) error {
	return instance.BoxCommand.DoWithWriter(func(writer *packed.Writer) error {
		bases, err := instance.resolveSourceFiles()
		if err != nil {
			return err
		}

		box := writer.Box()
		box.Name = instance.Name
		box.Version = instance.Version
		box.Description = instance.Description
		if instance.Build.Time != nil {
			box.Built = *instance.Build.Time
		} else {
			box.Built = time.Now().Truncate(time.Millisecond)
		}
		if instance.Revision != "" {
			box.Revision = instance.Revision
		} else {
			box.Revision = runtime.RandomRevision(box.Built)
		}
		box.BuiltBy = runtime.GetRuntime().ShortString()

		return f(writer, bases)
	}, om, wm)
}

func (instance *BaseCreateCommand) resolveSourceFiles() ([]string, error) {
	if len(instance.SourceFiles) > 0 {
		return instance.SourceFiles, nil
	} else if cwd, err := os.Getwd(); err != nil {
		return []string{}, err
	} else if usages, err := usagescanner.ScanForUsages(cwd); err != nil {
		return []string{}, err
	} else {
		return usages.Resolve(), nil
	}
}
