package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/api/handlers"
	"github.com/pcraw4d/business-verification/internal/auth"
	"github.com/pcraw4d/business-verification/internal/classification"
	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/database"
	"github.com/pcraw4d/business-verification/internal/observability"
	"github.com/pcraw4d/business-verification/internal/risk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// APITestSuite provides integration testing for API endpoints
type APITestSuite struct {
	server  *httptest.Server
	db      database.Database
	logger  *observability.Logger
	cleanup func()
}

// NewAPITestSuite creates a new API test suite
func NewAPITestSuite(t *testing.T) *APITestSuite {
	// Setup test database
	db, cleanup := setupTestDatabase(t)

	// Setup logger
	logger := observability.NewLogger(&config.ObservabilityConfig{
		LogLevel:  "debug",
		LogFormat: "json",
	})

	// Setup services
	authService := auth.NewAuthService(&config.AuthConfig{}, db, logger, nil)
	classificationService := classification.NewClassificationService(nil, db, logger, nil)
	riskService := risk.NewService(nil, logger)

	// Setup handlers
	authHandler := handlers.NewAuthHandler(authService, logger)
	classificationHandler := handlers.NewClassificationHandler(classificationService, logger)
	riskHandler := handlers.NewRiskHandler(riskService, logger)

	// Setup router
	mux := http.NewServeMux()

	// Auth routes
	mux.HandleFunc("POST /v1/auth/register", authHandler.Register)
	mux.HandleFunc("POST /v1/auth/login", authHandler.Login)
	mux.HandleFunc("POST /v1/auth/logout", authHandler.Logout)
	mux.HandleFunc("POST /v1/auth/refresh", authHandler.RefreshToken)

	// Classification routes
	mux.HandleFunc("POST /v1/classify", classificationHandler.ClassifyBusiness)
	mux.HandleFunc("POST /v1/classify/batch", classificationHandler.ClassifyBusinesses)
	mux.HandleFunc("GET /v1/classify/{business_id}", classificationHandler.GetClassification)

	// Risk routes
	mux.HandleFunc("POST /v1/risk/assess", riskHandler.AssessRisk)
	mux.HandleFunc("GET /v1/risk/{business_id}", riskHandler.GetRiskAssessment)
	mux.HandleFunc("GET /v1/risk/history/{business_id}", riskHandler.GetRiskHistory)

	// Create test server
	server := httptest.NewServer(mux)

	return &APITestSuite{
		server:  server,
		db:      db,
		logger:  logger,
		cleanup: cleanup,
	}
}

// setupTestDatabase sets up a test database
func setupTestDatabase(t *testing.T) (database.Database, func()) {
	// This would connect to a test database
	// For now, we'll return a mock
	return nil, func() {}
}

