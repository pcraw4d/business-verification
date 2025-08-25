package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
	"github.com/pcraw4d/business-verification/internal/shared"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// MockIntelligentRouter is a mock implementation of the intelligent router
type MockIntelligentRouter struct {
	mock.Mock
}

func (m *MockIntelligentRouter) RouteRequest(ctx context.Context, request *shared.BusinessClassificationRequest) (*shared.BusinessClassificationResponse, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*shared.BusinessClassificationResponse), args.Error(1)
}

func (m *MockIntelligentRouter) RouteBatchRequest(ctx context.Context, requests []shared.BusinessClassificationRequest) ([]shared.BusinessClassificationResponse, []shared.BatchError) {
	args := m.Called(ctx, requests)
	return args.Get(0).([]shared.BusinessClassificationResponse), args.Get(1).([]shared.BatchError)
}

// MockMetrics is a mock implementation of metrics
type MockMetrics struct {
	mock.Mock
}

func (m *MockMetrics) IncCounter(name string, labels map[string]string) {
	m.Called(name, labels)
}

func (m *MockMetrics) RecordHistogram(name string, value float64, labels map[string]string) {
	m.Called(name, value, labels)
}

func (m *MockMetrics) SetGauge(name string, value float64, labels map[string]string) {
	m.Called(name, value, labels)
}

// MockTracer is a mock implementation of tracer
type MockTracer struct {
	mock.Mock
}

func (m *MockTracer) Start(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	args := m.Called(ctx, name, opts)
	return args.Get(0).(context.Context), args.Get(1).(trace.Span)
}

// MockSpan is a mock implementation of span
type MockSpan struct {
	mock.Mock
}

func (m *MockSpan) End(options ...trace.SpanEndOption) {
	m.Called(options)
}

func (m *MockSpan) AddEvent(name string, opts ...trace.SpanStartOption) {
	m.Called(name, opts)
}

func (m *MockSpan) IsRecording() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockSpan) RecordError(err error, opts ...trace.SpanStartOption) {
	m.Called(err, opts)
}

func (m *MockSpan) SpanContext() trace.SpanContext {
	args := m.Called()
	return args.Get(0).(trace.SpanContext)
}

func (m *MockSpan) SetStatus(code trace.StatusCode, description string) {
	m.Called(code, description)
}

func (m *MockSpan) SetName(name string) {
	m.Called(name)
}

func (m *MockSpan) SetAttributes(kv ...trace.SpanStartOption) {
	m.Called(kv)
}

func (m *MockSpan) TracerProvider() trace.TracerProvider {
	args := m.Called()
	return args.Get(0).(trace.TracerProvider)
}

// Test setup helper
func setupTestHandler() (*IntelligentRoutingHandler, *MockIntelligentRouter, *MockMetrics, *MockTracer) {
	mockRouter := &MockIntelligentRouter{}
	mockMetrics := &MockMetrics{}
	mockTracer := &MockTracer{}

	logger := observability.NewLogger(zap.NewNop())

	handler := &IntelligentRoutingHandler{
		router:       mockRouter,
		logger:       logger,
		metrics:      mockMetrics,
		tracer:       mockTracer,
		requestIDGen: func() string { return "test-request-id" },
	}

	return handler, mockRouter, mockMetrics, mockTracer
}

func TestIntelligentRoutingHandler_ClassifyBusiness_Success(t *testing.T) {
	handler, mockRouter, mockMetrics, mockTracer := setupTestHandler()

	// Setup mock expectations
	mockSpan := &MockSpan{}
	mockTracer.On("Start", mock.Anything, "IntelligentRoutingHandler.ClassifyBusiness").Return(context.Background(), mockSpan)
	mockSpan.On("End", mock.Anything).Return()

	expectedResponse := &shared.BusinessClassificationResponse{
		ID:           "test-response-id",
		BusinessName: "Test Business",
		Classifications: []shared.IndustryClassification{
			{
				IndustryCode:    "1234",
				IndustryName:    "Test Industry",
				ConfidenceScore: 0.95,
			},
		},
		CreatedAt: time.Now(),
	}

	mockRouter.On("RouteRequest", mock.Anything, mock.AnythingOfType("*shared.BusinessClassificationRequest")).Return(expectedResponse, nil)

	// Create test request
	requestBody := shared.BusinessClassificationRequest{
		BusinessName: "Test Business",
		WebsiteURL:   "https://testbusiness.com",
	}

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/classify", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	// Execute request
	handler.ClassifyBusiness(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response shared.BusinessClassificationResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, expectedResponse.ID, response.ID)
	assert.Equal(t, expectedResponse.BusinessName, response.BusinessName)
	assert.Len(t, response.Classifications, 1)

	mockRouter.AssertExpectations(t)
	mockTracer.AssertExpectations(t)
	mockSpan.AssertExpectations(t)
}

