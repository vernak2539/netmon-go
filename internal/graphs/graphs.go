package graphs

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/vernak2539/netmon-go/internal/models"
	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
)

const OutputDir = "graphs"
const OutputFilename = "network_speed_test.png"

// Plot renders network metrics and device counts to a PNG image with dual Y-axes.
func Plot(metrics []models.NetworkMetric, deviceCounts []int) (string, error) {
	if len(metrics) == 0 {
		return "", fmt.Errorf("no metrics provided for plotting")
	}

	xValues := make([]time.Time, len(metrics))
	downloads := make([]float64, len(metrics))
	uploads := make([]float64, len(metrics))
	pings := make([]float64, len(metrics))
	devices := make([]float64, len(metrics))

	for i, m := range metrics {
		xValues[i] = m.Timestamp.Local()
		downloads[i] = m.Download
		uploads[i] = m.Upload
		pings[i] = m.Ping

		if i < len(deviceCounts) {
			devices[i] = float64(deviceCounts[i])
		} else {
			devices[i] = 0
		}
	}

	grayColor := drawing.ColorFromHex("808080")
	gridStyle := chart.Style{
		StrokeColor: drawing.ColorFromHex("D3D3D3"),
		StrokeWidth: 1.0,
	}

	graph := chart.Chart{
		Title: "Network Speed Test Results",
		TitleStyle: chart.Style{
			FontSize:  14,
			FontColor: drawing.ColorBlack,
		},
		XAxis: chart.XAxis{
			Name: "Time",
			ValueFormatter: func(v interface{}) string {
				if typed, isTyped := v.(time.Time); isTyped {
					return typed.Local().Format("02-01 15:04")
				}
				if typed, isTyped := v.(float64); isTyped {
					return time.Unix(0, int64(typed)).Local().Format("02-01 15:04")
				}
				return fmt.Sprintf("%v", v)
			},
			GridMajorStyle: gridStyle,
			Style: chart.Style{
				StrokeColor: grayColor,
			},
		},
		YAxis: chart.YAxis{
			Name:           "Speed (Mbps)",
			GridMajorStyle: gridStyle,
			Style: chart.Style{
				StrokeColor: grayColor,
			},
		},
		YAxisSecondary: chart.YAxis{
			Name: "Ping (ms) / Devices",
			Style: chart.Style{
				StrokeColor: grayColor,
			},
		},
		Series: []chart.Series{
			chart.TimeSeries{
				Name: "Download",
				Style: chart.Style{
					StrokeColor: drawing.ColorBlue,
					StrokeWidth: 2.0,
				},
				XValues: xValues,
				YValues: downloads,
			},
			chart.TimeSeries{
				Name: "Upload",
				Style: chart.Style{
					StrokeColor: drawing.ColorGreen,
					StrokeWidth: 2.0,
				},
				XValues: xValues,
				YValues: uploads,
			},
			chart.TimeSeries{
				Name:  "Ping",
				YAxis: chart.YAxisSecondary,
				Style: chart.Style{
					StrokeColor: drawing.ColorRed,
					StrokeWidth: 2.0,
				},
				XValues: xValues,
				YValues: pings,
			},
			chart.TimeSeries{
				Name:  "Devices",
				YAxis: chart.YAxisSecondary,
				Style: chart.Style{
					StrokeColor: drawing.ColorFromHex("800080"), // Purple
					StrokeWidth: 2.0,
				},
				XValues: xValues,
				YValues: devices,
			},
		},
	}

	// Enable legend
	graph.Elements = []chart.Renderable{
		chart.Legend(&graph),
	}

	if err := os.MkdirAll(OutputDir, 0755); err != nil {
		return "", fmt.Errorf("creating graph directory: %w", err)
	}

	targetPath := filepath.Join(OutputDir, OutputFilename)
	f, err := os.Create(targetPath)
	if err != nil {
		return "", fmt.Errorf("creating output image file: %w", err)
	}
	defer f.Close()

	if err := graph.Render(chart.PNG, f); err != nil {
		return "", fmt.Errorf("rendering graph PNG: %w", err)
	}

	return targetPath, nil
}
