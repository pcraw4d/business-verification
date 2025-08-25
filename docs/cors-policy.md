# CORS Policy Implementation

## Overview

The Enhanced Business Intelligence System implements a comprehensive Cross-Origin Resource Sharing (CORS) policy that provides secure and flexible cross-origin access control. The CORS middleware supports configurable origins, methods, headers, and path-based rules to meet various deployment scenarios.

## Features

### Core CORS Features
- **Configurable Origins**: Support for exact domains, wildcard subdomains, and global wildcards
- **HTTP Methods Control**: Configurable allowed HTTP methods (GET, POST, PUT, DELETE, etc.)
- **Header Management**: Control over allowed and exposed headers
- **Credentials Support**: Configurable credentials handling for cookies and authorization headers
- **Preflight Caching**: Configurable cache duration for preflight requests
- **Path-Based Rules**: Different CORS policies for specific API paths
- **Debug Mode**: Enhanced logging for development and troubleshooting

### Security Features
- **Origin Validation**: Strict origin checking with pattern matching
- **Method Restrictions**: Limit allowed HTTP methods per path
- **Header Validation**: Control over which headers can be sent and received
- **Credentials Security**: Proper handling of credentials with origin restrictions
- **Preflight Protection**: Secure handling of OPTIONS preflight requests

## Configuration

### Basic Configuration

```go
import (
    "time"
    "github.com/your-org/kyb-platform/internal/api/middleware"
)

// Default configuration
config := &middleware.CORSConfig{
    AllowedOrigins: []string{
        "https://app.example.com",
        "https://admin.example.com",
        "http://localhost:3000",
    },
    AllowAllOrigins:  false,
    AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowedHeaders: []string{
        "Origin",
        "Content-Type",
        "Accept",
        "Authorization",
        "X-API-Key",
    },
    ExposedHeaders: []string{
        "X-Total-Count",
        "X-RateLimit-Limit",
        "X-RateLimit-Remaining",
    },
    AllowCredentials: true,
    MaxAge:           3600 * time.Second, // 1 hour
    Debug:            false,
}
```

### Predefined Configurations

#### Development Configuration
```go
// Development-friendly configuration
config := middleware.GetDevelopmentCORSConfig()
// Features:
// - Allows all origins (*)
// - All HTTP methods
// - All headers
// - Credentials enabled
// - Short cache time (5 minutes)
// - Debug mode enabled
```

#### Production Configuration
```go
// Strict production configuration
config := middleware.GetStrictCORSConfig()
// Features:
// - No origins allowed by default (must be explicitly set)
// - Limited HTTP methods
// - Restricted headers
// - Credentials enabled
// - Longer cache time (1 hour)
// - Debug mode disabled
```

#### Default Configuration
```go
// Balanced configuration
config := middleware.GetDefaultCORSConfig()
// Features:
// - Common development origins
// - Standard HTTP methods
// - Essential headers
// - Credentials enabled
// - 24-hour cache time
// - Debug mode disabled
```

### Path-Based Rules

Configure different CORS policies for specific API paths:

```go
config := &middleware.CORSConfig{
    // Global settings
    AllowedOrigins:   []string{"https://app.example.com"},
    AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
    AllowCredentials: true,
    
    // Path-specific rules
    PathRules: []middleware.CORSPathRule{
        {
            Path:            "/api/public",
            AllowedOrigins:  []string{"*"}, // Allow all origins for public API
            AllowedMethods:  []string{"GET", "OPTIONS"},
            AllowCredentials: false, // No credentials for public endpoints
            MaxAge:          300 * time.Second, // 5 minutes
        },
        {
            Path:            "/api/admin",
            AllowedOrigins:  []string{"https://admin.example.com"},
            AllowedMethods:  []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
            AllowedHeaders:  []string{"Origin", "Content-Type", "Authorization", "X-Admin-Key"},
            ExposedHeaders:  []string{"X-Admin-Data"},
            AllowCredentials: true,
            MaxAge:          1800 * time.Second, // 30 minutes
        },
        {
            Path:            "/api/webhook",
            AllowedOrigins:  []string{"https://partner.example.com"},
            AllowedMethods:  []string{"POST", "OPTIONS"},
            AllowedHeaders:  []string{"Origin", "Content-Type", "X-Webhook-Signature"},
            AllowCredentials: false,
            MaxAge:          600 * time.Second, // 10 minutes
        },
    },
}
```

