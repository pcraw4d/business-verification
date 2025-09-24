package enrichment

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// CompanySizeClassifier combines employee count and revenue analysis to classify company size
type CompanySizeClassifier struct {
	config           *CompanySizeConfig
	employeeAnalyzer *EmployeeCountAnalyzer
	revenueAnalyzer  *RevenueAnalyzer
	logger           *zap.Logger
	tracer           trace.Tracer
}

// CompanySizeConfig contains configuration for company size classification
type CompanySizeConfig struct {
	// Classification settings
	EnableEmployeeAnalysis  bool `json:"enable_employee_analysis"`
	EnableRevenueAnalysis   bool `json:"enable_revenue_analysis"`
	EnableConfidenceScoring bool `json:"enable_confidence_scoring"`
	EnableValidation        bool `json:"enable_validation"`

	// Employee thresholds
	StartupMaxEmployees    int `json:"startup_max_employees"`    // ≤50 employees
	SMEMinEmployees        int `json:"sme_min_employees"`        // 51 employees
	SMEMaxEmployees        int `json:"sme_max_employees"`        // 250 employees
	EnterpriseMinEmployees int `json:"enterprise_min_employees"` // 251+ employees

	// Revenue thresholds (in USD)
	StartupMaxRevenue    int64 `json:"startup_max_revenue"`    // ≤$1M
	SMEMinRevenue        int64 `json:"sme_min_revenue"`        // $1M
	SMEMaxRevenue        int64 `json:"sme_max_revenue"`        // $10M
	EnterpriseMinRevenue int64 `json:"enterprise_min_revenue"` // $10M+

	// Weighting factors
	EmployeeWeight float64 `json:"employee_weight"` // Weight for employee data (0.0-1.0)
	RevenueWeight  float64 `json:"revenue_weight"`  // Weight for revenue data (0.0-1.0)

	// Confidence settings
	MinConfidenceThreshold float64 `json:"min_confidence_threshold"`
	RequireBothIndicators  bool    `json:"require_both_indicators"`

	// Quality settings
	EnableConsistencyCheck  bool `json:"enable_consistency_check"`
	EnableContextValidation bool `json:"enable_context_validation"`
}

