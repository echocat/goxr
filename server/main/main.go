package main

import (
	"github.com/echocat/goxr/common"
	"github.com/echocat/goxr/server"
)

func main() {
	app := common.NewApp()

	initiator := server.NewInitiatorFor(app)
	initiator.Execute()

	defer func() {
		if initiator.Server.Box != nil {
			_ = initiator.Server.Box.Close()
		}
	}()

	common.RunApp(app)
}
