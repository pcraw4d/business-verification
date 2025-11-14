package risk

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"
)

// TestConfig contains configuration for integration tests
type TestConfig struct {
	// Test Environment
	Environment   string        `json:"environment"`
	LogLevel      string        `json:"log_level"`
	TestTimeout   time.Duration `json:"test_timeout"`
	ParallelTests int           `json:"parallel_tests"`

	// Database Configuration
	DatabaseURL     string        `json:"database_url"`
	DatabaseTimeout time.Duration `json:"database_timeout"`

	// API Configuration
	APIPort    int           `json:"api_port"`
	APITimeout time.Duration `json:"api_timeout"`

	// Performance Configuration
	MaxConcurrency       int           `json:"max_concurrency"`
	LargeDatasetSize     int           `json:"large_dataset_size"`
	PerformanceThreshold time.Duration `json:"performance_threshold"`

	// File System Configuration
	TempDir   string `json:"temp_dir"`
	BackupDir string `json:"backup_dir"`
	ReportDir string `json:"report_dir"`

	// Test Data Configuration
	TestBusinessID string `json:"test_business_id"`
	TestDataSize   int    `json:"test_data_size"`

	// Feature Flags
	EnableAPITests    bool `json:"enable_api_tests"`
	EnableDBTests     bool `json:"enable_db_tests"`
	EnablePerfTests   bool `json:"enable_perf_tests"`
	EnableBackupTests bool `json:"enable_backup_tests"`
	EnableExportTests bool `json:"enable_export_tests"`
}

// DefaultTestConfig returns the default test configuration
func DefaultTestConfig() *TestConfig {
	return &TestConfig{
		Environment:          "test",
		LogLevel:             "info",
		TestTimeout:          30 * time.Minute,
		ParallelTests:        4,
		DatabaseURL:          "postgres://test:test@localhost:5432/test_db",
		DatabaseTimeout:      10 * time.Second,
		APIPort:              8080,
		APITimeout:           30 * time.Second,
		MaxConcurrency:       10,
		LargeDatasetSize:     1000,
		PerformanceThreshold: 5 * time.Second,
		TempDir:              "/tmp/risk_tests",
		BackupDir:            "/tmp/risk_backups",
		ReportDir:            "/tmp/risk_reports",
		TestBusinessID:       "test-business-123",
		TestDataSize:         100,
		EnableAPITests:       true,
		EnableDBTests:        true,
		EnablePerfTests:      true,
		EnableBackupTests:    true,
		EnableExportTests:    true,
	}
}

// LoadTestConfigFromEnv loads test configuration from environment variables
func LoadTestConfigFromEnv() *TestConfig {
	config := DefaultTestConfig()

	// Load from environment variables
	if env := os.Getenv("TEST_ENVIRONMENT"); env != "" {
		config.Environment = env
	}

	if level := os.Getenv("TEST_LOG_LEVEL"); level != "" {
		config.LogLevel = level
	}

	if timeout := os.Getenv("TEST_TIMEOUT"); timeout != "" {
		if duration, err := time.ParseDuration(timeout); err == nil {
			config.TestTimeout = duration
		}
	}

	if parallel := os.Getenv("TEST_PARALLEL"); parallel != "" {
		if p, err := strconv.Atoi(parallel); err == nil {
			config.ParallelTests = p
		}
	}

	if dbURL := os.Getenv("TEST_DATABASE_URL"); dbURL != "" {
		config.DatabaseURL = dbURL
	}

	if port := os.Getenv("TEST_API_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			config.APIPort = p
		}
	}

	if concurrency := os.Getenv("TEST_MAX_CONCURRENCY"); concurrency != "" {
		if c, err := strconv.Atoi(concurrency); err == nil {
			config.MaxConcurrency = c
		}
	}

	if datasetSize := os.Getenv("TEST_LARGE_DATASET_SIZE"); datasetSize != "" {
		if s, err := strconv.Atoi(datasetSize); err == nil {
			config.LargeDatasetSize = s
		}
	}

	if tempDir := os.Getenv("TEST_TEMP_DIR"); tempDir != "" {
		config.TempDir = tempDir
	}

	if backupDir := os.Getenv("TEST_BACKUP_DIR"); backupDir != "" {
		config.BackupDir = backupDir
	}

	if reportDir := os.Getenv("TEST_REPORT_DIR"); reportDir != "" {
		config.ReportDir = reportDir
	}

	if businessID := os.Getenv("TEST_BUSINESS_ID"); businessID != "" {
		config.TestBusinessID = businessID
	}

	if dataSize := os.Getenv("TEST_DATA_SIZE"); dataSize != "" {
		if s, err := strconv.Atoi(dataSize); err == nil {
			config.TestDataSize = s
		}
	}

	// Feature flags
	config.EnableAPITests = getBoolEnv("TEST_ENABLE_API", true)
	config.EnableDBTests = getBoolEnv("TEST_ENABLE_DB", true)
	config.EnablePerfTests = getBoolEnv("TEST_ENABLE_PERF", true)
	config.EnableBackupTests = getBoolEnv("TEST_ENABLE_BACKUP", true)
	config.EnableExportTests = getBoolEnv("TEST_ENABLE_EXPORT", true)

	return config
}

// getBoolEnv gets a boolean value from environment variable
func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// Validate validates the test configuration
func (tc *TestConfig) Validate() error {
	if tc.Environment == "" {
		return fmt.Errorf("environment is required")
	}

	if tc.TestTimeout <= 0 {
		return fmt.Errorf("test timeout must be positive")
	}

	if tc.ParallelTests <= 0 {
		return fmt.Errorf("parallel tests must be positive")
	}

	if tc.MaxConcurrency <= 0 {
		return fmt.Errorf("max concurrency must be positive")
	}

	if tc.LargeDatasetSize <= 0 {
		return fmt.Errorf("large dataset size must be positive")
	}

	if tc.TestDataSize <= 0 {
		return fmt.Errorf("test data size must be positive")
	}

	return nil
}

