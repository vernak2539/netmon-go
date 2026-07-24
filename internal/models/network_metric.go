package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type NetworkMetric struct {
	ID            uuid.UUID
	Download      float64
	Upload        float64
	Ping          float64
	Timestamp     time.Time
	Share         string
	Client        string
	Server        string
	BytesSent     int64
	BytesReceived int64
}

func NewNetworkMetric(download, upload, ping float64, share, client, server string, bytesSent, bytesReceived int64) (*NetworkMetric, error) {
	if download < 0 {
		return nil, fmt.Errorf("download must be non-negative")
	}
	if upload < 0 {
		return nil, fmt.Errorf("upload must be non-negative")
	}
	if ping < 0 {
		return nil, fmt.Errorf("ping must be non-negative")
	}
	if bytesSent < 0 {
		return nil, fmt.Errorf("bytes_sent must be non-negative")
	}
	if bytesReceived < 0 {
		return nil, fmt.Errorf("bytes_received must be non-negative")
	}
	if strings.TrimSpace(client) == "" {
		return nil, fmt.Errorf("client cannot be empty")
	}
	if strings.TrimSpace(server) == "" {
		return nil, fmt.Errorf("server cannot be empty")
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("generating UUID: %w", err)
	}

	return &NetworkMetric{
		ID:            id,
		Download:      download,
		Upload:        upload,
		Ping:          ping,
		Timestamp:     time.Now().UTC(),
		Share:         share,
		Client:        client,
		Server:        server,
		BytesSent:     bytesSent,
		BytesReceived: bytesReceived,
	}, nil
}
