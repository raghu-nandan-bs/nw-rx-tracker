package tui

import (
	"context"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/mum4k/termdash/cell"

	nwmbBPF "github.com/raghu-nandan-bs/nw-rx-tracker/pkg/bpf"
	log "github.com/sirupsen/logrus"
)

var globalPos uint64 = 0

type ringBufferByIP struct {
	rbByIP map[string]*ringBuffer
	ips    []string // because there is no map.keys() in go
}

func (r *ringBuffer) ItemsGlobal() []item {
	log.Tracef("ring buffer items: %v", r.items)
	pos := atomic.LoadUint64(&globalPos)
	if pos >= uint64(len(r.items)) {
		pos = pos % uint64(len(r.items))
	}
	return append(r.items[pos:], r.items[:pos]...)
}

func newRingBufferByIP(size int) *ringBufferByIP {
	if size <= 0 {
		panic("invalid size")
	}
	return &ringBufferByIP{
		rbByIP: make(map[string]*ringBuffer, size),
	}
}

func (r *ringBuffer) AddAtPos(i item, pos int) {
	// pos is the position in the ring buffer
	pos = pos % r.cap
	// if the buffer is growing for the first time,
	// and hasnt reached the capacity yet
	// we backfill the buffer with empty items
	if pos >= len(r.items) {
		backfill := make([]item, pos-len(r.items))
		r.items = append(r.items, backfill...)
		r.items = append(r.items, i)
	} else {
		// else we just override the item at the position
		r.items[pos] = i
	}
}

// func timeToMapForSeriesXLabels(t []item) map[int]string {
// 	m := make(map[int]string)
// 	for i, v := range t {
// 		m[i] = v.time.Format("15:04:05")
// 	}
// 	return m
// }

// ---------------------------------------------------------------------
func consumeByIP(
	ouputSubscriber chan *ringBufferByIP,
	resolution uint64,
	inputChan chan nwmbBPF.IngressStatsProcessed,
	ctx context.Context,
) {
	proceededSeries := newRingBufferByIP(int(resolution))
	timeSeriesRB := newRingBuffer(int(resolution))
	for {
		select {
		case <-ctx.Done():
			return
		case stats := <-inputChan:

			now := time.Now()
			timeSeriesRB.Add(item{
				bytes:   0,
				packets: 0,
				time:    now})

			// timeseries rb is used only for populating
			// x axis labels for the line chart
			proceededSeries.rbByIP["timeseries"] = timeSeriesRB
			log.Tracef("Added time series item to ring buffer : %v", timeSeriesRB.Items())

			for ip, rb := range proceededSeries.rbByIP {
				if ip == "timeseries" {
					continue
				}
				if pktStats, ok := stats.BySourceIPv4Addr[ip]; ok {
					rb.AddAtPos(
						item{
							bytes:   pktStats.Bytes,
							packets: pktStats.Packets,
						},
						int(atomic.LoadUint64(&globalPos)),
					)
					rb.expiry = uint64(resolution)

					rb.totalBytes += pktStats.Bytes
					rb.totalPackets += pktStats.Packets

					delete(stats.BySourceIPv4Addr, ip)
				} else if pktStats, ok := stats.BySourceIPv6Addr[ip]; ok {
					rb.AddAtPos(
						item{
							bytes:   pktStats.Bytes,
							packets: pktStats.Packets,
						},
						int(atomic.LoadUint64(&globalPos)),
					)
					rb.expiry = uint64(resolution)

					rb.totalBytes += pktStats.Bytes
					rb.totalPackets += pktStats.Packets

					delete(stats.BySourceIPv6Addr, ip)
				} else {
					rb.AddAtPos(
						item{
							bytes:   0,
							packets: 0,
						},
						int(atomic.LoadUint64(&globalPos)),
					)
				}
				rb.expiry--
				if rb.expiry == 0 {
					delete(proceededSeries.rbByIP, ip)
				}
			}
			log.Tracef("Adding new IPs to the ring buffer")
			for ip, pktStats := range stats.BySourceIPv4Addr {
				if _, ok := proceededSeries.rbByIP[ip]; !ok {
					proceededSeries.rbByIP[ip] = newRingBuffer(int(resolution))
					proceededSeries.rbByIP[ip].expiry = uint64(resolution)
					proceededSeries.rbByIP[ip].color = newRandomColor()
					proceededSeries.rbByIP[ip].AddAtPos(
						item{
							bytes:   pktStats.Bytes,
							packets: pktStats.Packets,
						},
						int(atomic.LoadUint64(&globalPos)),
					)

					proceededSeries.rbByIP[ip].totalBytes = pktStats.Bytes
					proceededSeries.rbByIP[ip].totalPackets = pktStats.Packets

					proceededSeries.rbByIP[ip].lastUpdatedIteration = atomic.LoadUint64(&globalPos)
				}
			}
			log.Tracef(("added ipv4 addresses to the ring buffer"))
			for ip, pktStats := range stats.BySourceIPv6Addr {
				if _, ok := proceededSeries.rbByIP[ip]; !ok {
					proceededSeries.rbByIP[ip] = newRingBuffer(int(resolution))
					proceededSeries.rbByIP[ip].expiry = uint64(resolution)
					proceededSeries.rbByIP[ip].color = newRandomColor()
					proceededSeries.rbByIP[ip].AddAtPos(
						item{
							bytes:   pktStats.Bytes,
							packets: pktStats.Packets,
						},
						int(atomic.LoadUint64(&globalPos)),
					)

					proceededSeries.rbByIP[ip].totalBytes = pktStats.Bytes
					proceededSeries.rbByIP[ip].totalPackets = pktStats.Packets

					proceededSeries.rbByIP[ip].lastUpdatedIteration = atomic.LoadUint64(&globalPos)
				}
			}
			ouputSubscriber <- proceededSeries
			atomic.AddUint64(&globalPos, 1)
		}
	}
}

func newRandomColor() cell.Color {
	return cell.ColorNumber(rand.Intn(255))
}
