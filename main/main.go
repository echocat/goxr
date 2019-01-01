package main

import (
	"github.com/echocat/goxr/common"
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

	common.RunApp(app)
}
