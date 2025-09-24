package risk

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// UserAcceptanceTesting provides comprehensive user acceptance testing for the KYB platform
type UserAcceptanceTesting struct {
	logger            *zap.Logger
	config            *UATConfig
	testCases         map[string]*UATTestCase
	results           *UATResults
	userSimulator     *UserSimulator
	feedbackCollector *FeedbackCollector
	reportGenerator   *UATReportGenerator
}

// UATConfig contains configuration for user acceptance testing
type UATConfig struct {
	TestEnvironment      string            `json:"test_environment"`
	TestTimeout          time.Duration     `json:"test_timeout"`
	ReportOutputPath     string            `json:"report_output_path"`
	LogLevel             string            `json:"log_level"`
	EnableUserSimulation bool              `json:"enable_user_simulation"`
	UserCount            int               `json:"user_count"`
	TestDuration         time.Duration     `json:"test_duration"`
	FeedbackCollection   bool              `json:"feedback_collection"`
	EnvironmentVariables map[string]string `json:"environment_variables"`
	DatabaseConfig       *DatabaseConfig   `json:"database_config"`
	APIConfig            *APIConfig        `json:"api_config"`
	ResourceLimits       *ResourceLimits   `json:"resource_limits"`
}

// UATTestCase represents a user acceptance test case
type UATTestCase struct {
	ID                 string                      `json:"id"`
	Name               string                      `json:"name"`
	Description        string                      `json:"description"`
	Category           string                      `json:"category"`
	Priority           string                      `json:"priority"`
	UserStory          string                      `json:"user_story"`
	AcceptanceCriteria []string                    `json:"acceptance_criteria"`
	Function           func(*UATContext) UATResult `json:"-"`
	SetupFunction      func(*UATContext) error     `json:"-"`
	CleanupFunction    func(*UATContext) error     `json:"-"`
	Parameters         map[string]interface{}      `json:"parameters"`
	ExpectedOutcome    *ExpectedOutcome            `json:"expected_outcome"`
	Tags               []string                    `json:"tags"`
}

// UATContext provides context for user acceptance test execution
type UATContext struct {
	ID          string                 `json:"id"`
	TestCaseID  string                 `json:"test_case_id"`
	UserID      string                 `json:"user_id"`
	UserRole    string                 `json:"user_role"`
	Parameters  map[string]interface{} `json:"parameters"`
	StartTime   time.Time              `json:"start_time"`
	EndTime     time.Time              `json:"end_time"`
	Duration    time.Duration          `json:"duration"`
	Context     context.Context        `json:"-"`
	Logger      *zap.Logger            `json:"-"`
	UserSession *UserSession           `json:"user_session"`
	TestData    map[string]interface{} `json:"test_data"`
}

// UATResult contains the result of a user acceptance test execution
type UATResult struct {
	TestCaseID         string                 `json:"test_case_id"`
	UserID             string                 `json:"user_id"`
	UserRole           string                 `json:"user_role"`
	StartTime          time.Time              `json:"start_time"`
	EndTime            time.Time              `json:"end_time"`
	Duration           time.Duration          `json:"duration"`
	Success            bool                   `json:"success"`
	ErrorMessage       string                 `json:"error_message,omitempty"`
	ExpectedOutcome    *ExpectedOutcome       `json:"expected_outcome"`
	ActualOutcome      *ActualOutcome         `json:"actual_outcome"`
	UserSatisfaction   *UserSatisfaction      `json:"user_satisfaction"`
	UsabilityMetrics   *UsabilityMetrics      `json:"usability_metrics"`
	PerformanceMetrics *PerformanceMetrics    `json:"performance_metrics"`
	Feedback           *UserFeedback          `json:"feedback"`
	Recommendations    []string               `json:"recommendations"`
	CustomMetrics      map[string]interface{} `json:"custom_metrics"`
}

