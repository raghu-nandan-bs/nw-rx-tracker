package tui

import (
	"context"
	"os"
	"os/signal"
	"syscall"
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
	window time.Duration, /* width of the TUI time series */
	saveFile string,
) {

	numberOfItems := int(window.Milliseconds() / refreshInterval.Milliseconds())
	rb := newRingBuffer(numberOfItems)

	log.Infof("Starting terminal display")

	if mode == "aggregate" {

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
		outputSubscriber := make(chan *ringBufferByIP)
		consumerCtx := context.Background()
		go consumeByIP(
			outputSubscriber,
			uint64(numberOfItems),
			inputChan,
			consumerCtx,
		)
		if mode == "text" {
			// recieve data from outputSubscriber, dont do anything with it
			go func() {
				for {
					select {
					case <-ctx.Done():
						return
					case rb := <-outputSubscriber:
						log.Tracef("Received ring buffer: %v", rb)
					}
				}
			}()
			log.Tracef("Started consumer process for IP wise aggregation")
			go printTextOutput(inputChan, ctx)
			// Set up signal handling
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

			func() {
				<-sigChan
				log.Info("Received termination signal, canceling context")
				ctx.Done()
				cancelFunc()
			}()
		} else if mode == "tui" {
			log.Tracef("Running TUI with aggregation by IP")
			err := RunTUIWithAggregationByIP(
				outputSubscriber,
				ctx,
				cancelFunc,
				refreshInterval,
			)
			if err != nil {
				log.Fatalf("Error running TUI: %v", err)
			}

		} else {
			log.Fatalf("Unknown display mode %s", mode)
			return
		}
		saveChart(outputSubscriber, saveFile)
		consumerCtx.Done()
	}
	log.Infof("TUI stopped!")

}
