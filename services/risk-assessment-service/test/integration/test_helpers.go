//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/cache"
	"kyb-platform/services/risk-assessment-service/internal/config"
	"kyb-platform/services/risk-assessment-service/internal/external"
	"kyb-platform/services/risk-assessment-service/internal/handlers"
	"kyb-platform/services/risk-assessment-service/internal/ml"
	"kyb-platform/services/risk-assessment-service/internal/models"
	"kyb-platform/services/risk-assessment-service/internal/supabase"
)

// TestServer represents a test HTTP server
type TestServer struct {
	server   *httptest.Server
	handlers *handlers.RiskAssessmentHandler
	logger   *zap.Logger
}

// TestDatabase represents a test database
type TestDatabase struct {
	client *supabase.Client
	logger *zap.Logger
}

// TestCache represents a test cache
type TestCache struct {
	cache  cache.Cache
	logger *zap.Logger
}

// SetupTestServer creates a test HTTP server
func SetupTestServer(t *testing.T) *TestServer {
	logger := zap.NewNop()
	cfg, err := config.Load()
	require.NoError(t, err)

	// Override config for testing
	cfg.Server.Port = "0" // Let the system choose a free port
	cfg.Supabase.URL = "https://test.supabase.co"
	cfg.Supabase.APIKey = "test-api-key"
	cfg.Redis.Addrs = []string{"localhost:6379"}
	cfg.Redis.DB = 1
	cfg.Redis.KeyPrefix = "test:"

	// Create handlers
	handlers := handlers.NewRiskAssessmentHandler(cfg, logger)

	// Create test server
	server := httptest.NewServer(handlers.SetupRoutes())

	return &TestServer{
		server:   server,
		handlers: handlers,
		logger:   logger,
	}
}

// TeardownTestServer closes the test server
func (ts *TestServer) TeardownTestServer() {
	if ts.server != nil {
		ts.server.Close()
	}
}

// URL returns the test server URL
func (ts *TestServer) URL() string {
	return ts.server.URL
}

// MakeRequest makes an HTTP request to the test server
func (ts *TestServer) MakeRequest(t *testing.T, method, path string, body interface{}) *http.Response {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		require.NoError(t, err)
	}

	req, err := http.NewRequest(method, ts.URL()+path, bytes.NewBuffer(reqBody))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err)

	return resp
}

// MockThomsonReutersServer creates a mock Thomson Reuters server
func MockThomsonReutersServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/api/v1/verify":
			response := map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"verified": true,
					"score":    0.95,
					"details": map[string]interface{}{
						"company_name": "Test Company Inc",
						"address":      "123 Test Street, Test City, TC 12345",
						"industry":     "technology",
					},
				},
			}
			json.NewEncoder(w).Encode(response)
		default:
			http.NotFound(w, r)
		}
	}))
}

// MockOFACServer creates a mock OFAC server
func MockOFACServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/api/v1/sanctions":
			response := map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"sanctions_found": false,
					"matches":         []interface{}{},
				},
			}
			json.NewEncoder(w).Encode(response)
		default:
			http.NotFound(w, r)
		}
	}))
}

// MockNewsAPIServer creates a mock News API server
func MockNewsAPIServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/v2/everything":
			response := map[string]interface{}{
				"status":       "ok",
				"totalResults": 0,
				"articles":     []interface{}{},
			}
			json.NewEncoder(w).Encode(response)
		default:
			http.NotFound(w, r)
		}
	}))
}

// MockOpenCorporatesServer creates a mock OpenCorporates server
func MockOpenCorporatesServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/api/0.4/companies/search":
			response := map[string]interface{}{
				"results": map[string]interface{}{
					"companies": []interface{}{
						map[string]interface{}{
							"company": map[string]interface{}{
								"name":              "Test Company Inc",
								"company_number":    "12345678",
								"jurisdiction_code": "us",
								"company_type":      "Corporation",
								"status":            "Active",
							},
						},
					},
				},
			}
			json.NewEncoder(w).Encode(response)
		default:
			http.NotFound(w, r)
		}
	}))
}

