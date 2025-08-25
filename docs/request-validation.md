# Request Validation Middleware

## Overview

The Request Validation Middleware provides comprehensive validation capabilities for HTTP requests in the Enhanced Business Intelligence System. It validates content types, request body sizes, query parameters, and request body content against configurable schemas, while providing input sanitization and security features.

## Features

### Core Validation Features
- **Content Type Validation**: Validates request content types against allowed types
- **Body Size Validation**: Prevents oversized requests to prevent DoS attacks
- **Query Parameter Validation**: Validates query parameter count, size, and content
- **JSON Schema Validation**: Validates request body against configurable schemas
- **Form Data Validation**: Supports form and multipart form data validation
- **Path-based Validation**: Different validation rules for different endpoints

### Security Features
- **Input Sanitization**: Automatic HTML entity encoding and whitespace trimming
- **Injection Prevention**: Detects and blocks SQL injection and XSS patterns
- **Sensitive Data Masking**: Masks sensitive data in error responses
- **Request Timeout**: Configurable validation timeout to prevent hanging requests

### Performance Features
- **Schema Caching**: Caches validation schemas for improved performance
- **Early Termination**: Option to stop validation on first error
- **Efficient Parsing**: Optimized JSON and form data parsing
- **Memory Management**: Proper body restoration for subsequent handlers

### Integration Features
- **Error Handling Integration**: Seamless integration with error handling middleware
- **Context Integration**: Provides validated and sanitized data through request context
- **Logging Integration**: Comprehensive logging of validation failures
- **Metrics Support**: Ready for integration with monitoring systems

## Configuration

### Basic Configuration

```go
config := &RequestValidationConfig{
    Enabled:              true,
    MaxBodySize:          10 * 1024 * 1024, // 10MB
    MaxQueryParams:       50,
    MaxQueryValueSize:    1000,
    AllowedContentTypes:  []string{"application/json", "application/x-www-form-urlencoded"},
    RequireContentType:   false,
    SanitizeInputs:       true,
    PreventInjection:     true,
    LogValidationFailures: true,
    CacheSchemas:         true,
    ValidationTimeout:    5 * time.Second,
    DetailedErrors:       true,
    StopOnFirstError:     false,
}
```

### Predefined Configurations

#### Default Configuration
```go
config := GetDefaultRequestValidationConfig()
```
- **Max Body Size**: 10MB
- **Max Query Params**: 50
- **Content Types**: JSON, form, multipart
- **Security**: Enabled (sanitization, injection prevention)
- **Performance**: Schema caching enabled
- **Timeout**: 5 seconds

#### Strict Configuration
```go
config := GetStrictRequestValidationConfig()
```
- **Max Body Size**: 1MB
- **Max Query Params**: 20
- **Content Types**: JSON only
- **Security**: Maximum security settings
- **Performance**: Fast validation with early termination
- **Timeout**: 3 seconds

#### Permissive Configuration
```go
config := GetPermissiveRequestValidationConfig()
```
- **Max Body Size**: 50MB
- **Max Query Params**: 100
- **Content Types**: All common types including text/plain
- **Security**: Minimal restrictions
- **Performance**: No caching, detailed errors
- **Timeout**: 10 seconds

## Usage

### Basic Usage

```go
package main

import (
    "net/http"
    "go.uber.org/zap"
    "github.com/your-project/internal/api/middleware"
)

func main() {
    logger := zap.NewProduction()
    
    // Create validation middleware
    validationConfig := middleware.GetDefaultRequestValidationConfig()
    validationMiddleware := middleware.NewRequestValidationMiddleware(validationConfig, logger)
    
    // Create your handler
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Get validated data from context
        if validatedData := middleware.GetValidatedData(r.Context()); validatedData != nil {
            // Use validated data
            data := validatedData.(map[string]interface{})
            // Process data...
        }
        
        // Get sanitized data from context
        if sanitizedData := middleware.GetSanitizedData(r.Context()); sanitizedData != nil {
            // Use sanitized data
            // Process sanitized data...
        }
        
        w.WriteHeader(http.StatusOK)
    })
    
    // Apply middleware
    finalHandler := validationMiddleware.Middleware(handler)
    
    // Start server
    http.ListenAndServe(":8080", finalHandler)
}
```

### Advanced Usage with Schema Validation

