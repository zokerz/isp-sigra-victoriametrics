package routeros

import (
	"testing"

	"mikrotik-victoriametrics-monitor/internal/config"
)

func TestParseInterface(t *testing.T) {
	got := ParseInterface(map[string]string{
		"name":      "pppoe-customer01",
		"running":   "true",
		"disabled":  "false",
		"rx-byte":   "100",
		"tx-byte":   "200",
		"rx-packet": "3",
		"tx-packet": "4",
		"rx-error":  "5",
		"tx-error":  "6",
		"rx-drop":   "7",
		"tx-drop":   "8",
		"comment":   "Customer 01",
		"type":      "pppoe-in",
	})
	if got.Name != "pppoe-customer01" || !got.Running || got.Disabled {
		t.Fatalf("unexpected status: %+v", got)
	}
	if got.RxBytes != 100 || got.TxDrop != 8 {
		t.Fatalf("unexpected counters: %+v", got)
	}
}

func TestIncludeInterface(t *testing.T) {
	filter := config.FilterConfig{
		IncludePrefixes: []string{"pppoe-"},
		ExcludePrefixes: []string{"ovpn-"},
	}
	if !IncludeInterface("pppoe-customer01", filter) {
		t.Fatal("expected PPPoE interface to be included")
	}
	if IncludeInterface("ovpn-user01", filter) {
		t.Fatal("expected excluded prefix to be excluded")
	}
	if IncludeInterface("ether1", filter) {
		t.Fatal("expected unmatched interface to be excluded")
	}
}
