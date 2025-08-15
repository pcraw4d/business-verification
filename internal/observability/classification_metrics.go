package observability

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// ClassificationMetrics provides comprehensive metrics for enhanced classification features
type ClassificationMetrics struct {
	config *config.ObservabilityConfig

	// Classification Accuracy Metrics
	classificationAccuracyByMethod     *prometheus.GaugeVec
	classificationAccuracyByIndustry   *prometheus.GaugeVec
	classificationAccuracyByRegion     *prometheus.GaugeVec
	classificationAccuracyByConfidence *prometheus.GaugeVec
	classificationAccuracyTrend        *prometheus.GaugeVec

	// Response Time Optimization Metrics
	classificationResponseTimeByMethod   *prometheus.HistogramVec
	classificationResponseTimeByType     *prometheus.HistogramVec
	classificationResponseTimePercentile *prometheus.GaugeVec
	classificationResponseTimeTrend      *prometheus.GaugeVec

	// Resource Usage Metrics
	classificationMemoryUsage      *prometheus.GaugeVec
	classificationCPUUsage         *prometheus.GaugeVec
	classificationCacheHitRate     *prometheus.GaugeVec
	classificationCacheSize        *prometheus.GaugeVec
	classificationDatabaseQueries  *prometheus.CounterVec
	classificationExternalAPICalls *prometheus.CounterVec

	// User Satisfaction Metrics
	userSatisfactionScore *prometheus.GaugeVec
	userFeedbackCount     *prometheus.CounterVec
	userFeedbackSentiment *prometheus.GaugeVec
	userSatisfactionTrend *prometheus.GaugeVec

	// Enhanced Classification Method Metrics
	websiteAnalysisAccuracy          *prometheus.GaugeVec
	websiteAnalysisResponseTime      *prometheus.HistogramVec
	websiteAnalysisSuccessRate       *prometheus.GaugeVec
	websiteAnalysisPageDiscoveryRate *prometheus.GaugeVec

	webSearchAccuracy             *prometheus.GaugeVec
	webSearchResponseTime         *prometheus.HistogramVec
	webSearchSuccessRate          *prometheus.GaugeVec
	webSearchProviderFallbackRate *prometheus.CounterVec

	mlModelAccuracy      *prometheus.GaugeVec
	mlModelResponseTime  *prometheus.HistogramVec
	mlModelInferenceRate *prometheus.CounterVec
	mlModelFallbackRate  *prometheus.CounterVec

	// Confidence Scoring Metrics
	confidenceScoreDistribution *prometheus.HistogramVec
	confidenceScoreByMethod     *prometheus.GaugeVec
	confidenceScoreAccuracy     *prometheus.GaugeVec
	confidenceScoreTrend        *prometheus.GaugeVec

	// Geographic and Industry-Specific Metrics
	geographicClassificationAccuracy *prometheus.GaugeVec
	industrySpecificAccuracy         *prometheus.GaugeVec
	regionSpecificResponseTime       *prometheus.HistogramVec
	industrySpecificResponseTime     *prometheus.HistogramVec

	// Quality Assurance Metrics
	qualityCheckPassRate    *prometheus.GaugeVec
	qualityCheckFailureRate *prometheus.CounterVec
	qualityCheckDuration    *prometheus.HistogramVec
	qualityCheckTrend       *prometheus.GaugeVec

	// Feedback and Validation Metrics
	feedbackCollectionRate      *prometheus.CounterVec
	feedbackProcessingTime      *prometheus.HistogramVec
	feedbackAccuracyImprovement *prometheus.GaugeVec
	feedbackValidationRate      *prometheus.GaugeVec

	// Performance Optimization Metrics
	cachePerformanceMetrics     *prometheus.GaugeVec
	modelOptimizationMetrics    *prometheus.GaugeVec
	resourceOptimizationMetrics *prometheus.GaugeVec
	scalabilityMetrics          *prometheus.GaugeVec

	// Custom metrics for specific business needs
	customMetrics map[string]prometheus.Collector
}

