package speedtest

import (
	"testing"
)

func TestMapResult(t *testing.T) {
	t.Run("Normal metrics mapping", func(t *testing.T) {
		downloadMbps := 150.25
		uploadMbps := 45.10
		pingMs := 15.2
		shareURL := "http://share/url.png"
		isp := "Test ISP"
		server := "New York"
		bytesSent := int64(5000)
		bytesReceived := int64(10000)

		m, err := mapResult(downloadMbps, uploadMbps, pingMs, shareURL, isp, server, bytesSent, bytesReceived)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if m.Download != downloadMbps || m.Upload != uploadMbps || m.Ping != pingMs {
			t.Errorf("mismatched values: %+v", m)
		}
		if m.Share != shareURL || m.Client != isp || m.Server != server {
			t.Errorf("mismatched metadata: %+v", m)
		}
	})

	t.Run("Ping >= 1000 set to 0 and empty share defaults to N/A", func(t *testing.T) {
		m, err := mapResult(100, 50, 1200, "", "", "", 10, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if m.Ping != 0 {
			t.Errorf("expected ping >= 1000 to be set to 0, got %f", m.Ping)
		}
		if m.Share != "N/A" {
			t.Errorf("expected empty share to default to N/A, got %s", m.Share)
		}
		if m.Client != "Unknown ISP" {
			t.Errorf("expected empty client to default to Unknown ISP, got %s", m.Client)
		}
		if m.Server != "Unknown Server" {
			t.Errorf("expected empty server to default to Unknown Server, got %s", m.Server)
		}
	})
}