```go
// Define validation schema
schema := &middleware.ValidationSchema{
    Required: []string{"name", "email"},
    Fields: map[string]middleware.ValidationRule{
        "name": {
            Type:        "string",
            Required:    true,
            MinLength:   2,
            MaxLength:   100,
            Description: "User's full name",
        },
        "email": {
            Type:        "string",
            Required:    true,
            Pattern:     `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
            Description: "Valid email address",
        },
        "age": {
            Type:        "number",
            MinValue:    18,
            MaxValue:    120,
            Description: "User's age (18-120)",
        },
        "status": {
            Type:        "string",
            Enum:        []string{"active", "inactive", "pending"},
            Description: "User account status",
        },
    },
    AllowExtra: false,
}

// Configure path-based validation
config := middleware.GetDefaultRequestValidationConfig()
config.PathRules["/api/users"] = schema

// Create middleware
validationMiddleware := middleware.NewRequestValidationMiddleware(config, logger)
```

### Integration with Error Handling Middleware

```go
// Create error handling middleware
errorConfig := middleware.GetDefaultErrorHandlingConfig()
errorMiddleware := middleware.NewErrorHandlingMiddleware(errorConfig, logger)

// Create validation middleware
validationConfig := middleware.GetDefaultRequestValidationConfig()
validationMiddleware := middleware.NewRequestValidationMiddleware(validationConfig, logger)

// Create your handler
handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // Your handler logic
    w.WriteHeader(http.StatusOK)
})

// Apply middleware in order
finalHandler := errorMiddleware.Middleware(
    validationMiddleware.Middleware(handler),
)

http.ListenAndServe(":8080", finalHandler)
```

## Validation Rules

### Field Validation Rules

```go
type ValidationRule struct {
    Field       string                    // Field name
    Type        string                    // Expected type (string, number, boolean, array, object)
    Required    bool                      // Whether field is required
    MinLength   int                       // Minimum length for strings
    MaxLength   int                       // Maximum length for strings
    MinValue    float64                   // Minimum value for numbers
    MaxValue    float64                   // Maximum value for numbers
    Pattern     string                    // Regex pattern for validation
    Enum        []string                  // Allowed values for enums
    CustomFunc  func(interface{}) error   // Custom validation function
    Description string                    // Human-readable description
}
```

### Supported Types

#### String Validation
```go
{
    Type:        "string",
    Required:    true,
    MinLength:   3,
    MaxLength:   50,
    Pattern:     `^[a-zA-Z0-9_]+$`,
    Enum:        []string{"admin", "user", "guest"},
}
```

#### Number Validation
```go
{
    Type:     "number",
    Required: true,
    MinValue: 0,
    MaxValue: 100,
}
```

#### Boolean Validation
```go
{
    Type:     "boolean",
    Required: true,
}
```

#### Array Validation
```go
{
    Type:     "array",
    Required: true,
}
```

#### Object Validation
```go
{
    Type:     "object",
    Required: true,
}
```

### Custom Validation Functions

```go
// Custom validation function
func validateEmail(value interface{}) error {
    if str, ok := value.(string); ok {
        if !strings.Contains(str, "@") {
            return fmt.Errorf("invalid email format")
        }
    }
    return nil
}

// Use in validation rule
rule := middleware.ValidationRule{
    Type:       "string",
    Required:   true,
    CustomFunc: validateEmail,
}
```

## Error Handling

### Validation Error Structure

```go
type RequestValidationError struct {
    Field   string      `json:"field"`   // Field name that failed validation
    Value   interface{} `json:"value"`   // Value that failed validation (may be masked)
    Rule    string      `json:"rule"`    // Rule that failed (required, type, pattern, etc.)
    Message string      `json:"message"` // Human-readable error message
}
```

### Error Response Format

```json
{
    "error": "Request validation failed",
    "errors": [
        {
            "field": "email",
            "value": "invalid-email",
            "rule": "pattern",
            "message": "must match pattern: ^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
        },
        {
            "field": "age",
            "value": 15,
            "rule": "min_value",
            "message": "minimum value is 18"
        }
    ],
    "success": false
}
```

### Integration with Error Handling Middleware

The request validation middleware integrates seamlessly with the error handling middleware:

```go
// Validation errors are automatically converted to CustomError types
validationErr := middleware.CreateValidationError("Request validation failed", "3 validation errors")

