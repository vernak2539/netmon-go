# Runner Implementation Plan (Native Go Speedtest & nmap)

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement `internal/runner/runner.go` to execute network speedtests using the native Go library `showwin/speedtest-go` and LAN device scans using `nmap`.

**Architecture:** A `Runner` struct that runs multi-threaded speed tests directly via Go API calls and executes `nmap` subnet ARP scans via `os/exec`. Results are mapped into domain models (`models.NetworkMetric`, `models.NetworkDevice`).

**Tech Stack:** Go stdlib (`os/exec`, `encoding/xml`, `regexp`), `github.com/showwin/speedtest-go/speedtest`, `github.com/vernak2539/netmon-go/internal/models`.

## Global Constraints

- Must map latency, download (Mbps/bits), upload (Mbps/bits), bytes sent/received, ISP client info, and server name into `models.NetworkMetric`.
- Ping > 1000ms is set to 0; empty share URL is set to "N/A".
- Parse `nmap --iflist` using regex `(\d+\.\d+\.\d+\.\d+/\d+)\s+ethernet\s+up` to detect local subnet automatically.
- Parse `nmap -sn -oX - <subnet>` XML output for host IP address and `srtt` latency.
- Provide clean unit tests using mocked test fixtures and interfaces.

---

## Evaluation of SmokePing (Issue #12)

> [!NOTE]
> **Evaluation Outcome for Issue #12:**
> 1. **Not a replacement for speed tests:** SmokePing measures continuous latency/jitter/packet loss via ICMP/FPing, but does **not** test bandwidth (download/upload Mbps).
> 2. **Heavy External Dependency:** SmokePing requires Perl, `rrdtool`, and a background daemon.
> 3. **Determination:** We use native Go `showwin/speedtest-go` for throughput metrics. Comment added to Issue #12 detailing these findings.

---

## Proposed File Structure

- Create: `internal/runner/runner.go` — Speedtest & nmap execution runner
- Create: `internal/runner/runner_test.go` — Unit tests for speedtest mapping & nmap XML/subnet parsing

---

### Task 1: Implement Native Go SpeedTest Runner (`internal/runner`)

**Files:**
- Create: `internal/runner/runner.go`
- Test: `internal/runner/runner_test.go`

**Interfaces:**
- Consumes: `models.NewNetworkMetric()`, `speedtest.FetchServers()`
- Produces: `RunSpeedtest(ctx context.Context) (*models.NetworkMetric, error)`

- [ ] **Step 1: Write the failing test for speedtest result mapping**

```go
package runner

import (
	"testing"
)

func TestMapSpeedtestResult(t *testing.T) {
	// Test mapping raw numbers into models.NetworkMetric
	downloadBytes := int64(150000000)
	uploadBytes := int64(45000000)
	downloadMbps := 150.23
	uploadMbps := 45.12
	latencyMs := 12.5
	isp := "TestISP"
	serverName := "TestServer"
	shareURL := "http://speedtest/share.png"

	metric, err := mapSpeedtestResult(downloadMbps, uploadMbps, latencyMs, shareURL, isp, serverName, uploadBytes, downloadBytes)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if metric.Download != downloadMbps || metric.Ping != latencyMs || metric.Client != isp {
		t.Errorf("unexpected metric values: %+v", metric)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/runner/...`
Expected: FAIL with "mapSpeedtestResult undefined"

- [ ] **Step 3: Implement `RunSpeedtest` and result mapping in `internal/runner/runner.go`**

```go
package runner

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/showwin/speedtest-go/speedtest"
	"github.com/vernak2539/netmon-go/internal/models"
)

type Runner struct{}

func New() *Runner {
	return &Runner{}
}

func mapSpeedtestResult(downloadMbps, uploadMbps, pingMs float64, shareURL, isp, serverName string, bytesSent, bytesReceived int64) (*models.NetworkMetric, error) {
	if pingMs >= 1000 {
		pingMs = 0
	}

	share := "N/A"
	if strings.TrimSpace(shareURL) != "" {
		share = shareURL
	}

	client := isp
	if strings.TrimSpace(client) == "" {
		client = "Unknown ISP"
	}

	server := serverName
	if strings.TrimSpace(server) == "" {
		server = "Unknown Server"
	}

	return models.NewNetworkMetric(
		downloadMbps,
		uploadMbps,
		pingMs,
		share,
		client,
		server,
		bytesSent,
		bytesReceived,
	)
}

func (r *Runner) RunSpeedtest(ctx context.Context) (*models.NetworkMetric, error) {
	user, err := speedtest.FetchUserInfo()
	isp := "Unknown ISP"
	if err == nil && user != nil {
		isp = user.Isp
	}

	serverList, err := speedtest.FetchServers()
	if err != nil {
		return nil, fmt.Errorf("fetching speedtest servers: %w", err)
	}

	targets, err := serverList.FindServer([]int{})
	if err != nil || len(targets) == 0 {
		return nil, fmt.Errorf("no speedtest servers found")
	}

	s := targets[0]

	if err := s.PingTestContext(ctx, nil); err != nil {
		return nil, fmt.Errorf("ping test failed: %w", err)
	}

	if err := s.DownloadTestContext(ctx); err != nil {
		return nil, fmt.Errorf("download test failed: %w", err)
	}

	if err := s.UploadTestContext(ctx); err != nil {
		return nil, fmt.Errorf("upload test failed: %w", err)
	}

	// Speedtest-go converts throughput to Mbps (float64) in s.DLSpeed and s.ULSpeed
	return mapSpeedtestResult(
		float64(s.DLSpeed),
		float64(s.ULSpeed),
		float64(s.Latency.Milliseconds()),
		s.Context.Result.Share,
		isp,
		s.Name,
		int64(s.Context.Result.BytesSent),
		int64(s.Context.Result.BytesReceived),
	)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `go test -v ./internal/runner/...`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/runner/
git commit -m "feat: implement native Go speedtest runner using showwin/speedtest-go"
```

