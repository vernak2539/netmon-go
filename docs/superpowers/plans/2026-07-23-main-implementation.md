# Main Entrypoint & Execution Loop Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement `cmd/netmon/main.go` to wire all components together into the main execution loop with signal handling, hourly speedtest/scan cycles, mini status reports, and 4-hourly AI-generated reports with graph image uploads.

**Architecture:** A standalone Go CLI binary entrypoint that loads `config`, initializes `db`, `ai`, `telegram`, and `runner` (or `speedtest` + `scanner`), and executes an hourly loop managed via Go `context` and `time.Ticker` for clean signal shutdown.

**Tech Stack:** Go stdlib (`context`, `flag`, `fmt`, `log`, `os`, `os/signal`, `strings`, `syscall`, `time`), `github.com/vernak2539/netmon-go/internal/...`.

## Global Constraints

- Must match Python `main.py` logic, status texts, prompt constants, and HTML formatting verbatim.
- Strip all `<br>`, `<br/>`, `<br />` tags from AI response using `strings.ReplaceAll`.
- Support graceful shutdown on `SIGINT` / `SIGTERM` using `signal.NotifyContext`.
- Verify build: `go build -o netmon ./cmd/netmon`.

---

## Proposed File Structure

- Create: `cmd/netmon/main.go` — Main entrypoint
- Create: `cmd/netmon/main_test.go` — Unit test verifying report text formatting and status text determination

---

### Task 1: Implement Report Formatting & Helper Logic (`cmd/netmon`)

**Files:**
- Create: `cmd/netmon/main.go`
- Test: `cmd/netmon/main_test.go`

**Interfaces:**
- Consumes: `models.NetworkMetric`, `models.NetworkDevice`
- Produces: `determineStatusText(dlSpeed, ping float64) string`, `formatMiniReport(...) string`, `formatUserReportItem(...) string`

- [ ] **Step 1: Write failing tests for status text and report formatters**

```go
package main

import (
	"testing"
)

func TestDetermineStatusText(t *testing.T) {
	t.Run("Good speed and low latency", func(t *testing.T) {
		status := determineStatusText(160.0, 15.0)
		if status != "Good speed and low latency" {
			t.Errorf("expected Good speed..., got %s", status)
		}
	})

	t.Run("Bad speed or high latency", func(t *testing.T) {
		status := determineStatusText(40.0, 15.0)
		if status != "A bunch of idiots decided to stream 4K movies all at once, or the ISP's mice were busy chewing on the fiber line again, whatever" {
			t.Errorf("unexpected status: %s", status)
		}

		statusPing := determineStatusText(100.0, 50.0)
		if statusPing != "A bunch of idiots decided to stream 4K movies all at once, or the ISP's mice were busy chewing on the fiber line again, whatever" {
			t.Errorf("unexpected status: %s", statusPing)
		}
	})

	t.Run("Average speed fallback", func(t *testing.T) {
		status := determineStatusText(100.0, 30.0)
		if status != "At least it works, I guess" {
			t.Errorf("expected 'At least it works, I guess', got %s", status)
		}
	})
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -v ./cmd/netmon/...`
Expected: FAIL with "determineStatusText undefined"

- [ ] **Step 3: Implement `determineStatusText` and string formatters in `cmd/netmon/main.go`**

```go
package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/vernak2539/netmon-go/internal/models"
)

const ReportSystemPrompt = `You are a sarcastic, cynical network analyst bot. Your job is to output a short network speed test and 24-hour trend report in Telegram HTML format.
You will receive a list of speed tests from the last 24 hours in chronological order (the last line is the latest test).

You must write the report in ENGLISH.
You must follow the EXACT structure below. Do not deviate from this layout, header naming, or formatting.

EXPECTED STRUCTURE:
<b>Network Speed Test Report (24h Analysis)</b>

Client: <b>[Client ISP]</b>
Server: <b>[Server Name]</b>

<b>Latest Test Metrics</b>
<pre>
Download: [Download Speed] Mbps
Upload: [Upload Speed] Mbps
Ping: [Ping Latency] ms
Devices Online: [Device Count]
</pre>

