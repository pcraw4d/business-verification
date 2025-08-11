# KYB Platform - Package Documentation

This document provides comprehensive documentation for all packages in the KYB Platform. Each package is documented with its purpose, API reference, usage examples, and best practices.

## Table of Contents

1. [Core Packages](#core-packages)
2. [API Packages](#api-packages)
3. [Service Packages](#service-packages)
4. [Database Packages](#database-packages)
5. [Utility Packages](#utility-packages)
6. [Configuration Packages](#configuration-packages)
7. [Observability Packages](#observability-packages)

## Core Packages

### Package: `cmd/api`

**Purpose**: Main application entry point for the KYB Platform API server.

**Location**: `cmd/api/`

**Description**: 
The `cmd/api` package contains the main application entry point that initializes and runs the KYB Platform API server. It handles server configuration, dependency injection, and graceful shutdown.

**Key Components**:
- `main.go`: Application entry point
- Server initialization and configuration
- Dependency injection setup
- Graceful shutdown handling

**Usage Example**:
```bash
# Run the API server
go run cmd/api/main.go

# Build the binary
go build -o kyb-api cmd/api/main.go

# Run with custom configuration
KYB_ENV=production ./kyb-api
```

**Configuration**:
```go
// Server configuration
type Server struct {
    config         *config.Config
    db             *database.Database
    authService    *auth.Service
    classification *classification.Service
    riskService    *risk.Service
    compliance     *compliance.Service
    logger         *logger.Logger
    metrics        *metrics.Metrics
}
```

**API Endpoints**:
- `GET /health` - Health check endpoint
- `GET /metrics` - Prometheus metrics endpoint
- `GET /docs` - API documentation
- `POST /v1/classify` - Business classification
- `POST /v1/risk/assess` - Risk assessment
- `POST /v1/compliance/check` - Compliance checking

### Package: `internal/config`

**Purpose**: Configuration management for the KYB Platform.

**Location**: `internal/config/`

**Description**:
The `internal/config` package provides centralized configuration management for the KYB Platform. It supports environment-based configuration, validation, and default values.

**Key Components**:
- `config.go`: Main configuration struct and loading logic
- `config_test.go`: Configuration tests
- Environment variable management
- Configuration validation

**Usage Example**:
```go
package main

import (
    "log"
    "github.com/kyb-platform/internal/config"
)

func main() {
    // Load configuration from environment
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }
    
    // Use configuration
    log.Printf("Server will run on port: %d", cfg.Server.Port)
    log.Printf("Database host: %s", cfg.Database.Host)
}
```

**Configuration Structure**:
```go
type Config struct {
    Server     ServerConfig     `mapstructure:"server"`
    Database   DatabaseConfig   `mapstructure:"database"`
    Redis      RedisConfig      `mapstructure:"redis"`
    Auth       AuthConfig       `mapstructure:"auth"`
    Logging    LoggingConfig    `mapstructure:"logging"`
    Metrics    MetricsConfig    `mapstructure:"metrics"`
    External   ExternalConfig   `mapstructure:"external"`
}
```

**Environment Variables**:
```bash
# Server configuration
KYB_SERVER_PORT=8080
KYB_SERVER_HOST=0.0.0.0
KYB_SERVER_TIMEOUT=30s

# Database configuration
KYB_DB_HOST=localhost
KYB_DB_PORT=5432
KYB_DB_NAME=kyb_platform
KYB_DB_USER=kyb_user
KYB_DB_PASSWORD=secure_password

# Redis configuration
KYB_REDIS_HOST=localhost
KYB_REDIS_PORT=6379
KYB_REDIS_DB=0

# Authentication
KYB_JWT_SECRET=your-jwt-secret
KYB_JWT_EXPIRY=24h
```

## API Packages

### Package: `internal/api/handlers`

**Purpose**: HTTP request handlers for the KYB Platform API.

**Location**: `internal/api/handlers/`

**Description**:
The `internal/api/handlers` package contains HTTP request handlers that process incoming API requests and return appropriate responses. Each handler is responsible for a specific API endpoint.

**Key Components**:
- `classification.go`: Business classification endpoints
- `risk.go`: Risk assessment endpoints
- `compliance.go`: Compliance checking endpoints
- `auth.go`: Authentication endpoints
- `admin.go`: Administrative endpoints
- `dashboard.go`: Dashboard and analytics endpoints

**Usage Example**:
```go
package main

import (
    "net/http"
    "github.com/kyb-platform/internal/api/handlers"
    "github.com/kyb-platform/internal/classification"
)

func main() {
    // Initialize services
    classificationService := classification.NewService()
    
    // Create handlers
    classificationHandler := handlers.NewClassificationHandler(classificationService)
    
    // Set up routes
    mux := http.NewServeMux()
    mux.HandleFunc("/v1/classify", classificationHandler.Classify)
    mux.HandleFunc("/v1/classify/batch", classificationHandler.ClassifyBatch)
    
    // Start server
    http.ListenAndServe(":8080", mux)
}
```

**Handler Interface**:
```go
type ClassificationHandler struct {
    service *classification.Service
    logger  *logger.Logger
}

func (h *ClassificationHandler) Classify(w http.ResponseWriter, r *http.Request) {
    // Parse request
    var req ClassificationRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    // Validate request
    if err := req.Validate(); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // Process classification
    result, err := h.service.Classify(r.Context(), &req)
    if err != nil {
        h.logger.Error("Classification failed", "error", err)
        http.Error(w, "Classification failed", http.StatusInternalServerError)
        return
    }
    
    // Return response
    json.NewEncoder(w).Encode(result)
}
```

### Package: `internal/api/middleware`

**Purpose**: HTTP middleware for authentication, logging, and request processing.

**Location**: `internal/api/middleware/`

**Description**:
The `internal/api/middleware` package provides HTTP middleware components for common functionality such as authentication, logging, rate limiting, and request validation.

**Key Components**:
- `auth.go`: Authentication middleware
- `logging.go`: Request logging middleware
- `rate_limit.go`: Rate limiting middleware
- `validation.go`: Request validation middleware
- `cors.go`: CORS middleware
- `recovery.go`: Panic recovery middleware

**Usage Example**:
```go
package main

import (
    "net/http"
    "github.com/kyb-platform/internal/api/middleware"
)

func main() {
    // Create middleware chain
    authMiddleware := middleware.Auth(authService)
    loggingMiddleware := middleware.Logging(logger)
    rateLimitMiddleware := middleware.RateLimit(100, time.Minute)
    
    // Apply middleware to routes
    handler := http.HandlerFunc(classificationHandler.Classify)
    handler = authMiddleware(handler)
    handler = loggingMiddleware(handler)
    handler = rateLimitMiddleware(handler)
    
    mux.HandleFunc("/v1/classify", handler)
}
```

**Middleware Interface**:
```go
type Middleware func(http.Handler) http.Handler

func Auth(authService *auth.Service) Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Extract token from request
            token := extractToken(r)
            if token == "" {
                http.Error(w, "Missing authentication token", http.StatusUnauthorized)
                return
            }
            
            // Validate token
            claims, err := authService.ValidateToken(token)
            if err != nil {
                http.Error(w, "Invalid authentication token", http.StatusUnauthorized)
                return
            }
            
            // Add claims to request context
            ctx := context.WithValue(r.Context(), "user", claims)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

## Service Packages

### Package: `internal/classification`

**Purpose**: Business classification functionality using industry-standard codes.

**Location**: `internal/classification/`

**Description**:
The `internal/classification` package provides business classification functionality using industry-standard codes (NAICS, SIC, MCC). It supports multiple classification methods including keyword-based, fuzzy matching, and hybrid approaches.

**Key Components**:
- `service.go`: Main classification service
- `fuzzy.go`: Fuzzy matching algorithms
- `normalize.go`: Text normalization utilities
- `data_loader.go`: Industry code data loading
- `mapping.go`: Code mapping and crosswalk logic

**Usage Example**:
```go
package main

import (
    "context"
    "github.com/kyb-platform/internal/classification"
)

func main() {
    // Initialize classification service
    service := classification.NewService()
    
    // Classify a business
    result, err := service.Classify(context.Background(), &classification.Request{
        BusinessName: "Acme Software Solutions Inc.",
        Address:      "123 Tech Street, San Francisco, CA 94105",
        Website:      "https://acmesoftware.com",
    })
    if err != nil {
        log.Fatalf("Classification failed: %v", err)
    }
    
    fmt.Printf("NAICS Code: %s\n", result.PrimaryClassification.NAICSCode)
    fmt.Printf("Confidence: %.2f\n", result.ConfidenceScore)
}
```

**Service Interface**:
```go
type Service struct {
    dataLoader    *DataLoader
    fuzzyMatcher  *FuzzyMatcher
    normalizer    *Normalizer
    logger        *logger.Logger
}

func (s *Service) Classify(ctx context.Context, req *Request) (*Result, error) {
    // Validate request
    if err := req.Validate(); err != nil {
        return nil, fmt.Errorf("invalid request: %w", err)
    }
    
    // Normalize business name
    normalizedName := s.normalizer.Normalize(req.BusinessName)
    
    // Perform classification using multiple methods
    keywordResult := s.classifyByKeywords(normalizedName)
    fuzzyResult := s.classifyByFuzzy(normalizedName)
    
    // Combine results using hybrid approach
    result := s.combineResults(keywordResult, fuzzyResult)
    
    // Calculate confidence score
    result.ConfidenceScore = s.calculateConfidence(result)
    
    return result, nil
}
```

### Package: `internal/risk`

**Purpose**: Risk assessment and analysis functionality.

**Location**: `internal/risk/`

**Description**:
The `internal/risk` package provides comprehensive risk assessment functionality for businesses. It evaluates multiple risk factors including financial, operational, compliance, and market risks.

**Key Components**:
- `service.go`: Main risk assessment service
- `scoring.go`: Risk scoring algorithms
- `models.go`: Risk data models
- `categories.go`: Risk category definitions
- `thresholds.go`: Risk threshold management

**Usage Example**:
```go
package main

import (
    "context"
    "github.com/kyb-platform/internal/risk"
)

func main() {
    // Initialize risk service
    service := risk.NewService()
    
    // Assess risk for a business
    assessment, err := service.AssessRisk(context.Background(), &risk.AssessmentRequest{
        BusinessID:      "business-123",
        AssessmentType:  "comprehensive",
        IncludeFactors:  true,
    })
    if err != nil {
        log.Fatalf("Risk assessment failed: %v", err)
    }
    
    fmt.Printf("Risk Level: %s\n", assessment.RiskLevel)
    fmt.Printf("Risk Score: %.2f\n", assessment.OverallRiskScore)
    
    // Print risk factors
    for _, factor := range assessment.RiskFactors {
        fmt.Printf("- %s: %.2f\n", factor.Name, factor.Score)
    }
}
```

**Risk Assessment Process**:
```go
func (s *Service) AssessRisk(ctx context.Context, req *AssessmentRequest) (*Assessment, error) {
    // Validate request
    if err := req.Validate(); err != nil {
        return nil, fmt.Errorf("invalid request: %w", err)
    }
    
    // Gather business data
    business, err := s.getBusinessData(ctx, req.BusinessID)
    if err != nil {
        return nil, fmt.Errorf("failed to get business data: %w", err)
    }
    
    // Assess different risk categories
    financialRisk := s.assessFinancialRisk(business)
    operationalRisk := s.assessOperationalRisk(business)
    complianceRisk := s.assessComplianceRisk(business)
    marketRisk := s.assessMarketRisk(business)
    
    // Calculate overall risk score
    overallScore := s.calculateOverallRisk(financialRisk, operationalRisk, complianceRisk, marketRisk)
    
    // Determine risk level
    riskLevel := s.determineRiskLevel(overallScore)
    
    return &Assessment{
        BusinessID:      req.BusinessID,
        OverallRiskScore: overallScore,
        RiskLevel:       riskLevel,
        RiskFactors:     []RiskFactor{financialRisk, operationalRisk, complianceRisk, marketRisk},
        CreatedAt:       time.Now(),
    }, nil
}
```

### Package: `internal/compliance`

**Purpose**: Compliance checking and framework management.

**Location**: `internal/compliance/`

**Description**:
The `internal/compliance` package provides compliance checking functionality for various regulatory frameworks including SOC 2, PCI DSS, GDPR, and regional compliance requirements.

**Key Components**:
- `service.go`: Main compliance service
- `frameworks/`: Framework-specific implementations
- `check_engine.go`: Compliance checking engine
- `reporting.go`: Compliance report generation
- `tracking.go`: Compliance status tracking

**Usage Example**:
```go
package main

import (
    "context"
    "github.com/kyb-platform/internal/compliance"
)

func main() {
    // Initialize compliance service
    service := compliance.NewService()
    
    // Check compliance for a business
    result, err := service.CheckCompliance(context.Background(), &compliance.CheckRequest{
        BusinessID: "business-123",
        Frameworks: []string{"soc2", "pci_dss", "gdpr"},
    })
    if err != nil {
        log.Fatalf("Compliance check failed: %v", err)
    }
    
    fmt.Printf("Overall Compliance: %.2f%%\n", result.OverallComplianceScore*100)
    
    // Print framework results
    for framework, score := range result.FrameworkResults {
        fmt.Printf("- %s: %.2f%%\n", framework, score.ComplianceScore*100)
    }
}
```

**Compliance Checking Process**:
```go
func (s *Service) CheckCompliance(ctx context.Context, req *CheckRequest) (*CheckResult, error) {
    // Validate request
    if err := req.Validate(); err != nil {
        return nil, fmt.Errorf("invalid request: %w", err)
    }
    
    // Get business data
    business, err := s.getBusinessData(ctx, req.BusinessID)
    if err != nil {
        return nil, fmt.Errorf("failed to get business data: %w", err)
    }
    
    // Check each framework
    frameworkResults := make(map[string]*FrameworkResult)
    for _, framework := range req.Frameworks {
        result, err := s.checkFramework(ctx, business, framework)
        if err != nil {
            s.logger.Error("Framework check failed", "framework", framework, "error", err)
            continue
        }
        frameworkResults[framework] = result
    }
    
    // Calculate overall compliance score
    overallScore := s.calculateOverallCompliance(frameworkResults)
    
    return &CheckResult{
        BusinessID:            req.BusinessID,
        OverallComplianceScore: overallScore,
        FrameworkResults:      frameworkResults,
        CheckedAt:            time.Now(),
    }, nil
}
```

## Database Packages

### Package: `internal/database`

**Purpose**: Database operations and data access layer.

**Location**: `internal/database/`

**Description**:
The `internal/database` package provides database operations and data access functionality for the KYB Platform. It includes models, migrations, and database utilities.

**Key Components**:
- `models.go`: Database models and schemas
- `postgres.go`: PostgreSQL database implementation
- `migrations/`: Database migration files
- `factory.go`: Database factory for instantiation
- `seeds.go`: Database seeding utilities

**Usage Example**:
```go
package main

import (
    "context"
    "github.com/kyb-platform/internal/database"
)

func main() {
    // Initialize database
    db, err := database.NewDatabase(&database.Config{
        Host:     "localhost",
        Port:     5432,
        Name:     "kyb_platform",
        User:     "kyb_user",
        Password: "secure_password",
    })
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    defer db.Close()
    
    // Create a business record
    business := &database.Business{
        Name:    "Acme Corporation",
        Address: "123 Business St, New York, NY 10001",
        Website: "https://acme.com",
    }
    
    err = db.CreateBusiness(context.Background(), business)
    if err != nil {
        log.Fatalf("Failed to create business: %v", err)
    }
    
    fmt.Printf("Created business with ID: %s\n", business.ID)
}
```

**Database Models**:
```go
type Business struct {
    ID          string    `db:"id" json:"id"`
    Name        string    `db:"name" json:"name"`
    Address     string    `db:"address" json:"address"`
    Website     string    `db:"website" json:"website"`
    Industry    string    `db:"industry" json:"industry"`
    CreatedAt   time.Time `db:"created_at" json:"created_at"`
    UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

type Classification struct {
    ID                    string    `db:"id" json:"id"`
    BusinessID           string    `db:"business_id" json:"business_id"`
    NAICSCode            string    `db:"naics_code" json:"naics_code"`
    SICCode              string    `db:"sic_code" json:"sic_code"`
    MCCCode              string    `db:"mcc_code" json:"mcc_code"`
    ConfidenceScore      float64   `db:"confidence_score" json:"confidence_score"`
    ClassificationMethod string    `db:"classification_method" json:"classification_method"`
    CreatedAt            time.Time `db:"created_at" json:"created_at"`
}

type RiskAssessment struct {
    ID              string    `db:"id" json:"id"`
    BusinessID     string    `db:"business_id" json:"business_id"`
    RiskScore      float64   `db:"risk_score" json:"risk_score"`
    RiskLevel      string    `db:"risk_level" json:"risk_level"`
    AssessmentType string    `db:"assessment_type" json:"assessment_type"`
    CreatedAt      time.Time `db:"created_at" json:"created_at"`
}
```

## Utility Packages

### Package: `pkg/validators`

**Purpose**: Input validation utilities.

**Location**: `pkg/validators/`

**Description**:
The `pkg/validators` package provides input validation utilities for the KYB Platform. It includes validation functions for common data types and business rules.

**Key Components**:
- `sanitize.go`: Input sanitization utilities
- `validation.go`: Validation functions
- `business.go`: Business-specific validation rules

**Usage Example**:
```go
package main

import (
    "github.com/kyb-platform/pkg/validators"
)

func main() {
    // Validate business name
    businessName := "Acme Corporation"
    if err := validators.ValidateBusinessName(businessName); err != nil {
        log.Printf("Invalid business name: %v", err)
        return
    }
    
    // Sanitize input
    sanitizedName := validators.SanitizeBusinessName(businessName)
    fmt.Printf("Sanitized name: %s\n", sanitizedName)
    
    // Validate email
    email := "user@example.com"
    if err := validators.ValidateEmail(email); err != nil {
        log.Printf("Invalid email: %v", err)
        return
    }
}
```

**Validation Functions**:
```go
// ValidateBusinessName validates a business name
func ValidateBusinessName(name string) error {
    if strings.TrimSpace(name) == "" {
        return ErrEmptyBusinessName
    }
    
    if len(strings.TrimSpace(name)) < 2 {
        return ErrBusinessNameTooShort
    }
    
    if len(name) > 200 {
        return ErrBusinessNameTooLong
    }
    
    // Check for invalid characters
    if !businessNameRegex.MatchString(name) {
        return ErrInvalidBusinessNameCharacters
    }
    
    return nil
}

// SanitizeBusinessName sanitizes a business name
func SanitizeBusinessName(name string) string {
    // Remove extra whitespace
    name = strings.TrimSpace(name)
    
    // Normalize whitespace
    name = strings.Join(strings.Fields(name), " ")
    
    // Remove invalid characters
    name = businessNameRegex.ReplaceAllString(name, "")
    
    return name
}
```

### Package: `pkg/encryption`

**Purpose**: Encryption and security utilities.

**Location**: `pkg/encryption/`

**Description**:
The `pkg/encryption` package provides encryption and security utilities for the KYB Platform. It includes functions for hashing, encryption, and secure key management.

**Key Components**:
- `hash.go`: Password hashing utilities
- `encrypt.go`: Data encryption utilities
- `keys.go`: Key management utilities

**Usage Example**:
```go
package main

import (
    "github.com/kyb-platform/pkg/encryption"
)

func main() {
    // Hash a password
    password := "secure_password"
    hashedPassword, err := encryption.HashPassword(password)
    if err != nil {
        log.Fatalf("Failed to hash password: %v", err)
    }
    
    // Verify password
    if encryption.VerifyPassword(password, hashedPassword) {
        fmt.Println("Password is valid")
    } else {
        fmt.Println("Password is invalid")
    }
    
    // Encrypt sensitive data
    sensitiveData := "sensitive_information"
    encryptedData, err := encryption.Encrypt(sensitiveData, encryptionKey)
    if err != nil {
        log.Fatalf("Failed to encrypt data: %v", err)
    }
    
    // Decrypt data
    decryptedData, err := encryption.Decrypt(encryptedData, encryptionKey)
    if err != nil {
        log.Fatalf("Failed to decrypt data: %v", err)
    }
    
    fmt.Printf("Decrypted data: %s\n", decryptedData)
}
```

## Configuration Packages

### Package: `configs`

**Purpose**: Configuration files and templates.

**Location**: `configs/`

**Description**:
The `configs` package contains configuration files and templates for different environments (development, staging, production).

**Key Components**:
- `development.env`: Development environment configuration
- `production.env`: Production environment configuration
- `staging.env`: Staging environment configuration

**Configuration Files**:
```bash
# Development environment (configs/development.env)
KYB_ENV=development
KYB_SERVER_PORT=8080
KYB_SERVER_HOST=localhost
KYB_DB_HOST=localhost
KYB_DB_PORT=5432
KYB_DB_NAME=kyb_platform_dev
KYB_REDIS_HOST=localhost
KYB_REDIS_PORT=6379
KYB_LOG_LEVEL=debug
```

```bash
# Production environment (configs/production.env)
KYB_ENV=production
KYB_SERVER_PORT=443
KYB_SERVER_HOST=0.0.0.0
KYB_DB_HOST=kyb-db.production.com
KYB_DB_PORT=5432
KYB_DB_NAME=kyb_platform_prod
KYB_REDIS_HOST=kyb-redis.production.com
KYB_REDIS_PORT=6379
KYB_LOG_LEVEL=info
KYB_JWT_SECRET=your-production-jwt-secret
```

## Observability Packages

### Package: `internal/observability`

**Purpose**: Logging, metrics, and tracing functionality.

**Location**: `internal/observability/`

**Description**:
The `internal/observability` package provides comprehensive observability functionality including structured logging, metrics collection, and distributed tracing.

**Key Components**:
- `logger.go`: Structured logging implementation
- `metrics.go`: Prometheus metrics collection
- `tracing.go`: OpenTelemetry tracing
- `health.go`: Health check endpoints
- `request_id.go`: Request ID propagation

**Usage Example**:
```go
package main

import (
    "context"
    "github.com/kyb-platform/internal/observability"
)

func main() {
    // Initialize logger
    logger := observability.NewLogger(&observability.LoggerConfig{
        Level:  "info",
        Format: "json",
    })
    
    // Initialize metrics
    metrics := observability.NewMetrics()
    
    // Initialize tracing
    tracer := observability.NewTracer("kyb-platform")
    
    // Create a span for an operation
    ctx, span := tracer.Start(context.Background(), "business_classification")
    defer span.End()
    
    // Log structured information
    logger.Info("Processing business classification",
        "business_name", "Acme Corporation",
        "request_id", "req-123",
        "user_id", "user-456",
    )
    
    // Record metrics
    metrics.IncCounter("classifications_total", map[string]string{
        "status": "success",
        "method": "api",
    })
    
    // Add span attributes
    span.SetAttributes(
        attribute.String("business.name", "Acme Corporation"),
        attribute.String("classification.method", "hybrid"),
    )
}
```

**Observability Components**:
```go
// Logger interface
type Logger interface {
    Debug(msg string, fields ...Field)
    Info(msg string, fields ...Field)
    Warn(msg string, fields ...Field)
    Error(msg string, fields ...Field)
    Fatal(msg string, fields ...Field)
}

// Metrics interface
type Metrics interface {
    IncCounter(name string, labels map[string]string)
    SetGauge(name string, value float64, labels map[string]string)
    ObserveHistogram(name string, value float64, labels map[string]string)
}

// Tracer interface
type Tracer interface {
    Start(ctx context.Context, name string, opts ...SpanStartOption) (context.Context, Span)
}

// Health checker interface
type HealthChecker interface {
    Check(ctx context.Context) HealthStatus
    RegisterCheck(name string, check HealthCheck)
}
```

---

## Package Dependencies

### Dependency Graph

```
cmd/api
├── internal/config
├── internal/observability
├── internal/database
├── internal/auth
├── internal/classification
├── internal/risk
├── internal/compliance
└── internal/api/handlers

internal/api/handlers
├── internal/classification
├── internal/risk
├── internal/compliance
├── internal/auth
└── internal/observability

internal/classification
├── internal/database
├── pkg/validators
└── internal/observability

internal/risk
├── internal/database
├── internal/classification
└── internal/observability

internal/compliance
├── internal/database
├── internal/risk
└── internal/observability
```

### Import Guidelines

**Internal Packages**:
- Use `internal/` prefix for packages that should not be imported by external code
- Internal packages can import other internal packages
- Internal packages should not import external packages unless necessary

**Public Packages**:
- Use `pkg/` prefix for packages that can be imported by external code
- Public packages should have stable APIs
- Public packages should have comprehensive documentation

**Dependency Management**:
- Minimize dependencies between packages
- Use interfaces to decouple packages
- Avoid circular dependencies
- Use dependency injection for service dependencies

---

## Package Testing

### Testing Guidelines

**Unit Tests**:
- Each package should have comprehensive unit tests
- Test all exported functions and methods
- Use table-driven tests for multiple scenarios
- Mock external dependencies

**Integration Tests**:
- Test package interactions
- Test database operations
- Test API endpoints
- Use test databases and external services

**Example Test Structure**:
```go
package classification_test

import (
    "testing"
    "github.com/kyb-platform/internal/classification"
)

func TestService_Classify(t *testing.T) {
    tests := []struct {
        name        string
        businessName string
        expectedNAICS string
        expectError  bool
    }{
        {
            name:         "Software company",
            businessName: "Acme Software Solutions",
            expectedNAICS: "541511",
            expectError:  false,
        },
        {
            name:         "Empty business name",
            businessName: "",
            expectedNAICS: "",
            expectError:  true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            service := classification.NewService()
            result, err := service.Classify(context.Background(), &classification.Request{
                BusinessName: tt.businessName,
            })
            
            if tt.expectError {
                if err == nil {
                    t.Error("Expected error but got none")
                }
                return
            }
            
            if err != nil {
                t.Errorf("Unexpected error: %v", err)
                return
            }
            
            if result.PrimaryClassification.NAICSCode != tt.expectedNAICS {
                t.Errorf("Expected NAICS code %s, got %s", 
                    tt.expectedNAICS, result.PrimaryClassification.NAICSCode)
            }
        })
    }
}
```

---

## Package Documentation Standards

### Documentation Requirements

**Package Documentation**:
- Every package must have a package-level comment
- Package comments should explain the package's purpose and usage
- Include examples in package comments

**Function Documentation**:
- All exported functions must have GoDoc comments
- Include parameter descriptions, return values, and examples
- Document error conditions and side effects

**Type Documentation**:
- All exported types must have documentation
- Include field descriptions and usage examples
- Document interface contracts and implementations

**Example Documentation**:
```go
// Package classification provides business classification functionality.
//
// Example usage:
//
//	service := classification.NewService()
//	result, err := service.Classify(context.Background(), &classification.Request{
//	    BusinessName: "Acme Corporation",
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("NAICS Code: %s\n", result.PrimaryClassification.NAICSCode)
package classification
```

---

*Last updated: January 2024*
