package config

import (
	"os"
	"path/filepath"
	"testing"
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
