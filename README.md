# MikroTik PPPoE VictoriaMetrics Monitor

Lightweight monitoring stack for MikroTik PPPoE interface traffic without SNMP `ifIndex`.

The collector uses the RouterOS API and performs one bulk interface request per router per polling interval. It stores no history locally. VictoriaMetrics stores the time series, and Grafana provides the dashboard.

## Architecture

```text
MikroTik RouterOS
  -> RouterOS API bulk /interface/print stats
  -> Go collector cached /metrics
  -> vmagent remote write
  -> VictoriaMetrics
  -> Grafana
```

## Why Not SNMP ifIndex

Dynamic PPPoE interfaces can receive a new SNMP `ifIndex` after reconnect. Monitoring based on `ifIndex` splits one customer into multiple time series. This collector uses stable PPPoE interface names as metric identity:

```text
router="core-pppoe-01", interface="pppoe-CUSTOMER"
```

## MikroTik Preparation

Create a read-only API user:

```routeros
/user group add name=monitoring policy=read,api,!local,!telnet,!ssh,!ftp,!reboot,!write,!policy,!test,!winbox,!password,!web,!sniff,!sensitive,!romon
/user add name=monitor password=CHANGE_STRONG_PASSWORD group=monitoring
/ip service enable api
/ip service set api port=8728 address=MONITORING_SERVER_IP/32
```

For TLS:

```routeros
/ip service enable api-ssl
/ip service set api-ssl port=8729 address=MONITORING_SERVER_IP/32
```

Use a strong password, restrict source IP, and do not use the admin user.

## Configure

Copy `.env.example` to `.env` and set:

```text
MIKROTIK_USERNAME=monitor
MIKROTIK_PASSWORD=strong-password
```

Edit `configs/config.example.yaml` and add routers. Sensitive username/password values can stay as placeholders when the environment variables above are set.

## Run

```bash
docker compose up -d
```

Services:

- Collector: http://localhost:9107
- VictoriaMetrics: http://localhost:8428
- Grafana: http://localhost:3000

Grafana default credentials come from `.env`; defaults are `admin/admin`.

## Validate

```bash
curl http://localhost:9107/healthz
curl http://localhost:9107/readyz
curl http://localhost:9107/metrics
curl "http://localhost:8428/api/v1/query?query=mikrotik_router_api_up"
```

Grafana auto-provisions the VictoriaMetrics datasource and imports the MikroTik PPPoE Overview dashboard.

## Metrics

The collector exposes:

- `mikrotik_interface_rx_bytes_total`
- `mikrotik_interface_tx_bytes_total`
- `mikrotik_interface_rx_packets_total`
- `mikrotik_interface_tx_packets_total`
- `mikrotik_interface_rx_errors_total`
- `mikrotik_interface_tx_errors_total`
- `mikrotik_interface_rx_drops_total`
- `mikrotik_interface_tx_drops_total`
- `mikrotik_interface_running`
- `mikrotik_interface_disabled`
- `mikrotik_interface_up`
- `mikrotik_router_api_up`
- `mikrotik_router_api_duration_seconds`
- `mikrotik_router_interfaces_total`

## Polling Guidance

Start with 60 seconds. For routers with many PPPoE sessions:

- 1,000 to 5,000 interfaces: 60 to 120 seconds
- 5,000 to 20,000 interfaces: 120 to 300 seconds

Do not reduce the interval unless router CPU and API latency are clearly healthy.

## Alert Query Examples

```promql
mikrotik_router_api_up == 0
sum(rate(mikrotik_interface_rx_bytes_total[5m])) == 0
mikrotik_router_api_duration_seconds > 10
mikrotik_interface_up{interface=~"pppoe-.*"} == 0
```

## Troubleshooting

- `/readyz` returns 503: no successful RouterOS API poll yet.
- `mikrotik_router_api_up == 0`: check address, API service, firewall, username, password, and TLS setting.
- No PPPoE interfaces: check `filter.include_prefixes`; some RouterOS versions report dynamic PPPoE names as `<pppoe-user>`.
- Grafana has no data: check vmagent logs and VictoriaMetrics query API.
