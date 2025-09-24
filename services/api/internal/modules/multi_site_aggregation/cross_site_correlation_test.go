package multi_site_aggregation

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestCrossSiteCorrelationService_AnalyzeCorrelations(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultCorrelationConfig()
	service := NewCrossSiteCorrelationService(config, logger)

	tests := []struct {
		name        string
		businessID  string
		sitesData   []SiteData
		expectError bool
		checkResult func(*testing.T, *CorrelationAnalysis)
	}{
		{
			name:       "successful correlation analysis with multiple sites",
			businessID: "business-123",
			sitesData: []SiteData{
				{
					ID:         "site-1",
					LocationID: "location-1",
					BusinessID: "business-123",
					DataType:   "contact_info",
					ExtractedData: map[string]interface{}{
						"phone":     "+1-555-123-4567",
						"email":     "contact@example.com",
						"employees": 50,
						"revenue":   1000000.0,
					},
					ConfidenceScore:  0.9,
					ExtractionMethod: "web_scraping",
					LastExtracted:    time.Now(),
					DataQuality:      0.85,
					IsValid:          true,
					CreatedAt:        time.Now(),
					UpdatedAt:        time.Now(),
				},
				{
					ID:         "site-2",
					LocationID: "location-2",
					BusinessID: "business-123",
					DataType:   "contact_info",
					ExtractedData: map[string]interface{}{
						"phone":     "+1-555-123-4567",
						"email":     "contact@example.com",
						"employees": 45,
						"revenue":   950000.0,
					},
					ConfidenceScore:  0.85,
					ExtractionMethod: "web_scraping",
					LastExtracted:    time.Now(),
					DataQuality:      0.80,
					IsValid:          true,
					CreatedAt:        time.Now(),
					UpdatedAt:        time.Now(),
				},
				{
					ID:         "site-3",
					LocationID: "location-3",
					BusinessID: "business-123",
					DataType:   "contact_info",
					ExtractedData: map[string]interface{}{
						"phone":     "+1-555-123-4567",
						"email":     "contact@example.com",
						"employees": 55,
						"revenue":   1050000.0,
					},
					ConfidenceScore:  0.88,
					ExtractionMethod: "web_scraping",
					LastExtracted:    time.Now(),
					DataQuality:      0.82,
					IsValid:          true,
					CreatedAt:        time.Now(),
					UpdatedAt:        time.Now(),
				},
			},
			expectError: false,
			checkResult: func(t *testing.T, result *CorrelationAnalysis) {
				assert.NotNil(t, result)
				assert.Equal(t, "business-123", result.BusinessID)
				assert.NotEmpty(t, result.ID)
				assert.NotZero(t, result.ConfidenceScore)
				assert.NotEmpty(t, result.CorrelationMatrix)
				assert.NotEmpty(t, result.DataPatterns)
				assert.NotEmpty(t, result.Insights)
				assert.Greater(t, result.ProcessingTime, time.Duration(0))
				assert.Equal(t, "comprehensive_correlation", result.AnalysisMethod)
			},
		},
		{
			name:       "insufficient data for correlation analysis",
			businessID: "business-456",
			sitesData: []SiteData{
				{
					ID:         "site-1",
					LocationID: "location-1",
					BusinessID: "business-456",
					DataType:   "contact_info",
					ExtractedData: map[string]interface{}{
						"phone": "+1-555-123-4567",
					},
					ConfidenceScore:  0.9,
					ExtractionMethod: "web_scraping",
					LastExtracted:    time.Now(),
					DataQuality:      0.85,
					IsValid:          true,
					CreatedAt:        time.Now(),
					UpdatedAt:        time.Now(),
				},
			},
			expectError: true,
			checkResult: func(t *testing.T, result *CorrelationAnalysis) {
				// Should not be called for error case
			},
		},
		{
			name:        "empty sites data",
			businessID:  "business-789",
			sitesData:   []SiteData{},
			expectError: true,
			checkResult: func(t *testing.T, result *CorrelationAnalysis) {
				// Should not be called for error case
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := service.AnalyzeCorrelations(ctx, tt.businessID, tt.sitesData)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				tt.checkResult(t, result)
			}
		})
	}
}