// Error handling middleware processes these consistently
errorMiddleware.handleErrorResponse(w, r, http.StatusBadRequest, validationErr)
```

## Security Features

### Input Sanitization

The middleware automatically sanitizes input data:

```go
// Input: "  <script>alert('xss')</script>  "
// Output: "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;"
```

### Injection Pattern Detection

Detects and blocks common injection patterns:

- SQL Injection: `SELECT`, `INSERT`, `UPDATE`, `DELETE`, etc.
- XSS: `<script>`, `javascript:`, `vbscript:`
- Command Injection: `exec`, `execute`, etc.

### Sensitive Data Masking

Sensitive data is masked in error responses:

```go
// Original: "Bearer secret-token-123"
// Masked: "[MASKED]"
```

## Performance Considerations

### Schema Caching

Enable schema caching for improved performance:

```go
config := &RequestValidationConfig{
    CacheSchemas: true,
}
```

### Validation Timeout

Set appropriate timeouts to prevent hanging requests:

```go
config := &RequestValidationConfig{
    ValidationTimeout: 5 * time.Second,
}
```

### Early Termination

Stop validation on first error for faster responses:

```go
config := &RequestValidationConfig{
    StopOnFirstError: true,
}
```

## Monitoring and Observability

### Logging

Validation failures are logged with structured data:

```go
logger.Warn("Request validation failed",
    zap.String("path", r.URL.Path),
    zap.String("method", r.Method),
    zap.Int("error_count", len(result.Errors)),
    zap.Any("errors", result.Errors),
)
```

### Metrics Integration

Ready for integration with monitoring systems:

```go
// Example metrics that could be collected:
// - validation_requests_total
// - validation_errors_total
// - validation_duration_seconds
// - validation_errors_by_type
// - validation_errors_by_field
```

## Testing

### Unit Testing

```go
func TestRequestValidation(t *testing.T) {
    // Create test configuration
    config := &RequestValidationConfig{
        Enabled:     true,
        MaxBodySize: 1024,
    }
    
    // Create middleware
    logger := zap.NewNop()
    middleware := NewRequestValidationMiddleware(config, logger)
    
    // Test request
    req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"name": "test"}`))
    req.Header.Set("Content-Type", "application/json")
    
    // Validate request
    result, err := middleware.ValidateRequest(req)
    
    assert.NoError(t, err)
    assert.True(t, result.Valid)
}
```

### Integration Testing

```go
func TestValidationMiddlewareIntegration(t *testing.T) {
    // Create middleware stack
    validationMiddleware := NewRequestValidationMiddleware(config, logger)
    errorMiddleware := NewErrorHandlingMiddleware(errorConfig, logger)
    
    // Create test handler
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Get validated data
        data := GetValidatedData(r.Context())
        assert.NotNil(t, data)
        
        w.WriteHeader(http.StatusOK)
    })
    
    // Apply middleware
    finalHandler := errorMiddleware.Middleware(
        validationMiddleware.Middleware(handler),
    )
    
    // Test request
    req := httptest.NewRequest("POST", "/test", strings.NewReader(`{"name": "test"}`))
    req.Header.Set("Content-Type", "application/json")
    
    rr := httptest.NewRecorder()
    finalHandler.ServeHTTP(rr, req)
    
    assert.Equal(t, http.StatusOK, rr.Code)
}
```

## Best Practices

### Configuration Best Practices

1. **Environment-Specific Configurations**
   ```go
   // Development
   config := GetPermissiveRequestValidationConfig()
   
   // Production
   config := GetStrictRequestValidationConfig()
   
   // Staging
   config := GetDefaultRequestValidationConfig()
   ```

2. **Path-Specific Validation**
   ```go
   // Different validation rules for different endpoints
   config.PathRules["/api/users"] = userSchema
   config.PathRules["/api/admin"] = adminSchema
   config.PathRules["/api/public"] = publicSchema
   ```

3. **Security-First Approach**
   ```go
   config := &RequestValidationConfig{
       SanitizeInputs:    true,
       PreventInjection:  true,
       LogValidationFailures: true,
       ValidationTimeout: 5 * time.Second,
   }
   ```

### Schema Design Best Practices

1. **Clear Field Descriptions**
   ```go
   rule := ValidationRule{
       Type:        "string",
       Required:    true,
       MinLength:   3,
       MaxLength:   50,
       Description: "User's full name (3-50 characters)",
   }
   ```

