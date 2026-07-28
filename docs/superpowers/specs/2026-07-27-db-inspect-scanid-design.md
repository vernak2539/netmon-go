# Design Spec: Device Scan ID Lookup for db-inspect Script

**Date:** 2026-07-27  
**Status:** Approved  

## Overview
Enhance `scripts/db-inspect.sh` and Makefile to support passing a `device_scan_id` (via `SCAN_ID` argument or environment variable).

## Design Details
1. **Behavior when `SCAN_ID` is provided**:
   - Query `device_scans` table for the matching UUID (`id = SCAN_ID`).
   - Output formatted tabular details showing IP addresses and latency ms recorded for that scan.
2. **Behavior when `SCAN_ID` is empty** (default):
   - Display the standard 10 most recent speedtest metrics table (showing `device_scan_id` for each row).
3. **Usage Syntax**:
   - CLI: `./scripts/db-inspect.sh [DB_PATH] [SCAN_ID]`
   - Make: `make db-inspect SCAN_ID=<device_scan_id>`

## Verification Plan
- Run `make db-inspect` to see standard metric list.
- Run `make db-inspect SCAN_ID=019fa67a-3941-7309-8e9a-71661b0f5c64` to verify single scan output.
