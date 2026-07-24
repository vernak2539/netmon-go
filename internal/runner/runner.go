package runner

import (
	"context"

	"github.com/vernak2539/netmon-go/internal/models"
	"github.com/vernak2539/netmon-go/internal/scanner"
	"github.com/vernak2539/netmon-go/internal/speedtest"
)

// Runner acts as a facade/orchestrator for network probes (speedtest & device scanning).
type Runner struct {
	speedtester *speedtest.Runner
	scanner     *scanner.DeviceScanner
}

func New() *Runner {
	return &Runner{
		speedtester: speedtest.New(),
		scanner:     scanner.New(),
	}
}

// RunSpeedtest delegates execution to the dedicated speedtest package.
func (r *Runner) RunSpeedtest(ctx context.Context) (*models.NetworkMetric, error) {
	return r.speedtester.Run(ctx)
}

// RunDevicesScan delegates execution to the dedicated scanner package.
func (r *Runner) RunDevicesScan(ctx context.Context) ([]models.NetworkDevice, error) {
	return r.scanner.Scan(ctx)
}
