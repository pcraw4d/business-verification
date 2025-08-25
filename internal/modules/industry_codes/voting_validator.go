package industry_codes

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"go.uber.org/zap"
)

// VotingValidationConfig defines configuration for voting result validation
type VotingValidationConfig struct {
	// Result validation settings
	MinResultCount          int     `json:"min_result_count"`
	MaxResultCount          int     `json:"max_result_count"`
	MinConfidenceThreshold  float64 `json:"min_confidence_threshold"`
	MaxConfidenceThreshold  float64 `json:"max_confidence_threshold"`
	MinVotingScoreThreshold float64 `json:"min_voting_score_threshold"`
	MinAgreementThreshold   float64 `json:"min_agreement_threshold"`
	MinConsistencyThreshold float64 `json:"min_consistency_threshold"`

	// Consistency check settings
	MaxConfidenceVariance float64 `json:"max_confidence_variance"`
	MinStrategyAgreement  float64 `json:"min_strategy_agreement"`
	MaxResultSpread       float64 `json:"max_result_spread"`

	// Quality assurance settings
	EnableAnomalyDetection   bool    `json:"enable_anomaly_detection"`
	AnomalyThreshold         float64 `json:"anomaly_threshold"`
	EnableCrossValidation    bool    `json:"enable_cross_validation"`
	CrossValidationThreshold float64 `json:"cross_validation_threshold"`

	// Code format validation
	ValidateCodeFormats      bool `json:"validate_code_formats"`
	ValidateCodeTypes        bool `json:"validate_code_types"`
	ValidateCodeDescriptions bool `json:"validate_code_descriptions"`

	// Advanced validation settings
	EnableStatisticalValidation bool          `json:"enable_statistical_validation"`
	StatisticalSignificance     float64       `json:"statistical_significance"`
	EnableTemporalValidation    bool          `json:"enable_temporal_validation"`
	TemporalWindow              time.Duration `json:"temporal_window"`
}

// VotingValidationResult represents the result of voting validation
type VotingValidationResult struct {
	IsValid           bool                     `json:"is_valid"`
	ValidationScore   float64                  `json:"validation_score"`
	Issues            []*VotingValidationIssue `json:"issues"`
	Warnings          []*ValidationWarning     `json:"warnings"`
	QualityMetrics    *VotingQualityMetrics    `json:"quality_metrics"`
	ConsistencyChecks *ConsistencyCheckResult  `json:"consistency_checks"`
	Recommendations   []string                 `json:"recommendations"`
	ValidationTime    time.Time                `json:"validation_time"`
}

// VotingValidationIssue represents a validation issue that affects result validity
type VotingValidationIssue struct {
	Type           string `json:"type"`
	Severity       string `json:"severity"` // "critical", "error", "warning"
	Message        string `json:"message"`
	Field          string `json:"field"`
	Value          string `json:"value"`
	Expected       string `json:"expected"`
	Recommendation string `json:"recommendation"`
}

// ValidationWarning represents a warning that doesn't affect validity
type ValidationWarning struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Field   string `json:"field"`
	Value   string `json:"value"`
	Impact  string `json:"impact"`
}

// VotingQualityMetrics provides detailed quality assessment
type VotingQualityMetrics struct {
	ResultCompleteness    float64 `json:"result_completeness"`
	ConfidenceReliability float64 `json:"confidence_reliability"`
	StrategyConsistency   float64 `json:"strategy_consistency"`
	CodeFormatCompliance  float64 `json:"code_format_compliance"`
	OverallQuality        float64 `json:"overall_quality"`
}

// ConsistencyCheckResult provides consistency analysis
type ConsistencyCheckResult struct {
	CrossStrategyAgreement float64             `json:"cross_strategy_agreement"`
	ConfidenceConsistency  float64             `json:"confidence_consistency"`
	ResultStability        float64             `json:"result_stability"`
	AnomalyScore           float64             `json:"anomaly_score"`
	ConsistencyIssues      []*ConsistencyIssue `json:"consistency_issues"`
}

// ConsistencyIssue represents a consistency-related issue
type ConsistencyIssue struct {
	Type        string   `json:"type"`
	Severity    string   `json:"severity"`
	Description string   `json:"description"`
	Impact      float64  `json:"impact"`
	Strategies  []string `json:"strategies"`
}

// VotingValidator provides comprehensive validation for voting results
type VotingValidator struct {
	config *VotingValidationConfig
	logger *zap.Logger
}

