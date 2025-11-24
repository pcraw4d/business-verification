package repository

import (
	"log"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// ClassificationMetrics holds Prometheus metrics for classification operations
type ClassificationMetrics struct {
	// Counter metrics
	PagesAnalyzedTotal          *prometheus.CounterVec
	StructuredDataFoundTotal    *prometheus.CounterVec
	BrandMatchesTotal           *prometheus.CounterVec
	
	// Histogram metrics
	AnalysisDurationSeconds     *prometheus.HistogramVec
	PageAnalysisDurationSeconds *prometheus.HistogramVec
	
	// Gauge metrics
	ConcurrentPagesAnalyzing    prometheus.Gauge
	MemoryUsageBytes            prometheus.Gauge
}

var (
	metrics *ClassificationMetrics
	once    sync.Once
)

// GetClassificationMetrics returns the singleton metrics instance
func GetClassificationMetrics() *ClassificationMetrics {
	once.Do(func() {
		metrics = &ClassificationMetrics{
			// Counter: Total pages analyzed by method
			PagesAnalyzedTotal: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Name: "classification_pages_analyzed_total",
					Help: "Total number of pages analyzed by analysis method",
				},
				[]string{"method"}, // method: multi_page, single_page, url_only
			),
			
			// Counter: Structured data found
			StructuredDataFoundTotal: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Name: "classification_structured_data_found_total",
					Help: "Count of structured data found during analysis",
				},
				[]string{"found"}, // found: true, false
			),
			
			// Counter: Brand matches by MCC range
			BrandMatchesTotal: promauto.NewCounterVec(
				prometheus.CounterOpts{
					Name: "classification_brand_matches_total",
					Help: "Count of brand matches by MCC range",
				},
				[]string{"mcc_range"}, // mcc_range: 3000-3831, other, none
			),
			
			// Histogram: Analysis duration by method
			AnalysisDurationSeconds: promauto.NewHistogramVec(
				prometheus.HistogramOpts{
					Name:    "classification_analysis_duration_seconds",
					Help:    "Duration of classification analysis by method",
					Buckets: []float64{0.1, 0.5, 1, 5, 10, 30, 60, 90, 120}, // Up to 120s
				},
				[]string{"method"}, // method: multi_page, single_page, url_only
			),
			
			// Histogram: Per-page analysis duration by page type
			PageAnalysisDurationSeconds: promauto.NewHistogramVec(
				prometheus.HistogramOpts{
					Name:    "classification_page_analysis_duration_seconds",
					Help:    "Duration of individual page analysis by page type",
					Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 15}, // Up to 15s per page
				},
				[]string{"page_type"}, // page_type: about, services, products, other
			),
			
			// Gauge: Current concurrent page analyses
			ConcurrentPagesAnalyzing: promauto.NewGauge(
				prometheus.GaugeOpts{
					Name: "classification_concurrent_pages_analyzing",
					Help: "Current number of concurrent page analyses",
				},
			),
			
			// Gauge: Memory usage per classification job
			MemoryUsageBytes: promauto.NewGauge(
				prometheus.GaugeOpts{
					Name: "classification_memory_usage_bytes",
					Help: "Memory usage per classification job in bytes",
				},
			),
		}
	})
	
	return metrics
}

// RecordPagesAnalyzed increments the pages analyzed counter
func (m *ClassificationMetrics) RecordPagesAnalyzed(method string, count int) {
	if m == nil {
		return
	}
	for i := 0; i < count; i++ {
		m.PagesAnalyzedTotal.WithLabelValues(method).Inc()
	}
}

// RecordStructuredDataFound increments the structured data found counter
func (m *ClassificationMetrics) RecordStructuredDataFound(found bool) {
	if m == nil {
		return
	}
	foundStr := "false"
	if found {
		foundStr = "true"
	}
	m.StructuredDataFoundTotal.WithLabelValues(foundStr).Inc()
}

// RecordBrandMatch increments the brand match counter
func (m *ClassificationMetrics) RecordBrandMatch(mccRange string) {
	if m == nil {
		return
	}
	if mccRange == "" {
		mccRange = "none"
	}
	m.BrandMatchesTotal.WithLabelValues(mccRange).Inc()
}

// RecordAnalysisDuration records the analysis duration
func (m *ClassificationMetrics) RecordAnalysisDuration(method string, duration time.Duration) {
	if m == nil {
		return
	}
	m.AnalysisDurationSeconds.WithLabelValues(method).Observe(duration.Seconds())
}

// RecordPageAnalysisDuration records the per-page analysis duration
func (m *ClassificationMetrics) RecordPageAnalysisDuration(pageType string, duration time.Duration) {
	if m == nil {
		return
	}
	if pageType == "" {
		pageType = "other"
	}
	m.PageAnalysisDurationSeconds.WithLabelValues(pageType).Observe(duration.Seconds())
}

// SetConcurrentPagesAnalyzing sets the current number of concurrent page analyses
func (m *ClassificationMetrics) SetConcurrentPagesAnalyzing(count int) {
	if m == nil {
		return
	}
	m.ConcurrentPagesAnalyzing.Set(float64(count))
}

// SetMemoryUsage sets the memory usage in bytes
func (m *ClassificationMetrics) SetMemoryUsage(bytes int64) {
	if m == nil {
		return
	}
	m.MemoryUsageBytes.Set(float64(bytes))
}

// InitializeMetrics initializes metrics with a logger (for backward compatibility)
func InitializeMetrics(logger *log.Logger) *ClassificationMetrics {
	return GetClassificationMetrics()
}

