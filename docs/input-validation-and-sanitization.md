# Input Validation and Sanitization System

## Overview

The Enhanced Business Intelligence System includes a comprehensive input validation and sanitization system designed to ensure data integrity, security, and compliance with business rules. This system provides both validation (checking if data meets requirements) and sanitization (cleaning and normalizing data) capabilities.

## Architecture

The system consists of two main components:

1. **Validation Middleware** (`internal/api/middleware/validation.go`)
   - HTTP request validation
   - Struct-based validation using tags
   - Security validation (SQL injection, XSS, path traversal)
   - Performance monitoring and logging

2. **Sanitization Utility** (`pkg/sanitizer/sanitizer.go`)
   - Input cleaning and normalization
   - Type-specific sanitization (email, URL, phone, UUID)
   - HTML and content sanitization
   - Secure token generation

## Features

### Validation Features

#### Basic Validation Rules
- **required**: Field must be present and not empty
- **email**: Valid email format validation
- **url**: Valid URL format validation
- **phone**: E.164 phone number format validation
- **uuid**: UUID format validation
- **min/max**: Numeric and string length validation
- **len**: Exact length validation
- **regex**: Custom regular expression validation

#### Security Validation Rules
- **sql_injection**: Detects SQL injection patterns
- **xss**: Detects cross-site scripting patterns
- **path_traversal**: Detects path traversal attempts
- **alphanumeric**: Ensures only alphanumeric characters
- **numeric**: Ensures only numeric characters
- **date**: Validates date formats

#### Advanced Features
- **Struct tag-based validation**: Use `validate` tags on struct fields
- **Custom error messages**: Detailed error reporting with field context
- **Severity levels**: Errors and warnings with different handling
- **Performance monitoring**: Validation duration tracking
- **Configurable limits**: Request size, field length, etc.

### Sanitization Features

#### General Sanitization
- **Null byte removal**: Removes null bytes from strings
- **Unicode normalization**: Normalizes Unicode characters
- **Whitespace trimming**: Removes leading/trailing whitespace
- **Line ending normalization**: Standardizes line endings
- **Length limiting**: Enforces maximum length constraints
- **Control character removal**: Removes non-printable characters

#### Type-Specific Sanitization
- **Email sanitization**: Cleans and validates email addresses
- **URL sanitization**: Cleans and validates URLs with protocol checking
- **Phone sanitization**: Cleans and validates E.164 phone numbers
- **UUID sanitization**: Cleans and validates UUIDs
- **HTML sanitization**: Removes dangerous HTML tags and attributes
- **Filename sanitization**: Safe filename generation
- **SQL sanitization**: Basic SQL injection pattern removal

#### Security Features
- **Script removal**: Removes JavaScript and other script content
- **Dangerous protocol detection**: Blocks javascript:, vbscript:, data: URLs
- **Path traversal prevention**: Removes directory traversal attempts
- **Secure token generation**: Cryptographically secure random tokens

## Usage Examples

### Basic Validation

```go
package main

import (
    "net/http"
    "github.com/your-org/kyb-platform/internal/api/middleware"
    "go.uber.org/zap"
)

// Define a struct with validation tags
type UserRegistration struct {
    ID          string  `json:"id" validate:"required,uuid"`
    Name        string  `json:"name" validate:"required,min=2,max=100"`
    Email       string  `json:"email" validate:"required,email"`
    Phone       string  `json:"phone" validate:"phone"`
    Age         int     `json:"age" validate:"min=18,max=120"`
    Website     string  `json:"website" validate:"url"`
    Description string  `json:"description" validate:"max=1000"`
}

func main() {
    logger := zap.NewProduction()
    
    // Create validator with configuration
    config := &middleware.ValidationConfig{
        MaxRequestSize:    10 * 1024 * 1024, // 10MB
        EnableSanitization: true,
        StrictMode:        false,
        LogValidationErrors: true,
    }
    
    validator := middleware.NewValidator(config, logger)
    
    // Use validation middleware
    handler := validator.ValidationMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        var user UserRegistration
        
        // Validate request
        result := validator.ValidateRequest(r, &user)
        if !result.IsValid {
            // Handle validation errors
            w.WriteHeader(http.StatusBadRequest)
            json.NewEncoder(w).Encode(result.Errors)
            return
        }
        
        // Process valid data
        // ...
    }))
    
    http.ListenAndServe(":8080", handler)
}
```