// NewVotingValidator creates a new voting validator
func NewVotingValidator(config *VotingValidationConfig, logger *zap.Logger) *VotingValidator {
	if config == nil {
		config = &VotingValidationConfig{
			MinResultCount:              1,
			MaxResultCount:              10,
			MinConfidenceThreshold:      0.0,
			MaxConfidenceThreshold:      1.0,
			MinVotingScoreThreshold:     0.3,
			MinAgreementThreshold:       0.4,
			MinConsistencyThreshold:     0.5,
			MaxConfidenceVariance:       0.3,
			MinStrategyAgreement:        0.6,
			MaxResultSpread:             0.5,
			EnableAnomalyDetection:      true,
			AnomalyThreshold:            2.0,
			EnableCrossValidation:       true,
			CrossValidationThreshold:    0.7,
			ValidateCodeFormats:         true,
			ValidateCodeTypes:           true,
			ValidateCodeDescriptions:    true,
			EnableStatisticalValidation: true,
			StatisticalSignificance:     0.05,
			EnableTemporalValidation:    true,
			TemporalWindow:              5 * time.Minute,
		}
	}

	return &VotingValidator{
		config: config,
		logger: logger,
	}
}

// ValidateVotingResult performs comprehensive validation of voting results
func (vv *VotingValidator) ValidateVotingResult(ctx context.Context, result *VotingResult, votes []*StrategyVote) (*VotingValidationResult, error) {
	startTime := time.Now()

	resultCount := 0
	if result != nil && result.FinalResults != nil {
		resultCount = len(result.FinalResults)
	}
	voteCount := 0
	if votes != nil {
		voteCount = len(votes)
	}

	vv.logger.Info("Starting voting result validation",
		zap.Int("result_count", resultCount),
		zap.Int("vote_count", voteCount))

	validationResult := &VotingValidationResult{
		IsValid:         true,
		ValidationScore: 1.0,
		Issues:          []*VotingValidationIssue{},
		Warnings:        []*ValidationWarning{},
		ValidationTime:  startTime,
	}

	// Validate input parameters
	if err := vv.validateInputs(result, votes); err != nil {
		validationResult.IsValid = false
		validationResult.ValidationScore = 0.0
		validationResult.Issues = append(validationResult.Issues, &VotingValidationIssue{
			Type:           "input_validation",
			Severity:       "critical",
			Message:        fmt.Sprintf("Input validation failed: %v", err),
			Field:          "inputs",
			Value:          "invalid",
			Expected:       "valid inputs",
			Recommendation: "Check input parameters and ensure they are properly formatted",
		})
		return validationResult, nil
	}

	// Perform basic result validation
	vv.validateBasicResults(result, validationResult)

	// Perform quality metrics calculation
	qualityMetrics := vv.calculateQualityMetrics(result, votes)
	validationResult.QualityMetrics = qualityMetrics

	// Perform consistency checks
	consistencyChecks := vv.performConsistencyChecks(result, votes)
	validationResult.ConsistencyChecks = consistencyChecks

	// Perform anomaly detection if enabled
	if vv.config.EnableAnomalyDetection {
		vv.detectAnomalies(result, votes, validationResult)
	}

	// Perform cross-validation if enabled
	if vv.config.EnableCrossValidation {
		vv.performCrossValidation(result, votes, validationResult)
	}

	// Perform statistical validation if enabled
	if vv.config.EnableStatisticalValidation {
		vv.performStatisticalValidation(result, votes, validationResult)
	}

	// Perform temporal validation if enabled
	if vv.config.EnableTemporalValidation {
		vv.performTemporalValidation(result, votes, validationResult)
	}

	// Calculate overall validation score
	validationResult.ValidationScore = vv.calculateOverallValidationScore(validationResult)

	// Determine if result is valid based on issues
	// If there are critical issues (like insufficient results), the result is always invalid
	hasCriticalIssues := false
	for _, issue := range validationResult.Issues {
		if issue.Severity == "critical" || issue.Type == "result_count" {
			hasCriticalIssues = true
			break
		}
	}

	validationResult.IsValid = !hasCriticalIssues && (len(validationResult.Issues) == 0 ||
		(len(validationResult.Issues) > 0 && validationResult.ValidationScore >= vv.config.MinVotingScoreThreshold))

	// Generate recommendations
	validationResult.Recommendations = vv.generateRecommendations(validationResult)

	vv.logger.Info("Voting result validation completed",
		zap.Bool("is_valid", validationResult.IsValid),
		zap.Float64("validation_score", validationResult.ValidationScore),
		zap.Int("issues", len(validationResult.Issues)),
		zap.Int("warnings", len(validationResult.Warnings)),
		zap.Duration("validation_time", time.Since(startTime)))

	return validationResult, nil
}

// validateInputs validates the input parameters
func (vv *VotingValidator) validateInputs(result *VotingResult, votes []*StrategyVote) error {
	if result == nil {
		return fmt.Errorf("voting result is nil")
	}

	if votes == nil {
		return fmt.Errorf("votes slice is nil")
	}

	if len(votes) == 0 {
		return fmt.Errorf("no votes provided")
	}

	if result.FinalResults == nil {
		return fmt.Errorf("final results is nil")
	}

	return nil
}