## Usage

### Basic Usage

```go
package main

import (
    "net/http"
    "github.com/your-org/kyb-platform/internal/api/middleware"
    "go.uber.org/zap"
)

func main() {
    logger := zap.NewProduction()
    
    // Create CORS configuration
    corsConfig := middleware.GetDefaultCORSConfig()
    corsConfig.AllowedOrigins = []string{
        "https://app.example.com",
        "https://admin.example.com",
    }
    
    // Create CORS middleware
    corsMiddleware := middleware.NewCORSMiddleware(corsConfig, logger)
    
    // Create your main handler
    mainHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, CORS-enabled API!"))
    })
    
    // Apply CORS middleware
    handler := corsMiddleware.Middleware(mainHandler)
    
    // Start server
    http.ListenAndServe(":8080", handler)
}
```

### Integration with Existing Middleware

```go
func setupMiddleware(router *http.ServeMux, logger *zap.Logger) {
    // CORS middleware (should be first)
    corsConfig := middleware.GetDefaultCORSConfig()
    corsMiddleware := middleware.NewCORSMiddleware(corsConfig, logger)
    
    // Security headers middleware
    securityConfig := &middleware.SecurityHeadersConfig{
        CSPEnabled:     true,
        HSTSEnabled:    true,
        FrameOptions:   "DENY",
        XSSProtection:  "1; mode=block",
    }
    securityMiddleware := middleware.NewSecurityHeadersMiddleware(securityConfig, logger)
    
    // Rate limiting middleware
    rateLimitConfig := &middleware.RateLimitConfig{
        RequestsPerMinute: 100,
        BurstSize:         20,
    }
    rateLimitMiddleware := middleware.NewRateLimitMiddleware(rateLimitConfig, logger)
    
    // Apply middleware chain
    handler := corsMiddleware.Middleware(
        securityMiddleware.Middleware(
            rateLimitMiddleware.Middleware(router),
        ),
    )
    
    return handler
}
```

## Origin Patterns

### Supported Origin Patterns

1. **Exact Match**
   ```
   "https://app.example.com"
   ```

2. **Wildcard Subdomain**
   ```
   "https://*.example.com"  // Matches api.example.com, admin.example.com, etc.
   ```

3. **Global Wildcard**
   ```
   "*"  // Matches any origin (use with caution)
   ```

4. **Localhost Development**
   ```
   "http://localhost:3000"
   "https://localhost:3000"
   ```

### Origin Matching Examples

```go
config := &middleware.CORSConfig{
    AllowedOrigins: []string{
        "https://app.example.com",        // Exact match
        "https://*.example.com",          // Subdomain wildcard
        "http://localhost:3000",          // Development
        "*",                             // Global wildcard (development only)
    },
}
```

## HTTP Headers

### Standard CORS Headers

#### Request Headers (Allowed)
- `Origin`: The origin of the request
- `Content-Type`: Content type of the request
- `Accept`: Accepted response types
- `Authorization`: Authentication credentials
- `X-Requested-With`: AJAX request indicator
- `X-API-Key`: API key for authentication

#### Response Headers (Exposed)
- `X-Total-Count`: Total number of items in pagination
- `X-Page-Count`: Number of pages in pagination
- `X-RateLimit-Limit`: Rate limit maximum
- `X-RateLimit-Remaining`: Remaining requests
- `X-RateLimit-Reset`: Rate limit reset time

### Custom Headers

