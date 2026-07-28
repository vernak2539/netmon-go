# Test Notification CLI Flag Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a `-test-notify` command-line flag to `netmon-go` that sends a test notification via the configured notifier backend and exits cleanly.

**Architecture:** Update `internal/config` to register `test-notify` flag and allow bypassing `DB_PATH` validation when `-test-notify` is passed. Update `cmd/netmon/main.go` to handle `-test-notify`, send test message, log result, and exit `0`.

**Tech Stack:** Go stdlib (`flag`, `log`, `os`), `github.com/vernak2539/netmon-go/internal/...`

## Global Constraints

- Verify using `make test` and `make build`.
- Maintain clean exit code `0` on successful test notification delivery.

---

## Task Breakdown

### Task 1: Add `TestNotify` field and flag to `internal/config`

**Files:**
- Modify: `internal/config/config.go`
- Modify: `internal/config/config_test.go`

**Interfaces:**
- Consumes: CLI flag `-test-notify`
- Produces: `Config.TestNotify` boolean field.

- [ ] **Step 1: Write failing test in `internal/config/config_test.go`**

```go
t.Run("Valid loading with -test-notify flag", func(t *testing.T) {
    cleanup()
    os.Setenv("TG_BOT_TOKEN", "test-token")
    os.Setenv("TG_CHAT_ID", "test-chat")
    // Note DB_PATH is omitted to test bypass when TestNotify is set

    *testNotify = true
    defer func() { *testNotify = false }()

    cfg, err := Load()
    if err != nil {
        t.Fatalf("unexpected error when test-notify is set: %v", err)
    }

    if !cfg.TestNotify {
        t.Errorf("expected TestNotify to be true, got false")
    }
})
```

- [ ] **Step 2: Run test to verify failure**

Run: `go test -v ./internal/config`
Expected: FAIL ("testNotify undefined" or "DB_PATH not found")

- [ ] **Step 3: Update `internal/config/config.go`**

```go
var (
    envFile    = flag.String("env", ".env", "Path to the .env file")
    testNotify = flag.Bool("test-notify", false, "Send a test notification and exit")
)

// Add TestNotify bool to Config struct:
type Config struct {
    ...
    TestNotify bool
}

// In Load():
    cfg := &Config{
        ...
        TestNotify: *testNotify,
    }

    if !cfg.TestNotify && strings.TrimSpace(cfg.DBPath) == "" {
        return nil, fmt.Errorf("DB_PATH not found or empty in environment")
    }
```

- [ ] **Step 4: Run test to verify pass**

Run: `go test -v ./internal/config`
Expected: PASS

- [ ] **Step 5: Commit changes**

```bash
git add internal/config/config.go internal/config/config_test.go
git commit -m "feat(config): add -test-notify CLI flag"
```

---

### Task 2: Implement `-test-notify` handling in `cmd/netmon/main.go`

**Files:**
- Modify: `cmd/netmon/main.go`

**Interfaces:**
- Consumes: `cfg.TestNotify` boolean flag, `bot notifier.Notifier`
- Produces: Sends test message and exits `os.Exit(0)` when `cfg.TestNotify` is `true`.

- [ ] **Step 1: Update `cmd/netmon/main.go` after notifier initialization**

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

- [ ] **Step 2: Run full build and test suite**

Run: `make test && make build`
Expected: PASS and `netmon` binary built successfully.

- [ ] **Step 3: Commit changes**

```bash
git add cmd/netmon/main.go
git commit -m "feat(main): handle -test-notify flag to send test message and exit"
```

---

## Verification Plan

### Automated Tests
- `make test`
- `make lint`

### Manual Verification
- Execute `./netmon -test-notify` and verify test message delivery and immediate clean exit.
