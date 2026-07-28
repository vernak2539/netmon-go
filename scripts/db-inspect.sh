#!/usr/bin/env bash
set -euo pipefail

# Determine DB path: first argument > DB_PATH env var > default metrics.sql
DB="${1:-${DB_PATH:-metrics.sql}}"

go run scripts/db_inspect.go "$DB"
