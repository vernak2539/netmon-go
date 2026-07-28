# Design Spec: Test Notification CLI Flag (`-test-notify`)

**Date:** 2026-07-27  
**Status:** Approved  

## Overview
Adds a `-test-notify` / `--test-notify` command-line flag to `netmon-go`. When passed, `netmon` loads the configuration, initializes the configured notifier (`telegram` or `discord`), sends a test status message, logs the result, and exits cleanly (`0`) without opening the database or starting the monitoring loop.

## Key Changes
1. **Config Struct Update (`internal/config/config.go`)**:
   - Add `TestNotify bool` field to `Config`.
   - Register `flag.Bool("test-notify", false, "Send a test notification and exit")`.
   - When `-test-notify` is `true`, bypass the `DB_PATH` non-empty requirement so users can test notifications even if `DB_PATH` isn't initialized yet.
2. **Main Entrypoint Update (`cmd/netmon/main.go`)**:
   - Check `cfg.TestNotify`:
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

## Verification Strategy
- Add unit test in `internal/config/config_test.go` checking `-test-notify` flag parsing.
- Run `make test` & `make lint`.
- Test running `./netmon -test-notify` directly.
