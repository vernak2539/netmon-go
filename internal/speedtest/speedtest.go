package speedtest

import (
	"context"
	"fmt"
	"strings"

	gospeedtest "github.com/showwin/speedtest-go/speedtest"
	"github.com/vernak2539/netmon-go/internal/models"
)

type Runner struct{}

func New() *Runner {
	return &Runner{}
}

func mapResult(downloadMbps, uploadMbps, pingMs float64, shareURL, isp, serverName string, bytesSent, bytesReceived int64) (*models.NetworkMetric, error) {
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

func (r *Runner) Run(ctx context.Context) (*models.NetworkMetric, error) {
	client := gospeedtest.New()
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

	bytesReceived := int64(s.DLSpeed.Mbps() * 1000000 / 8 * 10)
	bytesSent := int64(s.ULSpeed.Mbps() * 1000000 / 8 * 10)

	return mapResult(
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
