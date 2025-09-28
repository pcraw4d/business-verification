package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// E2ETestSuite provides comprehensive end-to-end testing for all KYB services
type E2ETestSuite struct {
	suite.Suite
	baseURL        string
	apiGateway     *APIGatewayClient
	classification *ClassificationClient
	merchant       *MerchantClient
	monitoring     *MonitoringClient
	pipeline       *PipelineClient
	frontend       *FrontendClient
	bi             *BusinessIntelligenceClient
	httpClient     *http.Client
}

// Test configuration
type TestConfig struct {
	BaseURL           string
	APIGatewayURL     string
	ClassificationURL string
	MerchantURL       string
	MonitoringURL     string
	PipelineURL       string
	FrontendURL       string
	BIURL             string
	Timeout           time.Duration
}

// Service clients for testing
type APIGatewayClient struct {
	baseURL string
	client  *http.Client
}

type ClassificationClient struct {
	baseURL string
	client  *http.Client
}

type MerchantClient struct {
	baseURL string
	client  *http.Client
}

type MonitoringClient struct {
	baseURL string
	client  *http.Client
}

type PipelineClient struct {
	baseURL string
	client  *http.Client
}

type FrontendClient struct {
	baseURL string
	client  *http.Client
}

type BusinessIntelligenceClient struct {
	baseURL string
	client  *http.Client
}

// Test data structures
type BusinessVerificationRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Address     string `json:"address"`
	Industry    string `json:"industry"`
	Website     string `json:"website,omitempty"`
	Phone       string `json:"phone,omitempty"`
	Email       string `json:"email,omitempty"`
}

