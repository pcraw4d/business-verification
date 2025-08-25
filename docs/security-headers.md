# Security Headers Documentation

## Overview

The Security Headers system provides comprehensive HTTP security headers to protect the KYB Platform against various web-based attacks and vulnerabilities. This system implements industry-standard security headers with configurable options for different deployment environments.

## Architecture

### Core Components

1. **SecurityHeadersMiddleware**: Main middleware component that applies security headers
2. **SecurityHeadersConfig**: Configuration structure for all security header options
3. **Predefined Configurations**: Ready-to-use configurations for different environments

### Security Headers Implemented

#### 1. Content Security Policy (CSP)
- **Purpose**: Prevents XSS attacks by controlling resource loading
- **Header**: `Content-Security-Policy`
- **Configuration**: `CSPEnabled`, `CSPDirectives`

#### 2. HTTP Strict Transport Security (HSTS)
- **Purpose**: Forces HTTPS connections and prevents protocol downgrade attacks
- **Header**: `Strict-Transport-Security`
- **Configuration**: `HSTSEnabled`, `HSTSMaxAge`, `HSTSIncludeSubdomains`, `HSTSPreload`

#### 3. X-Frame-Options
- **Purpose**: Prevents clickjacking attacks
- **Header**: `X-Frame-Options`
- **Configuration**: `FrameOptions`
- **Values**: `DENY`, `SAMEORIGIN`, `ALLOW-FROM`

#### 4. X-Content-Type-Options
- **Purpose**: Prevents MIME type sniffing
- **Header**: `X-Content-Type-Options`
- **Configuration**: `ContentTypeOptions`
- **Value**: `nosniff`

#### 5. X-XSS-Protection
- **Purpose**: Enables browser's XSS filtering
- **Header**: `X-XSS-Protection`
- **Configuration**: `XSSProtection`
- **Value**: `1; mode=block`

#### 6. Referrer Policy
- **Purpose**: Controls referrer information in requests
- **Header**: `Referrer-Policy`
- **Configuration**: `ReferrerPolicy`
- **Values**: `no-referrer`, `strict-origin-when-cross-origin`, `no-referrer-when-downgrade`

#### 7. Permissions Policy
- **Purpose**: Controls browser features and APIs
- **Header**: `Permissions-Policy`
- **Configuration**: `PermissionsPolicyEnabled`, `PermissionsPolicy`

#### 8. Server Information
- **Purpose**: Customizes server identification
- **Header**: `Server`
- **Configuration**: `ServerName`

#### 9. Additional Headers
- **Purpose**: Custom security headers
- **Configuration**: `AdditionalHeaders`

## Configuration

### SecurityHeadersConfig Structure

```go
type SecurityHeadersConfig struct {
    // Content Security Policy
    CSPEnabled     bool   `json:"csp_enabled" yaml:"csp_enabled"`
    CSPDirectives  string `json:"csp_directives" yaml:"csp_directives"`
    
    // HTTP Strict Transport Security
    HSTSEnabled    bool          `json:"hsts_enabled" yaml:"hsts_enabled"`
    HSTSMaxAge     time.Duration `json:"hsts_max_age" yaml:"hsts_max_age"`
    HSTSIncludeSubdomains bool   `json:"hsts_include_subdomains" yaml:"hsts_include_subdomains"`
    HSTSPreload    bool          `json:"hsts_preload" yaml:"hsts_preload"`
    
    // Frame Options
    FrameOptions   string `json:"frame_options" yaml:"frame_options"`
    
    // Content Type Options
    ContentTypeOptions string `json:"content_type_options" yaml:"content_type_options"`
    
    // XSS Protection
    XSSProtection string `json:"xss_protection" yaml:"xss_protection"`
    
    // Referrer Policy
    ReferrerPolicy string `json:"referrer_policy" yaml:"referrer_policy"`
    
    // Permissions Policy
    PermissionsPolicyEnabled bool   `json:"permissions_policy_enabled" yaml:"permissions_policy_enabled"`
    PermissionsPolicy       string `json:"permissions_policy" yaml:"permissions_policy"`
    
    // Server Information
    ServerName string `json:"server_name" yaml:"server_name"`
    
    // Additional Headers
    AdditionalHeaders map[string]string `json:"additional_headers" yaml:"additional_headers"`
    
    // Exclude Paths
    ExcludePaths []string `json:"exclude_paths" yaml:"exclude_paths"`
}
```

