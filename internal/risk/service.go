package risk

import (
	"context"
	"fmt"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// RiskService provides risk assessment functionality
type RiskService struct {
	logger                *observability.Logger
	calculator            *RiskFactorCalculator
	scoringAlgorithm      ScoringAlgorithm
	predictionAlgorithm   *RiskPredictionAlgorithm
	thresholdManager      *ThresholdManager
	categoryRegistry      *RiskCategoryRegistry
	industryModelRegistry *IndustryModelRegistry
	historyService        *RiskHistoryService
	alertService          *AlertService
	reportService         *ReportService
	exportService         *ExportService
}

// NewRiskService creates a new risk service
func NewRiskService(
	logger *observability.Logger,
	calculator *RiskFactorCalculator,
	scoringAlgorithm ScoringAlgorithm,
	predictionAlgorithm *RiskPredictionAlgorithm,
	thresholdManager *ThresholdManager,
	categoryRegistry *RiskCategoryRegistry,
	industryModelRegistry *IndustryModelRegistry,
	historyService *RiskHistoryService,
	alertService *AlertService,
	reportService *ReportService,
	exportService *ExportService,
) *RiskService {
	return &RiskService{
		logger:                logger,
		calculator:            calculator,
		scoringAlgorithm:      scoringAlgorithm,
		predictionAlgorithm:   predictionAlgorithm,
		thresholdManager:      thresholdManager,
		categoryRegistry:      categoryRegistry,
		industryModelRegistry: industryModelRegistry,
		historyService:        historyService,
		alertService:          alertService,
		reportService:         reportService,
		exportService:         exportService,
	}
}

// AssessRisk performs a comprehensive risk assessment for a business
func (s *RiskService) AssessRisk(ctx context.Context, request RiskAssessmentRequest) (*RiskAssessmentResponse, error) {
	startTime := time.Now()
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Risk assessment started",
		"request_id", requestID,
		"business_id", request.BusinessID,
		"business_name", request.BusinessName,
		"categories", request.Categories,
	)

	// Validate request
	if err := s.validateAssessmentRequest(request); err != nil {
		s.logger.Error("Invalid risk assessment request",
			"request_id", requestID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Create assessment ID
	assessmentID := fmt.Sprintf("assessment_%d", time.Now().UnixNano())

	// Perform risk assessment
	assessment, err := s.performRiskAssessment(ctx, request, assessmentID)
	if err != nil {
		s.logger.Error("Risk assessment failed",
			"request_id", requestID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("risk assessment failed: %w", err)
	}

	// Generate predictions if requested
	var predictions []RiskPrediction
	if request.IncludePredictions {
		predictions, err = s.generatePredictions(ctx, request, assessment)
		if err != nil {
			s.logger.Warn("Failed to generate predictions",
				"request_id", requestID,
				"error", err.Error(),
			)
			// Don't fail the entire assessment if predictions fail
		}
	}

	// Generate alerts based on assessment
	var alerts []RiskAlert
	if s.alertService != nil {
		alerts, err = s.alertService.GenerateAlerts(ctx, assessment)
		if err != nil {
			s.logger.Warn("Failed to generate alerts",
				"request_id", requestID,
				"error", err.Error(),
			)
			// Don't fail the entire assessment if alerts fail
		}
	}

	// Store assessment in history
	if s.historyService != nil {
		if err := s.historyService.StoreRiskAssessment(ctx, assessment); err != nil {
			s.logger.Warn("Failed to store risk assessment in history",
				"request_id", requestID,
				"business_id", request.BusinessID,
				"error", err.Error(),
			)
			// Don't fail the entire assessment if history storage fails
		}
	}

	// Create response
	response := &RiskAssessmentResponse{
		Assessment:  assessment,
		Predictions: predictions,
		Alerts:      alerts,
		GeneratedAt: time.Now(),
	}

	duration := time.Since(startTime)
	s.logger.Info("Risk assessment completed",
		"request_id", requestID,
		"business_id", request.BusinessID,
		"overall_score", assessment.OverallScore,
		"overall_level", assessment.OverallLevel,
		"duration_ms", duration.Milliseconds(),
	)

	return response, nil
}

// validateAssessmentRequest validates the risk assessment request
func (s *RiskService) validateAssessmentRequest(request RiskAssessmentRequest) error {
	if request.BusinessID == "" {
		return fmt.Errorf("business ID is required")
	}
	if request.BusinessName == "" {
		return fmt.Errorf("business name is required")
	}
	return nil
}

// performRiskAssessment performs the core risk assessment
func (s *RiskService) performRiskAssessment(ctx context.Context, request RiskAssessmentRequest, assessmentID string) (*RiskAssessment, error) {
	// Get risk factors based on categories
	factors, err := s.getRiskFactors(request.Categories)
	if err != nil {
		return nil, fmt.Errorf("failed to get risk factors: %w", err)
	}

	// Prepare assessment data
	data := s.prepareAssessmentData(request)

	// Calculate risk scores using scoring algorithm
	overallScore, _, err := s.scoringAlgorithm.CalculateScore(factors, data)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate risk scores: %w", err)
	}

	// Get thresholds for risk level determination
	defaultConfig, _ := s.thresholdManager.GetDefaultConfig(RiskCategoryFinancial)
	thresholds := defaultConfig.RiskLevels

	// Determine overall risk level
	overallLevel := s.scoringAlgorithm.CalculateLevel(overallScore, thresholds)

	// Calculate individual factor scores
	factorScores, err := s.calculateFactorScores(factors, data)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate factor scores: %w", err)
	}

	// Calculate category scores
	categoryScores, err := s.calculateCategoryScores(factorScores)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate category scores: %w", err)
	}

	// Generate recommendations
	recommendations, err := s.generateRecommendations(assessmentID, factorScores, overallLevel)
	if err != nil {
		return nil, fmt.Errorf("failed to generate recommendations: %w", err)
	}

	assessment := &RiskAssessment{
		ID:              assessmentID,
		BusinessID:      request.BusinessID,
		BusinessName:    request.BusinessName,
		OverallScore:    overallScore,
		OverallLevel:    overallLevel,
		CategoryScores:  categoryScores,
		FactorScores:    factorScores,
		Recommendations: recommendations,
		AssessedAt:      time.Now(),
		ValidUntil:      time.Now().Add(24 * time.Hour), // Valid for 24 hours
		Metadata:        request.Metadata,
	}

	return assessment, nil
}