// ExpectedOutcome defines the expected outcome of a UAT test case
type ExpectedOutcome struct {
	FunctionalityWorks    bool          `json:"functionality_works"`
	UserCanComplete       bool          `json:"user_can_complete"`
	PerformanceAcceptable bool          `json:"performance_acceptable"`
	NoErrors              bool          `json:"no_errors"`
	DataIntegrity         bool          `json:"data_integrity"`
	UserSatisfaction      float64       `json:"user_satisfaction"` // 0-10 scale
	CompletionTime        time.Duration `json:"completion_time"`
	ErrorRate             float64       `json:"error_rate"`
	SuccessRate           float64       `json:"success_rate"`
}

// ActualOutcome defines the actual outcome observed during UAT
type ActualOutcome struct {
	FunctionalityWorks    bool          `json:"functionality_works"`
	UserCanComplete       bool          `json:"user_can_complete"`
	PerformanceAcceptable bool          `json:"performance_acceptable"`
	NoErrors              bool          `json:"no_errors"`
	DataIntegrity         bool          `json:"data_integrity"`
	UserSatisfaction      float64       `json:"user_satisfaction"`
	CompletionTime        time.Duration `json:"completion_time"`
	ErrorRate             float64       `json:"error_rate"`
	SuccessRate           float64       `json:"success_rate"`
	IssuesEncountered     []string      `json:"issues_encountered"`
	WorkaroundsUsed       []string      `json:"workarounds_used"`
}

// UserSatisfaction contains user satisfaction metrics
type UserSatisfaction struct {
	OverallRating          float64  `json:"overall_rating"`  // 0-10 scale
	EaseOfUse              float64  `json:"ease_of_use"`     // 0-10 scale
	Functionality          float64  `json:"functionality"`   // 0-10 scale
	Performance            float64  `json:"performance"`     // 0-10 scale
	Reliability            float64  `json:"reliability"`     // 0-10 scale
	UserExperience         float64  `json:"user_experience"` // 0-10 scale
	WouldRecommend         bool     `json:"would_recommend"`
	Comments               string   `json:"comments"`
	ImprovementSuggestions []string `json:"improvement_suggestions"`
}

// UsabilityMetrics contains usability metrics
type UsabilityMetrics struct {
	TaskCompletionRate float64       `json:"task_completion_rate"`
	ErrorRate          float64       `json:"error_rate"`
	TimeToComplete     time.Duration `json:"time_to_complete"`
	TimeToFirstAction  time.Duration `json:"time_to_first_action"`
	ClickCount         int           `json:"click_count"`
	NavigationDepth    int           `json:"navigation_depth"`
	HelpRequests       int           `json:"help_requests"`
	ConfusionPoints    []string      `json:"confusion_points"`
	EfficiencyScore    float64       `json:"efficiency_score"`
	EffectivenessScore float64       `json:"effectiveness_score"`
	SatisfactionScore  float64       `json:"satisfaction_score"`
}

// UserFeedback contains user feedback data
type UserFeedback struct {
	OverallExperience  string   `json:"overall_experience"`
	LikedFeatures      []string `json:"liked_features"`
	DislikedFeatures   []string `json:"disliked_features"`
	MissingFeatures    []string `json:"missing_features"`
	BugReports         []string `json:"bug_reports"`
	ImprovementIdeas   []string `json:"improvement_ideas"`
	ComparisonToOther  string   `json:"comparison_to_other"`
	WillingnessToPay   string   `json:"willingness_to_pay"`
	AdditionalComments string   `json:"additional_comments"`
}

// UATResults contains the results of user acceptance testing
type UATResults struct {
	SessionID        string                 `json:"session_id"`
	StartTime        time.Time              `json:"start_time"`
	EndTime          time.Time              `json:"end_time"`
	TotalDuration    time.Duration          `json:"total_duration"`
	Environment      string                 `json:"environment"`
	TotalTestCases   int                    `json:"total_test_cases"`
	PassedTestCases  int                    `json:"passed_test_cases"`
	FailedTestCases  int                    `json:"failed_test_cases"`
	SkippedTestCases int                    `json:"skipped_test_cases"`
	PassRate         float64                `json:"pass_rate"`
	TestCaseResults  map[string][]UATResult `json:"test_case_results"`
	Summary          *UATSummary            `json:"summary"`
	Recommendations  []string               `json:"recommendations"`
	Issues           []UATIssue             `json:"issues"`
}