// GetLogger creates a logger based on the configuration
func (tc *TestConfig) GetLogger() *zap.Logger {
	var config zap.Config

	switch tc.LogLevel {
	case "debug":
		config = zap.NewDevelopmentConfig()
	case "info":
		config = zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		config = zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		config = zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		config = zap.NewProductionConfig()
	}

	logger, _ := config.Build()
	return logger
}

// TestDataGenerator generates test data for integration tests
type TestDataGenerator struct {
	config *TestConfig
	logger *zap.Logger
}

// NewTestDataGenerator creates a new test data generator
func NewTestDataGenerator(config *TestConfig) *TestDataGenerator {
	return &TestDataGenerator{
		config: config,
		logger: config.GetLogger(),
	}
}

// GenerateRiskAssessment generates a test risk assessment
func (tdg *TestDataGenerator) GenerateRiskAssessment(businessID string) *RiskAssessment {
	if businessID == "" {
		businessID = tdg.config.TestBusinessID
	}

	return &RiskAssessment{
		ID:           fmt.Sprintf("test-assessment-%d", time.Now().UnixNano()),
		BusinessID:   businessID,
		BusinessName: fmt.Sprintf("Test Business %s", businessID),
		OverallScore: 75.0 + float64(time.Now().UnixNano()%25),
		OverallLevel: RiskLevelHigh,
		AssessedAt:   time.Now(),
		ValidUntil:   time.Now().Add(24 * time.Hour),
		AlertLevel:   RiskLevelMedium,
		CategoryScores: map[RiskCategory]RiskScore{
			RiskCategoryFinancial: {
				FactorID:     "financial",
				FactorName:   "Financial",
				Category:     RiskCategoryFinancial,
				Score:        80.0,
				Level:        RiskLevelHigh,
				Confidence:   0.9,
				Explanation:  "Financial risk assessment",
				CalculatedAt: time.Now(),
			},
			RiskCategoryOperational: {
				FactorID:     "operational",
				FactorName:   "Operational",
				Category:     RiskCategoryOperational,
				Score:        70.0,
				Level:        RiskLevelMedium,
				Confidence:   0.8,
				Explanation:  "Operational risk assessment",
				CalculatedAt: time.Now(),
			},
		},
		FactorScores: []RiskScore{
			{
				FactorID:     "financial-1",
				FactorName:   "Financial Stability",
				Category:     RiskCategoryFinancial,
				Score:        80.0,
				Level:        RiskLevelHigh,
				Confidence:   0.9,
				Explanation:  "Strong financial position",
				CalculatedAt: time.Now(),
			},
		},
		Recommendations: []RiskRecommendation{
			{
				ID:          "rec-1",
				RiskFactor:  "financial",
				Priority:    RiskLevelHigh,
				Title:       "Improve Financial Monitoring",
				Description: "Implement regular financial monitoring",
				Action:      "Set up monthly reviews and implement alerts",
				Impact:      "Medium",
				Timeline:    "Within 30 days",
				CreatedAt:   time.Now(),
			},
		},
		Alerts: []RiskAlert{
			{
				ID:             "alert-1",
				BusinessID:     businessID,
				RiskFactor:     "financial-1",
				Level:          RiskLevelHigh,
				Message:        "High financial risk detected",
				Score:          80.0,
				Threshold:      75.0,
				TriggeredAt:    time.Now(),
				Acknowledged:   false,
				AcknowledgedAt: nil,
			},
		},
	}
}

// GenerateMultipleRiskAssessments generates multiple test risk assessments
func (tdg *TestDataGenerator) GenerateMultipleRiskAssessments(count int, businessID string) []*RiskAssessment {
	assessments := make([]*RiskAssessment, count)

	for i := 0; i < count; i++ {
		assessments[i] = tdg.GenerateRiskAssessment(businessID)
		// Add small delay to ensure unique timestamps
		time.Sleep(1 * time.Millisecond)
	}

	return assessments
}

// GenerateExportRequest generates a test export request
func (tdg *TestDataGenerator) GenerateExportRequest(businessID string) *ExportRequest {
	if businessID == "" {
		businessID = tdg.config.TestBusinessID
	}

	return &ExportRequest{
		BusinessID: businessID,
		ExportType: ExportTypeAssessments,
		Format:     ExportFormatJSON,
		Metadata:   map[string]interface{}{"test": "integration"},
	}
}

// GenerateBackupRequest generates a test backup request
func (tdg *TestDataGenerator) GenerateBackupRequest(businessID string) *BackupRequest {
	if businessID == "" {
		businessID = tdg.config.TestBusinessID
	}

	return &BackupRequest{
		BusinessID:  businessID,
		BackupType:  BackupTypeBusiness,
		IncludeData: []BackupDataType{BackupDataTypeAssessments},
		Metadata:    map[string]interface{}{"test": "integration"},
	}
}

// CleanupTestData cleans up test data
func (tdg *TestDataGenerator) CleanupTestData() error {
	tdg.logger.Info("Cleaning up test data")

	// Clean up temporary directories
	dirs := []string{
		tdg.config.TempDir,
		tdg.config.BackupDir,
		tdg.config.ReportDir,
	}

	for _, dir := range dirs {
		if err := os.RemoveAll(dir); err != nil {
			tdg.logger.Warn("Failed to clean up directory", zap.String("dir", dir), zap.Error(err))
		}
	}

	tdg.logger.Info("Test data cleanup completed")
	return nil
}
