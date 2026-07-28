# Design Spec: Upstream Synchronization & Full Parity (17 Python Commits)

**Date:** 2026-07-27  
**Status:** Draft  

## Overview
This design spec outlines bringing `netmon-go` (`main-go` branch) to full feature parity with the latest 17 commits synced to Python `main`.

## Key Objectives & Feature Areas

### 1. Multi-Notifier System (`internal/notifier` & `internal/discord`)
- **Config Updates (`internal/config`)**:
  - `NOTIFIER`: `"telegram"` (default) or `"discord"`.
  - `DISCORD_WEBHOOK_URL`: Optional when Telegram is selected; mandatory when `NOTIFIER=discord`.
  - `REQUEST_TIMEOUT`: Integer in seconds (default: `30s`).
- **Notifier Interface (`internal/notifier`)**:
  - Define `Notifier` interface in `internal/notifier`:
    ```go
    type ChatAction string

    const (
        ChatActionTyping      ChatAction = "typing"
        ChatActionUploadPhoto ChatAction = "upload_photo"
    )

    type Notifier interface {
        SendMessage(ctx context.Context, text string) error
        SendPhoto(ctx context.Context, photo []byte, caption string) error
        SendChatAction(ctx context.Context, action ChatAction) error
    }
    ```
  - `internal/telegram`: Implement `Notifier` interface.
  - `internal/discord`: Implement `Notifier` interface via Discord Webhook execution (`POST /api/webhooks/...` with multipart photo uploads or JSON message payloads).

### 2. Timezone Alignment for Graphs & Reports
- In `internal/graphs` and `cmd/netmon`: Convert stored metric timestamps (`m.Timestamp.Local()`) before formatting report text and plotting graph X-axis labels, ensuring text and visual chart timestamps match local system time.

### 3. Execution Loop Resiliency (`cmd/netmon`)
- Wrap each hourly cycle execution in `cmd/netmon/main.go` in an error-recovery block (`defer func() { if r := recover(); r != nil { ... } }()` & error checking).
- Ensure transient network timeouts or API failures (speedtest failure, nmap error, Discord/Telegram timeout) log the error cleanly and allow the daemon ticker to proceed to the next cycle without crash-looping.

## Verification Plan
1. **Unit Tests**:
   - `internal/config`: Test `NOTIFIER=discord`, `REQUEST_TIMEOUT` parsing & defaults.
   - `internal/discord`: Test mock webhook HTTP payload generation & multipart uploads.
   - `internal/notifier`: Interface compatibility tests for Telegram & Discord implementations.
2. **Build & Quality**:
   - Run `make test` & `make lint` across all packages.
   - Verify static binary compilation (`make build`).