// UATSummary contains a summary of UAT results
type UATSummary struct {
	OverallPassRate         float64                       `json:"overall_pass_rate"`
	OverallUserSatisfaction float64                       `json:"overall_user_satisfaction"`
	OverallUsabilityScore   float64                       `json:"overall_usability_score"`
	AverageCompletionTime   time.Duration                 `json:"average_completion_time"`
	AverageErrorRate        float64                       `json:"average_error_rate"`
	RecommendationRate      float64                       `json:"recommendation_rate"`
	CategoryMetrics         map[string]CategoryUATMetrics `json:"category_metrics"`
	UserPersonaResults      map[string]UserPersonaResults `json:"user_persona_results"`
	FeatureAnalysis         *FeatureAnalysis              `json:"feature_analysis"`
	CompetitiveAnalysis     *CompetitiveAnalysis          `json:"competitive_analysis"`
}

// CategoryUATMetrics contains UAT metrics for a specific category
type CategoryUATMetrics struct {
	CategoryName          string        `json:"category_name"`
	TestCaseCount         int           `json:"test_case_count"`
	PassRate              float64       `json:"pass_rate"`
	UserSatisfaction      float64       `json:"user_satisfaction"`
	UsabilityScore        float64       `json:"usability_score"`
	AverageCompletionTime time.Duration `json:"average_completion_time"`
	ErrorRate             float64       `json:"error_rate"`
	RecommendationRate    float64       `json:"recommendation_rate"`
}

// UserPersonaResults contains results for specific user personas
type UserPersonaResults struct {
	PersonaName             string        `json:"persona_name"`
	UserCount               int           `json:"user_count"`
	PassRate                float64       `json:"pass_rate"`
	UserSatisfaction        float64       `json:"user_satisfaction"`
	UsabilityScore          float64       `json:"usability_score"`
	AverageCompletionTime   time.Duration `json:"average_completion_time"`
	ErrorRate               float64       `json:"error_rate"`
	RecommendationRate      float64       `json:"recommendation_rate"`
	CommonIssues            []string      `json:"common_issues"`
	PersonaSpecificFeedback []string      `json:"persona_specific_feedback"`
}

// FeatureAnalysis contains analysis of feature performance
type FeatureAnalysis struct {
	MostLikedFeatures    []string                     `json:"most_liked_features"`
	MostDislikedFeatures []string                     `json:"most_disliked_features"`
	MissingFeatures      []string                     `json:"missing_features"`
	FeatureUsageStats    map[string]FeatureUsageStats `json:"feature_usage_stats"`
	FeatureSatisfaction  map[string]float64           `json:"feature_satisfaction"`
}

// FeatureUsageStats contains usage statistics for features
type FeatureUsageStats struct {
	UsageCount       int           `json:"usage_count"`
	UsageRate        float64       `json:"usage_rate"`
	CompletionRate   float64       `json:"completion_rate"`
	ErrorRate        float64       `json:"error_rate"`
	AverageTime      time.Duration `json:"average_time"`
	UserSatisfaction float64       `json:"user_satisfaction"`
}

// CompetitiveAnalysis contains competitive analysis results
type CompetitiveAnalysis struct {
	ComparedToCompetitors string   `json:"compared_to_competitors"`
	Advantages            []string `json:"advantages"`
	Disadvantages         []string `json:"disadvantages"`
	UniqueFeatures        []string `json:"unique_features"`
	MarketPosition        string   `json:"market_position"`
	PricingFeedback       string   `json:"pricing_feedback"`
}

// UATIssue represents an issue found during user acceptance testing
type UATIssue struct {
	ID                string    `json:"id"`
	TestCaseID        string    `json:"test_case_id"`
	Severity          string    `json:"severity"`
	Category          string    `json:"category"`
	Title             string    `json:"title"`
	Description       string    `json:"description"`
	UserImpact        string    `json:"user_impact"`
	BusinessImpact    string    `json:"business_impact"`
	ReproductionSteps []string  `json:"reproduction_steps"`
	ExpectedBehavior  string    `json:"expected_behavior"`
	ActualBehavior    string    `json:"actual_behavior"`
	Workaround        string    `json:"workaround"`
	Recommendation    string    `json:"recommendation"`
	DetectedAt        time.Time `json:"detected_at"`
	ReportedBy        string    `json:"reported_by"`
	Tags              []string  `json:"tags"`
}

