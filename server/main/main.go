package main

import (
	"github.com/blaubaer/goxr"
	"github.com/blaubaer/goxr/box/fs"
	"github.com/blaubaer/goxr/box/packed"
	"github.com/blaubaer/goxr/common"
	"github.com/blaubaer/goxr/log"
	"github.com/blaubaer/goxr/runtime"
	"github.com/blaubaer/goxr/server"
	"github.com/blaubaer/goxr/server/configuration"
	"github.com/urfave/cli"
	"os"
	"strings"
)

var (
	srv         = &server.Server{}
	config      *configuration.Configuration
	httpAddress = ":8080"
	logLevel    = &fixedLogLevel{}
)

func init() {
	goxr.AllowFallbackToFsBox = false
	if executable, err := runtime.Executable(); err != nil {
		fail(126, err)
	} else if b, err := packed.OpenBox(executable); common.IsDoesNotContainBox(err) {
		srv.Box = nil
	} else if err != nil {
		fail(126, err)
	} else if c, err := configuration.OfBox(b); err != nil {
		fail(127, err)
	} else if err := log.SetLevel(c.Logging.GetLevel()); err != nil {
		fail(127, err)
	} else {
		srv.Box = b
		config = &c
	}
}

func main() {
	app := common.NewApp()

	if srv.Box != nil {
		pb := srv.Box.(*packed.Box)
		app.Description = pb.Description
		app.Compiled = pb.Built
		app.Version = pb.ShortString()
	} else {
		app.Description = `Either serves the content of the given [box file] or from a given [base directory].
   If nothing is provided the current work directory is assumed as base directory.`
		app.ArgsUsage = "[box files or base directories]"
		oldBefore := app.Before
		app.Before = func(ctx *cli.Context) error {
			if err := oldBefore(ctx); err != nil {
				return err
			}
			var cb goxr.CombinedBox
			for _, base := range ctx.Args() {
				path := base
				parts := strings.SplitN(base, "=", 2)
				if len(parts) > 1 {
					path = parts[1]
				}
				if fi, err := os.Stat(path); err != nil {
					return err
				} else if fi.IsDir() {
					if box, err := fs.OpenBox(base); err != nil {
						return err
					} else {
						cb = cb.With(box)
					}
				} else {
					if box, err := packed.OpenBox(base); err != nil {
						return err
					} else {
						cb = cb.With(box)
					}
				}
			}
			if len(cb) == 0 {
				if box, err := fs.OpenBox("."); err != nil {
					return err
				} else {
					cb = cb.With(box)
				}
			}
			if c, err := configuration.OfBox(cb); err != nil {
				return err
			} else {
				if logLevel.value == nil {
					logLevel.value = c.Logging.GetLevel()
				}
				srv.Configuration = c
				srv.Box = cb
			}
			return nil
		}
	}
	if config != nil {
		srv.Configuration = *config
		httpAddress = config.Listen.GetHttpAddress()
		logLevel.value = config.Logging.GetLevel()
	}

	app.Flags = append(app.Flags, cli.StringFlag{
		Name:        "httpAddress",
		Value:       httpAddress,
		Destination: &httpAddress,
	})
	fixLogLevelFlag(&app.Flags)

	oldBefore := app.Before
	app.Before = func(ctx *cli.Context) error {
		if err := oldBefore(ctx); err != nil {
			return err
		}
		if logLevel.value != nil {
			if err := log.SetLevel(logLevel.value); err != nil {
				return err
			}
		}
		return nil
	}

	app.Action = func(ctx *cli.Context) error {
		return srv.Run()
	}

	defer func() {
		if srv.Box != nil {
			_ = srv.Box.Close()
		}
	}()
	common.RunApp(app)
}

func fixLogLevelFlag(flags *[]cli.Flag) {
	for i, flag := range *flags {
		if flag.GetName() == "logLevel" {
			gf := flag.(cli.GenericFlag)
			gf.Value = logLevel
			(*flags)[i] = gf
		}
	}
}

func fail(errCode int, arg interface{}) {
	common.MustWritef(os.Stderr, "%v\n\n", arg)
	os.Exit(errCode)
}

type fixedLogLevel struct {
	value log.Level
}

func (instance *fixedLogLevel) Set(plain string) error {
	if plain == "" {
		instance.value = nil
		return nil
	}
	if instance.value == nil {
		instance.value = log.GetLevel()
	}
	return instance.value.Set(plain)
}

func (instance fixedLogLevel) String() string {
	if instance.value == nil {
		return ""
	}
	return instance.value.String()
}
