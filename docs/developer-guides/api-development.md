# KYB Platform API Development Guide

## Overview

This guide provides comprehensive documentation for developing and maintaining the KYB Platform API. The API follows RESTful principles and provides merchant-centric business verification functionality.

## API Architecture

### Design Principles

1. **RESTful Design**: Use standard HTTP methods and status codes
2. **Resource-Based URLs**: URLs represent resources, not actions
3. **JSON Format**: All requests and responses use JSON
4. **Stateless**: Each request contains all necessary information
5. **Versioned**: API versions are clearly specified in URLs

### Base URL Structure

```
https://api.kyb-platform.com/v1/
```

### Authentication

The API uses JWT (JSON Web Token) authentication:

```http
Authorization: Bearer <jwt-token>
```

## API Endpoints

### Merchant Management

#### List Merchants

```http
GET /api/v1/merchants
```

**Query Parameters**:
- `page` (optional): Page number (default: 1)
- `limit` (optional): Items per page (default: 20, max: 100)
- `portfolio_type` (optional): Filter by portfolio type
- `risk_level` (optional): Filter by risk level
- `search` (optional): Search term for merchant name

**Response**:
```json
{
  "data": [
    {
      "id": "merchant_123",
      "name": "Acme Corporation",
      "business_type": "Retail",
      "industry_code": "5411",
      "portfolio_type": "onboarded",
      "risk_level": "low",
      "created_at": "2025-01-19T10:30:00Z",
      "updated_at": "2025-01-19T10:30:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 150,
    "total_pages": 8
  }
}
```

#### Get Merchant Details

```http
GET /api/v1/merchants/{id}
```

**Response**:
```json
{
  "id": "merchant_123",
  "name": "Acme Corporation",
  "business_type": "Retail",
  "industry_code": "5411",
  "portfolio_type": "onboarded",
  "risk_level": "low",
  "contact_info": {
    "email": "contact@acme.com",
    "phone": "+1-555-123-4567",
    "website": "https://www.acme.com"
  },
  "address": {
    "street": "123 Main St",
    "city": "Anytown",
    "state": "CA",
    "zip": "12345",
    "country": "US"
  },
  "verification_status": {
    "identity_verified": true,
    "address_verified": true,
    "business_verified": true,
    "last_verified": "2025-01-15T14:30:00Z"
  },
  "created_at": "2025-01-19T10:30:00Z",
  "updated_at": "2025-01-19T10:30:00Z"
}
```

#### Create Merchant

```http
POST /api/v1/merchants
```

**Request Body**:
```json
{
  "name": "New Business Inc",
  "business_type": "Technology",
  "industry_code": "7372",
  "contact_info": {
    "email": "info@newbusiness.com",
    "phone": "+1-555-987-6543",
    "website": "https://www.newbusiness.com"
  },
  "address": {
    "street": "456 Tech Ave",
    "city": "Tech City",
    "state": "CA",
    "zip": "90210",
    "country": "US"
  }
}
```

**Response**:
```json
{
  "id": "merchant_456",
  "name": "New Business Inc",
  "business_type": "Technology",
  "industry_code": "7372",
  "portfolio_type": "pending",
  "risk_level": "medium",
  "contact_info": {
    "email": "info@newbusiness.com",
    "phone": "+1-555-987-6543",
    "website": "https://www.newbusiness.com"
  },
  "address": {
    "street": "456 Tech Ave",
    "city": "Tech City",
    "state": "CA",
    "zip": "90210",
    "country": "US"
  },
  "verification_status": {
    "identity_verified": false,
    "address_verified": false,
    "business_verified": false,
    "last_verified": null
  },
  "created_at": "2025-01-19T11:00:00Z",
  "updated_at": "2025-01-19T11:00:00Z"
}
```

#### Update Merchant

```http
PUT /api/v1/merchants/{id}
```

**Request Body**:
```json
{
  "name": "Updated Business Name",
  "portfolio_type": "onboarded",
  "risk_level": "low"
}
```

**Response**:
```json
{
  "id": "merchant_123",
  "name": "Updated Business Name",
  "business_type": "Retail",
  "industry_code": "5411",
  "portfolio_type": "onboarded",
  "risk_level": "low",
  "contact_info": {
    "email": "contact@acme.com",
    "phone": "+1-555-123-4567",
    "website": "https://www.acme.com"
  },
  "address": {
    "street": "123 Main St",
    "city": "Anytown",
    "state": "CA",
    "zip": "12345",
    "country": "US"
  },
  "verification_status": {
    "identity_verified": true,
    "address_verified": true,
    "business_verified": true,
    "last_verified": "2025-01-15T14:30:00Z"
  },
  "created_at": "2025-01-19T10:30:00Z",
  "updated_at": "2025-01-19T12:00:00Z"
}
```