### Predefined Configurations

#### 1. Strict Security Configuration
```go
StrictSecurityConfig = &SecurityHeadersConfig{
    CSPEnabled:              true,
    CSPDirectives:           "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self'; font-src 'self'; connect-src 'self'; frame-ancestors 'none';",
    HSTSEnabled:             true,
    HSTSMaxAge:              31536000 * time.Second, // 1 year
    HSTSIncludeSubdomains:   true,
    HSTSPreload:             true,
    FrameOptions:            "DENY",
    ContentTypeOptions:      "nosniff",
    XSSProtection:           "1; mode=block",
    ReferrerPolicy:          "no-referrer",
    PermissionsPolicyEnabled: true,
    PermissionsPolicy:       "geolocation=(), microphone=(), camera=(), payment=(), usb=(), magnetometer=(), gyroscope=(), accelerometer=()",
    ServerName:              "KYB-Tool",
    AdditionalHeaders: map[string]string{
        "X-Download-Options": "noopen",
        "X-Permitted-Cross-Domain-Policies": "none",
    },
}
```

#### 2. Balanced Security Configuration
```go
BalancedSecurityConfig = &SecurityHeadersConfig{
    CSPEnabled:              true,
    CSPDirectives:           "default-src 'self'; style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net https://cdnjs.cloudflare.com; font-src 'self' https://cdn.jsdelivr.net https://cdnjs.cloudflare.com data:; script-src 'self' 'unsafe-inline'; img-src 'self' data: https:; connect-src 'self' https:;",
    HSTSEnabled:             true,
    HSTSMaxAge:              31536000 * time.Second, // 1 year
    HSTSIncludeSubdomains:   true,
    HSTSPreload:             false,
    FrameOptions:            "SAMEORIGIN",
    ContentTypeOptions:      "nosniff",
    XSSProtection:           "1; mode=block",
    ReferrerPolicy:          "strict-origin-when-cross-origin",
    PermissionsPolicyEnabled: true,
    PermissionsPolicy:       "geolocation=(), microphone=(), camera=()",
    ServerName:              "KYB-Tool",
    AdditionalHeaders: map[string]string{
        "X-Download-Options": "noopen",
    },
}
```

#### 3. Development Security Configuration
```go
DevelopmentSecurityConfig = &SecurityHeadersConfig{
    CSPEnabled:              false, // Disabled for development flexibility
    CSPDirectives:           "",
    HSTSEnabled:             false, // Disabled for development
    HSTSMaxAge:              0,
    HSTSIncludeSubdomains:   false,
    HSTSPreload:             false,
    FrameOptions:            "SAMEORIGIN",
    ContentTypeOptions:      "nosniff",
    XSSProtection:           "1; mode=block",
    ReferrerPolicy:          "no-referrer-when-downgrade",
    PermissionsPolicyEnabled: false,
    PermissionsPolicy:       "",
    ServerName:              "KYB-Tool-Dev",
    AdditionalHeaders:       make(map[string]string),
}
```

## Usage

### Basic Usage

```go
package main

import (
    "net/http"
    "go.uber.org/zap"
    "github.com/your-org/kyb-platform/internal/api/middleware"
)

func main() {
    logger := zap.NewProduction()
    
    // Create middleware with default configuration
    securityMiddleware := middleware.NewSecurityHeadersMiddleware(nil, logger)
    
    // Create your handler
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, World!"))
    })
    
    // Apply middleware
    http.Handle("/", securityMiddleware.Middleware(handler))
    
    http.ListenAndServe(":8080", nil)
}
```

### Using Predefined Configurations

```go
// For production with strict security
securityMiddleware := middleware.NewStrictSecurityHeadersMiddleware(logger)

// For production with balanced security
securityMiddleware := middleware.NewBalancedSecurityHeadersMiddleware(logger)

// For development
securityMiddleware := middleware.NewDevelopmentSecurityHeadersMiddleware(logger)
```

### Custom Configuration