```go
config := &middleware.CORSConfig{
    AllowedHeaders: []string{
        "Origin",
        "Content-Type",
        "Accept",
        "Authorization",
        "X-API-Key",
        "X-Custom-Header",  // Custom header
    },
    ExposedHeaders: []string{
        "X-Total-Count",
        "X-Custom-Response-Header",  // Custom response header
    },
}
```

## Security Considerations

### Best Practices

1. **Restrict Origins**
   ```go
   // Good: Specific origins
   AllowedOrigins: []string{"https://app.example.com"}
   
   // Avoid: Global wildcard in production
   AllowedOrigins: []string{"*"}
   ```

2. **Limit Methods**
   ```go
   // Good: Only necessary methods
   AllowedMethods: []string{"GET", "POST", "OPTIONS"}
   
   // Avoid: All methods unless needed
   AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
   ```

3. **Control Headers**
   ```go
   // Good: Only required headers
   AllowedHeaders: []string{"Origin", "Content-Type", "Authorization"}
   
   // Avoid: Wildcard headers
   AllowedHeaders: []string{"*"}
   ```

4. **Credentials Security**
   ```go
   // Good: Credentials with specific origins
   AllowCredentials: true
   AllowedOrigins: []string{"https://app.example.com"}
   
   // Avoid: Credentials with wildcard origins
   AllowCredentials: true
   AllowedOrigins: []string{"*"}
   ```

### Security Headers Integration

```go
// CORS and Security Headers work together
corsConfig := &middleware.CORSConfig{
    AllowedOrigins: []string{"https://app.example.com"},
    AllowCredentials: true,
}

securityConfig := &middleware.SecurityHeadersConfig{
    CSPEnabled:     true,
    HSTSEnabled:    true,
    FrameOptions:   "DENY",
    XSSProtection:  "1; mode=block",
}

// Apply both middlewares
handler := corsMiddleware.Middleware(
    securityMiddleware.Middleware(mainHandler),
)
```

## Environment-Specific Configurations

### Development Environment

```go
func getDevelopmentCORSConfig() *middleware.CORSConfig {
    return &middleware.CORSConfig{
        AllowedOrigins: []string{
            "http://localhost:3000",
            "https://localhost:3000",
            "http://localhost:8080",
            "https://localhost:8080",
            "http://127.0.0.1:3000",
        },
        AllowAllOrigins:  false,
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
        AllowedHeaders: []string{
            "Origin",
            "Content-Type",
            "Accept",
            "Authorization",
            "X-Requested-With",
        },
        ExposedHeaders: []string{
            "X-Total-Count",
            "X-Page-Count",
            "X-RateLimit-Limit",
            "X-RateLimit-Remaining",
        },
        AllowCredentials: true,
        MaxAge:           300 * time.Second, // 5 minutes
        Debug:            true,
    }
}
```

### Staging Environment

```go
func getStagingCORSConfig() *middleware.CORSConfig {
    return &middleware.CORSConfig{
        AllowedOrigins: []string{
            "https://staging-app.example.com",
            "https://staging-admin.example.com",
            "https://*.staging.example.com",
        },
        AllowAllOrigins:  false,
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders: []string{
            "Origin",
            "Content-Type",
            "Accept",
            "Authorization",
            "X-API-Key",
        },
        ExposedHeaders: []string{
            "X-Total-Count",
            "X-RateLimit-Limit",
            "X-RateLimit-Remaining",
        },
        AllowCredentials: true,
        MaxAge:           1800 * time.Second, // 30 minutes
        Debug:            false,
    }
}
```

### Production Environment

