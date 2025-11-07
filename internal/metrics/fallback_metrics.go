package metrics

import (
	"context"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

// FallbackMetrics tracks fallback data usage across services
type FallbackMetrics struct {
	mu        sync.RWMutex
	usage     map[string]*ServiceFallbackStats
	logger    *zap.Logger
	startTime time.Time

	// Prometheus metrics
	fallbackTotal      *prometheus.CounterVec
	fallbackRate       *prometheus.GaugeVec
	fallbackDuration   *prometheus.HistogramVec
	requestsTotal      *prometheus.CounterVec
	fallbackByCategory *prometheus.CounterVec
	fallbackBySource   *prometheus.CounterVec
}

// ServiceFallbackStats holds statistics for a service's fallback usage
type ServiceFallbackStats struct {
	ServiceName     string
	TotalRequests   int64
	FallbackCount   int64
	FallbackRate    float64
	ByCategory      map[string]int64 // Category -> count
	BySource        map[string]int64 // Source -> count
	LastFallback    time.Time
	TotalDuration   time.Duration
	AverageDuration time.Duration
}

// NewFallbackMetrics creates a new fallback metrics tracker
func NewFallbackMetrics(logger *zap.Logger) *FallbackMetrics {
	return &FallbackMetrics{
		usage:     make(map[string]*ServiceFallbackStats),
		logger:    logger,
		startTime: time.Now(),

		// Initialize Prometheus metrics
		fallbackTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kyb_fallback_total",
				Help: "Total number of fallback data usage events",
			},
			[]string{"service", "category", "source"},
		),
		fallbackRate: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "kyb_fallback_rate_percent",
				Help: "Fallback usage rate as percentage of total requests",
			},
			[]string{"service"},
		),
		fallbackDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "kyb_fallback_duration_seconds",
				Help:    "Duration of fallback operations in seconds",
				Buckets: prometheus.ExponentialBuckets(0.001, 2, 10), // 1ms to ~1s
			},
			[]string{"service", "category"},
		),
		requestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kyb_requests_total",
				Help: "Total number of requests (fallback and non-fallback)",
			},
			[]string{"service"},
		),
		fallbackByCategory: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kyb_fallback_by_category_total",
				Help: "Total fallback usage by category",
			},
			[]string{"service", "category"},
		),
		fallbackBySource: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kyb_fallback_by_source_total",
				Help: "Total fallback usage by source",
			},
			[]string{"service", "source"},
		),
	}
}

// RecordFallbackUsage records a fallback usage event
func (fm *FallbackMetrics) RecordFallbackUsage(ctx context.Context, serviceName, category, source string, duration time.Duration) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	stats, exists := fm.usage[serviceName]
	if !exists {
		stats = &ServiceFallbackStats{
			ServiceName: serviceName,
			ByCategory:  make(map[string]int64),
			BySource:    make(map[string]int64),
		}
		fm.usage[serviceName] = stats
	}

	stats.TotalRequests++
	stats.FallbackCount++
	stats.FallbackRate = float64(stats.FallbackCount) / float64(stats.TotalRequests) * 100
	stats.LastFallback = time.Now()
	stats.TotalDuration += duration
	stats.AverageDuration = stats.TotalDuration / time.Duration(stats.FallbackCount)

	if category != "" {
		stats.ByCategory[category]++
	}
	if source != "" {
		stats.BySource[source]++
	}

	// Update Prometheus metrics
	fm.fallbackTotal.WithLabelValues(serviceName, category, source).Inc()
	fm.fallbackRate.WithLabelValues(serviceName).Set(stats.FallbackRate)
	fm.fallbackDuration.WithLabelValues(serviceName, category).Observe(duration.Seconds())
	fm.requestsTotal.WithLabelValues(serviceName).Inc()
	if category != "" {
		fm.fallbackByCategory.WithLabelValues(serviceName, category).Inc()
	}
	if source != "" {
		fm.fallbackBySource.WithLabelValues(serviceName, source).Inc()
	}

	fm.logger.Info("Fallback usage recorded",
		zap.String("service", serviceName),
		zap.String("category", category),
		zap.String("source", source),
		zap.Duration("duration", duration))
}

