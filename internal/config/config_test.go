package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigAndEnvOverride(t *testing.T) {
	t.Setenv("MIKROTIK_USERNAME", "env-user")
	t.Setenv("MIKROTIK_PASSWORD", "env-pass")

	path := filepath.Join(t.TempDir(), "config.yaml")
	data := []byte(`
server:
  listen_addr: "127.0.0.1:9107"
polling:
  interval_seconds: 120
  timeout_seconds: 10
routers:
  - name: "r1"
    address: "192.0.2.1"
    api_port: 8728
    username: "file-user"
    password: "file-pass"
filter:
  include_prefixes: ["pppoe-"]
metrics:
  mode: "prometheus"
`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := cfg.Routers[0].Username; got != "env-user" {
		t.Fatalf("username override = %q", got)
	}
	if got := cfg.Routers[0].Password; got != "env-pass" {
		t.Fatalf("password override = %q", got)
	}
	if cfg.Polling.IntervalSeconds != 120 {
		t.Fatalf("interval = %d", cfg.Polling.IntervalSeconds)
	}
}
