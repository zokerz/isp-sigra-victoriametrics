package cache

import (
	"sync"
	"time"

	"mikrotik-victoriametrics-monitor/internal/routeros"
)

type RouterSnapshot struct {
	Name              string                       `json:"name"`
	Address           string                       `json:"address"`
	APIUp             bool                         `json:"api_up"`
	LastError         string                       `json:"last_error,omitempty"`
	LastPollAt        time.Time                    `json:"last_poll_at,omitempty"`
	LastSuccessAt     time.Time                    `json:"last_success_at,omitempty"`
	DurationSeconds   float64                      `json:"duration_seconds"`
	InterfacesTotal   int                          `json:"interfaces_total"`
	ExportedTotal     int                          `json:"exported_total"`
	Interfaces        []routeros.InterfaceStats    `json:"interfaces"`
}

type Memory struct {
	mu      sync.RWMutex
	routers map[string]RouterSnapshot
	ready   bool
}

func New() *Memory {
	return &Memory{routers: map[string]RouterSnapshot{}}
}

func (m *Memory) Upsert(snapshot RouterSnapshot, success bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !success {
		prev, ok := m.routers[snapshot.Name]
		if ok {
			snapshot.Interfaces = prev.Interfaces
			snapshot.ExportedTotal = prev.ExportedTotal
			snapshot.InterfacesTotal = prev.InterfacesTotal
			snapshot.LastSuccessAt = prev.LastSuccessAt
		}
	}
	if success {
		m.ready = true
		snapshot.LastSuccessAt = snapshot.LastPollAt
	}
	m.routers[snapshot.Name] = snapshot
}

func (m *Memory) Ready() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.ready
}

func (m *Memory) Snapshots() []RouterSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]RouterSnapshot, 0, len(m.routers))
	for _, snapshot := range m.routers {
		out = append(out, snapshot)
	}
	return out
}

func (m *Memory) Router(name string) (RouterSnapshot, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.routers[name]
	return s, ok
}

func (m *Memory) Interfaces(routerName string) []routeros.InterfaceStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []routeros.InterfaceStats
	for _, snapshot := range m.routers {
		if routerName != "" && snapshot.Name != routerName {
			continue
		}
		out = append(out, snapshot.Interfaces...)
	}
	return out
}
