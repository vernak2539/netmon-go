package scanner

import (
	"testing"
)

func TestParseSubnet(t *testing.T) {
	t.Run("Valid ethernet interface list", func(t *testing.T) {
		iflist := `
Starting Nmap
INTERFACES:
DEV    TYPE     IP/MASK         MAC
en0    192.168.1.5/24 ethernet up  AA:BB:CC:DD:EE:FF
	`
		subnet, err := parseSubnet(iflist)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if subnet != "192.168.1.5/24" {
			t.Errorf("expected 192.168.1.5/24, got %s", subnet)
		}
	})

	t.Run("No active ethernet interface error", func(t *testing.T) {
		iflist := `
Starting Nmap
INTERFACES:
DEV    TYPE     IP/MASK         MAC
lo0    loopback 127.0.0.1/8     00:00:00:00:00:00
	`
		_, err := parseSubnet(iflist)
		if err == nil {
			t.Error("expected error when no ethernet interface up, got nil")
		}
	})
}

func TestParseNmapXML(t *testing.T) {
	t.Run("Valid host XML parsing", func(t *testing.T) {
		xmlSample := `<?xml version="1.0"?>
<nmaprun>
  <host>
    <status state="up"/>
    <address addr="192.168.1.1" addrtype="ipv4"/>
    <times srtt="15000"/>
  </host>
  <host>
    <status state="down"/>
    <address addr="192.168.1.2" addrtype="ipv4"/>
  </host>
  <host>
    <status state="up"/>
    <address addr="192.168.1.100" addrtype="ipv4"/>
    <times srtt="2500"/>
  </host>
</nmaprun>`

		devices, err := parseNmapXML([]byte(xmlSample))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(devices) != 2 {
			t.Fatalf("expected 2 active devices, got %d", len(devices))
		}

		if devices[0].IP != "192.168.1.1" || devices[0].LatencyMs != 15.0 {
			t.Errorf("unexpected device 0: %+v", devices[0])
		}
		if devices[1].IP != "192.168.1.100" || devices[1].LatencyMs != 2.5 {
			t.Errorf("unexpected device 1: %+v", devices[1])
		}
	})
}
