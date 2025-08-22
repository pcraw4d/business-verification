package middleware

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

// ResourceAlertingScalingManager manages advanced alerting and auto-scaling for resource utilization
type ResourceAlertingScalingManager struct {
	config           *AlertingScalingConfig
	alertEngine      *AlertEngine
	scalingEngine    *AutoScalingEngine
	metricCollector  *EnhancedMetricCollector
	escalationEngine *EscalationEngine
	notificationMgr  *NotificationManager
	mu               sync.RWMutex
	ctx              context.Context
	cancel           context.CancelFunc
	alertingDone     chan struct{}
	scalingDone      chan struct{}
}

// AlertingScalingConfig holds configuration for enhanced alerting and scaling
type AlertingScalingConfig struct {
	AlertingInterval        time.Duration             // How often to check for alerts
	ScalingInterval         time.Duration             // How often to check for scaling
	MetricRetentionPeriod   time.Duration             // How long to retain metrics
	AlertRetentionPeriod    time.Duration             // How long to retain alert history
	ScalingCooldownPeriod   time.Duration             // Cooldown between scaling actions
	EnableProactiveScaling  bool                      // Enable predictive scaling
	EnableAdaptiveThresholds bool                     // Enable dynamic threshold adjustment
	AlertThresholds         *EnhancedAlertThresholds  // Enhanced alert thresholds
	ScalingPolicies         *AutoScalingPolicies      // Auto-scaling policies
	NotificationChannels    []*NotificationChannel    // Notification channels
	EscalationPolicies      []*EscalationPolicy       // Escalation policies
}

// EnhancedAlertThresholds defines comprehensive alerting thresholds
type EnhancedAlertThresholds struct {
	// CPU Thresholds
	CPUWarning            float64 // CPU usage warning threshold
	CPUCritical           float64 // CPU usage critical threshold
	CPUEmergency          float64 // CPU usage emergency threshold
	CPULoadAvgWarning     float64 // Load average warning threshold
	CPULoadAvgCritical    float64 // Load average critical threshold
	
	// Memory Thresholds  
	MemoryWarning         float64 // Memory usage warning threshold
	MemoryCritical        float64 // Memory usage critical threshold
	MemoryEmergency       float64 // Memory usage emergency threshold
	HeapWarning           float64 // Heap usage warning threshold
	HeapCritical          float64 // Heap usage critical threshold
	
	// Goroutine Thresholds
	GoroutineWarning      int     // Goroutine count warning threshold
	GoroutineCritical     int     // Goroutine count critical threshold
	GoroutineEmergency    int     // Goroutine count emergency threshold
	
	// Performance Thresholds
	ResponseTimeWarning   time.Duration // Response time warning threshold
	ResponseTimeCritical  time.Duration // Response time critical threshold
	ThroughputWarning     float64       // Throughput warning threshold (RPS)
	ThroughputCritical    float64       // Throughput critical threshold (RPS)
	ErrorRateWarning      float64       // Error rate warning threshold (%)
	ErrorRateCritical     float64       // Error rate critical threshold (%)
	
	// Resource Utilization Thresholds
	DiskIOWarning         float64 // Disk I/O utilization warning threshold
	DiskIOCritical        float64 // Disk I/O utilization critical threshold
	NetworkIOWarning      float64 // Network I/O utilization warning threshold
	NetworkIOCritical     float64 // Network I/O utilization critical threshold
	
	// Adaptive Thresholds
	AdaptiveEnabled       bool    // Enable adaptive threshold adjustment
	AdaptiveWindow        time.Duration // Window for adaptive calculations
	AdaptiveSensitivity   float64 // Sensitivity for adaptive adjustments (0.0-1.0)
}

// AutoScalingPolicies defines auto-scaling behavior
type AutoScalingPolicies struct {
	// Scaling Triggers
	CPUScaleUpThreshold     float64 // CPU threshold to trigger scale up
	CPUScaleDownThreshold   float64 // CPU threshold to trigger scale down
	MemoryScaleUpThreshold  float64 // Memory threshold to trigger scale up
	MemoryScaleDownThreshold float64 // Memory threshold to trigger scale down
	
	// Scaling Parameters
	MinInstances            int           // Minimum number of instances
	MaxInstances            int           // Maximum number of instances
	ScaleUpIncrement        int           // Number of instances to add on scale up
	ScaleDownDecrement      int           // Number of instances to remove on scale down
	ScaleUpCooldown         time.Duration // Cooldown after scale up
	ScaleDownCooldown       time.Duration // Cooldown after scale down
	
	// Advanced Scaling
	PredictiveScalingEnabled bool     // Enable predictive scaling
	PredictiveWindow         time.Duration // Window for predictive analysis
	PredictiveSensitivity    float64  // Sensitivity for predictive scaling (0.0-1.0)
	TargetUtilization        float64  // Target utilization percentage
	
	// Scaling Strategies
	ScaleUpStrategy         ScalingStrategy // Strategy for scaling up
	ScaleDownStrategy       ScalingStrategy // Strategy for scaling down
}

// ScalingStrategy defines how scaling should be performed
type ScalingStrategy string

const (
	ScalingStrategyConservative ScalingStrategy = "conservative" // Gradual scaling
	ScalingStrategyAggressive   ScalingStrategy = "aggressive"   // Rapid scaling
	ScalingStrategyPredictive   ScalingStrategy = "predictive"   // AI-based scaling
	ScalingStrategyAdaptive     ScalingStrategy = "adaptive"     // Self-adjusting scaling
)

// NotificationChannel defines notification delivery methods
type NotificationChannel struct {
	ID          string            `json:"id"`
	Type        NotificationType  `json:"type"`
	Endpoint    string            `json:"endpoint"`
	Enabled     bool              `json:"enabled"`
	Filters     []AlertLevel      `json:"filters"`
	Config      map[string]string `json:"config"`
	RateLimit   *NotificationRateLimit `json:"rate_limit,omitempty"`
}

// NotificationType defines the type of notification channel
type NotificationType string

const (
	NotificationTypeEmail    NotificationType = "email"
	NotificationTypeSlack    NotificationType = "slack"
	NotificationTypeWebhook  NotificationType = "webhook"
	NotificationTypeSMS      NotificationType = "sms"
	NotificationTypePagerDuty NotificationType = "pagerduty"
)

// EscalationPolicy defines how alerts should be escalated
type EscalationPolicy struct {
	ID                string               `json:"id"`
	AlertTypes        []string             `json:"alert_types"`
	Levels            []AlertLevel         `json:"levels"`
	EscalationSteps   []*EscalationStep    `json:"escalation_steps"`
	TimeoutDuration   time.Duration        `json:"timeout_duration"`
	MaxEscalations    int                  `json:"max_escalations"`
	Enabled           bool                 `json:"enabled"`
}

// EscalationStep defines a step in the escalation process
type EscalationStep struct {
	StepNumber          int               `json:"step_number"`
	DelayDuration       time.Duration     `json:"delay_duration"`
	NotificationChannels []string         `json:"notification_channels"`
	Actions             []*EscalationAction `json:"actions"`
	RequireAcknowledgment bool            `json:"require_acknowledgment"`
}

// EscalationAction defines an action to take during escalation
type EscalationAction struct {
	Type       EscalationActionType  `json:"type"`
	Parameters map[string]interface{} `json:"parameters"`
}

// EscalationActionType defines the type of escalation action
type EscalationActionType string

const (
	EscalationActionScale     EscalationActionType = "scale"
	EscalationActionRestart   EscalationActionType = "restart"
	EscalationActionThrottle  EscalationActionType = "throttle"
	EscalationActionProfile   EscalationActionType = "profile"
	EscalationActionNotify    EscalationActionType = "notify"
)

// NotificationRateLimit defines rate limiting for notifications
type NotificationRateLimit struct {
	MaxNotifications int           `json:"max_notifications"`
	TimeWindow       time.Duration `json:"time_window"`
}

// AlertEngine manages advanced alerting logic
type AlertEngine struct {
	config          *AlertingScalingConfig
	thresholds      *EnhancedAlertThresholds
	alertHistory    []*EnhancedAlert
	activeAlerts    map[string]*EnhancedAlert
	adaptiveMetrics *AdaptiveMetrics
	mu              sync.RWMutex
}

