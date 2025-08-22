# Legacy Classification Service Migration Guide

## Overview

This guide helps developers migrate from the legacy `ClassificationService` to the new modular architecture. The new architecture provides better separation of concerns, improved testability, and enhanced scalability.

## What Changed

### Before (Legacy Architecture)
```go
// Monolithic service with all classification logic
service := classification.NewClassificationService(config, db, logger, metrics)
result := service.Classify(ctx, request)
```

### After (Modular Architecture)
```go
// Modular approach with specialized modules
router := routing.NewIntelligentRouter()
keywordModule := keyword_classification.NewKeywordClassificationModule()
mlModule := ml_classification.NewMLClassificationModule()
websiteModule := website_analysis.NewWebsiteAnalysisModule()
searchModule := web_search_analysis.NewWebSearchAnalysisModule()

// Router automatically selects the best module
result := router.Process(ctx, request)
```

## Migration Steps

### Step 1: Update Dependencies

**Remove legacy service imports:**
```go
// OLD
import "github.com/pcraw4d/business-verification/internal/classification"

// NEW
import (
    "github.com/pcraw4d/business-verification/internal/routing"
    "github.com/pcraw4d/business-verification/internal/modules/keyword_classification"
    "github.com/pcraw4d/business-verification/internal/modules/ml_classification"
    "github.com/pcraw4d/business-verification/internal/modules/website_analysis"
    "github.com/pcraw4d/business-verification/internal/modules/web_search_analysis"
)
```

### Step 2: Replace Service Initialization

**Replace legacy service creation:**
```go
// OLD
service := classification.NewClassificationService(config, db, logger, metrics)

// NEW
router := routing.NewIntelligentRouter()
router.RegisterModule(keyword_classification.NewKeywordClassificationModule())
router.RegisterModule(ml_classification.NewMLClassificationModule())
router.RegisterModule(website_analysis.NewWebsiteAnalysisModule())
router.RegisterModule(web_search_analysis.NewWebSearchAnalysisModule())
```

### Step 3: Update Method Calls

**Replace legacy method calls:**
```go
// OLD
result := service.Classify(ctx, request)

// NEW
result := router.Process(ctx, request)
```

### Step 4: Update Response Handling

**Update response structure:**
```go
// OLD
if result.PrimaryClassification != nil {
    industryCode := result.PrimaryClassification.IndustryCode
    confidence := result.PrimaryClassification.ConfidenceScore
}

// NEW
if result.Classification != nil {
    industryCode := result.Classification.IndustryCode
    confidence := result.Classification.ConfidenceScore
    module := result.ModuleUsed // New field showing which module was used
}
```

## Module-Specific Migration

### Keyword Classification

**Legacy Method:** `classifyByKeywords()`
**New Module:** `internal/modules/keyword_classification/`

```go
// OLD
classifications := service.classifyByKeywords(request)

// NEW
module := keyword_classification.NewKeywordClassificationModule()
classifications := module.Classify(ctx, request)
```

### Machine Learning Classification

**Legacy Method:** `classifyByML()`
**New Module:** `internal/modules/ml_classification/`

```go
// OLD
classifications := service.classifyByML(request)

// NEW
module := ml_classification.NewMLClassificationModule()
classifications := module.Classify(ctx, request)
```

### Website Analysis

**Legacy Method:** `classifyByWebsiteAnalysis()`
**New Module:** `internal/modules/website_analysis/`

```go
// OLD
classifications := service.classifyByWebsiteAnalysis(ctx, request)

// NEW
module := website_analysis.NewWebsiteAnalysisModule()
classifications := module.Classify(ctx, request)
```

### Web Search Analysis

**Legacy Method:** `classifyBySearchAnalysis()`
**New Module:** `internal/modules/web_search_analysis/`

```go
// OLD
classifications := service.classifyBySearchAnalysis(ctx, request)

// NEW
module := web_search_analysis.NewWebSearchAnalysisModule()
classifications := module.Classify(ctx, request)
```

## Configuration Changes

### Legacy Configuration
```go
config := &config.ExternalServicesConfig{
    // All configuration in one place
}
```

### New Modular Configuration
```go
// Each module has its own configuration
keywordConfig := keyword_classification.Config{
    // Keyword-specific settings
}

mlConfig := ml_classification.Config{
    // ML-specific settings
}

websiteConfig := website_analysis.Config{
    // Website analysis settings
}

searchConfig := web_search_analysis.Config{
    // Search analysis settings
}
```

## Error Handling

### Legacy Error Handling
```go
result, err := service.Classify(ctx, request)
if err != nil {
    // Handle error
}
```

### New Error Handling
```go
result, err := router.Process(ctx, request)
if err != nil {
    // Handle error with module-specific information
    if moduleErr, ok := err.(*routing.ModuleError); ok {
        log.Printf("Module %s failed: %v", moduleErr.ModuleName, moduleErr.Err)
    }
}
```

## Testing Migration

### Legacy Testing
```go
func TestClassification(t *testing.T) {
    service := classification.NewClassificationService(config, db, logger, metrics)
    result := service.Classify(ctx, request)
    // Assertions
}
```

### New Testing
```go
func TestClassification(t *testing.T) {
    router := routing.NewIntelligentRouter()
    router.RegisterModule(keyword_classification.NewKeywordClassificationModule())
    
    result := router.Process(ctx, request)
    // Assertions
}
```

## Performance Considerations

### Benefits of New Architecture
- **Parallel Processing**: Multiple modules can run concurrently
- **Resource Optimization**: Only necessary modules are executed
- **Caching**: Module-specific caching strategies
- **Load Balancing**: Intelligent routing based on module availability

### Migration Checklist

- [ ] Update import statements
- [ ] Replace service initialization
- [ ] Update method calls
- [ ] Update response handling
- [ ] Update configuration
- [ ] Update error handling
- [ ] Update tests
- [ ] Verify functionality
- [ ] Update documentation

## Troubleshooting

### Common Issues

1. **Module Not Found Errors**
   ```
   module not registered: keyword_classification
   ```
   **Solution**: Ensure all required modules are registered with the router

2. **Configuration Errors**
   ```
   invalid configuration for module: ml_classification
   ```
   **Solution**: Check module-specific configuration requirements

3. **Response Structure Errors**
   ```
   unknown field PrimaryClassification
   ```
   **Solution**: Update field references to use new response structure

### Debugging Tips

1. **Enable Module Logging**
   ```go
   router.SetLogLevel(logging.Debug)
   ```

2. **Check Module Health**
   ```go
   for _, module := range router.GetModules() {
       if !module.IsHealthy() {
           log.Printf("Module %s is unhealthy", module.Name())
       }
   }
   ```

3. **Monitor Performance**
   ```go
   metrics := router.GetMetrics()
   log.Printf("Average response time: %v", metrics.AverageResponseTime)
   ```

## Support

If you encounter issues during migration:

1. Check the [Modular Architecture Documentation](../api/modular-architecture.md)
2. Review the [Intelligent Routing Guide](../api/intelligent-routing.md)
3. Use the troubleshooting section above
4. Contact the development team for assistance

## Conclusion

The new modular architecture provides significant benefits in terms of maintainability, scalability, and performance. By following this guide, you can successfully migrate from the legacy classification service to the new modular approach while maintaining backward compatibility during the transition period.