---

### Task 2: Implement Device Scan Runner (`internal/runner`)

**Files:**
- Modify: `internal/runner/runner.go`
- Test: `internal/runner/runner_test.go`

**Interfaces:**
- Consumes: `models.NewNetworkDevice()`
- Produces: `RunDevicesScan(ctx context.Context) ([]models.NetworkDevice, error)`

- [ ] **Step 1: Write failing test for subnet regex and nmap XML parsing**

```go
func TestParseSubnet(t *testing.T) {
	iflist := `
Starting Nmap
INTERFACES:
DEV    TYPE     IP/MASK         MAC
en0    ethernet 192.168.1.5/24  AA:BB:CC:DD:EE:FF
	`
	subnet, err := parseSubnet(iflist)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if subnet != "192.168.1.5/24" {
		t.Errorf("expected 192.168.1.5/24, got %s", subnet)
	}
}

func TestParseNmapXML(t *testing.T) {
	xmlSample := `<?xml version="1.0"?>
<nmaprun>
  <host>
    <status state="up"/>
    <address addr="192.168.1.1" addrtype="ipv4"/>
    <times srtt="15000"/>
  </host>
</nmaprun>`

	devices, err := parseNmapXML([]byte(xmlSample))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(devices) != 1 || devices[0].IP != "192.168.1.1" || devices[0].LatencyMs != 15.0 {
		t.Errorf("unexpected devices parsed: %+v", devices)
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test -v ./internal/runner/...`
Expected: FAIL with undefined `parseSubnet` and `parseNmapXML`

- [ ] **Step 3: Implement subnet matching and XML parsing in `internal/runner/runner.go`**

```go
var ethernetSubnetRegex = regexp.MustCompile(`(\d+\.\d+\.\d+\.\d+/\d+)\s+ethernet\s+up`)

type nmapTimes struct {
	SRTT string `xml:"srtt,attr"`
}

type nmapStatus struct {
	State string `xml:"state,attr"`
}

type nmapAddress struct {
	Addr     string `xml:"addr,attr"`
	AddrType string `xml:"addrtype,attr"`
}

type nmapHost struct {
	Status  nmapStatus  `xml:"status"`
	Address nmapAddress `xml:"address"`
	Times   *nmapTimes  `xml:"times"`
}

type nmapRun struct {
	Hosts []nmapHost `xml:"host"`
}

func parseSubnet(iflist string) (string, error) {
	match := ethernetSubnetRegex.FindStringSubmatch(iflist)
	if len(match) < 2 {
		return "", fmt.Errorf("no active ethernet interface found via nmap --iflist")
	}
	return match[1], nil
}

func parseNmapXML(data []byte) ([]models.NetworkDevice, error) {
	var run nmapRun
	if err := xml.Unmarshal(data, &run); err != nil {
		return nil, fmt.Errorf("parsing nmap XML: %w", err)
	}

	var devices []models.NetworkDevice
	for _, host := range run.Hosts {
		if host.Status.State != "up" || strings.TrimSpace(host.Address.Addr) == "" {
			continue
		}

		var latency float64
		if host.Times != nil && host.Times.SRTT != "" {
			var srtt int64
			if _, err := fmt.Sscanf(host.Times.SRTT, "%d", &srtt); err == nil {
				latency = float64(srtt) / 1000.0
			}
		}

		dev, err := models.NewNetworkDevice(host.Address.Addr, latency)
		if err == nil {
			devices = append(devices, *dev)
		}
	}

	return devices, nil
}

func (r *Runner) RunDevicesScan(ctx context.Context) ([]models.NetworkDevice, error) {
	iflistCmd := exec.CommandContext(ctx, "nmap", "--iflist")
	iflistOut, err := iflistCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("executing nmap --iflist: %w", err)
	}

	subnet, err := parseSubnet(string(iflistOut))
	if err != nil {
		return nil, err
	}

	scanCmd := exec.CommandContext(ctx, "sudo", "-n", "nmap", "-sn", "-oX", "-", subnet)
	xmlOut, err := scanCmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("nmap scan failed: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("executing nmap scan: %w", err)
	}

	return parseNmapXML(xmlOut)
}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `make test`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/runner/
git commit -m "feat: implement nmap device scan runner and XML parser"
```