// EnhancedAlert represents an enhanced alert with additional context
type EnhancedAlert struct {
	ID              string                 `json:"id"`
	Type            AlertType              `json:"type"`
	Level           AlertLevel             `json:"level"`
	Title           string                 `json:"title"`
	Description     string                 `json:"description"`
	Resource        string                 `json:"resource"`
	Metric          string                 `json:"metric"`
	CurrentValue    float64                `json:"current_value"`
	ThresholdValue  float64                `json:"threshold_value"`
	Severity        AlertSeverity          `json:"severity"`
	Tags            map[string]string      `json:"tags"`
	Metadata        map[string]interface{} `json:"metadata"`
	Timestamp       time.Time              `json:"timestamp"`
	ResolvedAt      *time.Time             `json:"resolved_at,omitempty"`
	AcknowledgedAt  *time.Time             `json:"acknowledged_at,omitempty"`
	EscalatedAt     *time.Time             `json:"escalated_at,omitempty"`
	Suppressed      bool                   `json:"suppressed"`
	SuppressionRule string                 `json:"suppression_rule,omitempty"`
}

// AlertType defines the type of alert
type AlertType string

const (
	AlertTypeCPU          AlertType = "cpu"
	AlertTypeMemory       AlertType = "memory"
	AlertTypeGoroutine    AlertType = "goroutine"
	AlertTypeDiskIO       AlertType = "disk_io"
	AlertTypeNetworkIO    AlertType = "network_io"
	AlertTypePerformance  AlertType = "performance"
	AlertTypeAvailability AlertType = "availability"
	AlertTypeSystem       AlertType = "system"
)

// AlertLevel defines the severity level of an alert
type AlertLevel string

const (
	AlertLevelInfo      AlertLevel = "info"
	AlertLevelWarning   AlertLevel = "warning"
	AlertLevelCritical  AlertLevel = "critical"
	AlertLevelEmergency AlertLevel = "emergency"
)

// AlertSeverity defines the business impact severity
type AlertSeverity int

const (
	AlertSeverityLow AlertSeverity = iota
	AlertSeverityMedium
	AlertSeverityHigh
	AlertSeverityCritical
)

// AutoScalingEngine manages automatic scaling operations
type AutoScalingEngine struct {
	config          *AlertingScalingConfig
	policies        *AutoScalingPolicies
	scalingHistory  []*ScalingEvent
	currentInstances int
	lastScalingTime time.Time
	predictiveModel *PredictiveModel
	mu              sync.RWMutex
}

// ScalingEvent represents a scaling operation
type ScalingEvent struct {
	ID              string          `json:"id"`
	Type            ScalingType     `json:"type"`
	Trigger         ScalingTrigger  `json:"trigger"`
	Strategy        ScalingStrategy `json:"strategy"`
	InstancesBefore int             `json:"instances_before"`
	InstancesAfter  int             `json:"instances_after"`
	Reason          string          `json:"reason"`
	Metadata        map[string]interface{} `json:"metadata"`
	Timestamp       time.Time       `json:"timestamp"`
	Duration        time.Duration   `json:"duration"`
	Success         bool            `json:"success"`
	Error           string          `json:"error,omitempty"`
}

// ScalingType defines the type of scaling operation
type ScalingType string

const (
	ScalingTypeUp   ScalingType = "scale_up"
	ScalingTypeDown ScalingType = "scale_down"
)

// ScalingTrigger defines what triggered the scaling
type ScalingTrigger string

const (
	ScalingTriggerThreshold  ScalingTrigger = "threshold"
	ScalingTriggerPredictive ScalingTrigger = "predictive"
	ScalingTriggerManual     ScalingTrigger = "manual"
	ScalingTriggerScheduled  ScalingTrigger = "scheduled"
)

// EnhancedMetricCollector collects comprehensive metrics for alerting and scaling
type EnhancedMetricCollector struct {
	metrics       *EnhancedMetrics
	metricHistory []*MetricSnapshot
	mu            sync.RWMutex
}

// EnhancedMetrics represents comprehensive system metrics
type EnhancedMetrics struct {
	Timestamp           time.Time              `json:"timestamp"`
	
	// CPU Metrics
	CPUUsage            float64                `json:"cpu_usage"`
	CPUUsagePerCore     []float64              `json:"cpu_usage_per_core"`
	LoadAverage         []float64              `json:"load_average"`
	CPUFrequency        float64                `json:"cpu_frequency"`
	CPUTemperature      float64                `json:"cpu_temperature,omitempty"`
	
	// Memory Metrics
	MemoryUsage         float64                `json:"memory_usage"`
	MemoryTotal         uint64                 `json:"memory_total"`
	MemoryUsed          uint64                 `json:"memory_used"`
	MemoryAvailable     uint64                 `json:"memory_available"`
	MemoryBuffered      uint64                 `json:"memory_buffered"`
	MemoryCached        uint64                 `json:"memory_cached"`
	SwapUsage           float64                `json:"swap_usage"`
	SwapTotal           uint64                 `json:"swap_total"`
	SwapUsed            uint64                 `json:"swap_used"`
	
	// Go Runtime Metrics
	HeapAlloc           uint64                 `json:"heap_alloc"`
	HeapSys             uint64                 `json:"heap_sys"`
	HeapInuse           uint64                 `json:"heap_inuse"`
	HeapIdle            uint64                 `json:"heap_idle"`
	HeapReleased        uint64                 `json:"heap_released"`
	GoroutineCount      int                    `json:"goroutine_count"`
	GCCycles            uint32                 `json:"gc_cycles"`
	GCPause             time.Duration          `json:"gc_pause"`
	GCCPUFraction       float64                `json:"gc_cpu_fraction"`
	
	// Performance Metrics
	ResponseTime        time.Duration          `json:"response_time"`
	Throughput          float64                `json:"throughput"`
	ErrorRate           float64                `json:"error_rate"`
	ActiveConnections   int                    `json:"active_connections"`
	QueueLength         int                    `json:"queue_length"`
	
	// I/O Metrics
	DiskIOReads         uint64                 `json:"disk_io_reads"`
	DiskIOWrites        uint64                 `json:"disk_io_writes"`
	DiskIOUtilization   float64                `json:"disk_io_utilization"`
	NetworkBytesIn      uint64                 `json:"network_bytes_in"`
	NetworkBytesOut     uint64                 `json:"network_bytes_out"`
	NetworkPacketsIn    uint64                 `json:"network_packets_in"`
	NetworkPacketsOut   uint64                 `json:"network_packets_out"`
	
	// Custom Metrics
	CustomMetrics       map[string]float64     `json:"custom_metrics"`
}

// MetricSnapshot represents a point-in-time metric capture
type MetricSnapshot struct {
	Timestamp time.Time        `json:"timestamp"`
	Metrics   *EnhancedMetrics `json:"metrics"`
}

// AdaptiveMetrics tracks adaptive threshold adjustments
type AdaptiveMetrics struct {
	BaselineMetrics     map[string]float64    `json:"baseline_metrics"`
	MovingAverages      map[string]float64    `json:"moving_averages"`
	StandardDeviations  map[string]float64    `json:"standard_deviations"`
	AdaptedThresholds   map[string]float64    `json:"adapted_thresholds"`
	LastUpdated         time.Time             `json:"last_updated"`
}

// PredictiveModel represents a simple predictive scaling model
type PredictiveModel struct {
	MetricWindows       map[string][]float64  `json:"metric_windows"`
	Trends             map[string]float64     `json:"trends"`
	Seasonality        map[string][]float64   `json:"seasonality"`
	Predictions        map[string]float64     `json:"predictions"`
	Confidence         map[string]float64     `json:"confidence"`
	LastTrainingTime   time.Time              `json:"last_training_time"`
}

// EscalationEngine manages alert escalation
type EscalationEngine struct {
	config             *AlertingScalingConfig
	policies           []*EscalationPolicy
	activeEscalations  map[string]*ActiveEscalation
	escalationHistory  []*EscalationEvent
	mu                 sync.RWMutex
}

// ActiveEscalation represents an ongoing escalation
type ActiveEscalation struct {
	ID                string             `json:"id"`
	AlertID           string             `json:"alert_id"`
	PolicyID          string             `json:"policy_id"`
	CurrentStep       int                `json:"current_step"`
	StartedAt         time.Time          `json:"started_at"`
	LastEscalatedAt   time.Time          `json:"last_escalated_at"`
	AcknowledgedAt    *time.Time         `json:"acknowledged_at,omitempty"`
	ResolvedAt        *time.Time         `json:"resolved_at,omitempty"`
	EscalationSteps   []*CompletedStep   `json:"escalation_steps"`
}

// CompletedStep represents a completed escalation step
type CompletedStep struct {
	StepNumber        int        `json:"step_number"`
	CompletedAt       time.Time  `json:"completed_at"`
	Acknowledged      bool       `json:"acknowledged"`
	ActionsExecuted   []string   `json:"actions_executed"`
}

// EscalationEvent represents an escalation event
type EscalationEvent struct {
	ID                string                 `json:"id"`
	Type              EscalationEventType    `json:"type"`
	AlertID           string                 `json:"alert_id"`
	EscalationID      string                 `json:"escalation_id"`
	PolicyID          string                 `json:"policy_id"`
	StepNumber        int                    `json:"step_number"`
	Details           string                 `json:"details"`
	Metadata          map[string]interface{} `json:"metadata"`
	Timestamp         time.Time              `json:"timestamp"`
}

