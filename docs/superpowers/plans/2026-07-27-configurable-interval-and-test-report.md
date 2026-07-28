# Configurable Interval & Test Report Flag Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the network monitoring cycle interval configurable via CLI flag (`-interval`) and environment variable (`SPEEDTEST_INTERVAL`), and add a `-test-report` CLI flag to generate and dispatch a full 24-hour report with graph before exiting.

**Architecture:** Extend `internal/config/config.go` to parse `-interval` (flag/env `SPEEDTEST_INTERVAL`, defaulting to 1 hour) and `-test-report` flag into `Config`. In `cmd/netmon/main.go`, pass `cfg.SpeedtestInterval` to `time.NewTicker`, and execute an early-exit `-test-report` flow after DB and notifier initialization that queries past 24h metrics, plots the graph via `graphs.Plot`, sends it via the configured `Notifier`, cleans up the PNG file, and exits 0.

**Tech Stack:** Go 1.25.6 stdlib (`flag`, `time`, `os`, `strconv`), `modernc.org/sqlite`, `go-chart/v2`, custom `notifier.Notifier` package.

## Global Constraints

- Go Version: `go 1.25.6`
- Target Branch: `main-go`
- Zero CGO (`CGO_ENABLED=0`)
- Standard testing package only (`testing`, `httptest`, `t.Run`, `t.Setenv`)

---

### Task 1: Add `SpeedtestInterval` and `TestReport` to `internal/config`

**Files:**
- Modify: `internal/config/config.go`
- Test: `internal/config/config_test.go`

**Interfaces:**
- Produces: `Config.SpeedtestInterval` (`time.Duration`), `Config.TestReport` (`bool`)

- [ ] **Step 1: Write failing unit tests for new config options**

Add test cases in `internal/config/config_test.go`:
- Test `-test-report` flag bypasses `DB_PATH` requirement check when set (or tests `TestReport` flag parsing).
- Test `-interval` CLI flag parsing (`-interval 15m` -> `15 * time.Minute`).
- Test `SPEEDTEST_INTERVAL` environment variable parsing (e.g. `SPEEDTEST_INTERVAL=1800` or `SPEEDTEST_INTERVAL=30m`).
- Test default `SpeedtestInterval` is `1 * time.Hour` when unspecified.
- Test invalid `SPEEDTEST_INTERVAL` values return error.

```go
	t.Run("Valid loading with -test-report flag", func(t *testing.T) {
		cleanup()
		os.Setenv("TG_BOT_TOKEN", "test-token")
		os.Setenv("TG_CHAT_ID", "test-chat")
		os.Setenv("DB_PATH", "test.db")

		*envFile = ".env"
		*testReport = true
		defer func() { *testReport = false }()

		cfg, err := Load()
		if err != nil {
			t.Fatalf("unexpected error when test-report is set: %v", err)
		}

		if !cfg.TestReport {
			t.Errorf("expected TestReport to be true, got false")
		}
	})

	t.Run("Custom SPEEDTEST_INTERVAL from env", func(t *testing.T) {
		cleanup()
		os.Setenv("TG_BOT_TOKEN", "test-token")
		os.Setenv("TG_CHAT_ID", "test-chat")
		os.Setenv("DB_PATH", "test.db")
		os.Setenv("SPEEDTEST_INTERVAL", "1800") // 1800 seconds = 30m

		*envFile = ".env"

		cfg, err := Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.SpeedtestInterval != 30*time.Minute {
			t.Errorf("expected SpeedtestInterval to be 30m, got %v", cfg.SpeedtestInterval)
		}
	})
```

- [ ] **Step 2: Run test to verify failure**

Run: `go test -v ./internal/config`
Expected: FAIL due to undefined variables `testReport` / struct fields.

- [ ] **Step 3: Implement config changes**

In `internal/config/config.go`:
1. Add `SpeedtestInterval time.Duration` and `TestReport bool` fields to `Config` struct.
2. Add flag declarations:
   ```go
   var (
       envFile           = flag.String("env", ".env", "Path to the .env file")
       testNotify        = flag.Bool("test-notify", false, "Send a test notification and exit")
       testReport        = flag.Bool("test-report", false, "Send a full 24h test report with graph and exit")
       speedtestInterval = flag.Duration("interval", 0, "Interval between speedtest cycles (e.g. 1h, 30m)")
   )
   ```
3. In `Load()`:
   - Parse `SPEEDTEST_INTERVAL` env var: accept duration string (e.g. `"30m"`, `"1h"`) or integer seconds string (e.g. `"1800"`).
   - Precedence: If `-interval` flag is set (> 0), use flag value. Else if `SPEEDTEST_INTERVAL` env var is non-empty, use parsed env value. Else default to `1 * time.Hour`.
   - Validate `SpeedtestInterval > 0`.
   - Assign `TestReport: *testReport` and `SpeedtestInterval: intervalVal`.

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -v ./internal/config`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/config/config.go internal/config/config_test.go
git commit -m "feat(config): add SpeedtestInterval and TestReport config options"
```

