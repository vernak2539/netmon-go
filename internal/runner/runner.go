package runner

import (
	"context"
	"encoding/xml"
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
	client := speedtest.New()
	user, err := client.FetchUserInfo()
	isp := "Unknown ISP"
	if err == nil && user != nil {
		isp = user.Isp
	}

	serverList, err := client.FetchServers()
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

	// Calculate bytes sent/received based on speed and test duration
	bytesReceived := int64(s.DLSpeed.Mbps() * 1000000 / 8 * 10)
	bytesSent := int64(s.ULSpeed.Mbps() * 1000000 / 8 * 10)

	return mapSpeedtestResult(
		s.DLSpeed.Mbps(),
		s.ULSpeed.Mbps(),
		float64(s.Latency.Milliseconds()),
		s.URL,
		isp,
		s.Name,
		bytesSent,
		bytesReceived,
	)
}

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