// EscalationEventType defines the type of escalation event
type EscalationEventType string

const (
	EscalationEventStarted     EscalationEventType = "started"
	EscalationEventEscalated   EscalationEventType = "escalated"
	EscalationEventAcknowledged EscalationEventType = "acknowledged"
	EscalationEventResolved    EscalationEventType = "resolved"
	EscalationEventTimedOut    EscalationEventType = "timed_out"
)

// NotificationManager manages notification delivery
type NotificationManager struct {
	channels         []*NotificationChannel
	notificationHistory []*NotificationEvent
	rateLimiters     map[string]*NotificationRateLimiter
	mu               sync.RWMutex
}

// NotificationEvent represents a notification event
type NotificationEvent struct {
	ID              string                 `json:"id"`
	ChannelID       string                 `json:"channel_id"`
	AlertID         string                 `json:"alert_id"`
	Type            NotificationType       `json:"type"`
	Recipient       string                 `json:"recipient"`
	Subject         string                 `json:"subject"`
	Message         string                 `json:"message"`
	Status          NotificationStatus     `json:"status"`
	Metadata        map[string]interface{} `json:"metadata"`
	SentAt          time.Time              `json:"sent_at"`
	DeliveredAt     *time.Time             `json:"delivered_at,omitempty"`
	Error           string                 `json:"error,omitempty"`
}

// NotificationStatus defines the status of a notification
type NotificationStatus string

const (
	NotificationStatusPending   NotificationStatus = "pending"
	NotificationStatusSent      NotificationStatus = "sent"
	NotificationStatusDelivered NotificationStatus = "delivered"
	NotificationStatusFailed    NotificationStatus = "failed"
)

// NotificationRateLimiter manages rate limiting for notifications
type NotificationRateLimiter struct {
	MaxNotifications int
	TimeWindow       time.Duration
	notifications    []time.Time
	mu               sync.Mutex
}

// DefaultAlertingScalingConfig creates a default configuration
func DefaultAlertingScalingConfig() *AlertingScalingConfig {
	return &AlertingScalingConfig{
		AlertingInterval:        30 * time.Second,
		ScalingInterval:         60 * time.Second,
		MetricRetentionPeriod:   24 * time.Hour,
		AlertRetentionPeriod:    7 * 24 * time.Hour,
		ScalingCooldownPeriod:   5 * time.Minute,
		EnableProactiveScaling:  true,
		EnableAdaptiveThresholds: true,
		AlertThresholds: &EnhancedAlertThresholds{
			CPUWarning:            70.0,
			CPUCritical:           85.0,
			CPUEmergency:          95.0,
			CPULoadAvgWarning:     2.0,
			CPULoadAvgCritical:    4.0,
			MemoryWarning:         70.0,
			MemoryCritical:        85.0,
			MemoryEmergency:       95.0,
			HeapWarning:           80.0,
			HeapCritical:          90.0,
			GoroutineWarning:      1000,
			GoroutineCritical:     2000,
			GoroutineEmergency:    5000,
			ResponseTimeWarning:   500 * time.Millisecond,
			ResponseTimeCritical:  1000 * time.Millisecond,
			ThroughputWarning:     100.0,
			ThroughputCritical:    50.0,
			ErrorRateWarning:      5.0,
			ErrorRateCritical:     10.0,
			DiskIOWarning:         80.0,
			DiskIOCritical:        90.0,
			NetworkIOWarning:      80.0,
			NetworkIOCritical:     90.0,
			AdaptiveEnabled:       true,
			AdaptiveWindow:        30 * time.Minute,
			AdaptiveSensitivity:   0.1,
		},
		ScalingPolicies: &AutoScalingPolicies{
			CPUScaleUpThreshold:      80.0,
			CPUScaleDownThreshold:    30.0,
			MemoryScaleUpThreshold:   80.0,
			MemoryScaleDownThreshold: 30.0,
			MinInstances:             1,
			MaxInstances:             10,
			ScaleUpIncrement:         1,
			ScaleDownDecrement:       1,
			ScaleUpCooldown:          5 * time.Minute,
			ScaleDownCooldown:        10 * time.Minute,
			PredictiveScalingEnabled: true,
			PredictiveWindow:         30 * time.Minute,
			PredictiveSensitivity:    0.7,
			TargetUtilization:        70.0,
			ScaleUpStrategy:          ScalingStrategyConservative,
			ScaleDownStrategy:        ScalingStrategyConservative,
		},
		NotificationChannels: []*NotificationChannel{
			{
				ID:       "default-log",
				Type:     NotificationTypeWebhook,
				Endpoint: "/dev/stdout",
				Enabled:  true,
				Filters:  []AlertLevel{AlertLevelWarning, AlertLevelCritical, AlertLevelEmergency},
			},
		},
		EscalationPolicies: []*EscalationPolicy{
			{
				ID:         "default-escalation",
				AlertTypes: []string{"cpu", "memory", "performance"},
				Levels:     []AlertLevel{AlertLevelCritical, AlertLevelEmergency},
				EscalationSteps: []*EscalationStep{
					{
						StepNumber:            1,
						DelayDuration:         2 * time.Minute,
						NotificationChannels:  []string{"default-log"},
						RequireAcknowledgment: false,
					},
					{
						StepNumber:            2,
						DelayDuration:         5 * time.Minute,
						NotificationChannels:  []string{"default-log"},
						RequireAcknowledgment: true,
						Actions: []*EscalationAction{
							{
								Type:       EscalationActionScale,
								Parameters: map[string]interface{}{"increment": 1},
							},
						},
					},
				},
				TimeoutDuration: 30 * time.Minute,
				MaxEscalations:  3,
				Enabled:         true,
			},
		},
	}
}

// NewResourceAlertingScalingManager creates a new resource alerting and scaling manager
func NewResourceAlertingScalingManager(config *AlertingScalingConfig) *ResourceAlertingScalingManager {
	if config == nil {
		config = DefaultAlertingScalingConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	rasm := &ResourceAlertingScalingManager{
		config:          config,
		alertEngine:     NewAlertEngine(config),
		scalingEngine:   NewAutoScalingEngine(config),
		metricCollector: NewEnhancedMetricCollector(),
		escalationEngine: NewEscalationEngine(config),
		notificationMgr: NewNotificationManager(config.NotificationChannels),
		ctx:             ctx,
		cancel:          cancel,
		alertingDone:    make(chan struct{}),
		scalingDone:     make(chan struct{}),
	}

	// Start background processes
	go rasm.startAlerting()
	go rasm.startScaling()

	return rasm
}

// NewAlertEngine creates a new alert engine
func NewAlertEngine(config *AlertingScalingConfig) *AlertEngine {
	return &AlertEngine{
		config:          config,
		thresholds:      config.AlertThresholds,
		alertHistory:    make([]*EnhancedAlert, 0),
		activeAlerts:    make(map[string]*EnhancedAlert),
		adaptiveMetrics: &AdaptiveMetrics{
			BaselineMetrics:    make(map[string]float64),
			MovingAverages:     make(map[string]float64),
			StandardDeviations: make(map[string]float64),
			AdaptedThresholds:  make(map[string]float64),
			LastUpdated:        time.Now(),
		},
	}
}

// NewAutoScalingEngine creates a new auto-scaling engine
func NewAutoScalingEngine(config *AlertingScalingConfig) *AutoScalingEngine {
	return &AutoScalingEngine{
		config:           config,
		policies:         config.ScalingPolicies,
		scalingHistory:   make([]*ScalingEvent, 0),
		currentInstances: config.ScalingPolicies.MinInstances,
		lastScalingTime:  time.Now(),
		predictiveModel: &PredictiveModel{
			MetricWindows: make(map[string][]float64),
			Trends:        make(map[string]float64),
			Seasonality:   make(map[string][]float64),
			Predictions:   make(map[string]float64),
			Confidence:    make(map[string]float64),
			LastTrainingTime: time.Now(),
		},
	}
}

// NewEnhancedMetricCollector creates a new enhanced metric collector
func NewEnhancedMetricCollector() *EnhancedMetricCollector {
	return &EnhancedMetricCollector{
		metrics: &EnhancedMetrics{
			CPUUsagePerCore: make([]float64, runtime.NumCPU()),
			LoadAverage:     make([]float64, 3),
			CustomMetrics:   make(map[string]float64),
		},
		metricHistory: make([]*MetricSnapshot, 0),
	}
}

// NewEscalationEngine creates a new escalation engine
func NewEscalationEngine(config *AlertingScalingConfig) *EscalationEngine {
	return &EscalationEngine{
		config:            config,
		policies:          config.EscalationPolicies,
		activeEscalations: make(map[string]*ActiveEscalation),
		escalationHistory: make([]*EscalationEvent, 0),
	}
}

// NewNotificationManager creates a new notification manager
func NewNotificationManager(channels []*NotificationChannel) *NotificationManager {
	return &NotificationManager{
		channels:            channels,
		notificationHistory: make([]*NotificationEvent, 0),
		rateLimiters:        make(map[string]*NotificationRateLimiter),
	}
}

// startAlerting starts the alerting process
func (rasm *ResourceAlertingScalingManager) startAlerting() {
	defer close(rasm.alertingDone)

	ticker := time.NewTicker(rasm.config.AlertingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-rasm.ctx.Done():
			return
		case <-ticker.C:
			if err := rasm.checkAlerts(); err != nil {
				log.Printf("Alert check error: %v", err)
			}
		}
	}
}