### Advanced Validation

```go
// Struct with comprehensive validation
type BusinessVerification struct {
    ID              string    `json:"id" validate:"required,uuid"`
    BusinessName    string    `json:"business_name" validate:"required,min=2,max=255"`
    Email           string    `json:"email" validate:"required,email"`
    Phone           string    `json:"phone" validate:"phone"`
    Website         string    `json:"website" validate:"url"`
    Address         string    `json:"address" validate:"required,min=10,max=500"`
    Industry        string    `json:"industry" validate:"required,alphanumeric"`
    TaxID           string    `json:"tax_id" validate:"numeric"`
    FoundedDate     string    `json:"founded_date" validate:"date"`
    FilePath        string    `json:"file_path" validate:"path_traversal"`
    Description     string    `json:"description" validate:"max=2000,xss"`
    SQLQuery        string    `json:"sql_query" validate:"sql_injection"`
}

// Custom validation handler
func handleBusinessVerification(w http.ResponseWriter, r *http.Request) {
    var business BusinessVerification
    
    result := validator.ValidateRequest(r, &business)
    if !result.IsValid {
        // Log validation errors
        logger.Error("validation failed",
            zap.Any("errors", result.Errors),
            zap.Any("warnings", result.Warnings),
            zap.Duration("duration", result.Duration))
        
        // Return detailed error response
        response := map[string]interface{}{
            "success": false,
            "errors":  result.Errors,
            "warnings": result.Warnings,
        }
        
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(response)
        return
    }
    
    // Process valid business verification
    // ...
}
```

### Sanitization Usage

```go
package main

import (
    "github.com/your-org/kyb-platform/pkg/sanitizer"
)

func main() {
    // Create sanitizer
    sanitizer := sanitizer.NewSanitizer()
    
    // Basic string sanitization
    dirtyInput := "  Hello <script>alert('xss')</script> World\x00  "
    cleanInput := sanitizer.SanitizeString(dirtyInput, nil)
    // Result: "Hello World"
    
    // Email sanitization
    dirtyEmail := "<script>user@example.com</script>"
    cleanEmail, err := sanitizer.SanitizeEmail(dirtyEmail)
    if err != nil {
        // Handle error
    }
    // Result: "user@example.com"
    
    // URL sanitization
    dirtyURL := "javascript:alert('xss')"
    cleanURL, err := sanitizer.SanitizeURL(dirtyURL)
    if err != nil {
        // Handle error - dangerous protocol detected
    }
    
    // HTML sanitization
    dirtyHTML := "<p>Hello <script>alert('xss')</script> <strong>World</strong></p>"
    cleanHTML := sanitizer.SanitizeHTML(dirtyHTML)
    // Result: "<p>Hello  <strong>World</strong></p>"
    
    // Filename sanitization
    dirtyFilename := "../../../etc/passwd"
    cleanFilename := sanitizer.SanitizeFilename(dirtyFilename)
    // Result: "etcpasswd"
    
    // Generate secure token
    token, err := sanitizer.GenerateSecureToken(32)
    if err != nil {
        // Handle error
    }
    // Result: cryptographically secure 32-character token
}
```

### Custom Configuration

```go
// Custom sanitization configuration
config := &sanitizer.SanitizationConfig{
    RemoveHTMLTags:     true,
    RemoveScripts:      true,
    NormalizeUnicode:   true,
    TrimWhitespace:     true,
    RemoveNullBytes:    true,
    NormalizeLineEndings: true,
    MaxLength:          5000,
    AllowHTML:          false,
    StrictMode:         true,
}

sanitizer := sanitizer.NewSanitizerWithConfig(config)

// Use custom configuration for specific sanitization
result := sanitizer.SanitizeString(input, config)
```

## Security Considerations

### Input Validation Security

1. **SQL Injection Prevention**
   - Validates against common SQL injection patterns
   - Detects UNION, SELECT, INSERT, UPDATE, DELETE, DROP, CREATE, ALTER, EXEC, EXECUTE
   - Should be used in conjunction with parameterized queries

2. **Cross-Site Scripting (XSS) Prevention**
   - Detects script tags and JavaScript code
   - Validates against dangerous HTML attributes
   - Removes event handlers (onload, onerror, onclick)