func TestCrossSiteCorrelationService_DetectPatterns(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultCorrelationConfig()
	service := NewCrossSiteCorrelationService(config, logger)

	tests := []struct {
		name      string
		sitesData []SiteData
		fields    []string
		expected  int // Expected number of patterns
	}{
		{
			name: "detect consistency patterns",
			sitesData: []SiteData{
				{
					ID:         "site-1",
					LocationID: "location-1",
					ExtractedData: map[string]interface{}{
						"phone": "+1-555-123-4567",
						"email": "contact@example.com",
					},
				},
				{
					ID:         "site-2",
					LocationID: "location-2",
					ExtractedData: map[string]interface{}{
						"phone": "+1-555-123-4567",
						"email": "contact@example.com",
					},
				},
			},
			fields:   []string{"phone", "email"},
			expected: 2, // One consistency pattern for each field
		},
		{
			name: "detect variation patterns",
			sitesData: []SiteData{
				{
					ID:         "site-1",
					LocationID: "location-1",
					ExtractedData: map[string]interface{}{
						"employees": 50,
						"revenue":   1000000.0,
					},
				},
				{
					ID:         "site-2",
					LocationID: "location-2",
					ExtractedData: map[string]interface{}{
						"employees": 45,
						"revenue":   950000.0,
					},
				},
				{
					ID:         "site-3",
					LocationID: "location-3",
					ExtractedData: map[string]interface{}{
						"employees": 55,
						"revenue":   1050000.0,
					},
				},
			},
			fields:   []string{"employees", "revenue"},
			expected: 2, // One variation pattern for each field
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patterns := service.detectPatterns(tt.sitesData, tt.fields)
			assert.Len(t, patterns, tt.expected)

			for _, pattern := range patterns {
				assert.NotEmpty(t, pattern.ID)
				assert.NotEmpty(t, pattern.FieldName)
				assert.NotEmpty(t, pattern.PatternType)
				assert.NotEmpty(t, pattern.Description)
				assert.NotEmpty(t, pattern.AffectedSites)
				assert.Greater(t, pattern.Confidence, 0.0)
				assert.LessOrEqual(t, pattern.Confidence, 1.0)
			}
		})
	}
}

func TestCrossSiteCorrelationService_DetectAnomalies(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultCorrelationConfig()
	service := NewCrossSiteCorrelationService(config, logger)

	tests := []struct {
		name      string
		sitesData []SiteData
		fields    []string
		expected  int // Expected number of anomalies
	}{
		{
			name: "detect outlier anomalies",
			sitesData: []SiteData{
				{
					ID:         "site-1",
					LocationID: "location-1",
					ExtractedData: map[string]interface{}{
						"employees": 50,
						"revenue":   1000000.0,
					},
				},
				{
					ID:         "site-2",
					LocationID: "location-2",
					ExtractedData: map[string]interface{}{
						"employees": 45,
						"revenue":   950000.0,
					},
				},
				{
					ID:         "site-3",
					LocationID: "location-3",
					ExtractedData: map[string]interface{}{
						"employees": 2000,       // Very extreme outlier
						"revenue":   50000000.0, // Very extreme outlier
					},
				},
			},
			fields:   []string{"employees", "revenue"},
			expected: 2, // One outlier for each field
		},
		{
			name: "detect missing value anomalies",
			sitesData: []SiteData{
				{
					ID:         "site-1",
					LocationID: "location-1",
					ExtractedData: map[string]interface{}{
						"phone": "+1-555-123-4567",
						"email": "contact@example.com",
					},
				},
				{
					ID:         "site-2",
					LocationID: "location-2",
					ExtractedData: map[string]interface{}{
						"phone": "+1-555-123-4567",
						// email is missing
					},
				},
			},
			fields:   []string{"phone", "email"},
			expected: 1, // One missing value anomaly for email
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			anomalies := service.detectAnomalies(tt.sitesData, tt.fields)
			assert.Len(t, anomalies, tt.expected)

			for _, anomaly := range anomalies {
				assert.NotEmpty(t, anomaly.ID)
				assert.NotEmpty(t, anomaly.FieldName)
				assert.NotEmpty(t, anomaly.AnomalyType)
				assert.NotEmpty(t, anomaly.Description)
				assert.NotEmpty(t, anomaly.AffectedSite)
				assert.NotEmpty(t, anomaly.Severity)
				assert.NotEmpty(t, anomaly.Recommendation)
			}
		})
	}
}

