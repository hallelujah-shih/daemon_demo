package main

import (
	"daemon_demo/api"
	"daemon_demo/config"
	"daemon_demo/process"
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
)

var (
	help     bool
	logLevel string
	confPath string
)

func init() {
	const (
		defaultLogLevel = "info"
		defaultConfPath = "config.yaml"
	)
	logrus.SetLevel(logrus.InfoLevel)
	logrus.SetOutput(os.Stdout)

	flag.BoolVar(&help, "h", false, "help")
	flag.StringVar(&logLevel, "log_level", defaultLogLevel, "log level in [debug|info|warn|error]")
	flag.StringVar(&confPath, "c", defaultConfPath, "yaml config path")
	flag.Usage = usage
}

func usage() {
	fmt.Fprintln(os.Stderr, os.Args[0], `
Options:
`)
	flag.PrintDefaults()
}

func main() {
	flag.Parse()
	if help {
		flag.Usage()
		return
	}

	var level logrus.Level
	if err := level.UnmarshalText([]byte(logLevel)); err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)

	cfg, err := config.LoadConfig(confPath)
	if err != nil {
		panic(err)
	}

	pm := process.New()
	for pname, pcfg := range cfg.ProcessConfigs {
		pm.ProcessCreate(pname, pcfg)
	}
	defer pm.Stop()

	apiSvc := api.New(&cfg.MConfig, pm)
	logrus.Errorln("serve:", apiSvc.Serve().ListenAndServe())
}
