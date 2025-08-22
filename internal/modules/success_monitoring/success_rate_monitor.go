package success_monitoring

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// SuccessRateMonitor tracks and analyzes business processing success rates
type SuccessRateMonitor struct {
	config    *SuccessMonitorConfig
	logger    *zap.Logger
	metrics   map[string]*ProcessMetrics
	mu        sync.RWMutex
	startTime time.Time
	alerts    []SuccessAlert
}

// SuccessMonitorConfig holds configuration for success rate monitoring
type SuccessMonitorConfig struct {
	TargetSuccessRate        float64       `json:"target_success_rate"` // 0.95 = 95%
	WarningThreshold         float64       `json:"warning_threshold"`   // 0.90 = 90%
	CriticalThreshold        float64       `json:"critical_threshold"`  // 0.85 = 85%
	EnableRealTimeMonitoring bool          `json:"enable_real_time_monitoring"`
	EnableFailureAnalysis    bool          `json:"enable_failure_analysis"`
	EnableTrendAnalysis      bool          `json:"enable_trend_analysis"`
	EnableAlerting           bool          `json:"enable_alerting"`
	MetricsRetentionPeriod   time.Duration `json:"metrics_retention_period"` // 30 days
	AnalysisWindow           time.Duration `json:"analysis_window"`          // 1 hour
	TrendWindow              time.Duration `json:"trend_window"`             // 24 hours
	MinDataPoints            int           `json:"min_data_points"`          // 100
	MaxDataPoints            int           `json:"max_data_points"`          // 10000
	AlertCooldownPeriod      time.Duration `json:"alert_cooldown_period"`    // 5 minutes
}

// ProcessMetrics holds aggregated success rate metrics for a process
type ProcessMetrics struct {
	ProcessName         string                `json:"process_name"`
	TotalAttempts       int64                 `json:"total_attempts"`
	SuccessfulAttempts  int64                 `json:"successful_attempts"`
	FailedAttempts      int64                 `json:"failed_attempts"`
	SuccessRate         float64               `json:"success_rate"`
	AverageResponseTime time.Duration         `json:"average_response_time"`
	LastUpdated         time.Time             `json:"last_updated"`
	DataPoints          []ProcessingDataPoint `json:"data_points"`
	FailurePatterns     map[string]int        `json:"failure_patterns"`
	SuccessTrend        TrendDirection        `json:"success_trend"`
	LastAlertTime       *time.Time            `json:"last_alert_time"`
}

// ProcessingDataPoint represents a single business processing attempt
type ProcessingDataPoint struct {
	Timestamp       time.Time              `json:"timestamp"`
	ProcessName     string                 `json:"process_name"`
	InputType       string                 `json:"input_type"`
	Success         bool                   `json:"success"`
	ResponseTime    time.Duration          `json:"response_time"`
	StatusCode      int                    `json:"status_code"`
	ErrorType       string                 `json:"error_type,omitempty"`
	ErrorMessage    string                 `json:"error_message,omitempty"`
	ProcessingStage string                 `json:"processing_stage,omitempty"`
	InputSize       int                    `json:"input_size,omitempty"`
	OutputSize      int                    `json:"output_size,omitempty"`
	ConfidenceScore float64                `json:"confidence_score,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// SuccessAlert represents an alert for low success rates
type SuccessAlert struct {
	ID          string     `json:"id"`
	ProcessName string     `json:"process_name"`
	AlertType   AlertType  `json:"alert_type"`
	Message     string     `json:"message"`
	CurrentRate float64    `json:"current_rate"`
	TargetRate  float64    `json:"target_rate"`
	Timestamp   time.Time  `json:"timestamp"`
	Resolved    bool       `json:"resolved"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
}

// AlertType represents the type of success rate alert
type AlertType string

const (
	AlertTypeWarning  AlertType = "warning"
	AlertTypeCritical AlertType = "critical"
	AlertTypeInfo     AlertType = "info"
)

// TrendDirection represents the direction of success rate trends
type TrendDirection string

const (
	TrendImproving    TrendDirection = "improving"
	TrendDegrading    TrendDirection = "degrading"
	TrendStable       TrendDirection = "stable"
	TrendInsufficient TrendDirection = "insufficient_data"
)

// SuccessRateReport represents a comprehensive success rate report
type SuccessRateReport struct {
	ProcessName     string                  `json:"process_name"`
	GeneratedAt     time.Time               `json:"generated_at"`
	AnalysisWindow  time.Duration           `json:"analysis_window"`
	OverallMetrics  OverallMetrics          `json:"overall_metrics"`
	ProcessMetrics  []ProcessMetrics        `json:"process_metrics"`
	Trends          []TrendAnalysis         `json:"trends"`
	Alerts          []SuccessAlert          `json:"alerts"`
	Recommendations []SuccessRecommendation `json:"recommendations"`
}

// OverallMetrics represents overall success rate metrics
type OverallMetrics struct {
	TotalProcesses     int            `json:"total_processes"`
	AverageSuccessRate float64        `json:"average_success_rate"`
	BestProcess        string         `json:"best_process"`
	WorstProcess       string         `json:"worst_process"`
	TotalAttempts      int64          `json:"total_attempts"`
	TotalSuccesses     int64          `json:"total_successes"`
	TotalFailures      int64          `json:"total_failures"`
	OverallTrend       TrendDirection `json:"overall_trend"`
}

