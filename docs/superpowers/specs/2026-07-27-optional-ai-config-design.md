# Design Spec: Making AI Configuration Optional in netmon-go

**Date:** 2026-07-27  
**Status:** Approved  

## Overview
Currently, `netmon-go` fails at startup if `AI_API_KEY`, `AI_MODEL`, or `AI_BASE_URL` are missing or empty in the environment / `.env` file. This design makes AI configuration optional so that `netmon-go` can run in environment setups where no OpenAI/AI API key is provided.

## Key Decisions
1. **Optional AI Config:** `AI_API_KEY` will no longer be mandatory in `internal/config/config.go`. If `AI_API_KEY` is empty or omitted, `cfg.AIAPIKey` will be empty (`""`) and `config.Load()` will succeed without error.
2. **Conditional Validation for Model & Base URL:** If `AI_API_KEY` is provided, `AI_MODEL` and `AI_BASE_URL` MUST also be set. If `AI_API_KEY` is empty, `AI_MODEL` and `AI_BASE_URL` are optional.
3. **`ai.Client` Handling:** `ai.New` will accept an empty `apiKey` and return `(nil, nil)`. Callers (such as `main.go`) check if `aiClient == nil`.
4. **Fallback behavior during 4-hour cycle:** When `aiClient == nil`, the 4-hour report cycle in `main.go` will send the mini report status text accompanied by the generated graph photo instead of attempting an AI completion.

## Component Changes

### 1. `internal/config/config.go`
- Remove check requiring `cfg.AIAPIKey != ""`.
- If `cfg.AIAPIKey != ""`, enforce that `cfg.AIModel` and `cfg.AIBaseURL` are non-empty.
- Keep `TG_BOT_TOKEN`, `TG_CHAT_ID`, and `DB_PATH` as mandatory required fields.

### 2. `internal/ai/ai.go`
- In `ai.New(apiKey, model, baseURL)`:
  - If `strings.TrimSpace(apiKey) == ""`, return `nil, nil` without error.

### 3. `cmd/netmon/main.go`
- Handle `aiClient == nil` cleanly:
  - Skip AI completion when `aiClient == nil`.
  - On 4-hour detailed report cycles, if `aiClient == nil`, format the mini report status text for the latest metric and send it with the dual Y-axis graph photo via Telegram `send_photo`.
  - If `aiClient != nil`, perform standard AI completion as before.

## Verification Strategy
- Run unit tests (`make test`).
- Verify `config_test.go` covers missing `AI_API_KEY` (passing config) vs missing mandatory Telegram/DB config (erroring).
- Verify `ai_test.go` covers `ai.New("", "", "")` returning `(nil, nil)`.
- Verify binary builds cleanly with `make build`.