// startScaling starts the auto-scaling process
func (rasm *ResourceAlertingScalingManager) startScaling() {
	defer close(rasm.scalingDone)

	ticker := time.NewTicker(rasm.config.ScalingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-rasm.ctx.Done():
			return
		case <-ticker.C:
			if err := rasm.checkScaling(); err != nil {
				log.Printf("Scaling check error: %v", err)
			}
		}
	}
}

// checkAlerts checks for alert conditions and triggers alerts
func (rasm *ResourceAlertingScalingManager) checkAlerts() error {
	// Collect current metrics
	metrics, err := rasm.metricCollector.CollectMetrics()
	if err != nil {
		return fmt.Errorf("failed to collect metrics: %w", err)
	}

	// Check each alert condition
	alerts := rasm.alertEngine.CheckAlerts(metrics)

	// Process new alerts
	for _, alert := range alerts {
		rasm.processAlert(alert)
	}

	// Update adaptive thresholds if enabled
	if rasm.config.AlertThresholds.AdaptiveEnabled {
		rasm.alertEngine.UpdateAdaptiveThresholds(metrics)
	}

	return nil
}

// checkScaling checks for scaling conditions and triggers scaling
func (rasm *ResourceAlertingScalingManager) checkScaling() error {
	// Collect current metrics
	metrics, err := rasm.metricCollector.CollectMetrics()
	if err != nil {
		return fmt.Errorf("failed to collect metrics: %w", err)
	}

	// Check scaling conditions
	scalingAction := rasm.scalingEngine.CheckScaling(metrics)
	if scalingAction != nil {
		return rasm.executeScaling(scalingAction)
	}

	return nil
}

// processAlert processes a new alert
func (rasm *ResourceAlertingScalingManager) processAlert(alert *EnhancedAlert) {
	// Add to alert engine
	rasm.alertEngine.AddAlert(alert)

	// Send notifications
	rasm.notificationMgr.SendNotifications(alert)

	// Check for escalation
	rasm.escalationEngine.CheckEscalation(alert)

	log.Printf("Alert processed: %s - %s: %s", alert.Level, alert.Type, alert.Description)
}

// executeScaling executes a scaling action
func (rasm *ResourceAlertingScalingManager) executeScaling(event *ScalingEvent) error {
	startTime := time.Now()

	// Execute the scaling operation (placeholder - in real implementation this would
	// interface with container orchestrator, cloud provider, etc.)
	err := rasm.performScalingOperation(event)
	
	event.Duration = time.Since(startTime)
	event.Success = (err == nil)
	if err != nil {
		event.Error = err.Error()
	}

	// Record the scaling event
	rasm.scalingEngine.AddScalingEvent(event)

	if err != nil {
		log.Printf("Scaling operation failed: %v", err)
		return err
	}

	log.Printf("Scaling operation completed: %s from %d to %d instances", 
		event.Type, event.InstancesBefore, event.InstancesAfter)
	return nil
}

// performScalingOperation performs the actual scaling operation
func (rasm *ResourceAlertingScalingManager) performScalingOperation(event *ScalingEvent) error {
	// Placeholder for actual scaling implementation
	// In a real system, this would integrate with:
	// - Kubernetes API for pod scaling
	// - Cloud provider APIs for instance scaling
	// - Load balancer configuration
	// - Service mesh updates
	
	// For now, we'll simulate the scaling operation
	log.Printf("Simulating scaling operation: %s", event.Type)
	
	// Update current instance count
	rasm.scalingEngine.mu.Lock()
	rasm.scalingEngine.currentInstances = event.InstancesAfter
	rasm.scalingEngine.lastScalingTime = time.Now()
	rasm.scalingEngine.mu.Unlock()
	
	return nil
}

// CollectMetrics collects comprehensive system metrics
func (emc *EnhancedMetricCollector) CollectMetrics() (*EnhancedMetrics, error) {
	emc.mu.Lock()
	defer emc.mu.Unlock()

	// Get current process
	pid := int32(os.Getpid())
	proc, err := process.NewProcess(pid)
	if err != nil {
		return nil, fmt.Errorf("failed to get process: %w", err)
	}

	// Collect CPU metrics
	cpuPercent, err := cpu.Percent(time.Second, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU metrics: %w", err)
	}

	// Collect memory metrics
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory metrics: %w", err)
	}

	// Collect process memory
	if _, err := proc.MemoryInfo(); err != nil {
		return nil, fmt.Errorf("failed to get process memory: %w", err)
	}

	// Collect runtime metrics
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)

	// Update metrics
	emc.metrics.Timestamp = time.Now()
	
	// CPU metrics
	if len(cpuPercent) > 0 {
		emc.metrics.CPUUsage = cpuPercent[0]
		if len(cpuPercent) > 1 {
			copy(emc.metrics.CPUUsagePerCore, cpuPercent)
		}
	}

	// Memory metrics
	emc.metrics.MemoryUsage = memInfo.UsedPercent
	emc.metrics.MemoryTotal = memInfo.Total
	emc.metrics.MemoryUsed = memInfo.Used
	emc.metrics.MemoryAvailable = memInfo.Available
	emc.metrics.MemoryBuffered = memInfo.Buffers
	emc.metrics.MemoryCached = memInfo.Cached

	// Go runtime metrics
	emc.metrics.HeapAlloc = rtm.Alloc
	emc.metrics.HeapSys = rtm.HeapSys
	emc.metrics.HeapInuse = rtm.HeapInuse
	emc.metrics.HeapIdle = rtm.HeapIdle
	emc.metrics.HeapReleased = rtm.HeapReleased
	emc.metrics.GoroutineCount = runtime.NumGoroutine()
	emc.metrics.GCCycles = rtm.NumGC
	emc.metrics.GCCPUFraction = rtm.GCCPUFraction

	// Calculate GC pause time
	if rtm.NumGC > 0 {
		emc.metrics.GCPause = time.Duration(rtm.PauseNs[(rtm.NumGC+255)%256])
	}

	// Store snapshot
	snapshot := &MetricSnapshot{
		Timestamp: emc.metrics.Timestamp,
		Metrics:   emc.copyMetrics(),
	}
	emc.metricHistory = append(emc.metricHistory, snapshot)

	// Trim history if needed
	maxHistory := 1000 // Keep last 1000 snapshots
	if len(emc.metricHistory) > maxHistory {
		emc.metricHistory = emc.metricHistory[len(emc.metricHistory)-maxHistory:]
	}

	return emc.copyMetrics(), nil
}

// copyMetrics creates a copy of current metrics
func (emc *EnhancedMetricCollector) copyMetrics() *EnhancedMetrics {
	metrics := *emc.metrics
	
	// Deep copy slices and maps
	metrics.CPUUsagePerCore = make([]float64, len(emc.metrics.CPUUsagePerCore))
	copy(metrics.CPUUsagePerCore, emc.metrics.CPUUsagePerCore)
	
	metrics.LoadAverage = make([]float64, len(emc.metrics.LoadAverage))
	copy(metrics.LoadAverage, emc.metrics.LoadAverage)
	
	metrics.CustomMetrics = make(map[string]float64)
	for k, v := range emc.metrics.CustomMetrics {
		metrics.CustomMetrics[k] = v
	}
	
	return &metrics
}

// CheckAlerts checks current metrics against thresholds and returns alerts
func (ae *AlertEngine) CheckAlerts(metrics *EnhancedMetrics) []*EnhancedAlert {
	ae.mu.Lock()
	defer ae.mu.Unlock()

	var alerts []*EnhancedAlert

	// Check CPU alerts
	alerts = append(alerts, ae.checkCPUAlerts(metrics)...)
	
	// Check memory alerts
	alerts = append(alerts, ae.checkMemoryAlerts(metrics)...)
	
	// Check goroutine alerts
	alerts = append(alerts, ae.checkGoroutineAlerts(metrics)...)
	
	// Check performance alerts
	alerts = append(alerts, ae.checkPerformanceAlerts(metrics)...)

	return alerts
}

