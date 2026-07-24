<p align="center">
  <img src="assets/logo.png" alt="netmon logo" width="180" />
</p>

<h1 align="center">netmon</h1>

<p align="center">
  <b>Self-hosted local network monitor with 24-hour speed charts & sarcastic AI commentary delivered straight to Telegram.</b>
</p>

<p align="center">
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-8bc34a?style=for-the-badge" alt="License MIT"></a>
  <img src="https://img.shields.io/badge/Go-1.23+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/Telegram-Bot_API-26A5E4?style=for-the-badge&logo=telegram&logoColor=white" alt="Telegram">
  <img src="https://img.shields.io/badge/SQLite-Storage-003B57?style=for-the-badge&logo=sqlite&logoColor=white" alt="SQLite">
</p>

---

A lightweight local bot that runs a speed test on your network every hour, scans active devices on your LAN using `nmap`, and logs everything to a local SQLite database.

Every 4 hours, it delivers a **detailed report** complete with a 24-hour trend graph and a sarcastic, LLM-generated commentary on your network's behavior (*"someone's hogging the bandwidth again"*).

> [!NOTE]
> **100% Private & Self-Hosted:** No external metric servers involved — everything runs locally on your machine or Raspberry Pi. Only text reports and graph images are dispatched to your Telegram chat.

---

## Acknowledgements