// NewClassificationMetrics creates a new classification metrics collector
func NewClassificationMetrics(cfg *config.ObservabilityConfig) (*ClassificationMetrics, error) {
	if !cfg.MetricsEnabled {
		return &ClassificationMetrics{
			config:        cfg,
			customMetrics: make(map[string]prometheus.Collector),
		}, nil
	}

	metrics := &ClassificationMetrics{
		config:        cfg,
		customMetrics: make(map[string]prometheus.Collector),
	}

	// Initialize Classification Accuracy Metrics
	metrics.classificationAccuracyByMethod = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "classification_accuracy_by_method",
			Help: "Classification accuracy by method (website_analysis, web_search, keyword_based, etc.)",
		},
		[]string{"method", "confidence_range"},
	)

	metrics.classificationAccuracyByIndustry = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "classification_accuracy_by_industry",
			Help: "Classification accuracy by industry sector",
		},
		[]string{"industry", "industry_code"},
	)

	metrics.classificationAccuracyByRegion = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "classification_accuracy_by_region",
			Help: "Classification accuracy by geographic region",
		},
		[]string{"region", "country"},
	)

	metrics.classificationAccuracyByConfidence = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "classification_accuracy_by_confidence",
			Help: "Classification accuracy by confidence score range",
		},
		[]string{"confidence_range", "method"},
	)

	metrics.classificationAccuracyTrend = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "classification_accuracy_trend",
			Help: "Classification accuracy trend over time",
		},
		[]string{"time_window", "method"},
	)

	// Initialize Response Time Optimization Metrics
	metrics.classificationResponseTimeByMethod = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "classification_response_time_by_method_seconds",
			Help:    "Classification response time by method in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "confidence_range"},
	)

	metrics.classificationResponseTimeByType = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "classification_response_time_by_type_seconds",
			Help:    "Classification response time by type (single, batch) in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"type", "method"},
	)

	metrics.classificationResponseTimePercentile = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "classification_response_time_percentile_seconds",
			Help: "Classification response time percentiles in seconds",
		},
		[]string{"percentile", "method"},
	)

	metrics.classificationResponseTimeTrend = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "classification_response_time_trend_seconds",
			Help: "Classification response time trend over time in seconds",
		},
		[]string{"time_window", "method"},
	)

	// Initialize Resource Usage Metrics
	metrics.classificationMemoryUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "classification_memory_usage_bytes",
			Help: "Memory usage for classification operations in bytes",
		},
		[]string{"component", "operation"},
	)

	metrics.classificationCPUUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "classification_cpu_usage_percent",
			Help: "CPU usage for classification operations in percentage",
		},
		[]string{"component", "operation"},
	)

	metrics.classificationCacheHitRate = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "classification_cache_hit_rate",
			Help: "Cache hit rate for classification operations",
		},
		[]string{"cache_type", "method"},
	)

	metrics.classificationCacheSize = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "classification_cache_size",
			Help: "Cache size for classification operations",
		},
		[]string{"cache_type", "unit"},
	)

	metrics.classificationDatabaseQueries = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "classification_database_queries_total",
			Help: "Total number of database queries for classification",
		},
		[]string{"operation", "table", "status"},
	)

	metrics.classificationExternalAPICalls = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "classification_external_api_calls_total",
			Help: "Total number of external API calls for classification",
		},
		[]string{"service", "endpoint", "status"},
	)

	// Initialize User Satisfaction Metrics
	metrics.userSatisfactionScore = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "user_satisfaction_score",
			Help: "User satisfaction score for classification results",
		},
		[]string{"method", "confidence_range"},
	)

	metrics.userFeedbackCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "user_feedback_count_total",
			Help: "Total number of user feedback submissions",
		},
		[]string{"feedback_type", "sentiment"},
	)

	metrics.userFeedbackSentiment = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "user_feedback_sentiment",
			Help: "User feedback sentiment score",
		},
		[]string{"method", "time_window"},
	)

	metrics.userSatisfactionTrend = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "user_satisfaction_trend",
			Help: "User satisfaction trend over time",
		},
		[]string{"time_window", "method"},
	)

	// Initialize Enhanced Classification Method Metrics
	metrics.websiteAnalysisAccuracy = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "website_analysis_accuracy",
			Help: "Website analysis classification accuracy",
		},
		[]string{"page_type", "content_quality"},
	)

	metrics.websiteAnalysisResponseTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "website_analysis_response_time_seconds",
			Help:    "Website analysis response time in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"page_type", "discovery_depth"},
	)

	metrics.websiteAnalysisSuccessRate = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "website_analysis_success_rate",
			Help: "Website analysis success rate",
		},
		[]string{"page_type", "content_quality"},
	)

	metrics.websiteAnalysisPageDiscoveryRate = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "website_analysis_page_discovery_rate",
			Help: "Website analysis page discovery rate",
		},
		[]string{"page_type", "priority_level"},
	)

	metrics.webSearchAccuracy = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "web_search_accuracy",
			Help: "Web search classification accuracy",
		},
		[]string{"provider", "search_type"},
	)

	metrics.webSearchResponseTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "web_search_response_time_seconds",
			Help:    "Web search response time in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"provider", "search_type"},
	)

	metrics.webSearchSuccessRate = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "web_search_success_rate",
			Help: "Web search success rate",
		},
		[]string{"provider", "search_type"},
	)

	metrics.webSearchProviderFallbackRate = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "web_search_provider_fallback_total",
			Help: "Total number of web search provider fallbacks",
		},
		[]string{"primary_provider", "fallback_provider"},
	)

	metrics.mlModelAccuracy = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "ml_model_accuracy",
			Help: "ML model classification accuracy",
		},
		[]string{"model_type", "model_version"},
	)

	metrics.mlModelResponseTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ml_model_response_time_seconds",
			Help:    "ML model response time in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"model_type", "model_version"},
	)

	metrics.mlModelInferenceRate = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ml_model_inference_total",
			Help: "Total number of ML model inferences",
		},
		[]string{"model_type", "model_version", "status"},
	)

	metrics.mlModelFallbackRate = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ml_model_fallback_total",
			Help: "Total number of ML model fallbacks",
		},
		[]string{"model_type", "fallback_method"},
	)

	// Initialize Confidence Scoring Metrics
	metrics.confidenceScoreDistribution = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "confidence_score_distribution",
			Help:    "Distribution of confidence scores",
			Buckets: []float64{0.0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0},
		},
		[]string{"method", "confidence_range"},
	)

	metrics.confidenceScoreByMethod = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "confidence_score_by_method",
			Help: "Average confidence score by method",
		},
		[]string{"method", "time_window"},
	)

	metrics.confidenceScoreAccuracy = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "confidence_score_accuracy",
			Help: "Accuracy of confidence score predictions",
		},
		[]string{"method", "confidence_range"},
	)

	metrics.confidenceScoreTrend = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "confidence_score_trend",
			Help: "Confidence score trend over time",
		},
		[]string{"time_window", "method"},
	)

	// Initialize Geographic and Industry-Specific Metrics
	metrics.geographicClassificationAccuracy = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "geographic_classification_accuracy",
			Help: "Classification accuracy by geographic region",
		},
		[]string{"region", "country", "method"},
	)

	metrics.industrySpecificAccuracy = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "industry_specific_accuracy",
			Help: "Classification accuracy by industry",
		},
		[]string{"industry", "industry_code", "method"},
	)

	metrics.regionSpecificResponseTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "region_specific_response_time_seconds",
			Help:    "Response time by geographic region in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"region", "country", "method"},
	)

	metrics.industrySpecificResponseTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "industry_specific_response_time_seconds",
			Help:    "Response time by industry in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"industry", "industry_code", "method"},
	)

	// Initialize Quality Assurance Metrics
	metrics.qualityCheckPassRate = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "quality_check_pass_rate",
			Help: "Quality check pass rate",
		},
		[]string{"check_type", "method"},
	)

	metrics.qualityCheckFailureRate = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "quality_check_failure_total",
			Help: "Total number of quality check failures",
		},
		[]string{"check_type", "failure_reason"},
	)

	metrics.qualityCheckDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "quality_check_duration_seconds",
			Help:    "Quality check duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"check_type", "method"},
	)

	metrics.qualityCheckTrend = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "quality_check_trend",
			Help: "Quality check trend over time",
		},
		[]string{"time_window", "check_type"},
	)

	// Initialize Feedback and Validation Metrics
	metrics.feedbackCollectionRate = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "feedback_collection_total",
			Help: "Total number of feedback collections",
		},
		[]string{"feedback_type", "method"},
	)

	metrics.feedbackProcessingTime = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "feedback_processing_time_seconds",
			Help:    "Feedback processing time in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"feedback_type", "processing_stage"},
	)

	metrics.feedbackAccuracyImprovement = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "feedback_accuracy_improvement",
			Help: "Accuracy improvement from feedback",
		},
		[]string{"method", "time_window"},
	)

	metrics.feedbackValidationRate = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "feedback_validation_rate",
			Help: "Feedback validation rate",
		},
		[]string{"feedback_type", "validation_status"},
	)

	// Initialize Performance Optimization Metrics
	metrics.cachePerformanceMetrics = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cache_performance_metrics",
			Help: "Cache performance metrics",
		},
		[]string{"metric_type", "cache_type"},
	)

	metrics.modelOptimizationMetrics = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "model_optimization_metrics",
			Help: "Model optimization metrics",
		},
		[]string{"metric_type", "model_type"},
	)

	metrics.resourceOptimizationMetrics = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "resource_optimization_metrics",
			Help: "Resource optimization metrics",
		},
		[]string{"metric_type", "resource_type"},
	)

	metrics.scalabilityMetrics = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "scalability_metrics",
			Help: "Scalability metrics",
		},
		[]string{"metric_type", "scale_factor"},
	)

	// Register all metrics
	collectors := []prometheus.Collector{
		metrics.classificationAccuracyByMethod,
		metrics.classificationAccuracyByIndustry,
		metrics.classificationAccuracyByRegion,
		metrics.classificationAccuracyByConfidence,
		metrics.classificationAccuracyTrend,
		metrics.classificationResponseTimeByMethod,
		metrics.classificationResponseTimeByType,
		metrics.classificationResponseTimePercentile,
		metrics.classificationResponseTimeTrend,
		metrics.classificationMemoryUsage,
		metrics.classificationCPUUsage,
		metrics.classificationCacheHitRate,
		metrics.classificationCacheSize,
		metrics.classificationDatabaseQueries,
		metrics.classificationExternalAPICalls,
		metrics.userSatisfactionScore,
		metrics.userFeedbackCount,
		metrics.userFeedbackSentiment,
		metrics.userSatisfactionTrend,
		metrics.websiteAnalysisAccuracy,
		metrics.websiteAnalysisResponseTime,
		metrics.websiteAnalysisSuccessRate,
		metrics.websiteAnalysisPageDiscoveryRate,
		metrics.webSearchAccuracy,
		metrics.webSearchResponseTime,
		metrics.webSearchSuccessRate,
		metrics.webSearchProviderFallbackRate,
		metrics.mlModelAccuracy,
		metrics.mlModelResponseTime,
		metrics.mlModelInferenceRate,
		metrics.mlModelFallbackRate,
		metrics.confidenceScoreDistribution,
		metrics.confidenceScoreByMethod,
		metrics.confidenceScoreAccuracy,
		metrics.confidenceScoreTrend,
		metrics.geographicClassificationAccuracy,
		metrics.industrySpecificAccuracy,
		metrics.regionSpecificResponseTime,
		metrics.industrySpecificResponseTime,
		metrics.qualityCheckPassRate,
		metrics.qualityCheckFailureRate,
		metrics.qualityCheckDuration,
		metrics.qualityCheckTrend,
		metrics.feedbackCollectionRate,
		metrics.feedbackProcessingTime,
		metrics.feedbackAccuracyImprovement,
		metrics.feedbackValidationRate,
		metrics.cachePerformanceMetrics,
		metrics.modelOptimizationMetrics,
		metrics.resourceOptimizationMetrics,
		metrics.scalabilityMetrics,
	}

	for _, collector := range collectors {
		if err := prometheus.Register(collector); err != nil {
			return nil, fmt.Errorf("failed to register classification metric: %w", err)
		}
	}

	return metrics, nil
}

