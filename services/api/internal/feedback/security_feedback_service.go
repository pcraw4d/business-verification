package feedback

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// SecurityFeedbackService provides a high-level service for security feedback collection and improvement
type SecurityFeedbackService struct {
	config                      *SecurityFeedbackServiceConfig
	logger                      *zap.Logger
	advancedSystem              *AdvancedSecurityFeedbackSystem
	securityAnalyzer            *SecurityFeedbackAnalyzer
	websiteVerificationImprover *WebsiteVerificationImprover
	feedbackRepository          FeedbackRepository
	mu                          sync.RWMutex
	isRunning                   bool
	stopChan                    chan struct{}
}

// SecurityFeedbackServiceConfig contains configuration for the security feedback service
type SecurityFeedbackServiceConfig struct {
	// Service settings
	AutoStart           bool          `json:"auto_start"`           // true
	CollectionInterval  time.Duration `json:"collection_interval"`  // 1 hour
	AnalysisInterval    time.Duration `json:"analysis_interval"`    // 2 hours
	ImprovementInterval time.Duration `json:"improvement_interval"` // 24 hours

	// Advanced system settings
	AdvancedSecurityConfig *AdvancedSecurityConfig `json:"advanced_security_config"`

	// Monitoring settings
	HealthCheckInterval      time.Duration `json:"health_check_interval"`      // 5 minutes
	MetricsReportingInterval time.Duration `json:"metrics_reporting_interval"` // 1 hour
}

// SecurityFeedbackServiceResult represents the result of a service operation
type SecurityFeedbackServiceResult struct {
	Operation         string                             `json:"operation"`
	Success           bool                               `json:"success"`
	Message           string                             `json:"message"`
	CollectionResult  *SecurityFeedbackCollectionResult  `json:"collection_result,omitempty"`
	AnalysisResult    *SecurityFeedbackAnalysisResult    `json:"analysis_result,omitempty"`
	ImprovementResult *SecurityFeedbackImprovementResult `json:"improvement_result,omitempty"`
	ProcessingTime    time.Duration                      `json:"processing_time"`
	Timestamp         time.Time                          `json:"timestamp"`
}

// NewSecurityFeedbackService creates a new security feedback service
func NewSecurityFeedbackService(
	config *SecurityFeedbackServiceConfig,
	logger *zap.Logger,
	feedbackRepository FeedbackRepository,
) *SecurityFeedbackService {
	if config == nil {
		config = &SecurityFeedbackServiceConfig{
			AutoStart:                true,
			CollectionInterval:       1 * time.Hour,
			AnalysisInterval:         2 * time.Hour,
			ImprovementInterval:      24 * time.Hour,
			HealthCheckInterval:      5 * time.Minute,
			MetricsReportingInterval: 1 * time.Hour,
			AdvancedSecurityConfig:   nil, // Will use defaults
		}
	}

	// Create security analyzer
	securityAnalyzer := NewSecurityFeedbackAnalyzer(&MLAnalysisConfig{
		MinFeedbackThreshold: 50,
	}, logger)

	// Create website verification improver
	websiteVerificationImprover := NewWebsiteVerificationImprover(&AdvancedLearningConfig{
		MinFeedbackThreshold:          50,
		MaxLearningBatchSize:          100,
		VerificationAccuracyThreshold: 0.9,
	}, logger)

	// Create advanced security feedback system
	advancedSystem := NewAdvancedSecurityFeedbackSystem(
		config.AdvancedSecurityConfig,
		logger,
		securityAnalyzer,
		websiteVerificationImprover,
		feedbackRepository,
	)

	service := &SecurityFeedbackService{
		config:                      config,
		logger:                      logger,
		advancedSystem:              advancedSystem,
		securityAnalyzer:            securityAnalyzer,
		websiteVerificationImprover: websiteVerificationImprover,
		feedbackRepository:          feedbackRepository,
		stopChan:                    make(chan struct{}),
	}

	// Auto-start if configured
	if config.AutoStart {
		go service.Start()
	}

	return service
}

// Start starts the security feedback service
func (sfs *SecurityFeedbackService) Start() error {
	sfs.mu.Lock()
	defer sfs.mu.Unlock()

	if sfs.isRunning {
		return fmt.Errorf("security feedback service is already running")
	}

	sfs.logger.Info("Starting security feedback service")

	// Start background goroutines
	go sfs.collectionWorker()
	go sfs.analysisWorker()
	go sfs.improvementWorker()
	go sfs.healthCheckWorker()
	go sfs.metricsReportingWorker()

	sfs.isRunning = true
	sfs.logger.Info("Security feedback service started successfully")

	return nil
}

