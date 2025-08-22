package risk_assessment

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewAntiDetectionService(t *testing.T) {
	config := &RiskAssessmentConfig{
		AntiDetectionConfig: AntiDetectionConfig{
			UserAgentRotationEnabled:   true,
			RequestDelayEnabled:        true,
			ProxyEnabled:               true,
			CustomHeadersEnabled:       true,
			DetectionMonitoringEnabled: true,
			MaxDetectionEvents:         100,
			DetectionAlertThreshold:    10,
		},
	}
	logger := zap.NewNop()

	service := NewAntiDetectionService(config, logger)
	require.NotNil(t, service)
	assert.NotEmpty(t, service.userAgents)
	assert.NotEmpty(t, service.proxies)
	assert.NotNil(t, service.requestDelays)
	assert.NotNil(t, service.detectionEvents)
}

func TestAntiDetectionService_CreateHTTPClient(t *testing.T) {
	config := &RiskAssessmentConfig{
		AntiDetectionConfig: AntiDetectionConfig{
			ProxyEnabled: true,
		},
	}
	logger := zap.NewNop()
	service := NewAntiDetectionService(config, logger)

	client, err := service.CreateHTTPClient(context.Background(), "https://example.com")
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, 30*time.Second, client.Timeout)
}

func TestAntiDetectionService_PrepareRequest(t *testing.T) {
	config := &RiskAssessmentConfig{
		AntiDetectionConfig: AntiDetectionConfig{
			UserAgentRotationEnabled: true,
			RequestDelayEnabled:      false, // Disable for testing
			CustomHeadersEnabled:     true,
			CustomHeaders: map[string]string{
				"Accept":          "text/html,application/xhtml+xml",
				"Accept-Language": "en-US,en;q=0.9",
			},
		},
	}
	logger := zap.NewNop()
	service := NewAntiDetectionService(config, logger)

	req, err := http.NewRequest("GET", "https://example.com", nil)
	require.NoError(t, err)

	err = service.PrepareRequest(context.Background(), req, "https://example.com")
	require.NoError(t, err)

	// Check that user agent is set
	assert.NotEmpty(t, req.Header.Get("User-Agent"))

	// Check that custom headers are set
	assert.Equal(t, "text/html,application/xhtml+xml", req.Header.Get("Accept"))
	assert.Equal(t, "en-US,en;q=0.9", req.Header.Get("Accept-Language"))

	// Check that referer is set
	assert.NotEmpty(t, req.Header.Get("Referer"))
}

func TestAntiDetectionService_MonitorResponse_Blocked(t *testing.T) {
	config := &RiskAssessmentConfig{
		AntiDetectionConfig: AntiDetectionConfig{
			DetectionMonitoringEnabled: true,
			MaxDetectionEvents:         10,
			DetectionAlertThreshold:    5,
		},
	}
	logger := zap.NewNop()
	service := NewAntiDetectionService(config, logger)

	// Create a test server that returns 403
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	req, err := http.NewRequest("GET", server.URL, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	err = service.MonitorResponse(context.Background(), req, resp, server.URL)
	require.NoError(t, err)

	report := service.GetDetectionReport()
	assert.Equal(t, 1, report.TotalEvents)
	assert.Equal(t, 1, report.EventsByType[DetectionEventBlocked])
	assert.Equal(t, 1, report.EventsBySeverity[DetectionSeverityHigh])
}

func TestAntiDetectionService_MonitorResponse_RateLimited(t *testing.T) {
	config := &RiskAssessmentConfig{
		AntiDetectionConfig: AntiDetectionConfig{
			DetectionMonitoringEnabled: true,
			MaxDetectionEvents:         10,
			DetectionAlertThreshold:    5,
		},
	}
	logger := zap.NewNop()
	service := NewAntiDetectionService(config, logger)

	// Create a test server that returns rate limiting headers
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "60")
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	req, err := http.NewRequest("GET", server.URL, nil)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	err = service.MonitorResponse(context.Background(), req, resp, server.URL)
	require.NoError(t, err)

	report := service.GetDetectionReport()
	assert.Equal(t, 1, report.TotalEvents)
	assert.Equal(t, 1, report.EventsByType[DetectionEventRateLimited])
	assert.Equal(t, 1, report.EventsBySeverity[DetectionSeverityMedium])
}