2. **Appropriate Validation Rules**
   ```go
   // Email validation
   emailRule := ValidationRule{
       Type:        "string",
       Required:    true,
       Pattern:     `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
       Description: "Valid email address",
   }
   
   // Age validation
   ageRule := ValidationRule{
       Type:        "number",
       MinValue:    18,
       MaxValue:    120,
       Description: "User's age (18-120)",
   }
   ```

3. **Custom Validation Functions**
   ```go
   func validateBusinessLogic(value interface{}) error {
       // Complex business logic validation
       return nil
   }
   ```

### Error Handling Best Practices

1. **Consistent Error Messages**
   ```go
   rule := ValidationRule{
       Type:        "string",
       Required:    true,
       Description: "User's full name is required",
   }
   ```

2. **Appropriate Error Detail Level**
   ```go
   // Development
   config.DetailedErrors = true
   
   // Production
   config.DetailedErrors = false
   ```

3. **Error Logging**
   ```go
   config.LogValidationFailures = true
   ```

## Troubleshooting

### Common Issues

1. **Body Not Restored**
   - Ensure middleware is applied before handlers that read the body
   - The middleware automatically restores the body for subsequent handlers

2. **Validation Timeouts**
   - Increase `ValidationTimeout` for complex validation rules
   - Consider optimizing custom validation functions

3. **Schema Cache Issues**
   - Clear schema cache if validation rules change
   - Monitor cache memory usage for large schemas

4. **Performance Issues**
   - Enable schema caching
   - Use early termination for simple validations
   - Optimize custom validation functions

### Debug Mode

Enable detailed logging for debugging:

```go
logger := zap.NewDevelopment()
config := &RequestValidationConfig{
    LogValidationFailures: true,
    DetailedErrors:       true,
}
```

## Migration Guide

### From Basic Validation

If you're migrating from basic validation:

1. **Replace Basic Checks**
   ```go
   // Old way
   if r.ContentLength > maxSize {
       http.Error(w, "Body too large", http.StatusBadRequest)
       return
   }
   
   // New way
   validationMiddleware := NewRequestValidationMiddleware(config, logger)
   handler := validationMiddleware.Middleware(yourHandler)
   ```

2. **Replace Manual JSON Validation**
   ```go
   // Old way
   var data map[string]interface{}
   if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
       http.Error(w, "Invalid JSON", http.StatusBadRequest)
       return
   }
   
   // New way
   if data := GetValidatedData(r.Context()); data != nil {
       // Use validated data
   }
   ```

3. **Replace Manual Sanitization**
   ```go
   // Old way
   name := html.EscapeString(strings.TrimSpace(input))
   
   // New way
   if sanitized := GetSanitizedData(r.Context()); sanitized != nil {
       name := sanitized["name"].(string)
   }
   ```

### Integration with Existing Middleware

1. **Order of Middleware**
   ```go
   // Recommended order
   finalHandler := errorMiddleware.Middleware(
       validationMiddleware.Middleware(
           loggingMiddleware.Middleware(
               yourHandler,
           ),
       ),
   )
   ```

2. **Context Integration**
   ```go
   // Access validated data in handlers
   func yourHandler(w http.ResponseWriter, r *http.Request) {
       if data := GetValidatedData(r.Context()); data != nil {
           // Process validated data
       }
   }
   ```

## API Reference

### Types

#### RequestValidationConfig
```go
type RequestValidationConfig struct {
    Enabled              bool
    MaxBodySize          int64
    MaxQueryParams       int
    MaxQueryValueSize    int
    AllowedContentTypes  []string
    RequireContentType   bool
    PathRules            map[string]*ValidationSchema
    SanitizeInputs       bool
    PreventInjection     bool
    LogValidationFailures bool
    CacheSchemas         bool
    ValidationTimeout    time.Duration
    DetailedErrors       bool
    StopOnFirstError     bool
}
```

#### ValidationSchema
```go
type ValidationSchema struct {
    Fields     map[string]ValidationRule
    Required   []string
    MaxSize    int64
    AllowExtra bool
}
```

#### ValidationRule
```go
type ValidationRule struct {
    Field       string
    Type        string
    Required    bool
    MinLength   int
    MaxLength   int
    MinValue    float64
    MaxValue    float64
    Pattern     string
    Enum        []string
    CustomFunc  func(interface{}) error
    Description string
}
```

### Functions

#### NewRequestValidationMiddleware
```go
func NewRequestValidationMiddleware(config *RequestValidationConfig, logger *zap.Logger) *RequestValidationMiddleware
```

#### GetValidatedData
```go
func GetValidatedData(ctx context.Context) interface{}
```

#### GetSanitizedData
```go
func GetSanitizedData(ctx context.Context) map[string]interface{}
```

#### Configuration Functions
```go
func GetDefaultRequestValidationConfig() *RequestValidationConfig
func GetStrictRequestValidationConfig() *RequestValidationConfig
func GetPermissiveRequestValidationConfig() *RequestValidationConfig
```

This comprehensive request validation middleware provides robust validation capabilities while maintaining security, performance, and ease of use. It integrates seamlessly with the existing middleware stack and provides the foundation for reliable request validation across the Enhanced Business Intelligence System.