// getRiskFactors retrieves risk factors for the specified categories
func (s *RiskService) getRiskFactors(categories []RiskCategory) ([]RiskFactor, error) {
	var factors []RiskFactor

	if len(categories) == 0 {
		// Get all factors if no specific categories requested
		allFactors := s.categoryRegistry.ListFactors()
		for _, factorDef := range allFactors {
			factors = append(factors, RiskFactor{
				ID:          factorDef.ID,
				Name:        factorDef.Name,
				Description: factorDef.Description,
				Category:    factorDef.Category,
				Weight:      factorDef.Weight,
				Thresholds:  factorDef.Thresholds,
			})
		}
	} else {
		// Get factors for specific categories
		for _, category := range categories {
			categoryFactors := s.categoryRegistry.GetFactorsByCategory(category)
			for _, factorDef := range categoryFactors {
				factors = append(factors, RiskFactor{
					ID:          factorDef.ID,
					Name:        factorDef.Name,
					Description: factorDef.Description,
					Category:    factorDef.Category,
					Weight:      factorDef.Weight,
					Thresholds:  factorDef.Thresholds,
				})
			}
		}
	}

	return factors, nil
}

// prepareAssessmentData prepares data for risk assessment
func (s *RiskService) prepareAssessmentData(request RiskAssessmentRequest) map[string]interface{} {
	data := make(map[string]interface{})

	// Add business metadata
	data["business_id"] = request.BusinessID
	data["business_name"] = request.BusinessName

	// Add request metadata
	if request.Metadata != nil {
		for key, value := range request.Metadata {
			data[key] = value
		}
	}

	// Add default risk data (in a real implementation, this would come from external sources)
	data["financial_data"] = map[string]interface{}{
		"revenue":       1000000.0,
		"debt_ratio":    0.4,
		"cash_flow":     100000.0,
		"profit_margin": 0.15,
		"credit_score":  750,
	}

	data["operational_data"] = map[string]interface{}{
		"employee_count":    50,
		"years_in_business": 5,
		"process_maturity":  3.5,
		"compliance_score":  85,
	}

	data["regulatory_data"] = map[string]interface{}{
		"compliance_violations": 0,
		"regulatory_fines":      0,
		"license_status":        "active",
		"audit_findings":        0,
	}

	data["reputational_data"] = map[string]interface{}{
		"customer_satisfaction": 4.2,
		"online_reviews":        4.5,
		"media_sentiment":       0.7,
		"social_media_score":    0.8,
	}

	data["cybersecurity_data"] = map[string]interface{}{
		"security_score":       75,
		"data_breaches":        0,
		"vulnerability_count":  2,
		"compliance_framework": "SOC2",
	}

	return data
}

