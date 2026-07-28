#!/usr/bin/env bash
set -euo pipefail

# Determine DB path: first argument > DB_PATH env var > default metrics.sql
DB="${1:-${DB_PATH:-metrics.sql}}"

if [ ! -f "$DB" ]; then
    echo "[db-inspect] Database file '$DB' not found."
    exit 1
fi

echo "=== Inspecting Database: $DB ==="

if command -v sqlite3 >/dev/null 2>&1; then
    sqlite3 -header -column "$DB" "
    SELECT 
        m.timestamp, 
        round(m.download / 1000000.0, 2) AS dl_mbps, 
        round(m.upload / 1000000.0, 2) AS ul_mbps, 
        m.ping, 
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

	fmt.Printf(\"Total Records: %d\n\", len(metrics))
	for i, m := range metrics {
		devCount := 0
		if i < len(deviceCounts) { devCount = deviceCounts[i] }
		fmt.Printf(\"[%s] DL: %.2f Mbps | UL: %.2f Mbps | Ping: %.2f ms | Devices: %d | Server: %s\n\",
			m.Timestamp.Local().Format(\"2006-01-02 15:04:05\"),
			m.Download/1e6, m.Upload/1e6, m.Ping, devCount, m.Server)
	}
}"
fi