func TestAntiDetectionService_GetDetectionReport(t *testing.T) {
	config := &RiskAssessmentConfig{
		AntiDetectionConfig: AntiDetectionConfig{
			DetectionMonitoringEnabled: true,
			MaxDetectionEvents:         10,
			DetectionAlertThreshold:    5,
		},
	}
	logger := zap.NewNop()
	service := NewAntiDetectionService(config, logger)

	// Add some test events
	service.eventMutex.Lock()
	service.detectionEvents = []DetectionEvent{
		{
			Timestamp:   time.Now(),
			URL:         "https://example1.com",
			EventType:   DetectionEventBlocked,
			Severity:    DetectionSeverityHigh,
			Description: "Test blocked event",
		},
		{
			Timestamp:   time.Now(),
			URL:         "https://example2.com",
			EventType:   DetectionEventRateLimited,
			Severity:    DetectionSeverityMedium,
			Description: "Test rate limited event",
		},
	}
	service.eventMutex.Unlock()

	report := service.GetDetectionReport()
	assert.Equal(t, 2, report.TotalEvents)
	assert.Equal(t, 1, report.EventsByType[DetectionEventBlocked])
	assert.Equal(t, 1, report.EventsByType[DetectionEventRateLimited])
	assert.Equal(t, 1, report.EventsBySeverity[DetectionSeverityHigh])
	assert.Equal(t, 1, report.EventsBySeverity[DetectionSeverityMedium])
	assert.Len(t, report.RecentEvents, 2)
	assert.Greater(t, report.RiskScore, 0.0)
}

func TestAntiDetectionService_UserAgentRotation(t *testing.T) {
	config := &RiskAssessmentConfig{
		AntiDetectionConfig: AntiDetectionConfig{
			UserAgentRotationEnabled: true,
		},
	}
	logger := zap.NewNop()
	service := NewAntiDetectionService(config, logger)

	// Test user agent rotation
	ua1 := service.getNextUserAgent()
	ua2 := service.getNextUserAgent()
	ua3 := service.getNextUserAgent()

	assert.NotEmpty(t, ua1)
	assert.NotEmpty(t, ua2)
	assert.NotEmpty(t, ua3)

	// Should be different user agents (unless we've cycled through all of them)
	if len(service.userAgents) > 2 {
		assert.NotEqual(t, ua1, ua2)
	}
}

func TestAntiDetectionService_HeaderRandomization(t *testing.T) {
	config := &RiskAssessmentConfig{
		AntiDetectionConfig: AntiDetectionConfig{
			HeaderRandomization: true,
		},
	}
	logger := zap.NewNop()
	service := NewAntiDetectionService(config, logger)

	// Test header value randomization
	originalValue := "en-US,en;q=0.9"
	randomized := service.randomizeHeaderValue("accept-language", originalValue)

	assert.NotEmpty(t, randomized)
	// Should be one of the predefined values
	validValues := []string{
		"en-US,en;q=0.9",
		"en-US,en;q=0.8,es;q=0.7",
		"en-GB,en;q=0.9",
		"en-CA,en;q=0.9",
	}

	found := false
	for _, valid := range validValues {
		if randomized == valid {
			found = true
			break
		}
	}
	assert.True(t, found, "Randomized value should be one of the predefined values")
}

func TestAntiDetectionService_ProxySelection(t *testing.T) {
	config := &RiskAssessmentConfig{
		AntiDetectionConfig: AntiDetectionConfig{
			ProxyEnabled: true,
		},
	}
	logger := zap.NewNop()
	service := NewAntiDetectionService(config, logger)

	// Test proxy selection
	proxy1 := service.selectProxy()
	proxy2 := service.selectProxy()
	proxy3 := service.selectProxy()

	assert.NotNil(t, proxy1)
	assert.NotNil(t, proxy2)
	assert.NotNil(t, proxy3)

	// Should be different proxies (unless we've cycled through all of them)
	if len(service.proxies) > 2 {
		assert.NotEqual(t, proxy1.Host, proxy2.Host)
	}
}

