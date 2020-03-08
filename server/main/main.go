package main

import (
	"github.com/echocat/goxr"
	"github.com/echocat/goxr/box/fs"
	"github.com/echocat/goxr/box/packed"
	"github.com/echocat/goxr/common"
	"github.com/echocat/goxr/log"
	"github.com/echocat/goxr/runtime"
	"github.com/echocat/goxr/server"
	"github.com/echocat/goxr/server/configuration"
	"github.com/urfave/cli"
	"os"
)

var (
	srv         = &server.Server{}
	httpAddress = &fixedString{def: ":8080"}
	logLevel    = &fixedLogLevel{}
	logFormat   = &fixedLogFormat{}
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
	} else if err := log.SetFormat(c.Logging.GetFormat()); err != nil {
		fail(127, err)
	} else {
		srv.Box = b
		srv.Configuration = c
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
				srv.Box = cb
			} else {
				if box, err := fs.OpenBox("."); err != nil {
					return err
				} else {
					srv.Box = box
				}
			}
			if c, err := configuration.OfBox(srv.Box); err != nil {
				return err
			} else {
				srv.Configuration = c
			}
			return nil
		}
	}
	app.Flags = append(app.Flags, cli.GenericFlag{
		Name:  "httpAddress",
		Value: httpAddress,
	})
	fixLogLevelFlag(&app.Flags)

	oldBefore := app.Before
	app.Before = func(ctx *cli.Context) error {
		if err := oldBefore(ctx); err != nil {
			return err
		}
		if err := log.SetLevel(logLevel.evaluate(srv.Configuration.Logging.GetLevel())); err != nil {
			return err
		}
		if err := log.SetFormat(logFormat.evaluate(srv.Configuration.Logging.GetFormat())); err != nil {
			return err
		}
		srv.Configuration.Listen.HttpAddress = httpAddress.evaluate(srv.Configuration.Listen.HttpAddress)
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

func (instance fixedLogLevel) evaluate(fromConfig log.Level) log.Level {
	v := instance.value
	if v != nil {
		return v
	}
	if fromConfig != nil {
		return fromConfig
	}
	return log.GetLevel()
}

type fixedLogFormat struct {
	value log.Format
}

func (instance *fixedLogFormat) Set(plain string) error {
	if plain == "" {
		instance.value = nil
		return nil
	}
	if instance.value == nil {
		instance.value = log.GetFormat()
	}
	return instance.value.Set(plain)
}

func (instance fixedLogFormat) String() string {
	if instance.value == nil {
		return ""
	}
	return instance.value.String()
}

func (instance fixedLogFormat) evaluate(fromConfig log.Format) log.Format {
	v := instance.value
	if v != nil {
		return v
	}
	if fromConfig != nil {
		return fromConfig
	}
	return log.GetFormat()
}

type fixedString struct {
	def   string
	value *string
}

func (instance *fixedString) Set(plain string) error {
	if plain == "" {
		instance.value = nil
		return nil
	}
	instance.value = &plain
	return nil
}

func (instance fixedString) String() string {
	if instance.value == nil {
		return ""
	}
	return *instance.value
}

func (instance fixedString) evaluate(fromConfig string) string {
	v := instance.value
	if v != nil {
		return *v
	}
	if fromConfig != "" {
		return fromConfig
	}
	return instance.def
}
