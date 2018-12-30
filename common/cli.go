package common

import (
	"github.com/blaubaer/goxr/log"
	"github.com/blaubaer/goxr/runtime"
	"github.com/urfave/cli"
	"os"
	"time"
)

var (
	CliHelpRequested bool
	versionCommand   = cli.Command{
		Name:  "version",
		Usage: "Print the actual version and other useful information.",
		Action: func(ctx *cli.Context) error {
			return Writef(ctx.App.Writer, "%v\n", runtime.GetRuntime())
		},
	}
)

func init() {
	cli.HelpFlag = cli.BoolFlag{
		Name:        "help, h",
		Usage:       "Show help",
		Destination: &CliHelpRequested,
	}

	cli.AppHelpTemplate = `{{.Name}}{{if .VisibleFlags}} [options]{{end}}{{if .VisibleCommands}} [command]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{end}}{{if .UsageText}}

USAGE:
   {{.UsageText}}{{end}}

VERSION:
   {{.Version}}{{if .Description}}

DESCRIPTION:
   {{.Description}}{{end}}{{if len .Authors}}

AUTHOR{{with $length := len .Authors}}{{if ne 1 $length}}S{{end}}{{end}}:
   {{range $index, $author := .Authors}}{{if $index}}
   {{end}}{{$author}}{{end}}{{end}}{{if .VisibleCommands}}

COMMANDS:{{range $index, $category := .VisibleCategories}}{{if $category.Name}}
   {{$category.Name}}:{{end}}{{range $category.VisibleCommands}}
   {{if $category.Name}}   {{end}}{{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}{{end}}{{end}}{{if .VisibleFlags}}

GLOBAL OPTIONS:
   {{range $index, $option := .VisibleFlags}}{{if $index}}
   {{end}}{{$option}}{{end}}{{end}}{{if .Copyright}}

COPYRIGHT:
   {{.Copyright}}{{end}}

`

	cli.CommandHelpTemplate = `{{.HelpName}}{{if .VisibleFlags}} [options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{end}}{{if .UsageText}}

USAGE:
   {{.UsageText}}{{end}}{{if .Category}}

CATEGORY:
   {{.Category}}{{end}}{{if .Description}}

DESCRIPTION:
   {{.Description}}{{end}}{{if .VisibleFlags}}

OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}

`
}

func NewApp() *cli.App {
	r := runtime.GetRuntime()

	result := cli.NewApp()
	result.Version = r.Version
	result.Compiled = r.Built

	result.HideHelp = true
	result.HideVersion = true
	result.Flags = append(result.Flags, log.DefaultLogger.Flags()...)
	result.Flags = append(result.Flags, cli.HelpFlag)

	result.Commands = append(result.Commands, versionCommand)

	result.Writer = os.Stderr
	result.ErrWriter = os.Stderr

	result.Before = func(*cli.Context) error {
		return log.DefaultLogger.Init()
	}

	return result
}

func RunApp(a *cli.App) {
	if err := a.Run(os.Args); err != nil {
		MustWritef(a.ErrWriter, "ERROR: %v\n", err)
		os.Exit(1)
	}
}

type CliTime struct {
	*time.Time
}

func (instance *CliTime) Set(plain string) error {
	if t, err := time.Parse(time.RFC3339, plain); err != nil {
		return err
	} else {
		*instance = CliTime{
			Time: &t,
		}
		return nil
	}
}

func (instance CliTime) String() string {
	t := instance.Time
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}
