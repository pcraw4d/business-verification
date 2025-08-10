# KYB Platform API - Error Codes and Responses

This document provides comprehensive information about error codes, response formats, and error handling best practices for the KYB Platform API.

## Table of Contents

1. [Error Response Format](#error-response-format)
2. [HTTP Status Codes](#http-status-codes)
3. [Error Types](#error-types)
4. [Authentication Errors](#authentication-errors)
5. [Validation Errors](#validation-errors)
6. [Rate Limiting](#rate-limiting)
7. [Business Logic Errors](#business-logic-errors)
8. [Server Errors](#server-errors)
9. [Error Handling Best Practices](#error-handling-best-practices)
10. [Error Recovery Strategies](#error-recovery-strategies)

---

## Error Response Format

All error responses follow a consistent JSON format:

```json
{
  "error": "error_type_identifier",
  "message": "Human-readable error message",
  "status_code": 400,
  "timestamp": "2024-01-15T10:30:00Z",
  "details": {
    "field": "business_name",
    "constraint": "required",
    "value": null
  },
  "retry_after": 60
}
```

### Response Fields

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `error` | string | Machine-readable error identifier | `validation_error` |
| `message` | string | Human-readable error description | `Business name is required` |
| `status_code` | integer | HTTP status code | `400` |
| `timestamp` | string | ISO 8601 timestamp | `2024-01-15T10:30:00Z` |
| `details` | object | Additional error context (optional) | `{"field": "email"}` |
| `retry_after` | integer | Seconds to wait before retry (optional) | `60` |

---

## HTTP Status Codes

### 2xx Success
- **200 OK**: Request successful
- **201 Created**: Resource created successfully
- **204 No Content**: Request successful, no content returned

### 4xx Client Errors
- **400 Bad Request**: Invalid request data or syntax
- **401 Unauthorized**: Authentication required or failed
- **403 Forbidden**: Insufficient permissions
- **404 Not Found**: Resource not found
- **409 Conflict**: Resource conflict (e.g., duplicate email)
- **422 Unprocessable Entity**: Valid request but business logic failed
- **429 Too Many Requests**: Rate limit exceeded

### 5xx Server Errors
- **500 Internal Server Error**: Unexpected server error
- **502 Bad Gateway**: Upstream service unavailable
- **503 Service Unavailable**: Service temporarily unavailable
- **504 Gateway Timeout**: Upstream service timeout

---

## Error Types

### Authentication Errors

#### `unauthorized`
**Status Code**: 401

**Description**: Invalid or missing authentication credentials.

**Common Causes**:
- Missing Authorization header
- Invalid JWT token format
- Expired access token
- Invalid API key

**Example Response**:
```json
{
  "error": "unauthorized",
  "message": "Invalid or expired authentication token",
  "status_code": 401,
  "timestamp": "2024-01-15T10:30:00Z"
}
```

**Resolution**:
- Include valid Authorization header: `Authorization: Bearer <token>`
- Refresh expired tokens using `/v1/auth/refresh`
- Verify token format and validity

#### `insufficient_permissions`
**Status Code**: 403

**Description**: User lacks required permissions for the requested operation.

**Common Causes**:
- User role doesn't have required permissions
- API key has insufficient scope
- Resource access restricted

**Example Response**:
```json
{
  "error": "insufficient_permissions",
  "message": "User does not have permission to access this resource",
  "status_code": 403,
  "timestamp": "2024-01-15T10:30:00Z",
  "details": {
    "required_permission": "admin:users:read",
    "user_permissions": ["user:profile:read"]
  }
}
```

**Resolution**:
- Contact administrator for permission upgrade
- Use API key with appropriate scope
- Verify resource ownership

### Validation Errors

#### `validation_error`
**Status Code**: 400

**Description**: Request data fails validation rules.

**Common Causes**:
- Missing required fields
- Invalid data types
- Field length constraints
- Invalid email format
- Invalid date format

**Example Response**:
```json
{
  "error": "validation_error",
  "message": "Business name is required",
  "status_code": 400,
  "timestamp": "2024-01-15T10:30:00Z",
  "details": {
    "field": "business_name",
    "constraint": "required",
    "value": null
  }
}
```

**Field-Specific Validation Errors**:

**Email Validation**:
```json
{
  "error": "validation_error",
  "message": "Invalid email format",
  "status_code": 400,
  "timestamp": "2024-01-15T10:30:00Z",
  "details": {
    "field": "email",
    "constraint": "email_format",
    "value": "invalid-email"
  }
}
```

**Password Validation**:
```json
{
  "error": "validation_error",
  "message": "Password must be at least 8 characters long",
  "status_code": 400,
  "timestamp": "2024-01-15T10:30:00Z",
  "details": {
    "field": "password",
    "constraint": "min_length",
    "min_length": 8,
    "actual_length": 5
  }
}
```

**Business Name Validation**:
```json
{
  "error": "validation_error",
  "message": "Business name must be between 2 and 200 characters",
  "status_code": 400,
  "timestamp": "2024-01-15T10:30:00Z",
  "details": {
    "field": "business_name",
    "constraint": "length",
    "min_length": 2,
    "max_length": 200,
    "actual_length": 250
  }
}
```

**Resolution**:
- Review field requirements and constraints
- Ensure data types match expected formats
- Validate input before sending requests

#### `invalid_json`
**Status Code**: 400

**Description**: Request body contains invalid JSON syntax.

**Example Response**:
```json
{
  "error": "invalid_json",
  "message": "Invalid JSON syntax in request body",
  "status_code": 400,
  "timestamp": "2024-01-15T10:30:00Z",
  "details": {
    "position": 45,
    "line": 3
  }
}
```

**Resolution**:
- Validate JSON syntax before sending
- Use proper JSON formatting tools
- Check for trailing commas or missing quotes

### Rate Limiting

#### `rate_limit_exceeded`
**Status Code**: 429

**Description**: Request rate exceeds allowed limits.

**Rate Limits**:
- **Authenticated users**: 100 requests per minute
- **Unauthenticated users**: 10 requests per minute
- **Classification endpoints**: 50 requests per minute
- **Risk assessment**: 30 requests per minute

**Example Response**:
```json
{
  "error": "rate_limit_exceeded",
  "message": "Rate limit exceeded. Please try again later.",
  "status_code": 429,
  "timestamp": "2024-01-15T10:30:00Z",
  "retry_after": 60,
  "details": {
    "limit": 100,
    "window": "1 minute",
    "reset_time": "2024-01-15T10:31:00Z"
  }
}
```

**Response Headers**:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1642248000
Retry-After: 60
```

**Resolution**:
- Implement exponential backoff
- Respect `Retry-After` header
- Monitor rate limit headers
- Consider batch operations for multiple requests

### Business Logic Errors

#### `resource_not_found`
**Status Code**: 404

**Description**: Requested resource does not exist.

**Common Scenarios**:
- Business ID not found
- User profile not found
- Classification history not found
- Compliance record not found

**Example Response**:
```json
{
  "error": "resource_not_found",
  "message": "Business with ID 'business_123' not found",
  "status_code": 404,
  "timestamp": "2024-01-15T10:30:00Z",
  "details": {
    "resource_type": "business",
    "resource_id": "business_123"
  }
}
```

**Resolution**:
- Verify resource ID exists
- Check resource ownership
- Use correct resource identifiers

#### `resource_conflict`
**Status Code**: 409

**Description**: Resource conflict prevents operation completion.

**Common Scenarios**:
- Email already registered
- Business name already exists
- Duplicate classification request

**Example Response**:
```json
{
  "error": "resource_conflict",
  "message": "User with email 'user@example.com' already exists",
  "status_code": 409,
  "timestamp": "2024-01-15T10:30:00Z",
  "details": {
    "conflicting_field": "email",
    "conflicting_value": "user@example.com"
  }
}
```

**Resolution**:
- Use unique identifiers
- Check for existing resources before creation
- Handle conflicts gracefully in application logic

#### `classification_failed`
**Status Code**: 422

**Description**: Business classification could not be completed.

**Common Causes**:
- Insufficient business information
- Unrecognized business type
- External service unavailable
- Classification confidence below threshold

**Example Response**:
```json
{
  "error": "classification_failed",
  "message": "Unable to classify business with provided information",
  "status_code": 422,
  "timestamp": "2024-01-15T10:30:00Z",
  "details": {
    "reason": "insufficient_information",
    "confidence_threshold": 0.5,
    "max_confidence": 0.3,
    "suggestions": [
      "Provide business description",
      "Include industry keywords",
      "Specify business type"
    ]
  }
}
```

**Resolution**:
- Provide more business details
- Include industry keywords
- Specify business type and description
- Retry with enhanced information

#### `risk_assessment_failed`
**Status Code**: 422

**Description**: Risk assessment could not be completed.

**Example Response**:
```json
{
  "error": "risk_assessment_failed",
  "message": "Risk assessment failed due to insufficient data",
  "status_code": 422,
  "timestamp": "2024-01-15T10:30:00Z",
  "details": {
    "reason": "insufficient_financial_data",
    "required_fields": ["annual_revenue", "employee_count"],
    "missing_fields": ["annual_revenue"]
  }
}
```

#### `compliance_check_failed`
**Status Code**: 422

**Description**: Compliance check could not be completed.

**Example Response**:
```json
{
  "error": "compliance_check_failed",
  "message": "Compliance check failed for specified frameworks",
  "status_code": 422,
  "timestamp": "2024-01-15T10:30:00Z",
  "details": {
    "reason": "framework_not_supported",
    "requested_frameworks": ["ISO_27001"],
    "supported_frameworks": ["SOC2", "PCI_DSS", "GDPR"]
  }
}
```

### Server Errors

#### `internal_server_error`
**Status Code**: 500

**Description**: Unexpected server error occurred.

**Example Response**:
```json
{
  "error": "internal_server_error",
  "message": "An unexpected error occurred. Please try again later.",
  "status_code": 500,
  "timestamp": "2024-01-15T10:30:00Z",
  "details": {
    "request_id": "req_123456789"
  }
}
```

**Resolution**:
- Retry request after delay
- Contact support if persistent
- Check system status page

#### `service_unavailable`
**Status Code**: 503

**Description**: Service temporarily unavailable.

**Common Causes**:
- Database maintenance
- External service outages
- High system load
- Scheduled maintenance

**Example Response**:
```json
{
  "error": "service_unavailable",
  "message": "Service temporarily unavailable. Please try again later.",
  "status_code": 503,
  "timestamp": "2024-01-15T10:30:00Z",
  "retry_after": 300,
  "details": {
    "maintenance_window": "2024-01-15T10:00:00Z - 2024-01-15T12:00:00Z"
  }
}
```

**Resolution**:
- Wait for service to resume
- Respect `Retry-After` header
- Check status page for updates

#### `gateway_timeout`
**Status Code**: 504

**Description**: Upstream service timeout.

**Example Response**:
```json
{
  "error": "gateway_timeout",
  "message": "Request timeout. Please try again.",
  "status_code": 504,
  "timestamp": "2024-01-15T10:30:00Z",
  "details": {
    "timeout_duration": "30s",
    "upstream_service": "classification_engine"
  }
}
```

**Resolution**:
- Retry request
- Consider reducing request complexity
- Contact support if persistent

---

## Error Handling Best Practices

### 1. Always Check Status Codes

```javascript
const response = await fetch('/v1/classify', {
  method: 'POST',
  headers: { 'Authorization': `Bearer ${token}` },
  body: JSON.stringify(requestData)
});

if (!response.ok) {
  const errorData = await response.json();
  handleError(errorData);
}
```

### 2. Implement Retry Logic

```javascript
async function makeRequestWithRetry(url, options, maxRetries = 3) {
  for (let attempt = 1; attempt <= maxRetries; attempt++) {
    try {
      const response = await fetch(url, options);
      
      if (response.status === 429) {
        const retryAfter = response.headers.get('Retry-After') || 60;
        await new Promise(resolve => setTimeout(resolve, retryAfter * 1000));
        continue;
      }
      
      if (response.status >= 500) {
        if (attempt === maxRetries) throw new Error('Max retries exceeded');
        await new Promise(resolve => setTimeout(resolve, Math.pow(2, attempt) * 1000));
        continue;
      }
      
      return response;
    } catch (error) {
      if (attempt === maxRetries) throw error;
    }
  }
}
```

### 3. Handle Specific Error Types

```javascript
function handleError(errorData) {
  switch (errorData.error) {
    case 'unauthorized':
      refreshToken();
      break;
    case 'validation_error':
      displayValidationErrors(errorData.details);
      break;
    case 'rate_limit_exceeded':
      showRateLimitMessage(errorData.retry_after);
      break;
    case 'resource_not_found':
      showNotFoundMessage();
      break;
    default:
      showGenericErrorMessage(errorData.message);
  }
}
```

### 4. Log Errors Appropriately

```javascript
function logError(errorData, context) {
  const logEntry = {
    timestamp: new Date().toISOString(),
    error: errorData.error,
    message: errorData.message,
    statusCode: errorData.status_code,
    context: context,
    requestId: errorData.details?.request_id
  };
  
  if (errorData.status_code >= 500) {
    console.error('Server error:', logEntry);
  } else {
    console.warn('Client error:', logEntry);
  }
}
```

---

## Error Recovery Strategies

### Authentication Errors
1. **Token Expired**: Use refresh token to get new access token
2. **Invalid Token**: Redirect to login page
3. **Insufficient Permissions**: Show appropriate error message

### Validation Errors
1. **Missing Fields**: Highlight required fields in UI
2. **Invalid Format**: Show field-specific error messages
3. **Length Constraints**: Display character limits

### Rate Limiting
1. **Implement Exponential Backoff**: Wait progressively longer between retries
2. **Show User Feedback**: Display rate limit status to users
3. **Batch Operations**: Combine multiple requests where possible

### Server Errors
1. **Retry with Backoff**: Implement intelligent retry logic
2. **Graceful Degradation**: Show cached data or fallback content
3. **User Communication**: Inform users of temporary issues

### Network Errors
1. **Connection Timeout**: Retry with increased timeout
2. **Network Unavailable**: Show offline mode if available
3. **Service Unavailable**: Direct users to status page

---

## Monitoring and Alerting

### Error Metrics to Track
- **Error Rate**: Percentage of requests resulting in errors
- **Error Distribution**: Breakdown by error type and status code
- **Response Time**: Impact of errors on performance
- **User Impact**: Number of users affected by errors

### Alerting Thresholds
- **Error Rate > 5%**: Investigate immediately
- **5xx Errors > 1%**: Critical alert
- **Rate Limit Hits > 10%**: Review rate limiting strategy
- **Authentication Failures > 20%**: Check token management

### Error Reporting
- **Request ID**: Include in all error responses for tracking
- **User Context**: Log user ID and session information
- **Request Details**: Capture request parameters and headers
- **Stack Traces**: Include for server errors (development only)

---

## Support and Troubleshooting

### Getting Help
- **Documentation**: Check this error reference guide
- **Status Page**: Monitor service health at `/health`
- **Support Email**: Contact support@kybplatform.com
- **Request ID**: Include request ID when reporting issues

### Common Issues
1. **Authentication Problems**: Verify token format and expiration
2. **Rate Limiting**: Implement proper retry logic
3. **Validation Errors**: Review API documentation for field requirements
4. **Server Errors**: Check system status and retry after delay

### Debugging Tips
1. **Enable Logging**: Set appropriate log levels for debugging
2. **Request Tracing**: Use request IDs to trace issues
3. **Error Context**: Capture relevant request and response data
4. **Reproduction Steps**: Document exact steps to reproduce errors
