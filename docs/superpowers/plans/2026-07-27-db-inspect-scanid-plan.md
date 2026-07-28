# Device Scan ID Input for db-inspect Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Update `scripts/db-inspect.sh` and `Makefile` to accept an optional `SCAN_ID` argument/variable and display the discovered devices (IP and latency) for that scan.

**Architecture:** Check if `$SCAN_ID` or `$2` is provided. If set, query `device_scans` for that ID. If not set, run standard top-10 speedtests table query.

**Tech Stack:** Bash, Go, SQLite

## Global Constraints

- Handle `SCAN_ID` gracefully if invalid or not found.
- Maintain `make db-inspect` backward compatibility.

---

## Task Breakdown

### Task 1: Update `scripts/db-inspect.sh` and `Makefile`

**Files:**
- Modify: `scripts/db-inspect.sh`
- Modify: `Makefile`

- [ ] **Step 1: Update `scripts/db-inspect.sh` to handle optional `SCAN_ID`**

```bash
#!/usr/bin/env bash
set -euo pipefail

# Determine DB path: first argument > DB_PATH env var > default metrics.sql
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
	// fallback print
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
```

- [ ] **Step 2: Update `Makefile` target `db-inspect`**

```makefile
db-inspect:
	@./scripts/db-inspect.sh "$(DB_PATH)" "$(SCAN_ID)"
```

- [ ] **Step 3: Commit changes**

```bash
git add scripts/db-inspect.sh Makefile
git commit -m "feat(scripts): support device_scan_id lookup in db-inspect"
```

---

## Verification Plan
- `make db-inspect`
- `make db-inspect SCAN_ID=019fa67a-3941-7309-8e9a-71661b0f5c64`
