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

func init() {
	cliArgs()
}

func cliArgs() {
	var err error
	flag.StringVar(&deviceName, "device", "eth0", "network device name")
	flag.StringVar(&intervalStr, "interval", "1s", "interval of refreshing stats")
	flag.StringVar(&windowStr, "window", "30s", "Width of the TUI time series")
	flag.StringVar(&displayMode, "display", "plain", "options: plain / aggregate / by-ip")
	flag.Parse()
	interval, err = time.ParseDuration(intervalStr)
	if err != nil {
		log.Error("Unable to parse cli arg (time interval) :%v", err)
		log.Info("Defaulting to duration of 1s")
		interval, _ = time.ParseDuration("1s")
	}
	window, err = time.ParseDuration(windowStr)
	if err != nil {
		log.Error("Unable to parse cli arg (time window) :%v", err)
		log.Info("Defaulting to duration of 30s")
		window, _ = time.ParseDuration("30s")
	}
}