// RecordRequest records a normal (non-fallback) request
func (fm *FallbackMetrics) RecordRequest(ctx context.Context, serviceName string) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	stats, exists := fm.usage[serviceName]
	if !exists {
		stats = &ServiceFallbackStats{
			ServiceName: serviceName,
			ByCategory:  make(map[string]int64),
			BySource:    make(map[string]int64),
		}
		fm.usage[serviceName] = stats
	}

	stats.TotalRequests++
	stats.FallbackRate = float64(stats.FallbackCount) / float64(stats.TotalRequests) * 100

	// Update Prometheus metrics
	fm.requestsTotal.WithLabelValues(serviceName).Inc()
	fm.fallbackRate.WithLabelValues(serviceName).Set(stats.FallbackRate)
}

// GetStats returns statistics for a specific service
func (fm *FallbackMetrics) GetStats(serviceName string) *ServiceFallbackStats {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	stats, exists := fm.usage[serviceName]
	if !exists {
		return &ServiceFallbackStats{
			ServiceName: serviceName,
			ByCategory:  make(map[string]int64),
			BySource:    make(map[string]int64),
		}
	}

	// Return a copy to avoid race conditions
	return &ServiceFallbackStats{
		ServiceName:     stats.ServiceName,
		TotalRequests:   stats.TotalRequests,
		FallbackCount:   stats.FallbackCount,
		FallbackRate:    stats.FallbackRate,
		ByCategory:      copyMap(stats.ByCategory),
		BySource:        copyMap(stats.BySource),
		LastFallback:    stats.LastFallback,
		TotalDuration:   stats.TotalDuration,
		AverageDuration: stats.AverageDuration,
	}
}

// GetAllStats returns statistics for all services
func (fm *FallbackMetrics) GetAllStats() map[string]*ServiceFallbackStats {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	result := make(map[string]*ServiceFallbackStats)
	for serviceName, stats := range fm.usage {
		result[serviceName] = &ServiceFallbackStats{
			ServiceName:     stats.ServiceName,
			TotalRequests:   stats.TotalRequests,
			FallbackCount:   stats.FallbackCount,
			FallbackRate:    stats.FallbackRate,
			ByCategory:      copyMap(stats.ByCategory),
			BySource:        copyMap(stats.BySource),
			LastFallback:    stats.LastFallback,
			TotalDuration:   stats.TotalDuration,
			AverageDuration: stats.AverageDuration,
		}
	}

	return result
}

// GetSummary returns a summary of all fallback metrics
func (fm *FallbackMetrics) GetSummary() FallbackMetricsSummary {
	fm.mu.RLock()
	defer fm.mu.RUnlock()

	summary := FallbackMetricsSummary{
		TotalServices:       len(fm.usage),
		TotalRequests:       0,
		TotalFallbacks:      0,
		OverallFallbackRate: 0,
		Uptime:              time.Since(fm.startTime),
		ByCategory:          make(map[string]int64),
		BySource:            make(map[string]int64),
	}

	for _, stats := range fm.usage {
		summary.TotalRequests += stats.TotalRequests
		summary.TotalFallbacks += stats.FallbackCount

		for category, count := range stats.ByCategory {
			summary.ByCategory[category] += count
		}

		for source, count := range stats.BySource {
			summary.BySource[source] += count
		}
	}

	if summary.TotalRequests > 0 {
		summary.OverallFallbackRate = float64(summary.TotalFallbacks) / float64(summary.TotalRequests) * 100
	}

	return summary
}

// FallbackMetricsSummary provides an overview of all fallback metrics
type FallbackMetricsSummary struct {
	TotalServices       int
	TotalRequests       int64
	TotalFallbacks      int64
	OverallFallbackRate float64
	Uptime              time.Duration
	ByCategory          map[string]int64
	BySource            map[string]int64
}

// copyMap creates a copy of a map
func copyMap(src map[string]int64) map[string]int64 {
	dst := make(map[string]int64)
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