func TestCrossSiteCorrelationService_AnalyzeTrends(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultCorrelationConfig()
	service := NewCrossSiteCorrelationService(config, logger)

	now := time.Now()
	oneHourAgo := now.Add(-1 * time.Hour)
	twoHoursAgo := now.Add(-2 * time.Hour)

	tests := []struct {
		name      string
		sitesData []SiteData
		fields    []string
		expected  int // Expected number of trends
	}{
		{
			name: "detect increasing trend",
			sitesData: []SiteData{
				{
					ID:            "site-1",
					LocationID:    "location-1",
					LastExtracted: twoHoursAgo,
					ExtractedData: map[string]interface{}{
						"revenue": 1000000.0,
					},
				},
				{
					ID:            "site-2",
					LocationID:    "location-2",
					LastExtracted: oneHourAgo,
					ExtractedData: map[string]interface{}{
						"revenue": 1100000.0,
					},
				},
				{
					ID:            "site-3",
					LocationID:    "location-3",
					LastExtracted: now,
					ExtractedData: map[string]interface{}{
						"revenue": 1200000.0,
					},
				},
			},
			fields:   []string{"revenue"},
			expected: 1, // One increasing trend
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trends := service.analyzeTrends(tt.sitesData, tt.fields)
			assert.Len(t, trends, tt.expected)

			for _, trend := range trends {
				assert.NotEmpty(t, trend.ID)
				assert.NotEmpty(t, trend.FieldName)
				assert.NotEmpty(t, trend.TrendType)
				assert.NotEmpty(t, trend.Direction)
				assert.NotEmpty(t, trend.Description)
				assert.NotEmpty(t, trend.AffectedSites)
				assert.Greater(t, trend.Confidence, 0.0)
				assert.LessOrEqual(t, trend.Confidence, 1.0)
				assert.Greater(t, trend.Timeframe, time.Duration(0))
			}
		})
	}
}

func TestCrossSiteCorrelationService_GenerateInsights(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultCorrelationConfig()
	service := NewCrossSiteCorrelationService(config, logger)

	// Create sample data for insight generation
	sitesData := []SiteData{
		{
			ID:         "site-1",
			LocationID: "location-1",
			ExtractedData: map[string]interface{}{
				"employees": 50,
				"revenue":   1000000.0,
			},
		},
		{
			ID:         "site-2",
			LocationID: "location-2",
			ExtractedData: map[string]interface{}{
				"employees": 45,
				"revenue":   950000.0,
			},
		},
		{
			ID:         "site-3",
			LocationID: "location-3",
			ExtractedData: map[string]interface{}{
				"employees": 55,
				"revenue":   1050000.0,
			},
		},
	}

	// Calculate actual correlation matrix
	fields := []string{"employees", "revenue"}
	correlationMatrix := service.calculateCorrelationMatrix(sitesData, fields)

	// Generate patterns, anomalies, and trends
	patterns := service.detectPatterns(sitesData, fields)
	anomalies := service.detectAnomalies(sitesData, fields)
	trends := service.analyzeTrends(sitesData, fields)

	insights := service.generateInsights(sitesData, correlationMatrix, patterns, anomalies, trends)

	assert.NotEmpty(t, insights)

	for _, insight := range insights {
		assert.NotEmpty(t, insight.ID)
		assert.NotEmpty(t, insight.Title)
		assert.NotEmpty(t, insight.Description)
		assert.NotEmpty(t, insight.InsightType)
		assert.NotEmpty(t, insight.Recommendation)
		assert.NotEmpty(t, insight.RelatedFields)
		assert.Greater(t, insight.Confidence, 0.0)
		assert.LessOrEqual(t, insight.Confidence, 1.0)
		assert.NotEmpty(t, insight.Impact)
	}
}

