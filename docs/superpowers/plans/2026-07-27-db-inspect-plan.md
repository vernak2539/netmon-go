# Database Inspection Script Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create `scripts/db-inspect.sh` script and `make db-inspect` Makefile target to easily display SQLite metric and device data.

**Architecture:** Create executable `scripts/db-inspect.sh` that checks for `sqlite3` CLI with fallback to Go query execution, and add `db-inspect` to `Makefile`.

**Tech Stack:** Bash, Go, SQLite

## Global Constraints

- Ensure `scripts/db-inspect.sh` is executable (`chmod +x`).
- Verify using `make db-inspect`.

---

## Task Breakdown

### Task 1: Create `scripts/db-inspect.sh` and update `Makefile`

**Files:**
- Create: `scripts/db-inspect.sh`
- Modify: `Makefile`

- [ ] **Step 1: Create `scripts/db-inspect.sh`**

```bash
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
        st.device_scan_id
    FROM speed_tests st
    JOIN network_metrics m ON st.metric_id = m.id
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
```

- [ ] **Step 2: Make `scripts/db-inspect.sh` executable**

Run: `chmod +x scripts/db-inspect.sh`

- [ ] **Step 3: Update `Makefile` to include `db-inspect`**

Add `db-inspect` to `.PHONY` and define:
```makefile
.PHONY: all build lint test clean release db-inspect

db-inspect:
	@./scripts/db-inspect.sh
```

- [ ] **Step 4: Commit changes**

```bash
git add scripts/db-inspect.sh Makefile
git commit -m "feat: add db-inspect script and Makefile target"
```

---

## Verification Plan

### Automated / Command Verification
- `make db-inspect`
