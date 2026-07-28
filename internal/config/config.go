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
	SpeedtestInterval time.Duration
	TestNotify        bool
	TestReport        bool
}

var (
	envFile           = flag.String("env", ".env", "Path to the .env file")
	testNotify        = flag.Bool("test-notify", false, "Send a test notification and exit")
	testReport        = flag.Bool("test-report", false, "Send a full 24h test report with graph and exit")
	speedtestInterval = flag.Duration("interval", 0, "Interval between speedtest cycles (e.g. 1h, 30m)")
)

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

	intervalVal := 1 * time.Hour
	if *speedtestInterval > 0 {
		intervalVal = *speedtestInterval
	} else if rawInterval := strings.TrimSpace(os.Getenv("SPEEDTEST_INTERVAL")); rawInterval != "" {
		if dur, err := time.ParseDuration(rawInterval); err == nil && dur > 0 {
			intervalVal = dur
		} else if sec, err := strconv.Atoi(rawInterval); err == nil && sec > 0 {
			intervalVal = time.Duration(sec) * time.Second
		} else {
			return nil, fmt.Errorf("invalid SPEEDTEST_INTERVAL: %s", rawInterval)
		}
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
		SpeedtestInterval: intervalVal,
		TestNotify:        *testNotify,
		TestReport:        *testReport,
	}

	if strings.TrimSpace(cfg.AIAPIKey) != "" {
		if strings.TrimSpace(cfg.AIModel) == "" {
			return nil, fmt.Errorf("AI_MODEL not found or empty in environment when AI_API_KEY is set")
		}
		if strings.TrimSpace(cfg.AIBaseURL) == "" {
			return nil, fmt.Errorf("AI_BASE_URL not found or empty in environment when AI_API_KEY is set")
		}
	}

	if !cfg.TestNotify && strings.TrimSpace(cfg.DBPath) == "" {
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