type BusinessVerificationResponse struct {
	ID              string                 `json:"id"`
	Status          string                 `json:"status"`
	Score           float64                `json:"score"`
	Classifications []ClassificationResult `json:"classifications"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

type ClassificationResult struct {
	Code        string  `json:"code"`
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
}

type MerchantRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Address     string `json:"address"`
	Industry    string `json:"industry"`
	Website     string `json:"website,omitempty"`
	Phone       string `json:"phone,omitempty"`
	Email       string `json:"email,omitempty"`
}

type MerchantResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Address     string    `json:"address"`
	Industry    string    `json:"industry"`
	Website     string    `json:"website"`
	Phone       string    `json:"phone"`
	Email       string    `json:"email"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type HealthResponse struct {
	Status    string                 `json:"status"`
	Version   string                 `json:"version"`
	Timestamp time.Time              `json:"timestamp"`
	Services  map[string]string      `json:"services,omitempty"`
	Metrics   map[string]interface{} `json:"metrics,omitempty"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// SetupSuite initializes the test suite
func (suite *E2ETestSuite) SetupSuite() {
	// Load test configuration
	config := suite.loadTestConfig()

	// Initialize HTTP client
	suite.httpClient = &http.Client{
		Timeout: config.Timeout,
	}

	// Initialize service clients
	suite.apiGateway = &APIGatewayClient{
		baseURL: config.APIGatewayURL,
		client:  suite.httpClient,
	}

	suite.classification = &ClassificationClient{
		baseURL: config.ClassificationURL,
		client:  suite.httpClient,
	}

	suite.merchant = &MerchantClient{
		baseURL: config.MerchantURL,
		client:  suite.httpClient,
	}

	suite.monitoring = &MonitoringClient{
		baseURL: config.MonitoringURL,
		client:  suite.httpClient,
	}

	suite.pipeline = &PipelineClient{
		baseURL: config.PipelineURL,
		client:  suite.httpClient,
	}

	suite.frontend = &FrontendClient{
		baseURL: config.FrontendURL,
		client:  suite.httpClient,
	}

	suite.bi = &BusinessIntelligenceClient{
		baseURL: config.BIURL,
		client:  suite.httpClient,
	}

	suite.baseURL = config.BaseURL
}

// loadTestConfig loads test configuration from environment or defaults
func (suite *E2ETestSuite) loadTestConfig() *TestConfig {
	config := &TestConfig{
		BaseURL:           getEnv("TEST_BASE_URL", "http://localhost:8080"),
		APIGatewayURL:     getEnv("TEST_API_GATEWAY_URL", "http://localhost:8080"),
		ClassificationURL: getEnv("TEST_CLASSIFICATION_URL", "http://localhost:8081"),
		MerchantURL:       getEnv("TEST_MERCHANT_URL", "http://localhost:8082"),
		MonitoringURL:     getEnv("TEST_MONITORING_URL", "http://localhost:8083"),
		PipelineURL:       getEnv("TEST_PIPELINE_URL", "http://localhost:8084"),
		FrontendURL:       getEnv("TEST_FRONTEND_URL", "http://localhost:8085"),
		BIURL:             getEnv("TEST_BI_URL", "http://localhost:8087"),
		Timeout:           30 * time.Second,
	}

	return config
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// TestServiceHealth tests health endpoints for all services
func (suite *E2ETestSuite) TestServiceHealth() {
	services := map[string]string{
		"API Gateway":           suite.apiGateway.baseURL,
		"Classification":        suite.classification.baseURL,
		"Merchant":              suite.merchant.baseURL,
		"Monitoring":            suite.monitoring.baseURL,
		"Pipeline":              suite.pipeline.baseURL,
		"Frontend":              suite.frontend.baseURL,
		"Business Intelligence": suite.bi.baseURL,
	}

	for serviceName, baseURL := range services {
		suite.Run(fmt.Sprintf("Health_%s", serviceName), func() {
			healthResp, err := suite.makeRequest("GET", baseURL+"/health", nil)
			require.NoError(suite.T(), err, "Health check failed for %s", serviceName)
			assert.Equal(suite.T(), http.StatusOK, healthResp.StatusCode, "Health check returned non-200 status for %s", serviceName)

			var health HealthResponse
			err = json.Unmarshal(healthResp.Body, &health)
			require.NoError(suite.T(), err, "Failed to parse health response for %s", serviceName)
			assert.Equal(suite.T(), "healthy", health.Status, "Service %s is not healthy", serviceName)
		})
	}
}

// TestEndToEndBusinessVerification tests complete business verification flow
func (suite *E2ETestSuite) TestEndToEndBusinessVerification() {
	// Test data
	testBusiness := BusinessVerificationRequest{
		Name:        "Test Restaurant Corp",
		Description: "A fine dining restaurant specializing in Italian cuisine with full bar service",
		Address:     "123 Main Street, New York, NY 10001",
		Industry:    "Restaurant",
		Website:     "https://testrestaurant.com",
		Phone:       "+1-555-123-4567",
		Email:       "info@testrestaurant.com",
	}

	suite.Run("Complete_Business_Verification_Flow", func() {
		// Step 1: Submit business verification request
		verificationResp, err := suite.apiGateway.VerifyBusiness(testBusiness)
		require.NoError(suite.T(), err, "Business verification request failed")
		assert.Equal(suite.T(), http.StatusOK, verificationResp.StatusCode, "Business verification returned non-200 status")

		var verification BusinessVerificationResponse
		err = json.Unmarshal(verificationResp.Body, &verification)
		require.NoError(suite.T(), err, "Failed to parse verification response")
		assert.NotEmpty(suite.T(), verification.ID, "Verification ID should not be empty")
		assert.Equal(suite.T(), "completed", verification.Status, "Verification should be completed")
		assert.Greater(suite.T(), verification.Score, 0.0, "Verification score should be greater than 0")
		assert.NotEmpty(suite.T(), verification.Classifications, "Classifications should not be empty")

		// Step 2: Verify classification results
		for _, classification := range verification.Classifications {
			assert.NotEmpty(suite.T(), classification.Code, "Classification code should not be empty")
			assert.NotEmpty(suite.T(), classification.Type, "Classification type should not be empty")
			assert.NotEmpty(suite.T(), classification.Description, "Classification description should not be empty")
			assert.Greater(suite.T(), classification.Confidence, 0.0, "Classification confidence should be greater than 0")
			assert.LessOrEqual(suite.T(), classification.Confidence, 1.0, "Classification confidence should be less than or equal to 1")
		}

		// Step 3: Verify merchant creation
		merchantResp, err := suite.merchant.GetMerchant(verification.ID)
		require.NoError(suite.T(), err, "Failed to retrieve merchant")
		assert.Equal(suite.T(), http.StatusOK, merchantResp.StatusCode, "Merchant retrieval returned non-200 status")

		var merchant MerchantResponse
		err = json.Unmarshal(merchantResp.Body, &merchant)
		require.NoError(suite.T(), err, "Failed to parse merchant response")
		assert.Equal(suite.T(), verification.ID, merchant.ID, "Merchant ID should match verification ID")
		assert.Equal(suite.T(), testBusiness.Name, merchant.Name, "Merchant name should match request")

		// Step 4: Verify monitoring data
		monitoringResp, err := suite.monitoring.GetMetrics()
		require.NoError(suite.T(), err, "Failed to retrieve monitoring metrics")
		assert.Equal(suite.T(), http.StatusOK, monitoringResp.StatusCode, "Monitoring metrics returned non-200 status")

		// Step 5: Verify business intelligence data
		biResp, err := suite.bi.GetDashboard()
		require.NoError(suite.T(), err, "Failed to retrieve BI dashboard")
		assert.Equal(suite.T(), http.StatusOK, biResp.StatusCode, "BI dashboard returned non-200 status")
	})
}

// TestServiceIntegration tests service-to-service communication
func (suite *E2ETestSuite) TestServiceIntegration() {
	suite.Run("API_Gateway_to_Classification", func() {
		testBusiness := BusinessVerificationRequest{
			Name:        "Integration Test Business",
			Description: "A test business for integration testing",
			Address:     "456 Test Street, Test City, TC 12345",
			Industry:    "Technology",
		}

		// Test API Gateway routing to Classification Service
		verificationResp, err := suite.apiGateway.VerifyBusiness(testBusiness)
		require.NoError(suite.T(), err, "API Gateway to Classification integration failed")
		assert.Equal(suite.T(), http.StatusOK, verificationResp.StatusCode, "Integration test returned non-200 status")
	})

	suite.Run("Pipeline_Service_Processing", func() {
		// Test pipeline service processing
		pipelineResp, err := suite.pipeline.GetStatus()
		require.NoError(suite.T(), err, "Pipeline service status check failed")
		assert.Equal(suite.T(), http.StatusOK, pipelineResp.StatusCode, "Pipeline status returned non-200 status")
	})

	suite.Run("Monitoring_Service_Integration", func() {
		// Test monitoring service integration
		monitoringResp, err := suite.monitoring.GetHealth()
		require.NoError(suite.T(), err, "Monitoring service health check failed")
		assert.Equal(suite.T(), http.StatusOK, monitoringResp.StatusCode, "Monitoring health returned non-200 status")
	})
}

// TestPerformanceRequirements tests performance requirements
func (suite *E2ETestSuite) TestPerformanceRequirements() {
	suite.Run("Response_Time_Requirements", func() {
		testBusiness := BusinessVerificationRequest{
			Name:        "Performance Test Business",
			Description: "A test business for performance testing",
			Address:     "789 Performance Street, Test City, TC 12345",
			Industry:    "Technology",
		}

		start := time.Now()
		verificationResp, err := suite.apiGateway.VerifyBusiness(testBusiness)
		duration := time.Since(start)

		require.NoError(suite.T(), err, "Performance test failed")
		assert.Equal(suite.T(), http.StatusOK, verificationResp.StatusCode, "Performance test returned non-200 status")
		assert.Less(suite.T(), duration, 5*time.Second, "Response time should be less than 5 seconds")
		assert.Less(suite.T(), duration, 1*time.Second, "Response time should be less than 1 second for optimal performance")
	})

	suite.Run("Concurrent_Requests", func() {
		// Test concurrent requests
		concurrency := 10
		results := make(chan error, concurrency)

		for i := 0; i < concurrency; i++ {
			go func(index int) {
				testBusiness := BusinessVerificationRequest{
					Name:        fmt.Sprintf("Concurrent Test Business %d", index),
					Description: "A test business for concurrent testing",
					Address:     fmt.Sprintf("%d Concurrent Street, Test City, TC 12345", index),
					Industry:    "Technology",
				}

				_, err := suite.apiGateway.VerifyBusiness(testBusiness)
				results <- err
			}(i)
		}

		// Collect results
		for i := 0; i < concurrency; i++ {
			err := <-results
			assert.NoError(suite.T(), err, "Concurrent request %d failed", i)
		}
	})
}

// TestErrorHandling tests error handling scenarios
func (suite *E2ETestSuite) TestErrorHandling() {
	suite.Run("Invalid_Request_Data", func() {
		// Test with invalid request data
		invalidData := map[string]interface{}{
			"name": "", // Empty name should cause validation error
		}

		resp, err := suite.makeRequest("POST", suite.apiGateway.baseURL+"/verify", invalidData)
		require.NoError(suite.T(), err, "Request should not fail at HTTP level")
		assert.Equal(suite.T(), http.StatusBadRequest, resp.StatusCode, "Should return 400 for invalid data")

		var errorResp ErrorResponse
		err = json.Unmarshal(resp.Body, &errorResp)
		require.NoError(suite.T(), err, "Failed to parse error response")
		assert.NotEmpty(suite.T(), errorResp.Error, "Error message should not be empty")
	})

	suite.Run("Non_Existent_Resource", func() {
		// Test accessing non-existent resource
		resp, err := suite.makeRequest("GET", suite.apiGateway.baseURL+"/verify/non-existent-id", nil)
		require.NoError(suite.T(), err, "Request should not fail at HTTP level")
		assert.Equal(suite.T(), http.StatusNotFound, resp.StatusCode, "Should return 404 for non-existent resource")
	})
}

// TestDataConsistency tests data consistency across services
func (suite *E2ETestSuite) TestDataConsistency() {
	suite.Run("Cross_Service_Data_Consistency", func() {
		testBusiness := BusinessVerificationRequest{
			Name:        "Data Consistency Test Business",
			Description: "A test business for data consistency testing",
			Address:     "321 Consistency Street, Test City, TC 12345",
			Industry:    "Technology",
		}

		// Submit verification
		verificationResp, err := suite.apiGateway.VerifyBusiness(testBusiness)
		require.NoError(suite.T(), err, "Business verification failed")
		assert.Equal(suite.T(), http.StatusOK, verificationResp.StatusCode, "Verification returned non-200 status")

		var verification BusinessVerificationResponse
		err = json.Unmarshal(verificationResp.Body, &verification)
		require.NoError(suite.T(), err, "Failed to parse verification response")

		// Verify data consistency across services
		merchantResp, err := suite.merchant.GetMerchant(verification.ID)
		require.NoError(suite.T(), err, "Failed to retrieve merchant")
		assert.Equal(suite.T(), http.StatusOK, merchantResp.StatusCode, "Merchant retrieval returned non-200 status")

		var merchant MerchantResponse
		err = json.Unmarshal(merchantResp.Body, &merchant)
		require.NoError(suite.T(), err, "Failed to parse merchant response")

		// Verify data consistency
		assert.Equal(suite.T(), verification.ID, merchant.ID, "IDs should be consistent across services")
		assert.Equal(suite.T(), testBusiness.Name, merchant.Name, "Names should be consistent across services")
		assert.Equal(suite.T(), testBusiness.Address, merchant.Address, "Addresses should be consistent across services")
	})
}

// Helper methods for making HTTP requests
func (suite *E2ETestSuite) makeRequest(method, url string, body interface{}) (*HTTPResponse, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := suite.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Body:       respBody,
		Headers:    resp.Header,
	}, nil
}

// HTTPResponse represents an HTTP response
type HTTPResponse struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

// Service client methods
func (c *APIGatewayClient) VerifyBusiness(business BusinessVerificationRequest) (*HTTPResponse, error) {
	return c.makeRequest("POST", c.baseURL+"/verify", business)
}

func (c *MerchantClient) GetMerchant(id string) (*HTTPResponse, error) {
	return c.makeRequest("GET", c.baseURL+"/merchants/"+id, nil)
}

func (c *MonitoringClient) GetMetrics() (*HTTPResponse, error) {
	return c.makeRequest("GET", c.baseURL+"/metrics", nil)
}

func (c *MonitoringClient) GetHealth() (*HTTPResponse, error) {
	return c.makeRequest("GET", c.baseURL+"/health", nil)
}

func (c *PipelineClient) GetStatus() (*HTTPResponse, error) {
	return c.makeRequest("GET", c.baseURL+"/status", nil)
}

func (c *BusinessIntelligenceClient) GetDashboard() (*HTTPResponse, error) {
	return c.makeRequest("GET", c.baseURL+"/dashboard/executive", nil)
}

// Generic makeRequest method for service clients
func (c *APIGatewayClient) makeRequest(method, url string, body interface{}) (*HTTPResponse, error) {
	return makeHTTPRequest(c.client, method, url, body)
}

func (c *MerchantClient) makeRequest(method, url string, body interface{}) (*HTTPResponse, error) {
	return makeHTTPRequest(c.client, method, url, body)
}

func (c *MonitoringClient) makeRequest(method, url string, body interface{}) (*HTTPResponse, error) {
	return makeHTTPRequest(c.client, method, url, body)
}

func (c *PipelineClient) makeRequest(method, url string, body interface{}) (*HTTPResponse, error) {
	return makeHTTPRequest(c.client, method, url, body)
}

func (c *BusinessIntelligenceClient) makeRequest(method, url string, body interface{}) (*HTTPResponse, error) {
	return makeHTTPRequest(c.client, method, url, body)
}

// makeHTTPRequest is a generic HTTP request helper
func makeHTTPRequest(client *http.Client, method, url string, body interface{}) (*HTTPResponse, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Body:       respBody,
		Headers:    resp.Header,
	}, nil
}

// TestE2ESuite runs the complete E2E test suite
func TestE2ESuite(t *testing.T) {
	suite.Run(t, new(E2ETestSuite))
}
