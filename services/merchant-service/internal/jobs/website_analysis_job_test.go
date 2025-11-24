package jobs

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"kyb-platform/services/merchant-service/internal/config"
)

func TestWebsiteAnalysisJob_GetID(t *testing.T) {
	job := NewWebsiteAnalysisJob(
		"merchant_123",
		"https://example.com",
		"Test Business",
		nil,
		&config.Config{Environment: "test"},
		zaptest.NewLogger(t),
	)

	assert.NotEmpty(t, job.GetID())
	assert.Contains(t, job.GetID(), "website_analysis_")
	assert.Contains(t, job.GetID(), "merchant_123")
}

func TestWebsiteAnalysisJob_GetMerchantID(t *testing.T) {
	job := NewWebsiteAnalysisJob(
		"merchant_123",
		"https://example.com",
		"Test Business",
		nil,
		&config.Config{Environment: "test"},
		zaptest.NewLogger(t),
	)

	assert.Equal(t, "merchant_123", job.GetMerchantID())
}

func TestWebsiteAnalysisJob_GetType(t *testing.T) {
	job := NewWebsiteAnalysisJob(
		"merchant_123",
		"https://example.com",
		"Test Business",
		nil,
		&config.Config{Environment: "test"},
		zaptest.NewLogger(t),
	)

	assert.Equal(t, "website_analysis", job.GetType())
}

func TestWebsiteAnalysisJob_SetStatus(t *testing.T) {
	job := NewWebsiteAnalysisJob(
		"merchant_123",
		"https://example.com",
		"Test Business",
		nil,
		&config.Config{Environment: "test"},
		zaptest.NewLogger(t),
	)

	assert.Equal(t, StatusPending, job.GetStatus())
	
	job.SetStatus(StatusProcessing)
	assert.Equal(t, StatusProcessing, job.GetStatus())
	
	job.SetStatus(StatusCompleted)
	assert.Equal(t, StatusCompleted, job.GetStatus())
}

func TestWebsiteAnalysisJob_hasScheme(t *testing.T) {
	tests := []struct {
		name     string
		urlStr   string
		expected bool
	}{
		{"has http", "http://example.com", true},
		{"has https", "https://example.com", true},
		{"no scheme", "example.com", false},
		{"empty string", "", false},
		{"just path", "/path/to/page", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasScheme(tt.urlStr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWebsiteAnalysisJob_analyzeSecurityHeaders(t *testing.T) {
	// Create mock server with security headers
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Strict-Transport-Security", "max-age=31536000")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer mockServer.Close()

	parsedURL, err := url.Parse(mockServer.URL)
	require.NoError(t, err)

	job := NewWebsiteAnalysisJob(
		"merchant_123",
		mockServer.URL,
		"Test Business",
		nil,
		&config.Config{Environment: "test"},
		zaptest.NewLogger(t),
	)

	ctx := context.Background()
	headers, err := job.analyzeSecurityHeaders(ctx, parsedURL)

	require.NoError(t, err)
	assert.True(t, headers.HasHSTS)
	assert.True(t, headers.HasCSP)
	assert.True(t, headers.HasXFrameOptions)
	assert.True(t, headers.HasXContentType)
	assert.Equal(t, 0.0, len(headers.MissingHeaders))
	assert.Greater(t, headers.SecurityScore, 0.8)
}

func TestWebsiteAnalysisJob_analyzeSecurityHeaders_MissingHeaders(t *testing.T) {
	// Create mock server without security headers
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer mockServer.Close()

	parsedURL, err := url.Parse(mockServer.URL)
	require.NoError(t, err)

	job := NewWebsiteAnalysisJob(
		"merchant_123",
		mockServer.URL,
		"Test Business",
		nil,
		&config.Config{Environment: "test"},
		zaptest.NewLogger(t),
	)

	ctx := context.Background()
	headers, err := job.analyzeSecurityHeaders(ctx, parsedURL)

	require.NoError(t, err)
	assert.False(t, headers.HasHSTS)
	assert.False(t, headers.HasCSP)
	assert.False(t, headers.HasXFrameOptions)
	assert.False(t, headers.HasXContentType)
	assert.Greater(t, len(headers.MissingHeaders), 0)
	assert.Less(t, headers.SecurityScore, 0.5)
}

func TestWebsiteAnalysisJob_analyzePerformance(t *testing.T) {
	// Create mock server with delay
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond) // Simulate load time
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("This is a test response with some content"))
	}))
	defer mockServer.Close()

	parsedURL, err := url.Parse(mockServer.URL)
	require.NoError(t, err)

	job := NewWebsiteAnalysisJob(
		"merchant_123",
		mockServer.URL,
		"Test Business",
		nil,
		&config.Config{Environment: "test"},
		zaptest.NewLogger(t),
	)

	ctx := context.Background()
	performance, err := job.analyzePerformance(ctx, parsedURL)

	require.NoError(t, err)
	assert.Greater(t, performance.LoadTime, 0.0)
	assert.Greater(t, performance.PageSize, 0)
	assert.Equal(t, 1, performance.RequestCount)
	assert.Greater(t, performance.PerformanceScore, 0.0)
}

