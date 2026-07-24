package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type SpeedTest struct {
	ID           uuid.UUID
	MetricID     uuid.UUID
	DeviceScanID uuid.UUID
	Timestamp    time.Time
}

func NewSpeedTest(metricID, deviceScanID uuid.UUID) (*SpeedTest, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("generating UUID: %w", err)
	}

	return &SpeedTest{
		ID:           id,
		MetricID:     metricID,
		DeviceScanID: deviceScanID,
		Timestamp:    time.Now().UTC(),
	}, nil
}
