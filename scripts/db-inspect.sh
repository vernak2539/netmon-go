#!/usr/bin/env bash
set -euo pipefail

# Determine DB path and Scan ID
DB="${1:-${DB_PATH:-metrics.sql}}"
SCAN_ID="${2:-${SCAN_ID:-}}"

if [ ! -f "$DB" ]; then
    echo "[db-inspect] Database file '$DB' not found."
    exit 1
fi

if [ -n "$SCAN_ID" ]; then
    echo "=== Inspecting Device Scan ID: $SCAN_ID in $DB ==="
    if command -v sqlite3 >/dev/null 2>&1; then
        sqlite3 -header -column "$DB" "
        SELECT id, ips, latencies FROM device_scans WHERE id = '$SCAN_ID';
        "
    else
        go run -e "
package main
import (
	\"context\"
	\"fmt\"
	\"log\"
	\"github.com/vernak2539/netmon-go/internal/db\"
)
func main() {
	database, err := db.Open(\"$DB\")
	if err != nil { log.Fatal(err) }
	defer database.Close()
	fmt.Println(\"Scan ID lookup: $SCAN_ID\")
}"
    fi
else
    echo "=== Inspecting Recent Speedtests: $DB ==="
    if command -v sqlite3 >/dev/null 2>&1; then
        sqlite3 -header -column "$DB" "
        SELECT 
            m.timestamp, 
            round(m.download, 1) AS dl_mbps, 
            round(m.upload, 1) AS ul_mbps, 
            m.ping, 
            m.client AS isp,
            m.server,
            st.device_scans_id AS device_scan_id
        FROM speedtest st
        JOIN metrics m ON st.metrics_id = m.id
        ORDER BY m.timestamp DESC 
        LIMIT 10;
        "
    else
        go run -e "
package main
import (
	\"context\"
	\"fmt\"
	\"log\"
	\"github.com/vernak2539/netmon-go/internal/db\"
)
func main() {
	database, err := db.Open(\"$DB\")
	if err != nil { log.Fatal(err) }
	defer database.Close()

	metrics, deviceCounts, err := database.GetMetricsWithDeviceCounts(context.Background())
	if err != nil { log.Fatal(err) }

	fmt.Printf(\"%-20s %-10s %-10s %-8s %-15s %-18s %-8s\n\", \"TIMESTAMP\", \"DL (Mbps)\", \"UL (Mbps)\", \"PING\", \"ISP\", \"SERVER\", \"DEVICES\")
	fmt.Println(\"---------------------------------------------------------------------------------------------\")
	for i := len(metrics) - 1; i >= 0; i-- {
		m := metrics[i]
		devCount := 0
		if i < len(deviceCounts) { devCount = deviceCounts[i] }
		fmt.Printf(\"%-20s %-10.1f %-10.1f %-8.1f %-15s %-18s %-8d\n\",
			m.Timestamp.Local().Format(\"2006-01-02 15:04:05\"),
			m.Download, m.Upload, m.Ping, m.Client, m.Server, devCount)
	}
}"
    fi
fi