```go
func getProductionCORSConfig() *middleware.CORSConfig {
    return &middleware.CORSConfig{
        AllowedOrigins: []string{
            "https://app.example.com",
            "https://admin.example.com",
            "https://api.example.com",
        },
        AllowAllOrigins:  false,
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders: []string{
            "Origin",
            "Content-Type",
            "Accept",
            "Authorization",
            "X-API-Key",
        },
        ExposedHeaders: []string{
            "X-Total-Count",
            "X-RateLimit-Limit",
            "X-RateLimit-Remaining",
        },
        AllowCredentials: true,
        MaxAge:           3600 * time.Second, // 1 hour
        Debug:            false,
        
        // Path-specific rules for production
        PathRules: []middleware.CORSPathRule{
            {
                Path:            "/api/public",
                AllowedOrigins:  []string{"*"},
                AllowedMethods:  []string{"GET", "OPTIONS"},
                AllowCredentials: false,
                MaxAge:          300 * time.Second,
            },
        },
    }
}
```

## Testing

### Unit Testing

```go
func TestCORSMiddleware(t *testing.T) {
    logger := zap.NewNop()
    config := &middleware.CORSConfig{
        AllowedOrigins:   []string{"https://example.com"},
        AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
        AllowedHeaders:   []string{"Origin", "Content-Type"},
        AllowCredentials: true,
    }
    
    middleware := middleware.NewCORSMiddleware(config, logger)
    
    // Test allowed origin
    req := httptest.NewRequest("OPTIONS", "/api/test", nil)
    req.Header.Set("Origin", "https://example.com")
    
    recorder := httptest.NewRecorder()
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    })
    
    middleware.Middleware(handler).ServeHTTP(recorder, req)
    
    assert.Equal(t, http.StatusNoContent, recorder.Code)
    assert.Equal(t, "https://example.com", recorder.Header().Get("Access-Control-Allow-Origin"))
}
```

### Integration Testing

```go
func TestCORSIntegration(t *testing.T) {
    // Setup test server with CORS middleware
    logger := zap.NewNop()
    corsConfig := middleware.GetDevelopmentCORSConfig()
    corsMiddleware := middleware.NewCORSMiddleware(corsConfig, logger)
    
    router := http.NewServeMux()
    router.HandleFunc("/api/test", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("success"))
    })
    
    handler := corsMiddleware.Middleware(router)
    server := httptest.NewServer(handler)
    defer server.Close()
    
    // Test preflight request
    req, _ := http.NewRequest("OPTIONS", server.URL+"/api/test", nil)
    req.Header.Set("Origin", "http://localhost:3000")
    req.Header.Set("Access-Control-Request-Method", "POST")
    req.Header.Set("Access-Control-Request-Headers", "Content-Type")
    
    client := &http.Client{}
    resp, err := client.Do(req)
    require.NoError(t, err)
    defer resp.Body.Close()
    
    assert.Equal(t, http.StatusNoContent, resp.StatusCode)
    assert.Equal(t, "http://localhost:3000", resp.Header.Get("Access-Control-Allow-Origin"))
}
```

## Troubleshooting

### Common Issues

1. **CORS Errors in Browser**
   ```
   Access to fetch at 'https://api.example.com/data' from origin 'https://app.example.com' 
   has been blocked by CORS policy
   ```
   
   **Solution**: Ensure the origin is in `AllowedOrigins`:
   ```go
   AllowedOrigins: []string{"https://app.example.com"}
   ```

2. **Credentials Not Sent**
   ```
   Credentials flag is 'true', but the 'Access-Control-Allow-Origin' header is '*'
   ```
   
   **Solution**: Use specific origins instead of wildcard when using credentials:
   ```go
   AllowCredentials: true
   AllowedOrigins: []string{"https://app.example.com"} // Not "*"
   ```

3. **Custom Headers Blocked**
   ```
   Request header field X-Custom-Header is not allowed by Access-Control-Allow-Headers
   ```
   
   **Solution**: Add the header to `AllowedHeaders`:
   ```go
   AllowedHeaders: []string{"Origin", "Content-Type", "X-Custom-Header"}
   ```

4. **Preflight Cache Issues**
   ```
   Preflight requests are cached for too long
   ```
   
   **Solution**: Reduce `MaxAge` for development:
   ```go
   MaxAge: 300 * time.Second // 5 minutes instead of 1 hour
   ```