// Stop stops the security feedback service
func (sfs *SecurityFeedbackService) Stop() error {
	sfs.mu.Lock()
	defer sfs.mu.Unlock()

	if !sfs.isRunning {
		return fmt.Errorf("security feedback service is not running")
	}

	sfs.logger.Info("Stopping security feedback service")

	// Signal all workers to stop
	close(sfs.stopChan)

	sfs.isRunning = false
	sfs.logger.Info("Security feedback service stopped successfully")

	return nil
}

// IsRunning returns whether the service is currently running
func (sfs *SecurityFeedbackService) IsRunning() bool {
	sfs.mu.RLock()
	defer sfs.mu.RUnlock()
	return sfs.isRunning
}

// CollectSecurityFeedback manually triggers security feedback collection
func (sfs *SecurityFeedbackService) CollectSecurityFeedback(ctx context.Context) (*SecurityFeedbackServiceResult, error) {
	startTime := time.Now()

	sfs.logger.Info("Manually triggering security feedback collection")

	result, err := sfs.advancedSystem.CollectSecurityFeedback(ctx)
	if err != nil {
		return &SecurityFeedbackServiceResult{
			Operation:      "collect_security_feedback",
			Success:        false,
			Message:        fmt.Sprintf("Failed to collect security feedback: %v", err),
			ProcessingTime: time.Since(startTime),
			Timestamp:      time.Now(),
		}, err
	}

	return &SecurityFeedbackServiceResult{
		Operation:        "collect_security_feedback",
		Success:          true,
		Message:          "Security feedback collection completed successfully",
		CollectionResult: result,
		ProcessingTime:   time.Since(startTime),
		Timestamp:        time.Now(),
	}, nil
}

// AnalyzeSecurityFeedback manually triggers security feedback analysis
func (sfs *SecurityFeedbackService) AnalyzeSecurityFeedback(ctx context.Context) (*SecurityFeedbackServiceResult, error) {
	startTime := time.Now()

	sfs.logger.Info("Manually triggering security feedback analysis")

	// First collect feedback
	collectionResult, err := sfs.advancedSystem.CollectSecurityFeedback(ctx)
	if err != nil {
		return &SecurityFeedbackServiceResult{
			Operation:      "analyze_security_feedback",
			Success:        false,
			Message:        fmt.Sprintf("Failed to collect feedback for analysis: %v", err),
			ProcessingTime: time.Since(startTime),
			Timestamp:      time.Now(),
		}, err
	}

	// Then analyze the collected feedback
	analysisResult, err := sfs.advancedSystem.AnalyzeSecurityFeedback(ctx, collectionResult)
	if err != nil {
		return &SecurityFeedbackServiceResult{
			Operation:        "analyze_security_feedback",
			Success:          false,
			Message:          fmt.Sprintf("Failed to analyze security feedback: %v", err),
			CollectionResult: collectionResult,
			ProcessingTime:   time.Since(startTime),
			Timestamp:        time.Now(),
		}, err
	}

	return &SecurityFeedbackServiceResult{
		Operation:        "analyze_security_feedback",
		Success:          true,
		Message:          "Security feedback analysis completed successfully",
		CollectionResult: collectionResult,
		AnalysisResult:   analysisResult,
		ProcessingTime:   time.Since(startTime),
		Timestamp:        time.Now(),
	}, nil
}

// ImproveSecurityAlgorithms manually triggers security algorithm improvement
func (sfs *SecurityFeedbackService) ImproveSecurityAlgorithms(ctx context.Context) (*SecurityFeedbackServiceResult, error) {
	startTime := time.Now()

	sfs.logger.Info("Manually triggering security algorithm improvement")

	// First collect and analyze feedback
	collectionResult, err := sfs.advancedSystem.CollectSecurityFeedback(ctx)
	if err != nil {
		return &SecurityFeedbackServiceResult{
			Operation:      "improve_security_algorithms",
			Success:        false,
			Message:        fmt.Sprintf("Failed to collect feedback for improvement: %v", err),
			ProcessingTime: time.Since(startTime),
			Timestamp:      time.Now(),
		}, err
	}

	analysisResult, err := sfs.advancedSystem.AnalyzeSecurityFeedback(ctx, collectionResult)
	if err != nil {
		return &SecurityFeedbackServiceResult{
			Operation:        "improve_security_algorithms",
			Success:          false,
			Message:          fmt.Sprintf("Failed to analyze feedback for improvement: %v", err),
			CollectionResult: collectionResult,
			ProcessingTime:   time.Since(startTime),
			Timestamp:        time.Now(),
		}, err
	}

	// Then improve algorithms based on analysis
	improvementResult, err := sfs.advancedSystem.ImproveSecurityAlgorithms(ctx, analysisResult)
	if err != nil {
		return &SecurityFeedbackServiceResult{
			Operation:        "improve_security_algorithms",
			Success:          false,
			Message:          fmt.Sprintf("Failed to improve security algorithms: %v", err),
			CollectionResult: collectionResult,
			AnalysisResult:   analysisResult,
			ProcessingTime:   time.Since(startTime),
			Timestamp:        time.Now(),
		}, err
	}

	return &SecurityFeedbackServiceResult{
		Operation:         "improve_security_algorithms",
		Success:           true,
		Message:           "Security algorithm improvement completed successfully",
		CollectionResult:  collectionResult,
		AnalysisResult:    analysisResult,
		ImprovementResult: improvementResult,
		ProcessingTime:    time.Since(startTime),
		Timestamp:         time.Now(),
	}, nil
}