3. **Path Traversal Prevention**
   - Detects directory traversal attempts (../, ..\\)
   - Blocks access to system directories (/etc/, /proc/, /sys/, /dev/)
   - Validates file paths for safe operations

4. **Content Type Validation**
   - Ensures proper Content-Type headers
   - Validates JSON format for API requests
   - Prevents MIME type confusion attacks

### Sanitization Security

1. **Data Cleaning**
   - Removes null bytes and control characters
   - Normalizes Unicode to prevent encoding attacks
   - Trims whitespace to prevent padding attacks

2. **Type-Specific Security**
   - Email: Validates format and removes dangerous content
   - URL: Blocks dangerous protocols (javascript:, vbscript:, data:)
   - Phone: Ensures E.164 format compliance
   - UUID: Validates format and normalizes case

3. **HTML Security**
   - Removes script tags and content
   - Strips dangerous attributes and event handlers
   - Preserves safe HTML tags and attributes
   - Normalizes HTML structure

## Performance Considerations

### Validation Performance

1. **Compiled Regex Patterns**
   - All regex patterns are compiled once at initialization
   - Reduces runtime compilation overhead
   - Improves validation performance

2. **Early Termination**
   - Validation stops on first error in strict mode
   - Reduces unnecessary validation checks
   - Improves response time for invalid requests

3. **Performance Monitoring**
   - Tracks validation duration
   - Logs slow validation operations
   - Provides metrics for optimization

### Sanitization Performance

1. **Efficient String Operations**
   - Uses optimized string replacement functions
   - Minimizes memory allocations
   - Reduces garbage collection pressure

2. **Configurable Processing**
   - Skip unnecessary sanitization steps
   - Configure based on input type and requirements
   - Balance security vs. performance

## Configuration Options

### Validation Configuration

```go
type ValidationConfig struct {
    MaxRequestSize       int64    // Maximum request size in bytes
    AllowedOrigins       []string // CORS allowed origins
    EnableSanitization   bool     // Enable automatic sanitization
    StrictMode           bool     // Strict validation mode
    LogValidationErrors  bool     // Log validation errors
}
```

### Sanitization Configuration

```go
type SanitizationConfig struct {
    RemoveHTMLTags       bool // Remove HTML tags
    RemoveScripts        bool // Remove script content
    NormalizeUnicode     bool // Normalize Unicode characters
    TrimWhitespace       bool // Trim leading/trailing whitespace
    RemoveNullBytes      bool // Remove null bytes
    NormalizeLineEndings bool // Normalize line endings
    MaxLength            int  // Maximum string length
    AllowHTML            bool // Allow HTML content
    StrictMode           bool // Strict sanitization mode
}
```

## Error Handling

### Validation Errors

```go
type ValidationError struct {
    Field    string      `json:"field"`    // Field name
    Message  string      `json:"message"`  // Error message
    Value    interface{} `json:"value"`    // Invalid value
    Rule     string      `json:"rule"`     // Validation rule
    Severity string      `json:"severity"` // Error or warning
    Code     string      `json:"code"`     // Error code
}
```

### Error Response Format

```json
{
    "success": false,
    "errors": [
        {
            "field": "email",
            "message": "Invalid email format",
            "value": "invalid-email",
            "rule": "email",
            "severity": "error",
            "code": "invalid_email"
        }
    ],
    "warnings": [
        {
            "field": "phone",
            "message": "Phone number format warning",
            "value": "+1234567890",
            "rule": "phone",
            "severity": "warning",
            "code": "phone_format_warning"
        }
    ]
}
```

## Testing

### Running Tests

```bash
# Run validation tests
go test ./internal/api/middleware -v

# Run sanitization tests
go test ./pkg/sanitizer -v

# Run benchmarks
go test ./internal/api/middleware -bench=.
go test ./pkg/sanitizer -bench=.
```

### Test Coverage

```bash
# Generate coverage report
go test ./internal/api/middleware -cover
go test ./pkg/sanitizer -cover

# Generate HTML coverage report
go test ./internal/api/middleware -coverprofile=validation.out
go tool cover -html=validation.out -o validation.html
```

## Best Practices

### Validation Best Practices

1. **Always Validate Input**
   - Validate all user input, regardless of source
   - Use appropriate validation rules for data types
   - Implement defense in depth

