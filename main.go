package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	nwmbBPF "github.com/raghu-nandan-bs/nw-rx-tracker/pkg/bpf"
	"github.com/raghu-nandan-bs/nw-rx-tracker/pkg/tui"
	log "github.com/sirupsen/logrus"
)

func main() {
	// manage logging
	f, err := os.OpenFile("nw-rx-tracker.log", os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		fmt.Errorf("unable to open log file: %v", err)
	}
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(
		&log.TextFormatter{
			FullTimestamp: true,
		},
	)
	log.SetOutput(f)

	// Remove resource limits for kernels <5.11.
	if err := nwmbBPF.RemoveMemolock(); err != nil {
		log.Fatalf("Removing memlock: %v", err)
	}
	log.Infof("Removed memlock")

	// loads bpf objects
	if err := nwmbBPF.LoadObjects(); err != nil {
		log.Fatalf("Loading eBPF objects: %v", err)
	}
	defer nwmbBPF.CloseObjects()
	log.Infof("Loaded eBPF objects")

	// link xdp program to the device
	cleanupFunc, err := nwmbBPF.LinkXDPProgramToDevice(deviceName)
	defer cleanupFunc()
	if err != nil {
		log.Fatalf("Attaching XDP: %v", err)
	}
	log.Infof("Attached XDP program to %s", deviceName)

	// listen to signal to stop the program
	stopChan := make(chan os.Signal, 5)
	stopSignalToChildren := make(chan bool, 1)
	signal.Notify(stopChan, os.Interrupt)

	// start user space program to read stats from eBPF
	statsChan := nwmbBPF.TrackIngress(interval, stopSignalToChildren)
	// context is for terminal display control
	ctx, cancel := context.WithCancel(context.Background())
	// start TUI
	tui.RunDisplay(
		statsChan,
		ctx,
		cancel,
		displayMode,
		interval, /*refresh interval*/
		window,   /*width of the TUI time series*/
	)

}