func TestIntelligentRoutingHandler_ClassifyBusiness_InvalidRequest(t *testing.T) {
	handler, _, _, mockTracer := setupTestHandler()

	// Setup mock expectations
	mockSpan := &MockSpan{}
	mockTracer.On("Start", mock.Anything, "IntelligentRoutingHandler.ClassifyBusiness").Return(context.Background(), mockSpan)
	mockSpan.On("End", mock.Anything).Return()

	// Create invalid request (missing business name)
	requestBody := map[string]interface{}{
		"website": "https://testbusiness.com",
	}

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/classify", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	// Execute request
	handler.ClassifyBusiness(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errorResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	require.NoError(t, err)

	assert.Contains(t, errorResponse["error"], "business_name is required")

	mockTracer.AssertExpectations(t)
	mockSpan.AssertExpectations(t)
}

func TestIntelligentRoutingHandler_ClassifyBusiness_RouterError(t *testing.T) {
	handler, mockRouter, mockMetrics, mockTracer := setupTestHandler()

	// Setup mock expectations
	mockSpan := &MockSpan{}
	mockTracer.On("Start", mock.Anything, "IntelligentRoutingHandler.ClassifyBusiness").Return(context.Background(), mockSpan)
	mockSpan.On("End", mock.Anything).Return()

	mockRouter.On("RouteRequest", mock.Anything, mock.AnythingOfType("*shared.BusinessClassificationRequest")).Return(nil, fmt.Errorf("router error"))

	mockMetrics.On("IncCounter", "classification_errors_total", mock.AnythingOfType("map[string]string")).Return()

	// Create test request
	requestBody := shared.BusinessClassificationRequest{
		BusinessName: "Test Business",
		WebsiteURL:   "https://testbusiness.com",
	}

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/classify", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	// Execute request
	handler.ClassifyBusiness(w, req)

	// Assertions
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var errorResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	require.NoError(t, err)

	assert.Contains(t, errorResponse["error"], "classification failed")

	mockRouter.AssertExpectations(t)
	mockTracer.AssertExpectations(t)
	mockSpan.AssertExpectations(t)
	mockMetrics.AssertExpectations(t)
}

func TestIntelligentRoutingHandler_ClassifyBusinessBatch_Success(t *testing.T) {
	handler, mockRouter, mockMetrics, mockTracer := setupTestHandler()

	// Setup mock expectations
	mockSpan := &MockSpan{}
	mockTracer.On("Start", mock.Anything, "IntelligentRoutingHandler.ClassifyBusinessBatch").Return(context.Background(), mockSpan)
	mockSpan.On("End", mock.Anything).Return()

	expectedResponses := []shared.BusinessClassificationResponse{
		{
			ID:           "response-1",
			BusinessName: "Business 1",
			Classifications: []shared.IndustryClassification{
				{
					IndustryCode:    "1234",
					IndustryName:    "Test Industry 1",
					ConfidenceScore: 0.95,
				},
			},
			CreatedAt: time.Now(),
		},
		{
			ID:           "response-2",
			BusinessName: "Business 2",
			Classifications: []shared.IndustryClassification{
				{
					IndustryCode:    "5678",
					IndustryName:    "Test Industry 2",
					ConfidenceScore: 0.90,
				},
			},
			CreatedAt: time.Now(),
		},
	}

	mockRouter.On("RouteBatchRequest", mock.Anything, mock.AnythingOfType("[]shared.BusinessClassificationRequest")).Return(expectedResponses, []shared.BatchError{})

	// Create test request
	requestBody := shared.BatchClassificationRequest{
		Requests: []shared.BusinessClassificationRequest{
			{
				BusinessName: "Business 1",
				Website:      "https://business1.com",
			},
			{
				BusinessName: "Business 2",
				Website:      "https://business2.com",
			},
		},
	}

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/classify/batch", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	// Execute request
	handler.ClassifyBusinessBatch(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response shared.BatchClassificationResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "test-request-id", response.ID)
	assert.Equal(t, "completed", response.Status)
	assert.Len(t, response.Responses, 2)
	assert.Len(t, response.Errors, 0)

	mockRouter.AssertExpectations(t)
	mockTracer.AssertExpectations(t)
	mockSpan.AssertExpectations(t)
}

func TestIntelligentRoutingHandler_ClassifyBusinessBatch_EmptyBatch(t *testing.T) {
	handler, _, _, mockTracer := setupTestHandler()

	// Setup mock expectations
	mockSpan := &MockSpan{}
	mockTracer.On("Start", mock.Anything, "IntelligentRoutingHandler.ClassifyBusinessBatch").Return(context.Background(), mockSpan)
	mockSpan.On("End", mock.Anything).Return()

	// Create empty batch request
	requestBody := shared.BatchClassificationRequest{
		Requests: []shared.BusinessClassificationRequest{},
	}

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/classify/batch", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	// Execute request
	handler.ClassifyBusinessBatch(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errorResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	require.NoError(t, err)

	assert.Contains(t, errorResponse["error"], "batch must contain at least one business")

	mockTracer.AssertExpectations(t)
	mockSpan.AssertExpectations(t)
}

func TestIntelligentRoutingHandler_ClassifyBusinessBatch_BatchSizeExceeded(t *testing.T) {
	handler, _, _, mockTracer := setupTestHandler()

	// Setup mock expectations
	mockSpan := &MockSpan{}
	mockTracer.On("Start", mock.Anything, "IntelligentRoutingHandler.ClassifyBusinessBatch").Return(context.Background(), mockSpan)
	mockSpan.On("End", mock.Anything).Return()

	// Create oversized batch request
	requests := make([]shared.BusinessClassificationRequest, 101)
	for i := 0; i < 101; i++ {
		requests[i] = shared.BusinessClassificationRequest{
			BusinessName: fmt.Sprintf("Business %d", i),
			Website:      fmt.Sprintf("https://business%d.com", i),
		}
	}

	requestBody := shared.BatchClassificationRequest{
		Requests: requests,
	}

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/classify/batch", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	// Execute request
	handler.ClassifyBusinessBatch(w, req)

	// Assertions
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errorResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	require.NoError(t, err)

	assert.Contains(t, errorResponse["error"], "batch size exceeds maximum of 100 businesses")

	mockTracer.AssertExpectations(t)
	mockSpan.AssertExpectations(t)
}

func TestIntelligentRoutingHandler_ClassifyBusinessBatch_PartialErrors(t *testing.T) {
	handler, mockRouter, _, mockTracer := setupTestHandler()

	// Setup mock expectations
	mockSpan := &MockSpan{}
	mockTracer.On("Start", mock.Anything, "IntelligentRoutingHandler.ClassifyBusinessBatch").Return(context.Background(), mockSpan)
	mockSpan.On("End", mock.Anything).Return()

	expectedResponses := []shared.BusinessClassificationResponse{
		{
			ID:           "response-1",
			BusinessName: "Business 1",
			Status:       "completed",
			Classifications: []shared.Classification{
				{
					Type:        "industry",
					Code:        "1234",
					Description: "Test Industry 1",
					Confidence:  0.95,
				},
			},
			CreatedAt: time.Now(),
		},
	}

	expectedErrors := []shared.BatchError{
		{
			Index: 1,
			Error: "classification failed for business 2",
		},
	}

	mockRouter.On("RouteBatchRequest", mock.Anything, mock.AnythingOfType("[]shared.BusinessClassificationRequest")).Return(expectedResponses, expectedErrors)

	// Create test request
	requestBody := shared.BatchClassificationRequest{
		Requests: []shared.BusinessClassificationRequest{
			{
				BusinessName: "Business 1",
				Website:      "https://business1.com",
			},
			{
				BusinessName: "Business 2",
				Website:      "https://business2.com",
			},
		},
	}

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/classify/batch", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	// Execute request
	handler.ClassifyBusinessBatch(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var response shared.BatchClassificationResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "test-request-id", response.ID)
	assert.Equal(t, "completed", response.Status)
	assert.Len(t, response.Responses, 1)
	assert.Len(t, response.Errors, 1)
	assert.Equal(t, 1, response.Errors[0].Index)
	assert.Contains(t, response.Errors[0].Error, "classification failed")

	mockRouter.AssertExpectations(t)
	mockTracer.AssertExpectations(t)
	mockSpan.AssertExpectations(t)
}

func TestIntelligentRoutingHandler_GetRoutingHealth(t *testing.T) {
	handler, _, _, mockTracer := setupTestHandler()

	// Setup mock expectations
	mockSpan := &MockSpan{}
	mockTracer.On("Start", mock.Anything, "IntelligentRoutingHandler.GetRoutingHealth").Return(context.Background(), mockSpan)
	mockSpan.On("End", mock.Anything).Return()

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Execute request
	handler.GetRoutingHealth(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var health map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &health)
	require.NoError(t, err)

	assert.Equal(t, "healthy", health["status"])
	assert.Equal(t, "intelligent_routing_system", health["router_id"])
	assert.Equal(t, "1.0.0", health["version"])

	mockTracer.AssertExpectations(t)
	mockSpan.AssertExpectations(t)
}

func TestIntelligentRoutingHandler_GetRoutingMetrics(t *testing.T) {
	handler, _, _, mockTracer := setupTestHandler()

	// Setup mock expectations
	mockSpan := &MockSpan{}
	mockTracer.On("Start", mock.Anything, "IntelligentRoutingHandler.GetRoutingMetrics").Return(context.Background(), mockSpan)
	mockSpan.On("End", mock.Anything).Return()

	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()

	// Execute request
	handler.GetRoutingMetrics(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var metrics map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &metrics)
	require.NoError(t, err)

	assert.Equal(t, float64(0), metrics["total_requests"])
	assert.Equal(t, float64(0), metrics["successful_requests"])
	assert.Equal(t, float64(0), metrics["failed_requests"])
	assert.Equal(t, float64(0), metrics["average_response_time"])
	assert.Equal(t, float64(0), metrics["requests_per_minute"])

	mockTracer.AssertExpectations(t)
	mockSpan.AssertExpectations(t)
}

func TestIntelligentRoutingHandler_parseClassificationRequest_Valid(t *testing.T) {
	handler, _, _, _ := setupTestHandler()

	// Create valid request
	requestBody := shared.BusinessClassificationRequest{
		BusinessName: "Test Business",
		Website:      "https://testbusiness.com",
	}

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/classify", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Parse request
	result, err := handler.parseClassificationRequest(req)

	// Assertions
	require.NoError(t, err)
	assert.Equal(t, "Test Business", result.BusinessName)
	assert.Equal(t, "https://testbusiness.com", result.Website)
}

func TestIntelligentRoutingHandler_parseClassificationRequest_InvalidJSON(t *testing.T) {
	handler, _, _, _ := setupTestHandler()

	// Create invalid JSON request
	req := httptest.NewRequest("POST", "/classify", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	// Parse request
	result, err := handler.parseClassificationRequest(req)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to decode request body")
}

func TestIntelligentRoutingHandler_parseClassificationRequest_MissingBusinessName(t *testing.T) {
	handler, _, _, _ := setupTestHandler()

	// Create request without business name
	requestBody := map[string]interface{}{
		"website": "https://testbusiness.com",
	}

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/classify", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Parse request
	result, err := handler.parseClassificationRequest(req)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "business_name is required")
}

func TestIntelligentRoutingHandler_parseBatchClassificationRequest_Valid(t *testing.T) {
	handler, _, _, _ := setupTestHandler()

	// Create valid batch request
	requestBody := shared.BatchClassificationRequest{
		Requests: []shared.BusinessClassificationRequest{
			{
				BusinessName: "Business 1",
				Website:      "https://business1.com",
			},
			{
				BusinessName: "Business 2",
				Website:      "https://business2.com",
			},
		},
	}

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/classify/batch", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Parse request
	result, err := handler.parseBatchClassificationRequest(req)

	// Assertions
	require.NoError(t, err)
	assert.Len(t, result.Requests, 2)
	assert.Equal(t, "Business 1", result.Requests[0].BusinessName)
	assert.Equal(t, "Business 2", result.Requests[1].BusinessName)
}

func TestIntelligentRoutingHandler_parseBatchClassificationRequest_InvalidJSON(t *testing.T) {
	handler, _, _, _ := setupTestHandler()

	// Create invalid JSON request
	req := httptest.NewRequest("POST", "/classify/batch", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	// Parse request
	result, err := handler.parseBatchClassificationRequest(req)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to decode batch request body")
}

func TestIntelligentRoutingHandler_parseBatchClassificationRequest_EmptyBatch(t *testing.T) {
	handler, _, _, _ := setupTestHandler()

	// Create empty batch request
	requestBody := shared.BatchClassificationRequest{
		Requests: []shared.BusinessClassificationRequest{},
	}

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/classify/batch", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Parse request
	result, err := handler.parseBatchClassificationRequest(req)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "batch must contain at least one business")
}

func TestIntelligentRoutingHandler_parseBatchClassificationRequest_MissingBusinessName(t *testing.T) {
	handler, _, _, _ := setupTestHandler()

	// Create batch request with missing business name
	requestBody := shared.BatchClassificationRequest{
		Requests: []shared.BusinessClassificationRequest{
			{
				Website: "https://business1.com",
			},
		},
	}

	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/classify/batch", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Parse request
	result, err := handler.parseBatchClassificationRequest(req)

	// Assertions
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "business_name is required for business at index 0")
}

func TestIntelligentRoutingHandler_writeResponse(t *testing.T) {
	handler, _, _, _ := setupTestHandler()

	w := httptest.NewRecorder()

	responseData := map[string]interface{}{
		"status":  "success",
		"message": "test message",
	}

	// Write response
	handler.writeResponse(w, responseData, http.StatusOK)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "test message", response["message"])
}

