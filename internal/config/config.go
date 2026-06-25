package config

import (
	"errors"
	"flag"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server  ServerConfig   `yaml:"server"`
	Polling PollingConfig  `yaml:"polling"`
	Routers []RouterConfig `yaml:"routers"`
	Filter  FilterConfig   `yaml:"filter"`
	Metrics MetricsConfig  `yaml:"metrics"`
}

type ServerConfig struct {
	ListenAddr string `yaml:"listen_addr"`
}

type PollingConfig struct {
	IntervalSeconds int `yaml:"interval_seconds"`
	TimeoutSeconds  int `yaml:"timeout_seconds"`
}

type RouterConfig struct {
	Name     string `yaml:"name"`
	Address  string `yaml:"address"`
	APIPort  int    `yaml:"api_port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	UseTLS   bool   `yaml:"use_tls"`
}

type FilterConfig struct {
	IncludePrefixes []string `yaml:"include_prefixes"`
	ExcludePrefixes []string `yaml:"exclude_prefixes"`
}

type MetricsConfig struct {
	Mode string `yaml:"mode"`
}

func Default() Config {
	return Config{
		Server:  ServerConfig{ListenAddr: "0.0.0.0:9107"},
		Polling: PollingConfig{IntervalSeconds: 60, TimeoutSeconds: 15},
		Filter: FilterConfig{
			IncludePrefixes: []string{"pppoe-"},
			ExcludePrefixes: []string{"ovpn-", "sstp-", "l2tp-"},
		},
		Metrics: MetricsConfig{Mode: "prometheus"},
	}
}

func ConfigPathFromFlags() string {
	path := flag.String("config", "configs/config.example.yaml", "path to config YAML")
	flag.Parse()
	return *path
}

func Load(path string) (Config, error) {
	cfg := Default()
	raw, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return cfg, err
	}
	applyEnvOverrides(&cfg)
	return cfg, cfg.Validate()
}

func applyEnvOverrides(cfg *Config) {
	username := os.Getenv("MIKROTIK_USERNAME")
	password := os.Getenv("MIKROTIK_PASSWORD")
	for i := range cfg.Routers {
		if username != "" {
			cfg.Routers[i].Username = username
		}
		if password != "" {
			cfg.Routers[i].Password = password
		}
	}
}

func (c Config) Validate() error {
	if c.Server.ListenAddr == "" {
		return errors.New("server.listen_addr is required")
	}
	if c.Polling.IntervalSeconds <= 0 {
		return errors.New("polling.interval_seconds must be greater than zero")
	}
	if c.Polling.TimeoutSeconds <= 0 {
		return errors.New("polling.timeout_seconds must be greater than zero")
	}
	if len(c.Routers) == 0 {
		return errors.New("at least one router is required")
	}
	for _, r := range c.Routers {
		if r.Name == "" || r.Address == "" || r.Username == "" || r.Password == "" {
			return errors.New("router name, address, username, and password are required")
		}
		if r.APIPort <= 0 {
			return errors.New("router api_port must be greater than zero")
		}
	}
	if c.Metrics.Mode == "" {
		c.Metrics.Mode = "prometheus"
	}
	if c.Metrics.Mode != "prometheus" {
		return errors.New("only metrics.mode=prometheus is implemented")
	}
	return nil
}

func (c Config) PollInterval() time.Duration {
	return time.Duration(c.Polling.IntervalSeconds) * time.Second
}

func (c Config) PollTimeout() time.Duration {
	return time.Duration(c.Polling.TimeoutSeconds) * time.Second
}
