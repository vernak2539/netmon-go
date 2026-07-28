package main

import (
	"fmt"
	"strings"

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

func cleanHTMLResponse(text string) string {
	r := strings.ReplaceAll(text, "<br>", "\n")
	r = strings.ReplaceAll(r, "<br/>", "\n")
	r = strings.ReplaceAll(r, "<br />", "\n")
	return r
}

func formatMiniReport(m *models.NetworkMetric, deviceCount int) string {
	timestampStr := m.Timestamp.Local().Format("2006-01-02 15:04:05")
	statusText := determineStatusText(m.Download, m.Ping)
	bytesReceivedMB := float64(m.BytesReceived) / 1000000.0
	bytesSentMB := float64(m.BytesSent) / 1000000.0

	return fmt.Sprintf(
		MiniReportTemplate,
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
