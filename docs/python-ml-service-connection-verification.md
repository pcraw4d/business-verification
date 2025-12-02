# Python ML Service Connection Verification Guide

## Overview

This document explains how the Go classification service connects to the Python ML service, potential connection issues, and how language differences (Go vs Python) are handled.

## Communication Architecture

### Language Independence ✅

**The language difference (Go vs Python) does NOT impact connectivity** because:

1. **HTTP/REST API Communication**: Both services communicate via standard HTTP/REST APIs
2. **JSON Data Exchange**: All data is exchanged as JSON, which is language-agnostic
3. **Standard Protocols**: Uses standard HTTP methods (GET, POST) and status codes
4. **No Direct Code Dependencies**: Go service doesn't import Python code - it's a pure HTTP client

### Connection Flow

```
Classification Service (Go)
    │
    ├─ Reads: PYTHON_ML_SERVICE_URL environment variable
    │
    ├─ Creates: HTTP client with 30s timeout
    │
    ├─ Initializes: PythonMLService client
    │   └─ Normalizes URL (removes trailing slash)
    │   └─ Creates circuit breaker for resilience
    │
    ├─ Tests Connection: GET /ping
    │   └─ 10 second timeout during initialization
    │
    ├─ Loads Models: GET /models
    │   └─ Returns empty list if models not loaded (no error)
    │
    └─ Ready for Requests: POST /classify-enhanced
        └─ 30 second timeout per request
        └─ Circuit breaker protects against cascading failures
```

## Connection Configuration

### Required Environment Variable

**For Classification Service:**
```bash
PYTHON_ML_SERVICE_URL=https://python-ml-service-production-xxx.up.railway.app
```

**Important Notes:**
- URL should NOT have trailing slash (automatically normalized)
- Must be accessible from Railway's network
- Should use HTTPS in production

### Connection Initialization

The classification service initializes the Python ML service in `cmd/main.go`:

```go
// Read environment variable
pythonMLServiceURL := os.Getenv("PYTHON_ML_SERVICE_URL")

if pythonMLServiceURL != "" {
    // Create client
    pythonMLService = infrastructure.NewPythonMLService(pythonMLServiceURL, stdLogger)
    
    // Test connection (10 second timeout)
    if err := pythonMLService.Initialize(initCtx); err != nil {
        // Graceful fallback - service continues without ML
        pythonMLService = nil
    }
}
```

## Potential Connection Issues

### 1. Environment Variable Not Set

**Symptom:**
- Classification service logs: "ℹ️ Python ML Service URL not configured"
- No Python ML initialization attempts

**Solution:**
```bash
# Set in Railway
railway variables set PYTHON_ML_SERVICE_URL="https://python-ml-service-production.up.railway.app" --service classification-service
```

### 2. Network Connectivity Issues

**Symptoms:**
- Initialization fails with connection errors
- Timeout errors during requests
- 502/503 errors from Python service

**Causes:**
- Python ML service not running
- Wrong URL (typo, wrong service)
- Network firewall blocking
- Railway service not publicly accessible

**Verification:**
```bash
# Test from command line
curl https://python-ml-service-production.up.railway.app/ping
# Expected: {"status":"ok","message":"Python ML Service is running"}

# Test health
curl https://python-ml-service-production.up.railway.app/health
# Expected: {"status":"healthy",...}
```

### 3. Service Discovery Issues

**In Railway:**
- Each service gets a unique public URL
- Services can communicate via public URLs (HTTPS)
- No internal service discovery needed
- Services are isolated by default

**Best Practice:**
- Use Railway's public domain for each service
- Set `PYTHON_ML_SERVICE_URL` explicitly
- Don't rely on internal networking (not available)

### 4. Timeout Issues

**Current Timeouts:**
- **Initialization**: 10 seconds (connection test)
- **HTTP Client**: 30 seconds (per request)
- **Model Loading**: 30 seconds (background)

**If timeouts occur:**
- Python service may be slow to respond
- Models may still be loading
- Network latency between services

**Solution:**
- Python service loads models in background (non-blocking)
- Classification service has graceful fallback
- Circuit breaker prevents cascading failures

### 5. Circuit Breaker Protection

The Go service includes a circuit breaker that:
- Opens after 5 consecutive failures
- Stays open for 30 seconds
- Needs 2 successes to close
- Prevents overwhelming a failing service

**This means:**
- Temporary failures are handled gracefully
- Service doesn't retry indefinitely
- Automatic recovery when service is healthy

## Connection Verification Steps

### Step 1: Verify Environment Variable

```bash
# Check if set in Railway
railway variables --service classification-service | grep PYTHON_ML_SERVICE_URL

# Or check logs during startup
# Look for: "🐍 Initializing Python ML Service at https://..."
```

### Step 2: Verify Python ML Service is Running

```bash
# Test ping endpoint
curl https://python-ml-service-production.up.railway.app/ping

# Test health endpoint
curl https://python-ml-service-production.up.railway.app/health

# Test models endpoint (should return 200, even if empty)
curl https://python-ml-service-production.up.railway.app/models
```

