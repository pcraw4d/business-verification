# Common Error Codes and Solutions

## Overview

This document provides a comprehensive reference for common error codes, their causes, and solutions when using the Risk Assessment Service API.

## HTTP Status Codes

### 2xx Success Codes

#### 200 OK
**Description:** Request completed successfully
**Example Response:**
```json
{
  "id": "risk_1234567890",
  "business_id": "biz_1234567890",
  "risk_score": 0.75,
  "risk_level": "medium",
  "status": "completed",
  "created_at": "2024-01-15T10:30:00Z"
}
```

#### 201 Created
**Description:** Resource created successfully
**Example Response:**
```json
{
  "webhook_id": "wh_1234567890",
  "url": "https://your-app.com/webhooks/risk-assessment",
  "active": true,
  "created_at": "2024-01-15T10:30:00Z"
}
```

#### 202 Accepted
**Description:** Request accepted for processing
**Example Response:**
```json
{
  "batch_id": "batch_1234567890",
  "status": "queued",
  "total_businesses": 100,
  "estimated_completion_time": "2024-01-15T11:30:00Z"
}
```

### 4xx Client Error Codes

#### 400 Bad Request
**Description:** Invalid request data or format

**Common Causes:**
- Missing required fields
- Invalid data types
- Malformed JSON
- Invalid field values

**Example Error Response:**
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Request validation failed",
    "details": "business_name is required",
    "field": "business_name",
    "validation": [
      {
        "field": "business_name",
        "message": "business_name is required",
        "code": "INVALID_BUSINESS_NAME"
      }
    ]
  },
  "request_id": "req_1234567890",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

**Solutions:**
1. **Check Required Fields:**
   ```javascript
   const requiredFields = ['business_name', 'business_address', 'industry', 'country'];
   const missingFields = requiredFields.filter(field => !data[field]);
   
   if (missingFields.length > 0) {
     throw new Error(`Missing required fields: ${missingFields.join(', ')}`);
   }
   ```

2. **Validate Data Types:**
   ```javascript
   function validateDataTypes(data) {
     if (typeof data.business_name !== 'string') {
       throw new Error('business_name must be a string');
     }
     if (typeof data.risk_score !== 'number') {
       throw new Error('risk_score must be a number');
     }
   }
   ```

3. **Check JSON Format:**
   ```javascript
   try {
     const requestBody = JSON.stringify(data);
     JSON.parse(requestBody); // Validate JSON
   } catch (error) {
     throw new Error('Invalid JSON format');
   }
   ```

#### 401 Unauthorized
**Description:** Authentication failed or missing

**Common Causes:**
- Invalid API key
- Missing API key
- Expired JWT token
- Malformed authorization header

