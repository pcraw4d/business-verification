# Classification Algorithm Root Cause Analysis

**Date**: 2025-11-10  
**Status**: Root Cause Identified

---

## Executive Summary

The classification service is **not using the actual classification algorithm**. Instead, it's using a **hardcoded placeholder function** that always returns "Food & Beverage" industry, regardless of the input business data.

---

## Root Cause

### Location

**File**: `services/classification-service/internal/handlers/classification.go`  
**Function**: `generateEnhancedClassification()` (lines 718-796)  
**Called from**: `processClassification()` (line 182)

### The Problem

The `generateEnhancedClassification()` function is a **placeholder/simulation** that always returns hardcoded "Food & Beverage" data:

```go
// Line 720-721: Comment indicates this is a placeholder
// For now, generate realistic data that simulates the unified classification approach
// In a full implementation, this would call the actual unified classifier

// Line 728: Hardcoded keywords
KeywordsExtracted: []string{"wine", "grape", "retail", "beverage", "store", "shop", "food", "drink"},

// Line 729: Hardcoded industry signals
IndustrySignals: []string{"food_beverage", "retail", "beverage_industry"},

// Line 736: Hardcoded industry
"industry": "Food & Beverage",

// Line 749: Hardcoded reasoning
reasoning := fmt.Sprintf("Primary industry identified as 'Food & Beverage' with 92%% confidence. ")

// Line 782: Hardcoded primary industry
PrimaryIndustry: "Food & Beverage",
```

### Why This Happens

1. The handler calls `generateEnhancedClassification(req)` which is a placeholder
2. The placeholder always returns "Food & Beverage" with hardcoded keywords and codes
3. The actual classification logic exists in `internal/classification/` but is **never called**

---

## Available Classification Services

The codebase contains actual classification implementations that are **not being used**:

### 1. UnifiedClassifier
- **Location**: `internal/classification/unified_classifier.go`
- **Purpose**: Unified classification using all available data sources
- **Method**: `ClassifyBusiness(ctx, input *ClassificationInput)`

### 2. MultiMethodClassifier
- **Location**: `internal/classification/multi_method_classifier.go`
- **Purpose**: Multi-method classification combining various approaches
- **Methods**: Multiple classification methods

### 3. IndustryDetectionService
- **Location**: `internal/classification/service.go`
- **Purpose**: Database-driven industry detection
- **Method**: `DetectIndustry(ctx, businessName, description, websiteURL)`

### 4. ClassificationCodeGenerator
- **Location**: `internal/classification/classifier.go`
- **Purpose**: Generates MCC, SIC, and NAICS codes
- **Method**: `GenerateClassificationCodes(ctx, keywords, detectedIndustry, confidence)`

### 5. Repository-Based Classification
- **Location**: `internal/classification/repository/supabase_repository.go`
- **Purpose**: Database-driven keyword matching and industry classification
- **Method**: `ClassifyBusiness(ctx, businessName, websiteURL)`

---

## The Fix

### Required Changes

1. **Replace the placeholder function** with actual classification logic
2. **Initialize classification services** in the handler (or pass them as dependencies)
3. **Call the actual classification methods** instead of returning hardcoded data

### Implementation Steps

#### Step 1: Update Handler Constructor

Add classification services to the handler:

```go
type ClassificationHandler struct {
    supabaseClient *supabase.Client
    logger         *zap.Logger
    config         *config.Config
    // ADD THESE:
    unifiedClassifier *classification.UnifiedClassifier
    // OR
    industryDetector  *classification.IndustryDetectionService
    codeGenerator      *classification.ClassificationCodeGenerator
}
```

#### Step 2: Replace generateEnhancedClassification

Replace the placeholder function with actual classification:

