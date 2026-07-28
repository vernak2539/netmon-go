package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/vernak2539/netmon-go/internal/db"
	"github.com/vernak2539/netmon-go/internal/models"
)

func determineStatusText(dlSpeed, ping float64) string {
	if dlSpeed >= 150 && ping <= 20 {
		return "Good speed and low latency"
	}
	if dlSpeed < 60 || ping > 40 {
		return "A bunch of idiots decided to stream 4K movies all at once, or the ISP's mice were busy chewing on the fiber line again, whatever"
	}
	return "At least it works, I guess"
}

func formatMiniReport(m *models.NetworkMetric, deviceCount int) string {
	timestampStr := m.Timestamp.Local().Format("2006-01-02 15:04:05")
	statusText := determineStatusText(m.Download, m.Ping)
	bytesReceivedMB := float64(m.BytesReceived) / 1000000.0
	bytesSentMB := float64(m.BytesSent) / 1000000.0

	return fmt.Sprintf(
		"Network Status Update\nHere is the latest snapshot of your internet speed:\n\nTime: %s\nISP: %s | Server: %s\n\nDevices online: %d\n\nDownload: %.1f Mbps\nUpload: %.1f Mbps\nLatency: %.1f ms\n\nTraffic used: %.1f MB down / %.1f MB up\n\nCurrent status: %s",
		timestampStr,
		m.Client,
		m.Server,
		deviceCount,
		m.Download,
		m.Upload,
		m.Ping,
		bytesReceivedMB,
		bytesSentMB,
		statusText,
	)
}

func main() {
	dbPath := "metrics.sql"
	if len(os.Args) > 1 && os.Args[1] != "" {
		dbPath = os.Args[1]
	} else if envPath := os.Getenv("DB_PATH"); envPath != "" {
		dbPath = envPath
	}

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		log.Fatalf("[db-inspect] Database file '%s' not found.", dbPath)
	}

	database, err := db.Open(dbPath)
	if err != nil {
		log.Fatalf("Opening database: %v", err)
	}
	defer database.Close()

	metrics, deviceCounts, err := database.GetMetricsWithDeviceCounts(context.Background())
	if err != nil {
		log.Fatalf("Querying metrics: %v", err)
	}

	if len(metrics) == 0 {
		fmt.Println("[db-inspect] No speedtest records found in database.")
		return
	}

	latestMetric := metrics[len(metrics)-1]
	latestDevCount := deviceCounts[len(deviceCounts)-1]

	fmt.Println(formatMiniReport(&latestMetric, latestDevCount))
}