// CompanySizeResult contains the results of company size classification
type CompanySizeResult struct {
	// Classification results
	CompanySize     string  `json:"company_size"`     // startup, sme, enterprise, unknown
	ConfidenceScore float64 `json:"confidence_score"` // 0.0-1.0
	Classification  string  `json:"classification"`   // Detailed classification
	SizeCategory    string  `json:"size_category"`    // Category with subcategories

	// Input data analysis
	EmployeeAnalysis *EmployeeCountResult `json:"employee_analysis,omitempty"`
	RevenueAnalysis  *RevenueResult       `json:"revenue_analysis,omitempty"`

	// Classification details
	EmployeeClassification string  `json:"employee_classification"`
	RevenueClassification  string  `json:"revenue_classification"`
	ConsistencyScore       float64 `json:"consistency_score"`
	DataQualityScore       float64 `json:"data_quality_score"`

	// Evidence and reasoning
	Evidence            []string `json:"evidence"`
	Reasoning           string   `json:"reasoning"`
	ClassificationBasis []string `json:"classification_basis"`

	// Weights used in classification
	EmployeeWeight float64 `json:"employee_weight"`
	RevenueWeight  float64 `json:"revenue_weight"`

	// Quality metrics
	DataQuality      DataQualityMetrics `json:"data_quality"`
	ValidationStatus ValidationStatus   `json:"validation_status"`
	IsValidated      bool               `json:"is_validated"`

	// Metadata
	ClassifiedAt time.Time              `json:"classified_at"`
	SourceURL    string                 `json:"source_url"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// CompanySizeDistribution represents size distribution analysis
type CompanySizeDistribution struct {
	StartupProbability    float64 `json:"startup_probability"`
	SMEProbability        float64 `json:"sme_probability"`
	EnterpriseProbability float64 `json:"enterprise_probability"`
	UnknownProbability    float64 `json:"unknown_probability"`
}

// NewCompanySizeClassifier creates a new company size classifier with default configuration
func NewCompanySizeClassifier(config *CompanySizeConfig, employeeAnalyzer *EmployeeCountAnalyzer, revenueAnalyzer *RevenueAnalyzer, logger *zap.Logger) *CompanySizeClassifier {
	if config == nil {
		config = &CompanySizeConfig{
			EnableEmployeeAnalysis:  true,
			EnableRevenueAnalysis:   true,
			EnableConfidenceScoring: true,
			EnableValidation:        true,

			StartupMaxEmployees:    50,
			SMEMinEmployees:        51,
			SMEMaxEmployees:        250,
			EnterpriseMinEmployees: 251,

			StartupMaxRevenue:    1000000,  // $1M
			SMEMinRevenue:        1000000,  // $1M
			SMEMaxRevenue:        10000000, // $10M
			EnterpriseMinRevenue: 10000000, // $10M+

			EmployeeWeight: 0.6, // Slightly favor employee data
			RevenueWeight:  0.4,

			MinConfidenceThreshold: 0.3,
			RequireBothIndicators:  false,

			EnableConsistencyCheck:  true,
			EnableContextValidation: true,
		}
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	return &CompanySizeClassifier{
		config:           config,
		employeeAnalyzer: employeeAnalyzer,
		revenueAnalyzer:  revenueAnalyzer,
		logger:           logger,
		tracer:           otel.Tracer("company-size-classifier"),
	}
}

// ClassifyCompanySize analyzes website content to classify company size
func (csc *CompanySizeClassifier) ClassifyCompanySize(ctx context.Context, content string) (*CompanySizeResult, error) {
	ctx, span := csc.tracer.Start(ctx, "company_size_classifier.classify",
		trace.WithAttributes(
			attribute.String("content_length", fmt.Sprintf("%d", len(content))),
			attribute.Bool("enable_employee_analysis", csc.config.EnableEmployeeAnalysis),
			attribute.Bool("enable_revenue_analysis", csc.config.EnableRevenueAnalysis),
		))
	defer span.End()

	csc.logger.Info("Starting company size classification",
		zap.String("content_length", fmt.Sprintf("%d", len(content))),
		zap.Bool("enable_employee_analysis", csc.config.EnableEmployeeAnalysis),
		zap.Bool("enable_revenue_analysis", csc.config.EnableRevenueAnalysis))

	result := &CompanySizeResult{
		ClassifiedAt:   time.Now(),
		Metadata:       make(map[string]interface{}),
		EmployeeWeight: csc.config.EmployeeWeight,
		RevenueWeight:  csc.config.RevenueWeight,
	}

	// Analyze employee count if enabled
	if csc.config.EnableEmployeeAnalysis && csc.employeeAnalyzer != nil {
		if employeeResult, err := csc.employeeAnalyzer.AnalyzeEmployeeCount(ctx, content, ""); err != nil {
			csc.logger.Error("Failed to analyze employee count", zap.Error(err))
			span.RecordError(err)
		} else {
			result.EmployeeAnalysis = employeeResult
			result.EmployeeClassification = csc.classifyByEmployees(employeeResult.EmployeeCount)
			result.Evidence = append(result.Evidence, employeeResult.Evidence...)
			result.ClassificationBasis = append(result.ClassificationBasis, "employee_count")
		}
	}

	// Analyze revenue if enabled
	if csc.config.EnableRevenueAnalysis && csc.revenueAnalyzer != nil {
		if revenueResult, err := csc.revenueAnalyzer.AnalyzeContent(ctx, content); err != nil {
			csc.logger.Error("Failed to analyze revenue", zap.Error(err))
			span.RecordError(err)
		} else {
			result.RevenueAnalysis = revenueResult
			result.RevenueClassification = csc.classifyByRevenue(revenueResult.RevenueAmount)
			result.Evidence = append(result.Evidence, revenueResult.Evidence...)
			result.ClassificationBasis = append(result.ClassificationBasis, "revenue_analysis")
		}
	}

	// Perform unified classification
	if err := csc.performUnifiedClassification(result); err != nil {
		csc.logger.Error("Failed to perform unified classification", zap.Error(err))
		span.RecordError(err)
		return nil, fmt.Errorf("unified classification failed: %w", err)
	}

	// Calculate confidence score
	if csc.config.EnableConfidenceScoring {
		result.ConfidenceScore = csc.calculateConfidence(result)
	}

	// Validate result
	if csc.config.EnableValidation {
		if err := csc.validateResult(result); err != nil {
			csc.logger.Error("Failed to validate result", zap.Error(err))
			span.RecordError(err)
		}
	}

	// Generate reasoning
	result.Reasoning = csc.generateReasoning(result)

	csc.logger.Info("Company size classification completed",
		zap.String("company_size", result.CompanySize),
		zap.String("classification", result.Classification),
		zap.Float64("confidence_score", result.ConfidenceScore))

	return result, nil
}

// classifyByEmployees classifies company size based on employee count
func (csc *CompanySizeClassifier) classifyByEmployees(employeeCount int) string {
	if employeeCount == 0 {
		return "unknown"
	}

	if employeeCount <= csc.config.StartupMaxEmployees {
		return "startup"
	} else if employeeCount <= csc.config.SMEMaxEmployees {
		return "sme"
	} else {
		return "enterprise"
	}
}

// classifyByRevenue classifies company size based on revenue
func (csc *CompanySizeClassifier) classifyByRevenue(revenue int64) string {
	if revenue == 0 {
		return "unknown"
	}

	if revenue <= csc.config.StartupMaxRevenue {
		return "startup"
	} else if revenue <= csc.config.SMEMaxRevenue {
		return "sme"
	} else {
		return "enterprise"
	}
}

// performUnifiedClassification combines employee and revenue classifications
func (csc *CompanySizeClassifier) performUnifiedClassification(result *CompanySizeResult) error {
	// Check if we have any classification data
	hasEmployeeData := result.EmployeeClassification != "" && result.EmployeeClassification != "unknown"
	hasRevenueData := result.RevenueClassification != "" && result.RevenueClassification != "unknown"

	if !hasEmployeeData && !hasRevenueData {
		result.CompanySize = "unknown"
		result.Classification = "Unknown - Insufficient data"
		result.SizeCategory = "unknown"
		return nil
	}

	// If only one type of data is available
	if hasEmployeeData && !hasRevenueData {
		result.CompanySize = result.EmployeeClassification
		result.Classification = fmt.Sprintf("%s (based on employee count)", strings.Title(result.EmployeeClassification))
		result.SizeCategory = result.EmployeeClassification
		return nil
	}

	if hasRevenueData && !hasEmployeeData {
		result.CompanySize = result.RevenueClassification
		result.Classification = fmt.Sprintf("%s (based on revenue)", strings.Title(result.RevenueClassification))
		result.SizeCategory = result.RevenueClassification
		return nil
	}

	// Both types of data are available - perform weighted classification
	result.ConsistencyScore = csc.calculateConsistency(result.EmployeeClassification, result.RevenueClassification)

	// Calculate weighted scores
	employeeScore := csc.getClassificationScore(result.EmployeeClassification)
	revenueScore := csc.getClassificationScore(result.RevenueClassification)

	weightedScore := (employeeScore * csc.config.EmployeeWeight) + (revenueScore * csc.config.RevenueWeight)

	// Determine final classification based on weighted score
	if weightedScore <= 1.5 {
		result.CompanySize = "startup"
	} else if weightedScore <= 2.5 {
		result.CompanySize = "sme"
	} else {
		result.CompanySize = "enterprise"
	}

	// Generate detailed classification
	if result.EmployeeClassification == result.RevenueClassification {
		result.Classification = fmt.Sprintf("%s (consistent across employee and revenue data)", strings.Title(result.CompanySize))
	} else {
		result.Classification = fmt.Sprintf("%s (weighted: %s by employees, %s by revenue)",
			strings.Title(result.CompanySize),
			result.EmployeeClassification,
			result.RevenueClassification)
	}

	result.SizeCategory = result.CompanySize

	return nil
}

// getClassificationScore converts classification to numeric score for weighting
func (csc *CompanySizeClassifier) getClassificationScore(classification string) float64 {
	switch classification {
	case "startup":
		return 1.0
	case "sme":
		return 2.0
	case "enterprise":
		return 3.0
	default:
		return 0.0
	}
}

// calculateConsistency calculates consistency between employee and revenue classifications
func (csc *CompanySizeClassifier) calculateConsistency(employeeClass, revenueClass string) float64 {
	if employeeClass == revenueClass {
		return 1.0 // Perfect consistency
	}

	// Define consistency matrix
	consistencyMatrix := map[string]map[string]float64{
		"startup": {
			"sme":        0.7, // Somewhat consistent
			"enterprise": 0.3, // Low consistency
		},
		"sme": {
			"startup":    0.7, // Somewhat consistent
			"enterprise": 0.8, // Good consistency (growth companies)
		},
		"enterprise": {
			"startup": 0.3, // Low consistency
			"sme":     0.8, // Good consistency
		},
	}

	if matrix, exists := consistencyMatrix[employeeClass]; exists {
		if score, exists := matrix[revenueClass]; exists {
			return score
		}
	}

	return 0.5 // Default moderate consistency
}

// calculateConfidence calculates overall confidence score for the classification
func (csc *CompanySizeClassifier) calculateConfidence(result *CompanySizeResult) float64 {
	confidence := 0.0

	// Base confidence from data availability
	hasEmployeeData := result.EmployeeAnalysis != nil && result.EmployeeAnalysis.EmployeeCount > 0
	hasRevenueData := result.RevenueAnalysis != nil && result.RevenueAnalysis.RevenueAmount > 0

	if hasEmployeeData && hasRevenueData {
		confidence += 0.4 // Both data types available
	} else if hasEmployeeData || hasRevenueData {
		confidence += 0.25 // One data type available
	}

	// Confidence from individual analyses
	if hasEmployeeData {
		confidence += result.EmployeeAnalysis.ConfidenceScore * csc.config.EmployeeWeight * 0.5
	}

	if hasRevenueData {
		confidence += result.RevenueAnalysis.ConfidenceScore * csc.config.RevenueWeight * 0.5
	}

	// Consistency bonus
	if hasEmployeeData && hasRevenueData {
		confidence += result.ConsistencyScore * 0.2
	}

	// Evidence quality
	if len(result.Evidence) > 0 {
		evidenceScore := float64(len(result.Evidence)) * 0.05
		if evidenceScore > 0.15 {
			evidenceScore = 0.15
		}
		confidence += evidenceScore
	}

	// Validation bonus
	if result.IsValidated {
		confidence += 0.05
	}

	// Cap at 1.0
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// validateResult validates the classification result
func (csc *CompanySizeClassifier) validateResult(result *CompanySizeResult) error {
	// Check minimum confidence threshold
	if result.ConfidenceScore < csc.config.MinConfidenceThreshold {
		return fmt.Errorf("confidence score %f below threshold %f",
			result.ConfidenceScore, csc.config.MinConfidenceThreshold)
	}

	// Validate company size
	validSizes := []string{"startup", "sme", "enterprise", "unknown"}
	isValid := false
	for _, size := range validSizes {
		if result.CompanySize == size {
			isValid = true
			break
		}
	}

	if !isValid {
		return fmt.Errorf("invalid company size: %s", result.CompanySize)
	}

	// Check if both indicators are required
	if csc.config.RequireBothIndicators {
		hasEmployeeData := result.EmployeeAnalysis != nil && result.EmployeeAnalysis.EmployeeCount > 0
		hasRevenueData := result.RevenueAnalysis != nil && result.RevenueAnalysis.RevenueAmount > 0

		if !hasEmployeeData || !hasRevenueData {
			return fmt.Errorf("both employee and revenue indicators required")
		}
	}

	result.IsValidated = true
	return nil
}

// generateReasoning generates human-readable reasoning for the classification
func (csc *CompanySizeClassifier) generateReasoning(result *CompanySizeResult) string {
	reasoning := ""

	// Classification result
	reasoning += fmt.Sprintf("Company classified as %s with %d%% confidence. ",
		strings.Title(result.CompanySize), int(result.ConfidenceScore*100))

	// Data sources
	if result.EmployeeAnalysis != nil && result.EmployeeAnalysis.EmployeeCount > 0 {
		reasoning += fmt.Sprintf("Employee analysis indicates %d employees suggesting %s classification. ",
			result.EmployeeAnalysis.EmployeeCount, result.EmployeeClassification)
	}

	if result.RevenueAnalysis != nil && result.RevenueAnalysis.RevenueAmount > 0 {
		reasoning += fmt.Sprintf("Revenue analysis indicates $%d annual revenue suggesting %s classification. ",
			result.RevenueAnalysis.RevenueAmount, result.RevenueClassification)
	}

	// Consistency analysis
	if result.EmployeeClassification != "" && result.RevenueClassification != "" {
		if result.EmployeeClassification == result.RevenueClassification {
			reasoning += "Employee and revenue data are consistent. "
		} else {
			reasoning += fmt.Sprintf("Employee and revenue data show different classifications with %d%% consistency. ",
				int(result.ConsistencyScore*100))
		}
	}

	// Weighting information
	if len(result.ClassificationBasis) > 1 {
		reasoning += fmt.Sprintf("Final classification uses %.0f%% employee weight and %.0f%% revenue weight. ",
			csc.config.EmployeeWeight*100, csc.config.RevenueWeight*100)
	}

	// Evidence summary
	if len(result.Evidence) > 0 {
		reasoning += fmt.Sprintf("Classification based on %d pieces of evidence. ", len(result.Evidence))
	}

	return reasoning
}

// GetSizeDistribution calculates probability distribution across size categories
func (csc *CompanySizeClassifier) GetSizeDistribution(result *CompanySizeResult) *CompanySizeDistribution {
	distribution := &CompanySizeDistribution{}

	// Base probabilities from confidence score
	baseConfidence := result.ConfidenceScore

	switch result.CompanySize {
	case "startup":
		distribution.StartupProbability = baseConfidence
		distribution.SMEProbability = (1.0 - baseConfidence) * 0.7
		distribution.EnterpriseProbability = (1.0 - baseConfidence) * 0.2
		distribution.UnknownProbability = (1.0 - baseConfidence) * 0.1
	case "sme":
		distribution.SMEProbability = baseConfidence
		distribution.StartupProbability = (1.0 - baseConfidence) * 0.4
		distribution.EnterpriseProbability = (1.0 - baseConfidence) * 0.5
		distribution.UnknownProbability = (1.0 - baseConfidence) * 0.1
	case "enterprise":
		distribution.EnterpriseProbability = baseConfidence
		distribution.SMEProbability = (1.0 - baseConfidence) * 0.6
		distribution.StartupProbability = (1.0 - baseConfidence) * 0.2
		distribution.UnknownProbability = (1.0 - baseConfidence) * 0.2
	default:
		distribution.UnknownProbability = 1.0
	}

	return distribution
}
