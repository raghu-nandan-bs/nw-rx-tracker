package tui

import (
	"context"
	"fmt"
	"sort"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/container/grid"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/linechart"
	"github.com/mum4k/termdash/widgets/text"

	log "github.com/sirupsen/logrus"
)

func RunTUIWithAggregationByIP(
	inputRingbufferByIP chan *ringBufferByIP,
	runContext context.Context,
	cancelFunc func(),
	refreshInterval time.Duration,
) error {
	t, err := tcell.New()
	if err != nil {
		panic(err)
	}
	defer t.Close()

	// @TODO: read this from a parameter
	var redrawInterval = refreshInterval

	// Below two charts are for displaying aggergate metrics
	aggrBytesChart, err := linechart.New(
		linechart.AxesCellOpts(cell.FgColor(cell.ColorRed)),
		linechart.YLabelCellOpts(cell.FgColor(cell.ColorGreen)),
		linechart.XLabelCellOpts(cell.FgColor(cell.ColorCyan)),
		linechart.YAxisFormattedValues(func(v float64) string {
			return humanize.Bytes(uint64(v))
		}),
	)
	aggrPktsChart, err := linechart.New(
		linechart.AxesCellOpts(cell.FgColor(cell.ColorRed)),
		linechart.YLabelCellOpts(cell.FgColor(cell.ColorGreen)),
		linechart.XLabelCellOpts(cell.FgColor(cell.ColorCyan)),
	)

	ipLister, err := text.New()
	if err != nil {
		return fmt.Errorf("text.New: %v", err)
	}

	if err := ipLister.Write("IPs\n"); err != nil {
		return fmt.Errorf("ipLister.Write: %v", err)
	}

	go populateChartsByIP(
		runContext,
		inputRingbufferByIP,
		redrawInterval,
		aggrBytesChart,
		aggrPktsChart,
	)

	go updateIPsLegend(
		runContext,
		inputRingbufferByIP,
		redrawInterval,
		ipLister,
	)

	builder := grid.New()
	builder.Add(
		grid.ColWidthPerc(72,
			grid.RowHeightPerc(47,
				grid.Widget(aggrBytesChart,
					container.Border(linestyle.Light),
					container.BorderTitle("RX - Bytes"),
				),
			),
			grid.RowHeightPerc(47,
				grid.Widget(aggrPktsChart,
					container.Border(linestyle.Light),
					container.BorderTitle("RX -Packets"),
				),
			),
		),
		grid.ColWidthFixed(20,
			grid.Widget(ipLister, container.Border(linestyle.Light)),
		),
	)

	gridOpts, err := builder.Build()
	if err != nil {
		return err
	}

	c, err := container.New(
		t, gridOpts...)
	if err != nil {
		return err
	}

	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == 'q' || k.Key == 'Q' {
			cancelFunc()
		}
	}

	if err := termdash.Run(runContext, t, c, termdash.KeyboardSubscriber(quitter), termdash.RedrawInterval(redrawInterval)); err != nil {
		return err
	}
	return nil

}

func populateChartsByIP(
	ctx context.Context,
	inputRingbufferByIP chan *ringBufferByIP,
	interval time.Duration,
	aggrBytesChart *linechart.LineChart,
	aggrPktsChart *linechart.LineChart) {
	log.Infof("Starting chart population loop with interval %v", interval)

	for {
		select {
		case <-ctx.Done():
			log.Infof("Context done, stopping chart population")
			return
		case rbByIP := <-inputRingbufferByIP:
			log.Tracef("Populating charts : %v", rbByIP)
			// Get the data from the ring buffer
			items := rbByIP.rbByIP["timeseries"].ItemsGlobal()

			log.Tracef("obtained time series items with length: %v", len(items))
			var bytesData []float64
			var packetsData []float64
			var xLabels []string

			for _, item := range items {
				bytesData = append(bytesData, float64(item.bytes))
				packetsData = append(packetsData, float64(item.packets))
				xLabels = append(xLabels, item.time.Format("15:04:05"))
			}

			// Add the data to the charts
			aggrBytesChart.Series("Bytes", bytesData, linechart.SeriesXLabels(timeToMapForSeriesXLabels(items)))
			aggrPktsChart.Series("Packets", packetsData, linechart.SeriesXLabels(timeToMapForSeriesXLabels(items)))
			for ip, rb := range rbByIP.rbByIP {

				if ip == "timeseries" {
					continue
				}

				items = rb.ItemsGlobal()
				bytesData = []float64{}
				packetsData = []float64{}

				log.Tracef("creating series for ip: %v; total points: %v\n", ip, rb.Len())

				for _, item := range items {
					bytesData = append(bytesData, float64(item.bytes))
					packetsData = append(packetsData, float64(item.packets))
				}

				aggrBytesChart.Series(ip, bytesData, linechart.SeriesCellOpts((cell.FgColor(rb.color))))
				aggrPktsChart.Series(ip, packetsData, linechart.SeriesCellOpts((cell.FgColor(rb.color))))
				log.Tracef("Added series for ip: %v", ip)
			}

		}
	}
	log.Infof("Stopping chart population loop")
}

func updateIPsLegend(
	ctx context.Context,
	inputRingbufferByIP chan *ringBufferByIP,
	interval time.Duration,
	ipLister *text.Text) {
	log.Infof("Starting IP lister update loop with interval %v", interval)

	for {
		select {
		case <-ctx.Done():
			log.Infof("Context done, stopping IP lister update")
			return
		case rbByIP := <-inputRingbufferByIP:
			ips := []string{}
			ipLister.Reset()
			log.Tracef("Updating IP lister : %v", rbByIP)
			// Get the data from the ring buffer
			for ip, _ := range rbByIP.rbByIP {

				if ip == "timeseries" {
					continue
				}
				ips = append(ips, ip)
			}
			sort.SliceStable(ips, func(i, j int) bool {
				return rbByIP.rbByIP[ips[i]].totalBytes > rbByIP.rbByIP[ips[j]].totalBytes
			})
			for _, ip := range ips {
				err := ipLister.Write(fmt.Sprintf("%v - %v\n", ip, humanize.Bytes(rbByIP.rbByIP[ip].totalBytes)), text.WriteCellOpts(cell.FgColor(rbByIP.rbByIP[ip].color)))
				if err != nil {
					log.Errorf("Error writing to IP lister: %v", err)
				}
			}
		}
	}
	log.Infof("Stopping IP lister update loop")
}