// checkCPUAlerts checks CPU-related alert conditions
func (ae *AlertEngine) checkCPUAlerts(metrics *EnhancedMetrics) []*EnhancedAlert {
	var alerts []*EnhancedAlert

	// CPU usage alerts
	if metrics.CPUUsage >= ae.thresholds.CPUEmergency {
		alerts = append(alerts, ae.createAlert(
			AlertTypeCPU, AlertLevelEmergency, "CPU Usage Emergency",
			fmt.Sprintf("CPU usage is critically high: %.2f%%", metrics.CPUUsage),
			"cpu_usage", metrics.CPUUsage, ae.thresholds.CPUEmergency,
		))
	} else if metrics.CPUUsage >= ae.thresholds.CPUCritical {
		alerts = append(alerts, ae.createAlert(
			AlertTypeCPU, AlertLevelCritical, "CPU Usage Critical",
			fmt.Sprintf("CPU usage is high: %.2f%%", metrics.CPUUsage),
			"cpu_usage", metrics.CPUUsage, ae.thresholds.CPUCritical,
		))
	} else if metrics.CPUUsage >= ae.thresholds.CPUWarning {
		alerts = append(alerts, ae.createAlert(
			AlertTypeCPU, AlertLevelWarning, "CPU Usage Warning",
			fmt.Sprintf("CPU usage is elevated: %.2f%%", metrics.CPUUsage),
			"cpu_usage", metrics.CPUUsage, ae.thresholds.CPUWarning,
		))
	}

	// Load average alerts
	if len(metrics.LoadAverage) > 0 {
		loadAvg := metrics.LoadAverage[0]
		if loadAvg >= ae.thresholds.CPULoadAvgCritical {
			alerts = append(alerts, ae.createAlert(
				AlertTypeCPU, AlertLevelCritical, "Load Average Critical",
				fmt.Sprintf("System load average is high: %.2f", loadAvg),
				"load_average", loadAvg, ae.thresholds.CPULoadAvgCritical,
			))
		} else if loadAvg >= ae.thresholds.CPULoadAvgWarning {
			alerts = append(alerts, ae.createAlert(
				AlertTypeCPU, AlertLevelWarning, "Load Average Warning",
				fmt.Sprintf("System load average is elevated: %.2f", loadAvg),
				"load_average", loadAvg, ae.thresholds.CPULoadAvgWarning,
			))
		}
	}

	return alerts
}

// checkMemoryAlerts checks memory-related alert conditions
func (ae *AlertEngine) checkMemoryAlerts(metrics *EnhancedMetrics) []*EnhancedAlert {
	var alerts []*EnhancedAlert

	// Memory usage alerts
	if metrics.MemoryUsage >= ae.thresholds.MemoryEmergency {
		alerts = append(alerts, ae.createAlert(
			AlertTypeMemory, AlertLevelEmergency, "Memory Usage Emergency",
			fmt.Sprintf("Memory usage is critically high: %.2f%%", metrics.MemoryUsage),
			"memory_usage", metrics.MemoryUsage, ae.thresholds.MemoryEmergency,
		))
	} else if metrics.MemoryUsage >= ae.thresholds.MemoryCritical {
		alerts = append(alerts, ae.createAlert(
			AlertTypeMemory, AlertLevelCritical, "Memory Usage Critical",
			fmt.Sprintf("Memory usage is high: %.2f%%", metrics.MemoryUsage),
			"memory_usage", metrics.MemoryUsage, ae.thresholds.MemoryCritical,
		))
	} else if metrics.MemoryUsage >= ae.thresholds.MemoryWarning {
		alerts = append(alerts, ae.createAlert(
			AlertTypeMemory, AlertLevelWarning, "Memory Usage Warning",
			fmt.Sprintf("Memory usage is elevated: %.2f%%", metrics.MemoryUsage),
			"memory_usage", metrics.MemoryUsage, ae.thresholds.MemoryWarning,
		))
	}

	// Heap usage alerts
	heapUsage := float64(metrics.HeapInuse) / float64(metrics.HeapSys) * 100
	if heapUsage >= ae.thresholds.HeapCritical {
		alerts = append(alerts, ae.createAlert(
			AlertTypeMemory, AlertLevelCritical, "Heap Usage Critical",
			fmt.Sprintf("Heap usage is high: %.2f%%", heapUsage),
			"heap_usage", heapUsage, ae.thresholds.HeapCritical,
		))
	} else if heapUsage >= ae.thresholds.HeapWarning {
		alerts = append(alerts, ae.createAlert(
			AlertTypeMemory, AlertLevelWarning, "Heap Usage Warning",
			fmt.Sprintf("Heap usage is elevated: %.2f%%", heapUsage),
			"heap_usage", heapUsage, ae.thresholds.HeapWarning,
		))
	}

	return alerts
}

// checkGoroutineAlerts checks goroutine-related alert conditions
func (ae *AlertEngine) checkGoroutineAlerts(metrics *EnhancedMetrics) []*EnhancedAlert {
	var alerts []*EnhancedAlert

	goroutineCount := float64(metrics.GoroutineCount)
	
	if metrics.GoroutineCount >= ae.thresholds.GoroutineEmergency {
		alerts = append(alerts, ae.createAlert(
			AlertTypeGoroutine, AlertLevelEmergency, "Goroutine Count Emergency",
			fmt.Sprintf("Goroutine count is critically high: %d", metrics.GoroutineCount),
			"goroutine_count", goroutineCount, float64(ae.thresholds.GoroutineEmergency),
		))
	} else if metrics.GoroutineCount >= ae.thresholds.GoroutineCritical {
		alerts = append(alerts, ae.createAlert(
			AlertTypeGoroutine, AlertLevelCritical, "Goroutine Count Critical",
			fmt.Sprintf("Goroutine count is high: %d", metrics.GoroutineCount),
			"goroutine_count", goroutineCount, float64(ae.thresholds.GoroutineCritical),
		))
	} else if metrics.GoroutineCount >= ae.thresholds.GoroutineWarning {
		alerts = append(alerts, ae.createAlert(
			AlertTypeGoroutine, AlertLevelWarning, "Goroutine Count Warning",
			fmt.Sprintf("Goroutine count is elevated: %d", metrics.GoroutineCount),
			"goroutine_count", goroutineCount, float64(ae.thresholds.GoroutineWarning),
		))
	}

	return alerts
}

// checkPerformanceAlerts checks performance-related alert conditions
func (ae *AlertEngine) checkPerformanceAlerts(metrics *EnhancedMetrics) []*EnhancedAlert {
	var alerts []*EnhancedAlert

	// Response time alerts
	if metrics.ResponseTime >= ae.thresholds.ResponseTimeCritical {
		alerts = append(alerts, ae.createAlert(
			AlertTypePerformance, AlertLevelCritical, "Response Time Critical",
			fmt.Sprintf("Response time is high: %v", metrics.ResponseTime),
			"response_time", float64(metrics.ResponseTime.Milliseconds()), float64(ae.thresholds.ResponseTimeCritical.Milliseconds()),
		))
	} else if metrics.ResponseTime >= ae.thresholds.ResponseTimeWarning {
		alerts = append(alerts, ae.createAlert(
			AlertTypePerformance, AlertLevelWarning, "Response Time Warning",
			fmt.Sprintf("Response time is elevated: %v", metrics.ResponseTime),
			"response_time", float64(metrics.ResponseTime.Milliseconds()), float64(ae.thresholds.ResponseTimeWarning.Milliseconds()),
		))
	}

	// Throughput alerts
	if metrics.Throughput <= ae.thresholds.ThroughputCritical {
		alerts = append(alerts, ae.createAlert(
			AlertTypePerformance, AlertLevelCritical, "Throughput Critical",
			fmt.Sprintf("Throughput is low: %.2f RPS", metrics.Throughput),
			"throughput", metrics.Throughput, ae.thresholds.ThroughputCritical,
		))
	} else if metrics.Throughput <= ae.thresholds.ThroughputWarning {
		alerts = append(alerts, ae.createAlert(
			AlertTypePerformance, AlertLevelWarning, "Throughput Warning",
			fmt.Sprintf("Throughput is low: %.2f RPS", metrics.Throughput),
			"throughput", metrics.Throughput, ae.thresholds.ThroughputWarning,
		))
	}

	// Error rate alerts
	if metrics.ErrorRate >= ae.thresholds.ErrorRateCritical {
		alerts = append(alerts, ae.createAlert(
			AlertTypePerformance, AlertLevelCritical, "Error Rate Critical",
			fmt.Sprintf("Error rate is high: %.2f%%", metrics.ErrorRate),
			"error_rate", metrics.ErrorRate, ae.thresholds.ErrorRateCritical,
		))
	} else if metrics.ErrorRate >= ae.thresholds.ErrorRateWarning {
		alerts = append(alerts, ae.createAlert(
			AlertTypePerformance, AlertLevelWarning, "Error Rate Warning",
			fmt.Sprintf("Error rate is elevated: %.2f%%", metrics.ErrorRate),
			"error_rate", metrics.ErrorRate, ae.thresholds.ErrorRateWarning,
		))
	}

	return alerts
}

