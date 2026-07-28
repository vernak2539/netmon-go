# Optional AI Configuration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make `AI_API_KEY` and AI configuration optional in `netmon-go` so that the application loads and runs seamlessly when AI credentials are missing, falling back to non-AI report delivery.

**Architecture:** Update `internal/config` to allow empty `AI_API_KEY` (while enforcing `AI_MODEL` and `AI_BASE_URL` only if `AI_API_KEY` is present), update `internal/ai` to return `(nil, nil)` on empty `AI_API_KEY`, and update `cmd/netmon/main.go` to handle `aiClient == nil` cleanly by sending status text and photo during the 4-hour cycle.

**Tech Stack:** Go stdlib, `github.com/vernak2539/netmon-go/internal/...`

## Global Constraints

- Never break existing Telegram notifications or database recording.
- Target `CGO_ENABLED=0` cross-compilation compatibility.
- Verify using `make test` and `make build`.

---

## Proposed Changes

### Component 1: `internal/config`
#### [MODIFY] `internal/config/config.go`
#### [MODIFY] `internal/config/config_test.go`

- Update `Load()` to no longer require `AI_API_KEY`.
- If `AI_API_KEY` is non-empty, enforce `AI_MODEL` and `AI_BASE_URL`.
- If `AI_API_KEY` is empty, ignore missing `AI_MODEL` / `AI_BASE_URL`.
- Add unit tests verifying `Load()` succeeds without `AI_API_KEY`.

### Component 2: `internal/ai`
#### [MODIFY] `internal/ai/ai.go`
#### [MODIFY] `internal/ai/ai_test.go`

- Update `New(apiKey, model, baseURL)` to return `(nil, nil)` when `apiKey` is empty.
- Add unit tests verifying `New("", "", "")` returns `(nil, nil)`.

### Component 3: `cmd/netmon`
#### [MODIFY] `cmd/netmon/main.go`

- Support `aiClient == nil` when initializing components in `main()`.
- During the 4-hour detailed report cycle:
  - If `aiClient != nil`, perform the AI analysis prompt & image report as before.
  - If `aiClient == nil`, construct the mini-report status message for the latest metric and upload the graph photo with this mini-report message.

---

## Task Breakdown

### Task 1: Update `internal/config` to make `AI_API_KEY` optional

**Files:**
- Modify: `internal/config/config.go`
- Modify: `internal/config/config_test.go`

**Interfaces:**
- Consumes: Environment variables (`AI_API_KEY`, `AI_MODEL`, `AI_BASE_URL`, `TG_BOT_TOKEN`, `TG_CHAT_ID`, `DB_PATH`)
- Produces: `config.Load() (*Config, error)` where `Config.AIAPIKey` may be empty.

- [ ] **Step 1: Write failing test in `internal/config/config_test.go`**

```go
t.Run("Valid loading without AI_API_KEY", func(t *testing.T) {
    cleanup()
    os.Setenv("TG_BOT_TOKEN", "test-token")
    os.Setenv("TG_CHAT_ID", "test-chat")
    os.Setenv("DB_PATH", "test.db")

    *envFile = ".env"

    cfg, err := Load()
    if err != nil {
        t.Fatalf("unexpected error when AI_API_KEY is omitted: %v", err)
    }

    if cfg.AIAPIKey != "" {
        t.Errorf("expected empty AIAPIKey, got %s", cfg.AIAPIKey)
    }
})
```

- [ ] **Step 2: Run tests to verify failure**

Run: `go test -v ./internal/config`
Expected: FAIL with "AI_API_KEY not found or empty in environment"

- [ ] **Step 3: Update `internal/config/config.go` implementation**

```go
	cfg := &Config{
		AIAPIKey:   os.Getenv("AI_API_KEY"),
		AIModel:    os.Getenv("AI_MODEL"),
		AIBaseURL:  os.Getenv("AI_BASE_URL"),
		TGBotToken: os.Getenv("TG_BOT_TOKEN"),
		TGChatID:   os.Getenv("TG_CHAT_ID"),
		DBPath:     os.Getenv("DB_PATH"),
	}

	if strings.TrimSpace(cfg.AIAPIKey) != "" {
		if strings.TrimSpace(cfg.AIModel) == "" {
			return nil, fmt.Errorf("AI_MODEL not found or empty in environment when AI_API_KEY is set")
		}
		if strings.TrimSpace(cfg.AIBaseURL) == "" {
			return nil, fmt.Errorf("AI_BASE_URL not found or empty in environment when AI_API_KEY is set")
		}
	}

	if strings.TrimSpace(cfg.DBPath) == "" {
		return nil, fmt.Errorf("DB_PATH not found or empty in environment")
	}
	if strings.TrimSpace(cfg.TGBotToken) == "" {
		return nil, fmt.Errorf("TG_BOT_TOKEN not found or empty in environment")
	}
	if strings.TrimSpace(cfg.TGChatID) == "" {
		return nil, fmt.Errorf("TG_CHAT_ID not found or empty in environment")
	}

	return cfg, nil
```

