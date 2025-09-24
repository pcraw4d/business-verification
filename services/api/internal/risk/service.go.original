package risk

import (
	"context"
	"fmt"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// RiskService provides risk assessment functionality
type RiskService struct {
	logger                     *observability.Logger
	calculator                 *RiskFactorCalculator
	scoringAlgorithm           ScoringAlgorithm
	predictionAlgorithm        *RiskPredictionAlgorithm
	thresholdManager           *ThresholdManager
	categoryRegistry           *RiskCategoryRegistry
	industryModelRegistry      *IndustryModelRegistry
	historyService             *RiskHistoryService
	alertService               *AlertService
	reportService              *ReportService
	exportService              *ExportService
	financialProviderManager   *FinancialProviderManager
	regulatoryProviderManager  *RegulatoryProviderManager
	mediaProviderManager       *MediaProviderManager
	marketDataProviderManager  *MarketDataProviderManager
	dataValidationManager      *DataValidationManager
	thresholdMonitoringManager *ThresholdMonitoringManager
	automatedAlertService      *AutomatedAlertService
	trendAnalysisService       *TrendAnalysisService
	reportingSystem            *ReportingSystem
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
	financialProviderManager *FinancialProviderManager,
	regulatoryProviderManager *RegulatoryProviderManager,
	mediaProviderManager *MediaProviderManager,
	marketDataProviderManager *MarketDataProviderManager,
	dataValidationManager *DataValidationManager,
	thresholdMonitoringManager *ThresholdMonitoringManager,
	automatedAlertService *AutomatedAlertService,
	trendAnalysisService *TrendAnalysisService,
	reportingSystem *ReportingSystem,
) *RiskService {
	return &RiskService{
		logger:                     logger,
		calculator:                 calculator,
		scoringAlgorithm:           scoringAlgorithm,
		predictionAlgorithm:        predictionAlgorithm,
		thresholdManager:           thresholdManager,
		categoryRegistry:           categoryRegistry,
		industryModelRegistry:      industryModelRegistry,
		historyService:             historyService,
		alertService:               alertService,
		reportService:              reportService,
		exportService:              exportService,
		financialProviderManager:   financialProviderManager,
		regulatoryProviderManager:  regulatoryProviderManager,
		mediaProviderManager:       mediaProviderManager,
		marketDataProviderManager:  marketDataProviderManager,
		dataValidationManager:      dataValidationManager,
		thresholdMonitoringManager: thresholdMonitoringManager,
		automatedAlertService:      automatedAlertService,
		trendAnalysisService:       trendAnalysisService,
		reportingSystem:            reportingSystem,
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

// GetCompanyFinancials retrieves financial data for a business
func (s *RiskService) GetCompanyFinancials(ctx context.Context, businessID string) (*FinancialData, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Retrieving company financials",
		"request_id", requestID,
		"business_id", businessID,
	)

	if s.financialProviderManager == nil {
		return nil, fmt.Errorf("financial provider manager not available")
	}

	data, err := s.financialProviderManager.GetCompanyFinancials(ctx, businessID)
	if err != nil {
		s.logger.Error("Failed to retrieve company financials",
			"request_id", requestID,
			"business_id", businessID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to retrieve company financials: %w", err)
	}

	s.logger.Info("Retrieved company financials successfully",
		"request_id", requestID,
		"business_id", businessID,
		"provider", data.Provider,
	)

	return data, nil
}

// GetCreditScore retrieves credit score for a business
func (s *RiskService) GetCreditScore(ctx context.Context, businessID string) (*CreditScore, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Retrieving credit score",
		"request_id", requestID,
		"business_id", businessID,
	)

	if s.financialProviderManager == nil {
		return nil, fmt.Errorf("financial provider manager not available")
	}

	score, err := s.financialProviderManager.GetCreditScore(ctx, businessID)
	if err != nil {
		s.logger.Error("Failed to retrieve credit score",
			"request_id", requestID,
			"business_id", businessID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to retrieve credit score: %w", err)
	}

	s.logger.Info("Retrieved credit score successfully",
		"request_id", requestID,
		"business_id", businessID,
		"provider", score.Provider,
		"score", score.Score,
	)

	return score, nil
}

// GetSanctionsData retrieves sanctions data for a business
func (s *RiskService) GetSanctionsData(ctx context.Context, businessID string) (*SanctionsData, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Retrieving sanctions data",
		"request_id", requestID,
		"business_id", businessID,
	)

	if s.regulatoryProviderManager == nil {
		return nil, fmt.Errorf("regulatory provider manager not available")
	}

	data, err := s.regulatoryProviderManager.GetSanctionsData(ctx, businessID)
	if err != nil {
		s.logger.Error("Failed to retrieve sanctions data",
			"request_id", requestID,
			"business_id", businessID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to retrieve sanctions data: %w", err)
	}

	s.logger.Info("Retrieved sanctions data successfully",
		"request_id", requestID,
		"business_id", businessID,
		"provider", data.Provider,
		"has_sanctions", data.HasSanctions,
	)

	return data, nil
}

// GetLicenseData retrieves license data for a business
func (s *RiskService) GetLicenseData(ctx context.Context, businessID string) (*LicenseData, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Retrieving license data",
		"request_id", requestID,
		"business_id", businessID,
	)

	if s.regulatoryProviderManager == nil {
		return nil, fmt.Errorf("regulatory provider manager not available")
	}

	data, err := s.regulatoryProviderManager.GetLicenseData(ctx, businessID)
	if err != nil {
		s.logger.Error("Failed to retrieve license data",
			"request_id", requestID,
			"business_id", businessID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to retrieve license data: %w", err)
	}

	s.logger.Info("Retrieved license data successfully",
		"request_id", requestID,
		"business_id", businessID,
		"provider", data.Provider,
		"overall_status", data.OverallStatus,
	)

	return data, nil
}

// GetComplianceData retrieves compliance data for a business
func (s *RiskService) GetComplianceData(ctx context.Context, businessID string) (*ComplianceData, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Retrieving compliance data",
		"request_id", requestID,
		"business_id", businessID,
	)

	if s.regulatoryProviderManager == nil {
		return nil, fmt.Errorf("regulatory provider manager not available")
	}

	data, err := s.regulatoryProviderManager.GetComplianceData(ctx, businessID)
	if err != nil {
		s.logger.Error("Failed to retrieve compliance data",
			"request_id", requestID,
			"business_id", businessID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to retrieve compliance data: %w", err)
	}

	s.logger.Info("Retrieved compliance data successfully",
		"request_id", requestID,
		"business_id", businessID,
		"provider", data.Provider,
		"overall_score", data.OverallScore,
	)

	return data, nil
}

// GetRegulatoryViolations retrieves regulatory violations for a business
func (s *RiskService) GetRegulatoryViolations(ctx context.Context, businessID string) (*RegulatoryViolations, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Retrieving regulatory violations",
		"request_id", requestID,
		"business_id", businessID,
	)

	if s.regulatoryProviderManager == nil {
		return nil, fmt.Errorf("regulatory provider manager not available")
	}

	data, err := s.regulatoryProviderManager.GetRegulatoryViolations(ctx, businessID)
	if err != nil {
		s.logger.Error("Failed to retrieve regulatory violations",
			"request_id", requestID,
			"business_id", businessID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to retrieve regulatory violations: %w", err)
	}

	s.logger.Info("Retrieved regulatory violations successfully",
		"request_id", requestID,
		"business_id", businessID,
		"provider", data.Provider,
		"total_violations", data.TotalViolations,
	)

	return data, nil
}

// GetTaxComplianceData retrieves tax compliance data for a business
func (s *RiskService) GetTaxComplianceData(ctx context.Context, businessID string) (*TaxComplianceData, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Retrieving tax compliance data",
		"request_id", requestID,
		"business_id", businessID,
	)

	if s.regulatoryProviderManager == nil {
		return nil, fmt.Errorf("regulatory provider manager not available")
	}

	data, err := s.regulatoryProviderManager.GetTaxComplianceData(ctx, businessID)
	if err != nil {
		s.logger.Error("Failed to retrieve tax compliance data",
			"request_id", requestID,
			"business_id", businessID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to retrieve tax compliance data: %w", err)
	}

	s.logger.Info("Retrieved tax compliance data successfully",
		"request_id", requestID,
		"business_id", businessID,
		"provider", data.Provider,
		"tax_id_status", data.TaxIDStatus,
	)

	return data, nil
}

// GetDataProtectionCompliance retrieves data protection compliance for a business
func (s *RiskService) GetDataProtectionCompliance(ctx context.Context, businessID string) (*DataProtectionCompliance, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Retrieving data protection compliance",
		"request_id", requestID,
		"business_id", businessID,
	)

	if s.regulatoryProviderManager == nil {
		return nil, fmt.Errorf("regulatory provider manager not available")
	}

	data, err := s.regulatoryProviderManager.GetDataProtectionCompliance(ctx, businessID)
	if err != nil {
		s.logger.Error("Failed to retrieve data protection compliance",
			"request_id", requestID,
			"business_id", businessID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to retrieve data protection compliance: %w", err)
	}

	s.logger.Info("Retrieved data protection compliance successfully",
		"request_id", requestID,
		"business_id", businessID,
		"provider", data.Provider,
		"overall_score", data.OverallScore,
	)

	return data, nil
}

// GetNewsArticles retrieves news articles for a business
func (s *RiskService) GetNewsArticles(ctx context.Context, businessID string, query NewsQuery) (*NewsResult, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Retrieving news articles",
		"request_id", requestID,
		"business_id", businessID,
	)

	if s.mediaProviderManager == nil {
		return nil, fmt.Errorf("media provider manager not available")
	}

	result, err := s.mediaProviderManager.GetNewsArticles(ctx, businessID, query)
	if err != nil {
		s.logger.Error("Failed to retrieve news articles",
			"request_id", requestID,
			"business_id", businessID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to retrieve news articles: %w", err)
	}

	s.logger.Info("Retrieved news articles successfully",
		"request_id", requestID,
		"business_id", businessID,
		"provider", result.Provider,
		"total_articles", result.TotalArticles,
	)

	return result, nil
}

// GetSocialMediaMentions retrieves social media mentions for a business
func (s *RiskService) GetSocialMediaMentions(ctx context.Context, businessID string, query SocialMediaQuery) (*SocialMediaResult, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Retrieving social media mentions",
		"request_id", requestID,
		"business_id", businessID,
	)

	if s.mediaProviderManager == nil {
		return nil, fmt.Errorf("media provider manager not available")
	}

	result, err := s.mediaProviderManager.GetSocialMediaMentions(ctx, businessID, query)
	if err != nil {
		s.logger.Error("Failed to retrieve social media mentions",
			"request_id", requestID,
			"business_id", businessID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to retrieve social media mentions: %w", err)
	}

	s.logger.Info("Retrieved social media mentions successfully",
		"request_id", requestID,
		"business_id", businessID,
		"provider", result.Provider,
		"total_mentions", result.TotalMentions,
	)

	return result, nil
}

// GetMediaSentiment retrieves media sentiment for a business
func (s *RiskService) GetMediaSentiment(ctx context.Context, businessID string) (*SentimentResult, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Retrieving media sentiment",
		"request_id", requestID,
		"business_id", businessID,
	)

	if s.mediaProviderManager == nil {
		return nil, fmt.Errorf("media provider manager not available")
	}

	result, err := s.mediaProviderManager.GetMediaSentiment(ctx, businessID)
	if err != nil {
		s.logger.Error("Failed to retrieve media sentiment",
			"request_id", requestID,
			"business_id", businessID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to retrieve media sentiment: %w", err)
	}

	s.logger.Info("Retrieved media sentiment successfully",
		"request_id", requestID,
		"business_id", businessID,
		"provider", result.Provider,
		"overall_score", result.OverallScore,
	)

	return result, nil
}

// GetReputationScore retrieves reputation score for a business
func (s *RiskService) GetReputationScore(ctx context.Context, businessID string) (*ReputationScore, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Retrieving reputation score",
		"request_id", requestID,
		"business_id", businessID,
	)

	if s.mediaProviderManager == nil {
		return nil, fmt.Errorf("media provider manager not available")
	}

	result, err := s.mediaProviderManager.GetReputationScore(ctx, businessID)
	if err != nil {
		s.logger.Error("Failed to retrieve reputation score",
			"request_id", requestID,
			"business_id", businessID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to retrieve reputation score: %w", err)
	}

	s.logger.Info("Retrieved reputation score successfully",
		"request_id", requestID,
		"business_id", businessID,
		"provider", result.Provider,
		"overall_score", result.OverallScore,
	)

	return result, nil
}

// GetMediaAlerts retrieves media alerts for a business
func (s *RiskService) GetMediaAlerts(ctx context.Context, businessID string) (*MediaAlerts, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Retrieving media alerts",
		"request_id", requestID,
		"business_id", businessID,
	)

	if s.mediaProviderManager == nil {
		return nil, fmt.Errorf("media provider manager not available")
	}

	result, err := s.mediaProviderManager.GetMediaAlerts(ctx, businessID)
	if err != nil {
		s.logger.Error("Failed to retrieve media alerts",
			"request_id", requestID,
			"business_id", businessID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to retrieve media alerts: %w", err)
	}

	s.logger.Info("Retrieved media alerts successfully",
		"request_id", requestID,
		"business_id", businessID,
		"provider", result.Provider,
		"total_alerts", result.TotalAlerts,
	)

	return result, nil
}

// GetEconomicIndicators retrieves economic indicators for a country
func (s *RiskService) GetEconomicIndicators(ctx context.Context, country string) (*EconomicIndicators, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Retrieving economic indicators",
		"request_id", requestID,
		"country", country,
	)

	if s.marketDataProviderManager == nil {
		return nil, fmt.Errorf("market data provider manager not available")
	}

	result, err := s.marketDataProviderManager.GetEconomicIndicators(ctx, country)
	if err != nil {
		s.logger.Error("Failed to retrieve economic indicators",
			"request_id", requestID,
			"country", country,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to retrieve economic indicators: %w", err)
	}

	s.logger.Info("Retrieved economic indicators successfully",
		"request_id", requestID,
		"country", country,
		"provider", result.Provider,
	)

	return result, nil
}

// GetMarketIndustryBenchmarks retrieves industry benchmarks for a market
func (s *RiskService) GetMarketIndustryBenchmarks(ctx context.Context, industry string, region string) (*MarketIndustryBenchmarks, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Retrieving market industry benchmarks",
		"request_id", requestID,
		"industry", industry,
		"region", region,
	)

	if s.marketDataProviderManager == nil {
		return nil, fmt.Errorf("market data provider manager not available")
	}

	result, err := s.marketDataProviderManager.GetIndustryBenchmarks(ctx, industry, region)
	if err != nil {
		s.logger.Error("Failed to retrieve market industry benchmarks",
			"request_id", requestID,
			"industry", industry,
			"region", region,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to retrieve market industry benchmarks: %w", err)
	}

	s.logger.Info("Retrieved market industry benchmarks successfully",
		"request_id", requestID,
		"industry", industry,
		"region", region,
		"provider", result.Provider,
	)

	return result, nil
}

// GetMarketRiskFactors retrieves market risk factors for a sector
func (s *RiskService) GetMarketRiskFactors(ctx context.Context, sector string) (*MarketRiskFactors, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Retrieving market risk factors",
		"request_id", requestID,
		"sector", sector,
	)

	if s.marketDataProviderManager == nil {
		return nil, fmt.Errorf("market data provider manager not available")
	}

	result, err := s.marketDataProviderManager.GetMarketRiskFactors(ctx, sector)
	if err != nil {
		s.logger.Error("Failed to retrieve market risk factors",
			"request_id", requestID,
			"sector", sector,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to retrieve market risk factors: %w", err)
	}

	s.logger.Info("Retrieved market risk factors successfully",
		"request_id", requestID,
		"sector", sector,
		"provider", result.Provider,
	)

	return result, nil
}

// GetCommodityPrices retrieves commodity prices
func (s *RiskService) GetCommodityPrices(ctx context.Context, commodities []string) (*CommodityPrices, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Retrieving commodity prices",
		"request_id", requestID,
		"commodities", commodities,
	)

	if s.marketDataProviderManager == nil {
		return nil, fmt.Errorf("market data provider manager not available")
	}

	result, err := s.marketDataProviderManager.GetCommodityPrices(ctx, commodities)
	if err != nil {
		s.logger.Error("Failed to retrieve commodity prices",
			"request_id", requestID,
			"commodities", commodities,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to retrieve commodity prices: %w", err)
	}

	s.logger.Info("Retrieved commodity prices successfully",
		"request_id", requestID,
		"commodities", commodities,
		"provider", result.Provider,
	)

	return result, nil
}

// GetCurrencyRates retrieves currency exchange rates
func (s *RiskService) GetCurrencyRates(ctx context.Context, baseCurrency string) (*CurrencyRates, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Retrieving currency rates",
		"request_id", requestID,
		"base_currency", baseCurrency,
	)

	if s.marketDataProviderManager == nil {
		return nil, fmt.Errorf("market data provider manager not available")
	}

	result, err := s.marketDataProviderManager.GetCurrencyRates(ctx, baseCurrency)
	if err != nil {
		s.logger.Error("Failed to retrieve currency rates",
			"request_id", requestID,
			"base_currency", baseCurrency,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to retrieve currency rates: %w", err)
	}

	s.logger.Info("Retrieved currency rates successfully",
		"request_id", requestID,
		"base_currency", baseCurrency,
		"provider", result.Provider,
	)

	return result, nil
}

// GetMarketTrends retrieves market trends
func (s *RiskService) GetMarketTrends(ctx context.Context, market string) (*MarketTrends, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Retrieving market trends",
		"request_id", requestID,
		"market", market,
	)

	if s.marketDataProviderManager == nil {
		return nil, fmt.Errorf("market data provider manager not available")
	}

	result, err := s.marketDataProviderManager.GetMarketTrends(ctx, market)
	if err != nil {
		s.logger.Error("Failed to retrieve market trends",
			"request_id", requestID,
			"market", market,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to retrieve market trends: %w", err)
	}

	s.logger.Info("Retrieved market trends successfully",
		"request_id", requestID,
		"market", market,
		"provider", result.Provider,
	)

	return result, nil
}

// ValidateFinancialData validates financial data
func (s *RiskService) ValidateFinancialData(ctx context.Context, data *FinancialData) (*ValidationResult, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Validating financial data",
		"request_id", requestID,
		"business_id", data.BusinessID,
	)

	if s.dataValidationManager == nil {
		return nil, fmt.Errorf("data validation manager not available")
	}

	result, err := s.dataValidationManager.ValidateFinancialData(data)
	if err != nil {
		s.logger.Error("Failed to validate financial data",
			"request_id", requestID,
			"business_id", data.BusinessID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to validate financial data: %w", err)
	}

	s.logger.Info("Financial data validation completed",
		"request_id", requestID,
		"business_id", data.BusinessID,
		"overall_score", result.OverallScore,
		"is_valid", result.IsValid,
	)

	return result, nil
}

// ValidateRegulatoryData validates regulatory data
func (s *RiskService) ValidateRegulatoryData(ctx context.Context, data *RegulatoryViolations) (*ValidationResult, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Validating regulatory data",
		"request_id", requestID,
		"business_id", data.BusinessID,
	)

	if s.dataValidationManager == nil {
		return nil, fmt.Errorf("data validation manager not available")
	}

	result, err := s.dataValidationManager.ValidateRegulatoryData(data)
	if err != nil {
		s.logger.Error("Failed to validate regulatory data",
			"request_id", requestID,
			"business_id", data.BusinessID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to validate regulatory data: %w", err)
	}

	s.logger.Info("Regulatory data validation completed",
		"request_id", requestID,
		"business_id", data.BusinessID,
		"overall_score", result.OverallScore,
		"is_valid", result.IsValid,
	)

	return result, nil
}

// ValidateMediaData validates media data
func (s *RiskService) ValidateMediaData(ctx context.Context, data *NewsResult) (*ValidationResult, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Validating media data",
		"request_id", requestID,
		"business_id", data.BusinessID,
	)

	if s.dataValidationManager == nil {
		return nil, fmt.Errorf("data validation manager not available")
	}

	result, err := s.dataValidationManager.ValidateMediaData(data)
	if err != nil {
		s.logger.Error("Failed to validate media data",
			"request_id", requestID,
			"business_id", data.BusinessID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to validate media data: %w", err)
	}

	s.logger.Info("Media data validation completed",
		"request_id", requestID,
		"business_id", data.BusinessID,
		"overall_score", result.OverallScore,
		"is_valid", result.IsValid,
	)

	return result, nil
}

// ValidateMarketData validates market data
func (s *RiskService) ValidateMarketData(ctx context.Context, data *EconomicIndicators) (*ValidationResult, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Validating market data",
		"request_id", requestID,
		"country", data.Country,
	)

	if s.dataValidationManager == nil {
		return nil, fmt.Errorf("data validation manager not available")
	}

	result, err := s.dataValidationManager.ValidateMarketData(data)
	if err != nil {
		s.logger.Error("Failed to validate market data",
			"request_id", requestID,
			"country", data.Country,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to validate market data: %w", err)
	}

	s.logger.Info("Market data validation completed",
		"request_id", requestID,
		"country", data.Country,
		"overall_score", result.OverallScore,
		"is_valid", result.IsValid,
	)

	return result, nil
}

// ValidateRiskAssessment validates risk assessment data
func (s *RiskService) ValidateRiskAssessment(ctx context.Context, assessment *RiskAssessment) (*ValidationResult, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Validating risk assessment",
		"request_id", requestID,
		"business_id", assessment.BusinessID,
	)

	if s.dataValidationManager == nil {
		return nil, fmt.Errorf("data validation manager not available")
	}

	result, err := s.dataValidationManager.ValidateRiskAssessment(assessment)
	if err != nil {
		s.logger.Error("Failed to validate risk assessment",
			"request_id", requestID,
			"business_id", assessment.BusinessID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to validate risk assessment: %w", err)
	}

	s.logger.Info("Risk assessment validation completed",
		"request_id", requestID,
		"business_id", assessment.BusinessID,
		"overall_score", result.OverallScore,
		"is_valid", result.IsValid,
	)

	return result, nil
}

// ValidateRiskFactor validates risk factor data
func (s *RiskService) ValidateRiskFactor(ctx context.Context, factor *RiskFactorResult) (*ValidationResult, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Validating risk factor",
		"request_id", requestID,
		"factor_id", factor.FactorID,
	)

	if s.dataValidationManager == nil {
		return nil, fmt.Errorf("data validation manager not available")
	}

	result, err := s.dataValidationManager.ValidateRiskFactor(factor)
	if err != nil {
		s.logger.Error("Failed to validate risk factor",
			"request_id", requestID,
			"factor_id", factor.FactorID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to validate risk factor: %w", err)
	}

	s.logger.Info("Risk factor validation completed",
		"request_id", requestID,
		"factor_id", factor.FactorID,
		"overall_score", result.OverallScore,
		"is_valid", result.IsValid,
	)

	return result, nil
}

// MonitorThresholds monitors risk thresholds for an assessment
func (s *RiskService) MonitorThresholds(ctx context.Context, assessment *RiskAssessment) ([]ThresholdAlert, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Monitoring thresholds for risk assessment",
		"request_id", requestID,
		"business_id", assessment.BusinessID,
	)

	if s.thresholdMonitoringManager == nil {
		return nil, fmt.Errorf("threshold monitoring manager not available")
	}

	alerts, err := s.thresholdMonitoringManager.MonitorThreshold(ctx, assessment)
	if err != nil {
		s.logger.Error("Failed to monitor thresholds",
			"request_id", requestID,
			"business_id", assessment.BusinessID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to monitor thresholds: %w", err)
	}

	s.logger.Info("Threshold monitoring completed",
		"request_id", requestID,
		"business_id", assessment.BusinessID,
		"alert_count", len(alerts),
	)

	return alerts, nil
}

// GetThresholdConfig retrieves threshold configuration for a category
func (s *RiskService) GetThresholdConfig(ctx context.Context, category RiskCategory) (*ThresholdMonitoringConfig, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Retrieving threshold config",
		"request_id", requestID,
		"category", category,
	)

	if s.thresholdMonitoringManager == nil {
		return nil, fmt.Errorf("threshold monitoring manager not available")
	}

	config, err := s.thresholdMonitoringManager.GetThresholdConfig(category)
	if err != nil {
		s.logger.Error("Failed to retrieve threshold config",
			"request_id", requestID,
			"category", category,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to retrieve threshold config: %w", err)
	}

	s.logger.Info("Retrieved threshold config",
		"request_id", requestID,
		"category", category,
		"warning_threshold", config.WarningThreshold,
		"critical_threshold", config.CriticalThreshold,
	)

	return config, nil
}

// UpdateThresholdConfig updates threshold configuration for a category
func (s *RiskService) UpdateThresholdConfig(ctx context.Context, category RiskCategory, config *ThresholdMonitoringConfig) error {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Updating threshold config",
		"request_id", requestID,
		"category", category,
	)

	if s.thresholdMonitoringManager == nil {
		return fmt.Errorf("threshold monitoring manager not available")
	}

	err := s.thresholdMonitoringManager.UpdateThresholdConfig(category, config)
	if err != nil {
		s.logger.Error("Failed to update threshold config",
			"request_id", requestID,
			"category", category,
			"error", err.Error(),
		)
		return fmt.Errorf("failed to update threshold config: %w", err)
	}

	s.logger.Info("Updated threshold config",
		"request_id", requestID,
		"category", category,
		"warning_threshold", config.WarningThreshold,
		"critical_threshold", config.CriticalThreshold,
	)

	return nil
}

// GetMonitoringStatus retrieves the current monitoring status
func (s *RiskService) GetMonitoringStatus(ctx context.Context) (*MonitoringStatus, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Retrieving monitoring status",
		"request_id", requestID,
	)

	if s.thresholdMonitoringManager == nil {
		return nil, fmt.Errorf("threshold monitoring manager not available")
	}

	status, err := s.thresholdMonitoringManager.GetMonitoringStatus()
	if err != nil {
		s.logger.Error("Failed to retrieve monitoring status",
			"request_id", requestID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to retrieve monitoring status: %w", err)
	}

	s.logger.Info("Retrieved monitoring status",
		"request_id", requestID,
		"active_monitors", status.ActiveMonitors,
		"total_alerts", status.TotalAlerts,
		"monitoring_health", status.MonitoringHealth,
	)

	return status, nil
}

// ProcessAutomatedAlerts processes automated alerts for an assessment
func (s *RiskService) ProcessAutomatedAlerts(ctx context.Context, assessment *RiskAssessment) ([]AutomatedAlert, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Processing automated alerts for assessment",
		"request_id", requestID,
		"business_id", assessment.BusinessID,
	)

	if s.automatedAlertService == nil {
		return nil, fmt.Errorf("automated alert service not available")
	}

	alerts, err := s.automatedAlertService.ProcessAssessment(ctx, assessment)
	if err != nil {
		s.logger.Error("Failed to process automated alerts",
			"request_id", requestID,
			"business_id", assessment.BusinessID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to process automated alerts: %w", err)
	}

	s.logger.Info("Automated alert processing completed",
		"request_id", requestID,
		"business_id", assessment.BusinessID,
		"alert_count", len(alerts),
	)

	return alerts, nil
}

// GetAutomatedAlertRules retrieves all automated alert rules
func (s *RiskService) GetAutomatedAlertRules(ctx context.Context) ([]*AutomatedAlertRule, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Retrieving automated alert rules",
		"request_id", requestID,
	)

	if s.automatedAlertService == nil {
		return nil, fmt.Errorf("automated alert service not available")
	}

	rules, err := s.automatedAlertService.GetAlertRules()
	if err != nil {
		s.logger.Error("Failed to retrieve automated alert rules",
			"request_id", requestID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to retrieve automated alert rules: %w", err)
	}

	s.logger.Info("Retrieved automated alert rules",
		"request_id", requestID,
		"rule_count", len(rules),
	)

	return rules, nil
}

// CreateAutomatedAlertRule creates a new automated alert rule
func (s *RiskService) CreateAutomatedAlertRule(ctx context.Context, rule *AutomatedAlertRule) error {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Creating automated alert rule",
		"request_id", requestID,
		"rule_name", rule.Name,
	)

	if s.automatedAlertService == nil {
		return fmt.Errorf("automated alert service not available")
	}

	err := s.automatedAlertService.CreateAlertRule(rule)
	if err != nil {
		s.logger.Error("Failed to create automated alert rule",
			"request_id", requestID,
			"rule_name", rule.Name,
			"error", err.Error(),
		)
		return fmt.Errorf("failed to create automated alert rule: %w", err)
	}

	s.logger.Info("Created automated alert rule",
		"request_id", requestID,
		"rule_id", rule.ID,
		"rule_name", rule.Name,
	)

	return nil
}

// UpdateAutomatedAlertRule updates an existing automated alert rule
func (s *RiskService) UpdateAutomatedAlertRule(ctx context.Context, ruleID string, rule *AutomatedAlertRule) error {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Updating automated alert rule",
		"request_id", requestID,
		"rule_id", ruleID,
	)

	if s.automatedAlertService == nil {
		return fmt.Errorf("automated alert service not available")
	}

	err := s.automatedAlertService.UpdateAlertRule(ruleID, rule)
	if err != nil {
		s.logger.Error("Failed to update automated alert rule",
			"request_id", requestID,
			"rule_id", ruleID,
			"error", err.Error(),
		)
		return fmt.Errorf("failed to update automated alert rule: %w", err)
	}

	s.logger.Info("Updated automated alert rule",
		"request_id", requestID,
		"rule_id", ruleID,
	)

	return nil
}

