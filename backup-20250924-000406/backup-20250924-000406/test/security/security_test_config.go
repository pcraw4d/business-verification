package security

import (
	"fmt"
	"time"
)

// SecurityTestConfig provides configuration for security testing
type SecurityTestConfig struct {
	// Test execution settings
	Timeout        time.Duration `json:"timeout"`
	RetryAttempts  int           `json:"retry_attempts"`
	RetryDelay     time.Duration `json:"retry_delay"`
	ParallelTests  bool          `json:"parallel_tests"`
	MaxConcurrency int           `json:"max_concurrency"`

	// Authentication settings
	JWTSecret         string        `json:"jwt_secret"`
	APIKeySecret      string        `json:"api_key_secret"`
	TokenExpiration   time.Duration `json:"token_expiration"`
	RefreshExpiration time.Duration `json:"refresh_expiration"`

	// Rate limiting settings
	RateLimitEnabled      bool `json:"rate_limit_enabled"`
	RequestsPerMinute     int  `json:"requests_per_minute"`
	BurstSize             int  `json:"burst_size"`
	RateLimitTestRequests int  `json:"rate_limit_test_requests"`

	// Security headers settings
	RequiredHeaders        []string `json:"required_headers"`
	OptionalHeaders        []string `json:"optional_headers"`
	HeaderValidationStrict bool     `json:"header_validation_strict"`

	// Input validation settings
	SQLInjectionTests     []string `json:"sql_injection_tests"`
	XSSAttackTests        []string `json:"xss_attack_tests"`
	PathTraversalTests    []string `json:"path_traversal_tests"`
	CommandInjectionTests []string `json:"command_injection_tests"`

	// Audit logging settings
	AuditLogEnabled   bool          `json:"audit_log_enabled"`
	AuditLogRetention time.Duration `json:"audit_log_retention"`
	AuditLogLevel     string        `json:"audit_log_level"`

	// Performance settings
	MaxResponseTime      time.Duration `json:"max_response_time"`
	MaxMemoryUsage       int64         `json:"max_memory_usage"`
	PerformanceThreshold float64       `json:"performance_threshold"`

	// Report settings
	ReportFormat           []string `json:"report_format"`
	ReportDirectory        string   `json:"report_directory"`
	IncludeDetails         bool     `json:"include_details"`
	IncludeRecommendations bool     `json:"include_recommendations"`

	// Compliance settings
	ComplianceFrameworks []string `json:"compliance_frameworks"`
	SecurityStandards    []string `json:"security_standards"`
	RequireCompliance    bool     `json:"require_compliance"`
}