// GetServiceStatus returns the current status of the security feedback service
func (sfs *SecurityFeedbackService) GetServiceStatus(ctx context.Context) map[string]interface{} {
	sfs.mu.RLock()
	defer sfs.mu.RUnlock()

	status := map[string]interface{}{
		"service_running":            sfs.isRunning,
		"collection_interval":        sfs.config.CollectionInterval.String(),
		"analysis_interval":          sfs.config.AnalysisInterval.String(),
		"improvement_interval":       sfs.config.ImprovementInterval.String(),
		"health_check_interval":      sfs.config.HealthCheckInterval.String(),
		"metrics_reporting_interval": sfs.config.MetricsReportingInterval.String(),
	}

	// Get advanced system health
	if sfs.advancedSystem != nil {
		advancedHealth := sfs.advancedSystem.GetSystemHealth(ctx)
		status["advanced_system_health"] = advancedHealth
	}

	// Get security metrics
	if sfs.advancedSystem != nil {
		metrics := sfs.advancedSystem.GetSecurityMetrics()
		status["security_metrics"] = metrics
	}

	return status
}

// collectionWorker runs the security feedback collection worker
func (sfs *SecurityFeedbackService) collectionWorker() {
	ticker := time.NewTicker(sfs.config.CollectionInterval)
	defer ticker.Stop()

	sfs.logger.Info("Security feedback collection worker started",
		zap.Duration("interval", sfs.config.CollectionInterval))

	for {
		select {
		case <-sfs.stopChan:
			sfs.logger.Info("Security feedback collection worker stopped")
			return
		case <-ticker.C:
			ctx := context.Background()
			result, err := sfs.advancedSystem.CollectSecurityFeedback(ctx)
			if err != nil {
				sfs.logger.Error("Failed to collect security feedback in worker",
					zap.Error(err))
			} else {
				sfs.logger.Info("Security feedback collection completed in worker",
					zap.Int("total_processed", result.TotalProcessed),
					zap.Duration("collection_time", result.CollectionTime))
			}
		}
	}
}

// analysisWorker runs the security feedback analysis worker
func (sfs *SecurityFeedbackService) analysisWorker() {
	ticker := time.NewTicker(sfs.config.AnalysisInterval)
	defer ticker.Stop()

	sfs.logger.Info("Security feedback analysis worker started",
		zap.Duration("interval", sfs.config.AnalysisInterval))

	for {
		select {
		case <-sfs.stopChan:
			sfs.logger.Info("Security feedback analysis worker stopped")
			return
		case <-ticker.C:
			ctx := context.Background()

			// Collect feedback first
			collectionResult, err := sfs.advancedSystem.CollectSecurityFeedback(ctx)
			if err != nil {
				sfs.logger.Error("Failed to collect feedback for analysis in worker",
					zap.Error(err))
				continue
			}

			// Then analyze
			analysisResult, err := sfs.advancedSystem.AnalyzeSecurityFeedback(ctx, collectionResult)
			if err != nil {
				sfs.logger.Error("Failed to analyze security feedback in worker",
					zap.Error(err))
			} else {
				sfs.logger.Info("Security feedback analysis completed in worker",
					zap.Float64("overall_security_score", analysisResult.OverallSecurityScore),
					zap.Float64("data_quality_score", analysisResult.DataQualityScore),
					zap.Duration("analysis_time", analysisResult.AnalysisTime))
			}
		}
	}
}