// DeleteAutomatedAlertRule deletes an automated alert rule
func (s *RiskService) DeleteAutomatedAlertRule(ctx context.Context, ruleID string) error {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Deleting automated alert rule",
		"request_id", requestID,
		"rule_id", ruleID,
	)

	if s.automatedAlertService == nil {
		return fmt.Errorf("automated alert service not available")
	}

	err := s.automatedAlertService.DeleteAlertRule(ruleID)
	if err != nil {
		s.logger.Error("Failed to delete automated alert rule",
			"request_id", requestID,
			"rule_id", ruleID,
			"error", err.Error(),
		)
		return fmt.Errorf("failed to delete automated alert rule: %w", err)
	}

	s.logger.Info("Deleted automated alert rule",
		"request_id", requestID,
		"rule_id", ruleID,
	)

	return nil
}

// GetAutomatedAlertHistory retrieves automated alert history for a business
func (s *RiskService) GetAutomatedAlertHistory(ctx context.Context, businessID string) ([]AutomatedAlert, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Retrieving automated alert history",
		"request_id", requestID,
		"business_id", businessID,
	)

	if s.automatedAlertService == nil {
		return nil, fmt.Errorf("automated alert service not available")
	}

	history, err := s.automatedAlertService.GetAlertHistory(businessID)
	if err != nil {
		s.logger.Error("Failed to retrieve automated alert history",
			"request_id", requestID,
			"business_id", businessID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to retrieve automated alert history: %w", err)
	}

	s.logger.Info("Retrieved automated alert history",
		"request_id", requestID,
		"business_id", businessID,
		"alert_count", len(history),
	)

	return history, nil
}

