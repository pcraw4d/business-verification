package datasource

import (
	"context"
	"time"
)

// SourceHealth represents the health status of a single data source
type SourceHealth struct {
	SourceName string        `json:"source_name"`
	Healthy    bool          `json:"healthy"`
	CheckedAt  time.Time     `json:"checked_at"`
	Latency    time.Duration `json:"latency"`
	Error      string        `json:"error,omitempty"`
}

// CheckHealth probes all configured sources and stores the latest status internally.
func (a *Aggregator) CheckHealth(ctx context.Context) []SourceHealth {
	statuses := make([]SourceHealth, len(a.sources))
	if len(a.sources) == 0 {
		return statuses
	}
	for i, src := range a.sources {
		started := time.Now()
		err := src.HealthCheck(ctx)
		statuses[i] = SourceHealth{
			SourceName: src.Name(),
			Healthy:    err == nil,
			CheckedAt:  time.Now(),
			Latency:    time.Since(started),
		}
		if err != nil {
			statuses[i].Error = err.Error()
		}
	}
	return statuses
}