// validateBasicResults performs basic validation of voting results
func (vv *VotingValidator) validateBasicResults(result *VotingResult, validationResult *VotingValidationResult) {
	// Validate result count
	if len(result.FinalResults) < vv.config.MinResultCount {
		validationResult.Issues = append(validationResult.Issues, &VotingValidationIssue{
			Type:           "result_count",
			Severity:       "error",
			Message:        fmt.Sprintf("Insufficient results: got %d, minimum required %d", len(result.FinalResults), vv.config.MinResultCount),
			Field:          "final_results",
			Value:          fmt.Sprintf("%d", len(result.FinalResults)),
			Expected:       fmt.Sprintf(">= %d", vv.config.MinResultCount),
			Recommendation: "Increase the number of classification strategies or adjust minimum result count",
		})
	}

	if len(result.FinalResults) > vv.config.MaxResultCount {
		validationResult.Warnings = append(validationResult.Warnings, &ValidationWarning{
			Type:    "result_count",
			Message: fmt.Sprintf("Too many results: got %d, maximum expected %d", len(result.FinalResults), vv.config.MaxResultCount),
			Field:   "final_results",
			Value:   fmt.Sprintf("%d", len(result.FinalResults)),
			Impact:  "May indicate over-classification or noise in results",
		})
	}

	// Validate voting score
	if result.VotingScore < vv.config.MinVotingScoreThreshold {
		validationResult.Issues = append(validationResult.Issues, &VotingValidationIssue{
			Type:           "voting_score",
			Severity:       "error",
			Message:        fmt.Sprintf("Voting score below threshold: got %.3f, minimum required %.3f", result.VotingScore, vv.config.MinVotingScoreThreshold),
			Field:          "voting_score",
			Value:          fmt.Sprintf("%.3f", result.VotingScore),
			Expected:       fmt.Sprintf(">= %.3f", vv.config.MinVotingScoreThreshold),
			Recommendation: "Review voting strategy configuration or improve classification strategies",
		})
	}

	// Validate agreement score
	if result.Agreement < vv.config.MinAgreementThreshold {
		validationResult.Issues = append(validationResult.Issues, &VotingValidationIssue{
			Type:           "agreement",
			Severity:       "error",
			Message:        fmt.Sprintf("Agreement score below threshold: got %.3f, minimum required %.3f", result.Agreement, vv.config.MinAgreementThreshold),
			Field:          "agreement",
			Value:          fmt.Sprintf("%.3f", result.Agreement),
			Expected:       fmt.Sprintf(">= %.3f", vv.config.MinAgreementThreshold),
			Recommendation: "Strategies are not agreeing on results, consider adjusting strategy weights or improving individual strategies",
		})
	}

	// Validate consistency score
	if result.Consistency < vv.config.MinConsistencyThreshold {
		validationResult.Issues = append(validationResult.Issues, &VotingValidationIssue{
			Type:           "consistency",
			Severity:       "error",
			Message:        fmt.Sprintf("Consistency score below threshold: got %.3f, minimum required %.3f", result.Consistency, vv.config.MinConsistencyThreshold),
			Field:          "consistency",
			Value:          fmt.Sprintf("%.3f", result.Consistency),
			Expected:       fmt.Sprintf(">= %.3f", vv.config.MinConsistencyThreshold),
			Recommendation: "Results are inconsistent across strategies, review strategy configuration and data quality",
		})
	}

	// Validate individual result confidences
	for i, res := range result.FinalResults {
		if res.Confidence < vv.config.MinConfidenceThreshold {
			validationResult.Warnings = append(validationResult.Warnings, &ValidationWarning{
				Type:    "confidence",
				Message: fmt.Sprintf("Low confidence for result %d: %.3f", i, res.Confidence),
				Field:   fmt.Sprintf("final_results[%d].confidence", i),
				Value:   fmt.Sprintf("%.3f", res.Confidence),
				Impact:  "May indicate uncertain classification",
			})
		}

		if res.Confidence > vv.config.MaxConfidenceThreshold {
			validationResult.Warnings = append(validationResult.Warnings, &ValidationWarning{
				Type:    "confidence",
				Message: fmt.Sprintf("Unusually high confidence for result %d: %.3f", i, res.Confidence),
				Field:   fmt.Sprintf("final_results[%d].confidence", i),
				Value:   fmt.Sprintf("%.3f", res.Confidence),
				Impact:  "May indicate overfitting or bias in classification",
			})
		}
	}
}

// calculateQualityMetrics calculates comprehensive quality metrics
func (vv *VotingValidator) calculateQualityMetrics(result *VotingResult, votes []*StrategyVote) *VotingQualityMetrics {
	metrics := &VotingQualityMetrics{}

	// Calculate result completeness
	metrics.ResultCompleteness = vv.calculateResultCompleteness(result, votes)

	// Calculate confidence reliability
	metrics.ConfidenceReliability = vv.calculateConfidenceReliability(result, votes)

	// Calculate strategy consistency
	metrics.StrategyConsistency = vv.calculateStrategyConsistency(result, votes)

	// Calculate code format compliance
	metrics.CodeFormatCompliance = vv.calculateCodeFormatCompliance(result)

	// Calculate overall quality
	metrics.OverallQuality = (metrics.ResultCompleteness + metrics.ConfidenceReliability +
		metrics.StrategyConsistency + metrics.CodeFormatCompliance) / 4.0

	return metrics
}

