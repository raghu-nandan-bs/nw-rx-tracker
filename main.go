package main

import (
	"os"
	"os/signal"

	nwmbBPF "github.com/raghu-nandan-bs/nw-rx-tracker/pkg/bpf"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(
		&log.TextFormatter{
			FullTimestamp: true,
		},
	)
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

	stopChan := make(chan os.Signal, 5)
	stopSignalToChildren := make(chan bool, 1)
	signal.Notify(stopChan, os.Interrupt)

	statsChan := nwmbBPF.TrackIngress(interval, stopSignalToChildren)
	go func() {
		for {
			select {
			case s := <-statsChan:
				log.Infof("Total bytes: %s\t total packets : %d", s.AggrStats.BytesAsHumanReadableStr(), s.AggrStats.Packets)
				// for k, v := range s.BySourceIPv4Addr {
				// 	log.Infof("Source IP: %s, Bytes: %s, Packets: %d", k, v.BytesAsHumanReadableStr(), v.Packets)
				// }
				// for k, v := range s.BySourceIPv6Addr {
				// 	log.Infof("Source IP: %s, Bytes: %s, Packets: %d", k, v.BytesAsHumanReadableStr(), v.Packets)
				// }
				log.Println("---------------------------------------------------")
			}
		}
	}()

	select {
	case <-stopChan:
		log.Infof("Received interrupt signal, exiting")
		stopSignalToChildren <- true
		return
	}
}
