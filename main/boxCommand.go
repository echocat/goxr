package main

import (
	"errors"
	"github.com/blaubaer/goxr/box/packed"
	"github.com/urfave/cli"
)

type BoxCommand struct {
	Filename string
}

func NewBoxCommand() BoxCommand {
	return BoxCommand{}
}

func (instance *BoxCommand) BeforeCli(cli *cli.Context) error {
	if cli.NArg() < 1 {
		return errors.New("too few arguments provided - <box filename> missing")
	}
	instance.Filename = cli.Args()[0]
	return nil
}

func (instance *BoxCommand) CliFlags() []cli.Flag {
	return []cli.Flag{}
}

type DoWithBoxFunc func(*packed.Box) error

func (instance *BoxCommand) DoWithBox(f DoWithBoxFunc) (rErr error) {
	filename := instance.Filename
	if filename == "" {
		return errors.New("no filename provided")
	}
	if box, err := packed.OpenBox(filename); err != nil {
		return err
	} else {
		defer func() {
			if err := box.Close(); err != nil {
				rErr = err
			}
		}()
		return f(box)
	}
}

type DoWithWriterFunc func(*packed.Writer) error

func (instance *BoxCommand) DoWithWriter(w DoWithWriterFunc, om packed.OpenMode, wm packed.WriteMode) (rErr error) {
	filename := instance.Filename
	if filename == "" {
		return errors.New("no filename provided")
	}
	if writer, err := packed.NewWriter(filename, om, wm); err != nil {
		return err
	} else {
		defer func() {
			if err := writer.Close(); err != nil {
				rErr = err
			}
		}()
		return w(writer)
	}
}
