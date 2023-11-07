package main

import (
	"github.com/echocat/goxr/common"
	"github.com/echocat/slf4g/native"
	"github.com/echocat/slf4g/native/facade/value"
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

	lv := value.NewProvider(native.DefaultProvider)
	app.Flags = append(app.Flags,
		cli.GenericFlag{
			Name:   "logLevel",
			Usage:  "Specifies the minimum required log level.",
			EnvVar: "GOXR_LOG_LEVEL",
			Value:  &lv.Level,
		},
		cli.GenericFlag{
			Name:   "logFormat",
			Usage:  "Specifies format output (text or json).",
			EnvVar: "GOXR_LOG_FORMAT",
			Value:  &lv.Consumer.Formatter,
		},
		cli.GenericFlag{
			Name:   "logColorMode",
			Usage:  "Specifies if the output is in colors or not (auto, never or always).",
			EnvVar: "GOXR_COLOR_MODE",
			Value:  lv.Consumer.Formatter.ColorMode,
		},
	)
	oldBefore := app.Before
	app.Before = func(context *cli.Context) error {
		if err := oldBefore(context); err != nil {
			return err
		}
		return nil
	}

	common.RunApp(app)
}