**Example Error Response:**
```json
{
  "error": {
    "code": "AUTHENTICATION_ERROR",
    "message": "Invalid API key",
    "details": "The provided API key is invalid or has been revoked"
  },
  "request_id": "req_1234567890",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

**Solutions:**
1. **Verify API Key Format:**
   ```bash
   # Test keys start with sk_test_
   # Live keys start with sk_live_
   echo "YOUR_API_KEY" | grep -E "^(sk_test_|sk_live_)"
   ```

2. **Check Authorization Header:**
   ```javascript
   const headers = {
     'Authorization': `Bearer ${API_KEY}`, // Correct format
     'Content-Type': 'application/json'
   };
   ```

3. **Test API Key:**
   ```bash
   curl -X GET "https://risk-assessment-service-production.up.railway.app/api/v1/health" \
     -H "Authorization: Bearer YOUR_API_KEY"
   ```

#### 403 Forbidden
**Description:** Access denied or insufficient permissions

**Common Causes:**
- IP address not whitelisted
- Insufficient API key permissions
- Resource access denied
- Rate limit exceeded

**Example Error Response:**
```json
{
  "error": {
    "code": "AUTHORIZATION_ERROR",
    "message": "IP address not allowed",
    "details": "Your IP address is not in the allowed list for this API key"
  },
  "request_id": "req_1234567890",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

**Solutions:**
1. **Check IP Whitelist:**
   ```bash
   # Get current IP
   curl https://ipinfo.io/ip
   
   # Add IP to whitelist
   curl -X PUT "https://api.kyb-platform.com/v1/keys/YOUR_API_KEY" \
     -H "Authorization: Bearer YOUR_API_KEY" \
     -H "Content-Type: application/json" \
     -d '{"allowed_ips": ["YOUR_IP_ADDRESS/32"]}'
   ```

2. **Check API Key Permissions:**
   ```bash
   curl -X GET "https://api.kyb-platform.com/v1/keys/YOUR_API_KEY" \
     -H "Authorization: Bearer YOUR_API_KEY"
   ```

#### 404 Not Found
**Description:** Resource not found

**Common Causes:**
- Invalid endpoint URL
- Resource ID doesn't exist
- Incorrect API version
- Deleted resource

**Example Error Response:**
```json
{
  "error": {
    "code": "RESOURCE_NOT_FOUND",
    "message": "Risk assessment not found",
    "details": "The requested risk assessment does not exist or has been deleted"
  },
  "request_id": "req_1234567890",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

**Solutions:**
1. **Verify Endpoint URL:**
   ```bash
   # Correct base URL
   https://risk-assessment-service-production.up.railway.app/api/v1
   
   # Check endpoint exists
   curl -X GET "https://risk-assessment-service-production.up.railway.app/api/v1/health"
   ```

2. **Check Resource ID:**
   ```javascript
   // Validate ID format
   if (!/^risk_\d+$/.test(assessmentId)) {
     throw new Error('Invalid assessment ID format');
   }
   ```

#### 429 Rate Limit Exceeded
**Description:** Too many requests

**Common Causes:**
- Exceeded requests per minute limit
- Exceeded daily request limit
- Burst of requests
- Missing rate limit handling

**Example Error Response:**
```json
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "Rate limit exceeded",
    "details": "You have exceeded your rate limit of 100 requests per minute"
  },
  "request_id": "req_1234567890",
  "timestamp": "2024-01-15T10:30:00Z",
  "rate_limit": {
    "limit": 100,
    "remaining": 0,
    "reset_time": "2024-01-15T11:00:00Z"
  }
}
```

**Solutions:**
1. **Implement Exponential Backoff:**
   ```javascript
   async function makeAPICallWithRetry(url, options, maxRetries = 3) {
     for (let attempt = 0; attempt < maxRetries; attempt++) {
       try {
         const response = await fetch(url, options);
         
         if (response.status === 429) {
           const waitTime = Math.pow(2, attempt) * 1000;
           console.log(`Rate limited. Waiting ${waitTime}ms before retry ${attempt + 1}`);
           await new Promise(resolve => setTimeout(resolve, waitTime));
           continue;
         }
         
         return response;
       } catch (error) {
         if (attempt === maxRetries - 1) throw error;
       }
     }
   }
   ```

2. **Monitor Rate Limit Headers:**
   ```javascript
   function checkRateLimit(response) {
     const remaining = parseInt(response.headers.get('X-RateLimit-Remaining'));
     const limit = parseInt(response.headers.get('X-RateLimit-Limit'));
     
     if (remaining / limit < 0.1) {
       console.warn('Rate limit nearly exceeded. Consider implementing request queuing.');
     }
   }
   ```

### 5xx Server Error Codes

#### 500 Internal Server Error
**Description:** Server encountered an unexpected error

**Common Causes:**
- Database connection issues
- External API failures
- Model prediction errors
- System overload

**Example Error Response:**
```json
{
  "error": {
    "code": "INTERNAL_SERVER_ERROR",
    "message": "An unexpected error occurred",
    "details": "Please try again later or contact support if the issue persists"
  },
  "request_id": "req_1234567890",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

**Solutions:**
1. **Retry with Backoff:**
   ```javascript
   async function retryOnServerError(apiCall, maxRetries = 3) {
     for (let attempt = 0; attempt < maxRetries; attempt++) {
       try {
         return await apiCall();
       } catch (error) {
         if (error.status >= 500 && attempt < maxRetries - 1) {
           const waitTime = Math.pow(2, attempt) * 1000;
           await new Promise(resolve => setTimeout(resolve, waitTime));
           continue;
         }
         throw error;
       }
     }
   }
   ```

2. **Check Service Status:**
   ```bash
   curl -X GET "https://status.kyb-platform.com/api/status"
   ```

#### 502 Bad Gateway
**Description:** Upstream service error

**Common Causes:**
- External API timeout
- Database connection issues
- Load balancer problems
- Service dependency failure

**Example Error Response:**
```json
{
  "error": {
    "code": "BAD_GATEWAY",
    "message": "Upstream service error",
    "details": "External API temporarily unavailable"
  },
  "request_id": "req_1234567890",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

**Solutions:**
1. **Implement Fallback Logic:**
   ```javascript
   async function assessWithFallback(businessData) {
     try {
       return await assessRisk(businessData, { include_external: true });
     } catch (error) {
       if (error.status === 502) {
         console.log('External APIs unavailable, using internal assessment');
         return await assessRisk(businessData, { include_external: false });
       }
       throw error;
     }
   }
   ```

#### 503 Service Unavailable
**Description:** Service temporarily unavailable

**Common Causes:**
- Maintenance mode
- System overload
- Database maintenance
- External service outage

**Example Error Response:**
```json
{
  "error": {
    "code": "SERVICE_UNAVAILABLE",
    "message": "Service temporarily unavailable",
    "details": "The service is currently undergoing maintenance"
  },
  "request_id": "req_1234567890",
  "timestamp": "2024-01-15T10:30:00Z",
  "retry_after": 300
}
```

**Solutions:**
1. **Respect Retry-After Header:**
   ```javascript
   if (response.status === 503) {
     const retryAfter = response.headers.get('Retry-After');
     if (retryAfter) {
       console.log(`Service unavailable. Retry after ${retryAfter} seconds`);
       await new Promise(resolve => setTimeout(resolve, parseInt(retryAfter) * 1000));
     }
   }
   ```

#### 504 Gateway Timeout
**Description:** Request timeout

**Common Causes:**
- Slow external API responses
- Database query timeout
- Network issues
- System overload

**Example Error Response:**
```json
{
  "error": {
    "code": "GATEWAY_TIMEOUT",
    "message": "Request timeout",
    "details": "The request took too long to process"
  },
  "request_id": "req_1234567890",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

**Solutions:**
1. **Increase Timeout:**
   ```javascript
   const controller = new AbortController();
   const timeoutId = setTimeout(() => controller.abort(), 30000); // 30 seconds
   
   try {
     const response = await fetch(url, {
       ...options,
       signal: controller.signal
     });
     clearTimeout(timeoutId);
     return response;
   } catch (error) {
     clearTimeout(timeoutId);
     throw error;
   }
   ```

## API-Specific Error Codes

### Authentication Errors

#### AUTHENTICATION_ERROR
**Code:** `AUTHENTICATION_ERROR`
**HTTP Status:** 401
**Description:** Authentication failed

**Sub-codes:**
- `INVALID_API_KEY`: API key is invalid or revoked
- `EXPIRED_TOKEN`: JWT token has expired
- `MISSING_AUTHENTICATION`: No authentication provided
- `INVALID_SIGNATURE`: Webhook signature verification failed

**Solutions:**
```javascript
function handleAuthenticationError(error) {
  switch (error.sub_code) {
    case 'INVALID_API_KEY':
      console.log('Check API key format and validity');
      break;
    case 'EXPIRED_TOKEN':
      console.log('Refresh JWT token');
      break;
    case 'MISSING_AUTHENTICATION':
      console.log('Add Authorization header');
      break;
    case 'INVALID_SIGNATURE':
      console.log('Verify webhook signature');
      break;
  }
}
```

### Validation Errors

#### VALIDATION_ERROR
**Code:** `VALIDATION_ERROR`
**HTTP Status:** 400
**Description:** Request validation failed

**Sub-codes:**
- `INVALID_BUSINESS_NAME`: Business name is invalid
- `INVALID_ADDRESS`: Business address is invalid
- `INVALID_INDUSTRY`: Industry is not supported
- `INVALID_COUNTRY`: Country code is invalid
- `INVALID_EMAIL`: Email format is invalid
- `INVALID_PHONE`: Phone format is invalid

**Solutions:**
```javascript
function handleValidationError(error) {
  if (error.validation) {
    error.validation.forEach(validationError => {
      switch (validationError.code) {
        case 'INVALID_BUSINESS_NAME':
          console.log('Business name must be 3-255 characters');
          break;
        case 'INVALID_ADDRESS':
          console.log('Address must be 10-500 characters');
          break;
        case 'INVALID_INDUSTRY':
          console.log('Industry must be one of: Technology, Healthcare, Finance, etc.');
          break;
        case 'INVALID_COUNTRY':
          console.log('Country must be a valid ISO 3166-1 alpha-2 code');
          break;
      }
    });
  }
}
```

### Model Errors

#### MODEL_ERROR
**Code:** `MODEL_ERROR`
**HTTP Status:** 500
**Description:** ML model error

**Sub-codes:**
- `MODEL_LOAD_FAILED`: Model failed to load
- `PREDICTION_FAILED`: Model prediction failed
- `INSUFFICIENT_DATA`: Not enough data for prediction
- `MODEL_TIMEOUT`: Model prediction timeout

**Solutions:**
```javascript
function handleModelError(error) {
  switch (error.sub_code) {
    case 'MODEL_LOAD_FAILED':
      console.log('Model service is temporarily unavailable');
      break;
    case 'PREDICTION_FAILED':
      console.log('Try with different model or parameters');
      break;
    case 'INSUFFICIENT_DATA':
      console.log('Provide more business data for accurate prediction');
      break;
    case 'MODEL_TIMEOUT':
      console.log('Model prediction took too long, try again');
      break;
  }
}
```

### External API Errors

#### EXTERNAL_API_ERROR
**Code:** `EXTERNAL_API_ERROR`
**HTTP Status:** 502
**Description:** External API error

**Sub-codes:**
- `EXTERNAL_API_TIMEOUT`: External API timeout
- `EXTERNAL_API_RATE_LIMIT`: External API rate limit
- `EXTERNAL_API_UNAVAILABLE`: External API unavailable
- `EXTERNAL_API_INVALID_RESPONSE`: Invalid response from external API

**Solutions:**
```javascript
function handleExternalAPIError(error) {
  switch (error.sub_code) {
    case 'EXTERNAL_API_TIMEOUT':
      console.log('External API timeout, retry later');
      break;
    case 'EXTERNAL_API_RATE_LIMIT':
      console.log('External API rate limited, implement queuing');
      break;
    case 'EXTERNAL_API_UNAVAILABLE':
      console.log('External API unavailable, use fallback');
      break;
    case 'EXTERNAL_API_INVALID_RESPONSE':
      console.log('External API returned invalid response');
      break;
  }
}
```

### Database Errors

#### DATABASE_ERROR
**Code:** `DATABASE_ERROR`
**HTTP Status:** 500
**Description:** Database error

**Sub-codes:**
- `DATABASE_CONNECTION_FAILED`: Database connection failed
- `DATABASE_QUERY_TIMEOUT`: Database query timeout
- `DATABASE_CONSTRAINT_VIOLATION`: Database constraint violation
- `DATABASE_TRANSACTION_FAILED`: Database transaction failed

**Solutions:**
```javascript
function handleDatabaseError(error) {
  switch (error.sub_code) {
    case 'DATABASE_CONNECTION_FAILED':
      console.log('Database connection failed, retry later');
      break;
    case 'DATABASE_QUERY_TIMEOUT':
      console.log('Database query timeout, optimize query');
      break;
    case 'DATABASE_CONSTRAINT_VIOLATION':
      console.log('Database constraint violation, check data');
      break;
    case 'DATABASE_TRANSACTION_FAILED':
      console.log('Database transaction failed, retry');
      break;
  }
}
```

## Error Handling Best Practices

### 1. Comprehensive Error Handling

```javascript
async function makeAPICall(url, options) {
  try {
    const response = await fetch(url, options);
    
    if (!response.ok) {
      const errorData = await response.json();
      throw new APIError(errorData, response.status);
    }
    
    return await response.json();
  } catch (error) {
    if (error instanceof APIError) {
      handleAPIError(error);
    } else {
      handleNetworkError(error);
    }
    throw error;
  }
}

class APIError extends Error {
  constructor(errorData, status) {
    super(errorData.message);
    this.name = 'APIError';
    this.code = errorData.code;
    this.details = errorData.details;
    this.status = status;
    this.requestId = errorData.request_id;
  }
}
```

### 2. Retry Logic

```javascript
async function retryAPICall(apiCall, maxRetries = 3) {
  for (let attempt = 0; attempt < maxRetries; attempt++) {
    try {
      return await apiCall();
    } catch (error) {
      if (shouldRetry(error) && attempt < maxRetries - 1) {
        const waitTime = calculateWaitTime(attempt, error);
        console.log(`Retrying in ${waitTime}ms... (${attempt + 1}/${maxRetries})`);
        await new Promise(resolve => setTimeout(resolve, waitTime));
        continue;
      }
      throw error;
    }
  }
}

function shouldRetry(error) {
  // Retry on server errors and rate limits
  return error.status >= 500 || error.status === 429;
}

function calculateWaitTime(attempt, error) {
  if (error.status === 429) {
    // Use Retry-After header if available
    return error.retryAfter ? error.retryAfter * 1000 : Math.pow(2, attempt) * 1000;
  }
  // Exponential backoff for server errors
  return Math.pow(2, attempt) * 1000;
}
```

### 3. Error Logging

```javascript
function logError(error, context = {}) {
  const errorLog = {
    timestamp: new Date().toISOString(),
    error: {
      name: error.name,
      message: error.message,
      code: error.code,
      status: error.status,
      stack: error.stack
    },
    context: context,
    requestId: error.requestId
  };
  
  console.error('API Error:', JSON.stringify(errorLog, null, 2));
  
  // Send to error tracking service
  if (process.env.NODE_ENV === 'production') {
    sendToErrorTracking(errorLog);
  }
}
```

### 4. User-Friendly Error Messages

```javascript
function getUserFriendlyMessage(error) {
  switch (error.code) {
    case 'AUTHENTICATION_ERROR':
      return 'Please check your API key and try again.';
    case 'VALIDATION_ERROR':
      return 'Please check your input data and try again.';
    case 'RATE_LIMIT_EXCEEDED':
      return 'Too many requests. Please wait a moment and try again.';
    case 'SERVICE_UNAVAILABLE':
      return 'Service is temporarily unavailable. Please try again later.';
    default:
      return 'An unexpected error occurred. Please try again or contact support.';
  }
}
```

## Error Monitoring and Alerting

### 1. Error Rate Monitoring

```javascript
class ErrorMonitor {
  constructor() {
    this.errorCounts = new Map();
    this.errorRates = new Map();
  }
  
  recordError(error) {
    const key = `${error.code}_${error.status}`;
    this.errorCounts.set(key, (this.errorCounts.get(key) || 0) + 1);
    
    // Calculate error rate
    const totalRequests = this.getTotalRequests();
    const errorRate = this.errorCounts.get(key) / totalRequests;
    this.errorRates.set(key, errorRate);
    
    // Alert if error rate is high
    if (errorRate > 0.1) { // 10% error rate
      this.sendAlert(key, errorRate);
    }
  }
  
  sendAlert(errorType, errorRate) {
    console.warn(`High error rate detected: ${errorType} (${(errorRate * 100).toFixed(2)}%)`);
    // Send to monitoring service
  }
}
```

### 2. Health Check Monitoring

```javascript
async function monitorAPIHealth() {
  try {
    const response = await fetch('/api/v1/health');
    const health = await response.json();
    
    if (health.status !== 'healthy') {
      console.warn('API health check failed:', health);
      // Send alert
    }
    
    return health;
  } catch (error) {
    console.error('Health check failed:', error);
    // Send critical alert
    throw error;
  }
}

// Run health check every 5 minutes
setInterval(monitorAPIHealth, 5 * 60 * 1000);
```

## Support and Resources

### Error Reporting

When reporting errors, include:
- Error code and message
- Request ID (if available)
- Request details (URL, method, headers)
- Response details (status, headers, body)
- Timestamp
- Steps to reproduce

### Support Channels

- **Email**: [api-support@kyb-platform.com](mailto:api-support@kyb-platform.com)
- **GitHub Issues**: [https://github.com/kyb-platform/risk-assessment-service/issues](https://github.com/kyb-platform/risk-assessment-service/issues)
- **Status Page**: [https://status.kyb-platform.com](https://status.kyb-platform.com)
- **Documentation**: [https://docs.kyb-platform.com](https://docs.kyb-platform.com)

---

**Last Updated**: January 15, 2024  
**Version**: 2.0.0  
**Next Review**: April 15, 2024