// RegisterNotificationProvider registers a notification provider with the automated alert service
func (s *RiskService) RegisterNotificationProvider(ctx context.Context, channel string, provider NotificationProvider) error {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Registering notification provider",
		"request_id", requestID,
		"channel", channel,
		"provider", provider.GetProviderName(),
	)

	if s.automatedAlertService == nil {
		return fmt.Errorf("automated alert service not available")
	}

	s.automatedAlertService.RegisterNotificationProvider(channel, provider)

	s.logger.Info("Registered notification provider",
		"request_id", requestID,
		"channel", channel,
		"provider", provider.GetProviderName(),
	)

	return nil
}

// GetPaymentHistory retrieves payment history for a business
func (s *RiskService) GetPaymentHistory(ctx context.Context, businessID string) (*PaymentHistory, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Retrieving payment history",
		"request_id", requestID,
		"business_id", businessID,
	)

	if s.financialProviderManager == nil {
		return nil, fmt.Errorf("financial provider manager not available")
	}

	history, err := s.financialProviderManager.GetPaymentHistory(ctx, businessID)
	if err != nil {
		s.logger.Error("Failed to retrieve payment history",
			"request_id", requestID,
			"business_id", businessID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to retrieve payment history: %w", err)
	}

	s.logger.Info("Retrieved payment history successfully",
		"request_id", requestID,
		"business_id", businessID,
		"provider", history.Provider,
		"payment_rate", history.PaymentRate,
	)

	return history, nil
}