- [ ] **Step 4: Run tests to verify pass**

Run: `go test -v ./internal/config`
Expected: PASS

- [ ] **Step 5: Commit changes**

```bash
git add internal/config/config.go internal/config/config_test.go
git commit -m "feat(config): make AI_API_KEY optional"
```

---

### Task 2: Update `internal/ai` to return `(nil, nil)` on empty API Key

**Files:**
- Modify: `internal/ai/ai.go`
- Modify: `internal/ai/ai_test.go`

**Interfaces:**
- Consumes: `apiKey`, `model`, `baseURL` strings
- Produces: `ai.New(apiKey, model, baseURL) (*Client, error)` returning `(nil, nil)` if `apiKey == ""`

- [ ] **Step 1: Write failing test in `internal/ai/ai_test.go`**

```go
func TestNewEmptyAPIKey(t *testing.T) {
    client, err := New("", "model", "http://localhost")
    if err != nil {
        t.Fatalf("expected nil error for empty API key, got: %v", err)
    }
    if client != nil {
        t.Errorf("expected nil client for empty API key, got: %v", client)
    }
}
```

- [ ] **Step 2: Run test to verify failure**

Run: `go test -v ./internal/ai`
Expected: FAIL ("api_key cannot be empty")

- [ ] **Step 3: Update `internal/ai/ai.go`**

```go
func New(apiKey, model, baseURL string) (*Client, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, nil
	}
	if strings.TrimSpace(model) == "" {
		return nil, fmt.Errorf("model cannot be empty")
	}
	if strings.TrimSpace(baseURL) == "" {
		return nil, fmt.Errorf("base_url cannot be empty")
	}

	config := openai.DefaultConfig(apiKey)
	config.BaseURL = baseURL

	return &Client{
		client: openai.NewClientWithConfig(config),
		model:  model,
	}, nil
}
```

- [ ] **Step 4: Run tests to verify pass**

Run: `go test -v ./internal/ai`
Expected: PASS

- [ ] **Step 5: Commit changes**

```bash
git add internal/ai/ai.go internal/ai/ai_test.go
git commit -m "feat(ai): return nil client gracefully when API key is empty"
```

---

### Task 3: Update `cmd/netmon/main.go` to handle optional AI client

**Files:**
- Modify: `cmd/netmon/main.go`

**Interfaces:**
- Consumes: `aiClient *ai.Client` (may be `nil`)
- Produces: Executable binary that runs continuously with or without AI client.

- [ ] **Step 1: Update `cmd/netmon/main.go` logic**

```go
	var report string
	if aiClient != nil {
		t.send_chat_action(tg.ChatAction.TYPING)
		var err error
		report, err = aiClient.SendMessage(ctx, userMessage, REPORT_SYSTEM_PROMPT)
		if err != nil {
			log.Printf("AI report generation failed: %v, falling back to status message", err)
		} else {
			report = cleanHTMLResponse(report)
		}
	}

	if report == "" {
		latestMetric := metrics[len(metrics)-1]
		latestDeviceCount := deviceCounts[len(deviceCounts)-1]
		dlSpeed := float64(latestMetric.Download) / 1000000.0
		statusText := determineStatusText(dlSpeed, latestMetric.Ping)
		report = formatMiniReport(latestMetric, latestDeviceCount, statusText)
	}

	bot.SendChatAction(telegram.ChatActionUploadPhoto)
	graph := graphs.New(metrics, deviceCounts)
	graphName, err := graph.Plot()
	if err != nil {
		log.Printf("Failed to generate graph: %v", err)
	} else {
		graphBytes, err := os.ReadFile(graphName)
		if err != nil {
			log.Printf("Failed to read graph file: %v", err)
		} else {
			if err := bot.SendPhoto(graphBytes, report); err != nil {
				log.Printf("Failed to send photo report: %v", err)
			}
		}
		_ = os.Remove(graphName)
	}
```

- [ ] **Step 2: Run full build and test suite**

Run: `make test && make build`
Expected: PASS and `netmon` binary built successfully.

- [ ] **Step 3: Commit changes**

```bash
git add cmd/netmon/main.go
git commit -m "feat(main): support optional AI client in report delivery cycle"
```

---

## Verification Plan

### Automated Tests
- `go test -v ./...`
- `make lint`

### Manual Verification
- Test binary launch without `AI_API_KEY` set in environment or `.env`: `./netmon`
- Verify binary logs `[netmon] The bot has been started.` instead of failing at `Failed to load configuration: AI_API_KEY not found...`.
