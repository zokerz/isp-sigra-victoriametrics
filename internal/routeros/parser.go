package routeros

import (
	"strconv"
	"strings"

	"mikrotik-victoriametrics-monitor/internal/config"
)

type InterfaceStats struct {
	Name     string `json:"name"`
	Running  bool   `json:"running"`
	Disabled bool   `json:"disabled"`
	RxBytes  uint64 `json:"rx_bytes"`
	TxBytes  uint64 `json:"tx_bytes"`
	RxPacket uint64 `json:"rx_packets"`
	TxPacket uint64 `json:"tx_packets"`
	RxError  uint64 `json:"rx_errors"`
	TxError  uint64 `json:"tx_errors"`
	RxDrop   uint64 `json:"rx_drops"`
	TxDrop   uint64 `json:"tx_drops"`
	Comment  string `json:"comment,omitempty"`
	Type     string `json:"type,omitempty"`
}

func ParseInterface(row map[string]string) InterfaceStats {
	return InterfaceStats{
		Name:     row["name"],
		Running:  parseBool(row["running"]),
		Disabled: parseBool(row["disabled"]),
		RxBytes:  parseUint(row["rx-byte"]),
		TxBytes:  parseUint(row["tx-byte"]),
		RxPacket: parseUint(row["rx-packet"]),
		TxPacket: parseUint(row["tx-packet"]),
		RxError:  parseUint(row["rx-error"]),
		TxError:  parseUint(row["tx-error"]),
		RxDrop:   parseUint(row["rx-drop"]),
		TxDrop:   parseUint(row["tx-drop"]),
		Comment:  row["comment"],
		Type:     row["type"],
	}
}

func IncludeInterface(name string, filter config.FilterConfig) bool {
	if name == "" {
		return false
	}
	for _, prefix := range filter.ExcludePrefixes {
		if prefix != "" && strings.HasPrefix(name, prefix) {
			return false
		}
	}
	if len(filter.IncludePrefixes) == 0 {
		return true
	}
	for _, prefix := range filter.IncludePrefixes {
		if prefix != "" && strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

func parseBool(v string) bool {
	return strings.EqualFold(v, "true") || strings.EqualFold(v, "yes") || v == "1"
}

func parseUint(v string) uint64 {
	if v == "" {
		return 0
	}
	n, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return 0
	}
	return n
}