// createAlert creates a new enhanced alert
func (ae *AlertEngine) createAlert(alertType AlertType, level AlertLevel, title, description, metric string, currentValue, thresholdValue float64) *EnhancedAlert {
	alert := &EnhancedAlert{
		ID:             fmt.Sprintf("%s-%s-%d", alertType, level, time.Now().UnixNano()),
		Type:           alertType,
		Level:          level,
		Title:          title,
		Description:    description,
		Resource:       "system",
		Metric:         metric,
		CurrentValue:   currentValue,
		ThresholdValue: thresholdValue,
		Severity:       ae.levelToSeverity(level),
		Tags:           make(map[string]string),
		Metadata:       make(map[string]interface{}),
		Timestamp:      time.Now(),
		Suppressed:     false,
	}

	// Add contextual tags
	alert.Tags["source"] = "resource_alerting_scaling"
	alert.Tags["host"] = "localhost" // In real implementation, get actual hostname

	// Add metadata
	alert.Metadata["go_version"] = runtime.Version()
	alert.Metadata["num_cpu"] = runtime.NumCPU()
	alert.Metadata["num_goroutine"] = runtime.NumGoroutine()

	return alert
}

// levelToSeverity converts alert level to severity
func (ae *AlertEngine) levelToSeverity(level AlertLevel) AlertSeverity {
	switch level {
	case AlertLevelEmergency:
		return AlertSeverityCritical
	case AlertLevelCritical:
		return AlertSeverityHigh
	case AlertLevelWarning:
		return AlertSeverityMedium
	default:
		return AlertSeverityLow
	}
}

// AddAlert adds an alert to the alert engine
func (ae *AlertEngine) AddAlert(alert *EnhancedAlert) {
	ae.mu.Lock()
	defer ae.mu.Unlock()

	// Add to active alerts
	ae.activeAlerts[alert.ID] = alert

	// Add to history
	ae.alertHistory = append(ae.alertHistory, alert)

	// Trim history if needed
	maxHistory := 10000 // Keep last 10,000 alerts
	if len(ae.alertHistory) > maxHistory {
		ae.alertHistory = ae.alertHistory[len(ae.alertHistory)-maxHistory:]
	}
}

// UpdateAdaptiveThresholds updates thresholds based on historical metrics
func (ae *AlertEngine) UpdateAdaptiveThresholds(metrics *EnhancedMetrics) {
	ae.mu.Lock()
	defer ae.mu.Unlock()

	// Simple adaptive threshold logic
	// In a real implementation, this would use more sophisticated algorithms
	
	// Update baseline metrics
	ae.adaptiveMetrics.BaselineMetrics["cpu_usage"] = ae.updateBaseline(ae.adaptiveMetrics.BaselineMetrics["cpu_usage"], metrics.CPUUsage)
	ae.adaptiveMetrics.BaselineMetrics["memory_usage"] = ae.updateBaseline(ae.adaptiveMetrics.BaselineMetrics["memory_usage"], metrics.MemoryUsage)
	
	// Calculate moving averages
	ae.adaptiveMetrics.MovingAverages["cpu_usage"] = ae.updateMovingAverage(ae.adaptiveMetrics.MovingAverages["cpu_usage"], metrics.CPUUsage)
	ae.adaptiveMetrics.MovingAverages["memory_usage"] = ae.updateMovingAverage(ae.adaptiveMetrics.MovingAverages["memory_usage"], metrics.MemoryUsage)
	
	// Adjust thresholds based on moving averages
	sensitivity := ae.config.AlertThresholds.AdaptiveSensitivity
	
	if cpuAvg := ae.adaptiveMetrics.MovingAverages["cpu_usage"]; cpuAvg > 0 {
		adaptedThreshold := cpuAvg * (1 + sensitivity)
		ae.adaptiveMetrics.AdaptedThresholds["cpu_warning"] = math.Min(adaptedThreshold, ae.thresholds.CPUWarning)
	}
	
	if memAvg := ae.adaptiveMetrics.MovingAverages["memory_usage"]; memAvg > 0 {
		adaptedThreshold := memAvg * (1 + sensitivity)
		ae.adaptiveMetrics.AdaptedThresholds["memory_warning"] = math.Min(adaptedThreshold, ae.thresholds.MemoryWarning)
	}
	
	ae.adaptiveMetrics.LastUpdated = time.Now()
}

// updateBaseline updates a baseline metric using exponential smoothing
func (ae *AlertEngine) updateBaseline(baseline, newValue float64) float64 {
	if baseline == 0 {
		return newValue
	}
	alpha := 0.1 // Smoothing factor
	return alpha*newValue + (1-alpha)*baseline
}

// updateMovingAverage updates a moving average
func (ae *AlertEngine) updateMovingAverage(average, newValue float64) float64 {
	if average == 0 {
		return newValue
	}
	alpha := 0.2 // Moving average factor
	return alpha*newValue + (1-alpha)*average
}

// CheckScaling checks if scaling is needed and returns a scaling event
func (ase *AutoScalingEngine) CheckScaling(metrics *EnhancedMetrics) *ScalingEvent {
	ase.mu.Lock()
	defer ase.mu.Unlock()

	// Check cooldown period
	if time.Since(ase.lastScalingTime) < ase.config.ScalingCooldownPeriod {
		return nil
	}

	// Check scale up conditions
	if ase.shouldScaleUp(metrics) {
		return ase.createScaleUpEvent(metrics)
	}

	// Check scale down conditions
	if ase.shouldScaleDown(metrics) {
		return ase.createScaleDownEvent(metrics)
	}

	return nil
}

// shouldScaleUp determines if scaling up is needed
func (ase *AutoScalingEngine) shouldScaleUp(metrics *EnhancedMetrics) bool {
	// Check if already at max instances
	if ase.currentInstances >= ase.policies.MaxInstances {
		return false
	}

	// Check CPU threshold
	if metrics.CPUUsage >= ase.policies.CPUScaleUpThreshold {
		return true
	}

	// Check memory threshold
	if metrics.MemoryUsage >= ase.policies.MemoryScaleUpThreshold {
		return true
	}

	// Check predictive scaling if enabled
	if ase.policies.PredictiveScalingEnabled {
		return ase.predictiveScaleUp(metrics)
	}

	return false
}

// shouldScaleDown determines if scaling down is needed
func (ase *AutoScalingEngine) shouldScaleDown(metrics *EnhancedMetrics) bool {
	// Check if already at min instances
	if ase.currentInstances <= ase.policies.MinInstances {
		return false
	}

	// Check CPU threshold
	if metrics.CPUUsage <= ase.policies.CPUScaleDownThreshold {
		return true
	}

	// Check memory threshold
	if metrics.MemoryUsage <= ase.policies.MemoryScaleDownThreshold {
		return true
	}

	return false
}

// predictiveScaleUp uses predictive modeling to determine if scale up is needed
func (ase *AutoScalingEngine) predictiveScaleUp(metrics *EnhancedMetrics) bool {
	// Simple predictive logic - in real implementation would use ML models
	
	// Update metric windows
	ase.updateMetricWindow("cpu_usage", metrics.CPUUsage)
	ase.updateMetricWindow("memory_usage", metrics.MemoryUsage)
	
	// Calculate trends
	cpuTrend := ase.calculateTrend("cpu_usage")
	memoryTrend := ase.calculateTrend("memory_usage")
	
	// Predict future values
	cpuPredicted := metrics.CPUUsage + cpuTrend*5 // 5 minutes ahead
	memoryPredicted := metrics.MemoryUsage + memoryTrend*5
	
	// Check if predicted values exceed thresholds
	return cpuPredicted >= ase.policies.CPUScaleUpThreshold || memoryPredicted >= ase.policies.MemoryScaleUpThreshold
}

// updateMetricWindow updates the metric window for predictive analysis
func (ase *AutoScalingEngine) updateMetricWindow(metric string, value float64) {
	window := ase.predictiveModel.MetricWindows[metric]
	window = append(window, value)
	
	// Keep only last N values (e.g., 30 for 30 data points)
	maxWindow := 30
	if len(window) > maxWindow {
		window = window[len(window)-maxWindow:]
	}
	
	ase.predictiveModel.MetricWindows[metric] = window
}

