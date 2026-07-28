# Upstream Sync Parity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement full feature parity with the 17 upstream Python commits: adding Discord webhook support, a unified `Notifier` abstraction, configurable request timeouts, local timezone chart alignment, and daemon cycle resiliency.

**Architecture:** Create `internal/notifier` package defining `Notifier` interface, implement `internal/discord` client, update `internal/config` with `NOTIFIER`, `DISCORD_WEBHOOK_URL`, and `REQUEST_TIMEOUT`, ensure `.Local()` timestamp formatting in `internal/graphs` and `cmd/netmon`, and add error resiliency in `cmd/netmon/main.go`.

**Tech Stack:** Go stdlib (`net/http`, `mime/multipart`, `time`), `github.com/vernak2539/netmon-go/internal/...`

## Global Constraints

- Target `CGO_ENABLED=0` cross-platform binary compatibility.
- Ensure all tests pass (`make test`) and linting is clean (`make lint`).

---

## Proposed Tasks

### Task 1: Update Configuration (`internal/config`)
- Add `Notifier` (`telegram`|`discord`), `DiscordWebhookURL`, and `RequestTimeout` (seconds) to `Config`.
- Default `Notifier` to `"telegram"`, `RequestTimeout` to `30s`.
- Validate `DiscordWebhookURL` when `NOTIFIER=discord`.

### Task 2: Implement Notifier Interface & Discord Webhook (`internal/notifier`, `internal/discord`, `internal/telegram`)
- Create `internal/notifier` defining `Notifier` interface and `ChatAction`.
- Implement `Notifier` in `internal/telegram` and `internal/discord`.
- Add unit tests for `internal/discord` (message sending & multipart photo uploads).

### Task 3: Local Timezone Alignment & Cycle Resiliency (`cmd/netmon`, `internal/graphs`)
- Format timestamps using `.Local()` in graph axis and report generator.
- Wrap main cycle execution in `cmd/netmon/main.go` so single-cycle failures log cleanly and retry on the next ticker interval without crashing the process.

---

## Verification Plan
- `make test` across all packages.
- `make lint` clean output.
- `make build` executable output.