<b>24-Hour Dynamics Analysis</b>
[Analyze the dynamics, drops, and load of the network over the last 24 hours. Note any major drops in download/upload speeds or ping spikes.
Also look at how the device count changed over the same period. ONLY claim a link between device count and speed/latency swings if the numbers actually move together (e.g. speed visibly drops in the same window device count rises). If device count swings around while speed/ping stay flat, say plainly that device count does NOT explain it this period, and point at the ISP/line instead. Never invent a correlation that isn't supported by the numbers.
If ping reads exactly 0.00 ms while download speed is very low (a few Mbps or less), do NOT describe that as a good/perfect ping. That reading means the real ping was too high to register and got floored to zero — call it a red flag, not a strength.
Use a sarcastic, informal tone when describing speed drops, latency spikes, or a sudden herd of new devices, blaming heavy users/leeches on the network or the ISP (e.g. "a bunch of idiots clogging the bandwidth", "ISP dropping the ball", "mice chewing the optic fiber cables", or "yet another gadget joining the freeloader party") — but only when the data actually supports that story.
CRITICAL: Do NOT blame server changes for fluctuations. Assume the server choice is optimal and fluctuations reflect real network load, device count, or ISP issues.
Wrap key numbers in <code> tags, e.g., <code>148.31 Mbps</code>, <code>15.18 ms</code>, or <code>7 devices</code>.]

<b>Data Transfer (Latest Test)</b>
<pre>
Downloaded: [Downloaded MB] MB
Uploaded: [Uploaded MB] MB
</pre>

<b>Conclusion</b>
[A sarcastic, witty 1 short sentence summary of the network's overall quality and reliability over the past day.]


TEMPLATE EXAMPLE OF THE OUTPUT:
<b>Network Speed Test Report (24h Analysis)</b>

Client: <b>nameserver</b>
Server: <b>New York</b>

<b>Latest Test Metrics</b>
<pre>
Download: 140.3 Mbps
Upload: 62.8 Mbps
Ping: 15.2 ms
Devices Online: 7
</pre>

<b>24-Hour Dynamics Analysis</b>
Over the last 24 hours, the download speed averaged <code>140 Mbps</code>, but we saw a massive drop to <code>20 Mbps</code> at 8:00 PM right as device count jumped from <code>4</code> to <code>11 devices</code>. Clearly, a bunch of idiots decided to stream 4K movies all at once, or the ISP's mice were busy chewing on the fiber line again. Latency remained stable except for a brief spike to <code>95 ms</code> during the speed dip.

<b>Data Transfer (Latest Test)</b>
<pre>
Downloaded: 160.0 MB
Uploaded: 70.0 MB
</pre>

<b>Conclusion</b>
Expect periodic speed deaths whenever the local leechers wake up or the ISP fails to maintain their potato infrastructure.


CRITICAL RULES:
1. Do NOT use <br> or <br/> tags. For line breaks, use normal newlines.
2. The entire report must be in English.
3. Keep the "24-Hour Dynamics Analysis" to exactly 2-3 short sentences.
4. Do NOT write any description text below the "Data Transfer (Latest Test)" pre-block.
5. Keep the "Conclusion" to exactly 1 short sentence.
6. Highlight all numeric metric values in the text using <code>[Value]</code>.
7. Do NOT output any markdown blocks like ```html. Output raw HTML tags directly.
8. Make sure all HTML tags are closed correctly.
9. Be sarcastic, informal, and funny when describing performance dips or network load.
10. The entire output MUST be under 800 characters to ensure it easily fits within Telegram limits.`

const MiniReportTemplate = `<b>Network Status Update</b>
Here is the latest snapshot of your internet speed:

Time: <b>%s</b>
ISP: <b>%s</b> | Server: <b>%s</b>

Devices online: <b>%d</b>

Download: <b>%.1f Mbps</b>
Upload: <b>%.1f Mbps</b>
Latency: <b>%.1f ms</b>

Traffic used: <b>%.1f MB</b> down / <b>%.1f MB</b> up

<b>Current status:</b> %s`

const ReportUserItemFormat = `Network speed test results:
- Date: %s
- Download: %.2f Mbps
- Upload: %.2f Mbps
- Ping: %.2f ms
- Client: %s
- Server: %s
- Downloaded: %.1f MB
- Uploaded: %.1f MB
- Share Link: %s
- Devices online: %d
`

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
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -v ./cmd/netmon/...`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add cmd/netmon/
git commit -m "feat: add report templates and status text logic"
```

---

### Task 2: Implement Main Loop & Signal Handling (`cmd/netmon`)

**Files:**
- Modify: `cmd/netmon/main.go`

**Interfaces:**
- Consumes: `config.Load()`, `db.Open()`, `ai.New()`, `telegram.New()`, `speedtest.New()`, `scanner.New()`, `graphs.Plot()`
- Produces: `main()` binary entrypoint

- [ ] **Step 1: Implement `main()` function in `cmd/netmon/main.go`**

```go
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
```

- [ ] **Step 2: Test building the binary**

Run: `make build`
Expected: Successfully compiles binary `netmon`

- [ ] **Step 3: Commit**

```bash
git add cmd/netmon/
git commit -m "feat: implement main entrypoint and hourly execution loop"
```