#### Delete Merchant

```http
DELETE /api/v1/merchants/{id}
```

**Response**:
```json
{
  "message": "Merchant deleted successfully"
}
```

### Bulk Operations

#### Bulk Update Merchants

```http
POST /api/v1/merchants/bulk
```

**Request Body**:
```json
{
  "operation": "update",
  "merchant_ids": ["merchant_123", "merchant_456"],
  "updates": {
    "portfolio_type": "onboarded",
    "risk_level": "low"
  }
}
```

**Response**:
```json
{
  "operation_id": "bulk_op_789",
  "status": "processing",
  "total_merchants": 2,
  "processed": 0,
  "failed": 0,
  "created_at": "2025-01-19T12:00:00Z"
}
```

#### Get Bulk Operation Status

```http
GET /api/v1/merchants/bulk/{operation_id}
```

**Response**:
```json
{
  "operation_id": "bulk_op_789",
  "status": "completed",
  "total_merchants": 2,
  "processed": 2,
  "failed": 0,
  "results": [
    {
      "merchant_id": "merchant_123",
      "status": "success",
      "updated_at": "2025-01-19T12:01:00Z"
    },
    {
      "merchant_id": "merchant_456",
      "status": "success",
      "updated_at": "2025-01-19T12:01:00Z"
    }
  ],
  "created_at": "2025-01-19T12:00:00Z",
  "completed_at": "2025-01-19T12:01:00Z"
}
```

### Session Management

#### Get Current Session

```http
GET /api/v1/session
```

**Response**:
```json
{
  "session_id": "session_abc123",
  "user_id": "user_456",
  "active_merchant_id": "merchant_123",
  "created_at": "2025-01-19T10:00:00Z",
  "last_accessed": "2025-01-19T12:00:00Z"
}
```

#### Set Active Merchant

```http
POST /api/v1/session/merchant
```

**Request Body**:
```json
{
  "merchant_id": "merchant_123"
}
```

**Response**:
```json
{
  "session_id": "session_abc123",
  "user_id": "user_456",
  "active_merchant_id": "merchant_123",
  "created_at": "2025-01-19T10:00:00Z",
  "last_accessed": "2025-01-19T12:00:00Z"
}
```

### Search and Filtering

#### Search Merchants

```http
GET /api/v1/merchants/search
```

**Query Parameters**:
- `q` (required): Search query
- `filters` (optional): JSON string of filters
- `page` (optional): Page number
- `limit` (optional): Items per page

**Example**:
```http
GET /api/v1/merchants/search?q=acme&filters={"portfolio_type":"onboarded","risk_level":"low"}&page=1&limit=10
```

**Response**:
```json
{
  "data": [
    {
      "id": "merchant_123",
      "name": "Acme Corporation",
      "business_type": "Retail",
      "portfolio_type": "onboarded",
      "risk_level": "low",
      "relevance_score": 0.95
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 1,
    "total_pages": 1
  },
  "search_metadata": {
    "query": "acme",
    "filters": {
      "portfolio_type": "onboarded",
      "risk_level": "low"
    },
    "execution_time_ms": 45
  }
}
```

## Error Handling

### Error Response Format

All errors follow a consistent format:

```json
{
  "error": {
    "code": "MERCHANT_NOT_FOUND",
    "message": "Merchant with ID 'merchant_123' not found",
    "details": {
      "merchant_id": "merchant_123",
      "timestamp": "2025-01-19T12:00:00Z",
      "request_id": "req_abc123"
    }
  }
}
```

### HTTP Status Codes

- `200 OK`: Successful request
- `201 Created`: Resource created successfully
- `400 Bad Request`: Invalid request data
- `401 Unauthorized`: Authentication required
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `409 Conflict`: Resource conflict
- `422 Unprocessable Entity`: Validation error
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Server error

### Common Error Codes

| Code | Description |
|------|-------------|
| `INVALID_REQUEST` | Request data is invalid |
| `MERCHANT_NOT_FOUND` | Merchant does not exist |
| `DUPLICATE_MERCHANT` | Merchant already exists |
| `VALIDATION_ERROR` | Data validation failed |
| `UNAUTHORIZED` | Authentication required |
| `FORBIDDEN` | Insufficient permissions |
| `RATE_LIMIT_EXCEEDED` | Too many requests |
| `INTERNAL_ERROR` | Server error |

## Rate Limiting

### Rate Limits

- **General API**: 1000 requests per hour per user
- **Bulk Operations**: 10 requests per hour per user
- **Search**: 500 requests per hour per user

