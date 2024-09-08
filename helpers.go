package main

import (
	"flag"
	"time"

	log "github.com/sirupsen/logrus"
)

var deviceName string
var intervalStr string
var windowStr string
var interval time.Duration
var window time.Duration
var displayMode string
var logFile string
var logLevel string
var savePath string

func init() {
	cliArgs()
}

func cliArgs() {
	var err error
	flag.StringVar(&deviceName, "device", "eth0", "network device name")
	flag.StringVar(&intervalStr, "interval", "1s", "interval of refreshing stats")
	flag.StringVar(&windowStr, "window", "30s", "Width of the TUI time series")
	flag.StringVar(&displayMode, "display", "text", "options: text (default) / tui")
	flag.StringVar(&logFile, "log", "/tmp/nwrxtrkr.log", "log file path")
	flag.StringVar(&logLevel, "log-level", "info", "log level")
	flag.StringVar(&savePath, "save", "/tmp/nwrxtrkr.html", "path to save the chart")
	flag.Parse()

	// Parse interval
	interval, err = time.ParseDuration(intervalStr)
	if err != nil {
		log.Error("Unable to parse cli arg (time interval) :%v", err)
		log.Info("Defaulting to duration of 1s")
		interval, _ = time.ParseDuration("1s")
	}

	// Parse window
	window, err = time.ParseDuration(windowStr)
	if err != nil {
		log.Error("Unable to parse cli arg (time window) :%v", err)
		log.Info("Defaulting to duration of 30s")
		window, _ = time.ParseDuration("30s")
	}

	// Parse log level
	parsedLogLevel, err := log.ParseLevel(logLevel)
	if err != nil {
		log.Warnf("Invalid log level: %s. Defaulting to 'info'", logLevel)
		parsedLogLevel = log.InfoLevel
	}
	log.SetLevel(parsedLogLevel)
}