// DefaultSecurityTestConfig returns the default security test configuration
func DefaultSecurityTestConfig() *SecurityTestConfig {
	return &SecurityTestConfig{
		// Test execution settings
		Timeout:        30 * time.Second,
		RetryAttempts:  3,
		RetryDelay:     1 * time.Second,
		ParallelTests:  true,
		MaxConcurrency: 10,

		// Authentication settings
		JWTSecret:         "test-jwt-secret-key-for-security-testing",
		APIKeySecret:      "test-api-key-secret-for-security-testing",
		TokenExpiration:   15 * time.Minute,
		RefreshExpiration: 7 * 24 * time.Hour,

		// Rate limiting settings
		RateLimitEnabled:      true,
		RequestsPerMinute:     100,
		BurstSize:             20,
		RateLimitTestRequests: 10,

		// Security headers settings
		RequiredHeaders: []string{
			"X-Content-Type-Options",
			"X-Frame-Options",
			"X-XSS-Protection",
			"Strict-Transport-Security",
		},
		OptionalHeaders: []string{
			"Content-Security-Policy",
			"Referrer-Policy",
			"Permissions-Policy",
		},
		HeaderValidationStrict: false,

		// Input validation settings
		SQLInjectionTests: []string{
			"'; DROP TABLE users; --",
			"' OR '1'='1",
			"'; INSERT INTO users VALUES ('hacker', 'password'); --",
			"' UNION SELECT * FROM users --",
			"'; UPDATE users SET password='hacked' WHERE id=1; --",
		},
		XSSAttackTests: []string{
			"<script>alert('xss')</script>",
			"<img src=x onerror=alert('xss')>",
			"javascript:alert('xss')",
			"<svg onload=alert('xss')>",
			"<iframe src=javascript:alert('xss')></iframe>",
		},
		PathTraversalTests: []string{
			"../../../etc/passwd",
			"..\\..\\..\\windows\\system32\\drivers\\etc\\hosts",
			"....//....//....//etc/passwd",
			"%2e%2e%2f%2e%2e%2f%2e%2e%2fetc%2fpasswd",
		},
		CommandInjectionTests: []string{
			"; ls -la",
			"| cat /etc/passwd",
			"&& whoami",
			"`id`",
			"$(whoami)",
		},

		// Audit logging settings
		AuditLogEnabled:   true,
		AuditLogRetention: 90 * 24 * time.Hour,
		AuditLogLevel:     "INFO",

		// Performance settings
		MaxResponseTime:      200 * time.Millisecond,
		MaxMemoryUsage:       100 * 1024 * 1024, // 100MB
		PerformanceThreshold: 0.95,

		// Report settings
		ReportFormat:           []string{"json", "markdown", "summary"},
		ReportDirectory:        "test/reports/security",
		IncludeDetails:         true,
		IncludeRecommendations: true,

		// Compliance settings
		ComplianceFrameworks: []string{
			"SOC2",
			"GDPR",
			"PCI-DSS",
			"ISO27001",
		},
		SecurityStandards: []string{
			"OWASP Top 10",
			"NIST Cybersecurity Framework",
			"CIS Controls",
		},
		RequireCompliance: false,
	}
}

// SecurityTestCategories defines the categories of security tests
type SecurityTestCategories struct {
	Authentication  bool `json:"authentication"`
	Authorization   bool `json:"authorization"`
	DataAccess      bool `json:"data_access"`
	AuditLogging    bool `json:"audit_logging"`
	InputValidation bool `json:"input_validation"`
	RateLimiting    bool `json:"rate_limiting"`
	SecurityHeaders bool `json:"security_headers"`
	Performance     bool `json:"performance"`
	Compliance      bool `json:"compliance"`
}

// DefaultSecurityTestCategories returns the default test categories
func DefaultSecurityTestCategories() *SecurityTestCategories {
	return &SecurityTestCategories{
		Authentication:  true,
		Authorization:   true,
		DataAccess:      true,
		AuditLogging:    true,
		InputValidation: true,
		RateLimiting:    true,
		SecurityHeaders: true,
		Performance:     false, // Optional for security testing
		Compliance:      false, // Optional for security testing
	}
}

// SecurityTestThresholds defines thresholds for security test results
type SecurityTestThresholds struct {
	MinPassRate           float64 `json:"min_pass_rate"`
	MaxCriticalFailures   int     `json:"max_critical_failures"`
	MaxHighSeverityIssues int     `json:"max_high_severity_issues"`
	MinSecurityScore      int     `json:"min_security_score"`
	MaxResponseTime       int     `json:"max_response_time_ms"`
	MaxMemoryUsage        int64   `json:"max_memory_usage_bytes"`
}

// DefaultSecurityTestThresholds returns the default security test thresholds
func DefaultSecurityTestThresholds() *SecurityTestThresholds {
	return &SecurityTestThresholds{
		MinPassRate:           0.80,              // 80% pass rate required
		MaxCriticalFailures:   0,                 // No critical failures allowed
		MaxHighSeverityIssues: 2,                 // Max 2 high severity issues
		MinSecurityScore:      70,                // Minimum security score of 70/100
		MaxResponseTime:       200,               // Max 200ms response time
		MaxMemoryUsage:        100 * 1024 * 1024, // Max 100MB memory usage
	}
}

// SecurityTestEnvironment defines the test environment configuration
type SecurityTestEnvironment struct {
	Name        string            `json:"name"`
	BaseURL     string            `json:"base_url"`
	DatabaseURL string            `json:"database_url"`
	Environment string            `json:"environment"` // dev, staging, production
	Variables   map[string]string `json:"variables"`
	Secrets     map[string]string `json:"secrets,omitempty"`
}