// RecordClassificationAccuracy records classification accuracy metrics
func (cm *ClassificationMetrics) RecordClassificationAccuracy(method, confidenceRange string, accuracy float64) {
	if !cm.config.MetricsEnabled {
		return
	}
	cm.classificationAccuracyByMethod.WithLabelValues(method, confidenceRange).Set(accuracy)
}

// RecordClassificationAccuracyByIndustry records classification accuracy by industry
func (cm *ClassificationMetrics) RecordClassificationAccuracyByIndustry(industry, industryCode string, accuracy float64) {
	if !cm.config.MetricsEnabled {
		return
	}
	cm.classificationAccuracyByIndustry.WithLabelValues(industry, industryCode).Set(accuracy)
}

// RecordClassificationAccuracyByRegion records classification accuracy by region
func (cm *ClassificationMetrics) RecordClassificationAccuracyByRegion(region, country string, accuracy float64) {
	if !cm.config.MetricsEnabled {
		return
	}
	cm.classificationAccuracyByRegion.WithLabelValues(region, country).Set(accuracy)
}

// RecordClassificationResponseTime records classification response time
func (cm *ClassificationMetrics) RecordClassificationResponseTime(method, confidenceRange string, duration time.Duration) {
	if !cm.config.MetricsEnabled {
		return
	}
	cm.classificationResponseTimeByMethod.WithLabelValues(method, confidenceRange).Observe(duration.Seconds())
}

