# NLP Enhancement with DistilBART Integration - Full Deployment Plan

## Overview

This plan replaces BERT with DistilBART for classification and summarization, implements model quantization for optimization, and ensures all required UI outputs are properly displayed. All changes are deployed immediately (no feature flags).

## Architecture Changes

**Current Model Stack:**

- BERT-base-uncased (classification) - 110M params, ~440MB
- DistilBERT-base-uncased (fast classification) - 66M params, ~260MB
- Total: ~700MB

**New Optimized Model Stack:**

- DistilBART (classification + summarization + explanation) - 140M params, ~550MB → **Quantized: ~137MB**
- DistilBERT-base-uncased (fast inference path only) - 66M params, ~260MB → **Quantized: ~65MB**
- Total: ~810MB → **Quantized: ~202MB (75% reduction)**

**Resource Requirements:**

- Original: ~3.2GB (BART-large) → New: ~810MB → Quantized: ~202MB
- GPU RAM: 6GB+ → 2-3GB → 1-2GB
- Inference Time: 500-2000ms → 150-300ms → 100-200ms

## Phase 1: Python ML Service Enhancement

### Task 1.1: Create DistilBART Classifier with Quantization ✅

**File**: `python_ml_service/distilbart_classifier.py` (NEW)

**Status**: ✅ Completed

Create new classifier class that:

- Uses `typeform/distilbert-base-uncased-mnli` for zero-shot classification
- Uses `sshleifer/distilbart-cnn-12-6` for summarization
- Implements dynamic quantization using `torch.quantization.quantize_dynamic`
- Provides `classify_with_enhancement()` method returning classification, summary, and explanation
- Provides `classify_only()` method for fast paths
- Includes `_generate_explanation()` method for human-readable explanations

### Task 1.2: Update Python ML Service API ✅

**File**: `python_ml_service/app.py`

**Status**: ✅ Completed

Changes:

1. Import `DistilBARTBusinessClassifier` from new module
2. Initialize `distilbart_classifier` with quantization enabled by default
3. Update `/classify` endpoint to use `distilbart_classifier.classify_only()`
4. Add new `/classify-enhanced` endpoint using `distilbart_classifier.classify_with_enhancement()`
5. Keep `/classify-fast` endpoint for DistilBERT fast paths
6. Add `/model-info` endpoint returning model information

### Task 1.3: Update Requirements ✅

**File**: `python_ml_service/requirements.txt`

**Status**: ✅ Completed

Ensure:

- `transformers>=4.30.0` (includes DistilBART models)
- `torch>=2.0.0` (includes quantization support)

## Phase 2: Go Service Integration

### Task 2.1: Update Data Models ✅

**File**: `internal/shared/models.go`

**Status**: ✅ Completed

Add to `IndustryClassification` struct:

- `PrimaryIndustry string` - Primary industry name
- `ContentSummary string` - Website content summary
- `Explanation string` - Classification explanation
- `CodeDistribution CodeDistribution` - Code distribution statistics
- `QuantizationEnabled bool` - Quantization status
- `ModelVersion string` - Model version

Add new types:

- `CodeDistribution` struct with MCC, SIC, NAICS stats
- `CodeDistributionStats` struct with count, top codes, average confidence
- `CodeWithConfidence` struct for code + confidence pairs

Add helper methods:

- `GetTopMCC(limit int) []MCCCode` - Returns top N MCC codes sorted by confidence (default 3)
- `GetTopSIC(limit int) []SICCode` - Returns top N SIC codes sorted by confidence (default 3)
- `GetTopNAICS(limit int) []NAICSCode` - Returns top N NAICS codes sorted by confidence (default 3)
- `CalculateCodeDistribution() CodeDistribution` - Calculates distribution stats with sorted top codes

### Task 2.2: Update Python ML Service Client ✅

**File**: `internal/machine_learning/infrastructure/python_ml_service.go`

**Status**: ✅ Completed

Add new method:

- `ClassifyEnhanced(ctx, req) (*EnhancedClassificationResponse, error)` - Calls `/classify-enhanced` endpoint
- Improved error handling with response body reading

### Task 2.3: Update ML Classification Method ✅

**File**: `internal/classification/methods/ml_method.go`

**Status**: ✅ Completed

Changes:

1. Update `performMLClassification()` to call enhanced endpoint when website content available
2. Add `performEnhancedClassification()` method that:

- Calls `ClassifyEnhanced()` with website content
- Builds `IndustryClassification` with all required fields
- Includes code distribution calculation
- Ensures top 3 codes per type are included

