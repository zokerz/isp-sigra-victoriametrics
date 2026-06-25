package metrics

import "mikrotik-victoriametrics-monitor/internal/cache"

type Registry struct {
	cache *cache.Memory
}

func NewRegistry(c *cache.Memory) *Registry {
	return &Registry{cache: c}
}

func (r *Registry) Render() string {
	return RenderPrometheus(r.cache.Snapshots())
}
