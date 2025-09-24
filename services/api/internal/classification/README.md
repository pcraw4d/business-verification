# Classification Services Integration

This package provides a modular, database-driven classification system that can be integrated into the existing main API.

## 🏗️ Architecture

The classification system is built using Clean Architecture principles with the following layers:

- **Repository Layer** (`repository/`) - Data access and database operations
- **Service Layer** (`service.go`, `classifier.go`) - Business logic and classification algorithms
- **Container Layer** (`container.go`) - Dependency injection and service management
- **Integration Layer** (`integration.go`) - Simple interface for external integration

## 🚀 Quick Integration

### 1. Initialize the Integration Service

```go
import (
    "github.com/pcraw4d/business-verification/internal/classification"
    "github.com/pcraw4d/business-verification/internal/database"
)

// Initialize Supabase client
supabaseClient, err := database.NewSupabaseClient(config, logger)
if err != nil {
    log.Fatalf("Failed to initialize Supabase client: %v", err)
}

// Create integration service
classificationService := classification.NewIntegrationService(supabaseClient, logger)
defer classificationService.Close()
```

### 2. Use in Your API Handler

```go
func (h *YourHandler) ClassifyBusiness(w http.ResponseWriter, r *http.Request) {
    // Parse request
    var request struct {
        BusinessName string `json:"business_name"`
        Description  string `json:"description"`
        WebsiteURL   string `json:"website_url"`
    }
    
    if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
        http.Error(w, "Invalid request format", http.StatusBadRequest)
        return
    }
    
    // Process classification using new services
    result := classificationService.ProcessBusinessClassification(
        r.Context(),
        request.BusinessName,
        request.Description,
        request.WebsiteURL,
    )
    
    // Return response
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}
```

### 3. Health Check Integration

```go
func (h *YourHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
    // Get classification services health
    classificationHealth := classificationService.GetHealthStatus()
    
    response := map[string]interface{}{
        "status": "healthy",
        "timestamp": time.Now().UTC().Format(time.RFC3339),
        "classification_services": classificationHealth,
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
```

## 🔧 Advanced Usage

### Using Individual Services

If you need more control, you can access individual services through the container:

```go
// Get the container
container := classification.NewClassificationContainer(supabaseClient, logger)

// Access individual services
industryDetectionService := container.GetIndustryDetectionService()
codeGenerator := container.GetCodeGenerator()

// Use services directly
result, err := industryDetectionService.DetectIndustryFromBusinessInfo(
    ctx,
    businessName,
    description,
    websiteURL,
)

codes, err := codeGenerator.GenerateClassificationCodes(
    ctx,
    keywords,
    detectedIndustry,
    confidence,
)
```

### Custom Repository Implementation

You can create custom repository implementations for testing or different data sources:

```go
// Create mock repository for testing
mockRepo := &MockKeywordRepository{}

// Create services with mock repository
industryService := classification.NewIndustryDetectionService(mockRepo, logger)
codeGen := classification.NewClassificationCodeGenerator(mockRepo, logger)
```

## 📊 Response Format

The integration service returns a structured response:

```json
{
  "success": true,
  "classification_data": {
    "industry_detection": {
      "detected_industry": "Technology",
      "confidence": 0.85,
      "keywords_matched": ["software", "platform", "digital"],
      "analysis_method": "Keyword-based detection",
      "evidence": "Keywords: software, platform, digital"
    },
    "classification_codes": {
      "mcc": [...],
      "sic": [...],
      "naics": [...]
    },
    "code_statistics": {
      "total_codes": 6,
      "mcc_count": 2,
      "sic_count": 2,
      "naics_count": 2,
      "avg_confidence": 0.85
    }
  },
  "enhanced_features": {
    "database_driven_classification": "active",
    "modular_architecture": "active"
  }
}
```

## 🧪 Testing

All services include comprehensive unit tests:

```bash
# Run all classification tests
go test ./internal/classification/ -v

# Run specific service tests
go test ./internal/classification/ -run TestIndustryDetectionService -v
go test ./internal/classification/ -run TestClassificationCodeGenerator -v
```

## 🔄 Migration from Old System

### Before (Hardcoded Logic)

```go
// Old hardcoded approach
func performRealKeywordClassification(businessName, description, websiteURL string) ClassificationResult {
    // Hardcoded industry detection logic
    if contains(businessName, "bank") || contains(businessName, "finance") {
        businessNameIndustry = "Financial Services"
        businessNameConfidence = 0.75
    }
    // ... more hardcoded logic
}
```

### After (Database-Driven)

```go
// New modular approach
result := classificationService.ProcessBusinessClassification(
    ctx,
    businessName,
    description,
    websiteURL,
)
```

## 🚨 Error Handling

The integration service includes comprehensive error handling:

- **Graceful fallbacks** - Returns default results if detection fails
- **Validation** - Validates classification codes for consistency
- **Logging** - Detailed logging for debugging and monitoring
- **Health checks** - Service health monitoring and status reporting

## 📈 Performance Features

- **Connection pooling** - Efficient database connection management
- **Caching ready** - Architecture supports future caching implementations
- **Async processing** - Services can be extended for background processing
- **Batch operations** - Support for processing multiple businesses at once

## 🔐 Security

- **Input validation** - All inputs are validated and sanitized
- **SQL injection protection** - Uses parameterized queries
- **Access control** - Repository layer supports Row-Level Security (RLS)
- **Audit logging** - Comprehensive logging for security monitoring

## 🚀 Future Enhancements

The modular architecture supports easy extension:

- **Machine Learning integration** - Add ML models for improved accuracy
- **Real-time updates** - Dynamic keyword weight updates
- **Multi-language support** - International business classification
- **Advanced analytics** - Business intelligence and reporting features
