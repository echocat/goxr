package main

import (
	"github.com/echocat/goxr/common"
	"github.com/echocat/goxr/server"
	"github.com/echocat/slf4g/native"
	"github.com/echocat/slf4g/native/facade/value"
	"github.com/urfave/cli"
)

func main() {
	app := common.NewApp()

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

	initiator := server.NewInitiatorFor(app)
	initiator.Execute()

	defer func() {
		if initiator.Server.Box != nil {
			_ = initiator.Server.Box.Close()
		}
	}()

	common.RunApp(app)
}
