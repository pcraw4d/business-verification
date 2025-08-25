# API Versioning Middleware

## Overview

The API Versioning Middleware provides comprehensive API versioning capabilities for the KYB Platform, enabling seamless version management, deprecation handling, and client compatibility validation. This middleware integrates with the existing version manager to provide a robust versioning solution.

## Features

### Core Features
- **Multi-Method Version Detection**: Support for URL path, headers, query parameters, and Accept header versioning
- **Version Validation**: Automatic validation against supported versions with fallback options
- **Deprecation Management**: Built-in deprecation warnings and sunset date handling
- **Client Version Validation**: Optional client version compatibility checking
- **Path Rewriting**: Automatic removal of version prefixes from URLs
- **Context Integration**: Version information available through request context

### Security Features
- **Strict Versioning**: Configurable strict mode for version enforcement
- **Error Handling**: Comprehensive error responses for unsupported versions
- **Logging**: Detailed logging of version detection and errors
- **Header Security**: Proper version-related response headers

### Customization Features
- **Flexible Configuration**: Extensive configuration options for different deployment scenarios
- **Predefined Presets**: Default, strict, and permissive configuration presets
- **Custom Headers**: Configurable header names and parameter names
- **Behavior Control**: Fine-grained control over versioning behavior

## Configuration

### Basic Configuration

```go
import (
    "github.com/pcraw4d/business-verification/internal/api/middleware"
    "github.com/pcraw4d/business-verification/internal/api/compatibility"
    "go.uber.org/zap"
)

// Create version manager
versionManager := compatibility.NewVersionManager(logger, nil)

// Create middleware with default configuration
apiVersioningMiddleware := middleware.NewAPIVersioningMiddleware(nil, versionManager, logger)
```

### Configuration Options

```go
config := &middleware.APIVersioningConfig{
    // Version Detection
    EnableURLVersioning:     true,  // Enable /v1/, /v2/ etc. in URLs
    EnableHeaderVersioning:  true,  // Enable X-API-Version header
    EnableQueryVersioning:   true,  // Enable ?version=v1 query parameter
    EnableAcceptVersioning:  true,  // Enable Accept header versioning
    
    // Version Headers
    VersionHeaderName:       "X-API-Version",           // Custom header name
    QueryVersionParam:       "version",                 // Query parameter name
    AcceptVersionPrefix:     "vnd.kyb-platform.v",      // Accept header prefix
    
    // Behavior
    StrictVersioning:        false,  // Require exact version match
    AllowVersionFallback:    true,   // Allow fallback to supported versions
    RemoveVersionFromPath:   true,   // Remove version from path for handlers
    
    // Error Handling
    ReturnVersionErrors:     true,   // Return detailed error responses
    LogVersionFailures:      true,   // Log version-related errors
    
    // Deprecation
    EnableDeprecationWarnings: true,  // Enable deprecation warnings
    DeprecationWarningDays:    30,    // Days to show deprecation warnings
    
    // Client Validation
    EnableClientValidation:   true,   // Enable client version validation
    ClientVersionHeader:      "X-Client-Version",       // Client version header
}
```

### Predefined Configurations

#### Default Configuration
```go
config := middleware.GetDefaultAPIVersioningConfig()
// Balanced configuration with all features enabled
```

#### Strict Configuration
```go
config := middleware.GetStrictAPIVersioningConfig()
// Strict versioning with minimal fallback options
```

#### Permissive Configuration
```go
config := middleware.GetPermissiveAPIVersioningConfig()
// Lenient versioning with maximum compatibility
```

## Usage Examples

### Basic Usage

