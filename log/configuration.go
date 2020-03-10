package log

import "github.com/urfave/cli"

type Configuration struct {
	Level        Level     `yaml:"level" json:"level"`
	Format       Format    `yaml:"format" json:"format"`
	ReportCaller bool      `yaml:"reportCaller,omitempty" json:"reportCaller,omitempty"`
	ColorMode    ColorMode `yaml:"colorMode" json:"colorMode" `
}

func (instance Configuration) Merge(with Configuration) Configuration {
	result := instance
	if with.Level != DefaultLevel {
		result.Level = with.Level
	}
	if with.Format != DefaultFormat {
		result.Format = with.Format
	}
	if with.ReportCaller {
		result.ReportCaller = with.ReportCaller
	}
	if with.ColorMode != DefaultColorMode {
		result.ColorMode = with.ColorMode
	}
	return result
}

func (instance Configuration) GetLevel(def Level) Level {
	if result := instance.Level; result != DefaultLevel {
		return result
	}
	return def
}

func (instance Configuration) GetFormat(def Format) Format {
	if result := instance.Format; result != DefaultFormat {
		return result
	}
	return def
}

func (instance Configuration) GetReportCaller() bool {
	return instance.ReportCaller
}

func (instance Configuration) GetColorMode(def ColorMode) ColorMode {
	if result := instance.ColorMode; result != DefaultColorMode {
		return result
	}
	return def
}

func (instance *Configuration) Flags() []cli.Flag {
	return []cli.Flag{
		cli.GenericFlag{
			Name:  "logLevel",
			Usage: "Specifies the minimum required log level.",
			Value: &instance.Level,
		},
		cli.GenericFlag{
			Name:  "logFormat",
			Usage: "Specifies format output (text or json).",
			Value: &instance.Format,
		},
		cli.GenericFlag{
			Name:  "logColorMode",
			Usage: "Specifies if the output is in colors or not (auto, never or always).",
			Value: &instance.ColorMode,
		},
		cli.BoolFlag{
			Name:        "logCaller",
			Usage:       "If true the caller details will be logged too.",
			Destination: &instance.ReportCaller,
		},
	}
}