// RecordClassificationResponseTimeByType records classification response time by type
func (cm *ClassificationMetrics) RecordClassificationResponseTimeByType(opType, method string, duration time.Duration) {
	if !cm.config.MetricsEnabled {
		return
	}
	cm.classificationResponseTimeByType.WithLabelValues(opType, method).Observe(duration.Seconds())
}

// RecordClassificationMemoryUsage records memory usage for classification
func (cm *ClassificationMetrics) RecordClassificationMemoryUsage(component, operation string, bytes int64) {
	if !cm.config.MetricsEnabled {
		return
	}
	cm.classificationMemoryUsage.WithLabelValues(component, operation).Set(float64(bytes))
}

// RecordClassificationCPUUsage records CPU usage for classification
func (cm *ClassificationMetrics) RecordClassificationCPUUsage(component, operation string, percentage float64) {
	if !cm.config.MetricsEnabled {
		return
	}
	cm.classificationCPUUsage.WithLabelValues(component, operation).Set(percentage)
}

// RecordClassificationCacheHitRate records cache hit rate
func (cm *ClassificationMetrics) RecordClassificationCacheHitRate(cacheType, method string, hitRate float64) {
	if !cm.config.MetricsEnabled {
		return
	}
	cm.classificationCacheHitRate.WithLabelValues(cacheType, method).Set(hitRate)
}

