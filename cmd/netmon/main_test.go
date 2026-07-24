package main

import (
	"testing"
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