// calculateResultCompleteness measures how complete the voting results are
func (vv *VotingValidator) calculateResultCompleteness(result *VotingResult, votes []*StrategyVote) float64 {
	if len(votes) == 0 {
		return 0.0
	}

	// Count strategies that contributed to final results
	contributingStrategies := make(map[string]bool)
	for _, vote := range votes {
		if len(vote.Results) > 0 {
			contributingStrategies[vote.StrategyName] = true
		}
	}

	completeness := float64(len(contributingStrategies)) / float64(len(votes))
	return math.Min(completeness, 1.0)
}

// calculateConfidenceReliability measures the reliability of confidence scores
func (vv *VotingValidator) calculateConfidenceReliability(result *VotingResult, votes []*StrategyVote) float64 {
	if len(result.FinalResults) == 0 {
		return 0.0
	}

	// Calculate confidence variance across strategies for each result
	var totalReliability float64
	for _, res := range result.FinalResults {
		var confidences []float64
		for _, vote := range votes {
			for _, voteRes := range vote.Results {
				if voteRes.Code != nil && res.Code != nil &&
					voteRes.Code.Type == res.Code.Type && voteRes.Code.Code == res.Code.Code {
					confidences = append(confidences, voteRes.Confidence)
				}
			}
		}

		if len(confidences) > 1 {
			mean := vv.calculateMean(confidences)
			variance := vv.calculateVariance(confidences, mean)
			reliability := 1.0 - math.Min(variance, 1.0)
			totalReliability += reliability
		} else {
			totalReliability += 0.5 // Default reliability for single confidence
		}
	}

	return totalReliability / float64(len(result.FinalResults))
}

// calculateStrategyConsistency measures consistency across strategies
func (vv *VotingValidator) calculateStrategyConsistency(result *VotingResult, votes []*StrategyVote) float64 {
	if len(votes) < 2 {
		return 1.0 // Perfect consistency for single strategy
	}

	// Calculate agreement on top results across strategies
	var agreements []float64
	for _, res := range result.FinalResults {
		agreement := 0.0
		for _, vote := range votes {
			for _, voteRes := range vote.Results {
				if voteRes.Code != nil && res.Code != nil &&
					voteRes.Code.Type == res.Code.Type && voteRes.Code.Code == res.Code.Code {
					agreement += 1.0
					break
				}
			}
		}
		agreements = append(agreements, agreement/float64(len(votes)))
	}

	if len(agreements) == 0 {
		return 0.0
	}

	return vv.calculateMean(agreements)
}

// calculateCodeFormatCompliance validates code format compliance
func (vv *VotingValidator) calculateCodeFormatCompliance(result *VotingResult) float64 {
	if len(result.FinalResults) == 0 {
		return 0.0
	}

	compliantCodes := 0
	for _, res := range result.FinalResults {
		if res.Code != nil && vv.isValidCodeFormat(res.Code) {
			compliantCodes++
		}
	}

	return float64(compliantCodes) / float64(len(result.FinalResults))
}

// isValidCodeFormat validates if a code has proper format
func (vv *VotingValidator) isValidCodeFormat(code *IndustryCode) bool {
	if code == nil {
		return false
	}

	// Validate code type
	if code.Type == "" {
		return false
	}

	// Validate code value
	if code.Code == "" {
		return false
	}

	// Validate description
	if vv.config.ValidateCodeDescriptions && code.Description == "" {
		return false
	}

	return true
}

// performConsistencyChecks performs comprehensive consistency analysis
func (vv *VotingValidator) performConsistencyChecks(result *VotingResult, votes []*StrategyVote) *ConsistencyCheckResult {
	consistency := &ConsistencyCheckResult{
		ConsistencyIssues: []*ConsistencyIssue{},
	}

	// Calculate cross-strategy agreement
	consistency.CrossStrategyAgreement = vv.calculateCrossStrategyAgreement(result, votes)

	// Calculate confidence consistency
	consistency.ConfidenceConsistency = vv.calculateConfidenceConsistency(result, votes)

	// Calculate result stability
	consistency.ResultStability = vv.calculateResultStability(result, votes)

	// Calculate anomaly score
	consistency.AnomalyScore = vv.calculateAnomalyScore(result, votes)

	// Identify consistency issues
	vv.identifyConsistencyIssues(result, votes, consistency)

	return consistency
}