---

### Task 2: Implement Configurable Interval and `-test-report` in `cmd/netmon`

**Files:**
- Modify: `cmd/netmon/main.go`
- Modify: `cmd/netmon/main_test.go`

**Interfaces:**
- Consumes: `config.Config.SpeedtestInterval`, `config.Config.TestReport`
- Produces: Executable CLI flags `-interval` and `-test-report`

- [ ] **Step 1: Write helper function & tests for test-report logic in `cmd/netmon`**

In `cmd/netmon/main_test.go`, add unit test verifying `generateAndSendReport` logic or helper formatting when metrics exist vs when metrics are empty.

- [ ] **Step 2: Implement `-test-report` handling and configurable ticker in `cmd/netmon/main.go`**

In `cmd/netmon/main.go`:
1. Use `cfg.SpeedtestInterval` for ticker creation:
   ```go
   ticker := time.NewTicker(cfg.SpeedtestInterval)
   ```
2. Handle `cfg.TestReport` after `database` and `aiClient` initialization:
   ```go
   if cfg.TestReport {
       log.Printf("Generating test report with graph via %s...", cfg.Notifier)
       metrics, deviceCounts, err := database.GetMetricsWithDeviceCounts(ctx)
       if err != nil {
           log.Fatalf("Failed to retrieve metrics for test report: %v", err)
       }
       if len(metrics) == 0 {
           log.Fatalf("No speedtest metrics found in database (%s) to generate graph report.", cfg.DBPath)
       }

       var report string
       if aiClient != nil {
           var userMessage strings.Builder
           for i, m := range metrics {
               devCount := 0
               if i < len(deviceCounts) {
                   devCount = deviceCounts[i]
               }
               userMessage.WriteString(fmt.Sprintf(
                   ReportUserItemFormat,
                   m.Timestamp.Local().Format("2006-01-02 15:04:05"),
                   m.Download,
                   m.Upload,
                   m.Ping,
                   m.Client,
                   m.Server,
                   float64(m.BytesReceived)/1000000.0,
                   float64(m.BytesSent)/1000000.0,
                   m.Share,
                   devCount,
               ))
               userMessage.WriteString("\n")
           }

           _ = bot.SendChatAction(ctx, notifier.ChatActionTyping)
           var aiErr error
           report, aiErr = aiClient.SendMessage(ctx, userMessage.String(), ReportSystemPrompt)
           if aiErr != nil {
               log.Printf("AI report generation failed: %v, falling back to status message", aiErr)
           } else {
               report = cleanHTMLResponse(report)
           }
       }

       if report == "" {
           latestMetric := metrics[len(metrics)-1]
           latestDeviceCount := deviceCounts[len(deviceCounts)-1]
           report = formatMiniReport(&latestMetric, latestDeviceCount)
       }

       _ = bot.SendChatAction(ctx, notifier.ChatActionUploadPhoto)
       graphPath, err := graphs.Plot(metrics, deviceCounts)
       if err != nil {
           log.Fatalf("Failed to generate graph: %v", err)
       }

       photoBytes, err := os.ReadFile(graphPath)
       if err != nil {
           _ = os.Remove(graphPath)
           log.Fatalf("Failed to read graph file: %v", err)
       }

       if err := bot.SendPhoto(ctx, photoBytes, report); err != nil {
           _ = os.Remove(graphPath)
           log.Fatalf("Failed to send photo: %v", err)
       }
       _ = os.Remove(graphPath)

       log.Println("Test report with graph delivered successfully.")
       os.Exit(0)
   }
   ```

- [ ] **Step 3: Run tests to verify compilation and execution**

Run: `go test -v ./cmd/netmon/...`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add cmd/netmon/main.go cmd/netmon/main_test.go
git commit -m "feat(netmon): support configurable ticker interval and -test-report flag"
```

---

### Task 3: Documentation and Verification

**Files:**
- Modify: `.env.example`
- Modify: `scripts/setup.sh`
- Modify: `AGENTS.md`

- [ ] **Step 1: Update `.env.example`**

Add `SPEEDTEST_INTERVAL` variable to `.env.example`:
```env
# Speedtest cycle interval in seconds or duration format (e.g. 3600 or 1h)
SPEEDTEST_INTERVAL=3600
```

- [ ] **Step 2: Update `scripts/setup.sh`**

Add prompt/export for `SPEEDTEST_INTERVAL` if user configures custom interval during setup.

- [ ] **Step 3: Update `AGENTS.md`**

Ensure `SPEEDTEST_INTERVAL`, `-interval`, and `-test-report` are noted under CLI flags / configuration gotchas.

- [ ] **Step 4: Run full project tests and linter**

Run: `make test && make lint && make build`
Expected: Clean pass with 0 errors.

- [ ] **Step 5: Commit**

```bash
git add .env.example scripts/setup.sh AGENTS.md
git commit -m "docs: update .env.example, setup.sh, and AGENTS.md with SPEEDTEST_INTERVAL and -test-report"
```

---