This project is a Go rewrite of [**netmon**](https://github.com/Role1776/netmon) by [@Role1776](https://github.com/Role1776), originally written in Python. All credit for the original concept, design, and architecture goes to the original author.

---

## Why Go?

This project was originally written in Python and has been rewritten in Go for several key advantages:

| | Python | Go |
| :--- | :--- | :--- |
| **Deployment** | Requires Python 3.13+, `uv`, virtualenv | Single static binary — just copy and run |
| **Memory** | ~50-80 MB (interpreter + dependencies) | ~10-15 MB |
| **Startup** | ~1-2s (interpreter + imports) | Instant |
| **Cross-compile** | Complex (wheels, platform deps) | `GOOS=linux GOARCH=arm64 go build` |
| **Dependencies** | pip/uv managed, virtualenv required | Vendored in binary, zero runtime deps |
| **Raspberry Pi** | Needs Python installed + venv setup | Copy one binary, done |

> [!TIP]
> **Looking for the Python version?** It lives on the [`main` branch](https://github.com/vernak2539/netmon-go/tree/main) and continues to receive updates.

---

## Features & Workflow

Every hour (`sleepTime` in `main.go`, default 3600 seconds):

1. **Speed Test:** Measures download/upload speeds, ping latency, ISP, and test server details using native Go `speedtest-go`.
2. **LAN Scan:** Scans the local subnet using `nmap` ARP scan to count active connected devices.
3. **Local Storage:** Saves metrics & device tallies directly to a local `metrics.sql` SQLite database.
4. **Status Alert:** Sends a concise status update to Telegram (*"all good"* or *"line is dying"*).
5. **24h AI Report:** Every 4th cycle (every 4h), generates a **24-hour trend graph** alongside a sarcastic LLM analysis of network load and speed fluctuations.

---

## Tech Stack

| Technology | Purpose |
| :--- | :--- |
| **Go 1.23+** | Core runtime |
| **SQLite** (`modernc.org/sqlite`) | Local metrics persistence (`metrics.sql`) — pure Go, no cgo |
| **`speedtest-go`** (`showwin/speedtest-go`) | Network bandwidth and ping measurements (pure Go, zero binary deps) |
| **`nmap`** | Subnet ARP scanning for device discovery |
| **`go-analyze/charts`** | 24-hour metrics visualization (pure Go PNG rendering) |
| **OpenAI-compatible API** (`sashabaranov/go-openai`) | Sarcastic report & trend analysis |
| **Telegram Bot API** | Alert and graph report delivery |

---

## Requirements

* **OS:** macOS or Linux (`nmap --iflist` required; Windows not supported out of the box).
* **Go 1.23+** — only needed to build from source. Pre-built binaries are available on the [Releases](https://github.com/vernak2539/netmon-go/releases) page.
* **System Binaries:** `nmap` installed system-wide.
* **Passwordless `sudo` for `nmap`** — device counting needs a real ARP scan (raw sockets), which requires root; see one-time setup below.
* **Tokens:** Telegram Bot Token, Telegram Chat ID, and an API key for your OpenAI-compatible provider (not needed if you point `AI_BASE_URL` at a local LLM server).

---

## Quick Start

### 1. System Dependencies

**macOS (Homebrew):**
```bash
brew install nmap
```

**Linux (Debian/Ubuntu):**
```bash
sudo apt update && sudo apt install -y nmap
```

### 2. Allow Passwordless `nmap` (one-time)

Device counting runs `nmap` as root for a real ARP scan — without it, host discovery silently falls back to ordinary TCP probing and undercounts devices that don't answer on common ports. Since the bot runs unattended, `sudo` needs to work without a password prompt on every cycle:

```bash
echo "$(whoami) ALL=(root) NOPASSWD: $(command -v nmap)" | sudo tee /etc/sudoers.d/netmon-nmap
sudo chmod 440 /etc/sudoers.d/netmon-nmap
```

This grants passwordless `sudo` only for the `nmap` binary — not your whole account.

### 3. Install

**Option A: Download a pre-built binary** (recommended)

Download the latest release for your platform from the [Releases](https://github.com/vernak2539/netmon-go/releases) page:

```bash
# Example for Linux ARM64 (Raspberry Pi)
curl -LO https://github.com/vernak2539/netmon-go/releases/latest/download/netmon-linux-arm64
chmod +x netmon-linux-arm64
mv netmon-linux-arm64 /usr/local/bin/netmon
```

**Option B: Build from source**

```bash
git clone https://github.com/vernak2539/netmon-go.git
cd netmon-go
git checkout main-go
make build
```

### 4. Configure `.env`

Copy the template file and fill in your secrets:

```bash
cp .env.example .env
```

`.env` variables:

| Variable | Description |
| :--- | :--- |
| `AI_API_KEY` | Your LLM provider API key (any string works for most local servers) |
| `AI_MODEL` | Model name (e.g. `gpt-4o-mini`, or a local model name — see below) |
| `AI_BASE_URL` | Base API URL (e.g., `https://api.openai.com/v1`, or your local server's URL) |
| `TG_BOT_TOKEN` | Telegram bot token from `@BotFather` |
| `TG_CHAT_ID` | Your Telegram Chat ID |
| `DB_PATH` | SQLite database file path (e.g. `metrics.sql`) |

> [!TIP]
> **You're not locked into OpenAI.** The AI client talks to any OpenAI-compatible endpoint, so a local inference server (e.g. [Ollama](https://ollama.com), LM Studio) works too — just point `AI_BASE_URL` at it.

### 5. Run the Bot

```bash
./netmon
```

Or with a custom env file:

```bash
./netmon --env /path/to/my.env
```

> [!TIP]
> Run the bot inside `tmux`/`screen` or set it up as a system service (`systemd`/`launchd`) to keep it running 24/7 in the background.

---

## Example Output

### Hourly Short Status Update

```text
Network Status Update
Time: 2026-07-21 14:00:00
ISP: MyISP | Server: New York

Devices online: 7
Download: 145.2 Mbps
Upload: 62.1 Mbps
Latency: 14.8 ms

Traffic used: 160.0 MB down / 70.0 MB up

Current status: Good speed and low latency
```

### 4-Hour Detailed Report (With Graph & AI Analysis)

Every 4 hours, the bot sends a **24-hour graph** accompanied by a sarcastic LLM-generated report:

<p align="center">
  <img src="assets/example_graph.png" alt="24h Network Speed Test Graph" width="650" />
</p>

```html
<b>Network Speed Test Report (24h Analysis)</b>

Client: <b>MyISP</b>
Server: <b>New York</b>

<b>Latest Test Metrics</b>
<pre>
Download: 178.5 Mbps
Upload: 45.2 Mbps
Ping: 23.1 ms
Devices Online: 9
</pre>

<b>24-Hour Dynamics Analysis</b>
Over the last 24 hours, the download speed averaged <code>140 Mbps</code>, but we saw a massive drop to <code>20 Mbps</code> at 8:00 PM right as device count jumped from <code>4</code> to <code>11 devices</code>. Clearly, someone's hogging the bandwidth or the ISP's mice were busy chewing on the fiber line again. Latency remained stable except for a brief spike during peak hours.

<b>Data Transfer (Latest Test)</b>
<pre>
Downloaded: 160.0 MB
Uploaded: 70.0 MB
</pre>

<b>Conclusion</b>
Expect periodic speed drops whenever local freeloaders stream 4K movies or the ISP potato infrastructure struggles.
```

---

## Project Structure

```text
netmon-go/
├── .github/workflows/  # CI (build+test) and Release (cross-compile) workflows
├── cmd/netmon/         # Go entrypoint (main.go)
├── internal/           # Go internal packages
│   ├── ai/             # OpenAI-compatible LLM client
│   ├── config/         # Environment config loader
│   ├── db/             # SQLite database layer
│   ├── graphs/         # Chart rendering (PNG)
│   ├── models/         # Domain data models
│   ├── runner/         # Speedtest + nmap execution
│   └── telegram/       # Telegram bot dispatch
├── assets/             # Logo & documentation media
├── graphs/             # Generated 24h graph images
├── .env.example        # Environment variable template
├── go.mod              # Go module definition
├── go.sum              # Dependency checksums
└── LICENSE             # MIT License
```

---

## License

Distributed under the **MIT License**. See [`LICENSE`](LICENSE) for more details.