// calculateCrossStrategyAgreement measures agreement across different strategies
func (vv *VotingValidator) calculateCrossStrategyAgreement(result *VotingResult, votes []*StrategyVote) float64 {
	if len(votes) < 2 {
		return 1.0
	}

	// Create strategy result maps
	strategyResults := make(map[string]map[string]*ClassificationResult)
	for _, vote := range votes {
		strategyResults[vote.StrategyName] = make(map[string]*ClassificationResult)
		for _, res := range vote.Results {
			if res.Code != nil {
				key := fmt.Sprintf("%s:%s", res.Code.Type, res.Code.Code)
				strategyResults[vote.StrategyName][key] = res
			}
		}
	}

	// Calculate pairwise agreement between strategies
	var agreements []float64
	strategies := make([]string, 0, len(strategyResults))
	for strategy := range strategyResults {
		strategies = append(strategies, strategy)
	}

	for i := 0; i < len(strategies); i++ {
		for j := i + 1; j < len(strategies); j++ {
			agreement := vv.calculatePairwiseAgreement(strategyResults[strategies[i]], strategyResults[strategies[j]])
			agreements = append(agreements, agreement)
		}
	}

	if len(agreements) == 0 {
		return 0.0
	}

	return vv.calculateMean(agreements)
}

// calculatePairwiseAgreement calculates agreement between two strategy result sets
func (vv *VotingValidator) calculatePairwiseAgreement(results1, results2 map[string]*ClassificationResult) float64 {
	if len(results1) == 0 && len(results2) == 0 {
		return 1.0
	}

	if len(results1) == 0 || len(results2) == 0 {
		return 0.0
	}

	// Find common codes
	commonCodes := 0
	totalCodes := len(results1) + len(results2)

	for key := range results1 {
		if _, exists := results2[key]; exists {
			commonCodes++
		}
	}

	// Calculate Jaccard similarity
	return float64(2*commonCodes) / float64(totalCodes)
}

// calculateConfidenceConsistency measures consistency of confidence scores
func (vv *VotingValidator) calculateConfidenceConsistency(result *VotingResult, votes []*StrategyVote) float64 {
	if len(result.FinalResults) == 0 {
		return 0.0
	}

	var consistencyScores []float64
	for _, res := range result.FinalResults {
		var confidences []float64
		for _, vote := range votes {
			for _, voteRes := range vote.Results {
				if voteRes.Code != nil && res.Code != nil &&
					voteRes.Code.Type == res.Code.Type && voteRes.Code.Code == res.Code.Code {
					confidences = append(confidences, voteRes.Confidence)
				}
			}
		}

		if len(confidences) > 1 {
			mean := vv.calculateMean(confidences)
			variance := vv.calculateVariance(confidences, mean)
			consistency := 1.0 - math.Min(variance, 1.0)
			consistencyScores = append(consistencyScores, consistency)
		}
	}

	if len(consistencyScores) == 0 {
		return 0.0
	}

	return vv.calculateMean(consistencyScores)
}

// calculateResultStability measures stability of results across strategies
func (vv *VotingValidator) calculateResultStability(result *VotingResult, votes []*StrategyVote) float64 {
	if len(votes) < 2 {
		return 1.0
	}

	// Calculate rank stability for each result
	var stabilityScores []float64
	for _, res := range result.FinalResults {
		var ranks []int
		for _, vote := range votes {
			for i, voteRes := range vote.Results {
				if voteRes.Code != nil && res.Code != nil &&
					voteRes.Code.Type == res.Code.Type && voteRes.Code.Code == res.Code.Code {
					ranks = append(ranks, i)
					break
				}
			}
		}

		if len(ranks) > 1 {
			mean := float64(vv.calculateIntMean(ranks))
			variance := vv.calculateIntVariance(ranks, int(mean))
			stability := 1.0 - math.Min(float64(variance)/float64(len(votes[0].Results)), 1.0)
			stabilityScores = append(stabilityScores, stability)
		}
	}

	if len(stabilityScores) == 0 {
		return 0.0
	}

	return vv.calculateMean(stabilityScores)
}

// calculateAnomalyScore calculates anomaly score for voting results
func (vv *VotingValidator) calculateAnomalyScore(result *VotingResult, votes []*StrategyVote) float64 {
	if len(votes) < 3 {
		return 0.0 // Need at least 3 strategies for anomaly detection
	}

	// Calculate confidence anomalies
	var confidenceAnomalies []float64
	for _, vote := range votes {
		allConfidences := []float64{}
		for _, v := range votes {
			allConfidences = append(allConfidences, v.Confidence)
		}
		overallMean := vv.calculateMean(allConfidences)
		overallStdDev := vv.calculateStandardDeviation(allConfidences, overallMean)

		if overallStdDev > 0 {
			zScore := math.Abs(vote.Confidence-overallMean) / overallStdDev
			if zScore > vv.config.AnomalyThreshold {
				confidenceAnomalies = append(confidenceAnomalies, zScore)
			}
		}
	}

	if len(confidenceAnomalies) == 0 {
		return 0.0
	}

	return vv.calculateMean(confidenceAnomalies)
}

