package main

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"mikrotik-victoriametrics-monitor/internal/cache"
	"mikrotik-victoriametrics-monitor/internal/config"
	"mikrotik-victoriametrics-monitor/internal/logger"
	"mikrotik-victoriametrics-monitor/internal/metrics"
	"mikrotik-victoriametrics-monitor/internal/routeros"
)

func main() {
	log := logger.New()
	cfg, err := config.Load(config.ConfigPathFromFlags())
	if err != nil {
		log.Error("config load failed", "error", err)
		os.Exit(1)
	}

	store := cache.New()
	client := routeros.NewClient()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Info("collector startup", "listen_addr", cfg.Server.ListenAddr, "routers", len(cfg.Routers))
	for _, r := range cfg.Routers {
		go pollRouter(ctx, log, cfg, r, client, store)
	}

	reg := metrics.NewRegistry(store)
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		if !store.Ready() {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		_, _ = w.Write([]byte(reg.Render()))
	})
	mux.HandleFunc("/api/status", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"ready": store.Ready(), "routers": store.Snapshots()})
	})
	mux.HandleFunc("/api/routers", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, store.Snapshots())
	})
	mux.HandleFunc("/api/interfaces", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, store.Interfaces(r.URL.Query().Get("router")))
	})
	mux.HandleFunc("/api/interfaces/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/api/interfaces/")
		for _, iface := range store.Interfaces("") {
			if iface.Name == name {
				writeJSON(w, http.StatusOK, iface)
				return
			}
		}
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "interface not found"})
	})

	server := &http.Server{Addr: cfg.Server.ListenAddr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error("http server failed", "error", err)
		os.Exit(1)
	}
	log.Info("collector stopped")
}

type apiClient interface {
	FetchInterfaces(context.Context, config.RouterConfig, config.FilterConfig) ([]routeros.InterfaceStats, int, error)
}

func pollRouter(ctx context.Context, log *slog.Logger, cfg config.Config, r config.RouterConfig, client apiClient, store *cache.Memory) {
	ticker := time.NewTicker(cfg.PollInterval())
	defer ticker.Stop()

	for {
		pollOnce(ctx, log, cfg, r, client, store)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func pollOnce(parent context.Context, log *slog.Logger, cfg config.Config, r config.RouterConfig, client apiClient, store *cache.Memory) {
	start := time.Now()
	log.Info("router poll started", "router", r.Name, "address", r.Address)

	var ifaces []routeros.InterfaceStats
	var total int
	var err error
	for attempt := 1; attempt <= 2; attempt++ {
		pollCtx, cancel := context.WithTimeout(parent, cfg.PollTimeout())
		ifaces, total, err = client.FetchInterfaces(pollCtx, r, cfg.Filter)
		cancel()
		if err == nil {
			break
		}
		log.Warn("router poll failed", "router", r.Name, "attempt", attempt, "error", err)
	}

	duration := time.Since(start).Seconds()
	snapshot := cache.RouterSnapshot{
		Name:            r.Name,
		Address:         r.Address,
		APIUp:           err == nil,
		LastPollAt:      time.Now().UTC(),
		DurationSeconds: duration,
		InterfacesTotal: total,
		ExportedTotal:   len(ifaces),
		Interfaces:      ifaces,
	}
	if err != nil {
		snapshot.LastError = err.Error()
		store.Upsert(snapshot, false)
		log.Error("router poll failure", "router", r.Name, "duration_seconds", duration, "error", err)
		return
	}
	store.Upsert(snapshot, true)
	log.Info(
		"router poll success",
		"router", r.Name,
		"interfaces_discovered", total,
		"interfaces_exported", len(ifaces),
		"duration_seconds", duration,
	)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