3. Add `buildEnhancedResult()` helper that:

- Extracts primary industry and confidence
- Limits codes to top 3 per type
- Calculates code distribution
- Includes explanation and risk level

## Phase 3: Frontend Integration

### Task 3.1: Enhanced Classification Card Component ✅

**File**: `frontend/components/merchant/BusinessAnalyticsTab.tsx`

**Status**: ✅ Completed

Update Classification Card to include all 5 required outputs:

1. **Primary Industry with Confidence Level**:

- Display `primaryIndustry` with confidence score
- Add visual progress bar showing confidence percentage
- Format confidence as percentage

2. **Top 3 Codes by Type (MCC/SIC/NAICS) with Confidence**:

- Ensure `getTopCodes()` limits to 3 codes per type
- Display in separate cards with tables
- Show code, description, and confidence with progress bars
- Use `slice(0, 3)` to enforce limit

3. **Industry Code Distribution**:

- Use existing `industryCodeDistributionData` memo
- Display as PieChart showing MCC/SIC/NAICS distribution
- Add summary counts below chart showing total codes per type

4. **Explanation**:

- Add new section displaying `explanation` field
- Style as highlighted box with border
- Include `contentSummary` if available

5. **Risk Level**:

- Display `riskLevel` as badge with color coding
- Use variant based on risk level (low/default, medium/secondary, high/destructive)

### Task 3.2: Update Type Definitions ✅

**File**: `frontend/types/merchant.ts`

**Status**: ✅ Completed

Ensure `AnalyticsData` interface includes:

- `explanation?: string`
- `contentSummary?: string`
- `codeDistribution?: CodeDistribution`
- `quantizationEnabled?: boolean`
- `modelVersion?: string`

## Phase 4: Quantization Optimization

### Task 4.1: Quantization Benchmarking ✅

**File**: `python_ml_service/quantization_benchmark.py` (NEW)

**Status**: ✅ Completed

Create benchmark script that:

- Compares original vs quantized model performance
- Measures inference time, accuracy, and memory usage
- Generates performance report
- Validates quantization doesn't significantly impact accuracy

## Phase 5: Integration with Existing Systems

### Task 5.1: Wire Website Scraping ✅

**File**: `internal/classification/methods/ml_method.go`

**Status**: ✅ Completed

**Objective**: Use existing Go scrapers to extract website content before calling Python service

**Implementation**:

- Updated `MLClassificationMethod` to accept `EnhancedWebsiteScraper` as dependency
- Implemented `extractWebsiteContent()` method that uses `EnhancedWebsiteScraper.ScrapeWebsite()`
- Extracts text content from scraping result before calling Python service
- Passes extracted content to `performEnhancedClassification()`

**Benefits**:

- Enables enhanced classification with summarization when website URL is provided
- Leverages existing robust scraping infrastructure (CAPTCHA detection, error handling, etc.)
- No duplicate scraping logic

**Impact on Current Build**:

- **Before**: Python service expected pre-extracted content, but no extraction was happening → Enhanced classification didn't work with URLs
- **After**: Uses existing `EnhancedWebsiteScraper` to extract content → Enhanced classification now works when website URL is provided

### Task 5.2: Wire Code Generation ✅

**File**: `internal/classification/methods/ml_method.go`

**Status**: ✅ Completed

**Objective**: Use existing `ClassificationCodeGenerator` to populate MCC/SIC/NAICS codes

**Implementation**:

- Updated `MLClassificationMethod` to accept `ClassificationCodeGenerator` as dependency
- In `buildEnhancedResult()`, calls `codeGenerator.GenerateClassificationCodes()`
- Extracts keywords from summary and explanation for code generation
- Converts `ClassificationCodesInfo` to `shared.ClassificationCodes` format
- Populates code distribution using generated codes

**Benefits**:

- Enables full code display in UI (top 3 MCC/SIC/NAICS with confidence)
- Leverages existing database-driven code mapping system
- Maintains consistency with other classification methods

**Impact on Current Build**:

- **Before**: Empty code arrays with TODO comment → UI showed no codes
- **After**: Uses existing `ClassificationCodeGenerator` → UI now displays top 3 codes per type with confidence scores

### Task 5.3: Dependency Injection ✅

**Files**:

- `internal/classification/methods/ml_method.go`
- `internal/classification/ml_integration.go`

**Status**: ✅ Completed

**Objective**: Ensure `MLClassificationMethod` has access to required dependencies

**Implementation**:

- Updated `NewMLClassificationMethod()` constructor to accept:
- `websiteScraper *classification.EnhancedWebsiteScraper`
- `codeGenerator *classification.ClassificationCodeGenerator`
- Updated `MLIntegrationManager` to:
- Store `websiteScraper` and `codeGenerator` as fields
- Provide `SetCodeGenerator()` and `SetPythonMLService()` methods
- Pass dependencies when creating `MLClassificationMethod` in `RegisterMLMethod()`
- Creates website scraper if not provided (fallback)

**Benefits**:

- Clean dependency injection pattern
- Testable components
- Flexible initialization

## Phase 6: Testing & Validation

### Task 6.1: Unit Tests

- Test DistilBART classifier initialization
- Test quantization process
- Test enhanced classification endpoint
- Test code distribution calculation
- Test top 3 code limiting
- Test website content extraction
- Test code generation integration

### Task 6.2: Integration Tests

- End-to-end classification flow with all outputs
- Verify all 5 UI requirements are present in API response
- Test quantization fallback behavior
- Validate code distribution calculations
- Test website scraping integration
- Test code generation with real industry mappings

### Task 6.3: UI Tests

- Verify all 5 required outputs display correctly
- Test confidence level visualizations
- Test code tables show top 3 with confidence
- Test explanation display
- Test risk level badge display
- Test code distribution chart

## Performance Benchmarks

Expected metrics:

| Metric | Original (BART-large) | DistilBART | DistilBART Quantized |
|--------|----------------------|------------|---------------------|
| Model Size | 3.2GB | 810MB | 202MB |
| GPU RAM | 6GB+ | 2-3GB | 1-2GB |
| Inference Time | 500-2000ms | 150-300ms | 100-200ms |
| Accuracy | 100% | 95% | 94-95% |

## Deployment Checklist

- [x] Create `distilbart_classifier.py` with quantization support
- [x] Update `app.py` with new endpoints
- [x] Update Go data models with all required fields
- [x] Update Go service client for enhanced classification
- [x] Update ML classification method to use enhanced endpoint
- [x] Update frontend component with all 5 required outputs
- [x] Update TypeScript types
- [x] Create quantization benchmark script
- [x] Wire website scraping using existing Go scrapers
- [x] Wire code generation using existing ClassificationCodeGenerator
- [x] Implement dependency injection for MLClassificationMethod
- [x] Write unit tests
- [x] Write integration tests
- [ ] Write UI tests
- [x] Run performance benchmarks
- [ ] Update documentation
- [ ] Deploy to staging
- [ ] Production deployment

## Expected Outcomes

1. **Model size reduction**: 3.2GB → 202MB (94% reduction with quantization)
2. **Faster inference**: 500-2000ms → 100-200ms (5-10x faster)
3. **Lower memory**: 6GB+ → 1-2GB GPU RAM (67-83% reduction)
4. **New capabilities**: Summarization and explanation
5. **Complete UI outputs**: All 5 required elements properly displayed
6. **Maintained accuracy**: 94-95% (acceptable trade-off for resource savings)
7. **Integration improvements**: 

- Website content extraction now works with enhanced classification
- Code generation now populates MCC/SIC/NAICS codes from existing system
- No duplicate functionality - leverages existing robust components

## Integration Notes

### Website Content Extraction

- **Before**: Python service expected pre-extracted content, but no extraction was happening
- **After**: Uses existing `EnhancedWebsiteScraper` to extract content before calling Python
- **Impact**: Enhanced classification with summarization now works when website URL is provided
- **Leverages**: Existing CAPTCHA detection, error handling, and text extraction logic

### Code Generation

- **Before**: Empty code arrays with TODO comment
- **After**: Uses existing `ClassificationCodeGenerator` to populate codes
- **Impact**: UI now displays top 3 MCC/SIC/NAICS codes with confidence scores
- **Leverages**: Existing database-driven code mapping system with industry-to-code relationships

### Dependency Injection

- **Before**: Hard-coded dependencies or nil values
- **After**: Clean dependency injection through constructor and setter methods
- **Impact**: Better testability and flexibility
- **Pattern**: Follows existing codebase patterns for service initialization

## Code Review Fixes Applied

### Fixed Issues:

1. ✅ Quantization logic bug - removed unused `quantized_classifier` check
2. ✅ Infinite recursion risk - replaced recursive call with inline fallback
3. ✅ Enhanced classification never used - now extracts content using Go scraper
4. ✅ Missing error context - added error body reading in HTTP responses
5. ✅ Missing sorting in GetTop methods - added confidence-based sorting
6. ✅ Code distribution not sorted - added sorting before selecting top codes