```go
package main

import (
    "net/http"
    "github.com/pcraw4d/business-verification/internal/api/middleware"
    "github.com/pcraw4d/business-verification/internal/api/compatibility"
    "go.uber.org/zap"
)

func main() {
    logger := zap.NewProduction()
    
    // Initialize version manager
    versionManager := compatibility.NewVersionManager(logger, nil)
    
    // Create API versioning middleware
    apiVersioningMiddleware := middleware.NewAPIVersioningMiddleware(nil, versionManager, logger)
    
    // Create your handlers
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Get version information from context
        versionInfo := middleware.GetVersionInfo(r.Context())
        apiVersion := middleware.GetAPIVersion(r.Context())
        
        // Use version information in your handler
        response := map[string]interface{}{
            "data": "Your API response",
            "version": apiVersion,
            "is_deprecated": versionInfo.IsDeprecated,
        }
        
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(response)
    })
    
    // Apply middleware
    finalHandler := apiVersioningMiddleware.Middleware(handler)
    
    // Start server
    http.ListenAndServe(":8080", finalHandler)
}
```

### Version Detection Methods

#### URL Path Versioning
```bash
# Version in URL path
GET /v3/businesses/123
GET /v2/verifications
GET /v1/health
```

#### Header Versioning
```bash
# Version in custom header
curl -H "X-API-Version: v3" https://api.kyb-platform.com/businesses/123
```

#### Query Parameter Versioning
```bash
# Version in query parameter
GET /businesses/123?version=v2
GET /verifications?version=v1
```

#### Accept Header Versioning
```bash
# Version in Accept header
curl -H "Accept: application/vnd.kyb-platform.v3+json" https://api.kyb-platform.com/businesses/123
```

### Handler Integration

```go
func BusinessHandler(w http.ResponseWriter, r *http.Request) {
    // Get version information
    versionInfo := middleware.GetVersionInfo(r.Context())
    apiVersion := middleware.GetAPIVersion(r.Context())
    
    // Handle different versions
    switch apiVersion {
    case "v1":
        handleBusinessV1(w, r)
    case "v2":
        handleBusinessV2(w, r)
    case "v3":
        handleBusinessV3(w, r)
    default:
        http.Error(w, "Unsupported version", http.StatusBadRequest)
    }
    
    // Add version-specific headers
    if versionInfo.IsDeprecated {
        w.Header().Set("X-Deprecation-Warning", "This version will be removed soon")
    }
}

func handleBusinessV1(w http.ResponseWriter, r *http.Request) {
    // V1 implementation
    response := map[string]interface{}{
        "id": "123",
        "name": "Business Name",
        // V1 has limited fields
    }
    json.NewEncoder(w).Encode(response)
}

func handleBusinessV3(w http.ResponseWriter, r *http.Request) {
    // V3 implementation with enhanced features
    response := map[string]interface{}{
        "id": "123",
        "name": "Business Name",
        "metadata": map[string]interface{}{
            "confidence_score": 0.95,
            "data_sources": []string{"government", "credit_bureau"},
            "last_updated": time.Now().ISO8601(),
        },
        "industry_codes": []map[string]string{
            {"type": "SIC", "code": "7372"},
            {"type": "NAICS", "code": "541511"},
        },
    }
    json.NewEncoder(w).Encode(response)
}
```

## Response Headers

The middleware automatically adds version-related headers to responses:

### Standard Headers
- `X-API-Version`: The resolved API version
- `X-API-Version-Requested`: The originally requested version

### Deprecation Headers (when applicable)
- `X-API-Deprecated`: Set to "true" for deprecated versions
- `X-API-Deprecated-At`: ISO 8601 timestamp of deprecation date
- `X-API-Sunset-Date`: ISO 8601 timestamp of removal date
- `X-API-Migration-Guide`: URL to migration documentation
- `X-API-Deprecation-Warning`: Human-readable deprecation warning

### Client Version Headers (when enabled)
- `X-Client-Version`: The client version from request
- `X-Client-Version-Warning`: Warning if client version is incompatible

## Error Responses

### Unsupported Version Error
```json
{
  "error": {
    "type": "unsupported_version",
    "message": "Unsupported API version: v99",
    "code": "UNSUPPORTED_VERSION",
    "requested_version": "v99",
    "supported_versions": ["v1", "v2", "v3"],
    "migration_guide": "https://docs.kyb-platform.com/migration"
  },
  "success": false
}
```

