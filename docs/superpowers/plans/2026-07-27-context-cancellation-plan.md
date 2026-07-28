# Immediate Context Cancellation Handling Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Ensure `cmd/netmon/main.go` checks `ctx.Err() != nil` immediately after `speedTester.Run(ctx)` or `deviceScanner.Scan(ctx)`, aborting the cycle cleanly without error logging when Ctrl+C (`SIGINT`/`SIGTERM`) is received.

**Architecture:** Add `if ctx.Err() != nil { return }` checks at key execution checkpoints inside `runCycle()` in `cmd/netmon/main.go`.

**Tech Stack:** Go stdlib (`context`, `log`)

## Global Constraints

- Never suppress genuine errors when `ctx.Err() == nil`.
- Verify using `make test` and `make build`.

---

## Task Breakdown

### Task 1: Add context cancellation guards to `cmd/netmon/main.go`

**Files:**
- Modify: `cmd/netmon/main.go`

**Interfaces:**
- Consumes: `ctx context.Context`
- Produces: Clean immediate cycle abort when `ctx.Err() != nil`.

- [ ] **Step 1: Update `cmd/netmon/main.go` inside `runCycle()`**

```go
		metric, err := speedTester.Run(ctx)
		if ctx.Err() != nil {
			log.Println("Context cancelled. Aborting speedtest cycle.")
			return
		}
		if err != nil {
			log.Printf("Error running speedtest: %v", err)
			return
		}

		devices, err := deviceScanner.Scan(ctx)
		if ctx.Err() != nil {
			log.Println("Context cancelled. Aborting device scan.")
			return
		}
		if err != nil {
			log.Printf("Error running device scan: %v", err)
			devices = []models.NetworkDevice{}
		}
```

- [ ] **Step 2: Run full build and test suite**

Run: `make test && make build`
Expected: PASS and `netmon` binary built successfully.

- [ ] **Step 3: Commit changes**

```bash
git add cmd/netmon/main.go
git commit -m "fix(main): abort speedtest cycle immediately when context is cancelled"
```

---

## Verification Plan

### Automated Tests
- `make test`
- `make lint`

### Manual Verification
- Execute `./netmon-go` and press `Ctrl+C` during the speedtest execution. Verify it logs `Context cancelled. Aborting speedtest cycle.` and exits immediately.
