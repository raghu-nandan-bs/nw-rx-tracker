package tui

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	humanize "github.com/dustin/go-humanize"
	nwmbBPF "github.com/raghu-nandan-bs/nw-rx-tracker/pkg/bpf"
	log "github.com/sirupsen/logrus"
)

// just displays data as text on terminal

func displayAggregates(stat nwmbBPF.IngressStatsProcessed) {
	log.Infof("Total packets: %d", stat.AggrStats.Packets)
	log.Infof("Total bytes: %s", humanize.Bytes(stat.AggrStats.Bytes))
	log.Infof("--------------------------------")
}

func displayAggregatesByIPAsPlainText(
	inputRingbufferByIP chan *ringBufferByIP,
	ctx context.Context,
	cancelFunc func()) {
	for {
		stopChan := make(chan os.Signal, 5)
		signal.Notify(stopChan, os.Interrupt)

		select {
		case <-stopChan:
			cancelFunc()
			return
		case <-ctx.Done():
			return
		case rbByIP := <-inputRingbufferByIP:
			for ip, rb := range rbByIP.rbByIP {
				fmt.Printf("IP: %s, size [%v]\n", ip, rb.Len())
				for _, item := range rb.Items() {
					fmt.Printf("Time: %s, Packets: %d, Bytes: %s\n\n",
						item.time.Format("15:04:05"),
						item.packets,
						humanize.Bytes(item.bytes),
					)
				}

				fmt.Printf("--------------------------------\n")
			}
		}
	}

}