// identifyConsistencyIssues identifies specific consistency issues
func (vv *VotingValidator) identifyConsistencyIssues(result *VotingResult, votes []*StrategyVote, consistency *ConsistencyCheckResult) {
	// Check for low cross-strategy agreement
	if consistency.CrossStrategyAgreement < vv.config.MinStrategyAgreement {
		consistency.ConsistencyIssues = append(consistency.ConsistencyIssues, &ConsistencyIssue{
			Type:        "low_agreement",
			Severity:    "error",
			Description: fmt.Sprintf("Low cross-strategy agreement: %.3f", consistency.CrossStrategyAgreement),
			Impact:      1.0 - consistency.CrossStrategyAgreement,
			Strategies:  vv.getStrategyNames(votes),
		})
	}

	// Check for high confidence variance
	if consistency.ConfidenceConsistency < 0.7 {
		consistency.ConsistencyIssues = append(consistency.ConsistencyIssues, &ConsistencyIssue{
			Type:        "confidence_variance",
			Severity:    "warning",
			Description: fmt.Sprintf("High confidence variance: %.3f", consistency.ConfidenceConsistency),
			Impact:      1.0 - consistency.ConfidenceConsistency,
			Strategies:  vv.getStrategyNames(votes),
		})
	}

	// Check for result instability
	if consistency.ResultStability < 0.6 {
		consistency.ConsistencyIssues = append(consistency.ConsistencyIssues, &ConsistencyIssue{
			Type:        "result_instability",
			Severity:    "warning",
			Description: fmt.Sprintf("Result instability detected: %.3f", consistency.ResultStability),
			Impact:      1.0 - consistency.ResultStability,
			Strategies:  vv.getStrategyNames(votes),
		})
	}

	// Check for anomalies
	if consistency.AnomalyScore > vv.config.AnomalyThreshold {
		consistency.ConsistencyIssues = append(consistency.ConsistencyIssues, &ConsistencyIssue{
			Type:        "anomaly_detected",
			Severity:    "error",
			Description: fmt.Sprintf("Anomaly detected with score: %.3f", consistency.AnomalyScore),
			Impact:      consistency.AnomalyScore / vv.config.AnomalyThreshold,
			Strategies:  vv.getStrategyNames(votes),
		})
	}
}

// detectAnomalies performs anomaly detection on voting results
func (vv *VotingValidator) detectAnomalies(result *VotingResult, votes []*StrategyVote, validationResult *VotingValidationResult) {
	if len(votes) < 3 {
		return // Need at least 3 strategies for anomaly detection
	}

	// Detect confidence anomalies
	var confidenceAnomalies []string
	allConfidences := []float64{}
	for _, vote := range votes {
		allConfidences = append(allConfidences, vote.Confidence)
	}

	mean := vv.calculateMean(allConfidences)
	stdDev := vv.calculateStandardDeviation(allConfidences, mean)

	for _, vote := range votes {
		if stdDev > 0 {
			zScore := math.Abs(vote.Confidence-mean) / stdDev
			if zScore > vv.config.AnomalyThreshold {
				confidenceAnomalies = append(confidenceAnomalies, vote.StrategyName)
				validationResult.Warnings = append(validationResult.Warnings, &ValidationWarning{
					Type:    "anomaly",
					Message: fmt.Sprintf("Anomalous confidence for strategy %s: z-score %.2f", vote.StrategyName, zScore),
					Field:   "confidence",
					Value:   fmt.Sprintf("%.3f", vote.Confidence),
					Impact:  "May indicate strategy malfunction or data quality issues",
				})
			}
		}
	}

	// Detect result distribution anomalies
	vv.detectResultDistributionAnomalies(result, votes, validationResult)
}

// detectResultDistributionAnomalies detects anomalies in result distribution
func (vv *VotingValidator) detectResultDistributionAnomalies(result *VotingResult, votes []*StrategyVote, validationResult *VotingValidationResult) {
	// Check for strategies with unusually many or few results
	resultCounts := make(map[string]int)
	for _, vote := range votes {
		resultCounts[vote.StrategyName] = len(vote.Results)
	}

	var counts []int
	for _, count := range resultCounts {
		counts = append(counts, count)
	}

	if len(counts) > 1 {
		mean := float64(vv.calculateIntMean(counts))
		variance := vv.calculateIntVariance(counts, int(mean))
		stdDev := math.Sqrt(float64(variance))

		for strategy, count := range resultCounts {
			if stdDev > 0 {
				zScore := math.Abs(float64(count)-mean) / stdDev
				if zScore > vv.config.AnomalyThreshold {
					validationResult.Warnings = append(validationResult.Warnings, &ValidationWarning{
						Type:    "result_distribution",
						Message: fmt.Sprintf("Anomalous result count for strategy %s: %d results (z-score %.2f)", strategy, count, zScore),
						Field:   "result_count",
						Value:   fmt.Sprintf("%d", count),
						Impact:  "May indicate strategy bias or configuration issues",
					})
				}
			}
		}
	}
}