func TestAntiDetectionService_RequestDelay(t *testing.T) {
	config := &RiskAssessmentConfig{
		AntiDetectionConfig: AntiDetectionConfig{
			RequestDelayEnabled: true,
			MinDelay:            10 * time.Millisecond,
			MaxDelay:            50 * time.Millisecond,
		},
	}
	logger := zap.NewNop()
	service := NewAntiDetectionService(config, logger)

	ctx := context.Background()

	// First request should not be delayed
	start := time.Now()
	err := service.applyRequestDelay(ctx, "https://example.com")
	duration := time.Since(start)

	require.NoError(t, err)
	assert.Less(t, duration, 10*time.Millisecond)

	// Second request should be delayed
	start = time.Now()
	err = service.applyRequestDelay(ctx, "https://example.com")
	duration = time.Since(start)

	require.NoError(t, err)
	assert.GreaterOrEqual(t, duration, 10*time.Millisecond)
}

func TestAntiDetectionService_DetectionRiskScore(t *testing.T) {
	config := &RiskAssessmentConfig{
		AntiDetectionConfig: AntiDetectionConfig{
			DetectionMonitoringEnabled: true,
			MaxDetectionEvents:         10,
		},
	}
	logger := zap.NewNop()
	service := NewAntiDetectionService(config, logger)

	// Test with no events
	score := service.calculateDetectionRiskScore()
	assert.Equal(t, 0.0, score)

	// Add some events
	service.eventMutex.Lock()
	service.detectionEvents = []DetectionEvent{
		{
			Timestamp: time.Now(),
			EventType: DetectionEventBlocked,
			Severity:  DetectionSeverityCritical,
		},
		{
			Timestamp: time.Now().Add(-1 * time.Hour),
			EventType: DetectionEventRateLimited,
			Severity:  DetectionSeverityMedium,
		},
	}
	service.eventMutex.Unlock()

	score = service.calculateDetectionRiskScore()
	assert.Greater(t, score, 0.0)
	assert.LessOrEqual(t, score, 1.0)
}

func TestAntiDetectionService_ExtractHeaders(t *testing.T) {
	config := &RiskAssessmentConfig{}
	logger := zap.NewNop()
	service := NewAntiDetectionService(config, logger)

	headers := http.Header{}
	headers.Set("Content-Type", "text/html")
	headers.Set("Server", "nginx")
	headers.Add("Set-Cookie", "session=123")
	headers.Add("Set-Cookie", "token=456")

	extracted := service.extractHeaders(headers)

	assert.Equal(t, "text/html", extracted["Content-Type"])
	assert.Equal(t, "nginx", extracted["Server"])
	assert.Equal(t, "session=123", extracted["Set-Cookie"]) // Should get first value
}

func TestAntiDetectionService_CleanupOldEvents(t *testing.T) {
	config := &RiskAssessmentConfig{
		AntiDetectionConfig: AntiDetectionConfig{
			DetectionMonitoringEnabled: true,
			MaxDetectionEvents:         10,
		},
	}
	logger := zap.NewNop()
	service := NewAntiDetectionService(config, logger)

	// Add old and new events
	service.eventMutex.Lock()
	service.detectionEvents = []DetectionEvent{
		{
			Timestamp: time.Now().Add(-8 * 24 * time.Hour), // 8 days old
			EventType: DetectionEventBlocked,
			Severity:  DetectionSeverityHigh,
		},
		{
			Timestamp: time.Now(), // Current
			EventType: DetectionEventRateLimited,
			Severity:  DetectionSeverityMedium,
		},
	}
	service.eventMutex.Unlock()

	// Run cleanup
	service.cleanupOldEvents()

	// Should only have the recent event
	service.eventMutex.RLock()
	eventCount := len(service.detectionEvents)
	service.eventMutex.RUnlock()

	assert.Equal(t, 1, eventCount)
}
