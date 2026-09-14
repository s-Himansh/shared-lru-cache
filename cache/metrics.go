package cache

import "sync/atomic"

// Metrics tracks cache performance statistics.
type Metrics struct {
	Hits      atomic.Int64
	Misses    atomic.Int64
	Evictions atomic.Int64
	Puts      atomic.Int64
	Gets      atomic.Int64
	Deletes   atomic.Int64
}

// Snapshot returns a point-in-time copy of the metrics.
func (m *Metrics) Snapshot() MetricsSnapshot {
	return MetricsSnapshot{
		Hits:      m.Hits.Load(),
		Misses:    m.Misses.Load(),
		Evictions: m.Evictions.Load(),
		Puts:      m.Puts.Load(),
		Gets:      m.Gets.Load(),
		Deletes:   m.Deletes.Load(),
		HitRate:   m.hitRate(),
	}
}

func (m *Metrics) hitRate() float64 {
	total := m.Hits.Load() + m.Misses.Load()
	if total == 0 {
		return 0
	}
	return float64(m.Hits.Load()) / float64(total) * 100
}

// MetricsSnapshot is a point-in-time copy of metrics.
type MetricsSnapshot struct {
	Hits      int64   `json:"hits"`
	Misses    int64   `json:"misses"`
	Evictions int64   `json:"evictions"`
	Puts      int64   `json:"puts"`
	Gets      int64   `json:"gets"`
	Deletes   int64   `json:"deletes"`
	HitRate   float64 `json:"hit_rate"`
}