// UserSession represents a user session during UAT
type UserSession struct {
	SessionID    string        `json:"session_id"`
	UserID       string        `json:"user_id"`
	UserRole     string        `json:"user_role"`
	StartTime    time.Time     `json:"start_time"`
	EndTime      time.Time     `json:"end_time"`
	Duration     time.Duration `json:"duration"`
	Actions      []UserAction  `json:"actions"`
	Pages        []string      `json:"pages"`
	Features     []string      `json:"features"`
	Errors       []string      `json:"errors"`
	HelpRequests []string      `json:"help_requests"`
	Feedback     *UserFeedback `json:"feedback"`
}

// UserAction represents a user action during UAT
type UserAction struct {
	ActionType string        `json:"action_type"`
	Timestamp  time.Time     `json:"timestamp"`
	Page       string        `json:"page"`
	Element    string        `json:"element"`
	Value      string        `json:"value"`
	Duration   time.Duration `json:"duration"`
	Success    bool          `json:"success"`
	Error      string        `json:"error,omitempty"`
}

// UserSimulator provides user simulation capabilities
type UserSimulator struct {
	logger    *zap.Logger
	config    *UATConfig
	enabled   bool
	userCount int
}

// FeedbackCollector provides feedback collection capabilities
type FeedbackCollector struct {
	logger   *zap.Logger
	config   *UATConfig
	enabled  bool
	feedback []UserFeedback
}

// UATReportGenerator generates UAT reports
type UATReportGenerator struct {
	logger *zap.Logger
	config *UATConfig
}

// NewUserAcceptanceTesting creates a new user acceptance testing instance
func NewUserAcceptanceTesting(config *UATConfig) *UserAcceptanceTesting {
	logger := zap.NewNop()
	if config.LogLevel == "debug" {
		logger, _ = zap.NewDevelopment()
	} else {
		logger, _ = zap.NewProduction()
	}

	return &UserAcceptanceTesting{
		logger:            logger,
		config:            config,
		testCases:         make(map[string]*UATTestCase),
		results:           &UATResults{},
		userSimulator:     NewUserSimulator(logger, config),
		feedbackCollector: NewFeedbackCollector(logger, config),
		reportGenerator:   NewUATReportGenerator(logger, config),
	}
}

// NewUserSimulator creates a new user simulator
func NewUserSimulator(logger *zap.Logger, config *UATConfig) *UserSimulator {
	return &UserSimulator{
		logger:    logger,
		config:    config,
		enabled:   config.EnableUserSimulation,
		userCount: config.UserCount,
	}
}

// NewFeedbackCollector creates a new feedback collector
func NewFeedbackCollector(logger *zap.Logger, config *UATConfig) *FeedbackCollector {
	return &FeedbackCollector{
		logger:   logger,
		config:   config,
		enabled:  config.FeedbackCollection,
		feedback: make([]UserFeedback, 0),
	}
}

// NewUATReportGenerator creates a new UAT report generator
func NewUATReportGenerator(logger *zap.Logger, config *UATConfig) *UATReportGenerator {
	return &UATReportGenerator{
		logger: logger,
		config: config,
	}
}

// AddTestCase adds a UAT test case to the testing suite
func (uat *UserAcceptanceTesting) AddTestCase(testCase *UATTestCase) {
	uat.testCases[testCase.ID] = testCase
	uat.logger.Info("Added UAT test case", zap.String("id", testCase.ID), zap.String("name", testCase.Name))
}

