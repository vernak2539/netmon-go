package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AIAPIKey          string
	AIModel           string
	AIBaseURL         string
	TGBotToken        string
	TGChatID          string
	DBPath            string
	Notifier          string
	DiscordWebhookURL string
	RequestTimeout    time.Duration
}

var envFile = flag.String("env", ".env", "Path to the .env file")

// Load parses command line flags and loads the configuration from the environment/dotenv file.
func Load() (*Config, error) {
	if !flag.Parsed() {
		flag.Parse()
	}

	// Only try to load dotenv if the file exists or a custom env file was specified
	if _, err := os.Stat(*envFile); err == nil {
		if err := godotenv.Load(*envFile); err != nil {
			return nil, fmt.Errorf("loading env file %s: %w", *envFile, err)
		}
	} else if *envFile != ".env" {
		// If a custom env file was requested but doesn't exist, return an error
		return nil, fmt.Errorf("env file not found: %s", *envFile)
	}

	notifier := strings.ToLower(strings.TrimSpace(os.Getenv("NOTIFIER")))
	if notifier == "" {
		notifier = "telegram"
	}
	if notifier != "telegram" && notifier != "discord" {
		return nil, fmt.Errorf("invalid NOTIFIER: %s (must be 'telegram' or 'discord')", notifier)
	}

	requestTimeout := 30 * time.Second
	if rawTimeout := strings.TrimSpace(os.Getenv("REQUEST_TIMEOUT")); rawTimeout != "" {
		sec, err := strconv.Atoi(rawTimeout)
		if err != nil || sec <= 0 {
			return nil, fmt.Errorf("invalid REQUEST_TIMEOUT: %s", rawTimeout)
		}
		requestTimeout = time.Duration(sec) * time.Second
	}

	cfg := &Config{
		AIAPIKey:          os.Getenv("AI_API_KEY"),
		AIModel:           os.Getenv("AI_MODEL"),
		AIBaseURL:         os.Getenv("AI_BASE_URL"),
		TGBotToken:        os.Getenv("TG_BOT_TOKEN"),
		TGChatID:          os.Getenv("TG_CHAT_ID"),
		DBPath:            os.Getenv("DB_PATH"),
		Notifier:          notifier,
		DiscordWebhookURL: os.Getenv("DISCORD_WEBHOOK_URL"),
		RequestTimeout:    requestTimeout,
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

	if cfg.Notifier == "telegram" {
		if strings.TrimSpace(cfg.TGBotToken) == "" {
			return nil, fmt.Errorf("TG_BOT_TOKEN not found or empty in environment")
		}
		if strings.TrimSpace(cfg.TGChatID) == "" {
			return nil, fmt.Errorf("TG_CHAT_ID not found or empty in environment")
		}
	} else if cfg.Notifier == "discord" {
		if strings.TrimSpace(cfg.DiscordWebhookURL) == "" {
			return nil, fmt.Errorf("DISCORD_WEBHOOK_URL not found or empty in environment when NOTIFIER is discord")
		}
	}

	return cfg, nil
}
