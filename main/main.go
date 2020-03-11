package main

import (
	"github.com/echocat/goxr/common"
	"github.com/echocat/goxr/log"
	"github.com/urfave/cli"
)

func main() {
	app := common.NewApp()
	app.Description = `Command line utility of goxr for interacting with boxes.
   See commands section for more details of supported features.`

	app.Commands = append(app.Commands, CatCommandInstance.NewCliCommands()...)
	app.Commands = append(app.Commands, CreateCommandInstance.NewCliCommands()...)
	app.Commands = append(app.Commands, CreateServerCommandInstance.NewCliCommands()...)
	app.Commands = append(app.Commands, ListCommandInstance.NewCliCommands()...)
	app.Commands = append(app.Commands, TruncateCommandInstance.NewCliCommands()...)

	lc := log.Configuration{}
	app.Flags = append(app.Flags, lc.Flags()...)
	oldBefore := app.Before
	app.Before = func(context *cli.Context) error {
		if err := oldBefore(context); err != nil {
			return err
		}
		if err := log.Default.SetConfiguration(lc); err != nil {
			return err
		}
		return nil
	}

	common.RunApp(app)
}
