package tui

import (
	"context"
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

	log "github.com/sirupsen/logrus"
)

func RunTUIWithAggregatedStats(
	rb *ringBuffer,
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

	go populateCharts(
		runContext,
		rb,
		redrawInterval,
		aggrBytesChart,
		aggrPktsChart,
	)

	builder := grid.New()
	builder.Add(
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

func populateCharts(
	ctx context.Context,
	rb *ringBuffer,
	interval time.Duration,
	aggrBytesChart *linechart.LineChart,
	aggrPktsChart *linechart.LineChart) {
	log.Infof("Starting chart population loop with interval %v", interval)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Infof("Context done, stopping chart population")
			return
		case <-ticker.C:
			log.Tracef("Populating charts")
			// Get the data from the ring buffer
			items := rb.Items()
			var bytesData []float64
			var packetsData []float64
			var xLabels []string

			for _, item := range items {
				bytesData = append(bytesData, float64(item.bytes))
				packetsData = append(packetsData, float64(item.packets))
				xLabels = append(xLabels, item.time.Format("15:04:05"))
			}
			log.Tracef("bytesData length: %v", len(bytesData))
			log.Tracef("packetsData length: %v", len(packetsData))
			log.Tracef("xLabels length: %v", len(xLabels))
			// Add the data to the charts
			aggrBytesChart.Series("Bytes", bytesData, linechart.SeriesXLabels(timeToMapForSeriesXLabels(items)))
			aggrPktsChart.Series("Packets", packetsData, linechart.SeriesXLabels(timeToMapForSeriesXLabels(items)))

		}
	}
}