```go
config := &middleware.SecurityHeadersConfig{
    CSPEnabled:     true,
    CSPDirectives:  "default-src 'self'; script-src 'self' 'unsafe-inline';",
    HSTSEnabled:    true,
    HSTSMaxAge:     31536000 * time.Second,
    FrameOptions:   "SAMEORIGIN",
    ServerName:     "My-Custom-Server",
    ExcludePaths:   []string{"/health", "/metrics"},
    AdditionalHeaders: map[string]string{
        "X-Custom-Security": "value",
    },
}

securityMiddleware := middleware.NewSecurityHeadersMiddleware(config, logger)
```

### Dynamic Configuration Updates

```go
// Update configuration at runtime
newConfig := &middleware.SecurityHeadersConfig{
    CSPDirectives: "default-src 'self'; script-src 'self';",
    FrameOptions:  "DENY",
}
securityMiddleware.UpdateConfig(newConfig)

// Add exclude paths
securityMiddleware.AddExcludePath("/api/public")

// Add custom headers
securityMiddleware.AddAdditionalHeader("X-Custom-Header", "value")
```

## Security Features

### 1. Content Security Policy (CSP)

CSP prevents XSS attacks by controlling which resources can be loaded:

```go
// Strict CSP
CSPDirectives: "default-src 'self'; script-src 'self'; style-src 'self';"

// Balanced CSP (allows CDNs)
CSPDirectives: "default-src 'self'; style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net;"
```

### 2. HTTP Strict Transport Security (HSTS)

HSTS forces HTTPS connections:

```go
// Basic HSTS
HSTSEnabled: true
HSTSMaxAge: 31536000 * time.Second // 1 year

// HSTS with subdomains and preload
HSTSIncludeSubdomains: true
HSTSPreload: true
```

### 3. Frame Protection

Prevents clickjacking attacks:

```go
// Deny all frames
FrameOptions: "DENY"

// Allow same-origin frames
FrameOptions: "SAMEORIGIN"
```

### 4. MIME Type Protection

Prevents MIME type sniffing:

```go
ContentTypeOptions: "nosniff"
```

### 5. XSS Protection

Enables browser's XSS filtering:

```go
XSSProtection: "1; mode=block"
```

### 6. Referrer Policy

Controls referrer information:

```go
// No referrer information
ReferrerPolicy: "no-referrer"

// Strict origin when cross-origin
ReferrerPolicy: "strict-origin-when-cross-origin"
```

### 7. Permissions Policy

Controls browser features:

```go
PermissionsPolicy: "geolocation=(), microphone=(), camera=(), payment=()"
```

## Path Exclusion

Exclude specific paths from security headers:

```go
config := &SecurityHeadersConfig{
    ExcludePaths: []string{
        "/health",           // Health check endpoint
        "/metrics",          // Metrics endpoint
        "/api/public/",      // Public API endpoints
    },
}
```

## Additional Headers

Add custom security headers:

```go
config := &SecurityHeadersConfig{
    AdditionalHeaders: map[string]string{
        "X-Download-Options": "noopen",
        "X-Permitted-Cross-Domain-Policies": "none",
        "X-Custom-Security": "value",
    },
}
```

## Testing

### Unit Tests

The security headers middleware includes comprehensive unit tests:

```bash
# Run all security headers tests
go test ./internal/api/middleware -run TestSecurityHeaders

# Run specific test
go test ./internal/api/middleware -run TestSecurityHeadersMiddleware_Middleware

# Run with coverage
go test ./internal/api/middleware -cover -run TestSecurityHeaders
```

### Test Coverage

Tests cover:
- Configuration creation and validation
- Header application
- Path exclusion logic
- Configuration updates
- Predefined configurations
- Edge cases and error conditions
- Performance benchmarks

## Monitoring and Logging

### Debug Logging

The middleware logs security header application:

```go
logger.Debug("Security headers applied",
    zap.String("path", r.URL.Path),
    zap.String("method", r.Method),
    zap.String("user_agent", r.UserAgent()),
    zap.String("remote_addr", r.RemoteAddr))
```

### Configuration Changes

Configuration updates are logged:

```go
logger.Info("Security headers configuration updated")
logger.Info("Security headers exclude path added", zap.String("path", path))
logger.Info("Additional security header added", zap.String("key", key), zap.String("value", value))
```

## Best Practices

### 1. Environment-Specific Configuration