// RecordClassificationCacheSize records cache size
func (cm *ClassificationMetrics) RecordClassificationCacheSize(cacheType, unit string, size float64) {
	if !cm.config.MetricsEnabled {
		return
	}
	cm.classificationCacheSize.WithLabelValues(cacheType, unit).Set(size)
}

// RecordClassificationDatabaseQueries records database queries
func (cm *ClassificationMetrics) RecordClassificationDatabaseQueries(operation, table, status string) {
	if !cm.config.MetricsEnabled {
		return
	}
	cm.classificationDatabaseQueries.WithLabelValues(operation, table, status).Inc()
}

// RecordClassificationExternalAPICalls records external API calls
func (cm *ClassificationMetrics) RecordClassificationExternalAPICalls(service, endpoint, status string) {
	if !cm.config.MetricsEnabled {
		return
	}
	cm.classificationExternalAPICalls.WithLabelValues(service, endpoint, status).Inc()
}

// RecordUserSatisfactionScore records user satisfaction score
func (cm *ClassificationMetrics) RecordUserSatisfactionScore(method, confidenceRange string, score float64) {
	if !cm.config.MetricsEnabled {
		return
	}
	cm.userSatisfactionScore.WithLabelValues(method, confidenceRange).Set(score)
}

// RecordUserFeedbackCount records user feedback count
func (cm *ClassificationMetrics) RecordUserFeedbackCount(feedbackType, sentiment string) {
	if !cm.config.MetricsEnabled {
		return
	}
	cm.userFeedbackCount.WithLabelValues(feedbackType, sentiment).Inc()
}

// RecordUserFeedbackSentiment records user feedback sentiment
func (cm *ClassificationMetrics) RecordUserFeedbackSentiment(method, timeWindow string, sentiment float64) {
	if !cm.config.MetricsEnabled {
		return
	}
	cm.userFeedbackSentiment.WithLabelValues(method, timeWindow).Set(sentiment)
}