### Invalid Version Error
```json
{
  "error": {
    "type": "invalid_version",
    "message": "Invalid API version: invalid",
    "code": "INVALID_VERSION",
    "requested_version": "invalid"
  },
  "success": false
}
```

## Integration with Other Middleware

### Middleware Order
```go
// Recommended middleware order
router := mux.NewRouter()

// 1. Request logging (first to capture all requests)
router.Use(requestLoggingMiddleware.Middleware)

// 2. CORS (early to handle preflight requests)
router.Use(corsMiddleware.Middleware)

// 3. API versioning (before authentication)
router.Use(apiVersioningMiddleware.Middleware)

// 4. Authentication
router.Use(authMiddleware.Middleware)

// 5. Rate limiting
router.Use(rateLimitMiddleware.Middleware)

// 6. Request validation
router.Use(requestValidationMiddleware.Middleware)

// 7. Error handling (last to catch all errors)
router.Use(errorHandlingMiddleware.Middleware)
```

### Context Integration
```go
// The middleware adds version information to request context
func MyHandler(w http.ResponseWriter, r *http.Request) {
    // Get version info
    versionInfo := middleware.GetVersionInfo(r.Context())
    apiVersion := middleware.GetAPIVersion(r.Context())
    
    // Use in logging
    logger.Info("Processing request",
        zap.String("api_version", apiVersion),
        zap.Bool("is_deprecated", versionInfo.IsDeprecated),
    )
    
    // Use in business logic
    if versionInfo.IsDeprecated {
        // Handle deprecated version logic
    }
}
```

## Version Detection Priority

The middleware detects versions in the following order:

1. **URL Path** (e.g., `/v3/businesses`)
2. **Custom Header** (e.g., `X-API-Version: v3`)
3. **Query Parameter** (e.g., `?version=v3`)
4. **Accept Header** (e.g., `Accept: application/vnd.kyb-platform.v3+json`)
5. **Default Version** (if no version detected)

## Deprecation Management

### Automatic Deprecation Warnings
```go
// The middleware automatically adds deprecation warnings for deprecated versions
// Warnings are shown for a configurable number of days after deprecation

config := &middleware.APIVersioningConfig{
    EnableDeprecationWarnings: true,
    DeprecationWarningDays:    30, // Show warnings for 30 days
}
```

### Migration Support
```go
// Version info includes migration guidance
versionInfo := middleware.GetVersionInfo(r.Context())
if versionInfo.IsDeprecated && versionInfo.MigrationGuide != "" {
    // Provide migration guidance to clients
    w.Header().Set("Link", fmt.Sprintf("<%s>; rel=\"migration\"", versionInfo.MigrationGuide))
}
```

## Client Version Validation

### Enable Client Validation
```go
config := &middleware.APIVersioningConfig{
    EnableClientValidation: true,
    ClientVersionHeader:    "X-Client-Version",
}
```

### Client Version Headers
```bash
# Valid client version
curl -H "X-Client-Version: 3.0.0" https://api.kyb-platform.com/v3/businesses

# Invalid client version (will show warning)
curl -H "X-Client-Version: 1.0.0" https://api.kyb-platform.com/v3/businesses
```

## Testing

### Unit Testing
```go
func TestAPIVersioningMiddleware(t *testing.T) {
    versionManager := compatibility.NewVersionManager(zap.NewNop(), nil)
    middleware := middleware.NewAPIVersioningMiddleware(nil, versionManager, zap.NewNop())
    
    // Test URL versioning
    req := httptest.NewRequest("GET", "/v3/test", nil)
    w := httptest.NewRecorder()
    
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        version := middleware.GetAPIVersion(r.Context())
        assert.Equal(t, "v3", version)
        w.WriteHeader(http.StatusOK)
    })
    
    middleware.Middleware(handler).ServeHTTP(w, req)
    assert.Equal(t, http.StatusOK, w.Code)
    assert.Equal(t, "v3", w.Header().Get("X-API-Version"))
}
```

