package metrics

import (
	"fmt"
	"sort"
	"strings"

	"mikrotik-victoriametrics-monitor/internal/cache"
)

func RenderPrometheus(snapshots []cache.RouterSnapshot) string {
	sort.Slice(snapshots, func(i, j int) bool { return snapshots[i].Name < snapshots[j].Name })

	var b strings.Builder
	writeHelp(&b, "mikrotik_interface_rx_bytes_total", "Interface received bytes")
	writeHelp(&b, "mikrotik_interface_tx_bytes_total", "Interface transmitted bytes")
	writeHelp(&b, "mikrotik_interface_rx_packets_total", "Interface received packets")
	writeHelp(&b, "mikrotik_interface_tx_packets_total", "Interface transmitted packets")
	writeHelp(&b, "mikrotik_interface_rx_errors_total", "Interface receive errors")
	writeHelp(&b, "mikrotik_interface_tx_errors_total", "Interface transmit errors")
	writeHelp(&b, "mikrotik_interface_rx_drops_total", "Interface receive drops")
	writeHelp(&b, "mikrotik_interface_tx_drops_total", "Interface transmit drops")
	writeGaugeHelp(&b, "mikrotik_interface_running", "Interface running flag")
	writeGaugeHelp(&b, "mikrotik_interface_disabled", "Interface disabled flag")
	writeGaugeHelp(&b, "mikrotik_interface_up", "Interface running and not disabled flag")
	writeGaugeHelp(&b, "mikrotik_router_api_up", "RouterOS API poll success")
	writeGaugeHelp(&b, "mikrotik_router_api_duration_seconds", "RouterOS API poll duration")
	writeGaugeHelp(&b, "mikrotik_router_interfaces_total", "Total interfaces returned by RouterOS before filtering")

	for _, router := range snapshots {
		routerLabels := fmt.Sprintf(`router="%s",address="%s"`, esc(router.Name), esc(router.Address))
		fmt.Fprintf(&b, "mikrotik_router_api_up{%s} %d\n", routerLabels, boolNum(router.APIUp))
		fmt.Fprintf(&b, "mikrotik_router_api_duration_seconds{%s} %.6f\n", routerLabels, router.DurationSeconds)
		fmt.Fprintf(&b, "mikrotik_router_interfaces_total{%s} %d\n", routerLabels, router.InterfacesTotal)

		sort.Slice(router.Interfaces, func(i, j int) bool { return router.Interfaces[i].Name < router.Interfaces[j].Name })
		for _, iface := range router.Interfaces {
			labels := fmt.Sprintf(
				`router="%s",address="%s",interface="%s",type="%s",comment="%s"`,
				esc(router.Name), esc(router.Address), esc(iface.Name), esc(iface.Type), esc(iface.Comment),
			)
			fmt.Fprintf(&b, "mikrotik_interface_rx_bytes_total{%s} %d\n", labels, iface.RxBytes)
			fmt.Fprintf(&b, "mikrotik_interface_tx_bytes_total{%s} %d\n", labels, iface.TxBytes)
			fmt.Fprintf(&b, "mikrotik_interface_rx_packets_total{%s} %d\n", labels, iface.RxPacket)
			fmt.Fprintf(&b, "mikrotik_interface_tx_packets_total{%s} %d\n", labels, iface.TxPacket)
			fmt.Fprintf(&b, "mikrotik_interface_rx_errors_total{%s} %d\n", labels, iface.RxError)
			fmt.Fprintf(&b, "mikrotik_interface_tx_errors_total{%s} %d\n", labels, iface.TxError)
			fmt.Fprintf(&b, "mikrotik_interface_rx_drops_total{%s} %d\n", labels, iface.RxDrop)
			fmt.Fprintf(&b, "mikrotik_interface_tx_drops_total{%s} %d\n", labels, iface.TxDrop)
			fmt.Fprintf(&b, "mikrotik_interface_running{%s} %d\n", labels, boolNum(iface.Running))
			fmt.Fprintf(&b, "mikrotik_interface_disabled{%s} %d\n", labels, boolNum(iface.Disabled))
			fmt.Fprintf(&b, "mikrotik_interface_up{%s} %d\n", labels, boolNum(iface.Running && !iface.Disabled))
		}
	}
	return b.String()
}

func writeHelp(b *strings.Builder, name, help string) {
	fmt.Fprintf(b, "# HELP %s %s\n# TYPE %s counter\n", name, help, name)
}

func writeGaugeHelp(b *strings.Builder, name, help string) {
	fmt.Fprintf(b, "# HELP %s %s\n# TYPE %s gauge\n", name, help, name)
}

func boolNum(v bool) int {
	if v {
		return 1
	}
	return 0
}

func esc(v string) string {
	v = strings.ReplaceAll(v, `\`, `\\`)
	v = strings.ReplaceAll(v, `"`, `\"`)
	v = strings.ReplaceAll(v, "\n", `\n`)
	return v
}
