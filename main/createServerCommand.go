package main

import (
	"bytes"
	"fmt"
	"github.com/echocat/goxr/box/packed"
	"github.com/echocat/goxr/log"
	"github.com/echocat/goxr/runtime"
	"github.com/urfave/cli"
	"io"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	sr "runtime"
	"strings"
	"text/template"
)

var CreateServerCommandInstance = NewCreateServerCommand()

type CreateServerCommand struct {
	BaseCreateCommand

	Arch        string
	Os          string
	Version     string
	DownloadUrl string
	Overwrite   bool
}

func NewCreateServerCommand() *CreateServerCommand {
	r := &CreateServerCommand{
		BaseCreateCommand: NewBaseCreateCommand(),
	}
	return r
}

func (instance *CreateServerCommand) NewCliCommands() []cli.Command {
	return []cli.Command{{
		Name:      "createServer",
		Usage:     "Creates a standalone server executable which contains a box.",
		ArgsUsage: "<box filename> <name> <version> <description> [[<prefix=>]<path to add>] ...",
		Before:    instance.BeforeCli,
		Flags:     instance.CliFlags(),
		Action:    instance.ExecuteFromCli,
		Description: `Creates a standalone server executable which contains a box inside the given <box filename>.

   It will set the given <name>, <version> and <description> of the created box.
   
   Either there is at least one <path to add> specified to add:
     in this case everything under the specified path will be explicitly added to the box.
   OR there is no [paths to add] specified:
     in this case this command searches in the current working directory for every *.go file
     that contains a goxr.OpenBox(..) or goxr.OpenBoxBy(..) statement and will use its specified
     bases as paths to add to the target box.`,
	}}
}

func (instance *CreateServerCommand) CliFlags() []cli.Flag {
	rt := runtime.GetRuntime()
	return append(instance.BaseCreateCommand.CliFlags(),
		cli.StringFlag{
			Name:        "os, o",
			Usage:       `Defines the operating system for the created server executable.`,
			Value:       rt.GOOS,
			Destination: &instance.Os,
		},
		cli.StringFlag{
			Name:        "arch, a",
			Usage:       `Defines the architecture for the created server executable.`,
			Value:       rt.GOARCH,
			Destination: &instance.Arch,
		},
		cli.StringFlag{
			Name:        "version, v",
			Usage:       `Defines the version for the created server executable.`,
			Value:       rt.Version,
			Destination: &instance.Version,
		},
		cli.StringFlag{
			Name:        "downloadUrl",
			Usage:       `Where to download the .`,
			Value:       `https://github.com/echocat/goxr/releases/download/{{.Version}}/goxr-server-{{.Os}}-{{.Arch}}{{.Os|ext}}`,
			Destination: &instance.DownloadUrl,
		},
		cli.BoolFlag{
			Name:        "overwrite",
			Usage:       `If set to <true> an already existing <box filename> will be overwritten if already exists.`,
			Destination: &instance.Overwrite,
		},
	)
}

func (instance *CreateServerCommand) ExecuteFromCli(ctx *cli.Context) error {
	if err := instance.createServerStub(instance.Filename); err != nil {
		return err
	}
	return instance.DoWithWriter(func(writer *packed.Writer, bases []string) error {
		box := writer.Box()
		l := log.
			WithField("box", instance.Filename)
		l.
			WithField("name", box.Name).
			WithField("description", box.Description).
			WithField("version", box.Version).
			WithField("revision", box.Revision).
			WithField("built", box.Built).
			Infof("Creating server %s...", instance.Filename)

		for _, base := range bases {
			sl := l.WithField("base", base)
			sl.Infof("Adding files of %s...", base)
			if err := writer.WriteFilesRecursive(base, func(candidate *packed.WriteCandidate) error {
				sl.
					WithField("target", candidate.Target.Filename).
					WithField("source", candidate.SourceFilename).
					Infof("  %s", candidate.Target.Filename)
				return nil
			}); err != nil {
				return err
			}
		}
		return nil
	}, packed.OpenModeOpenOnly, packed.WriteModeNewOnly)
}