### Integration Testing
```go
func TestAPIVersioningIntegration(t *testing.T) {
    // Test with real version manager and multiple detection methods
    versionManager := compatibility.NewVersionManager(zap.NewNop(), nil)
    middleware := middleware.NewAPIVersioningMiddleware(nil, versionManager, zap.NewNop())
    
    tests := []struct {
        name           string
        requestPath    string
        headers        map[string]string
        expectedVersion string
    }{
        {"URL versioning", "/v2/test", map[string]string{}, "v2"},
        {"Header versioning", "/test", map[string]string{"X-API-Version": "v3"}, "v3"},
        {"Query versioning", "/test?version=v1", map[string]string{}, "v1"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest("GET", tt.requestPath, nil)
            for key, value := range tt.headers {
                req.Header.Set(key, value)
            }
            
            w := httptest.NewRecorder()
            
            handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                version := middleware.GetAPIVersion(r.Context())
                assert.Equal(t, tt.expectedVersion, version)
                w.WriteHeader(http.StatusOK)
            })
            
            middleware.Middleware(handler).ServeHTTP(w, req)
            assert.Equal(t, http.StatusOK, w.Code)
        })
    }
}
```

## Performance Considerations

### Caching
- Version regex is compiled once during middleware creation
- Version manager caches version information
- Context lookups are O(1) operations

### Memory Usage
- Version info is stored in request context (minimal overhead)
- Headers are added to response (standard HTTP overhead)
- No persistent state maintained

### Processing Overhead
- URL parsing: O(1) with regex
- Header parsing: O(n) where n is number of headers
- Version validation: O(1) with map lookup
- Context operations: O(1)

## Security Considerations

### Version Validation
- All versions are validated against supported versions
- Strict mode prevents fallback to unsupported versions
- Error responses don't leak internal version information

### Header Security
- Custom header names are configurable
- Headers are properly sanitized
- No sensitive information in version headers

### Error Handling
- Detailed error responses are configurable
- 404 fallback option for security
- Proper HTTP status codes

## Best Practices

### Version Management
1. **Plan Deprecations**: Use deprecation periods to give clients time to migrate
2. **Document Changes**: Provide clear migration guides for breaking changes
3. **Monitor Usage**: Track version usage to plan deprecations
4. **Test Compatibility**: Ensure version compatibility matrix is accurate

### Configuration
1. **Use Appropriate Preset**: Choose default, strict, or permissive based on your needs
2. **Enable Logging**: Enable version failure logging for monitoring
3. **Set Timeouts**: Configure appropriate timeouts for version validation
4. **Customize Headers**: Use consistent header names across your API

### Client Integration
1. **Version Detection**: Implement robust version detection in clients
2. **Fallback Strategy**: Plan fallback strategies for unsupported versions
3. **Migration Planning**: Plan client migrations before version deprecation
4. **Error Handling**: Handle version errors gracefully in clients

## Troubleshooting

### Common Issues

#### Version Not Detected
```bash
# Check if version detection is enabled
config.EnableURLVersioning = true
config.EnableHeaderVersioning = true

# Verify URL format
GET /v3/businesses  # Correct
GET /version3/businesses  # Incorrect
```

#### Deprecation Warnings Not Showing
```bash
# Check deprecation configuration
config.EnableDeprecationWarnings = true
config.DeprecationWarningDays = 30

# Verify version is actually deprecated
# Check version manager configuration
```

#### Client Version Warnings
```bash
# Check client validation configuration
config.EnableClientValidation = true
config.ClientVersionHeader = "X-Client-Version"

# Verify client version format
X-Client-Version: 3.0.0  # Correct
X-Client-Version: v3.0.0  # May cause issues
```

### Debug Mode
```go
// Enable detailed logging for debugging
config.LogVersionFailures = true

// Check logs for version detection details
logger.Info("API version detected",
    zap.String("requested_version", requestedVersion),
    zap.String("resolved_version", resolvedVersion),
    zap.String("detection_method", detectionMethod),
)
```