func TestIntelligentRoutingHandler_recordBatchMetrics(t *testing.T) {
	handler, _, mockMetrics, _ := setupTestHandler()

	// Setup mock expectations
	mockMetrics.On("RecordHistogram", "batch_processing_duration_seconds", mock.AnythingOfType("float64"), mock.AnythingOfType("map[string]string")).Return()
	mockMetrics.On("SetGauge", "batch_size", float64(5), mock.AnythingOfType("map[string]string")).Return()
	mockMetrics.On("SetGauge", "batch_success_count", float64(3), mock.AnythingOfType("map[string]string")).Return()
	mockMetrics.On("SetGauge", "batch_error_count", float64(2), mock.AnythingOfType("map[string]string")).Return()

	ctx := context.Background()
	requestID := "test-request-id"
	totalBusinesses := 5
	successfulCount := 3
	duration := 2 * time.Second

	// Record metrics
	handler.recordBatchMetrics(ctx, requestID, totalBusinesses, successfulCount, duration)

	// Assertions
	mockMetrics.AssertExpectations(t)
}

func TestIntelligentRoutingHandler_NewIntelligentRoutingHandler(t *testing.T) {
	mockRouter := &MockIntelligentRouter{}
	logger := observability.NewLogger(zap.NewNop())
	mockMetrics := &MockMetrics{}
	mockTracer := &MockTracer{}

	// Create handler
	handler := NewIntelligentRoutingHandler(mockRouter, logger, mockMetrics, mockTracer)

	// Assertions
	assert.NotNil(t, handler)
	assert.Equal(t, mockRouter, handler.router)
	assert.Equal(t, logger, handler.logger)
	assert.Equal(t, mockMetrics, handler.metrics)
	assert.Equal(t, mockTracer, handler.tracer)
	assert.NotNil(t, handler.requestIDGen)
}