func (instance *CreateServerCommand) createServerStub(target string) error {
	if sourceFile, err := instance.downloadServerTemplate(); err != nil {
		return err
	} else if fi, err := os.Open(sourceFile); err != nil {
		return err
	} else {
		//noinspection GoUnhandledErrorResult
		defer fi.Close()
		flag := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
		if !instance.Overwrite {
			flag |= os.O_EXCL
		}
		if fo, err := os.OpenFile(target, flag, 0755); err != nil {
			return err
		} else {
			//noinspection GoUnhandledErrorResult
			defer fo.Close()
			if _, err := io.Copy(fo, fi); err != nil {
				return err
			} else {
				return nil
			}
		}
	}
}

func (instance *CreateServerCommand) downloadServerTemplate() (string, error) {
	if targetFile, err := instance.serverTemplateCache(); err != nil {
		return "", err
	} else if fi, err := os.Stat(targetFile); err == nil && !fi.IsDir() && fi.Size() > 0 {
		return targetFile, nil
	} else if err != nil && !os.IsNotExist(err) {
		return "", err
	} else if url, err := instance.downloadUrl(); err != nil {
		return "", err
	} else if targetFile, err := instance.serverTemplateCache(); err != nil {
		return "", err
	} else {
		client := instance.createHttpClient()
		log.Infof("Downloading server template %s...", filepath.Base(targetFile))
		if resp, err := client.Get(url); err != nil {
			return "", err
		} else if resp.StatusCode == 404 {
			return "", fmt.Errorf("the combination of os=%s, arch=%s, version=%s leads to download URL %s of the server executable which does not exist",
				instance.Os, instance.Arch, instance.Version, url)
		} else if resp.StatusCode < 200 || resp.StatusCode >= 400 {
			return "", fmt.Errorf("cannot download server executable from %s: %d - %s", url, resp.StatusCode, resp.Status)
		} else {
			//noinspection GoUnhandledErrorResult
			defer resp.Body.Close()
			if fo, err := os.OpenFile(targetFile, os.O_CREATE|os.O_WRONLY, 0755); err != nil {
				return "", err
			} else {
				//noinspection GoUnhandledErrorResult
				defer fo.Close()
				if size, err := io.Copy(fo, resp.Body); err != nil {
					return "", err
				} else {
					log.Infof("Downloading server template %s... DONE! (size: %d)", filepath.Base(targetFile), size)
					return targetFile, nil
				}
			}
		}
	}
}

func (instance *CreateServerCommand) createHttpClient() *http.Client {
	transport := &http.Transport{}
	transport.RegisterProtocol("file", http.NewFileTransport(&createServerFs{}))
	return &http.Client{
		Transport: transport,
	}
}

type createServerFs struct{}

func (instance *createServerFs) Open(name string) (http.File, error) {
	return os.Open(filepath.FromSlash(name))
}

func (instance *CreateServerCommand) serverTemplateCache() (string, error) {
	base := fmt.Sprintf("goxr-server-%s-%s-%s%s", instance.Os, instance.Arch, instance.Version, instance.ext(instance.Os))
	var dir string
	if lad := os.Getenv("LOCALAPPDATA"); lad != "" && sr.GOOS == "windows" {
		dir = filepath.Join(lad, "goxr", "cache", "server-template")
	} else if u, err := user.Current(); err != nil {
		return "", err
	} else if u.HomeDir != "" {
		dir = filepath.Join(u.HomeDir, ".goxr", "cache", "server-template")
	} else {
		dir = filepath.Join(os.TempDir(), "goxr", "server-template")
	}
	if absDir, err := filepath.Abs(dir); err != nil {
		return "", err
	} else if err := os.MkdirAll(absDir, 0755); err != nil {
		return "", err
	} else if result, err := filepath.Abs(filepath.Join(dir, base)); err != nil {
		return "", err
	} else {
		return result, nil
	}
}

func (instance *CreateServerCommand) downloadUrl() (string, error) {
	buf := new(bytes.Buffer)
	if tmpl, err := template.New("createServer.downloadUrl").
		Funcs(template.FuncMap{
			"ext": instance.ext,
		}).
		Option("missingkey=error").
		Parse(instance.DownloadUrl); err != nil {
		return "", err
	} else if err := tmpl.Execute(buf, instance); err != nil {
		return "", err
	} else {
		return buf.String(), nil
	}
}

func (instance *CreateServerCommand) ext(os string) string {
	switch strings.ToLower(os) {
	case "windows":
		return ".exe"
	default:
		return ""
	}
}