// calculateTrend calculates the trend for a metric
func (ase *AutoScalingEngine) calculateTrend(metric string) float64 {
	window := ase.predictiveModel.MetricWindows[metric]
	if len(window) < 2 {
		return 0
	}
	
	// Simple linear regression slope
	n := len(window)
	sumX, sumY, sumXY, sumX2 := 0.0, 0.0, 0.0, 0.0
	
	for i, y := range window {
		x := float64(i)
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}
	
	slope := (float64(n)*sumXY - sumX*sumY) / (float64(n)*sumX2 - sumX*sumX)
	ase.predictiveModel.Trends[metric] = slope
	
	return slope
}

// createScaleUpEvent creates a scale up event
func (ase *AutoScalingEngine) createScaleUpEvent(metrics *EnhancedMetrics) *ScalingEvent {
	instancesBefore := ase.currentInstances
	instancesAfter := instancesBefore + ase.policies.ScaleUpIncrement
	
	if instancesAfter > ase.policies.MaxInstances {
		instancesAfter = ase.policies.MaxInstances
	}

	return &ScalingEvent{
		ID:              fmt.Sprintf("scale-up-%d", time.Now().UnixNano()),
		Type:            ScalingTypeUp,
		Trigger:         ScalingTriggerThreshold,
		Strategy:        ase.policies.ScaleUpStrategy,
		InstancesBefore: instancesBefore,
		InstancesAfter:  instancesAfter,
		Reason:          fmt.Sprintf("CPU: %.2f%%, Memory: %.2f%%", metrics.CPUUsage, metrics.MemoryUsage),
		Metadata: map[string]interface{}{
			"cpu_usage":    metrics.CPUUsage,
			"memory_usage": metrics.MemoryUsage,
			"cpu_threshold": ase.policies.CPUScaleUpThreshold,
			"memory_threshold": ase.policies.MemoryScaleUpThreshold,
		},
		Timestamp: time.Now(),
	}
}

// createScaleDownEvent creates a scale down event
func (ase *AutoScalingEngine) createScaleDownEvent(metrics *EnhancedMetrics) *ScalingEvent {
	instancesBefore := ase.currentInstances
	instancesAfter := instancesBefore - ase.policies.ScaleDownDecrement
	
	if instancesAfter < ase.policies.MinInstances {
		instancesAfter = ase.policies.MinInstances
	}

	return &ScalingEvent{
		ID:              fmt.Sprintf("scale-down-%d", time.Now().UnixNano()),
		Type:            ScalingTypeDown,
		Trigger:         ScalingTriggerThreshold,
		Strategy:        ase.policies.ScaleDownStrategy,
		InstancesBefore: instancesBefore,
		InstancesAfter:  instancesAfter,
		Reason:          fmt.Sprintf("CPU: %.2f%%, Memory: %.2f%%", metrics.CPUUsage, metrics.MemoryUsage),
		Metadata: map[string]interface{}{
			"cpu_usage":    metrics.CPUUsage,
			"memory_usage": metrics.MemoryUsage,
			"cpu_threshold": ase.policies.CPUScaleDownThreshold,
			"memory_threshold": ase.policies.MemoryScaleDownThreshold,
		},
		Timestamp: time.Now(),
	}
}

// AddScalingEvent adds a scaling event to the history
func (ase *AutoScalingEngine) AddScalingEvent(event *ScalingEvent) {
	ase.mu.Lock()
	defer ase.mu.Unlock()

	ase.scalingHistory = append(ase.scalingHistory, event)

	// Trim history if needed
	maxHistory := 1000 // Keep last 1000 scaling events
	if len(ase.scalingHistory) > maxHistory {
		ase.scalingHistory = ase.scalingHistory[len(ase.scalingHistory)-maxHistory:]
	}
}

// CheckEscalation checks if an alert should be escalated
func (ee *EscalationEngine) CheckEscalation(alert *EnhancedAlert) {
	ee.mu.Lock()
	defer ee.mu.Unlock()

	// Find matching escalation policies
	for _, policy := range ee.policies {
		if ee.alertMatchesPolicy(alert, policy) {
			ee.startEscalation(alert, policy)
		}
	}
}

// alertMatchesPolicy checks if an alert matches an escalation policy
func (ee *EscalationEngine) alertMatchesPolicy(alert *EnhancedAlert, policy *EscalationPolicy) bool {
	if !policy.Enabled {
		return false
	}

	// Check alert type
	for _, alertType := range policy.AlertTypes {
		if string(alert.Type) == alertType {
			// Check alert level
			for _, level := range policy.Levels {
				if alert.Level == level {
					return true
				}
			}
		}
	}

	return false
}

// startEscalation starts an escalation for an alert
func (ee *EscalationEngine) startEscalation(alert *EnhancedAlert, policy *EscalationPolicy) {
	escalationID := fmt.Sprintf("esc-%s-%d", alert.ID, time.Now().UnixNano())

	escalation := &ActiveEscalation{
		ID:              escalationID,
		AlertID:         alert.ID,
		PolicyID:        policy.ID,
		CurrentStep:     0,
		StartedAt:       time.Now(),
		LastEscalatedAt: time.Now(),
		EscalationSteps: make([]*CompletedStep, 0),
	}

	ee.activeEscalations[escalationID] = escalation

	// Record escalation event
	event := &EscalationEvent{
		ID:           fmt.Sprintf("esc-event-%d", time.Now().UnixNano()),
		Type:         EscalationEventStarted,
		AlertID:      alert.ID,
		EscalationID: escalationID,
		PolicyID:     policy.ID,
		StepNumber:   0,
		Details:      "Escalation started",
		Metadata:     make(map[string]interface{}),
		Timestamp:    time.Now(),
	}

	ee.escalationHistory = append(ee.escalationHistory, event)

	log.Printf("Escalation started: %s for alert %s", escalationID, alert.ID)
}

// SendNotifications sends notifications for an alert
func (nm *NotificationManager) SendNotifications(alert *EnhancedAlert) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	for _, channel := range nm.channels {
		if nm.shouldSendNotification(channel, alert) {
			nm.sendNotification(channel, alert)
		}
	}
}

// shouldSendNotification determines if a notification should be sent
func (nm *NotificationManager) shouldSendNotification(channel *NotificationChannel, alert *EnhancedAlert) bool {
	if !channel.Enabled {
		return false
	}

	// Check filters
	if len(channel.Filters) > 0 {
		for _, filter := range channel.Filters {
			if alert.Level == filter {
				return true
			}
		}
		return false
	}

	return true
}

// sendNotification sends a notification
func (nm *NotificationManager) sendNotification(channel *NotificationChannel, alert *EnhancedAlert) {
	notificationID := fmt.Sprintf("notif-%s-%d", alert.ID, time.Now().UnixNano())

	notification := &NotificationEvent{
		ID:        notificationID,
		ChannelID: channel.ID,
		AlertID:   alert.ID,
		Type:      channel.Type,
		Recipient: channel.Endpoint,
		Subject:   fmt.Sprintf("Alert: %s", alert.Title),
		Message:   nm.formatMessage(alert),
		Status:    NotificationStatusPending,
		Metadata:  make(map[string]interface{}),
		SentAt:    time.Now(),
	}

	// Simulate sending notification
	err := nm.deliverNotification(channel, notification)
	if err != nil {
		notification.Status = NotificationStatusFailed
		notification.Error = err.Error()
	} else {
		notification.Status = NotificationStatusDelivered
		now := time.Now()
		notification.DeliveredAt = &now
	}

	nm.notificationHistory = append(nm.notificationHistory, notification)

	log.Printf("Notification sent: %s via %s", notification.Subject, channel.Type)
}

// formatMessage formats an alert message for notification
func (nm *NotificationManager) formatMessage(alert *EnhancedAlert) string {
	return fmt.Sprintf(`
Alert: %s
Level: %s
Description: %s
Metric: %s
Current Value: %.2f
Threshold: %.2f
Time: %s
`, alert.Title, alert.Level, alert.Description, alert.Metric,
		alert.CurrentValue, alert.ThresholdValue, alert.Timestamp.Format(time.RFC3339))
}

// deliverNotification delivers a notification via the specified channel
func (nm *NotificationManager) deliverNotification(channel *NotificationChannel, notification *NotificationEvent) error {
	// Placeholder for actual notification delivery
	// In a real implementation, this would integrate with:
	// - Email services (SMTP, SendGrid, etc.)
	// - Slack API
	// - SMS services (Twilio, etc.)
	// - PagerDuty API
	// - Webhook endpoints

	switch channel.Type {
	case NotificationTypeEmail:
		return nm.sendEmail(channel, notification)
	case NotificationTypeSlack:
		return nm.sendSlack(channel, notification)
	case NotificationTypeWebhook:
		return nm.sendWebhook(channel, notification)
	case NotificationTypeSMS:
		return nm.sendSMS(channel, notification)
	case NotificationTypePagerDuty:
		return nm.sendPagerDuty(channel, notification)
	default:
		return fmt.Errorf("unsupported notification type: %s", channel.Type)
	}
}