## Migration Guide

### From Manual Version Handling
```go
// Before: Manual version handling
func OldHandler(w http.ResponseWriter, r *http.Request) {
    version := r.URL.Query().Get("version")
    if version == "" {
        version = "v1" // Default
    }
    
    switch version {
    case "v1":
        handleV1(w, r)
    case "v2":
        handleV2(w, r)
    default:
        http.Error(w, "Unsupported version", http.StatusBadRequest)
    }
}

// After: Using middleware
func NewHandler(w http.ResponseWriter, r *http.Request) {
    version := middleware.GetAPIVersion(r.Context())
    versionInfo := middleware.GetVersionInfo(r.Context())
    
    // Handle deprecation
    if versionInfo.IsDeprecated {
        // Add deprecation handling
    }
    
    switch version {
    case "v1":
        handleV1(w, r)
    case "v2":
        handleV2(w, r)
    case "v3":
        handleV3(w, r)
    }
}
```

### From URL-Based Versioning Only
```go
// Before: Only URL versioning
router.HandleFunc("/v1/businesses", handleBusinessesV1)
router.HandleFunc("/v2/businesses", handleBusinessesV2)

// After: Multiple version detection methods
router.HandleFunc("/businesses", handleBusinesses) // Middleware handles versioning
```

## API Reference

### Types

#### APIVersioningConfig
```go
type APIVersioningConfig struct {
    // Version Detection
    EnableURLVersioning     bool
    EnableHeaderVersioning  bool
    EnableQueryVersioning   bool
    EnableAcceptVersioning  bool
    
    // Version Headers
    VersionHeaderName       string
    QueryVersionParam       string
    AcceptVersionPrefix     string
    
    // Behavior
    StrictVersioning        bool
    AllowVersionFallback    bool
    RemoveVersionFromPath   bool
    
    // Error Handling
    ReturnVersionErrors     bool
    LogVersionFailures      bool
    
    // Deprecation
    EnableDeprecationWarnings bool
    DeprecationWarningDays    int
    
    // Client Validation
    EnableClientValidation   bool
    ClientVersionHeader      string
}
```

#### VersionInfo
```go
type VersionInfo struct {
    RequestedVersion string
    ResolvedVersion  string
    IsDeprecated     bool
    DeprecationDate  *time.Time
    SunsetDate       *time.Time
    MigrationGuide   string
    ClientVersion    string
    IsValidClient    bool
}
```

#### VersionError
```go
type VersionError struct {
    Type        string
    Message     string
    Code        string
    RequestedVersion string
    SupportedVersions []string
    MigrationGuide string
}
```

### Functions

#### NewAPIVersioningMiddleware
```go
func NewAPIVersioningMiddleware(config *APIVersioningConfig, versionManager *compatibility.VersionManager, logger *zap.Logger) *APIVersioningMiddleware
```

#### GetVersionInfo
```go
func GetVersionInfo(ctx context.Context) *VersionInfo
```

#### GetAPIVersion
```go
func GetAPIVersion(ctx context.Context) string
```

#### Configuration Presets
```go
func GetDefaultAPIVersioningConfig() *APIVersioningConfig
func GetStrictAPIVersioningConfig() *APIVersioningConfig
func GetPermissiveAPIVersioningConfig() *APIVersioningConfig
```

## Monitoring and Alerting

### Key Metrics
- Version usage distribution
- Deprecated version usage
- Version detection failures
- Client version compatibility issues

### Alerts
- High usage of deprecated versions
- Version detection failures
- Client version compatibility warnings
- Version migration deadlines approaching

### Logging
```go
// Structured logging for monitoring
logger.Info("API version detected",
    zap.String("requested_version", requestedVersion),
    zap.String("resolved_version", resolvedVersion),
    zap.String("detection_method", detectionMethod),
    zap.Bool("is_deprecated", versionInfo.IsDeprecated),
    zap.String("client_version", versionInfo.ClientVersion),
    zap.Bool("is_valid_client", versionInfo.IsValidClient),
)
```
