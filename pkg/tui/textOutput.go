package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/dustin/go-humanize"
	nwmbBPF "github.com/raghu-nandan-bs/nw-rx-tracker/pkg/bpf"
	log "github.com/sirupsen/logrus"
)

func printTextOutput(inputChan chan nwmbBPF.IngressStatsProcessed,
	ctx context.Context) {

	printer := func(input nwmbBPF.IngressStatsProcessed) {
		fmt.Printf("\n--- Inbound Network Statistics at %s ---\n\n", time.Now().Format("2006-01-02 15:04:05.000000"))

		fmt.Println("Inbound IPv4 Traffic:")
		for ipv4, stats := range input.BySourceIPv4Addr {
			fmt.Printf("  %-39s  Bytes: %-10s  Packets: %d\n",
				ipv4,
				humanize.Bytes(uint64(stats.Bytes)),
				stats.Packets)
		}

		fmt.Println("\nInbound IPv6 Traffic:")
		for ipv6, stats := range input.BySourceIPv6Addr {
			fmt.Printf("  %-39s  Bytes: %-10s  Packets: %d\n",
				ipv6,
				humanize.Bytes(uint64(stats.Bytes)),
				stats.Packets)
		}

	}

	for {
		select {
		case <-ctx.Done():
			log.Info("Context done, stopping text output")
			return
		case rb := <-inputChan:
			log.Tracef("Received ring buffer: %v", rb)
			printer(rb)
		}
	}
}
