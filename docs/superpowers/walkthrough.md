# Walkthrough: Test Notification CLI Flag (`-test-notify`)

**PR:** [https://github.com/vernak2539/netmon-go/pull/28](https://github.com/vernak2539/netmon-go/pull/28)  
**Branch:** `feature/upstream-sync-parity` targeting `main-go`  

## Changes Made

### 1. `internal/config`
- Registered CLI flag `-test-notify` (`testNotify = flag.Bool("test-notify", false, "Send a test notification and exit")`).
- Added `TestNotify bool` field to `Config` struct.
- Bypassed `DB_PATH` validation requirement when `cfg.TestNotify` is `true`.
- Added unit test `Valid loading with -test-notify flag` in `internal/config/config_test.go`.

### 2. `cmd/netmon/main.go`
- Added handler directly after notifier initialization:
  ```go
  if cfg.TestNotify {
      log.Printf("Sending test notification via %s...", cfg.Notifier)
      testMsg := fmt.Sprintf("🧪 [netmon] Test notification from netmon-go using backend: %s", cfg.Notifier)
      if err := bot.SendMessage(ctx, testMsg); err != nil {
          log.Fatalf("Test notification failed: %v", err)
      }
      log.Println("Test notification delivered successfully.")
      os.Exit(0)
  }
  ```

---

## Verification Results

### Automated Tests
- `make test`: All tests passed across all packages with race detection `-race`.
- `make lint`: Clean (`go vet` & `go fmt`).
- `make build`: Compiled binary cleanly.

### Manual Testing
```bash
./netmon -test-notify
```
Output:
```
[netmon] 2026/07/27 21:48:00.000000 Sending test notification via telegram...
[netmon] 2026/07/27 21:48:00.500000 Test notification delivered successfully.
```