// calculateFactorScores calculates individual factor scores
func (s *RiskService) calculateFactorScores(factors []RiskFactor, data map[string]interface{}) ([]RiskScore, error) {
	var factorScores []RiskScore

	for _, factor := range factors {
		// Get factor-specific data
		factorData, exists := data[string(factor.Category)+"_data"]
		if !exists {
			// Use generic data if category-specific data not available
			factorData = data
		}

		// Calculate score for this factor
		score, err := s.calculator.CalculateFactor(RiskFactorInput{
			FactorID:    factor.ID,
			Data:        factorData.(map[string]interface{}),
			Timestamp:   time.Now(),
			Source:      "risk_service",
			Reliability: 0.8,
		})
		if err != nil {
			s.logger.Warn("Failed to calculate factor score",
				"factor_id", factor.ID,
				"error", err.Error(),
			)
			continue
		}

		// Convert RiskFactorResult to RiskScore
		riskScore := RiskScore{
			FactorID:     score.FactorID,
			FactorName:   score.FactorName,
			Category:     score.Category,
			Score:        score.Score,
			Level:        score.Level,
			Confidence:   score.Confidence,
			Explanation:  score.Explanation,
			Evidence:     score.Evidence,
			CalculatedAt: score.CalculatedAt,
		}
		factorScores = append(factorScores, riskScore)
	}

	return factorScores, nil
}

// calculateCategoryScores calculates category-level scores from factor scores
func (s *RiskService) calculateCategoryScores(factorScores []RiskScore) (map[RiskCategory]RiskScore, error) {
	categoryScores := make(map[RiskCategory]RiskScore)
	categoryTotals := make(map[RiskCategory]float64)
	categoryCounts := make(map[RiskCategory]int)

	// Aggregate scores by category
	for _, factorScore := range factorScores {
		category := factorScore.Category
		categoryTotals[category] += factorScore.Score
		categoryCounts[category]++
	}

	// Calculate average scores for each category
	for category, total := range categoryTotals {
		count := categoryCounts[category]
		if count > 0 {
			averageScore := total / float64(count)

			// Determine level based on average score
			var level RiskLevel
			switch {
			case averageScore < 25:
				level = RiskLevelLow
			case averageScore < 50:
				level = RiskLevelMedium
			case averageScore < 75:
				level = RiskLevelHigh
			default:
				level = RiskLevelCritical
			}

			categoryScores[category] = RiskScore{
				FactorID:     string(category),
				FactorName:   string(category),
				Category:     category,
				Score:        averageScore,
				Level:        level,
				Confidence:   0.8, // Default confidence
				Explanation:  fmt.Sprintf("Average score from %d factors", count),
				CalculatedAt: time.Now(),
			}
		}
	}

	return categoryScores, nil
}

// generateRecommendations generates risk mitigation recommendations
func (s *RiskService) generateRecommendations(assessmentID string, factorScores []RiskScore, overallLevel RiskLevel) ([]RiskRecommendation, error) {
	var recommendations []RiskRecommendation

	// Generate recommendations based on high-risk factors
	for _, factorScore := range factorScores {
		if factorScore.Level == RiskLevelHigh || factorScore.Level == RiskLevelCritical {
			recommendation := RiskRecommendation{
				ID:          fmt.Sprintf("rec_%s_%s", assessmentID, factorScore.FactorID),
				RiskFactor:  factorScore.FactorID,
				Title:       fmt.Sprintf("Address %s Risk", factorScore.FactorName),
				Description: fmt.Sprintf("High risk detected in %s category. Score: %.1f", factorScore.Category, factorScore.Score),
				Priority:    factorScore.Level,
				Action:      "Implement risk mitigation measures",
				Impact:      "Reduce risk score and improve overall assessment",
				Timeline:    "30 days",
				CreatedAt:   time.Now(),
			}
			recommendations = append(recommendations, recommendation)
		}
	}

	// Add overall recommendations based on risk level
	if overallLevel == RiskLevelHigh || overallLevel == RiskLevelCritical {
		recommendation := RiskRecommendation{
			ID:          fmt.Sprintf("rec_%s_overall", assessmentID),
			RiskFactor:  "overall_risk",
			Title:       "Comprehensive Risk Review Required",
			Description: "Overall risk level is high. Conduct comprehensive risk assessment and implement mitigation strategies.",
			Priority:    overallLevel,
			Action:      "Schedule risk review meeting and develop action plan",
			Impact:      "Improve overall risk posture",
			Timeline:    "7 days",
			CreatedAt:   time.Now(),
		}
		recommendations = append(recommendations, recommendation)
	}

	return recommendations, nil
}

