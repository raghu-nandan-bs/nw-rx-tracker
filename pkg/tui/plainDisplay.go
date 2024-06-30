package tui

import (
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
