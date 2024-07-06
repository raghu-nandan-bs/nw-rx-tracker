package tui

import (
	"context"
	"time"

	nwmbBPF "github.com/raghu-nandan-bs/nw-rx-tracker/pkg/bpf"
	log "github.com/sirupsen/logrus"
)

/*
Starts running terminal display
input: channel of IngressStatsProcessed, from github.com/raghu-nandan-bs/nw-rx-tracker/pkg/bpf api.go's  IngressStatsProcessed
*/
func RunDisplay(inputChan chan nwmbBPF.IngressStatsProcessed,
	ctx context.Context,
	cancelFunc func(),
	mode string,
	refreshInterval time.Duration,
	window time.Duration /* width of the TUI time series */) {

	numberOfItems := int(window.Milliseconds() / refreshInterval.Milliseconds())
	rb := newRingBuffer(numberOfItems)

	log.Infof("Starting terminal display")

	if mode == "plain" {
		// not implemented
		log.Infof("Plain display not implemented")
	} else if mode == "aggregate" {

		go ConsumeAggregates( // refresh interval not needed,
			// the function just reacts to the input channel
			rb,
			inputChan,
			ctx,
		)
		err := RunTUIWithAggregatedStats(
			rb,
			ctx,
			cancelFunc,
			refreshInterval,
		)
		if err != nil {
			log.Fatalf("Error running TUI: %v", err)
		}
	} else {
		log.Fatalf("Unknown display mode %s", mode)
	}

	log.Infof("TUI stopped!")

}
