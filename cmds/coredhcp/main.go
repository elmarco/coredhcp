// Copyright 2018-present the CoreDHCP Authors. All rights reserved
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/coredhcp/coredhcp"
	"github.com/coredhcp/coredhcp/config"
	"github.com/coredhcp/coredhcp/logger"
	"github.com/coredhcp/coredhcp/plugins"
	"github.com/coredhcp/coredhcp/plugins/dns"
	"github.com/coredhcp/coredhcp/plugins/file"
	"github.com/coredhcp/coredhcp/plugins/leasetime"
	"github.com/coredhcp/coredhcp/plugins/nbp"
	"github.com/coredhcp/coredhcp/plugins/netmask"
	rangepl "github.com/coredhcp/coredhcp/plugins/range"
	"github.com/coredhcp/coredhcp/plugins/router"
	"github.com/coredhcp/coredhcp/plugins/serverid"
	"github.com/sirupsen/logrus"
)

var (
	flagLogFile     = flag.String("logfile", "", "Name of the log file to append to. Default: stdout/stderr only")
	flagLogNoStdout = flag.Bool("nostdout", false, "Disable logging to stdout/stderr")
	flagLogLevel    = flag.String("loglevel", "info", fmt.Sprintf("Log level. One of %v", getLogLevels()))
)

var logLevels = map[string]func(*logrus.Logger){
	"none":    func(l *logrus.Logger) { l.SetOutput(ioutil.Discard) },
	"debug":   func(l *logrus.Logger) { l.SetLevel(logrus.DebugLevel) },
	"info":    func(l *logrus.Logger) { l.SetLevel(logrus.InfoLevel) },
	"warning": func(l *logrus.Logger) { l.SetLevel(logrus.WarnLevel) },
	"error":   func(l *logrus.Logger) { l.SetLevel(logrus.ErrorLevel) },
	"fatal":   func(l *logrus.Logger) { l.SetLevel(logrus.FatalLevel) },
}

func getLogLevels() []string {
	var levels []string
	for k := range logLevels {
		levels = append(levels, k)
	}
	return levels
}

var desiredPlugins = []*plugins.Plugin{
	&dns.Plugin,
	&file.Plugin,
	&leasetime.Plugin,
	&nbp.Plugin,
	&netmask.Plugin,
	&rangepl.Plugin,
	&router.Plugin,
	&serverid.Plugin,
}

func main() {
	flag.Parse()
	log := logger.GetLogger("main")
	fn, ok := logLevels[*flagLogLevel]
	if !ok {
		log.Fatalf("Invalid log level '%s'. Valid log levels are %v", *flagLogLevel, getLogLevels())
	}
	fn(log.Logger)
	log.Infof("Setting log level to '%s'", *flagLogLevel)
	if *flagLogFile != "" {
		log.Infof("Logging to file %s", *flagLogFile)
		logger.WithFile(log, *flagLogFile)
	}
	if *flagLogNoStdout {
		log.Infof("Disabling logging to stdout/stderr")
		logger.WithNoStdOutErr(log)
	}
	config, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	// register plugins
	for _, plugin := range desiredPlugins {
		if err := plugins.RegisterPlugin(plugin); err != nil {
			log.Fatalf("Failed to register plugin '%s': %v", plugin.Name, err)
		}
	}

	// start server
	server := coredhcp.NewServer(config)
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	if err := server.Wait(); err != nil {
		log.Error(err)
	}
	time.Sleep(time.Second)
}
