package main

import (
	"github.com/echocat/goxr/box/packed"
	"github.com/echocat/goxr/common"
	"github.com/urfave/cli"
	"os"
)

var TruncateCommandInstance = NewTruncateCommand()

type TruncateCommand struct {
	BoxCommand

	FailIfFileDoesNotContainBox bool
	FailIfFileDoesNotExist      bool
}

func NewTruncateCommand() *TruncateCommand {
	r := &TruncateCommand{
		BoxCommand: NewBoxCommand(),
	}
	return r
}

func (instance *TruncateCommand) NewCliCommands() []cli.Command {
	return []cli.Command{{
		Name:        "truncate",
		Usage:       "Truncates an existing box from existing file.",
		ArgsUsage:   "<box filename>",
		Before:      instance.BeforeCli,
		Flags:       instance.CliFlags(),
		Action:      instance.ExecuteFromCli,
		Description: `Will truncate a box from a the given <box filename>.`,
	}}
}

func (instance *TruncateCommand) CliFlags() []cli.Flag {
	return append(instance.BoxCommand.CliFlags(),
		cli.BoolTFlag{
			Name: "failIfFileDoesNotContainBox",
			Usage: `If set to true the command will fail if the file exists but does not contain a box
     which could be truncated.`,
			Destination: &instance.FailIfFileDoesNotContainBox,
		},
		cli.BoolTFlag{
			Name:        "failIfFileDoesNotExist",
			Usage:       `If set to true the command will fail if the file does not exists.`,
			Destination: &instance.FailIfFileDoesNotExist,
		},
	)
}

func (instance *TruncateCommand) ExecuteFromCli(ctx *cli.Context) error {
	if err := packed.Truncate(instance.Filename); os.IsNotExist(err) && !instance.FailIfFileDoesNotExist {
		return nil
	} else if common.IsDoesNotContainBox(err) && !instance.FailIfFileDoesNotContainBox {
		return nil
	} else if err != nil {
		return err
	} else {
		return nil
	}
}
