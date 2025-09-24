package risk_assessment

import (
	"context"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestFreeDomainAnalyzer_AnalyzeDomainFree(t *testing.T) {
	config := &RiskAssessmentConfig{
		WHOISLookupEnabled: true,
		RequestTimeout:     30 * time.Second,
	}
	logger := zap.NewNop()
	analyzer := NewFreeDomainAnalyzer(config, logger)

	tests := []struct {
		name           string
		domainName     string
		expectError    bool
		expectedFields []string
	}{
		{
			name:        "valid domain analysis",
			domainName:  "example.com",
			expectError: false,
			expectedFields: []string{
				"DomainName",
				"LastUpdated",
			},
		},
		{
			name:        "domain with subdomain",
			domainName:  "www.example.com",
			expectError: false,
			expectedFields: []string{
				"DomainName",
				"LastUpdated",
			},
		},
		{
			name:        "domain with protocol",
			domainName:  "https://example.com",
			expectError: false,
			expectedFields: []string{
				"DomainName",
				"LastUpdated",
			},
		},
		{
			name:        "invalid domain",
			domainName:  "invalid-domain-that-does-not-exist-12345.com",
			expectError: false, // Should not error, but may have risk factors
			expectedFields: []string{
				"DomainName",
				"LastUpdated",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := analyzer.AnalyzeDomainFree(ctx, tt.domainName)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if result != nil {
				// Check expected fields are present
				for _, field := range tt.expectedFields {
					switch field {
					case "DomainName":
						if result.DomainName == "" {
							t.Errorf("Expected DomainName to be set")
						}
					case "LastUpdated":
						if result.LastUpdated.IsZero() {
							t.Errorf("Expected LastUpdated to be set")
						}
					}
				}

				// Verify domain name extraction
				expectedDomain := analyzer.extractDomainName(tt.domainName)
				if result.DomainName != expectedDomain {
					t.Errorf("Expected domain name %s, got %s", expectedDomain, result.DomainName)
				}

				// Verify overall score is within valid range
				if result.OverallScore < 0.0 || result.OverallScore > 1.0 {
					t.Errorf("Overall score %f is outside valid range [0.0, 1.0]", result.OverallScore)
				}
			}
		})
	}
}

func TestFreeDomainAnalyzer_extractDomainName(t *testing.T) {
	analyzer := &FreeDomainAnalyzer{}

	tests := []struct {
		input    string
		expected string
	}{
		{"example.com", "example.com"},
		{"www.example.com", "example.com"},
		{"https://example.com", "example.com"},
		{"http://example.com", "example.com"},
		{"https://www.example.com", "example.com"},
		{"https://example.com/path", "example.com"},
		{"https://example.com:8080", "example.com"},
		{"https://example.com:8080/path", "example.com"},
		{"EXAMPLE.COM", "example.com"},
		{"WWW.EXAMPLE.COM", "example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := analyzer.extractDomainName(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestFreeDomainAnalyzer_calculateDomainAge(t *testing.T) {
	analyzer := &FreeDomainAnalyzer{}

	now := time.Now()
	creationDate := now.AddDate(-2, 0, 0)  // 2 years ago
	expirationDate := now.AddDate(8, 0, 0) // 8 years from now

	domainAge, err := analyzer.calculateDomainAge(creationDate, &expirationDate)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check age calculations
	expectedAgeInDays := 2 * 365
	if domainAge.AgeInDays < expectedAgeInDays-10 || domainAge.AgeInDays > expectedAgeInDays+10 {
		t.Errorf("Expected age around %d days, got %d", expectedAgeInDays, domainAge.AgeInDays)
	}

	expectedAgeInYears := 2.0
	if domainAge.AgeInYears < expectedAgeInYears-0.1 || domainAge.AgeInYears > expectedAgeInYears+0.1 {
		t.Errorf("Expected age around %.1f years, got %.1f", expectedAgeInYears, domainAge.AgeInYears)
	}

	// Check flags
	if domainAge.IsNewDomain {
		t.Errorf("Domain should not be marked as new (age: %d days)", domainAge.AgeInDays)
	}

	if domainAge.IsExpiringSoon {
		t.Errorf("Domain should not be expiring soon (expires in %d days)", domainAge.DaysToExpiry)
	}

	// Check age score
	if domainAge.AgeScore < 0.0 || domainAge.AgeScore > 1.0 {
		t.Errorf("Age score %f is outside valid range [0.0, 1.0]", domainAge.AgeScore)
	}
}

func TestFreeDomainAnalyzer_calculateAgeScore(t *testing.T) {
	analyzer := &FreeDomainAnalyzer{}

	tests := []struct {
		name     string
		age      *DomainAge
		expected float64
	}{
		{
			name: "new domain",
			age: &DomainAge{
				AgeInDays:      15,
				AgeInYears:     0.04,
				IsNewDomain:    true,
				IsExpiringSoon: false,
			},
			expected: 0.3, // 0.5 - 0.2 for new domain
		},
		{
			name: "mature domain",
			age: &DomainAge{
				AgeInDays:      1000,
				AgeInYears:     2.7,
				IsNewDomain:    false,
				IsExpiringSoon: false,
			},
			expected: 0.8, // 0.5 + 0.3 for mature domain
		},
		{
			name: "expiring domain",
			age: &DomainAge{
				AgeInDays:      1000,
				AgeInYears:     2.7,
				IsNewDomain:    false,
				IsExpiringSoon: true,
			},
			expected: 0.5, // 0.8 - 0.3 for expiring
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := analyzer.calculateAgeScore(tt.age)
			if score < tt.expected-0.1 || score > tt.expected+0.1 {
				t.Errorf("Expected score around %.1f, got %.1f", tt.expected, score)
			}
		})
	}
}

func TestFreeDomainAnalyzer_analyzeRegistrarInfo(t *testing.T) {
	analyzer := &FreeDomainAnalyzer{}

	tests := []struct {
		name              string
		registrarName     string
		expectedReputable bool
	}{
		{
			name:              "reputable registrar",
			registrarName:     "GoDaddy.com, LLC",
			expectedReputable: true,
		},
		{
			name:              "another reputable registrar",
			registrarName:     "Namecheap, Inc.",
			expectedReputable: true,
		},
		{
			name:              "unknown registrar",
			registrarName:     "Unknown Registrar Corp",
			expectedReputable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.analyzeRegistrarInfo(tt.registrarName)

			if result.RegistrarName != tt.registrarName {
				t.Errorf("Expected registrar name %s, got %s", tt.registrarName, result.RegistrarName)
			}

			if result.IsReputable != tt.expectedReputable {
				t.Errorf("Expected reputable %v, got %v", tt.expectedReputable, result.IsReputable)
			}

			// Check score is within valid range
			if result.RegistrarScore < 0.0 || result.RegistrarScore > 1.0 {
				t.Errorf("Registrar score %f is outside valid range [0.0, 1.0]", result.RegistrarScore)
			}
		})
	}
}

func TestFreeDomainAnalyzer_calculateOverallScore(t *testing.T) {
	analyzer := &FreeDomainAnalyzer{}

	// Create a test result with various components
	result := &DomainAnalysisResult{
		DomainName: "example.com",
		WHOISInfo: &WHOISInfo{
			DNSSEC: true,
			Status: []string{"clientTransferProhibited"},
		},
		DomainAge: &DomainAge{
			AgeScore: 0.8,
		},
		RegistrarInfo: &RegistrarInfo{
			RegistrarScore: 0.9,
		},
		DNSInfo: &DNSInfo{
			DNSScore: 0.7,
		},
	}

	score := analyzer.calculateOverallScore(result)

	// Score should be within valid range
	if score < 0.0 || score > 1.0 {
		t.Errorf("Overall score %f is outside valid range [0.0, 1.0]", score)
	}

	// With good components, score should be reasonably high
	if score < 0.6 {
		t.Errorf("Expected score to be reasonably high with good components, got %f", score)
	}
}

func TestFreeDomainAnalyzer_generateRecommendations(t *testing.T) {
	analyzer := &FreeDomainAnalyzer{}

	tests := []struct {
		name             string
		result           *DomainAnalysisResult
		expectedCount    int
		expectedKeywords []string
	}{
		{
			name: "new domain",
			result: &DomainAnalysisResult{
				DomainAge: &DomainAge{
					IsNewDomain: true,
				},
			},
			expectedCount:    1,
			expectedKeywords: []string{"new"},
		},
		{
			name: "expiring domain",
			result: &DomainAnalysisResult{
				DomainAge: &DomainAge{
					IsExpiringSoon: true,
				},
			},
			expectedCount:    1,
			expectedKeywords: []string{"expires"},
		},
		{
			name: "unreputable registrar",
			result: &DomainAnalysisResult{
				RegistrarInfo: &RegistrarInfo{
					IsReputable: false,
				},
			},
			expectedCount:    1,
			expectedKeywords: []string{"registrar"},
		},
		{
			name: "no MX records",
			result: &DomainAnalysisResult{
				DNSInfo: &DNSInfo{
					MXRecords:     []string{},
					DNSSECEnabled: false,
				},
			},
			expectedCount:    2, // MX and DNSSEC recommendations
			expectedKeywords: []string{"MX", "DNSSEC"},
		},
		{
			name: "comprehensive issues",
			result: &DomainAnalysisResult{
				DomainAge: &DomainAge{
					IsNewDomain:    true,
					IsExpiringSoon: true,
				},
				RegistrarInfo: &RegistrarInfo{
					IsReputable: false,
				},
				DNSInfo: &DNSInfo{
					MXRecords:     []string{},
					DNSSECEnabled: false,
				},
			},
			expectedCount:    5, // Multiple recommendations
			expectedKeywords: []string{"new", "expires", "registrar", "MX", "DNSSEC"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recommendations := analyzer.generateRecommendations(tt.result)

			if len(recommendations) != tt.expectedCount {
				t.Errorf("Expected %d recommendations, got %d", tt.expectedCount, len(recommendations))
			}

			// Check that expected keywords are present in recommendations
			recommendationText := ""
			for _, rec := range recommendations {
				recommendationText += rec + " "
			}

			for _, keyword := range tt.expectedKeywords {
				if !strings.Contains(strings.ToLower(recommendationText), strings.ToLower(keyword)) {
					t.Errorf("Expected keyword '%s' not found in recommendations: %v", keyword, recommendations)
				}
			}
		})
	}
}

func TestFreeDomainAnalyzer_parseDate(t *testing.T) {
	analyzer := &FreeDomainAnalyzer{}

	tests := []struct {
		name        string
		dateStr     string
		expectError bool
	}{
		{
			name:        "ISO format",
			dateStr:     "2022-01-15T10:30:00Z",
			expectError: false,
		},
		{
			name:        "simple date",
			dateStr:     "2022-01-15",
			expectError: false,
		},
		{
			name:        "space separated",
			dateStr:     "2022-01-15 10:30:00",
			expectError: false,
		},
		{
			name:        "invalid format",
			dateStr:     "invalid-date",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			date, err := analyzer.parseDate(tt.dateStr)

			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if !tt.expectError && date.IsZero() {
				t.Errorf("Expected valid date but got zero time")
			}
		})
	}
}

// Benchmark tests
func BenchmarkFreeDomainAnalyzer_AnalyzeDomainFree(b *testing.B) {
	config := &RiskAssessmentConfig{
		WHOISLookupEnabled: true,
		RequestTimeout:     30 * time.Second,
	}
	logger := zap.NewNop()
	analyzer := NewFreeDomainAnalyzer(config, logger)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = analyzer.AnalyzeDomainFree(ctx, "example.com")
	}
}

func BenchmarkFreeDomainAnalyzer_extractDomainName(b *testing.B) {
	analyzer := &FreeDomainAnalyzer{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = analyzer.extractDomainName("https://www.example.com/path?query=value")
	}
}