// RecordWebsiteAnalysisAccuracy records website analysis accuracy
func (cm *ClassificationMetrics) RecordWebsiteAnalysisAccuracy(pageType, contentQuality string, accuracy float64) {
	if !cm.config.MetricsEnabled {
		return
	}
	cm.websiteAnalysisAccuracy.WithLabelValues(pageType, contentQuality).Set(accuracy)
}

// RecordWebsiteAnalysisResponseTime records website analysis response time
func (cm *ClassificationMetrics) RecordWebsiteAnalysisResponseTime(pageType, discoveryDepth string, duration time.Duration) {
	if !cm.config.MetricsEnabled {
		return
	}
	cm.websiteAnalysisResponseTime.WithLabelValues(pageType, discoveryDepth).Observe(duration.Seconds())
}

// RecordWebsiteAnalysisSuccessRate records website analysis success rate
func (cm *ClassificationMetrics) RecordWebsiteAnalysisSuccessRate(pageType, contentQuality string, successRate float64) {
	if !cm.config.MetricsEnabled {
		return
	}
	cm.websiteAnalysisSuccessRate.WithLabelValues(pageType, contentQuality).Set(successRate)
}

// RecordWebsiteAnalysisPageDiscoveryRate records website analysis page discovery rate
func (cm *ClassificationMetrics) RecordWebsiteAnalysisPageDiscoveryRate(pageType, priorityLevel string, discoveryRate float64) {
	if !cm.config.MetricsEnabled {
		return
	}
	cm.websiteAnalysisPageDiscoveryRate.WithLabelValues(pageType, priorityLevel).Set(discoveryRate)
}

// RecordWebSearchAccuracy records web search accuracy
func (cm *ClassificationMetrics) RecordWebSearchAccuracy(provider, searchType string, accuracy float64) {
	if !cm.config.MetricsEnabled {
		return
	}
	cm.webSearchAccuracy.WithLabelValues(provider, searchType).Set(accuracy)
}

// RecordWebSearchResponseTime records web search response time
func (cm *ClassificationMetrics) RecordWebSearchResponseTime(provider, searchType string, duration time.Duration) {
	if !cm.config.MetricsEnabled {
		return
	}
	cm.webSearchResponseTime.WithLabelValues(provider, searchType).Observe(duration.Seconds())
}

// RecordWebSearchSuccessRate records web search success rate
func (cm *ClassificationMetrics) RecordWebSearchSuccessRate(provider, searchType string, successRate float64) {
	if !cm.config.MetricsEnabled {
		return
	}
	cm.webSearchSuccessRate.WithLabelValues(provider, searchType).Set(successRate)
}

// RecordWebSearchProviderFallback records web search provider fallback
func (cm *ClassificationMetrics) RecordWebSearchProviderFallback(primaryProvider, fallbackProvider string) {
	if !cm.config.MetricsEnabled {
		return
	}
	cm.webSearchProviderFallbackRate.WithLabelValues(primaryProvider, fallbackProvider).Inc()
}

// RecordMLModelAccuracy records ML model accuracy
func (cm *ClassificationMetrics) RecordMLModelAccuracy(modelType, modelVersion string, accuracy float64) {
	if !cm.config.MetricsEnabled {
		return
	}
	cm.mlModelAccuracy.WithLabelValues(modelType, modelVersion).Set(accuracy)
}

// RecordMLModelResponseTime records ML model response time
func (cm *ClassificationMetrics) RecordMLModelResponseTime(modelType, modelVersion string, duration time.Duration) {
	if !cm.config.MetricsEnabled {
		return
	}
	cm.mlModelResponseTime.WithLabelValues(modelType, modelVersion).Observe(duration.Seconds())
}

// RecordMLModelInference records ML model inference
func (cm *ClassificationMetrics) RecordMLModelInference(modelType, modelVersion, status string) {
	if !cm.config.MetricsEnabled {
		return
	}
	cm.mlModelInferenceRate.WithLabelValues(modelType, modelVersion, status).Inc()
}

