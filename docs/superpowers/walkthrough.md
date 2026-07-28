# Walkthrough: Upstream Sync Parity (17 Commits)

**PR:** [https://github.com/vernak2539/netmon-go/pull/27](https://github.com/vernak2539/netmon-go/pull/27)  
**Branch:** `feature/upstream-sync-parity` targeting `main-go`  

## Changes Made

### 1. `internal/config`
- Added `NOTIFIER` (`"telegram"` | `"discord"`, default `"telegram"`).
- Added `DISCORD_WEBHOOK_URL` (mandatory if `NOTIFIER=discord`).
- Added `REQUEST_TIMEOUT` parsed as `time.Duration` (default `30s`).
- Unit tests added in `internal/config/config_test.go`.

### 2. `internal/notifier`, `internal/telegram`, `internal/discord`
- Defined `notifier.Notifier` interface (`SendMessage`, `SendPhoto`, `SendChatAction`) and `ChatAction` constants.
- Updated `telegram.Client` to accept `context.Context` and `notifier.ChatAction`.
- Created `internal/discord` webhook client with JSON text delivery and multipart photo uploads.

### 3. `cmd/netmon` & `internal/graphs`
- Metric timestamps use `.Local()` formatting in graph X-axis labels and report text.
- `cmd/netmon/main.go` dynamically instantiates Discord or Telegram notifier based on `cfg.Notifier`.
- Main execution loop wraps cycle execution with panic recovery and error logging so single-cycle failures do not crash the daemon process.

---

## Verification Results

### Automated Tests
- Ran `make test` across all packages with race detection: **PASS**
- Ran `make lint` (`go vet` & `go fmt`): **PASS**
- Ran `make build`: **PASS** (compiled `netmon-go` binary cleanly with zero CGO dependencies)