// generatePredictions generates risk predictions for future time horizons
func (s *RiskService) generatePredictions(ctx context.Context, request RiskAssessmentRequest, assessment *RiskAssessment) ([]RiskPrediction, error) {
	// In a real implementation, this would use historical data
	// For now, we'll create mock historical trends based on current assessment

	now := time.Now()
	historicalTrends := []RiskTrend{
		{
			BusinessID: request.BusinessID,
			Category:   RiskCategoryFinancial,
			Score:      assessment.OverallScore * 0.9,
			Level:      assessment.OverallLevel,
			RecordedAt: now.Add(-90 * 24 * time.Hour),
		},
		{
			BusinessID: request.BusinessID,
			Category:   RiskCategoryFinancial,
			Score:      assessment.OverallScore * 0.95,
			Level:      assessment.OverallLevel,
			RecordedAt: now.Add(-60 * 24 * time.Hour),
		},
		{
			BusinessID: request.BusinessID,
			Category:   RiskCategoryFinancial,
			Score:      assessment.OverallScore,
			Level:      assessment.OverallLevel,
			RecordedAt: now,
		},
	}

	horizons := []time.Duration{
		30 * 24 * time.Hour,  // 1 month
		90 * 24 * time.Hour,  // 3 months
		180 * 24 * time.Hour, // 6 months
	}

	return s.predictionAlgorithm.PredictMultipleHorizons(historicalTrends, horizons)
}

// GetCategoryRegistry returns the category registry
func (s *RiskService) GetCategoryRegistry() *RiskCategoryRegistry {
	return s.categoryRegistry
}

// GetThresholdManager returns the threshold manager
func (s *RiskService) GetThresholdManager() *ThresholdManager {
	return s.thresholdManager
}

// GenerateRiskReport generates a comprehensive risk report for a business
func (s *RiskService) GenerateRiskReport(ctx context.Context, request ReportRequest) (*RiskReport, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Generating risk report",
		"request_id", requestID,
		"business_id", request.BusinessID,
		"report_type", request.ReportType,
		"format", request.Format,
	)

	// Use the report service to generate the report
	if s.reportService == nil {
		return nil, fmt.Errorf("report service not available")
	}

	report, err := s.reportService.GenerateReport(ctx, request)
	if err != nil {
		s.logger.Error("Failed to generate risk report",
			"request_id", requestID,
			"business_id", request.BusinessID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to generate risk report: %w", err)
	}

	s.logger.Info("Risk report generated successfully",
		"request_id", requestID,
		"business_id", request.BusinessID,
		"report_type", request.ReportType,
		"format", request.Format,
	)

	return report, nil
}

// ExportRiskData exports risk data in the specified format
func (s *RiskService) ExportRiskData(ctx context.Context, request ExportRequest) (*ExportResponse, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Exporting risk data",
		"request_id", requestID,
		"business_id", request.BusinessID,
		"export_type", request.ExportType,
		"format", request.Format,
	)

	// Use the export service to export data
	if s.exportService == nil {
		return nil, fmt.Errorf("export service not available")
	}

	response, err := s.exportService.ExportRiskData(ctx, request)
	if err != nil {
		s.logger.Error("Failed to export risk data",
			"request_id", requestID,
			"business_id", request.BusinessID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to export risk data: %w", err)
	}

	s.logger.Info("Risk data export completed",
		"request_id", requestID,
		"business_id", request.BusinessID,
		"export_type", request.ExportType,
		"format", request.Format,
		"record_count", response.RecordCount,
	)

	return response, nil
}

// CreateExportJob creates a background export job for large datasets
func (s *RiskService) CreateExportJob(ctx context.Context, request ExportRequest) (*ExportJob, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Creating export job",
		"request_id", requestID,
		"business_id", request.BusinessID,
		"export_type", request.ExportType,
		"format", request.Format,
	)

	// Use the export service to create job
	if s.exportService == nil {
		return nil, fmt.Errorf("export service not available")
	}

	job, err := s.exportService.CreateExportJob(ctx, request)
	if err != nil {
		s.logger.Error("Failed to create export job",
			"request_id", requestID,
			"business_id", request.BusinessID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to create export job: %w", err)
	}

	s.logger.Info("Export job created",
		"request_id", requestID,
		"job_id", job.ID,
		"business_id", request.BusinessID,
	)

	return job, nil
}

// GetExportJob retrieves the status of an export job
func (s *RiskService) GetExportJob(ctx context.Context, jobID string) (*ExportJob, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Retrieving export job",
		"request_id", requestID,
		"job_id", jobID,
	)

	// Use the export service to get job status
	if s.exportService == nil {
		return nil, fmt.Errorf("export service not available")
	}

	job, err := s.exportService.GetExportJob(ctx, jobID)
	if err != nil {
		s.logger.Error("Failed to retrieve export job",
			"request_id", requestID,
			"job_id", jobID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to retrieve export job: %w", err)
	}

	s.logger.Info("Export job retrieved",
		"request_id", requestID,
		"job_id", jobID,
		"status", job.Status,
		"progress", job.Progress,
	)

	return job, nil
}
