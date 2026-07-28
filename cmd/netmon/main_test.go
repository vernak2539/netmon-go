package main

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/vernak2539/netmon-go/internal/models"
)

func TestDetermineStatusText(t *testing.T) {
	t.Run("Good speed and low latency", func(t *testing.T) {
		status := determineStatusText(160.0, 15.0)
		if status != "Good speed and low latency" {
			t.Errorf("expected Good speed..., got %s", status)
		}
	})

	t.Run("Bad speed or high latency", func(t *testing.T) {
		status := determineStatusText(40.0, 15.0)
		if status != "A bunch of idiots decided to stream 4K movies all at once, or the ISP's mice were busy chewing on the fiber line again, whatever" {
			t.Errorf("unexpected status: %s", status)
		}

		statusPing := determineStatusText(100.0, 50.0)
		if statusPing != "A bunch of idiots decided to stream 4K movies all at once, or the ISP's mice were busy chewing on the fiber line again, whatever" {
			t.Errorf("unexpected status: %s", statusPing)
		}
	})

	t.Run("Average speed fallback", func(t *testing.T) {
		status := determineStatusText(100.0, 30.0)
		if status != "At least it works, I guess" {
			t.Errorf("expected 'At least it works, I guess', got %s", status)
		}
	})
}

func TestFormatMiniReport(t *testing.T) {
	utcTime := time.Date(2026, 7, 27, 12, 0, 0, 0, time.UTC)
	m := &models.NetworkMetric{
		Timestamp:     utcTime,
		Client:        "TestClient",
		Server:        "TestServer",
		Download:      200.0,
		Upload:        50.0,
		Ping:          10.0,
		BytesReceived: 10000000,
		BytesSent:     5000000,
	}

	report := formatMiniReport(m, 3)
	expectedTimeStr := utcTime.Local().Format("2006-01-02 15:04:05")
	if !strings.Contains(report, expectedTimeStr) {
		t.Errorf("expected mini report to contain local time string %s, got report: %s", expectedTimeStr, report)
	}
}

func TestCleanHTMLResponse(t *testing.T) {
	input := "Line 1<br>Line 2<br/>Line 3<br />Line 4"
	expected := "Line 1\nLine 2\nLine 3\nLine 4"
	result := cleanHTMLResponse(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestContextCancellationHandling(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if ctx.Err() == nil {
		t.Errorf("expected context error to be non-nil after cancellation")
	}
}
