package config

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	AIAPIKey   string
	AIModel    string
	AIBaseURL  string
	TGBotToken string
	TGChatID   string
	DBPath     string
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

	cfg := &Config{
		AIAPIKey:   os.Getenv("AI_API_KEY"),
		AIModel:    os.Getenv("AI_MODEL"),
		AIBaseURL:  os.Getenv("AI_BASE_URL"),
		TGBotToken: os.Getenv("TG_BOT_TOKEN"),
		TGChatID:   os.Getenv("TG_CHAT_ID"),
		DBPath:     os.Getenv("DB_PATH"),
	}

	if strings.TrimSpace(cfg.AIAPIKey) == "" {
		return nil, fmt.Errorf("AI_API_KEY not found or empty in environment")
	}
	if strings.TrimSpace(cfg.DBPath) == "" {
		return nil, fmt.Errorf("DB_PATH not found or empty in environment")
	}
	if strings.TrimSpace(cfg.AIModel) == "" {
		return nil, fmt.Errorf("AI_MODEL not found or empty in environment")
	}
	if strings.TrimSpace(cfg.AIBaseURL) == "" {
		return nil, fmt.Errorf("AI_BASE_URL not found or empty in environment")
	}
	if strings.TrimSpace(cfg.TGBotToken) == "" {
		return nil, fmt.Errorf("TG_BOT_TOKEN not found or empty in environment")
	}
	if strings.TrimSpace(cfg.TGChatID) == "" {
		return nil, fmt.Errorf("TG_CHAT_ID not found or empty in environment")
	}

	return cfg, nil
}