// RunTestCase runs a specific UAT test case
func (uat *UserAcceptanceTesting) RunTestCase(ctx context.Context, testCaseID string, userID string, userRole string) (*UATResult, error) {
	testCase, exists := uat.testCases[testCaseID]
	if !exists {
		return nil, fmt.Errorf("UAT test case with ID %s not found", testCaseID)
	}

	uat.logger.Info("Running UAT test case", zap.String("id", testCaseID), zap.String("name", testCase.Name), zap.String("user_id", userID))

	// Create UAT context
	uatCtx := &UATContext{
		ID:         testCaseID,
		TestCaseID: testCaseID,
		UserID:     userID,
		UserRole:   userRole,
		Parameters: testCase.Parameters,
		Context:    ctx,
		Logger:     uat.logger,
		StartTime:  time.Now(),
		UserSession: &UserSession{
			SessionID:    fmt.Sprintf("session_%s_%s", userID, testCaseID),
			UserID:       userID,
			UserRole:     userRole,
			StartTime:    time.Now(),
			Actions:      make([]UserAction, 0),
			Pages:        make([]string, 0),
			Features:     make([]string, 0),
			Errors:       make([]string, 0),
			HelpRequests: make([]string, 0),
		},
		TestData: make(map[string]interface{}),
	}

	// Run setup if provided
	if testCase.SetupFunction != nil {
		if err := testCase.SetupFunction(uatCtx); err != nil {
			return nil, fmt.Errorf("test case setup failed: %w", err)
		}
	}

	// Execute test case
	result := testCase.Function(uatCtx)
	result.TestCaseID = testCaseID
	result.UserID = userID
	result.UserRole = userRole
	result.StartTime = uatCtx.StartTime
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	// Update context
	uatCtx.EndTime = result.EndTime
	uatCtx.Duration = result.Duration
	uatCtx.UserSession.EndTime = result.EndTime
	uatCtx.UserSession.Duration = result.Duration

	// Run cleanup if provided
	if testCase.CleanupFunction != nil {
		if err := testCase.CleanupFunction(uatCtx); err != nil {
			uat.logger.Warn("Test case cleanup failed", zap.String("test_case_id", testCaseID), zap.Error(err))
		}
	}

	uat.logger.Info("UAT test case completed",
		zap.String("id", testCaseID),
		zap.String("user_id", userID),
		zap.Bool("success", result.Success),
		zap.Duration("duration", result.Duration))

	return &result, nil
}

// RunUATSuite runs all UAT test cases in the suite
func (uat *UserAcceptanceTesting) RunUATSuite(ctx context.Context) (*UATResults, error) {
	uat.logger.Info("Starting UAT suite execution")

	// Initialize results
	uat.results = &UATResults{
		SessionID:       fmt.Sprintf("uat_session_%d", time.Now().Unix()),
		StartTime:       time.Now(),
		Environment:     uat.config.TestEnvironment,
		TestCaseResults: make(map[string][]UATResult),
		Summary:         &UATSummary{},
		Recommendations: make([]string, 0),
		Issues:          make([]UATIssue, 0),
	}

	uat.results.TotalTestCases = len(uat.testCases)

	// Run each test case with multiple users
	for testCaseID, testCase := range uat.testCases {
		select {
		case <-ctx.Done():
			uat.results.EndTime = time.Now()
			uat.results.TotalDuration = uat.results.EndTime.Sub(uat.results.StartTime)
			return uat.results, ctx.Err()
		default:
		}

		uat.logger.Info("Running UAT test case", zap.String("id", testCaseID), zap.String("name", testCase.Name))

		// Run test case with multiple users
		testCaseResults := make([]UATResult, 0)
		userCount := uat.config.UserCount
		if userCount == 0 {
			userCount = 3 // Default to 3 users per test case
		}

		for i := 0; i < userCount; i++ {
			userID := fmt.Sprintf("user_%d", i+1)
			userRole := "standard_user" // Default role, could be parameterized

			result, err := uat.RunTestCase(ctx, testCaseID, userID, userRole)
			if err != nil {
				uat.logger.Error("UAT test case execution failed", zap.String("test_case_id", testCaseID), zap.Error(err))
				uat.results.FailedTestCases++
				continue
			}

			testCaseResults = append(testCaseResults, *result)

			// Check if test case passed
			if uat.isTestCasePassed(testCase, result) {
				uat.results.PassedTestCases++
			} else {
				uat.results.FailedTestCases++
			}
		}

		uat.results.TestCaseResults[testCaseID] = testCaseResults
	}

	// Calculate pass rate
	if uat.results.TotalTestCases > 0 {
		uat.results.PassRate = float64(uat.results.PassedTestCases) / float64(uat.results.TotalTestCases) * 100
	}

	uat.results.EndTime = time.Now()
	uat.results.TotalDuration = uat.results.EndTime.Sub(uat.results.StartTime)

	// Generate summary
	uat.generateSummary()

	// Generate recommendations
	uat.generateRecommendations()

	// Generate reports
	if err := uat.generateReports(); err != nil {
		uat.logger.Error("Failed to generate reports", zap.Error(err))
	}

	uat.logger.Info("UAT suite execution completed",
		zap.Int("total_test_cases", uat.results.TotalTestCases),
		zap.Int("passed_test_cases", uat.results.PassedTestCases),
		zap.Int("failed_test_cases", uat.results.FailedTestCases),
		zap.Float64("pass_rate", uat.results.PassRate))

	return uat.results, nil
}