// GetIndustryBenchmarks retrieves industry benchmarks
func (s *RiskService) GetIndustryBenchmarks(ctx context.Context, industry string) (*IndustryBenchmarks, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Retrieving industry benchmarks",
		"request_id", requestID,
		"industry", industry,
	)

	if s.financialProviderManager == nil {
		return nil, fmt.Errorf("financial provider manager not available")
	}

	benchmarks, err := s.financialProviderManager.GetIndustryBenchmarks(ctx, industry)
	if err != nil {
		s.logger.Error("Failed to retrieve industry benchmarks",
			"request_id", requestID,
			"industry", industry,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to retrieve industry benchmarks: %w", err)
	}

	s.logger.Info("Retrieved industry benchmarks successfully",
		"request_id", requestID,
		"industry", industry,
		"provider", benchmarks.Provider,
	)

	return benchmarks, nil
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

// AnalyzeRiskTrends performs comprehensive trend analysis for a business
func (s *RiskService) AnalyzeRiskTrends(ctx context.Context, businessID string, period string) (*TrendAnalysisResult, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Analyzing risk trends",
		"request_id", requestID,
		"business_id", businessID,
		"period", period,
	)

	if s.trendAnalysisService == nil {
		return nil, fmt.Errorf("trend analysis service not available")
	}

	// Get historical assessments for trend analysis
	history, err := s.historyService.GetRiskHistory(ctx, businessID, 100, 0)
	if err != nil {
		s.logger.Error("Failed to get risk history for trend analysis",
			"request_id", requestID,
			"business_id", businessID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to get risk history: %w", err)
	}

	// Extract assessments from history
	var assessments []*RiskAssessment
	for _, entry := range history.History {
		if entry.Assessment != nil {
			assessments = append(assessments, entry.Assessment)
		}
	}

	// Perform trend analysis
	result, err := s.trendAnalysisService.AnalyzeTrends(ctx, businessID, assessments, period)
	if err != nil {
		s.logger.Error("Failed to analyze risk trends",
			"request_id", requestID,
			"business_id", businessID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to analyze risk trends: %w", err)
	}

	s.logger.Info("Risk trend analysis completed",
		"request_id", requestID,
		"business_id", businessID,
		"overall_trend", result.OverallTrend,
		"trend_strength", result.OverallTrendStrength,
		"anomaly_count", len(result.Anomalies),
		"recommendation_count", len(result.Recommendations),
	)

	return result, nil
}

// GetTrendPredictions gets trend predictions for a business
func (s *RiskService) GetTrendPredictions(ctx context.Context, businessID string, horizon time.Duration) ([]TrendPrediction, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Getting trend predictions",
		"request_id", requestID,
		"business_id", businessID,
		"horizon", horizon,
	)

	if s.trendAnalysisService == nil {
		return nil, fmt.Errorf("trend analysis service not available")
	}

	// Get historical assessments
	history, err := s.historyService.GetRiskHistory(ctx, businessID, 50, 0)
	if err != nil {
		s.logger.Error("Failed to get risk history for predictions",
			"request_id", requestID,
			"business_id", businessID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to get risk history: %w", err)
	}

	// Extract assessments
	var assessments []*RiskAssessment
	for _, entry := range history.History {
		if entry.Assessment != nil {
			assessments = append(assessments, entry.Assessment)
		}
	}

	if len(assessments) < 3 {
		return nil, fmt.Errorf("insufficient historical data for predictions (minimum 3 assessments required)")
	}

	// Generate predictions
	result, err := s.trendAnalysisService.AnalyzeTrends(ctx, businessID, assessments, "6months")
	if err != nil {
		s.logger.Error("Failed to generate trend predictions",
			"request_id", requestID,
			"business_id", businessID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to generate predictions: %w", err)
	}

	// Filter predictions by horizon
	var filteredPredictions []TrendPrediction
	for _, prediction := range result.Predictions {
		if prediction.Horizon == horizon {
			filteredPredictions = append(filteredPredictions, prediction)
		}
	}

	s.logger.Info("Trend predictions retrieved",
		"request_id", requestID,
		"business_id", businessID,
		"horizon", horizon,
		"prediction_count", len(filteredPredictions),
	)

	return filteredPredictions, nil
}

// GetTrendAnomalies gets trend anomalies for a business
func (s *RiskService) GetTrendAnomalies(ctx context.Context, businessID string) ([]TrendAnomaly, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Getting trend anomalies",
		"request_id", requestID,
		"business_id", businessID,
	)

	if s.trendAnalysisService == nil {
		return nil, fmt.Errorf("trend analysis service not available")
	}

	// Get historical assessments
	history, err := s.historyService.GetRiskHistory(ctx, businessID, 100, 0)
	if err != nil {
		s.logger.Error("Failed to get risk history for anomaly detection",
			"request_id", requestID,
			"business_id", businessID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to get risk history: %w", err)
	}

	// Extract assessments
	var assessments []*RiskAssessment
	for _, entry := range history.History {
		if entry.Assessment != nil {
			assessments = append(assessments, entry.Assessment)
		}
	}

	if len(assessments) < 3 {
		return []TrendAnomaly{}, nil // No anomalies if insufficient data
	}

	// Perform trend analysis to get anomalies
	result, err := s.trendAnalysisService.AnalyzeTrends(ctx, businessID, assessments, "6months")
	if err != nil {
		s.logger.Error("Failed to detect trend anomalies",
			"request_id", requestID,
			"business_id", businessID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to detect anomalies: %w", err)
	}

	s.logger.Info("Trend anomalies retrieved",
		"request_id", requestID,
		"business_id", businessID,
		"anomaly_count", len(result.Anomalies),
	)

	return result.Anomalies, nil
}

// GenerateAdvancedReport generates an advanced risk report
func (s *RiskService) GenerateAdvancedReport(ctx context.Context, request AdvancedReportRequest) (*AdvancedRiskReport, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Generating advanced risk report",
		"request_id", requestID,
		"business_id", request.BusinessID,
		"report_type", request.ReportType,
		"format", request.Format,
	)

	if s.reportingSystem == nil {
		return nil, fmt.Errorf("reporting system not available")
	}

	report, err := s.reportingSystem.GenerateAdvancedReport(ctx, request)
	if err != nil {
		s.logger.Error("Failed to generate advanced risk report",
			"request_id", requestID,
			"business_id", request.BusinessID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to generate advanced risk report: %w", err)
	}

	s.logger.Info("Advanced risk report generated successfully",
		"request_id", requestID,
		"business_id", request.BusinessID,
		"report_type", request.ReportType,
		"format", request.Format,
	)

	return report, nil
}
