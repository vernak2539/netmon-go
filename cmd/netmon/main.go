package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/vernak2539/netmon-go/internal/ai"
	"github.com/vernak2539/netmon-go/internal/config"
	"github.com/vernak2539/netmon-go/internal/db"
	"github.com/vernak2539/netmon-go/internal/discord"
	"github.com/vernak2539/netmon-go/internal/graphs"
	"github.com/vernak2539/netmon-go/internal/models"
	"github.com/vernak2539/netmon-go/internal/notifier"
	"github.com/vernak2539/netmon-go/internal/scanner"
	"github.com/vernak2539/netmon-go/internal/speedtest"
	"github.com/vernak2539/netmon-go/internal/telegram"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.SetPrefix("[netmon] ")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	var bot notifier.Notifier
	if cfg.Notifier == "discord" {
		var err error
		bot, err = discord.New(cfg.DiscordWebhookURL, cfg.RequestTimeout)
		if err != nil {
			log.Fatalf("Failed to initialize Discord notifier: %v", err)
		}
	} else {
		var err error
		bot, err = telegram.New(cfg.TGBotToken, cfg.TGChatID, cfg.RequestTimeout)
		if err != nil {
			log.Fatalf("Failed to initialize Telegram bot: %v", err)
		}
	}

	database, err := db.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer database.Close()

	aiClient, err := ai.New(cfg.AIAPIKey, cfg.AIModel, cfg.AIBaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize AI client: %v", err)
	}
	if aiClient == nil {
		log.Println("AI API key not provided; AI report generation disabled.")
	}

	speedTester := speedtest.New()
	deviceScanner := scanner.New()

	log.Println("The bot has been started.")

	counter := 0
	ticker := time.NewTicker(3600 * time.Second)
	defer ticker.Stop()

	runCycle := func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Recovered from panic during speedtest cycle: %v", r)
			}
		}()

		log.Println("Starting speedtest cycle...")
		_ = bot.SendChatAction(ctx, notifier.ChatActionTyping)

		metric, err := speedTester.Run(ctx)
		if err != nil {
			log.Printf("Error running speedtest: %v", err)
			return
		}

		devices, err := deviceScanner.Scan(ctx)
		if err != nil {
			log.Printf("Error running device scan: %v", err)
			devices = []models.NetworkDevice{}
		}

		if err := database.AddMetric(ctx, metric); err != nil {
			log.Printf("Error adding metric to database: %v", err)
			return
		}

		scanID, err := database.AddDevices(ctx, devices)
		if err != nil {
			log.Printf("Error adding devices to database: %v", err)
			return
		}

		st, err := models.NewSpeedTest(metric.ID, scanID)
		if err != nil {
			log.Printf("Error creating speedtest model: %v", err)
			return
		}

		if err := database.AddSpeedTest(ctx, st); err != nil {
			log.Printf("Error linking speedtest in database: %v", err)
			return
		}

		log.Printf("Speedtest record added: %s", st.ID)

		if counter >= 4 {
			log.Println("Generating detailed 24h report and graph...")
			metrics, deviceCounts, err := database.GetMetricsWithDeviceCounts(ctx)
			if err != nil {
				log.Printf("Error getting metrics with device counts: %v", err)
				return
			}

			var report string
			if aiClient != nil {
				var userMessage strings.Builder
				for i, m := range metrics {
					devCount := 0
					if i < len(deviceCounts) {
						devCount = deviceCounts[i]
					}
					userMessage.WriteString(fmt.Sprintf(
						ReportUserItemFormat,
						m.Timestamp.Local().Format("2006-01-02 15:04:05"),
						m.Download,
						m.Upload,
						m.Ping,
						m.Client,
						m.Server,
						float64(m.BytesReceived)/1000000.0,
						float64(m.BytesSent)/1000000.0,
						m.Share,
						devCount,
					))
					userMessage.WriteString("\n")
				}

				_ = bot.SendChatAction(ctx, notifier.ChatActionTyping)
				var aiErr error
				report, aiErr = aiClient.SendMessage(ctx, userMessage.String(), ReportSystemPrompt)
				if aiErr != nil {
					log.Printf("AI report generation failed: %v, falling back to status message", aiErr)
				} else {
					report = cleanHTMLResponse(report)
				}
			}

			if report == "" {
				if len(metrics) > 0 && len(deviceCounts) > 0 {
					latestMetric := metrics[len(metrics)-1]
					latestDeviceCount := deviceCounts[len(deviceCounts)-1]
					report = formatMiniReport(&latestMetric, latestDeviceCount)
				}
			}

			_ = bot.SendChatAction(ctx, notifier.ChatActionUploadPhoto)
			graphPath, err := graphs.Plot(metrics, deviceCounts)
			if err != nil {
				log.Printf("Error plotting graph: %v", err)
				return
			}

			photoBytes, err := os.ReadFile(graphPath)
			if err != nil {
				log.Printf("Error reading graph PNG: %v", err)
				return
			}

			if err := bot.SendPhoto(ctx, photoBytes, report); err != nil {
				log.Printf("Error sending photo to notifier: %v", err)
				return
			}
			_ = os.Remove(graphPath)

			log.Println("Detailed report has been sent.")
			counter = 0
		} else {
			miniReport := formatMiniReport(metric, len(devices))
			if err := bot.SendMessage(ctx, miniReport); err != nil {
				log.Printf("Error sending mini report: %v", err)
			} else {
				log.Println("Mini report has been sent.")
			}
			counter++
		}
	}

	// Execute first cycle immediately
	runCycle()

	for {
		select {
		case <-ctx.Done():
			log.Println("Received termination signal. Exiting gracefully.")
			return
		case <-ticker.C:
			runCycle()
		}
	}
}