// improvementWorker runs the security algorithm improvement worker
func (sfs *SecurityFeedbackService) improvementWorker() {
	ticker := time.NewTicker(sfs.config.ImprovementInterval)
	defer ticker.Stop()

	sfs.logger.Info("Security algorithm improvement worker started",
		zap.Duration("interval", sfs.config.ImprovementInterval))

	for {
		select {
		case <-sfs.stopChan:
			sfs.logger.Info("Security algorithm improvement worker stopped")
			return
		case <-ticker.C:
			ctx := context.Background()

			// Collect and analyze feedback first
			collectionResult, err := sfs.advancedSystem.CollectSecurityFeedback(ctx)
			if err != nil {
				sfs.logger.Error("Failed to collect feedback for improvement in worker",
					zap.Error(err))
				continue
			}

			analysisResult, err := sfs.advancedSystem.AnalyzeSecurityFeedback(ctx, collectionResult)
			if err != nil {
				sfs.logger.Error("Failed to analyze feedback for improvement in worker",
					zap.Error(err))
				continue
			}

			// Then improve algorithms
			improvementResult, err := sfs.advancedSystem.ImproveSecurityAlgorithms(ctx, analysisResult)
			if err != nil {
				sfs.logger.Error("Failed to improve security algorithms in worker",
					zap.Error(err))
			} else {
				sfs.logger.Info("Security algorithm improvement completed in worker",
					zap.Int("algorithms_improved", len(improvementResult.AlgorithmsImproved)),
					zap.Float64("success_rate", improvementResult.SuccessRate),
					zap.Duration("improvement_time", improvementResult.ImprovementTime))
			}
		}
	}
}

// healthCheckWorker runs the health check worker
func (sfs *SecurityFeedbackService) healthCheckWorker() {
	ticker := time.NewTicker(sfs.config.HealthCheckInterval)
	defer ticker.Stop()

	sfs.logger.Info("Security feedback service health check worker started",
		zap.Duration("interval", sfs.config.HealthCheckInterval))

	for {
		select {
		case <-sfs.stopChan:
			sfs.logger.Info("Security feedback service health check worker stopped")
			return
		case <-ticker.C:
			ctx := context.Background()
			health := sfs.GetServiceStatus(ctx)

			// Log health status
			sfs.logger.Info("Security feedback service health check",
				zap.Bool("service_running", health["service_running"].(bool)))

			// Check for any health issues
			if advancedHealth, ok := health["advanced_system_health"].(map[string]interface{}); ok {
				if status, ok := advancedHealth["status"].(string); ok && status != "healthy" {
					sfs.logger.Warn("Advanced security system health issue detected",
						zap.String("status", status))
				}
			}
		}
	}
}

// metricsReportingWorker runs the metrics reporting worker
func (sfs *SecurityFeedbackService) metricsReportingWorker() {
	ticker := time.NewTicker(sfs.config.MetricsReportingInterval)
	defer ticker.Stop()

	sfs.logger.Info("Security feedback service metrics reporting worker started",
		zap.Duration("interval", sfs.config.MetricsReportingInterval))

	for {
		select {
		case <-sfs.stopChan:
			sfs.logger.Info("Security feedback service metrics reporting worker stopped")
			return
		case <-ticker.C:
			ctx := context.Background()
			status := sfs.GetServiceStatus(ctx)

			// Log metrics
			if metrics, ok := status["security_metrics"].(*SecurityFeedbackMetrics); ok {
				sfs.logger.Info("Security feedback service metrics report",
					zap.Int64("total_feedback_collected", metrics.TotalFeedbackCollected),
					zap.Int64("analysis_runs_completed", metrics.AnalysisRunsCompleted),
					zap.Int64("improvement_runs_completed", metrics.ImprovementRunsCompleted),
					zap.Int64("security_violations_found", metrics.SecurityViolationsFound),
					zap.Int64("trusted_source_issues_found", metrics.TrustedSourceIssuesFound),
					zap.Int64("website_verification_issues", metrics.WebsiteVerificationIssues))
			}
		}
	}
}

// GetAdvancedSystem returns the underlying advanced security feedback system
func (sfs *SecurityFeedbackService) GetAdvancedSystem() *AdvancedSecurityFeedbackSystem {
	return sfs.advancedSystem
}

// GetSecurityAnalyzer returns the security analyzer
func (sfs *SecurityFeedbackService) GetSecurityAnalyzer() *SecurityFeedbackAnalyzer {
	return sfs.securityAnalyzer
}

// GetWebsiteVerificationImprover returns the website verification improver
func (sfs *SecurityFeedbackService) GetWebsiteVerificationImprover() *WebsiteVerificationImprover {
	return sfs.websiteVerificationImprover
}