func TestCrossSiteCorrelationService_CalculateCorrelationMatrix(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultCorrelationConfig()
	service := NewCrossSiteCorrelationService(config, logger)

	sitesData := []SiteData{
		{
			ID:         "site-1",
			LocationID: "location-1",
			ExtractedData: map[string]interface{}{
				"employees": 50,
				"revenue":   1000000.0,
				"phone":     "+1-555-123-4567",
			},
		},
		{
			ID:         "site-2",
			LocationID: "location-2",
			ExtractedData: map[string]interface{}{
				"employees": 45,
				"revenue":   950000.0,
				"phone":     "+1-555-123-4567",
			},
		},
		{
			ID:         "site-3",
			LocationID: "location-3",
			ExtractedData: map[string]interface{}{
				"employees": 55,
				"revenue":   1050000.0,
				"phone":     "+1-555-123-4567",
			},
		},
	}

	fields := []string{"employees", "revenue", "phone"}
	matrix := service.calculateCorrelationMatrix(sitesData, fields)

	assert.NotEmpty(t, matrix)
	assert.Len(t, matrix, 3) // Three fields

	// Check that each field has correlations with all other fields
	for _, field := range fields {
		assert.Contains(t, matrix, field)
		assert.Len(t, matrix[field], 3) // Including self-correlation

		// Self-correlation should be 1.0
		assert.Equal(t, 1.0, matrix[field][field])
	}

	// Check that numeric fields have correlations
	assert.NotEqual(t, 0.0, matrix["employees"]["revenue"])
	assert.NotEqual(t, 0.0, matrix["revenue"]["employees"])

	// String fields should have 0.0 correlation with numeric fields
	assert.Equal(t, 0.0, matrix["phone"]["employees"])
	assert.Equal(t, 0.0, matrix["phone"]["revenue"])
	assert.Equal(t, 0.0, matrix["employees"]["phone"])
	assert.Equal(t, 0.0, matrix["revenue"]["phone"])

	// Check that employees and revenue have a correlation (should be high)
	if correlation, exists := matrix["employees"]["revenue"]; exists {
		assert.Greater(t, correlation, 0.5) // Should be positively correlated
	}
}

func TestCrossSiteCorrelationService_CalculatePearsonCorrelation(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultCorrelationConfig()
	service := NewCrossSiteCorrelationService(config, logger)

	tests := []struct {
		name     string
		x        []float64
		y        []float64
		expected float64
	}{
		{
			name:     "perfect positive correlation",
			x:        []float64{1, 2, 3, 4, 5},
			y:        []float64{1, 2, 3, 4, 5},
			expected: 1.0,
		},
		{
			name:     "perfect negative correlation",
			x:        []float64{1, 2, 3, 4, 5},
			y:        []float64{5, 4, 3, 2, 1},
			expected: -1.0,
		},
		{
			name:     "no correlation",
			x:        []float64{1, 2, 3, 4, 5},
			y:        []float64{1, 1, 1, 1, 1},
			expected: 0.0,
		},
		{
			name:     "insufficient data",
			x:        []float64{1},
			y:        []float64{1},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.calculatePearsonCorrelation(tt.x, tt.y)
			assert.InDelta(t, tt.expected, result, 0.01)
		})
	}
}

func TestCrossSiteCorrelationService_ExtractNumericValue(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultCorrelationConfig()
	service := NewCrossSiteCorrelationService(config, logger)

	tests := []struct {
		name     string
		value    interface{}
		expected *float64
	}{
		{
			name:     "float64 value",
			value:    123.45,
			expected: func() *float64 { f := 123.45; return &f }(),
		},
		{
			name:     "int value",
			value:    123,
			expected: func() *float64 { f := 123.0; return &f }(),
		},
		{
			name:     "int64 value",
			value:    int64(123),
			expected: func() *float64 { f := 123.0; return &f }(),
		},
		{
			name:     "string numeric value",
			value:    "123.45",
			expected: func() *float64 { f := 123.45; return &f }(),
		},
		{
			name:     "string with formatting",
			value:    "$1,234.56",
			expected: func() *float64 { f := 1234.56; return &f }(),
		},
		{
			name:     "non-numeric string",
			value:    "hello",
			expected: nil,
		},
		{
			name:     "nil value",
			value:    nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.extractNumericValue(tt.value)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.InDelta(t, *tt.expected, *result, 0.01)
			}
		})
	}
}