// isTestCasePassed checks if a UAT test case passed based on expected outcome
func (uat *UserAcceptanceTesting) isTestCasePassed(testCase *UATTestCase, result *UATResult) bool {
	if !result.Success {
		return false
	}

	if testCase.ExpectedOutcome == nil {
		return true
	}

	expected := testCase.ExpectedOutcome
	actual := result.ActualOutcome

	// Check functionality
	if expected.FunctionalityWorks && !actual.FunctionalityWorks {
		return false
	}

	// Check user completion
	if expected.UserCanComplete && !actual.UserCanComplete {
		return false
	}

	// Check performance
	if expected.PerformanceAcceptable && !actual.PerformanceAcceptable {
		return false
	}

	// Check errors
	if expected.NoErrors && !actual.NoErrors {
		return false
	}

	// Check data integrity
	if expected.DataIntegrity && !actual.DataIntegrity {
		return false
	}

	// Check user satisfaction (with tolerance)
	if expected.UserSatisfaction > 0 && actual.UserSatisfaction < expected.UserSatisfaction-1.0 {
		return false
	}

	// Check completion time (with tolerance)
	if expected.CompletionTime > 0 && actual.CompletionTime > expected.CompletionTime*1.5 {
		return false
	}

	// Check error rate
	if expected.ErrorRate > 0 && actual.ErrorRate > expected.ErrorRate*1.5 {
		return false
	}

	// Check success rate
	if expected.SuccessRate > 0 && actual.SuccessRate < expected.SuccessRate-0.1 {
		return false
	}

	return true
}