// SuccessRecommendation represents a recommendation for improving success rates
type SuccessRecommendation struct {
	ID          string    `json:"id"`
	ProcessName string    `json:"process_name"`
	Category    string    `json:"category"`
	Priority    int       `json:"priority"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Impact      string    `json:"impact"`
	Effort      string    `json:"effort"`
	Confidence  float64   `json:"confidence"`
	CreatedAt   time.Time `json:"created_at"`
}

// SuccessRateOptimization represents optimization strategies and improvements
type SuccessRateOptimization struct {
	ProcessName               string                    `json:"process_name"`
	GeneratedAt               time.Time                 `json:"generated_at"`
	CurrentSuccessRate        float64                   `json:"current_success_rate"`
	TargetSuccessRate         float64                   `json:"target_success_rate"`
	OptimizationGap           float64                   `json:"optimization_gap"`
	OptimizationStrategies    []OptimizationStrategy    `json:"optimization_strategies"`
	PerformanceTuning         PerformanceTuning         `json:"performance_tuning"`
	ProcessImprovements       []ProcessImprovement      `json:"process_improvements"`
	ResourceOptimization      ResourceOptimization      `json:"resource_optimization"`
	ConfigurationOptimization ConfigurationOptimization `json:"configuration_optimization"`
	ExpectedImprovement       float64                   `json:"expected_improvement"`
	ImplementationPlan        ImplementationPlan        `json:"implementation_plan"`
	OptimizationStatus        OptimizationStatus        `json:"optimization_status"`
}

// OptimizationStrategy represents a specific optimization approach
type OptimizationStrategy struct {
	ID                   string    `json:"id"`
	Name                 string    `json:"name"`
	Category             string    `json:"category"`
	Description          string    `json:"description"`
	ExpectedImpact       float64   `json:"expected_impact"`
	ImplementationEffort string    `json:"implementation_effort"`
	RiskLevel            string    `json:"risk_level"`
	Confidence           float64   `json:"confidence"`
	Prerequisites        []string  `json:"prerequisites"`
	Steps                []string  `json:"steps"`
	Metrics              []string  `json:"metrics"`
	CreatedAt            time.Time `json:"created_at"`
}

// PerformanceTuning represents performance optimization recommendations
type PerformanceTuning struct {
	ResponseTimeOptimization ResponseTimeOptimization `json:"response_time_optimization"`
	ThroughputOptimization   ThroughputOptimization   `json:"throughput_optimization"`
	ResourceOptimization     ResourceTuning           `json:"resource_optimization"`
	ConcurrencyOptimization  ConcurrencyOptimization  `json:"concurrency_optimization"`
	CacheOptimization        CacheOptimization        `json:"cache_optimization"`
}

// ResponseTimeOptimization represents response time improvement strategies
type ResponseTimeOptimization struct {
	CurrentAverageResponseTime time.Duration `json:"current_average_response_time"`
	TargetResponseTime         time.Duration `json:"target_response_time"`
	Bottlenecks                []Bottleneck  `json:"bottlenecks"`
	OptimizationStrategies     []string      `json:"optimization_strategies"`
	ExpectedImprovement        time.Duration `json:"expected_improvement"`
}

// Bottleneck represents a performance bottleneck
type Bottleneck struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Severity    string  `json:"severity"`
	Impact      float64 `json:"impact"`
	Description string  `json:"description"`
	Solution    string  `json:"solution"`
}

// ThroughputOptimization represents throughput improvement strategies
type ThroughputOptimization struct {
	CurrentThroughput      float64  `json:"current_throughput"`
	TargetThroughput       float64  `json:"target_throughput"`
	ThroughputGap          float64  `json:"throughput_gap"`
	OptimizationStrategies []string `json:"optimization_strategies"`
	ExpectedImprovement    float64  `json:"expected_improvement"`
}

// ResourceTuning represents resource optimization recommendations
type ResourceTuning struct {
	CPUOptimization     CPUOptimization     `json:"cpu_optimization"`
	MemoryOptimization  MemoryOptimization  `json:"memory_optimization"`
	NetworkOptimization NetworkOptimization `json:"network_optimization"`
	DiskOptimization    DiskOptimization    `json:"disk_optimization"`
}

// CPUOptimization represents CPU optimization strategies
type CPUOptimization struct {
	CurrentCPUUsage        float64  `json:"current_cpu_usage"`
	TargetCPUUsage         float64  `json:"target_cpu_usage"`
	OptimizationStrategies []string `json:"optimization_strategies"`
	ExpectedImprovement    float64  `json:"expected_improvement"`
}

// MemoryOptimization represents memory optimization strategies
type MemoryOptimization struct {
	CurrentMemoryUsage     float64  `json:"current_memory_usage"`
	TargetMemoryUsage      float64  `json:"target_memory_usage"`
	OptimizationStrategies []string `json:"optimization_strategies"`
	ExpectedImprovement    float64  `json:"expected_improvement"`
}

// NetworkOptimization represents network optimization strategies
type NetworkOptimization struct {
	CurrentNetworkLatency  time.Duration `json:"current_network_latency"`
	TargetNetworkLatency   time.Duration `json:"target_network_latency"`
	OptimizationStrategies []string      `json:"optimization_strategies"`
	ExpectedImprovement    time.Duration `json:"expected_improvement"`
}

// DiskOptimization represents disk optimization strategies
type DiskOptimization struct {
	CurrentDiskUsage       float64  `json:"current_disk_usage"`
	TargetDiskUsage        float64  `json:"target_disk_usage"`
	OptimizationStrategies []string `json:"optimization_strategies"`
	ExpectedImprovement    float64  `json:"expected_improvement"`
}

// ConcurrencyOptimization represents concurrency optimization strategies
type ConcurrencyOptimization struct {
	CurrentConcurrency     int      `json:"current_concurrency"`
	OptimalConcurrency     int      `json:"optimal_concurrency"`
	OptimizationStrategies []string `json:"optimization_strategies"`
	ExpectedImprovement    float64  `json:"expected_improvement"`
}

// CacheOptimization represents cache optimization strategies
type CacheOptimization struct {
	CurrentCacheHitRate    float64  `json:"current_cache_hit_rate"`
	TargetCacheHitRate     float64  `json:"target_cache_hit_rate"`
	OptimizationStrategies []string `json:"optimization_strategies"`
	ExpectedImprovement    float64  `json:"expected_improvement"`
}

// ProcessImprovement represents process improvement recommendations
type ProcessImprovement struct {
	ID                   string    `json:"id"`
	Name                 string    `json:"name"`
	Category             string    `json:"category"`
	Description          string    `json:"description"`
	CurrentState         string    `json:"current_state"`
	TargetState          string    `json:"target_state"`
	ImprovementSteps     []string  `json:"improvement_steps"`
	ExpectedImpact       float64   `json:"expected_impact"`
	ImplementationEffort string    `json:"implementation_effort"`
	RiskLevel            string    `json:"risk_level"`
	CreatedAt            time.Time `json:"created_at"`
}

// ResourceOptimization represents resource optimization recommendations
type ResourceOptimization struct {
	ScalingRecommendations ScalingRecommendations `json:"scaling_recommendations"`
	ResourceAllocation     ResourceAllocation     `json:"resource_allocation"`
	CostOptimization       CostOptimization       `json:"cost_optimization"`
	CapacityPlanning       CapacityPlanning       `json:"capacity_planning"`
}

// ScalingRecommendations represents scaling optimization strategies
type ScalingRecommendations struct {
	HorizontalScaling HorizontalScaling `json:"horizontal_scaling"`
	VerticalScaling   VerticalScaling   `json:"vertical_scaling"`
	AutoScaling       AutoScaling       `json:"auto_scaling"`
}

// HorizontalScaling represents horizontal scaling recommendations
type HorizontalScaling struct {
	RecommendedInstances int      `json:"recommended_instances"`
	ScalingTriggers      []string `json:"scaling_triggers"`
	ExpectedImprovement  float64  `json:"expected_improvement"`
	ImplementationSteps  []string `json:"implementation_steps"`
}

// VerticalScaling represents vertical scaling recommendations
type VerticalScaling struct {
	RecommendedResources ResourceSpec `json:"recommended_resources"`
	ScalingTriggers      []string     `json:"scaling_triggers"`
	ExpectedImprovement  float64      `json:"expected_improvement"`
	ImplementationSteps  []string     `json:"implementation_steps"`
}

// ResourceSpec represents resource specifications
type ResourceSpec struct {
	CPU     string `json:"cpu"`
	Memory  string `json:"memory"`
	Disk    string `json:"disk"`
	Network string `json:"network"`
}

// AutoScaling represents auto-scaling recommendations
type AutoScaling struct {
	MinInstances        int      `json:"min_instances"`
	MaxInstances        int      `json:"max_instances"`
	TargetCPUUsage      float64  `json:"target_cpu_usage"`
	ScalingPolicies     []string `json:"scaling_policies"`
	ExpectedImprovement float64  `json:"expected_improvement"`
}

// ResourceAllocation represents resource allocation optimization
type ResourceAllocation struct {
	CurrentAllocation      ResourceAllocationMap `json:"current_allocation"`
	OptimalAllocation      ResourceAllocationMap `json:"optimal_allocation"`
	OptimizationStrategies []string              `json:"optimization_strategies"`
	ExpectedImprovement    float64               `json:"expected_improvement"`
}

// ResourceAllocationMap represents resource allocation mapping
type ResourceAllocationMap struct {
	CPUAllocation     map[string]float64 `json:"cpu_allocation"`
	MemoryAllocation  map[string]float64 `json:"memory_allocation"`
	DiskAllocation    map[string]float64 `json:"disk_allocation"`
	NetworkAllocation map[string]float64 `json:"network_allocation"`
}

// CostOptimization represents cost optimization strategies
type CostOptimization struct {
	CurrentCost            float64  `json:"current_cost"`
	OptimizedCost          float64  `json:"optimized_cost"`
	CostSavings            float64  `json:"cost_savings"`
	OptimizationStrategies []string `json:"optimization_strategies"`
	ROI                    float64  `json:"roi"`
}

// CapacityPlanning represents capacity planning recommendations
type CapacityPlanning struct {
	CurrentCapacity   int                `json:"current_capacity"`
	RequiredCapacity  int                `json:"required_capacity"`
	CapacityGap       int                `json:"capacity_gap"`
	PlanningHorizon   time.Duration      `json:"planning_horizon"`
	GrowthProjections []GrowthProjection `json:"growth_projections"`
	Recommendations   []string           `json:"recommendations"`
}

// GrowthProjection represents growth projection data
type GrowthProjection struct {
	Timeframe        time.Duration `json:"timeframe"`
	ProjectedGrowth  float64       `json:"projected_growth"`
	RequiredCapacity int           `json:"required_capacity"`
	Confidence       float64       `json:"confidence"`
}

// ConfigurationOptimization represents configuration optimization recommendations
type ConfigurationOptimization struct {
	ParameterOptimization ParameterOptimization `json:"parameter_optimization"`
	ThresholdOptimization ThresholdOptimization `json:"threshold_optimization"`
	TimeoutOptimization   TimeoutOptimization   `json:"timeout_optimization"`
	RetryOptimization     RetryOptimization     `json:"retry_optimization"`
}

// ParameterOptimization represents parameter optimization strategies
type ParameterOptimization struct {
	CurrentParameters      map[string]interface{} `json:"current_parameters"`
	OptimalParameters      map[string]interface{} `json:"optimal_parameters"`
	OptimizationStrategies []string               `json:"optimization_strategies"`
	ExpectedImprovement    float64                `json:"expected_improvement"`
}

// ThresholdOptimization represents threshold optimization strategies
type ThresholdOptimization struct {
	CurrentThresholds      map[string]float64 `json:"current_thresholds"`
	OptimalThresholds      map[string]float64 `json:"optimal_thresholds"`
	OptimizationStrategies []string           `json:"optimization_strategies"`
	ExpectedImprovement    float64            `json:"expected_improvement"`
}

// TimeoutOptimization represents timeout optimization strategies
type TimeoutOptimization struct {
	CurrentTimeouts        map[string]time.Duration `json:"current_timeouts"`
	OptimalTimeouts        map[string]time.Duration `json:"optimal_timeouts"`
	OptimizationStrategies []string                 `json:"optimization_strategies"`
	ExpectedImprovement    time.Duration            `json:"expected_improvement"`
}

// RetryOptimization represents retry optimization strategies
type RetryOptimization struct {
	CurrentRetryConfig     RetryConfig `json:"current_retry_config"`
	OptimalRetryConfig     RetryConfig `json:"optimal_retry_config"`
	OptimizationStrategies []string    `json:"optimization_strategies"`
	ExpectedImprovement    float64     `json:"expected_improvement"`
}

// RetryConfig represents retry configuration
type RetryConfig struct {
	MaxRetries      int           `json:"max_retries"`
	RetryDelay      time.Duration `json:"retry_delay"`
	BackoffFactor   float64       `json:"backoff_factor"`
	RetryableErrors []string      `json:"retryable_errors"`
}

// ImplementationPlan represents the implementation plan for optimizations
type ImplementationPlan struct {
	Phases          []ImplementationPhase `json:"phases"`
	TotalDuration   time.Duration         `json:"total_duration"`
	TotalEffort     string                `json:"total_effort"`
	RiskAssessment  RiskAssessment        `json:"risk_assessment"`
	SuccessCriteria []string              `json:"success_criteria"`
	RollbackPlan    RollbackPlan          `json:"rollback_plan"`
}

// ImplementationPhase represents an implementation phase
type ImplementationPhase struct {
	PhaseNumber     int           `json:"phase_number"`
	Name            string        `json:"name"`
	Description     string        `json:"description"`
	Duration        time.Duration `json:"duration"`
	Effort          string        `json:"effort"`
	Prerequisites   []string      `json:"prerequisites"`
	Deliverables    []string      `json:"deliverables"`
	SuccessCriteria []string      `json:"success_criteria"`
	RiskLevel       string        `json:"risk_level"`
}

// RiskAssessment represents risk assessment for optimization
type RiskAssessment struct {
	OverallRiskLevel     string       `json:"overall_risk_level"`
	RiskFactors          []RiskFactor `json:"risk_factors"`
	MitigationStrategies []string     `json:"mitigation_strategies"`
	ContingencyPlans     []string     `json:"contingency_plans"`
}

// RiskFactor represents a risk factor
type RiskFactor struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	RiskLevel   string  `json:"risk_level"`
	Probability float64 `json:"probability"`
	Impact      float64 `json:"impact"`
	Description string  `json:"description"`
	Mitigation  string  `json:"mitigation"`
}

// RollbackPlan represents rollback plan for optimization
type RollbackPlan struct {
	RollbackTriggers []string      `json:"rollback_triggers"`
	RollbackSteps    []string      `json:"rollback_steps"`
	RollbackDuration time.Duration `json:"rollback_duration"`
	DataBackup       string        `json:"data_backup"`
}

// OptimizationStatus represents the status of optimization
type OptimizationStatus struct {
	Status            string    `json:"status"`
	LastOptimized     time.Time `json:"last_optimized"`
	OptimizationCount int       `json:"optimization_count"`
	SuccessRate       float64   `json:"success_rate"`
	Improvement       float64   `json:"improvement"`
	NextOptimization  time.Time `json:"next_optimization"`
}

// FailureAnalysis represents detailed failure analysis results
type FailureAnalysis struct {
	ProcessName             string                  `json:"process_name"`
	AnalysisWindow          time.Duration           `json:"analysis_window"`
	TotalFailures           int64                   `json:"total_failures"`
	FailureRate             float64                 `json:"failure_rate"`
	CommonErrorTypes        map[string]int          `json:"common_error_types"`
	CommonErrorMessages     map[string]int          `json:"common_error_messages"`
	ProblematicInputTypes   map[string]int          `json:"problematic_input_types"`
	ProcessingStageFailures map[string]int          `json:"processing_stage_failures"`
	FailureTrend            TrendDirection          `json:"failure_trend"`
	Recommendations         []FailureRecommendation `json:"recommendations"`
	Timestamp               time.Time               `json:"timestamp"`

	// Enhanced failure analysis fields
	RootCauseAnalysis    *RootCauseAnalysis    `json:"root_cause_analysis"`
	FailurePatterns      []FailurePattern      `json:"failure_patterns"`
	ErrorCategorization  ErrorCategorization   `json:"error_categorization"`
	TemporalAnalysis     TemporalAnalysis      `json:"temporal_analysis"`
	CorrelationAnalysis  CorrelationAnalysis   `json:"correlation_analysis"`
	ImpactAssessment     ImpactAssessment      `json:"impact_assessment"`
	ActionableInsights   []ActionableInsight   `json:"actionable_insights"`
	FailureTimeline      []FailureEvent        `json:"failure_timeline"`
	SimilarityAnalysis   SimilarityAnalysis    `json:"similarity_analysis"`
	PredictiveIndicators []PredictiveIndicator `json:"predictive_indicators"`
}

// FailureRecommendation represents a recommendation for improving success rates
type FailureRecommendation struct {
	Type                 string  `json:"type"`
	Description          string  `json:"description"`
	Impact               string  `json:"impact"`
	Effort               string  `json:"effort"`
	Priority             int     `json:"priority"`
	EstimatedImprovement float64 `json:"estimated_improvement"`
}

// RootCauseAnalysis represents the root cause analysis of failures
type RootCauseAnalysis struct {
	PrimaryRootCause     RootCause         `json:"primary_root_cause"`
	SecondaryRootCauses  []RootCause       `json:"secondary_root_causes"`
	CauseConfidence      float64           `json:"cause_confidence"`
	ContributingFactors  []Contributing    `json:"contributing_factors"`
	SystemicIssues       []SystemicIssue   `json:"systemic_issues"`
	EnvironmentalFactors []Environmental   `json:"environmental_factors"`
	HumanFactors         []HumanFactor     `json:"human_factors"`
	TechnicalFactors     []TechnicalFactor `json:"technical_factors"`
}

// RootCause represents a specific root cause
type RootCause struct {
	ID          string    `json:"id"`
	Category    string    `json:"category"`    // "technical", "process", "environmental", "human"
	Subcategory string    `json:"subcategory"` // specific area
	Description string    `json:"description"`
	Evidence    []string  `json:"evidence"`
	Confidence  float64   `json:"confidence"`
	Impact      string    `json:"impact"`    // "high", "medium", "low"
	Frequency   int       `json:"frequency"` // how often this cause appears
	FirstSeen   time.Time `json:"first_seen"`
	LastSeen    time.Time `json:"last_seen"`
}

// Contributing represents a contributing factor
type Contributing struct {
	Factor      string  `json:"factor"`
	Weight      float64 `json:"weight"` // 0-1.0
	Description string  `json:"description"`
	Category    string  `json:"category"`
}

// SystemicIssue represents a systemic issue
type SystemicIssue struct {
	Issue         string    `json:"issue"`
	Scope         string    `json:"scope"`    // "global", "service", "component"
	Severity      string    `json:"severity"` // "critical", "high", "medium", "low"
	AffectedAreas []string  `json:"affected_areas"`
	FirstDetected time.Time `json:"first_detected"`
}

// Environmental represents environmental factors
type Environmental struct {
	Factor      string        `json:"factor"`
	Description string        `json:"description"`
	Impact      string        `json:"impact"`
	Duration    time.Duration `json:"duration"`
	Frequency   string        `json:"frequency"` // "constant", "intermittent", "rare"
}

// HumanFactor represents human-related factors
type HumanFactor struct {
	Factor      string `json:"factor"`
	Category    string `json:"category"` // "configuration", "operation", "process"
	Description string `json:"description"`
	Prevention  string `json:"prevention"`
	Training    bool   `json:"training_needed"`
}

// TechnicalFactor represents technical factors
type TechnicalFactor struct {
	Component  string `json:"component"`
	Issue      string `json:"issue"`
	Severity   string `json:"severity"`
	Resolution string `json:"resolution"`
	Prevention string `json:"prevention"`
	TechDebt   bool   `json:"technical_debt"`
}

// FailurePattern represents a detected failure pattern
type FailurePattern struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Pattern         string                 `json:"pattern"`    // regex or description
	Frequency       int                    `json:"frequency"`  // occurrences
	Confidence      float64                `json:"confidence"` // 0-1.0
	Severity        string                 `json:"severity"`   // "critical", "high", "medium", "low"
	Category        string                 `json:"category"`   // "input", "processing", "output", "system"
	Characteristics map[string]interface{} `json:"characteristics"`
	Examples        []string               `json:"examples"`
	FirstDetected   time.Time              `json:"first_detected"`
	LastDetected    time.Time              `json:"last_detected"`
	Trend           TrendDirection         `json:"trend"`
	RelatedPatterns []string               `json:"related_patterns"`
}

// ErrorCategorization represents categorized error analysis
type ErrorCategorization struct {
	UserErrors          ErrorCategory `json:"user_errors"`
	SystemErrors        ErrorCategory `json:"system_errors"`
	IntegrationErrors   ErrorCategory `json:"integration_errors"`
	ValidationErrors    ErrorCategory `json:"validation_errors"`
	TimeoutErrors       ErrorCategory `json:"timeout_errors"`
	ResourceErrors      ErrorCategory `json:"resource_errors"`
	SecurityErrors      ErrorCategory `json:"security_errors"`
	DataErrors          ErrorCategory `json:"data_errors"`
	NetworkErrors       ErrorCategory `json:"network_errors"`
	ConfigurationErrors ErrorCategory `json:"configuration_errors"`
}

// ErrorCategory represents a category of errors
type ErrorCategory struct {
	Count           int                    `json:"count"`
	Percentage      float64                `json:"percentage"`
	TrendDirection  TrendDirection         `json:"trend_direction"`
	CommonMessages  []string               `json:"common_messages"`
	Severity        string                 `json:"severity"`
	ImpactLevel     string                 `json:"impact_level"`
	ResolutionTime  time.Duration          `json:"avg_resolution_time"`
	RecurrenceRate  float64                `json:"recurrence_rate"`
	Examples        []string               `json:"examples"`
	Recommendations []string               `json:"recommendations"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// TemporalAnalysis represents time-based failure analysis
type TemporalAnalysis struct {
	HourlyDistribution map[int]int        `json:"hourly_distribution"` // 0-23 hours
	DailyDistribution  map[string]int     `json:"daily_distribution"`  // Mon-Sun
	WeeklyTrends       map[string]float64 `json:"weekly_trends"`       // week numbers
	MonthlyTrends      map[string]float64 `json:"monthly_trends"`      // months
	SeasonalPatterns   []SeasonalPattern  `json:"seasonal_patterns"`
	PeakFailureTimes   []PeakTime         `json:"peak_failure_times"`
	FailureFrequency   map[string]int     `json:"failure_frequency"` // frequency distribution
	Duration           TemporalDuration   `json:"duration"`
	Cyclical           []CyclicalPattern  `json:"cyclical_patterns"`
}

// SeasonalPattern represents seasonal failure patterns
type SeasonalPattern struct {
	Season     string    `json:"season"` // "spring", "summer", "fall", "winter"
	Pattern    string    `json:"pattern"`
	Impact     string    `json:"impact"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	Confidence float64   `json:"confidence"`
}

// PeakTime represents peak failure times
type PeakTime struct {
	TimeRange   string  `json:"time_range"` // "09:00-10:00"
	FailureRate float64 `json:"failure_rate"`
	Volume      int     `json:"volume"`
	Reason      string  `json:"reason"`
}

// TemporalDuration represents temporal duration analysis
type TemporalDuration struct {
	AverageFailureDuration time.Duration `json:"average_failure_duration"`
	MedianFailureDuration  time.Duration `json:"median_failure_duration"`
	MaxFailureDuration     time.Duration `json:"max_failure_duration"`
	MinFailureDuration     time.Duration `json:"min_failure_duration"`
}

// CyclicalPattern represents cyclical patterns
type CyclicalPattern struct {
	Cycle       string        `json:"cycle"` // "daily", "weekly", "monthly"
	Period      time.Duration `json:"period"`
	Amplitude   float64       `json:"amplitude"`
	Phase       float64       `json:"phase"`
	Confidence  float64       `json:"confidence"`
	Description string        `json:"description"`
}

// CorrelationAnalysis represents correlation analysis between failures and other factors
type CorrelationAnalysis struct {
	LoadCorrelation          float64                  `json:"load_correlation"`
	TimeCorrelation          map[string]float64       `json:"time_correlation"`
	InputSizeCorrelation     float64                  `json:"input_size_correlation"`
	ProcessStageCorrelation  map[string]float64       `json:"process_stage_correlation"`
	ErrorTypeCorrelation     map[string]float64       `json:"error_type_correlation"`
	ResponseTimeCorrelation  float64                  `json:"response_time_correlation"`
	ConcurrencyCorrelation   float64                  `json:"concurrency_correlation"`
	ExternalFactors          []ExternalFactor         `json:"external_factors"`
	DependencyCorrelation    map[string]float64       `json:"dependency_correlation"`
	ResourceCorrelation      ResourceCorrelation      `json:"resource_correlation"`
	EnvironmentalCorrelation EnvironmentalCorrelation `json:"environmental_correlation"`
}

// ExternalFactor represents external factors that correlate with failures
type ExternalFactor struct {
	Factor      string  `json:"factor"`
	Correlation float64 `json:"correlation"`
	Impact      string  `json:"impact"`
	Source      string  `json:"source"`
	Confidence  float64 `json:"confidence"`
}

// ResourceCorrelation represents correlation with resource usage
type ResourceCorrelation struct {
	CPUCorrelation     float64 `json:"cpu_correlation"`
	MemoryCorrelation  float64 `json:"memory_correlation"`
	DiskCorrelation    float64 `json:"disk_correlation"`
	NetworkCorrelation float64 `json:"network_correlation"`
}

// EnvironmentalCorrelation represents correlation with environmental factors
type EnvironmentalCorrelation struct {
	TimeOfDayCorrelation float64 `json:"time_of_day_correlation"`
	DayOfWeekCorrelation float64 `json:"day_of_week_correlation"`
	LoadCorrelation      float64 `json:"load_correlation"`
	WeatherCorrelation   float64 `json:"weather_correlation"`
}

// ImpactAssessment represents impact assessment of failures
type ImpactAssessment struct {
	BusinessImpact     BusinessImpact     `json:"business_impact"`
	TechnicalImpact    TechnicalImpact    `json:"technical_impact"`
	UserImpact         UserImpact         `json:"user_impact"`
	FinancialImpact    FinancialImpact    `json:"financial_impact"`
	ReputationalImpact ReputationalImpact `json:"reputational_impact"`
	OperationalImpact  OperationalImpact  `json:"operational_impact"`
	SecurityImpact     SecurityImpact     `json:"security_impact"`
	ComplianceImpact   ComplianceImpact   `json:"compliance_impact"`
}

// BusinessImpact represents business impact
type BusinessImpact struct {
	Severity               string        `json:"severity"`
	AffectedProcesses      []string      `json:"affected_processes"`
	BusinessContinuity     string        `json:"business_continuity"`
	CustomerSatisfaction   float64       `json:"customer_satisfaction"`
	ServiceLevelAgreement  string        `json:"sla_impact"`
	RevenueLoss            float64       `json:"estimated_revenue_loss"`
	ProductivityLoss       float64       `json:"productivity_loss"`
	RecoveryTimeObjective  time.Duration `json:"recovery_time_objective"`
	RecoveryPointObjective time.Duration `json:"recovery_point_objective"`
}

// TechnicalImpact represents technical impact
type TechnicalImpact struct {
	SystemAvailability     float64  `json:"system_availability"`
	PerformanceDegradation float64  `json:"performance_degradation"`
	DataIntegrity          string   `json:"data_integrity"`
	SystemStability        string   `json:"system_stability"`
	AffectedComponents     []string `json:"affected_components"`
	CascadingFailures      bool     `json:"cascading_failures"`
	RecoveryComplexity     string   `json:"recovery_complexity"`
}

// UserImpact represents user impact
type UserImpact struct {
	AffectedUsers       int     `json:"affected_users"`
	UserExperience      string  `json:"user_experience"`
	ServiceDisruption   string  `json:"service_disruption"`
	FeatureAvailability float64 `json:"feature_availability"`
	UserSatisfaction    float64 `json:"user_satisfaction"`
	SupportTickets      int     `json:"support_tickets"`
}

// FinancialImpact represents financial impact
type FinancialImpact struct {
	DirectCosts        float64 `json:"direct_costs"`
	IndirectCosts      float64 `json:"indirect_costs"`
	OpportunityCosts   float64 `json:"opportunity_costs"`
	RecoveryCosts      float64 `json:"recovery_costs"`
	PenaltyCosts       float64 `json:"penalty_costs"`
	TotalEstimatedCost float64 `json:"total_estimated_cost"`
	CostPerIncident    float64 `json:"cost_per_incident"`
}

// ReputationalImpact represents reputational impact
type ReputationalImpact struct {
	BrandImpact          string  `json:"brand_impact"`
	CustomerTrust        float64 `json:"customer_trust"`
	MediaCoverage        string  `json:"media_coverage"`
	SocialMediaSentiment float64 `json:"social_media_sentiment"`
	CompetitiveImpact    string  `json:"competitive_impact"`
}

// OperationalImpact represents operational impact
type OperationalImpact struct {
	StaffProductivity      float64 `json:"staff_productivity"`
	ProcessEfficiency      float64 `json:"process_efficiency"`
	ResourceUtilization    float64 `json:"resource_utilization"`
	OperationalOverhead    float64 `json:"operational_overhead"`
	MaintenanceRequirement string  `json:"maintenance_requirement"`
}

// SecurityImpact represents security impact
type SecurityImpact struct {
	SecurityLevel         string   `json:"security_level"`
	VulnerabilityExposure string   `json:"vulnerability_exposure"`
	DataExposure          string   `json:"data_exposure"`
	AccessControl         string   `json:"access_control"`
	ComplianceStatus      string   `json:"compliance_status"`
	ThreatLevel           string   `json:"threat_level"`
	AffectedAssets        []string `json:"affected_assets"`
}

// ComplianceImpact represents compliance impact
type ComplianceImpact struct {
	RegulationsAffected  []string `json:"regulations_affected"`
	ComplianceRisk       string   `json:"compliance_risk"`
	AuditImpact          string   `json:"audit_impact"`
	ReportingRequirement string   `json:"reporting_requirement"`
	PenaltyRisk          float64  `json:"penalty_risk"`
}

// ActionableInsight represents an actionable insight
type ActionableInsight struct {
	ID              string    `json:"id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	Category        string    `json:"category"`   // "immediate", "short-term", "long-term"
	Priority        int       `json:"priority"`   // 1-5
	Urgency         string    `json:"urgency"`    // "critical", "high", "medium", "low"
	Impact          string    `json:"impact"`     // "high", "medium", "low"
	Effort          string    `json:"effort"`     // "low", "medium", "high"
	Confidence      float64   `json:"confidence"` // 0-1.0
	Evidence        []string  `json:"evidence"`
	Actions         []string  `json:"actions"`
	ExpectedOutcome string    `json:"expected_outcome"`
	Timeline        string    `json:"timeline"`
	Owner           string    `json:"owner"`
	Dependencies    []string  `json:"dependencies"`
	Risks           []string  `json:"risks"`
	Success         []string  `json:"success_metrics"`
	CreatedAt       time.Time `json:"created_at"`
}

// FailureEvent represents a failure event in the timeline
type FailureEvent struct {
	Timestamp       time.Time              `json:"timestamp"`
	EventType       string                 `json:"event_type"` // "failure", "recovery", "escalation"
	Severity        string                 `json:"severity"`
	Description     string                 `json:"description"`
	ProcessName     string                 `json:"process_name"`
	ProcessingStage string                 `json:"processing_stage"`
	ErrorType       string                 `json:"error_type"`
	ErrorMessage    string                 `json:"error_message"`
	Impact          string                 `json:"impact"`
	Duration        time.Duration          `json:"duration"`
	Resolution      string                 `json:"resolution"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// SimilarityAnalysis represents similarity analysis between failures
type SimilarityAnalysis struct {
	SimilarFailures   []SimilarFailure   `json:"similar_failures"`
	FailureClusters   []FailureCluster   `json:"failure_clusters"`
	AnomalousFailures []AnomalousFailure `json:"anomalous_failures"`
	PatternMatches    []PatternMatch     `json:"pattern_matches"`
}

// SimilarFailure represents a similar failure
type SimilarFailure struct {
	FailureID      string    `json:"failure_id"`
	Similarity     float64   `json:"similarity"` // 0-1.0
	CommonFeatures []string  `json:"common_features"`
	Differences    []string  `json:"differences"`
	Resolution     string    `json:"resolution"`
	Timestamp      time.Time `json:"timestamp"`
}

// FailureCluster represents a cluster of similar failures
type FailureCluster struct {
	ClusterID       string    `json:"cluster_id"`
	FailureCount    int       `json:"failure_count"`
	Centroid        string    `json:"centroid"`
	CommonFeatures  []string  `json:"common_features"`
	Variance        float64   `json:"variance"`
	FirstOccurrence time.Time `json:"first_occurrence"`
	LastOccurrence  time.Time `json:"last_occurrence"`
}

// AnomalousFailure represents an anomalous failure
type AnomalousFailure struct {
	FailureID      string    `json:"failure_id"`
	AnomalyScore   float64   `json:"anomaly_score"` // 0-1.0
	UniqueFeatures []string  `json:"unique_features"`
	Investigation  string    `json:"investigation_status"`
	Timestamp      time.Time `json:"timestamp"`
}

// PatternMatch represents a pattern match
type PatternMatch struct {
	PatternID       string   `json:"pattern_id"`
	MatchConfidence float64  `json:"match_confidence"` // 0-1.0
	MatchedFeatures []string `json:"matched_features"`
	Variations      []string `json:"variations"`
}

// PredictiveIndicator represents a predictive indicator
type PredictiveIndicator struct {
	IndicatorType   string         `json:"indicator_type"` // "early_warning", "risk_factor", "trend"
	Name            string         `json:"name"`
	Description     string         `json:"description"`
	CurrentValue    float64        `json:"current_value"`
	ThresholdValue  float64        `json:"threshold_value"`
	Confidence      float64        `json:"confidence"` // 0-1.0
	Trend           TrendDirection `json:"trend"`
	TimeToImpact    time.Duration  `json:"time_to_impact"`
	Severity        string         `json:"severity"`
	Recommendations []string       `json:"recommendations"`
	LastUpdated     time.Time      `json:"last_updated"`
}

// TrendAnalysis represents trend analysis results
type TrendAnalysis struct {
	ProcessName       string         `json:"process_name"`
	AnalysisWindow    time.Duration  `json:"analysis_window"`
	SuccessRateTrend  float64        `json:"success_rate_trend"`
	VolumeTrend       float64        `json:"volume_trend"`
	ResponseTimeTrend float64        `json:"response_time_trend"`
	TrendDirection    TrendDirection `json:"trend_direction"`
	Confidence        float64        `json:"confidence"`
	Predictions       []Prediction   `json:"predictions"`
	Timestamp         time.Time      `json:"timestamp"`
}

// Prediction represents a future prediction
type Prediction struct {
	Timeframe   time.Duration `json:"timeframe"`
	SuccessRate float64       `json:"success_rate"`
	Confidence  float64       `json:"confidence"`
	Factors     []string      `json:"factors"`
}

// NewSuccessRateMonitor creates a new success rate monitor
func NewSuccessRateMonitor(config *SuccessMonitorConfig, logger *zap.Logger) *SuccessRateMonitor {
	if config == nil {
		config = DefaultSuccessMonitorConfig()
	}

	monitor := &SuccessRateMonitor{
		config:    config,
		logger:    logger,
		metrics:   make(map[string]*ProcessMetrics),
		startTime: time.Now(),
		alerts:    make([]SuccessAlert, 0),
	}

	// Start background monitoring if enabled
	if config.EnableRealTimeMonitoring {
		go monitor.startBackgroundMonitoring()
	}

	return monitor
}

// DefaultSuccessMonitorConfig returns default configuration
func DefaultSuccessMonitorConfig() *SuccessMonitorConfig {
	return &SuccessMonitorConfig{
		TargetSuccessRate:        0.95, // 95%
		WarningThreshold:         0.90, // 90%
		CriticalThreshold:        0.85, // 85%
		EnableRealTimeMonitoring: true,
		EnableFailureAnalysis:    true,
		EnableTrendAnalysis:      true,
		EnableAlerting:           true,
		MetricsRetentionPeriod:   30 * 24 * time.Hour, // 30 days
		AnalysisWindow:           1 * time.Hour,       // 1 hour
		TrendWindow:              24 * time.Hour,      // 24 hours
		MinDataPoints:            100,
		MaxDataPoints:            10000,
		AlertCooldownPeriod:      5 * time.Minute, // 5 minutes
	}
}

// RecordProcessingAttempt records a business processing attempt
func (srm *SuccessRateMonitor) RecordProcessingAttempt(ctx context.Context, dataPoint ProcessingDataPoint) error {
	srm.mu.Lock()
	defer srm.mu.Unlock()

	// Add timestamp if not set
	if dataPoint.Timestamp.IsZero() {
		dataPoint.Timestamp = time.Now()
	}

	// Get or create process metrics
	metrics := srm.getOrCreateProcessMetrics(dataPoint.ProcessName)

	// Update metrics
	metrics.TotalAttempts++
	if dataPoint.Success {
		metrics.SuccessfulAttempts++
	} else {
		metrics.FailedAttempts++
		// Track failure patterns
		if dataPoint.ErrorType != "" {
			metrics.FailurePatterns[dataPoint.ErrorType]++
		}
	}

	// Calculate success rate
	if metrics.TotalAttempts > 0 {
		metrics.SuccessRate = float64(metrics.SuccessfulAttempts) / float64(metrics.TotalAttempts)
	}

	// Update average response time
	if metrics.TotalAttempts == 1 {
		metrics.AverageResponseTime = dataPoint.ResponseTime
	} else {
		// Weighted average
		total := metrics.AverageResponseTime * time.Duration(metrics.TotalAttempts-1)
		metrics.AverageResponseTime = (total + dataPoint.ResponseTime) / time.Duration(metrics.TotalAttempts)
	}

	// Add data point
	metrics.DataPoints = append(metrics.DataPoints, dataPoint)
	metrics.LastUpdated = time.Now()

	// Cleanup old data points
	srm.cleanupOldDataPoints(metrics)

	// Update success trend
	metrics.SuccessTrend = srm.calculateSuccessTrend(metrics)

	// Check for alerts
	if srm.config.EnableAlerting {
		srm.checkAlerts(ctx, metrics)
	}

	srm.logger.Debug("Recorded processing attempt",
		zap.String("process", dataPoint.ProcessName),
		zap.String("input_type", dataPoint.InputType),
		zap.Bool("success", dataPoint.Success),
		zap.Duration("response_time", dataPoint.ResponseTime),
		zap.Float64("current_success_rate", metrics.SuccessRate))

	return nil
}

// GetProcessMetrics returns current metrics for a process
func (srm *SuccessRateMonitor) GetProcessMetrics(processName string) *ProcessMetrics {
	srm.mu.RLock()
	defer srm.mu.RUnlock()

	metrics, exists := srm.metrics[processName]
	if !exists {
		return nil
	}

	// Return a copy to avoid race conditions
	metricsCopy := *metrics
	metricsCopy.DataPoints = make([]ProcessingDataPoint, len(metrics.DataPoints))
	copy(metricsCopy.DataPoints, metrics.DataPoints)

	return &metricsCopy
}

// GetAllProcessMetrics returns metrics for all processes
func (srm *SuccessRateMonitor) GetAllProcessMetrics() map[string]*ProcessMetrics {
	srm.mu.RLock()
	defer srm.mu.RUnlock()

	result := make(map[string]*ProcessMetrics)
	for processName, metrics := range srm.metrics {
		// Return a copy to avoid race conditions
		metricsCopy := *metrics
		metricsCopy.DataPoints = make([]ProcessingDataPoint, len(metrics.DataPoints))
		copy(metricsCopy.DataPoints, metrics.DataPoints)
		result[processName] = &metricsCopy
	}

	return result
}

// AnalyzeFailures performs comprehensive failure analysis for a process with root cause identification
func (srm *SuccessRateMonitor) AnalyzeFailures(ctx context.Context, processName string) (*FailureAnalysis, error) {
	srm.mu.RLock()
	metrics, exists := srm.metrics[processName]
	srm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("process %s not found", processName)
	}

	// Get data points within analysis window
	endTime := time.Now()
	startTime := endTime.Add(-srm.config.AnalysisWindow)

	var relevantDataPoints []ProcessingDataPoint
	for _, dp := range metrics.DataPoints {
		if dp.Timestamp.After(startTime) && dp.Timestamp.Before(endTime) {
			relevantDataPoints = append(relevantDataPoints, dp)
		}
	}

	if len(relevantDataPoints) < srm.config.MinDataPoints {
		return nil, fmt.Errorf("insufficient data points for analysis: %d < %d", len(relevantDataPoints), srm.config.MinDataPoints)
	}

	analysis := &FailureAnalysis{
		ProcessName:             processName,
		AnalysisWindow:          srm.config.AnalysisWindow,
		CommonErrorTypes:        make(map[string]int),
		CommonErrorMessages:     make(map[string]int),
		ProblematicInputTypes:   make(map[string]int),
		ProcessingStageFailures: make(map[string]int),
		Timestamp:               time.Now(),
	}

	// Basic failure counting (existing logic)
	var failureDataPoints []ProcessingDataPoint
	for _, dp := range relevantDataPoints {
		if !dp.Success {
			analysis.TotalFailures++
			failureDataPoints = append(failureDataPoints, dp)

			if dp.ErrorType != "" {
				analysis.CommonErrorTypes[dp.ErrorType]++
			}
			if dp.ErrorMessage != "" {
				analysis.CommonErrorMessages[dp.ErrorMessage]++
			}
			if dp.InputType != "" {
				analysis.ProblematicInputTypes[dp.InputType]++
			}
			if dp.ProcessingStage != "" {
				analysis.ProcessingStageFailures[dp.ProcessingStage]++
			}
		}
	}

	// Calculate failure rate
	analysis.FailureRate = float64(analysis.TotalFailures) / float64(len(relevantDataPoints))

	// Determine failure trend
	analysis.FailureTrend = srm.calculateFailureTrend(relevantDataPoints)

	// Enhanced failure analysis - Root Cause Analysis
	analysis.RootCauseAnalysis = srm.performRootCauseAnalysis(failureDataPoints)

	// Enhanced failure analysis - Pattern Detection
	analysis.FailurePatterns = srm.detectFailurePatterns(failureDataPoints)

	// Enhanced failure analysis - Error Categorization
	analysis.ErrorCategorization = srm.categorizeErrors(failureDataPoints)

	// Enhanced failure analysis - Temporal Analysis
	analysis.TemporalAnalysis = srm.performTemporalAnalysis(failureDataPoints)

	// Enhanced failure analysis - Correlation Analysis
	analysis.CorrelationAnalysis = srm.performCorrelationAnalysis(failureDataPoints, relevantDataPoints)

	// Enhanced failure analysis - Impact Assessment
	analysis.ImpactAssessment = srm.assessFailureImpact(failureDataPoints, analysis.FailureRate)

	// Enhanced failure analysis - Actionable Insights
	analysis.ActionableInsights = srm.generateActionableInsights(analysis)

	// Enhanced failure analysis - Failure Timeline
	analysis.FailureTimeline = srm.buildFailureTimeline(failureDataPoints)

	// Enhanced failure analysis - Similarity Analysis
	analysis.SimilarityAnalysis = srm.performSimilarityAnalysis(failureDataPoints)

	// Enhanced failure analysis - Predictive Indicators
	analysis.PredictiveIndicators = srm.identifyPredictiveIndicators(failureDataPoints, relevantDataPoints)

	// Generate enhanced recommendations
	analysis.Recommendations = srm.generateFailureRecommendations(analysis)

	srm.logger.Info("Comprehensive failure analysis completed",
		zap.String("process", processName),
		zap.Int64("total_failures", analysis.TotalFailures),
		zap.Float64("failure_rate", analysis.FailureRate),
		zap.Int("failure_patterns", len(analysis.FailurePatterns)),
		zap.Int("actionable_insights", len(analysis.ActionableInsights)),
		zap.Int("predictive_indicators", len(analysis.PredictiveIndicators)))

	return analysis, nil
}

// AnalyzeTrends performs trend analysis for a process
func (srm *SuccessRateMonitor) AnalyzeTrends(ctx context.Context, processName string) (*TrendAnalysis, error) {
	srm.mu.RLock()
	metrics, exists := srm.metrics[processName]
	srm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("process %s not found", processName)
	}

	// Get data points within trend window
	endTime := time.Now()
	startTime := endTime.Add(-srm.config.TrendWindow)

	var relevantDataPoints []ProcessingDataPoint
	for _, dp := range metrics.DataPoints {
		if dp.Timestamp.After(startTime) && dp.Timestamp.Before(endTime) {
			relevantDataPoints = append(relevantDataPoints, dp)
		}
	}

	if len(relevantDataPoints) < srm.config.MinDataPoints {
		return nil, fmt.Errorf("insufficient data points for trend analysis: %d < %d", len(relevantDataPoints), srm.config.MinDataPoints)
	}

	analysis := &TrendAnalysis{
		ProcessName:    processName,
		AnalysisWindow: srm.config.TrendWindow,
		Timestamp:      time.Now(),
	}

	// Calculate trends
	analysis.SuccessRateTrend = srm.calculateSuccessRateTrend(relevantDataPoints)
	analysis.VolumeTrend = srm.calculateVolumeTrend(relevantDataPoints)
	analysis.ResponseTimeTrend = srm.calculateResponseTimeTrend(relevantDataPoints)
	analysis.TrendDirection = srm.determineTrendDirection(analysis.SuccessRateTrend)
	analysis.Confidence = srm.calculateTrendConfidence(relevantDataPoints)

	// Generate predictions
	analysis.Predictions = srm.generatePredictions(analysis, relevantDataPoints)

	return analysis, nil
}

// GetAlerts returns current alerts
func (srm *SuccessRateMonitor) GetAlerts() []SuccessAlert {
	srm.mu.RLock()
	defer srm.mu.RUnlock()

	alerts := make([]SuccessAlert, len(srm.alerts))
	copy(alerts, srm.alerts)
	return alerts
}

// ResolveAlert resolves an alert by ID
func (srm *SuccessRateMonitor) ResolveAlert(alertID string) error {
	srm.mu.Lock()
	defer srm.mu.Unlock()

	for i, alert := range srm.alerts {
		if alert.ID == alertID && !alert.Resolved {
			now := time.Now()
			srm.alerts[i].Resolved = true
			srm.alerts[i].ResolvedAt = &now

			srm.logger.Info("Alert resolved",
				zap.String("alert_id", alertID),
				zap.String("process", alert.ProcessName),
				zap.String("alert_type", string(alert.AlertType)))

			return nil
		}
	}

	return fmt.Errorf("alert %s not found or already resolved", alertID)
}

// GetSuccessRateReport generates a comprehensive success rate report
func (srm *SuccessRateMonitor) GetSuccessRateReport(ctx context.Context, processName string) (*SuccessRateReport, error) {
	srm.mu.RLock()
	defer srm.mu.RUnlock()

	// Get process metrics
	metrics, exists := srm.metrics[processName]
	if !exists {
		return nil, fmt.Errorf("no metrics found for process: %s", processName)
	}

	// Calculate overall metrics
	overallMetrics := srm.calculateOverallMetrics()

	// Generate trends
	trends := srm.calculateTrends(processName)

	// Get active alerts
	alerts := srm.getActiveAlerts(processName)

	// Generate recommendations
	recommendations := srm.generateOverallRecommendations(metrics, overallMetrics)

	report := &SuccessRateReport{
		ProcessName:     processName,
		GeneratedAt:     time.Now(),
		AnalysisWindow:  srm.config.AnalysisWindow,
		OverallMetrics:  overallMetrics,
		ProcessMetrics:  []ProcessMetrics{*metrics},
		Trends:          trends,
		Alerts:          alerts,
		Recommendations: recommendations,
	}

	srm.logger.Info("Generated success rate report",
		zap.String("process_name", processName),
		zap.Float64("success_rate", metrics.SuccessRate),
		zap.Int("recommendations_count", len(recommendations)))

	return report, nil
}

// Helper methods

func (srm *SuccessRateMonitor) getOrCreateProcessMetrics(processName string) *ProcessMetrics {
	metrics, exists := srm.metrics[processName]
	if !exists {
		metrics = &ProcessMetrics{
			ProcessName:     processName,
			FailurePatterns: make(map[string]int),
			DataPoints:      make([]ProcessingDataPoint, 0),
		}
		srm.metrics[processName] = metrics
	}
	return metrics
}

func (srm *SuccessRateMonitor) cleanupOldDataPoints(metrics *ProcessMetrics) {
	cutoffTime := time.Now().Add(-srm.config.MetricsRetentionPeriod)

	var validDataPoints []ProcessingDataPoint
	for _, dp := range metrics.DataPoints {
		if dp.Timestamp.After(cutoffTime) {
			validDataPoints = append(validDataPoints, dp)
		}
	}

	// Limit data points if exceeding max
	if len(validDataPoints) > srm.config.MaxDataPoints {
		validDataPoints = validDataPoints[len(validDataPoints)-srm.config.MaxDataPoints:]
	}

	metrics.DataPoints = validDataPoints
}

func (srm *SuccessRateMonitor) calculateSuccessTrend(metrics *ProcessMetrics) TrendDirection {
	if len(metrics.DataPoints) < 10 {
		return TrendInsufficient
	}

	// Simple trend calculation based on recent vs older data points
	midPoint := len(metrics.DataPoints) / 2
	recentDataPoints := metrics.DataPoints[midPoint:]
	olderDataPoints := metrics.DataPoints[:midPoint]

	recentSuccessRate := srm.calculateSuccessRateFromDataPoints(recentDataPoints)
	olderSuccessRate := srm.calculateSuccessRateFromDataPoints(olderDataPoints)

	threshold := 0.02 // 2% change threshold

	if recentSuccessRate > olderSuccessRate+threshold {
		return TrendImproving
	} else if recentSuccessRate < olderSuccessRate-threshold {
		return TrendDegrading
	}
	return TrendStable
}

func (srm *SuccessRateMonitor) calculateSuccessRateFromDataPoints(dataPoints []ProcessingDataPoint) float64 {
	if len(dataPoints) == 0 {
		return 0
	}

	successes := 0
	for _, dp := range dataPoints {
		if dp.Success {
			successes++
		}
	}

	return float64(successes) / float64(len(dataPoints))
}

func (srm *SuccessRateMonitor) checkAlerts(ctx context.Context, metrics *ProcessMetrics) {
	// Check if enough time has passed since last alert
	if metrics.LastAlertTime != nil {
		if time.Since(*metrics.LastAlertTime) < srm.config.AlertCooldownPeriod {
			return
		}
	}

	var alertType AlertType
	var message string

	if metrics.SuccessRate < srm.config.CriticalThreshold {
		alertType = AlertTypeCritical
		message = fmt.Sprintf("Critical: Success rate %.2f%% is below critical threshold %.2f%%",
			metrics.SuccessRate*100, srm.config.CriticalThreshold*100)
	} else if metrics.SuccessRate < srm.config.WarningThreshold {
		alertType = AlertTypeWarning
		message = fmt.Sprintf("Warning: Success rate %.2f%% is below warning threshold %.2f%%",
			metrics.SuccessRate*100, srm.config.WarningThreshold*100)
	} else {
		return // No alert needed
	}

	// Create alert
	alert := SuccessAlert{
		ID:          fmt.Sprintf("alert_%s_%d", metrics.ProcessName, time.Now().Unix()),
		ProcessName: metrics.ProcessName,
		AlertType:   alertType,
		Message:     message,
		CurrentRate: metrics.SuccessRate,
		TargetRate:  srm.config.TargetSuccessRate,
		Timestamp:   time.Now(),
		Resolved:    false,
	}

	srm.alerts = append(srm.alerts, alert)
	now := time.Now()
	metrics.LastAlertTime = &now

	srm.logger.Warn("Success rate alert created",
		zap.String("alert_id", alert.ID),
		zap.String("process", metrics.ProcessName),
		zap.String("alert_type", string(alertType)),
		zap.Float64("current_rate", metrics.SuccessRate),
		zap.Float64("target_rate", srm.config.TargetSuccessRate))
}

func (srm *SuccessRateMonitor) startBackgroundMonitoring() {
	ticker := time.NewTicker(srm.config.AnalysisWindow)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx := context.Background()

			// Perform failure analysis for all processes
			if srm.config.EnableFailureAnalysis {
				srm.mu.RLock()
				processNames := make([]string, 0, len(srm.metrics))
				for processName := range srm.metrics {
					processNames = append(processNames, processName)
				}
				srm.mu.RUnlock()

				for _, processName := range processNames {
					analysis, err := srm.AnalyzeFailures(ctx, processName)
					if err != nil {
						srm.logger.Debug("Background failure analysis skipped",
							zap.String("process", processName),
							zap.Error(err))
					} else {
						srm.logger.Debug("Background failure analysis completed",
							zap.String("process", processName),
							zap.Int64("total_failures", analysis.TotalFailures),
							zap.Float64("failure_rate", analysis.FailureRate))
					}
				}
			}

			// Perform trend analysis for all processes
			if srm.config.EnableTrendAnalysis {
				srm.mu.RLock()
				processNames := make([]string, 0, len(srm.metrics))
				for processName := range srm.metrics {
					processNames = append(processNames, processName)
				}
				srm.mu.RUnlock()

				for _, processName := range processNames {
					trends, err := srm.AnalyzeTrends(ctx, processName)
					if err != nil {
						srm.logger.Debug("Background trend analysis skipped",
							zap.String("process", processName),
							zap.Error(err))
					} else {
						srm.logger.Debug("Background trend analysis completed",
							zap.String("process", processName),
							zap.Float64("success_rate_trend", trends.SuccessRateTrend),
							zap.String("trend_direction", string(trends.TrendDirection)))
					}
				}
			}
		}
	}
}

// Additional helper methods for trend analysis
func (srm *SuccessRateMonitor) calculateFailureTrend(dataPoints []ProcessingDataPoint) TrendDirection {
	if len(dataPoints) < 10 {
		return TrendInsufficient
	}

	midPoint := len(dataPoints) / 2
	recentDataPoints := dataPoints[midPoint:]
	olderDataPoints := dataPoints[:midPoint]

	recentFailureRate := srm.calculateFailureRateFromDataPoints(recentDataPoints)
	olderFailureRate := srm.calculateFailureRateFromDataPoints(olderDataPoints)

	threshold := 0.02 // 2% change threshold

	if recentFailureRate < olderFailureRate-threshold {
		return TrendImproving
	} else if recentFailureRate > olderFailureRate+threshold {
		return TrendDegrading
	}
	return TrendStable
}

func (srm *SuccessRateMonitor) calculateFailureRateFromDataPoints(dataPoints []ProcessingDataPoint) float64 {
	if len(dataPoints) == 0 {
		return 0
	}

	failures := 0
	for _, dp := range dataPoints {
		if !dp.Success {
			failures++
		}
	}

	return float64(failures) / float64(len(dataPoints))
}

func (srm *SuccessRateMonitor) calculateSuccessRateTrend(dataPoints []ProcessingDataPoint) float64 {
	if len(dataPoints) < 2 {
		return 0
	}

	// Simple linear trend calculation
	recentRate := srm.calculateSuccessRateFromDataPoints(dataPoints[len(dataPoints)/2:])
	olderRate := srm.calculateSuccessRateFromDataPoints(dataPoints[:len(dataPoints)/2])

	return recentRate - olderRate
}

func (srm *SuccessRateMonitor) calculateVolumeTrend(dataPoints []ProcessingDataPoint) float64 {
	if len(dataPoints) < 2 {
		return 0
	}

	// Calculate volume trend (requests per time period)
	recentVolume := len(dataPoints[len(dataPoints)/2:])
	olderVolume := len(dataPoints[:len(dataPoints)/2])

	if olderVolume == 0 {
		return 0
	}

	return float64(recentVolume-olderVolume) / float64(olderVolume)
}

func (srm *SuccessRateMonitor) calculateResponseTimeTrend(dataPoints []ProcessingDataPoint) float64 {
	if len(dataPoints) < 2 {
		return 0
	}

	// Calculate average response time trend
	recentDataPoints := dataPoints[len(dataPoints)/2:]
	olderDataPoints := dataPoints[:len(dataPoints)/2]

	recentAvg := srm.calculateAverageResponseTime(recentDataPoints)
	olderAvg := srm.calculateAverageResponseTime(olderDataPoints)

	if olderAvg == 0 {
		return 0
	}

	return float64(recentAvg-olderAvg) / float64(olderAvg)
}

func (srm *SuccessRateMonitor) calculateAverageResponseTime(dataPoints []ProcessingDataPoint) time.Duration {
	if len(dataPoints) == 0 {
		return 0
	}

	totalDuration := time.Duration(0)
	for _, dp := range dataPoints {
		totalDuration += dp.ResponseTime
	}

	return totalDuration / time.Duration(len(dataPoints))
}

func (srm *SuccessRateMonitor) determineTrendDirection(trend float64) TrendDirection {
	threshold := 0.01 // 1% threshold

	if trend > threshold {
		return TrendImproving
	} else if trend < -threshold {
		return TrendDegrading
	}
	return TrendStable
}

func (srm *SuccessRateMonitor) calculateTrendConfidence(dataPoints []ProcessingDataPoint) float64 {
	if len(dataPoints) < 10 {
		return 0
	}

	// Simple confidence calculation based on data point consistency
	successRates := make([]float64, 0)
	windowSize := len(dataPoints) / 5

	for i := 0; i < len(dataPoints)-windowSize; i += windowSize {
		window := dataPoints[i : i+windowSize]
		rate := srm.calculateSuccessRateFromDataPoints(window)
		successRates = append(successRates, rate)
	}

	// Calculate variance
	if len(successRates) < 2 {
		return 0
	}

	mean := 0.0
	for _, rate := range successRates {
		mean += rate
	}
	mean /= float64(len(successRates))

	variance := 0.0
	for _, rate := range successRates {
		variance += (rate - mean) * (rate - mean)
	}
	variance /= float64(len(successRates))

	// Lower variance = higher confidence
	confidence := 1.0 - variance
	if confidence < 0 {
		confidence = 0
	}
	if confidence > 1 {
		confidence = 1
	}

	return confidence
}

func (srm *SuccessRateMonitor) generatePredictions(analysis *TrendAnalysis, dataPoints []ProcessingDataPoint) []Prediction {
	predictions := make([]Prediction, 0)

	// Generate predictions for different timeframes
	timeframes := []time.Duration{
		1 * time.Hour,
		6 * time.Hour,
		24 * time.Hour,
		7 * 24 * time.Hour,
	}

	currentRate := srm.calculateSuccessRateFromDataPoints(dataPoints)

	for _, timeframe := range timeframes {
		// Simple linear prediction
		predictedRate := currentRate + analysis.SuccessRateTrend*float64(timeframe.Hours())/24.0

		// Ensure prediction is within reasonable bounds
		if predictedRate < 0 {
			predictedRate = 0
		}
		if predictedRate > 1 {
			predictedRate = 1
		}

		// Calculate confidence based on trend confidence and data quality
		confidence := analysis.Confidence * 0.8 // Reduce confidence for predictions

		prediction := Prediction{
			Timeframe:   timeframe,
			SuccessRate: predictedRate,
			Confidence:  confidence,
			Factors:     []string{"historical_trend", "current_performance"},
		}

		predictions = append(predictions, prediction)
	}

	return predictions
}

func (srm *SuccessRateMonitor) generateFailureRecommendations(analysis *FailureAnalysis) []FailureRecommendation {
	recommendations := make([]FailureRecommendation, 0)

	// Analyze common error types
	for errorType, count := range analysis.CommonErrorTypes {
		if count > 5 { // Only recommend for frequently occurring errors
			recommendation := FailureRecommendation{
				Type:                 "error_type_optimization",
				Description:          fmt.Sprintf("Address frequent %s errors", errorType),
				Impact:               "high",
				Effort:               "medium",
				Priority:             1,
				EstimatedImprovement: 0.05, // 5% improvement
			}
			recommendations = append(recommendations, recommendation)
		}
	}

	// Analyze problematic input types
	for inputType, count := range analysis.ProblematicInputTypes {
		if count > 3 {
			recommendation := FailureRecommendation{
				Type:                 "input_validation",
				Description:          fmt.Sprintf("Improve validation for %s input type", inputType),
				Impact:               "medium",
				Effort:               "low",
				Priority:             2,
				EstimatedImprovement: 0.03, // 3% improvement
			}
			recommendations = append(recommendations, recommendation)
		}
	}

	// Analyze processing stage failures
	for stage, count := range analysis.ProcessingStageFailures {
		if count > 2 {
			recommendation := FailureRecommendation{
				Type:                 "processing_optimization",
				Description:          fmt.Sprintf("Optimize %s processing stage", stage),
				Impact:               "high",
				Effort:               "high",
				Priority:             1,
				EstimatedImprovement: 0.08, // 8% improvement
			}
			recommendations = append(recommendations, recommendation)
		}
	}

	return recommendations
}

// calculateImpactLevel calculates impact level
func (srm *SuccessRateMonitor) calculateImpactLevel(dp ProcessingDataPoint) string {
	if dp.ResponseTime > 30*time.Second {
		return "high"
	} else if dp.ResponseTime > 10*time.Second {
		return "medium"
	}
	return "low"
}

// calculateSeverity calculates severity based on percentage
func (srm *SuccessRateMonitor) calculateSeverity(percentage float64) string {
	if percentage > 0.5 {
		return "critical"
	} else if percentage > 0.2 {
		return "high"
	} else if percentage > 0.1 {
		return "medium"
	}
	return "low"
}

// performRootCauseAnalysis analyzes failures to identify root causes
func (srm *SuccessRateMonitor) performRootCauseAnalysis(failureDataPoints []ProcessingDataPoint) *RootCauseAnalysis {
	if len(failureDataPoints) == 0 {
		return &RootCauseAnalysis{
			PrimaryRootCause:     RootCause{},
			SecondaryRootCauses:  []RootCause{},
			CauseConfidence:      0.0,
			ContributingFactors:  []Contributing{},
			SystemicIssues:       []SystemicIssue{},
			EnvironmentalFactors: []Environmental{},
			HumanFactors:         []HumanFactor{},
			TechnicalFactors:     []TechnicalFactor{},
		}
	}

	rootCauses := make(map[string]*RootCause)

	for _, dp := range failureDataPoints {
		causeID := fmt.Sprintf("%s_%s", dp.ErrorType, dp.ProcessingStage)
		if cause, exists := rootCauses[causeID]; exists {
			cause.Frequency++
			cause.LastSeen = dp.Timestamp
		} else {
			rootCauses[causeID] = &RootCause{
				ID:          causeID,
				Category:    srm.categorizeRootCause(dp.ErrorType, dp.ProcessingStage),
				Subcategory: dp.ProcessingStage,
				Description: fmt.Sprintf("Failures in %s stage with error type: %s", dp.ProcessingStage, dp.ErrorType),
				Evidence:    []string{dp.ErrorMessage},
				Confidence:  0.0,
				Impact:      srm.calculateImpactLevel(dp),
				Frequency:   1,
				FirstSeen:   dp.Timestamp,
				LastSeen:    dp.Timestamp,
			}
		}
	}

	// Find primary root cause (most frequent)
	var primaryCause *RootCause
	maxFrequency := 0

	for _, cause := range rootCauses {
		if cause.Frequency > maxFrequency {
			maxFrequency = cause.Frequency
			primaryCause = cause
		}
		cause.Confidence = float64(cause.Frequency) / float64(len(failureDataPoints))
	}

	// Create secondary root causes
	secondaryCauses := make([]RootCause, 0)
	for _, cause := range rootCauses {
		if cause != primaryCause && cause.Frequency > 1 {
			secondaryCauses = append(secondaryCauses, *cause)
		}
	}

	confidence := 0.5
	if primaryCause != nil {
		confidence = primaryCause.Confidence
	}

	return &RootCauseAnalysis{
		PrimaryRootCause:     *primaryCause,
		SecondaryRootCauses:  secondaryCauses,
		CauseConfidence:      confidence,
		ContributingFactors:  srm.identifyContributingFactors(failureDataPoints),
		SystemicIssues:       srm.identifySystemicIssues(failureDataPoints),
		EnvironmentalFactors: srm.identifyEnvironmentalFactors(failureDataPoints),
		HumanFactors:         srm.identifyHumanFactors(failureDataPoints),
		TechnicalFactors:     srm.identifyTechnicalFactors(failureDataPoints),
	}
}

// categorizeRootCause categorizes root causes
func (srm *SuccessRateMonitor) categorizeRootCause(errorType, processingStage string) string {
	switch errorType {
	case "validation_error", "input_error":
		return "process"
	case "timeout_error", "resource_error":
		return "technical"
	case "configuration_error":
		return "human"
	case "network_error", "integration_error":
		return "environmental"
	default:
		return "technical"
	}
}

// identifyContributingFactors identifies contributing factors
func (srm *SuccessRateMonitor) identifyContributingFactors(failureDataPoints []ProcessingDataPoint) []Contributing {
	factors := []Contributing{
		{
			Factor:      "High Response Time",
			Weight:      0.7,
			Description: "Failures often correlated with high response times",
			Category:    "performance",
		},
	}

	avgResponseTime := srm.calculateAverageResponseTime(failureDataPoints)
	if avgResponseTime > 15*time.Second {
		factors[0].Weight = 0.9
	}

	return factors
}

// identifySystemicIssues identifies systemic issues
func (srm *SuccessRateMonitor) identifySystemicIssues(failureDataPoints []ProcessingDataPoint) []SystemicIssue {
	issues := []SystemicIssue{}

	stageFailures := make(map[string]int)
	for _, dp := range failureDataPoints {
		if dp.ProcessingStage != "" {
			stageFailures[dp.ProcessingStage]++
		}
	}

	for stage, count := range stageFailures {
		if float64(count)/float64(len(failureDataPoints)) > 0.3 {
			issues = append(issues, SystemicIssue{
				Issue:         fmt.Sprintf("High failure rate in %s stage", stage),
				Scope:         "service",
				Severity:      "high",
				AffectedAreas: []string{stage},
				FirstDetected: time.Now().Add(-srm.config.AnalysisWindow),
			})
		}
	}

	return issues
}

// identifyEnvironmentalFactors identifies environmental factors
func (srm *SuccessRateMonitor) identifyEnvironmentalFactors(failureDataPoints []ProcessingDataPoint) []Environmental {
	factors := []Environmental{}

	hourCounts := make(map[int]int)
	for _, dp := range failureDataPoints {
		hour := dp.Timestamp.Hour()
		hourCounts[hour]++
	}

	for hour, count := range hourCounts {
		if float64(count)/float64(len(failureDataPoints)) > 0.2 {
			factors = append(factors, Environmental{
				Factor:      fmt.Sprintf("Peak failures at hour %d", hour),
				Description: "High failure rate during specific hours",
				Impact:      "medium",
				Duration:    time.Hour,
				Frequency:   "daily",
			})
		}
	}

	return factors
}

// identifyHumanFactors identifies human factors
func (srm *SuccessRateMonitor) identifyHumanFactors(failureDataPoints []ProcessingDataPoint) []HumanFactor {
	factors := []HumanFactor{}

	configErrors := 0
	for _, dp := range failureDataPoints {
		if dp.ErrorType == "configuration_error" {
			configErrors++
		}
	}

	if configErrors > 0 {
		factors = append(factors, HumanFactor{
			Factor:      "Configuration Errors",
			Category:    "configuration",
			Description: "Failures due to incorrect configuration",
			Prevention:  "Implement configuration validation and automated testing",
			Training:    true,
		})
	}

	return factors
}

// identifyTechnicalFactors identifies technical factors
func (srm *SuccessRateMonitor) identifyTechnicalFactors(failureDataPoints []ProcessingDataPoint) []TechnicalFactor {
	factors := []TechnicalFactor{}

	resourceErrors := 0
	timeoutErrors := 0

	for _, dp := range failureDataPoints {
		if dp.ErrorType == "resource_error" {
			resourceErrors++
		}
		if dp.ErrorType == "timeout_error" {
			timeoutErrors++
		}
	}

	if resourceErrors > 0 {
		factors = append(factors, TechnicalFactor{
			Component:  "Resource Management",
			Issue:      "Resource exhaustion causing failures",
			Severity:   "high",
			Resolution: "Implement better resource allocation and monitoring",
			Prevention: "Add resource usage alerts and auto-scaling",
			TechDebt:   true,
		})
	}

	if timeoutErrors > 0 {
		factors = append(factors, TechnicalFactor{
			Component:  "Timeout Handling",
			Issue:      "Timeouts causing failures",
			Severity:   "medium",
			Resolution: "Optimize processing times and adjust timeout values",
			Prevention: "Implement better timeout handling and retry logic",
			TechDebt:   false,
		})
	}

	return factors
}

// detectFailurePatterns detects patterns in failures
func (srm *SuccessRateMonitor) detectFailurePatterns(failureDataPoints []ProcessingDataPoint) []FailurePattern {
	patterns := []FailurePattern{}

	errorTypeCount := make(map[string]int)
	for _, dp := range failureDataPoints {
		if dp.ErrorType != "" {
			errorTypeCount[dp.ErrorType]++
		}
	}

	for errorType, count := range errorTypeCount {
		if count >= 3 {
			confidence := float64(count) / float64(len(failureDataPoints))
			patterns = append(patterns, FailurePattern{
				ID:              fmt.Sprintf("pattern_%s", errorType),
				Name:            fmt.Sprintf("%s Pattern", errorType),
				Description:     fmt.Sprintf("Recurring pattern of %s errors", errorType),
				Pattern:         errorType,
				Frequency:       count,
				Confidence:      confidence,
				Severity:        srm.calculateSeverity(confidence),
				Category:        "error_type",
				Characteristics: map[string]interface{}{"error_type": errorType},
				Examples:        []string{errorType},
				FirstDetected:   time.Now().Add(-srm.config.AnalysisWindow),
				LastDetected:    time.Now(),
				Trend:           TrendDegrading,
				RelatedPatterns: []string{},
			})
		}
	}

	return patterns
}

// categorizeErrors categorizes errors by type
func (srm *SuccessRateMonitor) categorizeErrors(failureDataPoints []ProcessingDataPoint) ErrorCategorization {
	categorization := ErrorCategorization{}

	total := len(failureDataPoints)
	if total == 0 {
		return categorization
	}

	counts := make(map[string]int)

	for _, dp := range failureDataPoints {
		category := srm.categorizeError(dp.ErrorType)
		counts[category]++
	}

	for category, count := range counts {
		percentage := float64(count) / float64(total)
		errorCat := ErrorCategory{
			Count:          count,
			Percentage:     percentage,
			TrendDirection: TrendStable,
			Severity:       srm.calculateSeverity(percentage),
		}

		switch category {
		case "validation":
			categorization.ValidationErrors = errorCat
		case "timeout":
			categorization.TimeoutErrors = errorCat
		case "resource":
			categorization.ResourceErrors = errorCat
		case "system":
			categorization.SystemErrors = errorCat
		}
	}

	return categorization
}

// categorizeError categorizes an error type
func (srm *SuccessRateMonitor) categorizeError(errorType string) string {
	switch errorType {
	case "validation_error", "input_error":
		return "validation"
	case "timeout_error":
		return "timeout"
	case "resource_error", "memory_error":
		return "resource"
	default:
		return "system"
	}
}

// performTemporalAnalysis performs temporal analysis of failures
func (srm *SuccessRateMonitor) performTemporalAnalysis(failureDataPoints []ProcessingDataPoint) TemporalAnalysis {
	analysis := TemporalAnalysis{
		HourlyDistribution: make(map[int]int),
		DailyDistribution:  make(map[string]int),
		WeeklyTrends:       make(map[string]float64),
		MonthlyTrends:      make(map[string]float64),
		SeasonalPatterns:   []SeasonalPattern{},
		PeakFailureTimes:   []PeakTime{},
		FailureFrequency:   make(map[string]int),
		Duration:           TemporalDuration{},
		Cyclical:           []CyclicalPattern{},
	}

	if len(failureDataPoints) == 0 {
		return analysis
	}

	for _, dp := range failureDataPoints {
		hour := dp.Timestamp.Hour()
		analysis.HourlyDistribution[hour]++

		day := dp.Timestamp.Weekday().String()
		analysis.DailyDistribution[day]++
	}

	for hour, count := range analysis.HourlyDistribution {
		if count > len(failureDataPoints)/24*2 {
			analysis.PeakFailureTimes = append(analysis.PeakFailureTimes, PeakTime{
				TimeRange:   fmt.Sprintf("%02d:00-%02d:00", hour, hour+1),
				FailureRate: float64(count) / float64(len(failureDataPoints)),
				Volume:      count,
				Reason:      "High failure concentration",
			})
		}
	}

	return analysis
}

// performCorrelationAnalysis performs correlation analysis
func (srm *SuccessRateMonitor) performCorrelationAnalysis(failureDataPoints, allDataPoints []ProcessingDataPoint) CorrelationAnalysis {
	return CorrelationAnalysis{
		LoadCorrelation:         srm.calculateLoadCorrelation(failureDataPoints, allDataPoints),
		TimeCorrelation:         make(map[string]float64),
		InputSizeCorrelation:    srm.calculateInputSizeCorrelation(failureDataPoints),
		ProcessStageCorrelation: make(map[string]float64),
		ErrorTypeCorrelation:    make(map[string]float64),
		ResponseTimeCorrelation: srm.calculateResponseTimeCorrelation(failureDataPoints),
		ConcurrencyCorrelation:  0.5,
		ExternalFactors:         []ExternalFactor{},
		DependencyCorrelation:   make(map[string]float64),
		ResourceCorrelation: ResourceCorrelation{
			CPUCorrelation:     0.6,
			MemoryCorrelation:  0.7,
			DiskCorrelation:    0.3,
			NetworkCorrelation: 0.8,
		},
		EnvironmentalCorrelation: EnvironmentalCorrelation{
			TimeOfDayCorrelation: 0.4,
			DayOfWeekCorrelation: 0.2,
			LoadCorrelation:      0.8,
			WeatherCorrelation:   0.0,
		},
	}
}

// calculateLoadCorrelation calculates load correlation
func (srm *SuccessRateMonitor) calculateLoadCorrelation(failureDataPoints, allDataPoints []ProcessingDataPoint) float64 {
	if len(allDataPoints) == 0 {
		return 0.0
	}

	failureRate := float64(len(failureDataPoints)) / float64(len(allDataPoints))

	if failureRate > 0.2 {
		return 0.8
	} else if failureRate > 0.1 {
		return 0.5
	}
	return 0.2
}

// calculateInputSizeCorrelation calculates input size correlation
func (srm *SuccessRateMonitor) calculateInputSizeCorrelation(failureDataPoints []ProcessingDataPoint) float64 {
	if len(failureDataPoints) == 0 {
		return 0.0
	}

	avgInputSize := 0
	for _, dp := range failureDataPoints {
		avgInputSize += dp.InputSize
	}

	if len(failureDataPoints) > 0 {
		avgInputSize /= len(failureDataPoints)
	}

	if avgInputSize > 10000 {
		return 0.7
	} else if avgInputSize > 5000 {
		return 0.5
	}
	return 0.3
}

// calculateResponseTimeCorrelation calculates response time correlation
func (srm *SuccessRateMonitor) calculateResponseTimeCorrelation(failureDataPoints []ProcessingDataPoint) float64 {
	if len(failureDataPoints) == 0 {
		return 0.0
	}

	avgResponseTime := srm.calculateAverageResponseTime(failureDataPoints)

	if avgResponseTime > 30*time.Second {
		return 0.9
	} else if avgResponseTime > 10*time.Second {
		return 0.7
	}
	return 0.4
}

// assessFailureImpact assesses the impact of failures
func (srm *SuccessRateMonitor) assessFailureImpact(failureDataPoints []ProcessingDataPoint, failureRate float64) ImpactAssessment {
	return ImpactAssessment{
		BusinessImpact: BusinessImpact{
			Severity:             srm.calculateSeverity(failureRate),
			AffectedProcesses:    []string{"business_processing"},
			CustomerSatisfaction: 1.0 - failureRate,
			RevenueLoss:          failureRate * 1000.0,
		},
		TechnicalImpact: TechnicalImpact{
			SystemAvailability:     1.0 - failureRate,
			PerformanceDegradation: failureRate * 0.8,
			DataIntegrity:          "good",
			SystemStability:        "stable",
		},
		UserImpact: UserImpact{
			AffectedUsers:    int(float64(100) * failureRate),
			UserExperience:   "degraded",
			UserSatisfaction: 1.0 - failureRate,
		},
		FinancialImpact: FinancialImpact{
			DirectCosts:        failureRate * 500.0,
			TotalEstimatedCost: failureRate * 1100.0,
			CostPerIncident:    50.0,
		},
		ReputationalImpact: ReputationalImpact{
			BrandImpact:   srm.calculateImpactLevel(ProcessingDataPoint{}),
			CustomerTrust: 1.0 - failureRate*0.5,
		},
		OperationalImpact: OperationalImpact{
			StaffProductivity: 1.0 - failureRate*0.3,
			ProcessEfficiency: 1.0 - failureRate*0.5,
		},
		SecurityImpact: SecurityImpact{
			SecurityLevel: "normal",
			DataExposure:  "none",
		},
		ComplianceImpact: ComplianceImpact{
			ComplianceRisk: "low",
			PenaltyRisk:    0.0,
		},
	}
}

// generateActionableInsights generates actionable insights
func (srm *SuccessRateMonitor) generateActionableInsights(analysis *FailureAnalysis) []ActionableInsight {
	insights := []ActionableInsight{}

	for _, pattern := range analysis.FailurePatterns {
		if pattern.Confidence > 0.3 {
			insights = append(insights, ActionableInsight{
				ID:          fmt.Sprintf("insight_%s", pattern.ID),
				Title:       fmt.Sprintf("Address %s Pattern", pattern.Name),
				Description: pattern.Description,
				Category:    "immediate",
				Priority:    srm.calculatePriority(pattern.Severity),
				Urgency:     pattern.Severity,
				Impact:      "medium",
				Confidence:  pattern.Confidence,
				CreatedAt:   time.Now(),
			})
		}
	}

	return insights
}

// calculatePriority calculates priority based on severity
func (srm *SuccessRateMonitor) calculatePriority(severity string) int {
	switch severity {
	case "critical":
		return 1
	case "high":
		return 2
	case "medium":
		return 3
	default:
		return 4
	}
}

// buildFailureTimeline builds a timeline of failure events
func (srm *SuccessRateMonitor) buildFailureTimeline(failureDataPoints []ProcessingDataPoint) []FailureEvent {
	timeline := []FailureEvent{}

	for _, dp := range failureDataPoints {
		timeline = append(timeline, FailureEvent{
			Timestamp:       dp.Timestamp,
			EventType:       "failure",
			Severity:        srm.calculateImpactLevel(dp),
			Description:     fmt.Sprintf("Failure in %s stage", dp.ProcessingStage),
			ProcessName:     dp.ProcessName,
			ProcessingStage: dp.ProcessingStage,
			ErrorType:       dp.ErrorType,
			ErrorMessage:    dp.ErrorMessage,
			Impact:          srm.calculateImpactLevel(dp),
			Duration:        dp.ResponseTime,
			Resolution:      "pending",
			Metadata: map[string]interface{}{
				"input_size":       dp.InputSize,
				"output_size":      dp.OutputSize,
				"confidence_score": dp.ConfidenceScore,
			},
		})
	}

	return timeline
}

// performSimilarityAnalysis performs similarity analysis between failures
func (srm *SuccessRateMonitor) performSimilarityAnalysis(failureDataPoints []ProcessingDataPoint) SimilarityAnalysis {
	return SimilarityAnalysis{
		SimilarFailures:   []SimilarFailure{},
		FailureClusters:   []FailureCluster{},
		AnomalousFailures: []AnomalousFailure{},
		PatternMatches:    []PatternMatch{},
	}
}

// identifyPredictiveIndicators identifies predictive indicators
func (srm *SuccessRateMonitor) identifyPredictiveIndicators(failureDataPoints, allDataPoints []ProcessingDataPoint) []PredictiveIndicator {
	indicators := []PredictiveIndicator{}

	avgResponseTime := srm.calculateAverageResponseTime(failureDataPoints)
	thresholdResponseTime := 15 * time.Second

	if avgResponseTime > thresholdResponseTime {
		indicators = append(indicators, PredictiveIndicator{
			IndicatorType:   "early_warning",
			Name:            "High Response Time",
			Description:     "Response times are increasing, indicating potential future failures",
			CurrentValue:    avgResponseTime.Seconds(),
			ThresholdValue:  thresholdResponseTime.Seconds(),
			Confidence:      0.8,
			Trend:           TrendDegrading,
			TimeToImpact:    2 * time.Hour,
			Severity:        "medium",
			Recommendations: []string{"Optimize processing"},
			LastUpdated:     time.Now(),
		})
	}

	return indicators
}

// calculateOverallMetrics calculates overall metrics across all processes
func (srm *SuccessRateMonitor) calculateOverallMetrics() OverallMetrics {
	var totalAttempts, totalSuccesses, totalFailures int64
	var totalSuccessRate float64
	var processCount int
	var bestProcess, worstProcess string
	var bestRate, worstRate float64

	for processName, metrics := range srm.metrics {
		totalAttempts += metrics.TotalAttempts
		totalSuccesses += metrics.SuccessfulAttempts
		totalFailures += metrics.FailedAttempts
		totalSuccessRate += metrics.SuccessRate
		processCount++

		if bestProcess == "" || metrics.SuccessRate > bestRate {
			bestProcess = processName
			bestRate = metrics.SuccessRate
		}

		if worstProcess == "" || metrics.SuccessRate < worstRate {
			worstProcess = processName
			worstRate = metrics.SuccessRate
		}
	}

	averageSuccessRate := 0.0
	if processCount > 0 {
		averageSuccessRate = totalSuccessRate / float64(processCount)
	}

	return OverallMetrics{
		TotalProcesses:     processCount,
		AverageSuccessRate: averageSuccessRate,
		BestProcess:        bestProcess,
		WorstProcess:       worstProcess,
		TotalAttempts:      totalAttempts,
		TotalSuccesses:     totalSuccesses,
		TotalFailures:      totalFailures,
		OverallTrend:       srm.calculateOverallTrend(),
	}
}

// calculateOverallTrend calculates the overall trend across all processes
func (srm *SuccessRateMonitor) calculateOverallTrend() TrendDirection {
	if len(srm.metrics) == 0 {
		return TrendInsufficient
	}

	improvingCount := 0
	degradingCount := 0

	for _, metrics := range srm.metrics {
		switch metrics.SuccessTrend {
		case TrendImproving:
			improvingCount++
		case TrendDegrading:
			degradingCount++
		}
	}

	if improvingCount > degradingCount {
		return TrendImproving
	} else if degradingCount > improvingCount {
		return TrendDegrading
	}
	return TrendStable
}

// calculateTrends calculates trends for a specific process
func (srm *SuccessRateMonitor) calculateTrends(processName string) []TrendAnalysis {
	metrics, exists := srm.metrics[processName]
	if !exists {
		return []TrendAnalysis{}
	}

	if len(metrics.DataPoints) < 2 {
		return []TrendAnalysis{}
	}

	// Calculate trend over the last 24 hours
	trendWindow := 24 * time.Hour
	cutoffTime := time.Now().Add(-trendWindow)

	var recentDataPoints []ProcessingDataPoint
	for _, dp := range metrics.DataPoints {
		if dp.Timestamp.After(cutoffTime) {
			recentDataPoints = append(recentDataPoints, dp)
		}
	}

	if len(recentDataPoints) < 2 {
		return []TrendAnalysis{}
	}

	// Split data into two halves for comparison
	midPoint := len(recentDataPoints) / 2
	firstHalf := recentDataPoints[:midPoint]
	secondHalf := recentDataPoints[midPoint:]

	startRate := srm.calculateSuccessRate(firstHalf)
	endRate := srm.calculateSuccessRate(secondHalf)
	change := endRate - startRate

	trend := TrendStable
	if change > 0.05 {
		trend = TrendImproving
	} else if change < -0.05 {
		trend = TrendDegrading
	}

	return []TrendAnalysis{
		{
			ProcessName:       processName,
			AnalysisWindow:    trendWindow,
			SuccessRateTrend:  change,
			VolumeTrend:       0.0, // Not calculated in this implementation
			ResponseTimeTrend: 0.0, // Not calculated in this implementation
			TrendDirection:    trend,
			Confidence:        0.8,
			Predictions:       []Prediction{},
			Timestamp:         time.Now(),
		},
	}
}

// calculateSuccessRate calculates success rate from data points
func (srm *SuccessRateMonitor) calculateSuccessRate(dataPoints []ProcessingDataPoint) float64 {
	if len(dataPoints) == 0 {
		return 0.0
	}

	successCount := 0
	for _, dp := range dataPoints {
		if dp.Success {
			successCount++
		}
	}

	return float64(successCount) / float64(len(dataPoints))
}

// getActiveAlerts gets active alerts for a process
func (srm *SuccessRateMonitor) getActiveAlerts(processName string) []SuccessAlert {
	var activeAlerts []SuccessAlert
	for _, alert := range srm.alerts {
		if alert.ProcessName == processName && !alert.Resolved {
			activeAlerts = append(activeAlerts, alert)
		}
	}
	return activeAlerts
}

// generateOverallRecommendations generates recommendations for improving success rates
func (srm *SuccessRateMonitor) generateOverallRecommendations(metrics *ProcessMetrics, overallMetrics OverallMetrics) []SuccessRecommendation {
	var recommendations []SuccessRecommendation

	// Check if success rate is below target
	if metrics.SuccessRate < srm.config.TargetSuccessRate {
		recommendations = append(recommendations, SuccessRecommendation{
			ID:          "rec_low_success_rate",
			ProcessName: metrics.ProcessName,
			Category:    "performance",
			Priority:    1,
			Title:       "Improve Success Rate",
			Description: fmt.Sprintf("Current success rate %.2f%% is below target %.2f%%. Investigate failure patterns and implement improvements.", metrics.SuccessRate*100, srm.config.TargetSuccessRate*100),
			Impact:      "high",
			Effort:      "medium",
			Confidence:  0.9,
			CreatedAt:   time.Now(),
		})
	}

	// Check for high failure patterns
	if len(metrics.FailurePatterns) > 0 {
		recommendations = append(recommendations, SuccessRecommendation{
			ID:          "rec_failure_patterns",
			ProcessName: metrics.ProcessName,
			Category:    "analysis",
			Priority:    2,
			Title:       "Analyze Failure Patterns",
			Description: fmt.Sprintf("Found %d failure patterns. Conduct detailed analysis to identify root causes.", len(metrics.FailurePatterns)),
			Impact:      "medium",
			Effort:      "low",
			Confidence:  0.8,
			CreatedAt:   time.Now(),
		})
	}

	// Check response time
	if metrics.AverageResponseTime > 5*time.Second {
		recommendations = append(recommendations, SuccessRecommendation{
			ID:          "rec_high_response_time",
			ProcessName: metrics.ProcessName,
			Category:    "performance",
			Priority:    3,
			Title:       "Optimize Response Time",
			Description: fmt.Sprintf("Average response time %.2fs is high. Consider performance optimizations.", metrics.AverageResponseTime.Seconds()),
			Impact:      "medium",
			Effort:      "high",
			Confidence:  0.7,
			CreatedAt:   time.Now(),
		})
	}

	// Check if this is the worst performing process
	if overallMetrics.WorstProcess == metrics.ProcessName {
		recommendations = append(recommendations, SuccessRecommendation{
			ID:          "rec_worst_performer",
			ProcessName: metrics.ProcessName,
			Category:    "priority",
			Priority:    1,
			Title:       "Priority Improvement Required",
			Description: "This process has the lowest success rate among all processes. Prioritize improvements.",
			Impact:      "high",
			Effort:      "high",
			Confidence:  0.9,
			CreatedAt:   time.Now(),
		})
	}

	return recommendations
}

// CreateSuccessRateOptimization generates comprehensive optimization strategies for improving success rates
func (srm *SuccessRateMonitor) CreateSuccessRateOptimization(ctx context.Context, processName string) (*SuccessRateOptimization, error) {
	srm.mu.RLock()
	defer srm.mu.RUnlock()

	// Get process metrics
	metrics, exists := srm.metrics[processName]
	if !exists {
		return nil, fmt.Errorf("no metrics found for process: %s", processName)
	}

	// Get failure analysis for optimization insights
	failureAnalysis, err := srm.AnalyzeFailures(ctx, processName)
	if err != nil {
		srm.logger.Warn("Failed to analyze failures for optimization",
			zap.String("process_name", processName),
			zap.Error(err))
	}

	// Calculate optimization gap
	optimizationGap := srm.config.TargetSuccessRate - metrics.SuccessRate
	if optimizationGap < 0 {
		optimizationGap = 0 // Already meeting target
	}

	// Generate optimization strategies
	optimizationStrategies := srm.generateOptimizationStrategies(metrics, failureAnalysis)

	// Generate performance tuning recommendations
	performanceTuning := srm.generatePerformanceTuning(metrics, failureAnalysis)

	// Generate process improvements
	processImprovements := srm.generateProcessImprovements(metrics, failureAnalysis)

	// Generate resource optimization
	resourceOptimization := srm.generateResourceOptimization(metrics, failureAnalysis)

	// Generate configuration optimization
	configurationOptimization := srm.generateConfigurationOptimization(metrics, failureAnalysis)

	// Calculate expected improvement
	expectedImprovement := srm.calculateExpectedImprovement(optimizationStrategies, performanceTuning, processImprovements)

	// Generate implementation plan
	implementationPlan := srm.generateImplementationPlan(optimizationStrategies, expectedImprovement)

	// Get optimization status
	optimizationStatus := srm.getOptimizationStatus(processName)

	optimization := &SuccessRateOptimization{
		ProcessName:               processName,
		GeneratedAt:               time.Now(),
		CurrentSuccessRate:        metrics.SuccessRate,
		TargetSuccessRate:         srm.config.TargetSuccessRate,
		OptimizationGap:           optimizationGap,
		OptimizationStrategies:    optimizationStrategies,
		PerformanceTuning:         performanceTuning,
		ProcessImprovements:       processImprovements,
		ResourceOptimization:      resourceOptimization,
		ConfigurationOptimization: configurationOptimization,
		ExpectedImprovement:       expectedImprovement,
		ImplementationPlan:        implementationPlan,
		OptimizationStatus:        optimizationStatus,
	}

	srm.logger.Info("Generated success rate optimization",
		zap.String("process_name", processName),
		zap.Float64("current_rate", metrics.SuccessRate),
		zap.Float64("target_rate", srm.config.TargetSuccessRate),
		zap.Float64("optimization_gap", optimizationGap),
		zap.Float64("expected_improvement", expectedImprovement),
		zap.Int("strategies_count", len(optimizationStrategies)))

	return optimization, nil
}

// generateOptimizationStrategies generates optimization strategies based on metrics and failure analysis
func (srm *SuccessRateMonitor) generateOptimizationStrategies(metrics *ProcessMetrics, failureAnalysis *FailureAnalysis) []OptimizationStrategy {
	var strategies []OptimizationStrategy

	// Strategy 1: Error handling optimization
	if len(metrics.FailurePatterns) > 0 {
		strategies = append(strategies, OptimizationStrategy{
			ID:                   "error_handling_optimization",
			Name:                 "Error Handling Optimization",
			Category:             "error_management",
			Description:          "Improve error handling and recovery mechanisms based on failure patterns",
			ExpectedImpact:       0.05, // 5% improvement
			ImplementationEffort: "medium",
			RiskLevel:            "low",
			Confidence:           0.8,
			Prerequisites:        []string{"failure_analysis_complete"},
			Steps: []string{
				"Analyze common error patterns",
				"Implement retry mechanisms",
				"Add circuit breaker patterns",
				"Improve error logging and monitoring",
			},
			Metrics:   []string{"error_rate", "recovery_time", "success_rate"},
			CreatedAt: time.Now(),
		})
	}

	// Strategy 2: Response time optimization
	if metrics.AverageResponseTime > 2*time.Second {
		strategies = append(strategies, OptimizationStrategy{
			ID:                   "response_time_optimization",
			Name:                 "Response Time Optimization",
			Category:             "performance",
			Description:          "Optimize response times to improve user experience and reduce timeouts",
			ExpectedImpact:       0.03, // 3% improvement
			ImplementationEffort: "high",
			RiskLevel:            "medium",
			Confidence:           0.7,
			Prerequisites:        []string{"performance_analysis"},
			Steps: []string{
				"Identify performance bottlenecks",
				"Optimize database queries",
				"Implement caching strategies",
				"Add connection pooling",
			},
			Metrics:   []string{"response_time", "throughput", "success_rate"},
			CreatedAt: time.Now(),
		})
	}

	// Strategy 3: Input validation optimization
	if failureAnalysis != nil && len(failureAnalysis.CommonErrorTypes) > 0 {
		strategies = append(strategies, OptimizationStrategy{
			ID:                   "input_validation_optimization",
			Name:                 "Input Validation Optimization",
			Category:             "validation",
			Description:          "Improve input validation to prevent invalid data from causing failures",
			ExpectedImpact:       0.04, // 4% improvement
			ImplementationEffort: "low",
			RiskLevel:            "low",
			Confidence:           0.9,
			Prerequisites:        []string{"error_analysis"},
			Steps: []string{
				"Analyze validation error patterns",
				"Enhance input validation rules",
				"Add client-side validation",
				"Improve error messages",
			},
			Metrics:   []string{"validation_error_rate", "success_rate"},
			CreatedAt: time.Now(),
		})
	}

	// Strategy 4: Resource optimization
	if metrics.TotalAttempts > 1000 {
		strategies = append(strategies, OptimizationStrategy{
			ID:                   "resource_optimization",
			Name:                 "Resource Optimization",
			Category:             "infrastructure",
			Description:          "Optimize resource allocation and usage to improve performance",
			ExpectedImpact:       0.02, // 2% improvement
			ImplementationEffort: "medium",
			RiskLevel:            "medium",
			Confidence:           0.6,
			Prerequisites:        []string{"resource_monitoring"},
			Steps: []string{
				"Monitor resource usage",
				"Optimize memory allocation",
				"Improve CPU utilization",
				"Add resource scaling",
			},
			Metrics:   []string{"cpu_usage", "memory_usage", "success_rate"},
			CreatedAt: time.Now(),
		})
	}

	return strategies
}

// generatePerformanceTuning generates performance tuning recommendations
func (srm *SuccessRateMonitor) generatePerformanceTuning(metrics *ProcessMetrics, failureAnalysis *FailureAnalysis) PerformanceTuning {
	// Response time optimization
	responseTimeOptimization := srm.generateResponseTimeOptimization(metrics, failureAnalysis)

	// Throughput optimization
	throughputOptimization := srm.generateThroughputOptimization(metrics)

	// Resource optimization
	resourceOptimization := srm.generateResourceTuning(metrics)

	// Concurrency optimization
	concurrencyOptimization := srm.generateConcurrencyOptimization(metrics)

	// Cache optimization
	cacheOptimization := srm.generateCacheOptimization(metrics)

	return PerformanceTuning{
		ResponseTimeOptimization: responseTimeOptimization,
		ThroughputOptimization:   throughputOptimization,
		ResourceOptimization:     resourceOptimization,
		ConcurrencyOptimization:  concurrencyOptimization,
		CacheOptimization:        cacheOptimization,
	}
}

// generateResponseTimeOptimization generates response time optimization recommendations
func (srm *SuccessRateMonitor) generateResponseTimeOptimization(metrics *ProcessMetrics, failureAnalysis *FailureAnalysis) ResponseTimeOptimization {
	targetResponseTime := 1 * time.Second
	bottlenecks := []Bottleneck{}

	// Identify bottlenecks based on failure analysis
	if failureAnalysis != nil {
		if len(failureAnalysis.ProcessingStageFailures) > 0 {
			bottlenecks = append(bottlenecks, Bottleneck{
				ID:          "processing_stage_bottleneck",
				Name:        "Processing Stage Bottleneck",
				Type:        "processing",
				Severity:    "medium",
				Impact:      0.3,
				Description: "Some processing stages are taking longer than expected",
				Solution:    "Optimize slow processing stages and add parallel processing",
			})
		}

		if failureAnalysis.CorrelationAnalysis.LoadCorrelation > 0.7 {
			bottlenecks = append(bottlenecks, Bottleneck{
				ID:          "high_load_bottleneck",
				Name:        "High Load Bottleneck",
				Type:        "load",
				Severity:    "high",
				Impact:      0.5,
				Description: "Response times increase significantly under high load",
				Solution:    "Implement load balancing and horizontal scaling",
			})
		}
	}

	// Calculate expected improvement
	expectedImprovement := metrics.AverageResponseTime - targetResponseTime
	if expectedImprovement < 0 {
		expectedImprovement = 0
	}

	optimizationStrategies := []string{
		"Implement connection pooling",
		"Add caching layers",
		"Optimize database queries",
		"Use async processing where possible",
	}

	return ResponseTimeOptimization{
		CurrentAverageResponseTime: metrics.AverageResponseTime,
		TargetResponseTime:         targetResponseTime,
		Bottlenecks:                bottlenecks,
		OptimizationStrategies:     optimizationStrategies,
		ExpectedImprovement:        expectedImprovement,
	}
}

// generateThroughputOptimization generates throughput optimization recommendations
func (srm *SuccessRateMonitor) generateThroughputOptimization(metrics *ProcessMetrics) ThroughputOptimization {
	// Calculate current throughput (requests per second)
	currentThroughput := float64(metrics.TotalAttempts) / time.Since(metrics.LastUpdated).Seconds()
	targetThroughput := currentThroughput * 1.5 // 50% improvement target
	throughputGap := targetThroughput - currentThroughput

	if throughputGap < 0 {
		throughputGap = 0
	}

	optimizationStrategies := []string{
		"Increase concurrency limits",
		"Implement request batching",
		"Add horizontal scaling",
		"Optimize resource allocation",
	}

	return ThroughputOptimization{
		CurrentThroughput:      currentThroughput,
		TargetThroughput:       targetThroughput,
		ThroughputGap:          throughputGap,
		OptimizationStrategies: optimizationStrategies,
		ExpectedImprovement:    throughputGap,
	}
}

// generateResourceTuning generates resource tuning recommendations
func (srm *SuccessRateMonitor) generateResourceTuning(metrics *ProcessMetrics) ResourceTuning {
	// CPU optimization
	cpuOptimization := CPUOptimization{
		CurrentCPUUsage: 70.0, // Placeholder - would come from actual monitoring
		TargetCPUUsage:  80.0,
		OptimizationStrategies: []string{
			"Optimize CPU-intensive operations",
			"Implement worker pools",
			"Add CPU profiling",
		},
		ExpectedImprovement: 0.02,
	}

	// Memory optimization
	memoryOptimization := MemoryOptimization{
		CurrentMemoryUsage: 60.0, // Placeholder
		TargetMemoryUsage:  75.0,
		OptimizationStrategies: []string{
			"Implement memory pooling",
			"Add garbage collection tuning",
			"Optimize data structures",
		},
		ExpectedImprovement: 0.03,
	}

	// Network optimization
	networkOptimization := NetworkOptimization{
		CurrentNetworkLatency: 100 * time.Millisecond, // Placeholder
		TargetNetworkLatency:  50 * time.Millisecond,
		OptimizationStrategies: []string{
			"Use connection pooling",
			"Implement request compression",
			"Add CDN for static content",
		},
		ExpectedImprovement: 50 * time.Millisecond,
	}

	// Disk optimization
	diskOptimization := DiskOptimization{
		CurrentDiskUsage: 50.0, // Placeholder
		TargetDiskUsage:  70.0,
		OptimizationStrategies: []string{
			"Implement disk I/O optimization",
			"Add SSD storage",
			"Optimize file operations",
		},
		ExpectedImprovement: 0.01,
	}

	return ResourceTuning{
		CPUOptimization:     cpuOptimization,
		MemoryOptimization:  memoryOptimization,
		NetworkOptimization: networkOptimization,
		DiskOptimization:    diskOptimization,
	}
}

// generateConcurrencyOptimization generates concurrency optimization recommendations
func (srm *SuccessRateMonitor) generateConcurrencyOptimization(metrics *ProcessMetrics) ConcurrencyOptimization {
	currentConcurrency := 10 // Placeholder - would come from actual monitoring
	optimalConcurrency := currentConcurrency * 2

	optimizationStrategies := []string{
		"Increase worker pool size",
		"Implement request queuing",
		"Add concurrency monitoring",
		"Optimize thread management",
	}

	return ConcurrencyOptimization{
		CurrentConcurrency:     currentConcurrency,
		OptimalConcurrency:     optimalConcurrency,
		OptimizationStrategies: optimizationStrategies,
		ExpectedImprovement:    0.04,
	}
}

// generateCacheOptimization generates cache optimization recommendations
func (srm *SuccessRateMonitor) generateCacheOptimization(metrics *ProcessMetrics) CacheOptimization {
	currentCacheHitRate := 0.6 // Placeholder - would come from actual monitoring
	targetCacheHitRate := 0.8

	optimizationStrategies := []string{
		"Implement multi-level caching",
		"Add cache warming strategies",
		"Optimize cache eviction policies",
		"Add cache monitoring",
	}

	return CacheOptimization{
		CurrentCacheHitRate:    currentCacheHitRate,
		TargetCacheHitRate:     targetCacheHitRate,
		OptimizationStrategies: optimizationStrategies,
		ExpectedImprovement:    0.02,
	}
}

// calculateExpectedImprovement calculates the expected improvement from all optimizations
func (srm *SuccessRateMonitor) calculateExpectedImprovement(strategies []OptimizationStrategy, performanceTuning PerformanceTuning, processImprovements []ProcessImprovement) float64 {
	totalImprovement := 0.0

	// Add improvements from strategies
	for _, strategy := range strategies {
		totalImprovement += strategy.ExpectedImpact
	}

	// Add improvements from performance tuning
	totalImprovement += 0.02 // Estimated 2% from performance tuning

	// Add improvements from process improvements
	for _, improvement := range processImprovements {
		totalImprovement += improvement.ExpectedImpact
	}

	// Cap improvement at 20% to be realistic
	if totalImprovement > 0.20 {
		totalImprovement = 0.20
	}

	return totalImprovement
}

// getOptimizationStatus gets the current optimization status for a process
func (srm *SuccessRateMonitor) getOptimizationStatus(processName string) OptimizationStatus {
	// This would typically come from a database or persistent storage
	// For now, return a default status
	return OptimizationStatus{
		Status:            "pending",
		LastOptimized:     time.Time{},
		OptimizationCount: 0,
		SuccessRate:       0.0,
		Improvement:       0.0,
		NextOptimization:  time.Now().Add(24 * time.Hour),
	}
}

// generateProcessImprovements generates process improvement recommendations
func (srm *SuccessRateMonitor) generateProcessImprovements(metrics *ProcessMetrics, failureAnalysis *FailureAnalysis) []ProcessImprovement {
	var improvements []ProcessImprovement

	// Improvement 1: Error handling process
	if failureAnalysis != nil && len(failureAnalysis.CommonErrorTypes) > 0 {
		improvements = append(improvements, ProcessImprovement{
			ID:           "error_handling_process",
			Name:         "Error Handling Process Improvement",
			Category:     "error_management",
			Description:  "Improve error handling and recovery processes",
			CurrentState: "Basic error handling",
			TargetState:  "Comprehensive error handling with recovery",
			ImprovementSteps: []string{
				"Implement structured error handling",
				"Add error categorization",
				"Create error recovery procedures",
				"Add error monitoring and alerting",
			},
			ExpectedImpact:       0.03,
			ImplementationEffort: "medium",
			RiskLevel:            "low",
			CreatedAt:            time.Now(),
		})
	}

	// Improvement 2: Input validation process
	if failureAnalysis != nil && len(failureAnalysis.ProblematicInputTypes) > 0 {
		improvements = append(improvements, ProcessImprovement{
			ID:           "input_validation_process",
			Name:         "Input Validation Process Improvement",
			Category:     "validation",
			Description:  "Improve input validation and preprocessing",
			CurrentState: "Basic validation",
			TargetState:  "Comprehensive validation with preprocessing",
			ImprovementSteps: []string{
				"Implement multi-stage validation",
				"Add input preprocessing",
				"Create validation rules engine",
				"Add validation feedback loops",
			},
			ExpectedImpact:       0.04,
			ImplementationEffort: "low",
			RiskLevel:            "low",
			CreatedAt:            time.Now(),
		})
	}

	// Improvement 3: Monitoring and alerting process
	improvements = append(improvements, ProcessImprovement{
		ID:           "monitoring_process",
		Name:         "Monitoring and Alerting Process Improvement",
		Category:     "monitoring",
		Description:  "Improve monitoring and alerting processes",
		CurrentState: "Basic monitoring",
		TargetState:  "Comprehensive monitoring with predictive alerts",
		ImprovementSteps: []string{
			"Implement real-time monitoring",
			"Add predictive alerting",
			"Create escalation procedures",
			"Add performance dashboards",
		},
		ExpectedImpact:       0.02,
		ImplementationEffort: "medium",
		RiskLevel:            "low",
		CreatedAt:            time.Now(),
	})

	return improvements
}

// generateResourceOptimization generates resource optimization recommendations
func (srm *SuccessRateMonitor) generateResourceOptimization(metrics *ProcessMetrics, failureAnalysis *FailureAnalysis) ResourceOptimization {
	// Scaling recommendations
	scalingRecommendations := ScalingRecommendations{
		HorizontalScaling: HorizontalScaling{
			RecommendedInstances: 3,
			ScalingTriggers:      []string{"high_cpu_usage", "high_memory_usage", "high_response_time"},
			ExpectedImprovement:  0.03,
			ImplementationSteps:  []string{"Deploy load balancer", "Configure auto-scaling", "Set up monitoring"},
		},
		VerticalScaling: VerticalScaling{
			RecommendedResources: ResourceSpec{
				CPU:     "4 cores",
				Memory:  "8GB",
				Disk:    "100GB SSD",
				Network: "1Gbps",
			},
			ScalingTriggers:     []string{"cpu_bottleneck", "memory_bottleneck"},
			ExpectedImprovement: 0.02,
			ImplementationSteps: []string{"Upgrade instance type", "Monitor performance", "Validate improvements"},
		},
		AutoScaling: AutoScaling{
			MinInstances:        2,
			MaxInstances:        10,
			TargetCPUUsage:      70.0,
			ScalingPolicies:     []string{"cpu_based", "memory_based", "response_time_based"},
			ExpectedImprovement: 0.04,
		},
	}

	// Resource allocation
	resourceAllocation := ResourceAllocation{
		CurrentAllocation: ResourceAllocationMap{
			CPUAllocation:     map[string]float64{"process": 0.7, "system": 0.3},
			MemoryAllocation:  map[string]float64{"process": 0.6, "system": 0.4},
			DiskAllocation:    map[string]float64{"data": 0.5, "logs": 0.3, "temp": 0.2},
			NetworkAllocation: map[string]float64{"inbound": 0.5, "outbound": 0.5},
		},
		OptimalAllocation: ResourceAllocationMap{
			CPUAllocation:     map[string]float64{"process": 0.8, "system": 0.2},
			MemoryAllocation:  map[string]float64{"process": 0.7, "system": 0.3},
			DiskAllocation:    map[string]float64{"data": 0.6, "logs": 0.2, "temp": 0.2},
			NetworkAllocation: map[string]float64{"inbound": 0.6, "outbound": 0.4},
		},
		OptimizationStrategies: []string{"Optimize resource allocation", "Implement resource quotas", "Add resource monitoring"},
		ExpectedImprovement:    0.02,
	}

	// Cost optimization
	costOptimization := CostOptimization{
		CurrentCost:            1000.0, // Placeholder
		OptimizedCost:          800.0,  // Placeholder
		CostSavings:            200.0,
		OptimizationStrategies: []string{"Use reserved instances", "Implement auto-scaling", "Optimize resource usage"},
		ROI:                    0.25,
	}

	// Capacity planning
	capacityPlanning := CapacityPlanning{
		CurrentCapacity:  1000,
		RequiredCapacity: 1500,
		CapacityGap:      500,
		PlanningHorizon:  30 * 24 * time.Hour, // 30 days
		GrowthProjections: []GrowthProjection{
			{
				Timeframe:        7 * 24 * time.Hour, // 1 week
				ProjectedGrowth:  0.1,                // 10%
				RequiredCapacity: 1100,
				Confidence:       0.8,
			},
			{
				Timeframe:        30 * 24 * time.Hour, // 1 month
				ProjectedGrowth:  0.5,                 // 50%
				RequiredCapacity: 1500,
				Confidence:       0.6,
			},
		},
		Recommendations: []string{"Scale horizontally", "Add monitoring", "Implement auto-scaling"},
	}

	return ResourceOptimization{
		ScalingRecommendations: scalingRecommendations,
		ResourceAllocation:     resourceAllocation,
		CostOptimization:       costOptimization,
		CapacityPlanning:       capacityPlanning,
	}
}

// generateConfigurationOptimization generates configuration optimization recommendations
func (srm *SuccessRateMonitor) generateConfigurationOptimization(metrics *ProcessMetrics, failureAnalysis *FailureAnalysis) ConfigurationOptimization {
	// Parameter optimization
	parameterOptimization := ParameterOptimization{
		CurrentParameters: map[string]interface{}{
			"max_connections": 100,
			"timeout":         30,
			"retry_attempts":  3,
			"cache_size":      1000,
		},
		OptimalParameters: map[string]interface{}{
			"max_connections": 200,
			"timeout":         15,
			"retry_attempts":  5,
			"cache_size":      2000,
		},
		OptimizationStrategies: []string{"Tune connection pool", "Optimize timeouts", "Increase retry attempts", "Expand cache size"},
		ExpectedImprovement:    0.03,
	}

	// Threshold optimization
	thresholdOptimization := ThresholdOptimization{
		CurrentThresholds: map[string]float64{
			"success_rate_warning":   0.90,
			"success_rate_critical":  0.85,
			"response_time_warning":  2.0,
			"response_time_critical": 5.0,
		},
		OptimalThresholds: map[string]float64{
			"success_rate_warning":   0.92,
			"success_rate_critical":  0.88,
			"response_time_warning":  1.5,
			"response_time_critical": 3.0,
		},
		OptimizationStrategies: []string{"Adjust warning thresholds", "Optimize critical thresholds", "Add predictive thresholds"},
		ExpectedImprovement:    0.01,
	}

	// Timeout optimization
	timeoutOptimization := TimeoutOptimization{
		CurrentTimeouts: map[string]time.Duration{
			"http_timeout":     30 * time.Second,
			"database_timeout": 10 * time.Second,
			"cache_timeout":    5 * time.Second,
		},
		OptimalTimeouts: map[string]time.Duration{
			"http_timeout":     15 * time.Second,
			"database_timeout": 5 * time.Second,
			"cache_timeout":    2 * time.Second,
		},
		OptimizationStrategies: []string{"Reduce HTTP timeout", "Optimize database timeout", "Decrease cache timeout"},
		ExpectedImprovement:    2 * time.Second,
	}

	// Retry optimization
	retryOptimization := RetryOptimization{
		CurrentRetryConfig: RetryConfig{
			MaxRetries:      3,
			RetryDelay:      1 * time.Second,
			BackoffFactor:   2.0,
			RetryableErrors: []string{"timeout", "connection_error", "temporary_error"},
		},
		OptimalRetryConfig: RetryConfig{
			MaxRetries:      5,
			RetryDelay:      500 * time.Millisecond,
			BackoffFactor:   1.5,
			RetryableErrors: []string{"timeout", "connection_error", "temporary_error", "rate_limit"},
		},
		OptimizationStrategies: []string{"Increase retry attempts", "Optimize retry delay", "Add exponential backoff", "Expand retryable errors"},
		ExpectedImprovement:    0.02,
	}

	return ConfigurationOptimization{
		ParameterOptimization: parameterOptimization,
		ThresholdOptimization: thresholdOptimization,
		TimeoutOptimization:   timeoutOptimization,
		RetryOptimization:     retryOptimization,
	}
}

// generateImplementationPlan generates implementation plan for optimizations
func (srm *SuccessRateMonitor) generateImplementationPlan(strategies []OptimizationStrategy, expectedImprovement float64) ImplementationPlan {
	phases := []ImplementationPhase{
		{
			PhaseNumber:     1,
			Name:            "Quick Wins",
			Description:     "Implement low-risk, high-impact optimizations",
			Duration:        7 * 24 * time.Hour, // 1 week
			Effort:          "low",
			Prerequisites:   []string{},
			Deliverables:    []string{"Configuration optimizations", "Basic monitoring improvements"},
			SuccessCriteria: []string{"5% success rate improvement", "Reduced response times"},
			RiskLevel:       "low",
		},
		{
			PhaseNumber:     2,
			Name:            "Performance Optimization",
			Description:     "Implement performance tuning and resource optimization",
			Duration:        14 * 24 * time.Hour, // 2 weeks
			Effort:          "medium",
			Prerequisites:   []string{"Phase 1 complete"},
			Deliverables:    []string{"Performance optimizations", "Resource improvements"},
			SuccessCriteria: []string{"10% success rate improvement", "Improved throughput"},
			RiskLevel:       "medium",
		},
		{
			PhaseNumber:     3,
			Name:            "Advanced Optimizations",
			Description:     "Implement advanced optimizations and scaling",
			Duration:        21 * 24 * time.Hour, // 3 weeks
			Effort:          "high",
			Prerequisites:   []string{"Phase 2 complete"},
			Deliverables:    []string{"Advanced optimizations", "Scaling improvements"},
			SuccessCriteria: []string{"15% success rate improvement", "Achieve target success rate"},
			RiskLevel:       "high",
		},
	}

	// Calculate total duration
	totalDuration := 0 * time.Hour
	for _, phase := range phases {
		totalDuration += phase.Duration
	}

	// Risk assessment
	riskAssessment := RiskAssessment{
		OverallRiskLevel: "medium",
		RiskFactors: []RiskFactor{
			{
				ID:          "performance_risk",
				Name:        "Performance Risk",
				RiskLevel:   "medium",
				Probability: 0.3,
				Impact:      0.4,
				Description: "Performance optimizations may introduce new issues",
				Mitigation:  "Implement changes gradually with monitoring",
			},
			{
				ID:          "scaling_risk",
				Name:        "Scaling Risk",
				RiskLevel:   "high",
				Probability: 0.2,
				Impact:      0.6,
				Description: "Scaling changes may affect system stability",
				Mitigation:  "Test thoroughly in staging environment",
			},
		},
		MitigationStrategies: []string{
			"Implement changes gradually",
			"Monitor performance closely",
			"Have rollback plan ready",
			"Test in staging environment",
		},
		ContingencyPlans: []string{
			"Rollback to previous configuration",
			"Disable problematic optimizations",
			"Scale back to previous capacity",
		},
	}

	// Rollback plan
	rollbackPlan := RollbackPlan{
		RollbackTriggers: []string{
			"Success rate drops below 85%",
			"Response time increases by 50%",
			"Error rate increases by 25%",
		},
		RollbackSteps: []string{
			"Stop new optimizations",
			"Revert configuration changes",
			"Scale back to previous capacity",
			"Monitor system stability",
		},
		RollbackDuration: 1 * time.Hour,
		DataBackup:       "Daily automated backups",
	}

	return ImplementationPlan{
		Phases:         phases,
		TotalDuration:  totalDuration,
		TotalEffort:    "medium",
		RiskAssessment: riskAssessment,
		SuccessCriteria: []string{
			"Achieve 95% success rate",
			"Reduce response time by 50%",
			"Improve throughput by 100%",
		},
		RollbackPlan: rollbackPlan,
	}
}
