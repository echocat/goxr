package server

import (
	"fmt"
	"github.com/echocat/goxr"
	"github.com/echocat/goxr/box/fs"
	"github.com/echocat/goxr/box/packed"
	"github.com/echocat/goxr/common"
	"github.com/echocat/goxr/runtime"
	"github.com/echocat/goxr/server/configuration"
	"github.com/echocat/slf4g"
	"github.com/urfave/cli"
	"os"
)

type InitiatorPhase func(initiator *Initiator) error

type Initiator struct {
	Server *Server
	App    *cli.App

	Phases []InitiatorPhase
	Fail   func(initiator *Initiator, err error)
}

func NewInitiatorFor(app *cli.App) *Initiator {
	result := &Initiator{
		Server: &Server{
			Logger: log.GetLogger("server"),
		},
		App: app,

		Phases: []InitiatorPhase{
			InitiatorPrepare,
			InitiatorBaseConfigureCli,
			InitiatorPhaseFixLogLevelFlag,
			InitiatorConfigureCliAction,
		},
		Fail: default_Initiator_Fail,
	}

	return result
}

func (instance *Initiator) Execute() {
	for _, phase := range instance.Phases {
		if err := phase(instance); err != nil {
			instance.Fail(instance, err)
			return
		}
	}
}

func InitiatorPrepare(instance *Initiator) error {
	goxr.AllowFallbackToFsBox = false
	if executable, err := runtime.Executable(); err != nil {
		return InitiatorError{err, 127}
	} else if b, err := packed.OpenBox(executable); common.IsDoesNotContainBox(err) {
		instance.Server.Box = nil
		return nil
	} else if err != nil {
		return InitiatorError{err, 127}
	} else if c, err := configuration.OfBox(b); err != nil {
		return InitiatorError{err, 128}
	} else {
		instance.Server.Box = b
		instance.Server.Configuration = c
		return nil
	}
}

func InitiatorBaseConfigureCli(instance *Initiator) error {
	if instance.Server.Box != nil {
		pb := instance.Server.Box.(*packed.Box)
		instance.App.Description = pb.Description
		instance.App.Compiled = pb.Built
		instance.App.Version = pb.ShortString()
	} else {
		instance.App.Description = `Either serves the content of the given [box file] or from a given [base directory].
   If nothing is provided the current work directory is assumed as base directory.`
		instance.App.ArgsUsage = "[box files or base directories]"
		oldBefore := instance.App.Before
		instance.App.Before = func(ctx *cli.Context) error {
			if err := oldBefore(ctx); err != nil {
				return err
			}
			if ctx.NArg() > 0 {
				var cb goxr.CombinedBox
				for _, base := range ctx.Args() {
					if box, err := packed.OpenBox(base); err == nil {
						cb = cb.With(box)
					} else if !common.IsDoesNotContainBox(err) {
						return err
					} else if box, err := fs.OpenBox(base); err != nil {
						return err
					} else {
						cb = cb.With(box)
					}
				}
				instance.Server.Box = cb
			} else {
				if box, err := fs.OpenBox("."); err != nil {
					return err
				} else {
					instance.Server.Box = box
				}
			}
			if c, err := configuration.OfBox(instance.Server.Box); err != nil {
				return err
			} else {
				instance.Server.Configuration = c.Merge(instance.Server.Configuration)
			}
			return nil
		}
	}
	return nil
}

func InitiatorConfigureCliAction(instance *Initiator) error {
	instance.App.Action = func(ctx *cli.Context) error {
		return instance.Server.Run()
	}
	return nil
}

// noinspection GoSnakeCaseUsage
func default_Initiator_Fail(_ *Initiator, err error) {
	ie := InitiatorErrorFor(err)
	common.MustWritef(os.Stderr, "%v\n\n", ie.Cause)
	os.Exit(ie.ExitCode)
}

func InitiatorPhaseFixLogLevelFlag(instance *Initiator) error {
	instance.App.Flags = append(instance.App.Flags, instance.Server.Configuration.Flags()...)
	return nil
}

func InitiatorErrorFor(err error) InitiatorError {
	if ie, ok := err.(InitiatorError); ok {
		return ie
	}
	return InitiatorError{
		Cause:    err,
		ExitCode: 126,
	}
}

type InitiatorError struct {
	Cause    error
	ExitCode int
}

func (instance InitiatorError) Error() string {
	return fmt.Sprintf("[%d] %v", instance.ExitCode, instance.Cause)
}