// TestAuthEndpoints tests authentication endpoints
func TestAuthEndpoints(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	suite := NewAPITestSuite(t)
	defer suite.cleanup()

	t.Run("Register User", func(t *testing.T) {
		payload := map[string]interface{}{
			"email":      "test@example.com",
			"password":   "securepassword123",
			"first_name": "John",
			"last_name":  "Doe",
		}

		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", suite.server.URL+"/v1/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Contains(t, response, "user_id")
		assert.Contains(t, response, "message")
	})

	t.Run("Login User", func(t *testing.T) {
		payload := map[string]interface{}{
			"email":    "test@example.com",
			"password": "securepassword123",
		}

		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", suite.server.URL+"/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Contains(t, response, "access_token")
		assert.Contains(t, response, "refresh_token")
		assert.Contains(t, response, "expires_in")
	})

	t.Run("Invalid Login", func(t *testing.T) {
		payload := map[string]interface{}{
			"email":    "test@example.com",
			"password": "wrongpassword",
		}

		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", suite.server.URL+"/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

// TestClassificationEndpoints tests classification endpoints
func TestClassificationEndpoints(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	suite := NewAPITestSuite(t)
	defer suite.cleanup()

	// Get auth token first
	token := getAuthToken(t, suite)

	t.Run("Classify Single Business", func(t *testing.T) {
		payload := map[string]interface{}{
			"business_name": "Acme Corporation",
			"business_type": "Corporation",
			"industry":      "Technology",
			"location": map[string]interface{}{
				"country": "US",
				"state":   "CA",
				"city":    "San Francisco",
			},
		}

		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", suite.server.URL+"/v1/classify", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Contains(t, response, "classification_id")
		assert.Contains(t, response, "naics_code")
		assert.Contains(t, response, "confidence_score")
		assert.Contains(t, response, "business_type")
	})

	t.Run("Classify Multiple Businesses", func(t *testing.T) {
		payload := map[string]interface{}{
			"businesses": []map[string]interface{}{
				{
					"business_name": "Tech Solutions Inc",
					"business_type": "Corporation",
					"industry":      "Technology",
				},
				{
					"business_name": "Green Energy LLC",
					"business_type": "LLC",
					"industry":      "Energy",
				},
			},
		}

		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", suite.server.URL+"/v1/classify/batch", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Contains(t, response, "classifications")
		assert.Contains(t, response, "total_processed")
		assert.Contains(t, response, "successful")
		assert.Contains(t, response, "failed")
	})

	t.Run("Get Classification", func(t *testing.T) {
		// First create a classification
		classificationID := createTestClassification(t, suite, token)

		req, _ := http.NewRequest("GET", suite.server.URL+"/v1/classify/"+classificationID, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Contains(t, response, "classification_id")
		assert.Contains(t, response, "business_name")
		assert.Contains(t, response, "naics_code")
	})
}

// TestRiskEndpoints tests risk assessment endpoints
func TestRiskEndpoints(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	suite := NewAPITestSuite(t)
	defer suite.cleanup()

	token := getAuthToken(t, suite)

	t.Run("Assess Risk", func(t *testing.T) {
		payload := map[string]interface{}{
			"business_id":   "test-business-123",
			"business_name": "High Risk Corp",
			"business_type": "Corporation",
			"industry":      "Financial Services",
			"location": map[string]interface{}{
				"country": "US",
				"state":   "NY",
				"city":    "New York",
			},
			"financial_data": map[string]interface{}{
				"annual_revenue":    1000000,
				"employee_count":    50,
				"years_in_business": 2,
			},
		}

		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", suite.server.URL+"/v1/risk/assess", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Contains(t, response, "risk_assessment_id")
		assert.Contains(t, response, "risk_score")
		assert.Contains(t, response, "risk_level")
		assert.Contains(t, response, "risk_factors")
		assert.Contains(t, response, "recommendations")
	})

	t.Run("Get Risk Assessment", func(t *testing.T) {
		// First create a risk assessment
		assessmentID := createTestRiskAssessment(t, suite, token)

		req, _ := http.NewRequest("GET", suite.server.URL+"/v1/risk/"+assessmentID, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Contains(t, response, "risk_assessment_id")
		assert.Contains(t, response, "risk_score")
		assert.Contains(t, response, "risk_level")
	})

	t.Run("Get Risk History", func(t *testing.T) {
		businessID := "test-business-123"

		req, _ := http.NewRequest("GET", suite.server.URL+"/v1/risk/history/"+businessID, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Contains(t, response, "assessments")
		assert.Contains(t, response, "total_count")
	})
}

// TestAPIErrorHandling tests error handling in API endpoints
func TestAPIErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	suite := NewAPITestSuite(t)
	defer suite.cleanup()

	t.Run("Invalid JSON", func(t *testing.T) {
		req, _ := http.NewRequest("POST", suite.server.URL+"/v1/classify", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Missing Required Fields", func(t *testing.T) {
		payload := map[string]interface{}{
			"business_type": "Corporation",
			// Missing business_name
		}

		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", suite.server.URL+"/v1/classify", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("Unauthorized Access", func(t *testing.T) {
		req, _ := http.NewRequest("GET", suite.server.URL+"/v1/classify/test-id", nil)
		// No Authorization header

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Invalid Token", func(t *testing.T) {
		req, _ := http.NewRequest("GET", suite.server.URL+"/v1/classify/test-id", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

// TestAPIPerformance tests API performance under load
func TestAPIPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	suite := NewAPITestSuite(t)
	defer suite.cleanup()

	token := getAuthToken(t, suite)

	t.Run("Concurrent Classification Requests", func(t *testing.T) {
		const numRequests = 10
		results := make(chan int, numRequests)

		// Start concurrent requests
		for i := 0; i < numRequests; i++ {
			go func(id int) {
				payload := map[string]interface{}{
					"business_name": fmt.Sprintf("Test Business %d", id),
					"business_type": "Corporation",
					"industry":      "Technology",
				}

				body, _ := json.Marshal(payload)
				req, _ := http.NewRequest("POST", suite.server.URL+"/v1/classify", bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+token)

				start := time.Now()
				resp, err := http.DefaultClient.Do(req)
				duration := time.Since(start)

				if err == nil && resp.StatusCode == http.StatusOK {
					results <- int(duration.Milliseconds())
				} else {
					results <- -1
				}
				resp.Body.Close()
			}(i)
		}

		// Collect results
		var durations []int
		for i := 0; i < numRequests; i++ {
			duration := <-results
			if duration > 0 {
				durations = append(durations, duration)
			}
		}

		// Calculate statistics
		if len(durations) > 0 {
			var total int
			for _, d := range durations {
				total += d
			}
			avgDuration := total / len(durations)

			t.Logf("Average response time: %dms", avgDuration)
			t.Logf("Successful requests: %d/%d", len(durations), numRequests)

			// Assert reasonable performance
			assert.Less(t, avgDuration, 1000, "Average response time should be under 1 second")
			assert.Greater(t, len(durations), numRequests/2, "At least half of requests should succeed")
		}
	})
}

// Helper functions

func getAuthToken(t *testing.T, suite *APITestSuite) string {
	// Register a test user
	registerPayload := map[string]interface{}{
		"email":      "test@example.com",
		"password":   "securepassword123",
		"first_name": "John",
		"last_name":  "Doe",
	}

	body, _ := json.Marshal(registerPayload)
	req, _ := http.NewRequest("POST", suite.server.URL+"/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()

	// Login to get token
	loginPayload := map[string]interface{}{
		"email":    "test@example.com",
		"password": "securepassword123",
	}

	body, _ = json.Marshal(loginPayload)
	req, _ = http.NewRequest("POST", suite.server.URL+"/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	token, ok := response["access_token"].(string)
	require.True(t, ok)
	return token
}

func createTestClassification(t *testing.T, suite *APITestSuite, token string) string {
	payload := map[string]interface{}{
		"business_name": "Test Business",
		"business_type": "Corporation",
		"industry":      "Technology",
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", suite.server.URL+"/v1/classify", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	classificationID, ok := response["classification_id"].(string)
	require.True(t, ok)
	return classificationID
}

func createTestRiskAssessment(t *testing.T, suite *APITestSuite, token string) string {
	payload := map[string]interface{}{
		"business_id":   "test-business-123",
		"business_name": "Test Business",
		"business_type": "Corporation",
		"industry":      "Technology",
	}

	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", suite.server.URL+"/v1/risk/assess", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assessmentID, ok := response["risk_assessment_id"].(string)
	require.True(t, ok)
	return assessmentID
}
