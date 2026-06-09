package analytics

import (
	"sync"
	"time"
)

type ToolMetric struct {
	ToolName  string
	OrgID     string
	Success   bool
	LatencyMs int64
	CreatedAt time.Time
}

type Collector struct {
	mu      sync.RWMutex
	metrics []ToolMetric
}

func NewCollector() *Collector {
	return &Collector{}
}

func (c *Collector) Record(metric ToolMetric) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if metric.CreatedAt.IsZero() {
		metric.CreatedAt = time.Now()
	}
	c.metrics = append(c.metrics, metric)
}

func (c *Collector) Snapshot() []ToolMetric {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]ToolMetric, len(c.metrics))
	copy(out, c.metrics)
	return out
}
