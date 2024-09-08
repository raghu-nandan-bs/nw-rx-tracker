package tui

import (
	"fmt"
	"os"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	log "github.com/sirupsen/logrus"
)

func saveChart(outputSubscriber chan *ringBufferByIP, savePath string) {
	log.Info("Saving chart")
	rb := <-outputSubscriber
	log.Infof("Received ring buffer")
	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "Traffic Statistics",
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type:   "slider",
			Start:  0,
			End:    100,
			Orient: "horizontal",
		}),
		charts.WithYAxisOpts(
			opts.YAxis{
				Name: "Bytes",
			},
		),
		charts.WithXAxisOpts(
			opts.XAxis{
				Name: "Time",
			},
		),
		charts.WithLegendOpts(opts.Legend{
			Show:         opts.Bool(true),
			Align:        "right",
			Orient:       "vertical",
			Right:        "0%",
			SelectedMode: "multiple",
			Selected:     make(map[string]bool),
			Type:         "scroll",
		}),
	)

	line.SetSeriesOptions(
		charts.WithLineChartOpts(opts.LineChart{
			Smooth: opts.Bool(true),
		}),
		charts.WithAreaStyleOpts(opts.AreaStyle{
			Opacity: 0.7,
		}),
	)

	timeSeries := make([]string, 0)

	timeSeriesItems := rb.rbByIP["timeseries"].items
	totalCapturedItems := 0
	for _, item := range timeSeriesItems {

		if item.time.Unix() <= 0 {
			break
		}
		totalCapturedItems++
		timeSeries = append(timeSeries, fmt.Sprintf("%v", item.time))
	}
	delete(rb.rbByIP, "timeseries")
	for ip, item := range rb.rbByIP {
		dataSeries := make([]opts.LineData, 0)
		itemsAdded := 0
		for _, i := range item.items {
			if itemsAdded >= totalCapturedItems {
				break
			}
			dataSeries = append(dataSeries, opts.LineData{Value: i.bytes})
			itemsAdded++
		}
		line.AddSeries(ip, dataSeries)
	}

	line.SetXAxis(timeSeries)
	f, err := os.Create(savePath)
	if err != nil {
		log.Errorf("Error creating file: %v", err)
	}
	defer f.Close()

	line.Render(f)
	log.Infof("Saved chart to %s", savePath)
}
