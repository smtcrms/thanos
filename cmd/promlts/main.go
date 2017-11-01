package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/version"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

func init() {
	prometheus.MustRegister(version.NewCollector("prometheus"))
}

type runFunc func(log.Logger, prometheus.Registerer) error

func main() {
	app := kingpin.New(filepath.Base(os.Args[0]), "A block storage based long-term storage for Prometheus")

	app.Version(version.Print("promlts"))
	app.HelpFlag.Short('h')

	logLevel := app.Flag("log.level", "log filtering level").
		Default("info").Enum("error", "warn", "info", "debug")

	cmds := map[string]runFunc{
		"sidecar": registerSidecar(app, "sidecar"),
	}
	cmd, err := app.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "Error parsing commandline arguments"))
		app.Usage(os.Args[1:])
		os.Exit(2)
	}

	var logger log.Logger
	{
		var lvl level.Option
		switch *logLevel {
		case "error":
			lvl = level.AllowError()
		case "warn":
			lvl = level.AllowWarn()
		case "info":
			lvl = level.AllowInfo()
		case "debug":
			lvl = level.AllowDebug()
		default:
			panic("unexpected log level")
		}
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.NewSyncLogger(logger)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = level.NewFilter(logger, lvl)
	}

	metrics := prometheus.NewRegistry()
	metrics.MustRegister(version.NewCollector("prometheus"))

	if err := cmds[cmd](logger, metrics); err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrap(err, "command failed"))
		os.Exit(1)
	}
}
