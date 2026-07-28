# Design Spec: Immediate Cancellation Handling in Main Loop

**Date:** 2026-07-27  
**Status:** Approved  

## Overview
When `netmon-go` receives `SIGINT` (Ctrl+C) or `SIGTERM` during a speed test or device scan cycle, `ctx` is cancelled. However, `runCycle()` previously did not check `ctx.Err()` immediately after `speedTester.Run(ctx)` returned. This caused:
1. Context cancellation errors to be logged as generic errors (`Error running speedtest: ...`).
2. Delay in exiting because the application waited until `runCycle()` finished completely before checking `ctx.Done()`.

## Key Changes
1. **Context Cancellation Guard (`cmd/netmon/main.go`)**:
   - Check `if ctx.Err() != nil` immediately after `speedTester.Run(ctx)` and `deviceScanner.Scan(ctx)`.
   - If `ctx.Err() != nil`, log `Context cancelled. Aborting speedtest cycle.` and return immediately without logging generic errors or attempting database writes / notifier calls.
2. **Immediate Exit Check**:
   - In `runCycle()`, returning early on `ctx.Err() != nil` allows the main loop's `select { case <-ctx.Done(): ... }` or immediate check to exit `main()` cleanly without delay.

## Verification Plan
- Unit tests: Add unit test in `cmd/netmon` checking that cancelled context aborts `runCycle` without error logging.
- Manual test: Run `./netmon` and press `Ctrl+C` during speedtest execution to verify immediate clean exit.
