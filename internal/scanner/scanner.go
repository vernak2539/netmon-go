package scanner

import (
	"context"
	"encoding/xml"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/vernak2539/netmon-go/internal/models"
)

type DeviceScanner struct{}

func New() *DeviceScanner {
	return &DeviceScanner{}
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

func (s *DeviceScanner) Scan(ctx context.Context) ([]models.NetworkDevice, error) {
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