// sendEmail sends an email notification (placeholder)
func (nm *NotificationManager) sendEmail(channel *NotificationChannel, notification *NotificationEvent) error {
	log.Printf("Sending email to %s: %s", channel.Endpoint, notification.Subject)
	return nil
}

// sendSlack sends a Slack notification (placeholder)
func (nm *NotificationManager) sendSlack(channel *NotificationChannel, notification *NotificationEvent) error {
	log.Printf("Sending Slack message to %s: %s", channel.Endpoint, notification.Subject)
	return nil
}

// sendWebhook sends a webhook notification (placeholder)
func (nm *NotificationManager) sendWebhook(channel *NotificationChannel, notification *NotificationEvent) error {
	log.Printf("Sending webhook to %s: %s", channel.Endpoint, notification.Subject)
	return nil
}

// sendSMS sends an SMS notification (placeholder)
func (nm *NotificationManager) sendSMS(channel *NotificationChannel, notification *NotificationEvent) error {
	log.Printf("Sending SMS to %s: %s", channel.Endpoint, notification.Subject)
	return nil
}

// sendPagerDuty sends a PagerDuty notification (placeholder)
func (nm *NotificationManager) sendPagerDuty(channel *NotificationChannel, notification *NotificationEvent) error {
	log.Printf("Sending PagerDuty alert to %s: %s", channel.Endpoint, notification.Subject)
	return nil
}

// GetCurrentMetrics returns the current metrics
func (rasm *ResourceAlertingScalingManager) GetCurrentMetrics() (*EnhancedMetrics, error) {
	return rasm.metricCollector.CollectMetrics()
}

// GetActiveAlerts returns all active alerts
func (rasm *ResourceAlertingScalingManager) GetActiveAlerts() []*EnhancedAlert {
	rasm.alertEngine.mu.RLock()
	defer rasm.alertEngine.mu.RUnlock()

	alerts := make([]*EnhancedAlert, 0, len(rasm.alertEngine.activeAlerts))
	for _, alert := range rasm.alertEngine.activeAlerts {
		if alert.ResolvedAt == nil {
			alerts = append(alerts, alert)
		}
	}

	return alerts
}

// GetAlertHistory returns alert history
func (rasm *ResourceAlertingScalingManager) GetAlertHistory(limit int) []*EnhancedAlert {
	rasm.alertEngine.mu.RLock()
	defer rasm.alertEngine.mu.RUnlock()

	history := rasm.alertEngine.alertHistory
	if limit > 0 && len(history) > limit {
		history = history[len(history)-limit:]
	}

	// Return a copy
	result := make([]*EnhancedAlert, len(history))
	copy(result, history)
	return result
}

// GetScalingHistory returns scaling history
func (rasm *ResourceAlertingScalingManager) GetScalingHistory(limit int) []*ScalingEvent {
	rasm.scalingEngine.mu.RLock()
	defer rasm.scalingEngine.mu.RUnlock()

	history := rasm.scalingEngine.scalingHistory
	if limit > 0 && len(history) > limit {
		history = history[len(history)-limit:]
	}

	// Return a copy
	result := make([]*ScalingEvent, len(history))
	copy(result, history)
	return result
}

// GetCurrentInstances returns the current number of instances
func (rasm *ResourceAlertingScalingManager) GetCurrentInstances() int {
	rasm.scalingEngine.mu.RLock()
	defer rasm.scalingEngine.mu.RUnlock()
	return rasm.scalingEngine.currentInstances
}

// ManualScale manually triggers a scaling operation
func (rasm *ResourceAlertingScalingManager) ManualScale(targetInstances int, reason string) error {
	rasm.scalingEngine.mu.Lock()
	defer rasm.scalingEngine.mu.Unlock()

	currentInstances := rasm.scalingEngine.currentInstances

	if targetInstances < rasm.config.ScalingPolicies.MinInstances {
		return fmt.Errorf("target instances (%d) below minimum (%d)", targetInstances, rasm.config.ScalingPolicies.MinInstances)
	}

	if targetInstances > rasm.config.ScalingPolicies.MaxInstances {
		return fmt.Errorf("target instances (%d) above maximum (%d)", targetInstances, rasm.config.ScalingPolicies.MaxInstances)
	}

	if targetInstances == currentInstances {
		return fmt.Errorf("target instances (%d) same as current (%d)", targetInstances, currentInstances)
	}

	var scalingType ScalingType
	if targetInstances > currentInstances {
		scalingType = ScalingTypeUp
	} else {
		scalingType = ScalingTypeDown
	}

	event := &ScalingEvent{
		ID:              fmt.Sprintf("manual-scale-%d", time.Now().UnixNano()),
		Type:            scalingType,
		Trigger:         ScalingTriggerManual,
		Strategy:        ScalingStrategyConservative,
		InstancesBefore: currentInstances,
		InstancesAfter:  targetInstances,
		Reason:          reason,
		Metadata: map[string]interface{}{
			"manual": true,
			"user":   "system", // In real implementation, would track actual user
		},
		Timestamp: time.Now(),
	}

	return rasm.executeScaling(event)
}

// AcknowledgeAlert acknowledges an alert
func (rasm *ResourceAlertingScalingManager) AcknowledgeAlert(alertID string) error {
	rasm.alertEngine.mu.Lock()
	defer rasm.alertEngine.mu.Unlock()

	alert, exists := rasm.alertEngine.activeAlerts[alertID]
	if !exists {
		return fmt.Errorf("alert not found: %s", alertID)
	}

	now := time.Now()
	alert.AcknowledgedAt = &now

	log.Printf("Alert acknowledged: %s", alertID)
	return nil
}

// ResolveAlert resolves an alert
func (rasm *ResourceAlertingScalingManager) ResolveAlert(alertID string) error {
	rasm.alertEngine.mu.Lock()
	defer rasm.alertEngine.mu.Unlock()

	alert, exists := rasm.alertEngine.activeAlerts[alertID]
	if !exists {
		return fmt.Errorf("alert not found: %s", alertID)
	}

	now := time.Now()
	alert.ResolvedAt = &now

	// Remove from active alerts
	delete(rasm.alertEngine.activeAlerts, alertID)

	log.Printf("Alert resolved: %s", alertID)
	return nil
}

// UpdateConfig updates the configuration
func (rasm *ResourceAlertingScalingManager) UpdateConfig(config *AlertingScalingConfig) error {
	rasm.mu.Lock()
	defer rasm.mu.Unlock()

	rasm.config = config
	rasm.alertEngine.config = config
	rasm.alertEngine.thresholds = config.AlertThresholds
	rasm.scalingEngine.config = config
	rasm.scalingEngine.policies = config.ScalingPolicies

	log.Printf("Configuration updated")
	return nil
}

// GetStatus returns the current status of the alerting and scaling manager
func (rasm *ResourceAlertingScalingManager) GetStatus() map[string]interface{} {
	rasm.mu.RLock()
	defer rasm.mu.RUnlock()

	status := map[string]interface{}{
		"status":              "active",
		"alerting_enabled":    true,
		"scaling_enabled":     true,
		"active_alerts":       len(rasm.GetActiveAlerts()),
		"current_instances":   rasm.GetCurrentInstances(),
		"min_instances":       rasm.config.ScalingPolicies.MinInstances,
		"max_instances":       rasm.config.ScalingPolicies.MaxInstances,
		"last_scaling_time":   rasm.scalingEngine.lastScalingTime,
		"predictive_enabled":  rasm.config.ScalingPolicies.PredictiveScalingEnabled,
		"adaptive_enabled":    rasm.config.AlertThresholds.AdaptiveEnabled,
		"notification_channels": len(rasm.config.NotificationChannels),
		"escalation_policies": len(rasm.config.EscalationPolicies),
	}

	return status
}

// Shutdown gracefully shuts down the alerting and scaling manager
func (rasm *ResourceAlertingScalingManager) Shutdown() error {
	rasm.mu.Lock()
	defer rasm.mu.Unlock()

	log.Printf("Shutting down resource alerting and scaling manager...")

	// Cancel context to stop background processes
	rasm.cancel()

	// Wait for processes to complete
	select {
	case <-rasm.alertingDone:
		log.Printf("Alerting process stopped")
	case <-time.After(5 * time.Second):
		log.Printf("Alerting process shutdown timeout")
	}

	select {
	case <-rasm.scalingDone:
		log.Printf("Scaling process stopped")
	case <-time.After(5 * time.Second):
		log.Printf("Scaling process shutdown timeout")
	}

	log.Printf("Resource alerting and scaling manager shutdown complete")
	return nil
}