func TestCrossSiteCorrelationService_CalculateConfidenceScore(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultCorrelationConfig()
	service := NewCrossSiteCorrelationService(config, logger)

	correlationMatrix := map[string]map[string]float64{
		"field1": {"field2": 0.8},
		"field2": {"field1": 0.8},
	}

	patterns := []DataPattern{
		{Confidence: 0.9},
		{Confidence: 0.85},
	}

	anomalies := []DataAnomaly{
		{Severity: "high"},
		{Severity: "medium"},
	}

	trends := []DataTrend{
		{Confidence: 0.8},
	}

	insights := []DataInsight{
		{Confidence: 0.75},
	}

	confidence := service.calculateConfidenceScore(correlationMatrix, patterns, anomalies, trends, insights)

	assert.Greater(t, confidence, 0.0)
	assert.LessOrEqual(t, confidence, 1.0)
}

func TestCrossSiteCorrelationService_UtilityMethods(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultCorrelationConfig()
	service := NewCrossSiteCorrelationService(config, logger)

	// Test calculateMean
	values := []float64{1, 2, 3, 4, 5}
	mean := service.calculateMean(values)
	assert.Equal(t, 3.0, mean)

	// Test calculateStandardDeviation
	stdDev := service.calculateStandardDeviation(values, mean)
	assert.InDelta(t, 1.58, stdDev, 0.01)

	// Test calculateLinearRegression
	x := []float64{1, 2, 3, 4, 5}
	y := []float64{2, 4, 6, 8, 10}
	slope, intercept := service.calculateLinearRegression(x, y)
	assert.InDelta(t, 2.0, slope, 0.01)
	assert.InDelta(t, 0.0, intercept, 0.01)

	// Test calculateRSquared
	rSquared := service.calculateRSquared(x, y, slope, intercept)
	assert.InDelta(t, 1.0, rSquared, 0.01)

	// Test determineAnomalySeverity
	assert.Equal(t, "low", service.determineAnomalySeverity(1.5))
	assert.Equal(t, "medium", service.determineAnomalySeverity(2.1))
	assert.Equal(t, "high", service.determineAnomalySeverity(2.6))
	assert.Equal(t, "critical", service.determineAnomalySeverity(3.1))

	// Test getCorrelationDirection
	assert.Equal(t, "positive", service.getCorrelationDirection(0.8))
	assert.Equal(t, "negative", service.getCorrelationDirection(-0.8))
	assert.Equal(t, "weak", service.getCorrelationDirection(0.5))
}

func TestParseNumericString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
		hasError bool
	}{
		{
			name:     "simple number",
			input:    "123.45",
			expected: 123.45,
			hasError: false,
		},
		{
			name:     "number with comma",
			input:    "1,234.56",
			expected: 1234.56,
			hasError: false,
		},
		{
			name:     "number with dollar sign",
			input:    "$123.45",
			expected: 123.45,
			hasError: false,
		},
		{
			name:     "number with percentage",
			input:    "50%",
			expected: 50.0,
			hasError: false,
		},
		{
			name:     "non-numeric string",
			input:    "hello",
			expected: 0.0,
			hasError: true,
		},
		{
			name:     "empty string",
			input:    "",
			expected: 0.0,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseNumericString(tt.input)
			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.InDelta(t, tt.expected, result, 0.01)
			}
		})
	}
}

func TestDefaultCorrelationConfig(t *testing.T) {
	config := DefaultCorrelationConfig()

	assert.NotNil(t, config)
	assert.Equal(t, 0.3, config.MinCorrelationThreshold)
	assert.Equal(t, 30*time.Second, config.MaxAnalysisTime)
	assert.True(t, config.EnablePatternDetection)
	assert.True(t, config.EnableAnomalyDetection)
	assert.True(t, config.EnableTrendAnalysis)
	assert.True(t, config.EnableInsightGeneration)
	assert.Equal(t, 0.7, config.ConfidenceThreshold)
	assert.Equal(t, 5, config.MaxPatternsPerField)
	assert.Equal(t, 10, config.MaxAnomaliesPerField)
	assert.Equal(t, 3, config.MaxTrendsPerField)
	assert.Equal(t, 10, config.MaxInsightsPerAnalysis)
}
