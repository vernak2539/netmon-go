package graphs

import (
	"os"
	"testing"
	"time"

	"github.com/vernak2539/netmon-go/internal/models"
)

func TestPlot(t *testing.T) {
	t.Run("Empty metrics error", func(t *testing.T) {
		_, err := Plot(nil, nil)
		if err == nil {
			t.Error("expected error for empty metrics, got nil")
		}
	})

	t.Run("Valid chart generation", func(t *testing.T) {
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

		// Clean up generated file after test
		_ = os.Remove(outputPath)
	})
}
