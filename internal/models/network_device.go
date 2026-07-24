package models

import (
	"fmt"
	"strings"
	"time"
)

type NetworkDevice struct {
	IP        string
	LatencyMs float64
	Timestamp time.Time
}

func NewNetworkDevice(ip string, latencyMs float64) (*NetworkDevice, error) {
	if strings.TrimSpace(ip) == "" {
		return nil, fmt.Errorf("IP cannot be empty")
	}
	if latencyMs < 0 {
		return nil, fmt.Errorf("latency must be non-negative")
	}

	return &NetworkDevice{
		IP:        ip,
		LatencyMs: latencyMs,
		Timestamp: time.Now().UTC(),
	}, nil
}