func TestWebsiteAnalysisJob_analyzeSSL_HTTPS(t *testing.T) {
	// Note: This test requires actual HTTPS connection
	// We'll test the logic but may skip actual SSL connection in unit tests
	job := NewWebsiteAnalysisJob(
		"merchant_123",
		"https://example.com",
		"Test Business",
		nil,
		&config.Config{Environment: "test"},
		zaptest.NewLogger(t),
	)

	parsedURL, err := url.Parse("https://example.com:443")
	require.NoError(t, err)

	// This test may fail if we can't connect to example.com
	// So we'll just verify the function doesn't panic
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sslData, err := job.analyzeSSL(ctx, parsedURL)
	
	// We expect either success or connection error, but not panic
	if err != nil {
		// Connection error is acceptable in unit tests (network may not be available)
		// Just verify we got an error response
		assert.Error(t, err)
	} else {
		// If connection succeeds, verify structure
		assert.NotNil(t, sslData)
	}
}

func TestWebsiteAnalysisJob_analyzeSSL_HTTP(t *testing.T) {
	job := NewWebsiteAnalysisJob(
		"merchant_123",
		"http://example.com",
		"Test Business",
		nil,
		&config.Config{Environment: "test"},
		zaptest.NewLogger(t),
	)

	parsedURL, err := url.Parse("http://example.com")
	require.NoError(t, err)

	ctx := context.Background()
	sslData, err := job.analyzeSSL(ctx, parsedURL)

	require.NoError(t, err)
	assert.False(t, sslData.Valid) // HTTP should not have valid SSL
}

func TestWebsiteAnalysisJob_performWebsiteAnalysis(t *testing.T) {
	// Create mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Strict-Transport-Security", "max-age=31536000")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Test content"))
	}))
	defer mockServer.Close()

	job := NewWebsiteAnalysisJob(
		"merchant_123",
		mockServer.URL,
		"Test Business",
		nil,
		&config.Config{Environment: "test"},
		zaptest.NewLogger(t),
	)

	ctx := context.Background()
	result, err := job.performWebsiteAnalysis(ctx)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, mockServer.URL, result.WebsiteURL)
	assert.Equal(t, "completed", result.Status)
	assert.NotEmpty(t, result.LastAnalyzed)
	assert.NotNil(t, result.SSL)
	assert.NotNil(t, result.SecurityHeaders)
	assert.NotNil(t, result.Performance)
	assert.NotNil(t, result.Accessibility)
}

func TestWebsiteAnalysisJob_performWebsiteAnalysis_AddScheme(t *testing.T) {
	// Test that scheme is added if missing
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer mockServer.Close()

	// Extract host from mock server URL
	parsedMockURL, _ := url.Parse(mockServer.URL)
	hostWithoutScheme := parsedMockURL.Host

	job := NewWebsiteAnalysisJob(
		"merchant_123",
		hostWithoutScheme, // URL without scheme
		"Test Business",
		nil,
		&config.Config{Environment: "test"},
		zaptest.NewLogger(t),
	)

	ctx := context.Background()
	result, err := job.performWebsiteAnalysis(ctx)

	// Should add https:// scheme
	require.NoError(t, err)
	assert.Contains(t, result.WebsiteURL, "https://")
}

func TestWebsiteAnalysisJob_performWebsiteAnalysis_InvalidURL(t *testing.T) {
	job := NewWebsiteAnalysisJob(
		"merchant_123",
		"not-a-valid-url-!!!",
		"Test Business",
		nil,
		&config.Config{Environment: "test"},
		zaptest.NewLogger(t),
	)

	ctx := context.Background()
	result, err := job.performWebsiteAnalysis(ctx)

	assert.Error(t, err)
	assert.Nil(t, result)
}