// performCrossValidation performs cross-validation of voting results
func (vv *VotingValidator) performCrossValidation(result *VotingResult, votes []*StrategyVote, validationResult *VotingValidationResult) {
	if len(votes) < 2 {
		return
	}

	// Perform leave-one-out cross-validation
	var crossValidationScores []float64
	for i := range votes {
		// Create subset without strategy i
		subsetVotes := make([]*StrategyVote, 0, len(votes)-1)
		for j, vote := range votes {
			if i != j {
				subsetVotes = append(subsetVotes, vote)
			}
		}

		// Calculate agreement between subset and full result
		agreement := vv.calculateCrossValidationAgreement(result, subsetVotes)
		crossValidationScores = append(crossValidationScores, agreement)
	}

	if len(crossValidationScores) > 0 {
		meanScore := vv.calculateMean(crossValidationScores)
		if meanScore < vv.config.CrossValidationThreshold {
			validationResult.Issues = append(validationResult.Issues, &VotingValidationIssue{
				Type:           "cross_validation",
				Severity:       "error",
				Message:        fmt.Sprintf("Cross-validation score below threshold: %.3f", meanScore),
				Field:          "cross_validation",
				Value:          fmt.Sprintf("%.3f", meanScore),
				Expected:       fmt.Sprintf(">= %.3f", vv.config.CrossValidationThreshold),
				Recommendation: "Results are not stable across strategy subsets, review voting configuration",
			})
		}
	}
}

// calculateCrossValidationAgreement calculates agreement for cross-validation
func (vv *VotingValidator) calculateCrossValidationAgreement(result *VotingResult, subsetVotes []*StrategyVote) float64 {
	// Create subset result map
	subsetResults := make(map[string]*ClassificationResult)
	for _, vote := range subsetVotes {
		for _, res := range vote.Results {
			if res.Code != nil {
				key := fmt.Sprintf("%s:%s", res.Code.Type, res.Code.Code)
				if existing, exists := subsetResults[key]; !exists || res.Confidence > existing.Confidence {
					subsetResults[key] = res
				}
			}
		}
	}

	// Calculate agreement with full result
	agreement := 0.0
	for _, res := range result.FinalResults {
		if res.Code != nil {
			key := fmt.Sprintf("%s:%s", res.Code.Type, res.Code.Code)
			if _, exists := subsetResults[key]; exists {
				agreement += 1.0
			}
		}
	}

	if len(result.FinalResults) == 0 {
		return 0.0
	}

	return agreement / float64(len(result.FinalResults))
}

// performStatisticalValidation performs statistical validation of results
func (vv *VotingValidator) performStatisticalValidation(result *VotingResult, votes []*StrategyVote, validationResult *VotingValidationResult) {
	if len(votes) < 2 {
		return
	}

	// Perform chi-square test for independence
	chiSquareScore := vv.calculateChiSquareScore(result, votes)
	if chiSquareScore > vv.config.StatisticalSignificance {
		validationResult.Warnings = append(validationResult.Warnings, &ValidationWarning{
			Type:    "statistical",
			Message: fmt.Sprintf("High chi-square score indicates potential strategy dependence: %.3f", chiSquareScore),
			Field:   "statistical_independence",
			Value:   fmt.Sprintf("%.3f", chiSquareScore),
			Impact:  "Strategies may not be independent, consider strategy diversity",
		})
	}
}

// calculateChiSquareScore calculates chi-square score for statistical independence
func (vv *VotingValidator) calculateChiSquareScore(result *VotingResult, votes []*StrategyVote) float64 {
	// Simplified chi-square calculation for strategy independence
	// This is a basic implementation - in practice, you might want more sophisticated statistical tests

	var chiSquare float64
	for _, res := range result.FinalResults {
		if res.Code != nil {
			expected := float64(len(votes)) / float64(len(result.FinalResults))
			observed := 0.0
			for _, vote := range votes {
				for _, voteRes := range vote.Results {
					if voteRes.Code != nil && voteRes.Code.Type == res.Code.Type && voteRes.Code.Code == res.Code.Code {
						observed += 1.0
						break
					}
				}
			}

			if expected > 0 {
				chiSquare += math.Pow(observed-expected, 2) / expected
			}
		}
	}

	return chiSquare
}

// performTemporalValidation performs temporal validation of results
func (vv *VotingValidator) performTemporalValidation(result *VotingResult, votes []*StrategyVote, validationResult *VotingValidationResult) {
	// Check for temporal consistency in voting times
	var voteTimes []time.Time
	for _, vote := range votes {
		voteTimes = append(voteTimes, vote.VoteTime)
	}

	if len(voteTimes) > 1 {
		// Sort times to check for temporal spread
		sort.Slice(voteTimes, func(i, j int) bool {
			return voteTimes[i].Before(voteTimes[j])
		})

		timeSpread := voteTimes[len(voteTimes)-1].Sub(voteTimes[0])
		if timeSpread > vv.config.TemporalWindow {
			validationResult.Warnings = append(validationResult.Warnings, &ValidationWarning{
				Type:    "temporal",
				Message: fmt.Sprintf("Large temporal spread in voting: %v", timeSpread),
				Field:   "vote_times",
				Value:   timeSpread.String(),
				Impact:  "May indicate system performance issues or strategy delays",
			})
		}
	}
}

