package main

import (
	"flag"
	"time"

	log "github.com/sirupsen/logrus"
)

var deviceName string
var intervalStr string
var interval time.Duration

func init() {
	cliArgs()

}

func cliArgs() {
	flag.StringVar(&deviceName, "device", "eth0", "network device name")
	flag.StringVar(&intervalStr, "interval", "1s", "interval of refreshing stats")
	flag.Parse()
	refreshRate, err := time.ParseDuration(intervalStr)
	if err != nil {
		log.Error("Unable to parse cli arg (time interval) :%v", err)
		log.Info("Defaulting to duration of 1s")
		refreshRate, _ = time.ParseDuration("1s")
	}
	interval = refreshRate
}