### Step 3: Check Classification Service Logs

**Successful Initialization:**
```
🐍 Initializing Python ML Service at https://...
✅ Python ML Service initialized successfully
✅ Classification services initialized
  python_ml_service: true
```

**Failed Initialization (with fallback):**
```
🐍 Initializing Python ML Service at https://...
⚠️ Failed to initialize Python ML Service, continuing without enhanced classification
  error: failed to connect to Python ML service: ...
✅ Classification services initialized
  python_ml_service: false
```

### Step 4: Test Classification Request

```bash
curl -X POST https://classification-service-production.up.railway.app/v1/classify \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Test Business",
    "description": "Software development",
    "website_url": "https://example.com"
  }'
```

**Expected Behavior:**
- If Python ML service is available: Uses ensemble voting (Python ML + Go)
- If Python ML service unavailable: Uses Go classification only
- Both scenarios return valid responses

## Language Difference Handling

### HTTP/REST Communication

**Go Service (Client):**
```go
// Creates HTTP request
httpReq, err := http.NewRequestWithContext(ctx, "POST", 
    pms.endpoint+"/classify-enhanced", bytes.NewBuffer(requestBody))
httpReq.Header.Set("Content-Type", "application/json")

// Sends request
resp, err := pms.httpClient.Do(httpReq)

// Parses JSON response
json.NewDecoder(resp.Body).Decode(&response)
```

**Python Service (Server):**
```python
# FastAPI endpoint
@app.post("/classify-enhanced", response_model=EnhancedClassificationResponse)
async def classify_enhanced(request: EnhancedClassificationRequest):
    # Process request
    result = distilbart_classifier.classify(...)
    # Return JSON response
    return EnhancedClassificationResponse(...)
```

### Data Serialization

**Request (Go → Python):**
```go
type EnhancedClassificationRequest struct {
    BusinessName     string `json:"business_name"`
    Description      string `json:"description"`
    WebsiteURL       string `json:"website_url"`
    MaxResults       int    `json:"max_results"`
    MaxContentLength int    `json:"max_content_length"`
}
```

**Response (Python → Go):**
```go
type EnhancedClassificationResponse struct {
    Success         bool                `json:"success"`
    Classifications []Classification     `json:"classifications"`
    Confidence      float64             `json:"confidence"`
    Explanation     string              `json:"explanation"`
    Summary         string              `json:"summary"`
    // ... more fields
}
```

**No Language-Specific Issues:**
- JSON is standardized across languages
- HTTP is language-agnostic
- FastAPI automatically handles JSON serialization
- Go's `encoding/json` handles deserialization

## Network Architecture in Railway

### Service Isolation

```
┌─────────────────────────────────────┐
│  Railway Project                    │
│                                     │
│  ┌──────────────────────────────┐  │
│  │ Classification Service (Go)  │  │
│  │ Port: 8080                   │  │
│  │ Public: https://...          │  │
│  └──────────┬───────────────────┘  │
│             │                       │
│             │ HTTPS                 │
│             │ (Public Internet)     │
│             │                       │
│  ┌──────────▼───────────────────┐  │
│  │ Python ML Service (Python)   │  │
│  │ Port: 8080                   │  │
│  │ Public: https://...          │  │
│  └──────────────────────────────┘  │
└─────────────────────────────────────┘
```

**Key Points:**
- Services communicate via public HTTPS URLs
- No internal networking (each service is isolated)
- Railway handles SSL/TLS termination
- Services can be in different regions/data centers

## Troubleshooting Checklist

- [ ] `PYTHON_ML_SERVICE_URL` is set in Railway for classification-service
- [ ] Python ML service is deployed and running
- [ ] Python ML service `/ping` endpoint returns 200 OK
- [ ] Python ML service `/health` endpoint returns healthy
- [ ] Classification service logs show successful initialization
- [ ] Network connectivity works (test with curl)
- [ ] No firewall blocking HTTPS traffic
- [ ] URLs are correct (no typos, correct service names)
- [ ] Services are in the same Railway project (or URLs are publicly accessible)

## Best Practices

1. **Always Set Environment Variable**: Use Railway CLI or dashboard
2. **Monitor Initialization Logs**: Check for connection errors
3. **Test Endpoints Manually**: Verify Python service is accessible
4. **Use HTTPS**: Always use HTTPS in production
5. **Handle Failures Gracefully**: Service should work without Python ML (fallback)
6. **Monitor Circuit Breaker**: Check if it's opening/closing
7. **Set Appropriate Timeouts**: Current 30s is reasonable for ML inference

## Summary

✅ **Language difference does NOT impact connectivity** - HTTP/REST is language-agnostic
✅ **Connection is via standard HTTPS** - No special networking needed
✅ **Graceful fallback** - Service works even if Python ML is unavailable
✅ **Circuit breaker protection** - Prevents cascading failures
✅ **Standard JSON communication** - Works seamlessly between Go and Python

The main potential issues are:
1. Environment variable not set
2. Network connectivity problems
3. Python service not running
4. Timeout issues (rare, handled by circuit breaker)

