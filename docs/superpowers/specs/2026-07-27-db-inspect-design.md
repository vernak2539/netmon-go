# Design Spec: Database Inspection Helper Script & Makefile Target

**Date:** 2026-07-27  
**Status:** Approved  

## Overview
Adds a `db-inspect` helper script in `scripts/db-inspect.sh` and a corresponding `make db-inspect` Makefile target to easily query and inspect the SQLite database metrics, speedtests, and online device counts.

## Key Design & Behavior

1. **Script Location & Name**: `scripts/db-inspect.sh`
2. **Database Path Determination**: Accepts `$DB_PATH` from `.env` or defaults to `metrics.sql`. Allows passing a path argument (e.g. `./scripts/db-inspect.sh my_custom.db`).
3. **Execution Strategy**:
   - Primary: Uses `sqlite3` CLI if installed on the system with formatted column output.
   - Fallback: Uses `go run` with `internal/db` if `sqlite3` CLI is not installed.
4. **Makefile Target**:
   ```makefile
   db-inspect:
   	@./scripts/db-inspect.sh
   ```

## Verification Strategy
- Execute `chmod +x scripts/db-inspect.sh`
- Test running `./scripts/db-inspect.sh` and `make db-inspect` on a sample DB file.
