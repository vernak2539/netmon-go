# Chart Rendering Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement `internal/graphs/graphs.go` using pure Go charting (`github.com/wcharczuk/go-chart/v2`) to render 24-hour network metrics & device count trends to a PNG file.

**Architecture:** A `GraphPlotter` struct / `Plot()` function that converts slices of `models.NetworkMetric` and device counts into a dual Y-axis PNG chart (Download & Upload on primary left axis in Mbps; Ping & Devices on secondary right axis), saving to `graphs/network_speed_test.png`.

**Tech Stack:** Go stdlib (`os`, `path/filepath`), `github.com/wcharczuk/go-chart/v2`, `github.com/vernak2539/netmon-go/internal/models`.

## Global Constraints

- Primary Y-axis: Speed in Mbps (Download in blue, Upload in green). Note: Convert metric download/upload from bytes/bits if necessary (in Python it divides by $10^6$ for Mbps).
- Secondary Y-axis: Ping in ms (Red) and Device Count (Purple).
- X-axis: Time formatted as `DD-MM HH:MM` (`02-01 15:04`).
- Save PNG output to `graphs/network_speed_test.png` (creating parent directory `graphs/` if it doesn't exist).
- Return exact file path string.

---

## Proposed File Structure

- Create: `internal/graphs/graphs.go` — PNG chart rendering implementation
- Create: `internal/graphs/graphs_test.go` — Unit test verifying PNG file creation and rendering

---

### Task 1: Implement Pure Go Graph Plotter (`internal/graphs`)

**Files:**
- Create: `internal/graphs/graphs.go`
- Test: `internal/graphs/graphs_test.go`

**Interfaces:**
- Consumes: `models.NetworkMetric`, `deviceCounts []int`
- Produces: `Plot(metrics []models.NetworkMetric, deviceCounts []int) (string, error)`

- [ ] **Step 1: Write the failing test for chart generation**

```go
package graphs

import (
	"os"
	"testing"
	"time"

	"github.com/vernak2539/netmon-go/internal/models"
)

func TestPlot(t *testing.T) {
	now := time.Now().UTC()
	m1, _ := models.NewNetworkMetric(150.0, 50.0, 15.0, "", "Client", "Server", 1000, 2000)
	m1.Timestamp = now.Add(-2 * time.Hour)

	m2, _ := models.NewNetworkMetric(180.0, 60.0, 12.0, "", "Client", "Server", 1000, 2000)
	m2.Timestamp = now.Add(-1 * time.Hour)

	metrics := []models.NetworkMetric{*m1, *m2}
	counts := []int{5, 7}

	outputPath, err := Plot(metrics, counts)
	if err != nil {
		t.Fatalf("unexpected error during plotting: %v", err)
	}

	if outputPath != "graphs/network_speed_test.png" {
		t.Errorf("expected output path graphs/network_speed_test.png, got %s", outputPath)
	}

	info, err := os.Stat(outputPath)
	if err != nil {
		t.Fatalf("output PNG file does not exist: %v", err)
	}

	if info.Size() == 0 {
		t.Errorf("output PNG file is empty")
	}

	// Cleanup generated file after test
	_ = os.Remove(outputPath)
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -v ./internal/graphs/...`
Expected: FAIL with "Plot undefined"

- [ ] **Step 3: Implement `Plot` in `internal/graphs/graphs.go`**

```go
package graphs

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/wcharczuk/go-chart/v2"
	"github.com/wcharczuk/go-chart/v2/drawing"
	"github.com/vernak2539/netmon-go/internal/models"
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
		xValues[i] = m.Timestamp
		// Convert to Mbps if download/upload stored as raw bps, or use direct Mbps values
		downloads[i] = m.Download
		uploads[i] = m.Upload
		pings[i] = m.Ping

		if i < len(deviceCounts) {
			devices[i] = float64(deviceCounts[i])
		} else {
			devices[i] = 0
		}
	}

	graph := chart.Chart{
		Title: "Network Speed Test Results",
		TitleStyle: chart.Style{
			Show:      true,
			FontSize:  14,
			FontColor: drawing.ColorBlack,
		},
		XAxis: chart.XAxis{
			Name:           "Time",
			ValueFormatter: chart.TimeValueFormatterWithFormat("02-01 15:04"),
			Style: chart.Style{
				Show:        true,
				StrokeColor: drawing.ColorGray,
			},
		},
		YAxis: chart.YAxis{
			Name: "Speed (Mbps)",
			Style: chart.Style{
				Show:        true,
				StrokeColor: drawing.ColorGray,
			},
		},
		YAxisSecondary: chart.YAxis{
			Name: "Ping (ms) / Devices",
			Style: chart.Style{
				Show:        true,
				StrokeColor: drawing.ColorGray,
			},
		},
		Series: []chart.Series{
			chart.TimeSeries{
				Name: "Download",
				Style: chart.Style{
					Show:        true,
					StrokeColor: drawing.ColorBlue,
					StrokeWidth: 2.0,
				},
				XValues: xValues,
				YValues: downloads,
			},
			chart.TimeSeries{
				Name: "Upload",
				Style: chart.Style{
					Show:        true,
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
					Show:        true,
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
					Show:        true,
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
```

- [ ] **Step 4: Run test to verify it passes**

Run: `make test`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add go.mod go.sum internal/graphs/
git commit -m "feat: implement PNG chart rendering using go-chart/v2"
```
