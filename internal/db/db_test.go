package db

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/vernak2539/netmon-go/internal/models"
)

func TestDB(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dbtest")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	t.Run("Insert and Retrieve Metrics", func(t *testing.T) {
		m, err := models.NewNetworkMetric(
			100.5,
			50.2,
			12.5,
			"http://share",
			"192.168.1.1",
			"New York",
			1000,
			2000,
		)
		if err != nil {
			t.Fatalf("failed to create metric: %v", err)
		}

		err = db.AddMetric(ctx, m)
		if err != nil {
			t.Fatalf("failed to add metric: %v", err)
		}

		metrics, err := db.GetMetrics(ctx)
		if err != nil {
			t.Fatalf("failed to retrieve metrics: %v", err)
		}

		if len(metrics) != 1 {
			t.Fatalf("expected 1 metric, got %d", len(metrics))
		}

		retrieved := metrics[0]
		if retrieved.Download != m.Download || retrieved.Upload != m.Upload || retrieved.Ping != m.Ping {
			t.Errorf("mismatched values. sent: %+v, got: %+v", m, retrieved)
		}
	})

	t.Run("SpeedTest Join Queries", func(t *testing.T) {
		// Create new Metric
		m, err := models.NewNetworkMetric(120.0, 60.0, 10.0, "", "192.168.1.1", "Atlanta", 5000, 6000)
		if err != nil {
			t.Fatalf("failed to create metric: %v", err)
		}
		if err := db.AddMetric(ctx, m); err != nil {
			t.Fatalf("failed to save metric: %v", err)
		}

		// Create Devices
		d1, _ := models.NewNetworkDevice("192.168.1.100", 5.2)
		d2, _ := models.NewNetworkDevice("192.168.1.101", 10.1)
		devices := []models.NetworkDevice{*d1, *d2}

		scanID, err := db.AddDevices(ctx, devices)
		if err != nil {
			t.Fatalf("failed to add devices: %v", err)
		}

		// Create SpeedTest connection
		st, err := models.NewSpeedTest(m.ID, scanID)
		if err != nil {
			t.Fatalf("failed to create speedtest: %v", err)
		}

		if err := db.AddSpeedTest(ctx, st); err != nil {
			t.Fatalf("failed to add speedtest link: %v", err)
		}

		// Fetch joint metrics
		metrics, counts, err := db.GetMetricsWithDeviceCounts(ctx)
		if err != nil {
			t.Fatalf("failed to query metrics with counts: %v", err)
		}

		if len(metrics) != 1 {
			// Only metrics that have a speedtest record joined will be returned here
			t.Fatalf("expected 1 joint metric, got %d", len(metrics))
		}

		if counts[0] != 2 {
			t.Errorf("expected 2 devices online, got %d", counts[0])
		}

		if metrics[0].ID != m.ID {
			t.Errorf("expected metric ID %s, got %s", m.ID, metrics[0].ID)
		}
	})

	t.Run("Datetime Default Parsing fallback", func(t *testing.T) {
		// Manually insert with a sqlite-native date string using datetime('now') to test fallback parsing
		metricID, _ := models.NewNetworkMetric(90, 45, 15, "", "client", "server", 10, 10)
		_, err := db.db.ExecContext(ctx, `
			INSERT INTO metrics (id, download, upload, ping, share, client, server, bytes_sent, bytes_received, timestamp)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'));
		`, metricID.ID.String(), 90.0, 45.0, 15.0, "", "client", "server", 10, 10)
		if err != nil {
			t.Fatalf("manual insert failed: %v", err)
		}

		metrics, err := db.GetMetrics(ctx)
		if err != nil {
			t.Fatalf("failed to parse fallback metrics: %v", err)
		}

		found := false
		for _, met := range metrics {
			if met.ID == metricID.ID {
				found = true
				if met.Timestamp.IsZero() {
					t.Error("timestamp parsed as zero time")
				}
				break
			}
		}
		if !found {
			t.Error("inserted metric was not found")
		}
	})
}
