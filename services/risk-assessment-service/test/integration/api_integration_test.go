//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"kyb-platform/services/risk-assessment-service/internal/handlers"
	"kyb-platform/services/risk-assessment-service/internal/models"
)

// TestSetup holds test setup information
type TestSetup struct {
	server   *TestServer
	database *TestDatabase
	cache    *TestCache
}

// Note: SetupTestServer and TeardownTestServer are defined in test_helpers.go

// createTestRouter creates a test router with all endpoints
func createTestRouter(riskHandler *handlers.RiskAssessmentHandler) http.Handler {
	mux := http.NewServeMux()

	// Risk assessment endpoints
	mux.HandleFunc("POST /api/v1/assess", riskHandler.HandleRiskAssessment)
	mux.HandleFunc("GET /api/v1/assess/{id}", riskHandler.HandleGetRiskAssessment)
	mux.HandleFunc("POST /api/v1/assess/batch", riskHandler.HandleBatchRiskAssessment)
	mux.HandleFunc("GET /api/v1/assess/batch/{id}", riskHandler.HandleGetBatchAssessment)

	// Health check endpoint
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	return mux
}

func TestRiskAssessmentAPI_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Setup test server
	server := SetupTestServer(t)
	defer server.TeardownTestServer()

	tests := []struct {
		name             string
		method           string
		path             string
		requestBody      interface{}
		expectedStatus   int
		validateResponse func(*testing.T, *http.Response)
	}{
		{
			name:   "POST /api/v1/assess - valid request",
			method: "POST",
			path:   "/api/v1/assess",
			requestBody: models.RiskAssessmentRequest{
				BusinessName:      "Test Company Inc",
				BusinessAddress:   "123 Test Street, Test City, TC 12345",
				Industry:          "technology",
				Country:           "US",
				Phone:             "+1-555-123-4567",
				Email:             "test@company.com",
				Website:           "https://testcompany.com",
				PredictionHorizon: 3,
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, resp *http.Response) {
				var response models.RiskAssessmentResponse
				err := json.NewDecoder(resp.Body).Decode(&response)
				require.NoError(t, err)

				assert.NotEmpty(t, response.ID)
				assert.NotEmpty(t, response.BusinessID)
				assert.GreaterOrEqual(t, response.RiskScore, 0.0)
				assert.LessOrEqual(t, response.RiskScore, 1.0)
				assert.NotEmpty(t, response.RiskLevel)
				assert.NotEmpty(t, response.RiskFactors)
				assert.Equal(t, 3, response.PredictionHorizon)
				assert.GreaterOrEqual(t, response.ConfidenceScore, 0.0)
				assert.LessOrEqual(t, response.ConfidenceScore, 1.0)
				assert.Equal(t, models.StatusCompleted, response.Status)
				assert.NotZero(t, response.CreatedAt)
				assert.NotZero(t, response.UpdatedAt)
			},
		},
		{
			name:   "POST /api/v1/assess - invalid request",
			method: "POST",
			path:   "/api/v1/assess",
			requestBody: models.RiskAssessmentRequest{
				BusinessName: "", // Invalid: empty name
				Industry:     "technology",
				Country:      "US",
			},
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(t *testing.T, resp *http.Response) {
				// Should return error response
				assert.Contains(t, resp.Header.Get("Content-Type"), "application/json")
			},
		},
		{
			name:   "POST /api/v1/assess - missing required fields",
			method: "POST",
			path:   "/api/v1/assess",
			requestBody: map[string]interface{}{
				"business_name": "Test Company",
				// Missing required fields
			},
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(t *testing.T, resp *http.Response) {
				// Should return validation error
				assert.Contains(t, resp.Header.Get("Content-Type"), "application/json")
			},
		},
		{
			name:           "GET /api/v1/assess/{id} - valid ID",
			method:         "GET",
			path:           "/api/v1/assess/test-id",
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, resp *http.Response) {
				var response models.RiskAssessmentResponse
				err := json.NewDecoder(resp.Body).Decode(&response)
				require.NoError(t, err)

				assert.Equal(t, "test-id", response.ID)
			},
		},
		{
			name:           "GET /api/v1/assess/{id} - invalid ID",
			method:         "GET",
			path:           "/api/v1/assess/invalid-id",
			expectedStatus: http.StatusNotFound,
			validateResponse: func(t *testing.T, resp *http.Response) {
				// Should return not found error
			},
		},
		{
			name:   "POST /api/v1/assess/batch - valid batch request",
			method: "POST",
			path:   "/api/v1/assess/batch",
			requestBody: map[string]interface{}{
				"businesses": []models.RiskAssessmentRequest{
					{
						BusinessName:    "Company 1",
						BusinessAddress: "123 Street 1, City, TC 12345",
						Industry:        "technology",
						Country:         "US",
					},
					{
						BusinessName:    "Company 2",
						BusinessAddress: "456 Street 2, City, TC 12345",
						Industry:        "finance",
						Country:         "US",
					},
				},
			},
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, resp *http.Response) {
				var response map[string]interface{}
				err := json.NewDecoder(resp.Body).Decode(&response)
				require.NoError(t, err)

				assert.Contains(t, response, "batch_id")
				assert.Contains(t, response, "status")
			},
		},
		{
			name:           "GET /health - health check",
			method:         "GET",
			path:           "/health",
			expectedStatus: http.StatusOK,
			validateResponse: func(t *testing.T, resp *http.Response) {
				body := make([]byte, 100)
				n, err := resp.Body.Read(body)
				require.NoError(t, err)

				assert.Equal(t, "OK", string(body[:n]))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			var reqBody []byte
			var err error

			if tt.requestBody != nil {
				reqBody, err = json.Marshal(tt.requestBody)
				require.NoError(t, err)
			}

			req, err := http.NewRequest(tt.method, server.server.URL+tt.path, bytes.NewBuffer(reqBody))
			require.NoError(t, err)

			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json")

			// Make request
			client := &http.Client{Timeout: 30 * time.Second}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			// Validate response
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			if tt.validateResponse != nil {
				tt.validateResponse(t, resp)
			}
		})
	}
}

