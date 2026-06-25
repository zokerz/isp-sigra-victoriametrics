package metrics

import (
	"strings"
	"testing"

	"mikrotik-victoriametrics-monitor/internal/cache"
	"mikrotik-victoriametrics-monitor/internal/routeros"
)

func TestRenderPrometheus(t *testing.T) {
	out := RenderPrometheus([]cache.RouterSnapshot{{
		Name:            "core",
		Address:         "192.0.2.1",
		APIUp:           true,
		DurationSeconds: 0.25,
		InterfacesTotal: 2,
		Interfaces: []routeros.InterfaceStats{{
			Name:    "pppoe-customer01",
			Running: true,
			RxBytes: 100,
			TxBytes: 200,
			Type:    "pppoe-in",
		}},
	}})

	needles := []string{
		"mikrotik_router_api_up{router=\"core\",address=\"192.0.2.1\"} 1",
		"mikrotik_interface_rx_bytes_total{router=\"core\",address=\"192.0.2.1\",interface=\"pppoe-customer01\"",
		"mikrotik_interface_up",
	}
	for _, needle := range needles {
		if !strings.Contains(out, needle) {
			t.Fatalf("missing %q in:\n%s", needle, out)
		}
	}
	if strings.Contains(strings.ToLower(out), "ifindex") {
		t.Fatalf("must not expose ifIndex: %s", out)
	}
}
