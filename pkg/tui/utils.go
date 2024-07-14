package tui

import (
	"context"
	"fmt"
	"time"

	"github.com/mum4k/termdash/cell"
	nwmbBPF "github.com/raghu-nandan-bs/nw-rx-tracker/pkg/bpf"
	log "github.com/sirupsen/logrus"
)

type item struct {
	bytes   uint64
	packets uint64
	time    time.Time
}

// inspired from https://github.com/surki/network-microburst/blob/main/chart.go#L260C1-L292C2

type ringBuffer struct {
	pos                  int
	items                []item
	lastUpdatedIteration uint64
	cap                  int
	color                cell.Color
	expiry               uint64
	mode                 string
	totalBytes           uint64
	totalPackets         uint64
}

func newRingBuffer(size int) *ringBuffer {
	if size <= 0 {
		panic(fmt.Sprintf("invalid size %d", size))
	}

	return &ringBuffer{
		items: make([]item, size, size),
		cap:   size,
	}
}

func (r *ringBuffer) Add(item item) {
	if r.pos >= len(r.items) {
		r.items = append(r.items, item)
	} else {
		r.items[r.pos] = item
	}
	r.pos = (r.pos + 1) % cap(r.items)
}

func (r *ringBuffer) Len() int {
	return len(r.items)
}

func (r *ringBuffer) Items() []item {
	log.Tracef("ring buffer items: %v", r.items)
	return append(r.items[r.pos:], r.items[:r.pos]...)
}

func timeToMapForSeriesXLabels(t []item) map[int]string {
	m := make(map[int]string)
	for i, v := range t {
		m[i] = v.time.Format("15:04:05")
	}
	return m
}

// ---------------------------------------------------------------------

func ConsumeAggregates(
	rb *ringBuffer,
	inputChan chan nwmbBPF.IngressStatsProcessed,
	ctx context.Context,
) {
	for {
		select {
		case packetStats := <-inputChan:
			rb.Add(item{
				bytes:   packetStats.AggrStats.Bytes,
				packets: packetStats.AggrStats.Packets,
				time:    time.Now(),
			})
			log.Tracef("ring buffer size: %v", rb.Len())

		case <-ctx.Done():
			log.Infof("Context done, ring buffer consumer stopping")
			return
		}
	}
}