### Rate Limit Headers

```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1642593600
```

### Rate Limit Exceeded Response

```json
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Rate limit exceeded. Try again later.",
    "details": {
      "limit": 1000,
      "remaining": 0,
      "reset_time": "2025-01-19T13:00:00Z"
    }
  }
}
```

## API Development Guidelines

### Handler Implementation

```go
// MerchantHandler handles merchant-related HTTP requests
type MerchantHandler struct {
    service MerchantService
    logger  *zap.Logger
}

// NewMerchantHandler creates a new merchant handler
func NewMerchantHandler(service MerchantService, logger *zap.Logger) *MerchantHandler {
    return &MerchantHandler{
        service: service,
        logger:  logger,
    }
}

// GetMerchants handles GET /api/v1/merchants
func (h *MerchantHandler) GetMerchants(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Parse query parameters
    filters, err := parseMerchantFilters(r.URL.Query())
    if err != nil {
        h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
        return
    }
    
    // Get merchants from service
    merchants, pagination, err := h.service.GetMerchants(ctx, filters)
    if err != nil {
        h.logger.Error("failed to get merchants", zap.Error(err))
        h.writeErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve merchants")
        return
    }
    
    // Write response
    response := MerchantListResponse{
        Data:       merchants,
        Pagination: pagination,
    }
    
    h.writeJSONResponse(w, http.StatusOK, response)
}
```

### Request Validation

```go
// ValidateMerchantRequest validates merchant creation/update requests
func ValidateMerchantRequest(req *MerchantRequest) error {
    var errors []string
    
    if req.Name == "" {
        errors = append(errors, "name is required")
    }
    
    if len(req.Name) > 255 {
        errors = append(errors, "name exceeds maximum length of 255 characters")
    }
    
    if req.ContactInfo != nil {
        if req.ContactInfo.Email != "" && !isValidEmail(req.ContactInfo.Email) {
            errors = append(errors, "invalid email format")
        }
        
        if req.ContactInfo.Phone != "" && !isValidPhone(req.ContactInfo.Phone) {
            errors = append(errors, "invalid phone format")
        }
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("validation failed: %s", strings.Join(errors, "; "))
    }
    
    return nil
}
```

### Response Helpers

```go
// writeJSONResponse writes a JSON response
func (h *MerchantHandler) writeJSONResponse(w http.ResponseWriter, status int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    
    if err := json.NewEncoder(w).Encode(data); err != nil {
        h.logger.Error("failed to encode JSON response", zap.Error(err))
    }
}

// writeErrorResponse writes an error response
func (h *MerchantHandler) writeErrorResponse(w http.ResponseWriter, status int, code, message string) {
    errorResponse := ErrorResponse{
        Error: ErrorDetail{
            Code:    code,
            Message: message,
            Details: map[string]interface{}{
                "timestamp":  time.Now().UTC(),
                "request_id": getRequestID(w),
            },
        },
    }
    
    h.writeJSONResponse(w, status, errorResponse)
}
```

### Middleware Implementation

```go
// AuthMiddleware handles authentication
func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Extract token from Authorization header
            authHeader := r.Header.Get("Authorization")
            if !strings.HasPrefix(authHeader, "Bearer ") {
                writeErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or invalid authorization header")
                return
            }
            
            tokenString := strings.TrimPrefix(authHeader, "Bearer ")
            
            // Validate token
            claims, err := validateJWT(tokenString, jwtSecret)
            if err != nil {
                writeErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid token")
                return
            }
            
            // Add user info to context
            ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
            ctx = context.WithValue(ctx, "user_roles", claims.Roles)
            
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// LoggingMiddleware logs HTTP requests
func LoggingMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            
            // Wrap response writer to capture status code
            wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
            
            next.ServeHTTP(wrapped, r)
            
            // Log request
            logger.Info("HTTP request",
                zap.String("method", r.Method),
                zap.String("url", r.URL.String()),
                zap.Int("status", wrapped.statusCode),
                zap.Duration("duration", time.Since(start)),
                zap.String("user_agent", r.UserAgent()),
                zap.String("remote_addr", r.RemoteAddr),
            )
        })
    }
}
```

## Testing API Endpoints

### Unit Testing Handlers