// CreateTestRiskAssessment creates a test risk assessment
func CreateTestRiskAssessment() *models.RiskAssessment {
	return &models.RiskAssessment{
		ID:         "test-assessment-123",
		BusinessID: "test-business-456",
		Status:     models.AssessmentStatusCompleted,
		RiskLevel:  models.RiskLevelMedium,
		RiskScore:  0.65,
		RiskFactors: []models.RiskFactor{
			{
				Category:    models.RiskCategoryFinancial,
				Subcategory: "credit_score",
				Score:       0.7,
				Description: "Credit score analysis",
				Confidence:  0.85,
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// CreateTestRiskAssessmentRequest creates a test risk assessment request
func CreateTestRiskAssessmentRequest() *models.RiskAssessmentRequest {
	return &models.RiskAssessmentRequest{
		BusinessName:      "Test Company Inc",
		BusinessAddress:   "123 Test Street, Test City, TC 12345",
		Industry:          "technology",
		Country:           "US",
		PredictionHorizon: 3,
		ModelType:         "xgboost",
	}
}

// CreateTestRiskAssessmentResponse creates a test risk assessment response
func CreateTestRiskAssessmentResponse() *models.RiskAssessmentResponse {
	return &models.RiskAssessmentResponse{
		ID:        "test-assessment-123",
		Status:    string(models.AssessmentStatusCompleted),
		RiskLevel: string(models.RiskLevelMedium),
		RiskScore: 0.65,
		RiskFactors: []models.RiskFactor{
			{
				Category:    models.RiskCategoryFinancial,
				Subcategory: "credit_score",
				Score:       0.7,
				Description: "Credit score analysis",
				Confidence:  0.85,
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// AssertRiskAssessmentResponse validates a risk assessment response
func AssertRiskAssessmentResponse(t *testing.T, resp *models.RiskAssessmentResponse) {
	assert.NotEmpty(t, resp.ID)
	assert.NotEmpty(t, resp.Status)
	assert.NotEmpty(t, resp.RiskLevel)
	assert.GreaterOrEqual(t, resp.RiskScore, 0.0)
	assert.LessOrEqual(t, resp.RiskScore, 1.0)
	assert.NotNil(t, resp.RiskFactors)
	assert.NotZero(t, resp.CreatedAt)
	assert.NotZero(t, resp.UpdatedAt)
}

// AssertErrorResponse validates an error response
func AssertErrorResponse(t *testing.T, resp *http.Response, expectedStatus int) {
	assert.Equal(t, expectedStatus, resp.StatusCode)

	var errorResp map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&errorResp)
	require.NoError(t, err)

	assert.Contains(t, errorResp, "error")
	assert.Contains(t, errorResp, "message")
}

// CleanupTestData cleans up test data
func CleanupTestData(t *testing.T, cache cache.Cache, db *supabase.Client) {
	ctx := context.Background()

	// Clean up cache
	if cache != nil {
		cache.Clear(ctx)
	}

	// Clean up database
	if db != nil {
		// Add database cleanup logic here
		// This would depend on the specific database implementation
	}
}

// WaitForService waits for a service to be ready
func WaitForService(t *testing.T, url string, timeout time.Duration) {
	client := &http.Client{Timeout: 5 * time.Second}
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		resp, err := client.Get(url + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(100 * time.Millisecond)
	}

	t.Fatalf("Service at %s not ready after %v", url, timeout)
}

// GenerateTestData generates test data for various scenarios
func GenerateTestData() map[string]interface{} {
	return map[string]interface{}{
		"valid_business": CreateTestRiskAssessmentRequest(),
		"invalid_business": &models.RiskAssessmentRequest{
			BusinessName:    "", // Invalid: empty name
			BusinessAddress: "123 Test Street",
			Industry:        "technology",
			Country:         "US",
		},
		"high_risk_business": &models.RiskAssessmentRequest{
			BusinessName:      "High Risk Company",
			BusinessAddress:   "456 High Risk Street, Risk City, RC 54321",
			Industry:          "financial_services",
			Country:           "US",
			PredictionHorizon: 6,
		},
		"low_risk_business": &models.RiskAssessmentRequest{
			BusinessName:      "Low Risk Company",
			BusinessAddress:   "789 Low Risk Street, Safe City, SC 98765",
			Industry:          "technology",
			Country:           "US",
			PredictionHorizon: 3,
		},
	}
}

// BenchmarkTestData generates data for benchmark tests
func BenchmarkTestData() []*models.RiskAssessmentRequest {
	var requests []*models.RiskAssessmentRequest

	for i := 0; i < 1000; i++ {
		requests = append(requests, &models.RiskAssessmentRequest{
			BusinessName:      fmt.Sprintf("Test Company %d", i),
			BusinessAddress:   fmt.Sprintf("%d Test Street, Test City, TC %05d", i, i),
			Industry:          "technology",
			Country:           "US",
			PredictionHorizon: 3,
			ModelType:         "xgboost",
		})
	}

	return requests
}

// MockExternalAPIManager creates a mock external API manager
func MockExternalAPIManager(t *testing.T, mockServers map[string]*httptest.Server) *external.ExternalAPIManager {
	logger := zap.NewNop()
	cfg := &config.ExternalConfig{
		ThomsonReuters: config.ThomsonReutersConfig{
			APIKey:  "test-api-key",
			BaseURL: mockServers["thomson_reuters"].URL,
			Timeout: 30 * time.Second,
		},
		OFAC: config.OFACConfig{
			APIKey:  "test-api-key",
			BaseURL: mockServers["ofac"].URL,
			Timeout: 30 * time.Second,
		},
		NewsAPI: config.NewsAPIConfig{
			APIKey:  "test-api-key",
			BaseURL: mockServers["newsapi"].URL,
			Timeout: 30 * time.Second,
		},
		OpenCorporates: config.OpenCorporatesConfig{
			APIKey:  "test-api-key",
			BaseURL: mockServers["opencorporates"].URL,
			Timeout: 30 * time.Second,
		},
	}

	manager, err := external.NewExternalAPIManager(cfg, logger)
	require.NoError(t, err)

	return manager
}

// MockMLService creates a mock ML service
func MockMLService(t *testing.T) *ml.Service {
	logger := zap.NewNop()
	cfg := &config.MLConfig{
		ModelPath:    "/tmp/test-models",
		ModelType:    "xgboost",
		BatchSize:    100,
		MaxWorkers:   4,
		CacheEnabled: true,
		CacheTTL:     300 * time.Second,
	}

	service, err := ml.NewService(cfg, logger)
	require.NoError(t, err)

	return service
}

// SetupTestEnvironment sets up the test environment
func SetupTestEnvironment(t *testing.T) {
	// Set test environment variables
	os.Setenv("ENV", "test")
	os.Setenv("LOG_LEVEL", "error")
	os.Setenv("SUPABASE_URL", "https://test.supabase.co")
	os.Setenv("SUPABASE_API_KEY", "test-api-key")
	os.Setenv("REDIS_ADDRS", "localhost:6379")
	os.Setenv("REDIS_DB", "1")
	os.Setenv("REDIS_KEY_PREFIX", "test:")
}

// TeardownTestEnvironment cleans up the test environment
func TeardownTestEnvironment(t *testing.T) {
	// Clean up environment variables
	os.Unsetenv("ENV")
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("SUPABASE_URL")
	os.Unsetenv("SUPABASE_API_KEY")
	os.Unsetenv("REDIS_ADDRS")
	os.Unsetenv("REDIS_DB")
	os.Unsetenv("REDIS_KEY_PREFIX")
}

// TestMain runs before all tests
func TestMain(m *testing.M) {
	// Setup test environment
	SetupTestEnvironment(&testing.T{})

	// Run tests
	code := m.Run()

	// Cleanup test environment
	TeardownTestEnvironment(&testing.T{})

	// Exit with the same code as the tests
	os.Exit(code)
}
