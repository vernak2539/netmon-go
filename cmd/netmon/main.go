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
	"github.com/vernak2539/netmon-go/internal/graphs"
	"github.com/vernak2539/netmon-go/internal/models"
	"github.com/vernak2539/netmon-go/internal/scanner"
	"github.com/vernak2539/netmon-go/internal/speedtest"
	"github.com/vernak2539/netmon-go/internal/telegram"
)

func determineStatusText(dlSpeed, ping float64) string {
	if dlSpeed >= 150 && ping <= 20 {
		return "Good speed and low latency"
	}
	if dlSpeed < 60 || ping > 40 {
		return "A bunch of idiots decided to stream 4K movies all at once, or the ISP's mice were busy chewing on the fiber line again, whatever"
	}
	return "At least it works, I guess"
}

func cleanHTMLResponse(text string) string {
	r := strings.ReplaceAll(text, "<br>", "\n")
	r = strings.ReplaceAll(r, "<br/>", "\n")
	r = strings.ReplaceAll(r, "<br />", "\n")
	return r
}

func formatMiniReport(m *models.NetworkMetric, deviceCount int) string {
	timestampStr := m.Timestamp.Format("2006-01-02 15:04:05")
	statusText := determineStatusText(m.Download, m.Ping)
	bytesReceivedMB := float64(m.BytesReceived) / 1000000.0
	bytesSentMB := float64(m.BytesSent) / 1000000.0

	return fmt.Sprintf(
		MiniReportTemplate,
		timestampStr,
		m.Client,
		m.Server,
		deviceCount,
		m.Download,
		m.Upload,
		m.Ping,
		bytesReceivedMB,
		bytesSentMB,
		statusText,
	)
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.SetPrefix("[netmon] ")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	bot, err := telegram.New(cfg.TGBotToken, cfg.TGChatID)
	if err != nil {
		log.Fatalf("Failed to initialize Telegram bot: %v", err)
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

	speedTester := speedtest.New()
	deviceScanner := scanner.New()

	log.Println("The bot has been started.")

	counter := 0
	ticker := time.NewTicker(3600 * time.Second)
	defer ticker.Stop()

	runCycle := func() {
		log.Println("Starting speedtest cycle...")
		_ = bot.SendChatAction(telegram.Typing)

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

			var userMessage strings.Builder
			for i, m := range metrics {
				devCount := 0
				if i < len(deviceCounts) {
					devCount = deviceCounts[i]
				}
				userMessage.WriteString(fmt.Sprintf(
					ReportUserItemFormat,
					m.Timestamp.Format("2006-01-02 15:04:05"),
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

			_ = bot.SendChatAction(telegram.Typing)
			report, err := aiClient.SendMessage(ctx, userMessage.String(), ReportSystemPrompt)
			if err != nil {
				log.Printf("Error generating AI report: %v", err)
				return
			}
			report = cleanHTMLResponse(report)

			_ = bot.SendChatAction(telegram.UploadPhoto)
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

			if err := bot.SendPhoto(photoBytes, report); err != nil {
				log.Printf("Error sending photo to Telegram: %v", err)
				return
			}

			log.Println("Detailed report has been sent.")
			counter = 0
		} else {
			miniReport := formatMiniReport(metric, len(devices))
			if err := bot.SendMessage(miniReport); err != nil {
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