```go
func (h *ClassificationHandler) generateEnhancedClassification(ctx context.Context, req *ClassificationRequest) (*EnhancedClassificationResult, error) {
    // Option 1: Use UnifiedClassifier
    input := &classification.ClassificationInput{
        BusinessName: req.BusinessName,
        Description:  req.Description,
        WebsiteURL:   req.WebsiteURL,
    }
    
    result, err := h.unifiedClassifier.ClassifyBusiness(ctx, input)
    if err != nil {
        return nil, fmt.Errorf("classification failed: %w", err)
    }
    
    // Convert result to EnhancedClassificationResult
    // ...
    
    // Option 2: Use IndustryDetectionService + CodeGenerator
    industryResult, err := h.industryDetector.DetectIndustry(ctx, req.BusinessName, req.Description, req.WebsiteURL)
    if err != nil {
        return nil, fmt.Errorf("industry detection failed: %w", err)
    }
    
    codes, err := h.codeGenerator.GenerateClassificationCodes(
        ctx,
        industryResult.Keywords,
        industryResult.IndustryName,
        industryResult.Confidence,
    )
    if err != nil {
        return nil, fmt.Errorf("code generation failed: %w", err)
    }
    
    // Convert to EnhancedClassificationResult
    // ...
}
```

#### Step 3: Update processClassification

Update to handle errors and pass context:

```go
func (h *ClassificationHandler) processClassification(ctx context.Context, req *ClassificationRequest, startTime time.Time) (*ClassificationResponse, error) {
    // Call actual classification (not placeholder)
    enhancedResult, err := h.generateEnhancedClassification(ctx, req)
    if err != nil {
        return nil, fmt.Errorf("classification failed: %w", err)
    }
    
    // Rest of the function...
}
```

---

## Dependencies Required

### Repository Dependencies

The classification services need:
- `repository.KeywordRepository` - For keyword matching
- Database connection (Supabase) - For industry/code lookups

### Initialization

Services need to be initialized in `cmd/main.go`:

```go
// Initialize repository
keywordRepo := repository.NewSupabaseKeywordRepository(supabaseClient, logger)

// Initialize classification services
industryDetector := classification.NewIndustryDetectionService(keywordRepo, logger)
codeGenerator := classification.NewClassificationCodeGenerator(keywordRepo, logger)

// Or use UnifiedClassifier
unifiedClassifier := classification.NewUnifiedClassifier(keywordRepo, logger, ...)

// Pass to handler
handler := handlers.NewClassificationHandler(supabaseClient, logger, config, industryDetector, codeGenerator)
```

---

## Testing After Fix

### Test Cases to Verify

1. **Software Development Company**
   - Expected: Technology/Software industry
   - Expected Codes: MCC 5734/7372, NAICS 541511/541512, SIC 7371/7372

2. **Medical Clinic**
   - Expected: Healthcare industry
   - Expected Codes: MCC 8011, NAICS 621111, SIC 8011

3. **Restaurant Chain**
   - Expected: Food & Beverage industry
   - Expected Codes: MCC 5812, NAICS 722511, SIC 5812

4. **Financial Services**
   - Expected: Financial Services industry
   - Expected Codes: Financial services codes

5. **Retail Store**
   - Expected: Retail industry
   - Expected Codes: Retail codes

---

## Impact Assessment

### Current State
- ❌ **0% accuracy** for non-restaurant businesses
- ❌ All businesses classified as "Food & Beverage"
- ❌ Hardcoded keywords and codes
- ❌ No actual classification logic running

### After Fix
- ✅ **Expected >90% accuracy** for diverse business types
- ✅ Industry classification based on actual business data
- ✅ Dynamic keyword extraction and matching
- ✅ Proper MCC, SIC, NAICS code generation

---

## Priority

**CRITICAL** - This is a **core functionality blocker** for beta release. The classification service is completely non-functional for its primary purpose.

---

## Estimated Effort

- **Code Changes**: 2-4 hours
- **Testing**: 2-3 hours
- **Integration Testing**: 1-2 hours
- **Total**: 5-9 hours

---

## Recommendations

1. **Immediate Action**: Replace the placeholder function with actual classification logic
2. **Testing**: Test with diverse business types to verify accuracy
3. **Monitoring**: Add logging to track classification accuracy in production
4. **Documentation**: Update API documentation to reflect actual classification behavior

---

**Last Updated**: 2025-11-10 03:30 UTC