```go
func TestMerchantHandler_GetMerchants(t *testing.T) {
    tests := []struct {
        name           string
        queryParams    map[string]string
        mockSetup      func(*mocks.MerchantService)
        expectedStatus int
        expectedError  string
    }{
        {
            name: "successful request",
            queryParams: map[string]string{
                "page": "1",
                "limit": "10",
            },
            mockSetup: func(service *mocks.MerchantService) {
                service.On("GetMerchants", mock.Anything, mock.Anything).
                    Return([]Merchant{{ID: "1", Name: "Test"}}, Pagination{}, nil)
            },
            expectedStatus: http.StatusOK,
        },
        {
            name: "invalid page parameter",
            queryParams: map[string]string{
                "page": "invalid",
            },
            expectedStatus: http.StatusBadRequest,
            expectedError:  "INVALID_REQUEST",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup
            mockService := &mocks.MerchantService{}
            if tt.mockSetup != nil {
                tt.mockSetup(mockService)
            }
            
            handler := NewMerchantHandler(mockService, zap.NewNop())
            
            // Create request
            req := httptest.NewRequest("GET", "/api/v1/merchants", nil)
            q := req.URL.Query()
            for key, value := range tt.queryParams {
                q.Add(key, value)
            }
            req.URL.RawQuery = q.Encode()
            
            // Create response recorder
            w := httptest.NewRecorder()
            
            // Execute
            handler.GetMerchants(w, req)
            
            // Assert
            assert.Equal(t, tt.expectedStatus, w.Code)
            
            if tt.expectedError != "" {
                var errorResp ErrorResponse
                err := json.Unmarshal(w.Body.Bytes(), &errorResp)
                require.NoError(t, err)
                assert.Equal(t, tt.expectedError, errorResp.Error.Code)
            }
            
            mockService.AssertExpectations(t)
        })
    }
}
```

### Integration Testing

```go
func TestMerchantAPI_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    
    // Setup test server
    server := setupTestServer(t)
    defer server.Close()
    
    t.Run("create and retrieve merchant", func(t *testing.T) {
        // Create merchant
        merchantData := MerchantRequest{
            Name:         "Test Merchant",
            BusinessType: "Retail",
            ContactInfo: &ContactInfo{
                Email: "test@merchant.com",
            },
        }
        
        createResp, err := http.Post(
            server.URL+"/api/v1/merchants",
            "application/json",
            strings.NewReader(toJSON(merchantData)),
        )
        require.NoError(t, err)
        require.Equal(t, http.StatusCreated, createResp.StatusCode)
        
        var createdMerchant Merchant
        err = json.NewDecoder(createResp.Body).Decode(&createdMerchant)
        require.NoError(t, err)
        
        // Retrieve merchant
        getResp, err := http.Get(server.URL + "/api/v1/merchants/" + createdMerchant.ID)
        require.NoError(t, err)
        require.Equal(t, http.StatusOK, getResp.StatusCode)
        
        var retrievedMerchant Merchant
        err = json.NewDecoder(getResp.Body).Decode(&retrievedMerchant)
        require.NoError(t, err)
        
        assert.Equal(t, createdMerchant.ID, retrievedMerchant.ID)
        assert.Equal(t, createdMerchant.Name, retrievedMerchant.Name)
    })
}
```

## API Documentation

### OpenAPI Specification

The API is documented using OpenAPI 3.0 specification:

```yaml
openapi: 3.0.3
info:
  title: KYB Platform API
  description: Merchant-centric business verification platform
  version: 1.0.0
  contact:
    name: API Support
    email: api-support@kyb-platform.com

servers:
  - url: https://api.kyb-platform.com/v1
    description: Production server
  - url: https://staging-api.kyb-platform.com/v1
    description: Staging server

paths:
  /merchants:
    get:
      summary: List merchants
      description: Retrieve a paginated list of merchants
      parameters:
        - name: page
          in: query
          schema:
            type: integer
            minimum: 1
            default: 1
        - name: limit
          in: query
          schema:
            type: integer
            minimum: 1
            maximum: 100
            default: 20
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/MerchantListResponse'
        '400':
          description: Bad request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ErrorResponse'
```

### API Documentation Generation

```bash
# Generate API documentation
swagger generate spec -o api-docs.json

# Serve API documentation
swagger serve api-docs.json
```

## Performance Considerations

### Database Optimization