// generateSummary generates a summary of UAT results
func (uat *UserAcceptanceTesting) generateSummary() {
	summary := &UATSummary{
		CategoryMetrics:     make(map[string]CategoryUATMetrics),
		UserPersonaResults:  make(map[string]UserPersonaResults),
		FeatureAnalysis:     &FeatureAnalysis{},
		CompetitiveAnalysis: &CompetitiveAnalysis{},
	}

	// Aggregate metrics across all test cases
	totalUserSatisfaction := 0.0
	totalUsabilityScore := 0.0
	totalCompletionTime := time.Duration(0)
	totalErrorRate := 0.0
	totalRecommendationRate := 0.0
	userCount := 0

	categoryMetrics := make(map[string]*CategoryUATMetrics)

	for testCaseID, results := range uat.results.TestCaseResults {
		testCase := uat.testCases[testCaseID]
		category := testCase.Category

		// Initialize category metrics if not exists
		if categoryMetrics[category] == nil {
			categoryMetrics[category] = &CategoryUATMetrics{
				CategoryName: category,
			}
		}

		// Aggregate results for this test case
		for _, result := range results {
			catMetrics := categoryMetrics[category]
			catMetrics.TestCaseCount++

			if result.Success {
				catMetrics.PassRate++
			}

			if result.UserSatisfaction != nil {
				totalUserSatisfaction += result.UserSatisfaction.OverallRating
				catMetrics.UserSatisfaction += result.UserSatisfaction.OverallRating
				userCount++

				if result.UserSatisfaction.WouldRecommend {
					totalRecommendationRate++
					catMetrics.RecommendationRate++
				}
			}

			if result.UsabilityMetrics != nil {
				totalUsabilityScore += result.UsabilityMetrics.SatisfactionScore
				catMetrics.UsabilityScore += result.UsabilityMetrics.SatisfactionScore
				totalCompletionTime += result.UsabilityMetrics.TimeToComplete
				catMetrics.AverageCompletionTime += result.UsabilityMetrics.TimeToComplete
				totalErrorRate += result.UsabilityMetrics.ErrorRate
				catMetrics.ErrorRate += result.UsabilityMetrics.ErrorRate
			}
		}
	}

	// Calculate averages
	summary.OverallPassRate = uat.results.PassRate
	if userCount > 0 {
		summary.OverallUserSatisfaction = totalUserSatisfaction / float64(userCount)
		summary.OverallUsabilityScore = totalUsabilityScore / float64(userCount)
		summary.AverageCompletionTime = totalCompletionTime / time.Duration(userCount)
		summary.AverageErrorRate = totalErrorRate / float64(userCount)
		summary.RecommendationRate = totalRecommendationRate / float64(userCount) * 100
	}

	// Calculate category averages
	for category, catMetrics := range categoryMetrics {
		if catMetrics.TestCaseCount > 0 {
			catMetrics.PassRate = catMetrics.PassRate / float64(catMetrics.TestCaseCount) * 100
			catMetrics.UserSatisfaction = catMetrics.UserSatisfaction / float64(catMetrics.TestCaseCount)
			catMetrics.UsabilityScore = catMetrics.UsabilityScore / float64(catMetrics.TestCaseCount)
			catMetrics.AverageCompletionTime = catMetrics.AverageCompletionTime / time.Duration(catMetrics.TestCaseCount)
			catMetrics.ErrorRate = catMetrics.ErrorRate / float64(catMetrics.TestCaseCount)
			catMetrics.RecommendationRate = catMetrics.RecommendationRate / float64(catMetrics.TestCaseCount) * 100
		}
		summary.CategoryMetrics[category] = *catMetrics
	}

	uat.results.Summary = summary
}

// generateRecommendations generates recommendations based on UAT results
func (uat *UserAcceptanceTesting) generateRecommendations() {
	recommendations := make([]string, 0)

	// Low pass rate recommendation
	if uat.results.PassRate < 80 {
		recommendations = append(recommendations, "Low pass rate detected. Review failed test cases and improve user experience.")
	}

	// Low user satisfaction recommendation
	if uat.results.Summary.OverallUserSatisfaction < 7.0 {
		recommendations = append(recommendations, "Low user satisfaction detected. Improve user interface and user experience design.")
	}

	// High error rate recommendation
	if uat.results.Summary.AverageErrorRate > 0.1 {
		recommendations = append(recommendations, "High error rate detected. Improve error handling and user guidance.")
	}

	// Long completion time recommendation
	if uat.results.Summary.AverageCompletionTime > 10*time.Minute {
		recommendations = append(recommendations, "Long completion times detected. Optimize user workflows and reduce complexity.")
	}

	// Low recommendation rate
	if uat.results.Summary.RecommendationRate < 70 {
		recommendations = append(recommendations, "Low recommendation rate detected. Address user pain points and improve overall experience.")
	}

	uat.results.Recommendations = recommendations
}

// generateReports generates UAT reports
func (uat *UserAcceptanceTesting) generateReports() error {
	uat.logger.Info("Generating UAT reports")
	return uat.reportGenerator.GenerateReports(uat.results)
}

// GetResults returns the UAT results
func (uat *UserAcceptanceTesting) GetResults() *UATResults {
	return uat.results
}

// GetTestCases returns all UAT test cases
func (uat *UserAcceptanceTesting) GetTestCases() map[string]*UATTestCase {
	return uat.testCases
}
