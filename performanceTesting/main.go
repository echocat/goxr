package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"github.com/echocat/goxr/common"
	"github.com/echocat/goxr/log"
	"github.com/shirou/gopsutil/cpu"
	"github.com/urfave/cli"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	cpuCores       int
	cpuInfos       []cpu.InfoStat
	cpuInfosString string
)

func init() {
	var err error
	cpuCores, err = cpu.Counts(true)
	must(err)
	cpuInfos, err = cpu.Info()
	must(err)
	buf := new(bytes.Buffer)
	for i, cpuInfo := range cpuInfos {
		if i != 0 {
			common.MustWritef(buf, "|")
		}
		common.MustWritef(buf, "%d#%s",
			cpuInfo.CPU,
			cpuInfo.ModelName,
		)
	}
	cpuInfosString = buf.String()
}

func main() {
	defer func() {
		r := recover()
		if r != nil {
			_, _ = fmt.Fprintf(os.Stderr, "\n\n\n\n\nERR: %v\n\n\n\n\n\n", r)
			panic(r)
		}
	}()
	app := common.NewApp()
	app.Name = "goxr-performance-testing"
	app.Flags = append(app.Flags,
		cli.DurationFlag{
			Name:        "executionTimePerTest",
			Destination: &executionTimePerTest,
			Value:       executionTimePerTest,
		},
		cli.UintFlag{
			Name:        "numberOfParallelExecutions",
			Destination: &numberOfParallelExecutions,
			Value:       numberOfParallelExecutions,
		},
		cli.GenericFlag{
			Name:  "files",
			Value: &files,
			Usage: fmt.Sprintf(`Could be either one of the static defined files from %s
    or a unit definition like 1K, 1M, ...`, filesDirectory),
		},
		cli.GenericFlag{
			Name:  "processes",
			Value: &processes,
			Usage: fmt.Sprintf(`Defines with which type of process to test - comma separated. (Options are: %s)`, strings.Join(availableProcessTypes(), ", ")),
		},
	)
	app.Action = func(*cli.Context) {
		files.prepare()
		processes.prepare()
		e := createExecutor()

		cw := csv.NewWriter(os.Stdout)
		must(cw.Write([]string{
			"time",
			"os",
			"arch",
			"cpu",
			"cpuCores",
			"process",
			"test",
			"fileSize(b)",
			"runners",
			"executionTime(ns)",
			"executions",
			"ops/s",
			"totalDuration(ns)",
			"minDuration(ns)",
			"avgDuration(ns)",
			"maxDuration(ns)",
			"failures",
			"errors",
			"url",
		}))
		cw.Flush()

		runProcesses(e, cw)
	}

	common.RunApp(app)
}

func runProcesses(e executor, cw *csv.Writer) {
	for _, p := range processes {
		runProcess(e, p, cw)
	}
}

func runProcess(e executor, p process, cw *csv.Writer) {
	start := time.Now()
	log.Infof("Run tests for %s...", p.name())

	p.start()
	defer p.shutdown()
	runFiles(e, p, cw)

	d := time.Now().Sub(start)
	log.
		WithField("duration", d).
		Infof("Run tests for %s... DONE!", p.name())
}

func runFiles(e executor, ta target, cw *csv.Writer) {
	for _, f := range files {
		runTest(e, ta, f, cw)
	}
}

func runTest(e executor, ta target, te test, cw *csv.Writer) {
	start := time.Now()
	log.Infof("    Run test %s for %s...", te.name(), ta.name())

	result := e.execute(te, ta)

	d := time.Now().Sub(start)
	opsPs := float64(result.Executions) / result.TotalDuration.Seconds()
	avgDuration := result.TotalDuration / time.Duration(result.Executions)

	log.
		WithField("duration", d).
		WithField("ops/s", opsPs).
		WithField("avgDuration", avgDuration).
		WithField("executions", result.Executions).
		WithField("failures", result.Failures).
		Infof("    Run test %s for for %s... DONE!", te.name(), ta.name())

	must(cw.Write([]string{
		time.Now().Format(time.RFC3339),
		runtime.GOOS,
		runtime.GOARCH,
		cpuInfosString,
		strconv.FormatInt(int64(cpuCores), 10),
		te.name(),
		strconv.FormatUint(uint64(te.getSize()), 10),
		strconv.FormatUint(uint64(e.numberOfParallelExecutions), 10),
		strconv.FormatInt(int64(e.executionTime), 10),
		strconv.FormatUint(result.Executions, 10),
		strconv.FormatFloat(opsPs, 'f', 10, 64),
		strconv.FormatInt(int64(result.TotalDuration), 10),
		strconv.FormatInt(int64(result.MinDuration), 10),
		strconv.FormatInt(int64(avgDuration), 10),
		strconv.FormatInt(int64(result.MaxDuration), 10),
		strconv.FormatUint(uint64(result.Failures), 10),
		strconv.FormatUint(uint64(result.Errors), 10),
		ta.createUriFor(te),
	}))
	cw.Flush()
}