// RecordMLModelFallback records ML model fallback
func (cm *ClassificationMetrics) RecordMLModelFallback(modelType, fallbackMethod string) {
	if !cm.config.MetricsEnabled {
		return
	}
	cm.mlModelFallbackRate.WithLabelValues(modelType, fallbackMethod).Inc()
}

// RecordConfidenceScoreDistribution records confidence score distribution
func (cm *ClassificationMetrics) RecordConfidenceScoreDistribution(method, confidenceRange string, score float64) {
	if !cm.config.MetricsEnabled {
		return
	}
	cm.confidenceScoreDistribution.WithLabelValues(method, confidenceRange).Observe(score)
}

// RecordConfidenceScoreByMethod records confidence score by method
func (cm *ClassificationMetrics) RecordConfidenceScoreByMethod(method, timeWindow string, score float64) {
	if !cm.config.MetricsEnabled {
		return
	}
	cm.confidenceScoreByMethod.WithLabelValues(method, timeWindow).Set(score)
}

// RecordConfidenceScoreAccuracy records confidence score accuracy
func (cm *ClassificationMetrics) RecordConfidenceScoreAccuracy(method, confidenceRange string, accuracy float64) {
	if !cm.config.MetricsEnabled {
		return
	}
	cm.confidenceScoreAccuracy.WithLabelValues(method, confidenceRange).Set(accuracy)
}

// RecordGeographicClassificationAccuracy records geographic classification accuracy
func (cm *ClassificationMetrics) RecordGeographicClassificationAccuracy(region, country, method string, accuracy float64) {
	if !cm.config.MetricsEnabled {
		return
	}
	cm.geographicClassificationAccuracy.WithLabelValues(region, country, method).Set(accuracy)
}

// RecordIndustrySpecificAccuracy records industry-specific accuracy
func (cm *ClassificationMetrics) RecordIndustrySpecificAccuracy(industry, industryCode, method string, accuracy float64) {
	if !cm.config.MetricsEnabled {
		return
	}
	cm.industrySpecificAccuracy.WithLabelValues(industry, industryCode, method).Set(accuracy)
}

// RecordRegionSpecificResponseTime records region-specific response time
func (cm *ClassificationMetrics) RecordRegionSpecificResponseTime(region, country, method string, duration time.Duration) {
	if !cm.config.MetricsEnabled {
		return
	}
	cm.regionSpecificResponseTime.WithLabelValues(region, country, method).Observe(duration.Seconds())
}

// RecordIndustrySpecificResponseTime records industry-specific response time
func (cm *ClassificationMetrics) RecordIndustrySpecificResponseTime(industry, industryCode, method string, duration time.Duration) {
	if !cm.config.MetricsEnabled {
		return
	}
	cm.industrySpecificResponseTime.WithLabelValues(industry, industryCode, method).Observe(duration.Seconds())
}

// RecordQualityCheckPassRate records quality check pass rate
func (cm *ClassificationMetrics) RecordQualityCheckPassRate(checkType, method string, passRate float64) {
	if !cm.config.MetricsEnabled {
		return
	}
	cm.qualityCheckPassRate.WithLabelValues(checkType, method).Set(passRate)
}

// RecordQualityCheckFailure records quality check failure
func (cm *ClassificationMetrics) RecordQualityCheckFailure(checkType, failureReason string) {
	if !cm.config.MetricsEnabled {
		return
	}
	cm.qualityCheckFailureRate.WithLabelValues(checkType, failureReason).Inc()
}

// RecordQualityCheckDuration records quality check duration
func (cm *ClassificationMetrics) RecordQualityCheckDuration(checkType, method string, duration time.Duration) {
	if !cm.config.MetricsEnabled {
		return
	}
	cm.qualityCheckDuration.WithLabelValues(checkType, method).Observe(duration.Seconds())
}

// RecordFeedbackCollection records feedback collection
func (cm *ClassificationMetrics) RecordFeedbackCollection(feedbackType, method string) {
	if !cm.config.MetricsEnabled {
		return
	}
	cm.feedbackCollectionRate.WithLabelValues(feedbackType, method).Inc()
}

