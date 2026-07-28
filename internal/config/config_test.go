package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	// Helper to clear env vars after test
	cleanup := func() {
		os.Unsetenv("AI_API_KEY")
		os.Unsetenv("AI_MODEL")
		os.Unsetenv("AI_BASE_URL")
		os.Unsetenv("TG_BOT_TOKEN")
		os.Unsetenv("TG_CHAT_ID")
		os.Unsetenv("DB_PATH")
		os.Unsetenv("NOTIFIER")
		os.Unsetenv("DISCORD_WEBHOOK_URL")
		os.Unsetenv("REQUEST_TIMEOUT")
	}
	defer cleanup()

	t.Run("Missing variables error", func(t *testing.T) {
		cleanup()
		_, err := Load()
		if err == nil {
			t.Error("expected error due to missing config, got nil")
		}
	})

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
		if cfg.Notifier != "telegram" {
			t.Errorf("expected default Notifier to be telegram, got %s", cfg.Notifier)
		}
		if cfg.RequestTimeout != 30*time.Second {
			t.Errorf("expected default RequestTimeout to be 30s, got %v", cfg.RequestTimeout)
		}
	})

	t.Run("Valid loading from environment", func(t *testing.T) {
		cleanup()
		os.Setenv("AI_API_KEY", "test-key")
		os.Setenv("AI_MODEL", "test-model")
		os.Setenv("AI_BASE_URL", "http://localhost:8000")
		os.Setenv("TG_BOT_TOKEN", "test-token")
		os.Setenv("TG_CHAT_ID", "test-chat")
		os.Setenv("DB_PATH", "test.db")

		// Use default .env path which is skipped if not found
		*envFile = ".env"

		cfg, err := Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.AIAPIKey != "test-key" {
			t.Errorf("expected AI_API_KEY to be test-key, got %s", cfg.AIAPIKey)
		}
		if cfg.AIModel != "test-model" {
			t.Errorf("expected AI_MODEL to be test-model, got %s", cfg.AIModel)
		}
		if cfg.AIBaseURL != "http://localhost:8000" {
			t.Errorf("expected AI_BASE_URL to be http://localhost:8000, got %s", cfg.AIBaseURL)
		}
		if cfg.TGBotToken != "test-token" {
			t.Errorf("expected TG_BOT_TOKEN to be test-token, got %s", cfg.TGBotToken)
		}
		if cfg.TGChatID != "test-chat" {
			t.Errorf("expected TG_CHAT_ID to be test-chat, got %s", cfg.TGChatID)
		}
		if cfg.DBPath != "test.db" {
			t.Errorf("expected DB_PATH to be test.db, got %s", cfg.DBPath)
		}
		if cfg.Notifier != "telegram" {
			t.Errorf("expected default Notifier to be telegram, got %s", cfg.Notifier)
		}
		if cfg.RequestTimeout != 30*time.Second {
			t.Errorf("expected default RequestTimeout to be 30s, got %v", cfg.RequestTimeout)
		}
	})

	t.Run("Valid loading with NOTIFIER=discord", func(t *testing.T) {
		cleanup()
		os.Setenv("NOTIFIER", "discord")
		os.Setenv("DISCORD_WEBHOOK_URL", "https://discord.com/api/webhooks/123/abc")
		os.Setenv("DB_PATH", "test.db")

		*envFile = ".env"

		cfg, err := Load()
		if err != nil {
			t.Fatalf("unexpected error when NOTIFIER=discord: %v", err)
		}

		if cfg.Notifier != "discord" {
			t.Errorf("expected Notifier to be discord, got %s", cfg.Notifier)
		}
		if cfg.DiscordWebhookURL != "https://discord.com/api/webhooks/123/abc" {
			t.Errorf("expected DiscordWebhookURL to be https://discord.com/api/webhooks/123/abc, got %s", cfg.DiscordWebhookURL)
		}
	})

	t.Run("Missing DISCORD_WEBHOOK_URL when NOTIFIER=discord", func(t *testing.T) {
		cleanup()
		os.Setenv("NOTIFIER", "discord")
		os.Setenv("DB_PATH", "test.db")

		*envFile = ".env"

		_, err := Load()
		if err == nil {
			t.Error("expected error when DISCORD_WEBHOOK_URL is missing for discord notifier, got nil")
		}
	})

	t.Run("Invalid NOTIFIER value", func(t *testing.T) {
		cleanup()
		os.Setenv("NOTIFIER", "slack")
		os.Setenv("DB_PATH", "test.db")

		*envFile = ".env"

		_, err := Load()
		if err == nil {
			t.Error("expected error when NOTIFIER is invalid, got nil")
		}
	})

	t.Run("Custom valid REQUEST_TIMEOUT", func(t *testing.T) {
		cleanup()
		os.Setenv("TG_BOT_TOKEN", "test-token")
		os.Setenv("TG_CHAT_ID", "test-chat")
		os.Setenv("DB_PATH", "test.db")
		os.Setenv("REQUEST_TIMEOUT", "15")

		*envFile = ".env"

		cfg, err := Load()
		if err != nil {
			t.Fatalf("unexpected error with custom REQUEST_TIMEOUT: %v", err)
		}

		if cfg.RequestTimeout != 15*time.Second {
			t.Errorf("expected RequestTimeout to be 15s, got %v", cfg.RequestTimeout)
		}
	})

	t.Run("Invalid REQUEST_TIMEOUT", func(t *testing.T) {
		cleanup()
		os.Setenv("TG_BOT_TOKEN", "test-token")
		os.Setenv("TG_CHAT_ID", "test-chat")
		os.Setenv("DB_PATH", "test.db")
		os.Setenv("REQUEST_TIMEOUT", "invalid")

		*envFile = ".env"

		_, err := Load()
		if err == nil {
			t.Error("expected error when REQUEST_TIMEOUT is non-numeric, got nil")
		}
	})

	t.Run("Valid loading from dotenv file", func(t *testing.T) {
		cleanup()

		// Create temp .env file
		tmpDir, err := os.MkdirTemp("", "configtest")
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		dotenvPath := filepath.Join(tmpDir, ".env")
		content := `
AI_API_KEY=file-key
AI_MODEL=file-model
AI_BASE_URL=http://file-host
TG_BOT_TOKEN=file-token
TG_CHAT_ID=file-chat
DB_PATH=file.db
`
		if err := os.WriteFile(dotenvPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write temp env file: %v", err)
		}

		*envFile = dotenvPath
		cfg, err := Load()
		if err != nil {
			t.Fatalf("unexpected error loading from file: %v", err)
		}

		if cfg.AIAPIKey != "file-key" {
			t.Errorf("expected file-key, got %s", cfg.AIAPIKey)
		}
		if cfg.DBPath != "file.db" {
			t.Errorf("expected file.db, got %s", cfg.DBPath)
		}
	})
}