### Debug Mode

Enable debug mode to see detailed CORS logs:

```go
config := &middleware.CORSConfig{
    // ... other config
    Debug: true,
}
```

Debug logs will show:
- Origin validation results
- Preflight request handling
- Header processing
- Path rule matching

### Monitoring

Monitor CORS activity through logs:

```go
// CORS middleware logs
logger.Info("CORS request processed",
    zap.String("origin", origin),
    zap.String("method", r.Method),
    zap.String("path", r.URL.Path),
    zap.Bool("allowed", allowed),
)
```

## Performance Considerations

### Caching Strategy

1. **Preflight Caching**
   ```go
   // Longer cache for stable APIs
   MaxAge: 3600 * time.Second // 1 hour
   
   // Shorter cache for development
   MaxAge: 300 * time.Second // 5 minutes
   ```

2. **Path-Based Optimization**
   ```go
   PathRules: []middleware.CORSPathRule{
       {
           Path: "/api/static", // Static content
           MaxAge: 86400 * time.Second, // 24 hours
       },
       {
           Path: "/api/dynamic", // Dynamic content
           MaxAge: 300 * time.Second, // 5 minutes
       },
   }
   ```

### Memory Usage

- CORS middleware has minimal memory footprint
- Configuration is loaded once at startup
- No per-request state storage
- Path rules are evaluated efficiently

## Migration Guide

### From Basic CORS

If you're migrating from a basic CORS implementation:

1. **Replace manual header setting**:
   ```go
   // Old way
   w.Header().Set("Access-Control-Allow-Origin", "*")
   w.Header().Set("Access-Control-Allow-Methods", "GET, POST")
   
   // New way
   corsMiddleware := middleware.NewCORSMiddleware(config, logger)
   handler := corsMiddleware.Middleware(yourHandler)
   ```

2. **Add configuration**:
   ```go
   config := &middleware.CORSConfig{
       AllowedOrigins: []string{"https://your-app.com"},
       AllowedMethods: []string{"GET", "POST", "OPTIONS"},
       AllowCredentials: true,
   }
   ```

3. **Test thoroughly**:
   - Test all origins
   - Test all methods
   - Test with credentials
   - Test preflight requests

### From Third-Party CORS Libraries

1. **Replace library imports**:
   ```go
   // Remove third-party CORS library
   // import "github.com/rs/cors"
   
   // Use our CORS middleware
   import "github.com/your-org/kyb-platform/internal/api/middleware"
   ```

2. **Update configuration**:
   ```go
   // Old third-party config
   cors := cors.New(cors.Options{
       AllowedOrigins: []string{"https://app.example.com"},
       AllowedMethods: []string{"GET", "POST"},
   })
   
   // New config
   config := &middleware.CORSConfig{
       AllowedOrigins: []string{"https://app.example.com"},
       AllowedMethods: []string{"GET", "POST", "OPTIONS"},
   }
   ```

3. **Update middleware chain**:
   ```go
   // Old way
   handler := cors.Handler(yourHandler)
   
   // New way
   corsMiddleware := middleware.NewCORSMiddleware(config, logger)
   handler := corsMiddleware.Middleware(yourHandler)
   ```

## Conclusion

The CORS policy implementation provides a comprehensive, secure, and flexible solution for handling cross-origin requests in the Enhanced Business Intelligence System. With configurable origins, methods, headers, and path-based rules, it supports various deployment scenarios while maintaining security best practices.

Key benefits:
- **Security**: Strict origin validation and method restrictions
- **Flexibility**: Path-based rules and environment-specific configurations
- **Performance**: Efficient preflight caching and minimal overhead
- **Maintainability**: Clean, testable code with comprehensive documentation
- **Integration**: Seamless integration with other security middleware

For production deployments, always use specific origins, limit methods and headers to what's necessary, and enable debug mode only in development environments.