// calculateOverallValidationScore calculates the overall validation score
func (vv *VotingValidator) calculateOverallValidationScore(validationResult *VotingValidationResult) float64 {
	baseScore := 1.0

	// Deduct points for issues
	for _, issue := range validationResult.Issues {
		switch issue.Severity {
		case "critical":
			baseScore -= 0.3
		case "error":
			baseScore -= 0.2
		case "warning":
			baseScore -= 0.1
		}
	}

	// Deduct points for warnings
	for range validationResult.Warnings {
		baseScore -= 0.05
	}

	// Add quality metrics contribution
	if validationResult.QualityMetrics != nil {
		baseScore = (baseScore + validationResult.QualityMetrics.OverallQuality) / 2.0
	}

	// Add consistency contribution
	if validationResult.ConsistencyChecks != nil {
		consistencyScore := (validationResult.ConsistencyChecks.CrossStrategyAgreement +
			validationResult.ConsistencyChecks.ConfidenceConsistency +
			validationResult.ConsistencyChecks.ResultStability) / 3.0
		baseScore = (baseScore + consistencyScore) / 2.0
	}

	return math.Max(0.0, math.Min(1.0, baseScore))
}

// generateRecommendations generates improvement recommendations
func (vv *VotingValidator) generateRecommendations(validationResult *VotingValidationResult) []string {
	var recommendations []string

	// Add recommendations based on issues
	for _, issue := range validationResult.Issues {
		if issue.Recommendation != "" {
			recommendations = append(recommendations, issue.Recommendation)
		}
	}

	// Add quality-based recommendations
	if validationResult.QualityMetrics != nil {
		if validationResult.QualityMetrics.ResultCompleteness < 0.8 {
			recommendations = append(recommendations, "Increase the number of contributing strategies or improve strategy coverage")
		}
		if validationResult.QualityMetrics.ConfidenceReliability < 0.7 {
			recommendations = append(recommendations, "Review confidence calculation methods and ensure consistent confidence scoring")
		}
		if validationResult.QualityMetrics.StrategyConsistency < 0.6 {
			recommendations = append(recommendations, "Improve strategy consistency by reviewing strategy configuration and data quality")
		}
		if validationResult.QualityMetrics.CodeFormatCompliance < 0.9 {
			recommendations = append(recommendations, "Ensure all classification strategies return properly formatted industry codes")
		}
	}

	// Add consistency-based recommendations
	if validationResult.ConsistencyChecks != nil {
		if validationResult.ConsistencyChecks.CrossStrategyAgreement < 0.6 {
			recommendations = append(recommendations, "Strategies are not agreeing on results, consider adjusting strategy weights or improving individual strategies")
		}
		if validationResult.ConsistencyChecks.ConfidenceConsistency < 0.7 {
			recommendations = append(recommendations, "High confidence variance detected, review confidence calculation methods across strategies")
		}
		if validationResult.ConsistencyChecks.ResultStability < 0.6 {
			recommendations = append(recommendations, "Result instability detected, review strategy ranking and result aggregation methods")
		}
		if validationResult.ConsistencyChecks.AnomalyScore > 2.0 {
			recommendations = append(recommendations, "Anomalies detected in voting results, investigate strategy performance and data quality")
		}
	}

	// Add general recommendations based on validation score
	if validationResult.ValidationScore < 0.7 {
		recommendations = append(recommendations, "Overall validation score is low, review voting configuration and strategy performance")
	}

	// Remove duplicates
	seen := make(map[string]bool)
	var uniqueRecommendations []string
	for _, rec := range recommendations {
		if !seen[rec] {
			seen[rec] = true
			uniqueRecommendations = append(uniqueRecommendations, rec)
		}
	}

	return uniqueRecommendations
}

// Helper methods for statistical calculations
func (vv *VotingValidator) calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func (vv *VotingValidator) calculateVariance(values []float64, mean float64) float64 {
	if len(values) == 0 {
		return 0.0
	}
	sum := 0.0
	for _, v := range values {
		sum += math.Pow(v-mean, 2)
	}
	return sum / float64(len(values))
}

func (vv *VotingValidator) calculateStandardDeviation(values []float64, mean float64) float64 {
	return math.Sqrt(vv.calculateVariance(values, mean))
}

func (vv *VotingValidator) calculateIntMean(values []int) int {
	if len(values) == 0 {
		return 0
	}
	sum := 0
	for _, v := range values {
		sum += v
	}
	return sum / len(values)
}

func (vv *VotingValidator) calculateIntVariance(values []int, mean int) int {
	if len(values) == 0 {
		return 0
	}
	sum := 0
	for _, v := range values {
		sum += (v - mean) * (v - mean)
	}
	return sum / len(values)
}

func (vv *VotingValidator) getStrategyNames(votes []*StrategyVote) []string {
	names := make([]string, 0, len(votes))
	for _, vote := range votes {
		names = append(names, vote.StrategyName)
	}
	return names
}