2. **Use Struct Tags**
   - Define validation rules in struct tags
   - Keep validation logic close to data structures
   - Make validation requirements explicit

3. **Handle Errors Gracefully**
   - Provide clear error messages
   - Log validation failures for monitoring
   - Return appropriate HTTP status codes

4. **Performance Optimization**
   - Use compiled regex patterns
   - Implement early termination for invalid data
   - Monitor validation performance

### Sanitization Best Practices

1. **Sanitize Before Storage**
   - Clean data before storing in database
   - Normalize data for consistent processing
   - Remove dangerous content early

2. **Type-Specific Sanitization**
   - Use appropriate sanitization for data types
   - Validate after sanitization
   - Consider context-specific requirements

3. **Security First**
   - Prioritize security over convenience
   - Use strict sanitization for untrusted input
   - Regularly update security patterns

4. **Performance Considerations**
   - Balance security with performance
   - Use efficient string operations
   - Configure sanitization based on needs

## Integration with Existing Code

### Middleware Integration

```go
// Add validation middleware to existing handlers
func setupRoutes() {
    validator := middleware.NewValidator(nil, logger)
    
    // Apply to specific routes
    router.Handle("/api/verify", validator.ValidationMiddleware(verifyHandler))
    router.Handle("/api/register", validator.ValidationMiddleware(registerHandler))
    
    // Apply globally
    router.Use(validator.ValidationMiddleware)
}
```

### Handler Integration

```go
// Integrate with existing handlers
func existingHandler(w http.ResponseWriter, r *http.Request) {
    // Get validator from context
    validator := r.Context().Value("validator").(*middleware.Validator)
    
    var requestData YourRequestStruct
    result := validator.ValidateRequest(r, &requestData)
    
    if !result.IsValid {
        // Handle validation errors
        return
    }
    
    // Process validated data
    // ...
}
```

## Monitoring and Logging

### Validation Logging

```go
// Configure validation logging
config := &middleware.ValidationConfig{
    LogValidationErrors: true,
}

// Log validation events
logger.Info("validation completed",
    zap.Bool("is_valid", result.IsValid),
    zap.Int("error_count", len(result.Errors)),
    zap.Int("warning_count", len(result.Warnings)),
    zap.Duration("duration", result.Duration))
```

### Sanitization Logging

```go
// Log sanitization events
logger.Info("input sanitized",
    zap.String("input_type", "email"),
    zap.String("original_length", fmt.Sprintf("%d", len(original))),
    zap.String("sanitized_length", fmt.Sprintf("%d", len(sanitized))),
    zap.Bool("was_modified", original != sanitized))
```

## Troubleshooting

### Common Issues

1. **Validation Not Working**
   - Check struct tags are properly formatted
   - Ensure validation middleware is applied
   - Verify configuration is correct

2. **Performance Issues**
   - Monitor validation duration
   - Check regex pattern complexity
   - Consider caching validation results

3. **Security Concerns**
   - Review validation rules regularly
   - Update security patterns
   - Test with malicious input

### Debug Mode

```go
// Enable debug logging
logger := zap.NewDevelopment()

// Add debug information to validation results
if debug {
    result.Debug = map[string]interface{}{
        "validation_rules": rules,
        "field_values":     values,
        "processing_time":  duration,
    }
}
```

## Future Enhancements

### Planned Features

1. **Custom Validation Rules**
   - User-defined validation functions
   - Plugin architecture for custom rules
   - Rule composition and chaining

2. **Advanced Sanitization**
   - Machine learning-based content analysis
   - Context-aware sanitization
   - Adaptive security rules

3. **Performance Improvements**
   - Parallel validation processing
   - Caching and memoization
   - Optimized regex patterns

4. **Integration Enhancements**
   - GraphQL validation support
   - gRPC validation middleware
   - Database constraint validation

### Contributing

To contribute to the validation and sanitization system:

1. Follow the existing code style and patterns
2. Add comprehensive tests for new features
3. Update documentation for changes
4. Consider security implications
5. Test performance impact

## Conclusion

The input validation and sanitization system provides a robust foundation for ensuring data integrity and security in the Enhanced Business Intelligence System. By following the guidelines and best practices outlined in this documentation, developers can effectively protect against common security vulnerabilities while maintaining high performance and usability.

For additional support or questions, refer to the API documentation or contact the development team.