- **Production**: Use strict or balanced security configuration
- **Development**: Use development configuration for flexibility
- **Staging**: Use balanced configuration to test production-like settings

### 2. CSP Configuration

- Start with strict CSP and relax as needed
- Use `report-uri` directive for CSP violation monitoring
- Test CSP thoroughly before deployment

### 3. HSTS Configuration

- Start with shorter max-age in testing
- Gradually increase max-age in production
- Use preload only after thorough testing

### 4. Path Exclusion

- Only exclude paths that absolutely need it
- Document why paths are excluded
- Regularly review excluded paths

### 5. Monitoring

- Monitor CSP violations
- Track security header effectiveness
- Log configuration changes

## Troubleshooting

### Common Issues

#### 1. CSP Blocking Resources

**Problem**: CSP is blocking legitimate resources
**Solution**: Adjust CSP directives or exclude problematic paths

```go
// More permissive CSP
CSPDirectives: "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline';"
```

#### 2. HSTS Issues in Development

**Problem**: HSTS causing issues in development environment
**Solution**: Disable HSTS in development

```go
DevelopmentSecurityConfig.HSTSEnabled = false
```

#### 3. Frame Options Blocking Integration

**Problem**: Frame options blocking legitimate iframe usage
**Solution**: Use SAMEORIGIN instead of DENY

```go
FrameOptions: "SAMEORIGIN"
```

### Debug Mode

Enable debug logging to troubleshoot issues:

```go
logger := zap.NewDevelopment()
securityMiddleware := middleware.NewSecurityHeadersMiddleware(config, logger)
```

## Integration with Other Systems

### 1. Load Balancers

Configure load balancers to preserve security headers:

```nginx
# Nginx configuration
proxy_hide_header Server;
add_header X-Frame-Options "SAMEORIGIN" always;
add_header X-Content-Type-Options "nosniff" always;
```

### 2. CDNs

Ensure CDNs don't strip security headers:

```yaml
# Cloudflare configuration
headers:
  - name: "X-Frame-Options"
    value: "SAMEORIGIN"
  - name: "X-Content-Type-Options"
    value: "nosniff"
```

### 3. Reverse Proxies

Configure reverse proxies to handle security headers properly:

```yaml
# Traefik configuration
headers:
  customRequestHeaders:
    X-Forwarded-Proto: "https"
  customResponseHeaders:
    X-Frame-Options: "SAMEORIGIN"
```

## Security Considerations

### 1. Header Order

Security headers are applied in a specific order for maximum effectiveness:
1. Content Security Policy
2. HTTP Strict Transport Security
3. Frame Options
4. Content Type Options
5. XSS Protection
6. Referrer Policy
7. Permissions Policy
8. Server Information
9. Additional Headers

### 2. Header Conflicts

Some headers may conflict with others:
- CSP and XSS Protection: CSP is more effective
- HSTS and mixed content: HSTS prevents mixed content
- Frame Options and CSP frame-ancestors: CSP frame-ancestors takes precedence

### 3. Browser Compatibility

Consider browser compatibility:
- CSP: Modern browsers
- HSTS: Modern browsers
- Permissions Policy: Modern browsers
- XSS Protection: Legacy browsers

## Performance Impact

### Minimal Overhead

The security headers middleware has minimal performance impact:
- Header application: ~1-2 microseconds per request
- Path exclusion check: ~0.1 microseconds per request
- Configuration lookup: ~0.1 microseconds per request

### Optimization Tips

1. Use path exclusion for high-traffic endpoints
2. Minimize additional headers
3. Use predefined configurations when possible
4. Cache configuration objects

## Future Enhancements

### Planned Features

1. **Dynamic CSP**: Generate CSP based on application analysis
2. **Header Analytics**: Track security header effectiveness
3. **Automated Testing**: Automated security header testing
4. **Configuration Validation**: Validate configuration at startup
5. **Header Monitoring**: Monitor security header compliance

### Extension Points

The middleware is designed for easy extension:
- Custom header providers
- Dynamic configuration sources
- Header validation plugins
- Monitoring integrations

## Conclusion

The Security Headers system provides comprehensive protection against web-based attacks while maintaining flexibility for different deployment environments. By following the best practices outlined in this documentation, you can effectively secure your KYB Platform application.