// DefaultSecurityTestEnvironment returns the default test environment
func DefaultSecurityTestEnvironment() *SecurityTestEnvironment {
	return &SecurityTestEnvironment{
		Name:        "KYB Platform Security Tests",
		BaseURL:     "http://localhost:8080",
		DatabaseURL: "postgres://test:test@localhost:5432/kyb_test",
		Environment: "test",
		Variables: map[string]string{
			"LOG_LEVEL":    "debug",
			"ENVIRONMENT":  "test",
			"TEST_MODE":    "true",
			"DISABLE_AUTH": "false",
		},
		Secrets: map[string]string{
			"JWT_SECRET":     "test-jwt-secret-key",
			"API_KEY_SECRET": "test-api-key-secret",
		},
	}
}

// SecurityTestSuiteConfig combines all security test configuration
type SecurityTestSuiteConfig struct {
	Config      *SecurityTestConfig      `json:"config"`
	Categories  *SecurityTestCategories  `json:"categories"`
	Thresholds  *SecurityTestThresholds  `json:"thresholds"`
	Environment *SecurityTestEnvironment `json:"environment"`
}

// DefaultSecurityTestSuiteConfig returns the default security test suite configuration
func DefaultSecurityTestSuiteConfig() *SecurityTestSuiteConfig {
	return &SecurityTestSuiteConfig{
		Config:      DefaultSecurityTestConfig(),
		Categories:  DefaultSecurityTestCategories(),
		Thresholds:  DefaultSecurityTestThresholds(),
		Environment: DefaultSecurityTestEnvironment(),
	}
}

// ValidateConfig validates the security test configuration
func (config *SecurityTestSuiteConfig) ValidateConfig() error {
	// Validate timeout
	if config.Config.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	// Validate retry attempts
	if config.Config.RetryAttempts < 0 {
		return fmt.Errorf("retry attempts must be non-negative")
	}

	// Validate rate limiting
	if config.Config.RateLimitEnabled {
		if config.Config.RequestsPerMinute <= 0 {
			return fmt.Errorf("requests per minute must be positive when rate limiting is enabled")
		}
		if config.Config.BurstSize <= 0 {
			return fmt.Errorf("burst size must be positive when rate limiting is enabled")
		}
	}

	// Validate thresholds
	if config.Thresholds.MinPassRate < 0 || config.Thresholds.MinPassRate > 1 {
		return fmt.Errorf("min pass rate must be between 0 and 1")
	}

	if config.Thresholds.MinSecurityScore < 0 || config.Thresholds.MinSecurityScore > 100 {
		return fmt.Errorf("min security score must be between 0 and 100")
	}

	// Validate environment
	if config.Environment.BaseURL == "" {
		return fmt.Errorf("base URL must be specified")
	}

	return nil
}

// GetTestCategories returns the enabled test categories
func (config *SecurityTestSuiteConfig) GetTestCategories() []string {
	var categories []string

	if config.Categories.Authentication {
		categories = append(categories, "AUTHENTICATION")
	}
	if config.Categories.Authorization {
		categories = append(categories, "AUTHORIZATION")
	}
	if config.Categories.DataAccess {
		categories = append(categories, "DATA_ACCESS")
	}
	if config.Categories.AuditLogging {
		categories = append(categories, "AUDIT_LOGGING")
	}
	if config.Categories.InputValidation {
		categories = append(categories, "INPUT_VALIDATION")
	}
	if config.Categories.RateLimiting {
		categories = append(categories, "RATE_LIMITING")
	}
	if config.Categories.SecurityHeaders {
		categories = append(categories, "SECURITY_HEADERS")
	}
	if config.Categories.Performance {
		categories = append(categories, "PERFORMANCE")
	}
	if config.Categories.Compliance {
		categories = append(categories, "COMPLIANCE")
	}

	return categories
}

// IsCategoryEnabled checks if a specific test category is enabled
func (config *SecurityTestSuiteConfig) IsCategoryEnabled(category string) bool {
	categories := config.GetTestCategories()
	for _, cat := range categories {
		if cat == category {
			return true
		}
	}
	return false
}