```go
// Optimized query with proper indexing
func (r *MerchantRepository) GetMerchants(ctx context.Context, filters MerchantFilters) ([]Merchant, error) {
    query := `
        SELECT id, name, business_type, portfolio_type, risk_level, created_at, updated_at
        FROM merchants
        WHERE 1=1
    `
    
    args := []interface{}{}
    argIndex := 1
    
    if filters.PortfolioType != "" {
        query += fmt.Sprintf(" AND portfolio_type = $%d", argIndex)
        args = append(args, filters.PortfolioType)
        argIndex++
    }
    
    if filters.RiskLevel != "" {
        query += fmt.Sprintf(" AND risk_level = $%d", argIndex)
        args = append(args, filters.RiskLevel)
        argIndex++
    }
    
    query += " ORDER BY created_at DESC"
    
    if filters.Limit > 0 {
        query += fmt.Sprintf(" LIMIT $%d", argIndex)
        args = append(args, filters.Limit)
        argIndex++
    }
    
    if filters.Offset > 0 {
        query += fmt.Sprintf(" OFFSET $%d", argIndex)
        args = append(args, filters.Offset)
    }
    
    rows, err := r.db.QueryContext(ctx, query, args...)
    if err != nil {
        return nil, fmt.Errorf("failed to query merchants: %w", err)
    }
    defer rows.Close()
    
    var merchants []Merchant
    for rows.Next() {
        var merchant Merchant
        err := rows.Scan(
            &merchant.ID,
            &merchant.Name,
            &merchant.BusinessType,
            &merchant.PortfolioType,
            &merchant.RiskLevel,
            &merchant.CreatedAt,
            &merchant.UpdatedAt,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to scan merchant: %w", err)
        }
        merchants = append(merchants, merchant)
    }
    
    return merchants, nil
}
```

### Caching Strategy

```go
// Cached merchant service
type CachedMerchantService struct {
    service MerchantService
    cache   Cache
    ttl     time.Duration
}

func (s *CachedMerchantService) GetMerchant(ctx context.Context, id string) (*Merchant, error) {
    // Try cache first
    cacheKey := fmt.Sprintf("merchant:%s", id)
    if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
        var merchant Merchant
        if err := json.Unmarshal(cached, &merchant); err == nil {
            return &merchant, nil
        }
    }
    
    // Get from service
    merchant, err := s.service.GetMerchant(ctx, id)
    if err != nil {
        return nil, err
    }
    
    // Cache result
    if data, err := json.Marshal(merchant); err == nil {
        s.cache.Set(ctx, cacheKey, data, s.ttl)
    }
    
    return merchant, nil
}
```

## Security Considerations

### Input Validation

```go
// Comprehensive input validation
func ValidateMerchantInput(input *MerchantInput) error {
    var errors []string
    
    // Name validation
    if input.Name == "" {
        errors = append(errors, "name is required")
    } else if len(input.Name) > 255 {
        errors = append(errors, "name exceeds maximum length")
    } else if containsSQLInjection(input.Name) {
        errors = append(errors, "name contains invalid characters")
    }
    
    // Email validation
    if input.Email != "" {
        if !isValidEmail(input.Email) {
            errors = append(errors, "invalid email format")
        }
    }
    
    // Phone validation
    if input.Phone != "" {
        if !isValidPhone(input.Phone) {
            errors = append(errors, "invalid phone format")
        }
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("validation failed: %s", strings.Join(errors, "; "))
    }
    
    return nil
}
```

### SQL Injection Prevention

```go
// Use parameterized queries
func (r *MerchantRepository) GetMerchantByID(ctx context.Context, id string) (*Merchant, error) {
    query := "SELECT id, name, business_type FROM merchants WHERE id = $1"
    
    var merchant Merchant
    err := r.db.QueryRowContext(ctx, query, id).Scan(
        &merchant.ID,
        &merchant.Name,
        &merchant.BusinessType,
    )
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, ErrMerchantNotFound
        }
        return nil, fmt.Errorf("failed to get merchant: %w", err)
    }
    
    return &merchant, nil
}
```

## Monitoring and Observability

### Request Metrics

```go
// Metrics middleware
func MetricsMiddleware() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            start := time.Now()
            
            // Wrap response writer
            wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
            
            next.ServeHTTP(wrapped, r)
            
            // Record metrics
            duration := time.Since(start)
            
            // Increment request counter
            requestCounter.WithLabelValues(
                r.Method,
                r.URL.Path,
                strconv.Itoa(wrapped.statusCode),
            ).Inc()
            
            // Record request duration
            requestDuration.WithLabelValues(
                r.Method,
                r.URL.Path,
            ).Observe(duration.Seconds())
        })
    }
}
```

### Health Checks

```go
// Health check endpoint
func HealthCheck(w http.ResponseWriter, r *http.Request) {
    health := map[string]interface{}{
        "status":    "healthy",
        "timestamp": time.Now().UTC(),
        "version":   version,
        "services": map[string]string{
            "database": checkDatabase(),
            "redis":    checkRedis(),
        },
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(health)
}
```

## Conclusion

This API development guide provides comprehensive documentation for developing and maintaining the KYB Platform API. Follow these guidelines to ensure consistent, secure, and performant API development.

For additional information, refer to the architecture documentation and deployment guides.