func TestRiskAssessmentAPI_ConcurrentRequests(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Setup test server
	server := SetupTestServer(t)
	defer server.TeardownTestServer()

	// Test concurrent requests
	numRequests := 10
	results := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		go func(id int) {
			request := models.RiskAssessmentRequest{
				BusinessName:      fmt.Sprintf("Test Company %d", id),
				BusinessAddress:   fmt.Sprintf("123 Test Street %d, Test City, TC 12345", id),
				Industry:          "technology",
				Country:           "US",
				PredictionHorizon: 3,
			}

			reqBody, err := json.Marshal(request)
			if err != nil {
				results <- err
				return
			}

			req, err := http.NewRequest("POST", server.server.URL+"/api/v1/assess", bytes.NewBuffer(reqBody))
			if err != nil {
				results <- err
				return
			}

			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{Timeout: 30 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				results <- err
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				results <- fmt.Errorf("unexpected status code: %d", resp.StatusCode)
				return
			}

			var response models.RiskAssessmentResponse
			err = json.NewDecoder(resp.Body).Decode(&response)
			if err != nil {
				results <- err
				return
			}

			if response.ID == "" {
				results <- fmt.Errorf("empty response ID")
				return
			}

			results <- nil
		}(i)
	}

	// Wait for all requests to complete
	for i := 0; i < numRequests; i++ {
		err := <-results
		assert.NoError(t, err, "Request %d failed", i)
	}
}

func TestRiskAssessmentAPI_ErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Setup test server
	server := SetupTestServer(t)
	defer server.TeardownTestServer()

	tests := []struct {
		name           string
		method         string
		path           string
		requestBody    string
		headers        map[string]string
		expectedStatus int
	}{
		{
			name:           "invalid JSON",
			method:         "POST",
			path:           "/api/v1/assess",
			requestBody:    `{"invalid": json}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unsupported content type",
			method:         "POST",
			path:           "/api/v1/assess",
			requestBody:    "plain text",
			headers:        map[string]string{"Content-Type": "text/plain"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "method not allowed",
			method:         "DELETE",
			path:           "/api/v1/assess",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "invalid endpoint",
			method:         "GET",
			path:           "/api/v1/invalid",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, server.server.URL+tt.path, bytes.NewBufferString(tt.requestBody))
			require.NoError(t, err)

			req.Header.Set("Content-Type", "application/json")
			if tt.headers != nil {
				for key, value := range tt.headers {
					req.Header.Set(key, value)
				}
			}

			client := &http.Client{Timeout: 30 * time.Second}
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func TestRiskAssessmentAPI_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Setup test server
	server := SetupTestServer(t)
	defer server.TeardownTestServer()

	// Test response time
	request := models.RiskAssessmentRequest{
		BusinessName:      "Performance Test Company",
		BusinessAddress:   "123 Performance Street, Test City, TC 12345",
		Industry:          "technology",
		Country:           "US",
		PredictionHorizon: 3,
	}

	reqBody, err := json.Marshal(request)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", server.server.URL+"/api/v1/assess", bytes.NewBuffer(reqBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}

	// Measure response time
	start := time.Now()
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	duration := time.Since(start)

	// Response should be fast (less than 2 seconds)
	assert.Less(t, duration, 2*time.Second, "Response time should be less than 2 seconds")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Validate response
	var response models.RiskAssessmentResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.NotEmpty(t, response.ID)
	assert.GreaterOrEqual(t, response.RiskScore, 0.0)
	assert.LessOrEqual(t, response.RiskScore, 1.0)
}

// Benchmark tests
func BenchmarkRiskAssessmentAPI_SingleRequest(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping integration benchmark")
	}

	// Setup test server
	server := SetupTestServer(&testing.T{})
	defer server.TeardownTestServer()

	request := models.RiskAssessmentRequest{
		BusinessName:      "Benchmark Company",
		BusinessAddress:   "123 Benchmark Street, Test City, TC 12345",
		Industry:          "technology",
		Country:           "US",
		PredictionHorizon: 3,
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		b.Fatal(err)
	}

	client := &http.Client{Timeout: 30 * time.Second}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, err := http.NewRequest("POST", server.server.URL+"/api/v1/assess", bytes.NewBuffer(reqBody))
		if err != nil {
			b.Fatal(err)
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			b.Fatal(err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			b.Fatalf("unexpected status code: %d", resp.StatusCode)
		}
	}
}

func BenchmarkRiskAssessmentAPI_ConcurrentRequests(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping integration benchmark")
	}

	// Setup test server
	server := SetupTestServer(&testing.T{})
	defer server.TeardownTestServer()

	request := models.RiskAssessmentRequest{
		BusinessName:      "Concurrent Benchmark Company",
		BusinessAddress:   "123 Concurrent Street, Test City, TC 12345",
		Industry:          "technology",
		Country:           "US",
		PredictionHorizon: 3,
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		b.Fatal(err)
	}

	client := &http.Client{Timeout: 30 * time.Second}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, err := http.NewRequest("POST", server.server.URL+"/api/v1/assess", bytes.NewBuffer(reqBody))
			if err != nil {
				b.Fatal(err)
			}

			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			if err != nil {
				b.Fatal(err)
			}
			resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				b.Fatalf("unexpected status code: %d", resp.StatusCode)
			}
		}
	})
}