// RecordFeedbackProcessingTime records feedback processing time
func (cm *ClassificationMetrics) RecordFeedbackProcessingTime(feedbackType, processingStage string, duration time.Duration) {
	if !cm.config.MetricsEnabled {
		return
	}
	cm.feedbackProcessingTime.WithLabelValues(feedbackType, processingStage).Observe(duration.Seconds())
}

// RecordFeedbackAccuracyImprovement records feedback accuracy improvement
func (cm *ClassificationMetrics) RecordFeedbackAccuracyImprovement(method, timeWindow string, improvement float64) {
	if !cm.config.MetricsEnabled {
		return
	}
	cm.feedbackAccuracyImprovement.WithLabelValues(method, timeWindow).Set(improvement)
}

// RecordFeedbackValidationRate records feedback validation rate
func (cm *ClassificationMetrics) RecordFeedbackValidationRate(feedbackType, validationStatus string, rate float64) {
	if !cm.config.MetricsEnabled {
		return
	}
	cm.feedbackValidationRate.WithLabelValues(feedbackType, validationStatus).Set(rate)
}

// RecordCachePerformanceMetrics records cache performance metrics
func (cm *ClassificationMetrics) RecordCachePerformanceMetrics(metricType, cacheType string, value float64) {
	if !cm.config.MetricsEnabled {
		return
	}
	cm.cachePerformanceMetrics.WithLabelValues(metricType, cacheType).Set(value)
}

// RecordModelOptimizationMetrics records model optimization metrics
func (cm *ClassificationMetrics) RecordModelOptimizationMetrics(metricType, modelType string, value float64) {
	if !cm.config.MetricsEnabled {
		return
	}
	cm.modelOptimizationMetrics.WithLabelValues(metricType, modelType).Set(value)
}

// RecordResourceOptimizationMetrics records resource optimization metrics
func (cm *ClassificationMetrics) RecordResourceOptimizationMetrics(metricType, resourceType string, value float64) {
	if !cm.config.MetricsEnabled {
		return
	}
	cm.resourceOptimizationMetrics.WithLabelValues(metricType, resourceType).Set(value)
}

// RecordScalabilityMetrics records scalability metrics
func (cm *ClassificationMetrics) RecordScalabilityMetrics(metricType, scaleFactor string, value float64) {
	if !cm.config.MetricsEnabled {
		return
	}
	cm.scalabilityMetrics.WithLabelValues(metricType, scaleFactor).Set(value)
}

// AddCustomClassificationMetric adds a custom classification metric
func (cm *ClassificationMetrics) AddCustomClassificationMetric(name string, collector prometheus.Collector) error {
	if !cm.config.MetricsEnabled {
		return nil
	}

	if err := prometheus.Register(collector); err != nil {
		return fmt.Errorf("failed to register custom classification metric %s: %w", name, err)
	}

	cm.customMetrics[name] = collector
	return nil
}

// RemoveCustomClassificationMetric removes a custom classification metric
func (cm *ClassificationMetrics) RemoveCustomClassificationMetric(name string) error {
	if !cm.config.MetricsEnabled {
		return nil
	}

	if collector, exists := cm.customMetrics[name]; exists {
		if prometheus.Unregister(collector) {
			delete(cm.customMetrics, name)
		}
	}

	return nil
}

// ServeHTTP serves the classification metrics endpoint
func (cm *ClassificationMetrics) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !cm.config.MetricsEnabled {
		http.Error(w, "Classification metrics disabled", http.StatusServiceUnavailable)
		return
	}

	promhttp.Handler().ServeHTTP(w, r)
}

// StartClassificationMetricsServer starts the classification metrics server
func (cm *ClassificationMetrics) StartClassificationMetricsServer(ctx context.Context) error {
	if !cm.config.MetricsEnabled {
		return nil
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics/classification", cm)

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cm.config.MetricsPort+1), // Use different port
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Classification metrics server error: %v\n", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return server.Shutdown(shutdownCtx)
}

// IsEnabled returns whether classification metrics are enabled
func (cm *ClassificationMetrics) IsEnabled() bool {
	return cm.config.MetricsEnabled
}

// String returns a string representation of the classification metrics configuration
func (cm *ClassificationMetrics) String() string {
	return fmt.Sprintf("ClassificationMetrics{enabled=%t, port=%d}", cm.config.MetricsEnabled, cm.config.MetricsPort+1)
}
