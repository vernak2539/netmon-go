# Agent Developer Guide (`AGENTS.md`)

Welcome to **netmon-go**! This repository is a pure Go port of the [netmon](https://github.com/Role1776/netmon) network monitoring and reporting application.

This document serves as the authoritative operational guide for AI coding assistants working in this repository.

---

## 1. Branching & PR Approval Rules

### Permanent Dual Branch Strategy
- `main`: Tracks the original Python codebase.
- `main-go`: The primary production branch for the Go port.
- **CRITICAL**: `main` and `main-go` are permanent parallel branches. **NEVER merge `main` into `main-go` or `main-go` into `main`**. Feature branches for Go work MUST be created off `main-go` (e.g. `issue-N-feature`) and targeted to `main-go` in PRs.

### PR Approval Rule
- **NEVER merge Pull Requests directly** without explicit user approval in the chat.
- Always open PRs, update task checkboxes on GitHub issues, update project board status to `In review`, and notify the user for approval.

---

## 2. Project Architecture & Directory Layout

```
netmon-go/
├── cmd/
│   └── netmon/         # Main entrypoint, prompts (prompts.go), report formatters (reports.go), main loop (main.go)
├── internal/
│   ├── ai/             # OpenAI API client (go-openai wrapper with custom BaseURL)
│   ├── config/         # Environment configuration loading & validation (.env support)
│   ├── db/             # SQLite persistence layer (modernc.org/sqlite, zero CGO)
│   ├── graphs/         # Dual Y-axis PNG chart rendering (go-chart/v2)
│   ├── models/         # Domain models (NetworkMetric, NetworkDevice, SpeedTest with UUIDv7)
│   ├── runner/         # Composite facade orchestrating speedtest and scanner
│   ├── scanner/        # LAN device scanning via nmap ARP discovery
│   ├── speedtest/      # Multi-threaded speed tests via showwin/speedtest-go
│   └── telegram/       # Telegram Bot API client (pure stdlib http, multipart photo uploads)
├── docs/               # Architecture docs & implementation plans
├── scripts/            # Release & helper scripts (release.sh)
├── .github/workflows/  # CI/CD workflows (ci.yml, release.yml)
├── Makefile            # Standard build targets (build, lint, test, release, clean)
└── AGENTS.md           # This agent guide
```

---

## 3. Common Developer Workflows & Commands

All standard tasks are defined in the `Makefile`:

```bash
# Build the local netmon-go binary
make build

# Run linter and formatting checks
make lint

# Run unit tests with race detection across all packages
make test

# Inspect database metrics and device scans
make db-inspect                        # Inspect 10 most recent speedtest metrics
make db-inspect SCAN_ID=<device_scan_id> # Inspect devices for a specific scan ID

# Tag a new release and push to remote (triggers GitHub Actions release workflow)
make release VERSION=v1.0.0

# Clean build artifacts
make clean
```

---

## 4. Critical Code Conventions & Gotchas

### ⚠️ Go Filename Rule
- **NEVER name regular Go domain source files `*_test.go`**. In Go, any file ending in `_test.go` is strictly treated as a test file by `go build` and will be excluded from normal package compilation. (Example: use `speedtest.go`, not `speed_test.go`).

### Zero CGO Requirement
- The project targets cross-platform static binary generation (`CGO_ENABLED=0`).
- SQLite uses `modernc.org/sqlite` (pure Go). Do NOT introduce dependencies that require CGO or gcc.

### Signal Handling & Context Cancellation
- Long-running processes (`cmd/netmon/main.go`, `speedtest`, `scanner`) MUST accept `context.Context` and handle cancellation via `signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)`.

### HTML Formatting & Telegram AI Prompts
- AI responses sent to Telegram MUST NOT contain `<br>`, `<br/>`, or `<br />` tags. Use `cleanHTMLResponse()` to convert them to standard newlines.

---

## 5. Project Board Automation (GitHub Project #3)

When managing ticket lifecycles, update GitHub Project #3 (`PVT_kwHOAAf0Ns4BeScO`) item statuses using the `gh` CLI:

- **Backlog**: `f75ad846`
- **Ready**: `61e4505c`
- **In progress**: `47fc9ee4`
- **In review**: `df73e18b`
- **Done**: `98236657`

Example command:
```bash
gh project item-edit --id <ITEM_ID> --project-id PVT_kwHOAAf0Ns4BeScO --field-id PVTSSF_lAHOAAf0Ns4BeScOzhYtzO0 --single-select-option-id <OPTION_ID>
```

---

## 6. AI-Assisted Feature Development & Superpowers Workflow

AI agents working on new features, refactoring, or bug fixes MUST follow this systematic lifecycle combining GitHub Project #3 automation and Superpowers skills:

### Step 1: Issue & Project Board Setup
1. Identify or create the target GitHub Issue (`gh issue create --title "..." --body "..."`).
2. Add the issue to GitHub Project #3 (`PVT_kwHOAAf0Ns4BeScO`) and set status to **In progress** (`47fc9ee4`).
3. Create a feature branch off `main-go` named `issue-<NUMBER>-<feature>` (`git checkout -b issue-12-feature main-go`).

### Step 2: Implementation Planning (Superpowers Skills)
1. Use `writing-plans` (or `brainstorming` for design decisions) to author a comprehensive, bite-sized implementation plan.
2. Save the plan to `docs/superpowers/plans/YYYY-MM-DD-<feature-name>.md`.
3. Plan MUST follow TDD principles (failing test -> pass test -> lint -> commit) with exact file paths and code snippets.

### Step 3: Execution & Verification
1. Execute the plan using `subagent-driven-development` (recommended for task isolation and automated review gates) or `executing-plans`.
2. As each task finishes:
   - Run `make test` and `make lint`.
   - Update task checkboxes in the GitHub Issue body (`gh issue edit <NUMBER> --body "..."`).
   - Append completed task status to the plan's ledger file (`.superpowers/sdd/<PLAN_BASENAME>/progress.md`).

### Step 4: Code Review & PR Submission
1. Complete whole-branch code review verification.
2. Submit a Pull Request targeting `main-go` with `Resolves #<NUMBER>` in the body (`gh pr create --base main-go --title "..." --body "...")`.
3. Update GitHub Project #3 status to **In review** (`df73e18b`).
4. **Notify the user for PR review & approval** (do NOT merge directly).

### Step 5: Post-Approval Merge & Completion
1. Upon user approval and PR merge to `main-go`:
2. Update GitHub Project #3 status to **Done** (`98236657`).
3. Clean up the feature branch locally (`git checkout main-go && git pull origin main-go && git branch -d issue-<NUMBER>-<feature>`).

