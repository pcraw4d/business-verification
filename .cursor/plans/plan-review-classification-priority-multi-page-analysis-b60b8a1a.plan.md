<!-- b60b8a1a-ba14-43a9-bbdf-5fdc2b4b8059 1e3648be-d1c9-4263-8355-fb2f076c8089 -->
# Hybrid Code Generation Implementation Plan

## Overview

Transform the classification code generation from industry-only to a hybrid approach that combines:

1. Direct keyword-to-code matching via `code_keywords` table (new)
2. Industry-based code lookup (existing, enhanced)
3. Confidence-based filtering and ranking
4. Optional multi-industry code generation

## Current State Analysis

### Existing Implementation

- **File**: `internal/classification/classifier.go`
- **Method**: `GenerateClassificationCodes()` - currently only uses industry-based lookup
- **Method**: `generateCodesInParallel()` - generates MCC/SIC/NAICS codes by industry_id only
- **Limitation**: No direct keyword-to-code matching, single industry only

### Available but Unused Capabilities

1. **`code_keywords` table** - Schema exists, indexed, but no repository methods query it
2. **Enhanced scoring algorithm** - Has semantic boost enabled but not used for codes
3. **Multi-method classifier** - Can provide top N industries but only uses #1 for codes

## Implementation Plan

### Phase 1: Add code_keywords Repository Method

**File**: `internal/classification/repository/supabase_repository.go`

**Add new method**:

```go
// GetClassificationCodesByKeywords retrieves classification codes directly from keywords
// This bypasses industry detection and matches keywords to codes via code_keywords table
func (r *SupabaseKeywordRepository) GetClassificationCodesByKeywords(
    ctx context.Context,
    keywords []string,
    codeType string, // "MCC", "SIC", or "NAICS"
    minRelevance float64, // Minimum relevance_score threshold (default 0.5)
) ([]*ClassificationCode, error)
```

**Implementation details**:

- Query `code_keywords` table joined with `classification_codes`
- Match keywords using case-insensitive comparison
- Filter by `code_type` and `relevance_score >= minRelevance`
- Join with `classification_codes` to get full code details
- Order by `relevance_score DESC`, then by `code`
- Return codes with relevance scores for confidence calculation

**SQL Query Structure**:

```sql
SELECT DISTINCT cc.id, cc.industry_id, cc.code_type, cc.code, cc.description, 
       cc.is_active, ck.relevance_score, ck.match_type
FROM code_keywords ck
JOIN classification_codes cc ON ck.code_id = cc.id
WHERE LOWER(ck.keyword) = ANY($1) -- keywords array
  AND cc.code_type = $2
  AND cc.is_active = true
  AND ck.relevance_score >= $3
ORDER BY ck.relevance_score DESC, cc.code ASC
```

**Add to interface**: `internal/classification/repository/interface.go`

- Add method signature to `KeywordRepository` interface

### Phase 2: Implement Hybrid Code Generation

**File**: `internal/classification/classifier.go`

**Modify `generateCodesInParallel()` method**:

- Add parallel goroutine for keyword-based code lookup
- Keep existing industry-based lookup
- Merge results with confidence weighting
- Apply deduplication

**New helper method**:

```go
// generateCodesFromKeywords generates codes using direct keyword matching
func (g *ClassificationCodeGenerator) generateCodesFromKeywords(
    ctx context.Context,
    keywords []string,
    codeType string,
) ([]CodeMatch, error)
```

**New merge method**:

```go
// mergeCodeResults combines industry-based and keyword-based code results
func (g *ClassificationCodeGenerator) mergeCodeResults(
    industryCodes []*ClassificationCode,
    keywordCodes []CodeMatch,
    industryConfidence float64,
) []RankedCode
```

**CodeMatch struct** (new):

```go
type CodeMatch struct {
    Code            *ClassificationCode
    RelevanceScore  float64
    MatchType       string // "exact", "partial", "synonym"
    Source          string // "industry" or "keyword"
    Confidence      float64
}
```

**RankedCode struct** (new):

```go
type RankedCode struct {
    Code            *ClassificationCode
    CombinedConfidence float64
    Sources         []string // Which sources contributed
    MatchDetails    []CodeMatch
}
```

### Phase 3: Add Confidence Filtering and Ranking

**File**: `internal/classification/classifier.go`

**Add confidence calculation**:

- Industry-based codes: `confidence * 0.9` (existing)
- Keyword-based codes: `relevance_score * industry_confidence * 0.85`
- Combined codes: Weighted average of both sources

**Add filtering**:

- Filter codes below confidence threshold (default 0.6)
- Limit to top N codes per type (default: top 10 MCC, top 10 SIC, top 10 NAICS)
- Prioritize codes with `is_primary = true`

**Add ranking logic**:

1. Sort by combined confidence (descending)
2. Boost `is_primary` codes by 1.2x
3. Boost codes matched by both sources by 1.3x
4. Apply deduplication (same code from multiple sources)

### Phase 4: Optional Multi-Industry Support

**File**: `internal/classification/classifier.go`

**Modify `GenerateClassificationCodes()` signature** (optional enhancement):

```go
// Add optional parameter for multi-industry support
func (g *ClassificationCodeGenerator) GenerateClassificationCodes(
    ctx context.Context,
    keywords []string,
    detectedIndustry string,
    confidence float64,
    // New optional parameter:
    additionalIndustries []IndustryResult, // Top N industries from ensemble
) (*ClassificationCodesInfo, error)
```

**Implementation**:

- Generate codes for primary industry (existing)
- Generate codes for additional industries (if provided)
- Merge all results with weighted confidence
- Deduplicate across industries
- Rank by combined confidence

**Integration point**: `internal/classification/multi_method_classifier.go`

- Modify `calculateEnsembleResult()` to pass top 3 industries to code generator
- Or add separate method `GenerateCodesForMultipleIndustries()`

## Testing Requirements

### Unit Tests

**File**: `internal/classification/repository/supabase_repository_test.go`

- Test `GetClassificationCodesByKeywords()` with various keyword sets
- Test relevance score filtering
- Test code type filtering
- Test empty keyword array handling

**File**: `internal/classification/classifier_test.go`

- Test hybrid code generation (industry + keywords)
- Test confidence filtering
- Test code merging and deduplication
- Test ranking logic

### Integration Tests

**File**: `services/classification-service/test/integration/hybrid_code_generation_test.go`

- Test end-to-end hybrid code generation
- Test fallback when code_keywords table is empty
- Test performance with large keyword sets
- Test multi-industry code generation

## Configuration

**File**: `services/classification-service/internal/config/config.go`

**Add configuration options**:

```go
type ClassificationConfig struct {
    // ... existing fields ...
    
    // Hybrid code generation settings
    EnableKeywordCodeMatching    bool    `env:"ENABLE_KEYWORD_CODE_MATCHING" envDefault:"true"`
    KeywordCodeMinRelevance      float64 `env:"KEYWORD_CODE_MIN_RELEVANCE" envDefault:"0.5"`
    CodeConfidenceThreshold      float64 `env:"CODE_CONFIDENCE_THRESHOLD" envDefault:"0.6"`
    MaxCodesPerType              int     `env:"MAX_CODES_PER_TYPE" envDefault:"10"`
    EnableMultiIndustryCodes     bool    `env:"ENABLE_MULTI_INDUSTRY_CODES" envDefault:"false"`
    MultiIndustryCodeWeight       float64 `env:"MULTI_INDUSTRY_CODE_WEIGHT" envDefault:"0.7"` // Weight for secondary industries
}
```

## Database Considerations

### Verify code_keywords Table Population

- Check if `code_keywords` table has data
- If empty, document data population requirements
- Add migration script if needed to populate initial data

### Performance Optimization

- Ensure indexes exist: `idx_code_keywords_keyword`, `idx_code_keywords_code`
- Consider adding composite index: `(keyword, code_type, relevance_score)`
- Add query result caching for frequently matched keywords

## Rollout Strategy

### Step 1: Implement Phase 1 (code_keywords query)

- Add repository method
- Add unit tests
- Verify with test data

### Step 2: Implement Phase 2 (hybrid generation)

- Modify code generator
- Add merge logic
- Test with existing classification flow

### Step 3: Add Phase 3 (confidence filtering)

- Add filtering and ranking
- Tune confidence thresholds
- Monitor accuracy improvements

### Step 4: Optional Phase 4 (multi-industry)

- Add multi-industry support
- Integrate with ensemble classifier
- Test with multi-industry businesses

## Success Metrics

- **Accuracy**: Increase code match accuracy by 50-60% over industry-only approach
- **Coverage**: Generate codes for businesses where industry detection fails
- **Performance**: Maintain <100ms overhead for hybrid approach
- **Resilience**: Fallback gracefully when code_keywords table is empty

## Risk Mitigation

1. **Empty code_keywords table**: Fallback to industry-only (existing behavior)
2. **Performance degradation**: Add caching, limit keyword set size
3. **Too many codes returned**: Apply confidence threshold and top-N limiting
4. **Breaking changes**: Maintain backward compatibility, add feature flags

### Phase 5: API Response Enhancement for Frontend

**File**: `services/merchant-service/internal/jobs/classification_job.go`

**Enhance `IndustryCode` struct**:

```go
type IndustryCode struct {
    Code            string   `json:"code"`
    Description     string   `json:"description"`
    Confidence      float64  `json:"confidence"`
    // New fields for advanced features:
    Source          []string `json:"source,omitempty"`          // ["industry", "keyword", "both"]
    MatchType       string   `json:"matchType,omitempty"`       // "exact", "partial", "synonym"
    RelevanceScore  float64  `json:"relevanceScore,omitempty"`  // From code_keywords table
    Industries      []string `json:"industries,omitempty"`      // Industries that contributed this code
    IsPrimary       bool     `json:"isPrimary,omitempty"`       // From classification_codes.is_primary
}
```

**Update `extractClassificationFromResponse()` method**:

- Extract new metadata fields from classification service response
- Map source information to `IndustryCode.Source`
- Include match types and relevance scores
- Include contributing industries

**File**: `services/merchant-service/internal/handlers/merchant.go`

**Update `HandleMerchantSpecificAnalytics()` method**:

- Ensure new `IndustryCode` fields are included in API response
- Add code generation metadata to response:
  ```go
  CodeGenerationMetadata: map[string]interface{}{
      "method": "hybrid", // "industry_only", "keyword_only", "hybrid"
      "sources": []string{"industry", "keyword"},
      "industriesAnalyzed": []string{...},
      "keywordMatches": int,
      "industryMatches": int,
  }
  ```


### Phase 6: Frontend Type Updates and UI Enhancement

**File**: `frontend/types/merchant.ts`

**Update `IndustryCode` interface**:

```typescript
export interface IndustryCode {
  code: string;
  description: string;
  confidence: number;
  // New optional fields:
  source?: ('industry' | 'keyword' | 'both')[];
  matchType?: 'exact' | 'partial' | 'synonym';
  relevanceScore?: number;
  industries?: string[];
  isPrimary?: boolean;
}
```

**Update `ClassificationData` interface**:

```typescript
export interface ClassificationData {
  // ... existing fields ...
  metadata?: {
    // ... existing metadata ...
    codeGeneration?: {
      method: 'industry_only' | 'keyword_only' | 'hybrid';
      sources: string[];
      industriesAnalyzed: string[];
      keywordMatches: number;
      industryMatches: number;
      totalCodesGenerated: number;
    };
  };
}
```

**File**: `frontend/components/merchant/ClassificationMetadata.tsx`

**Enhance component to display**:

- Code generation method badge (Hybrid/Industry/Keyword)
- Source indicators for each code (industry/keyword/both icons)
- Match type badges (exact/partial/synonym)
- Confidence score visualization
- Contributing industries list
- Primary code indicators

**New component**: `frontend/components/merchant/CodeSourceBadge.tsx`

- Display source information for each code
- Show match type with color coding
- Display confidence and relevance scores

## Files to Modify

1. `internal/classification/repository/supabase_repository.go` - Add `GetClassificationCodesByKeywords()`
2. `internal/classification/repository/interface.go` - Add interface method
3. `internal/classification/classifier.go` - Modify `generateCodesInParallel()` and add helper methods
4. `services/classification-service/internal/config/config.go` - Add configuration options
5. `internal/classification/multi_method_classifier.go` - Integrate multi-industry support (required)
6. `services/merchant-service/internal/jobs/classification_job.go` - Enhance `IndustryCode` struct and extraction logic
7. `services/merchant-service/internal/handlers/merchant.go` - Update API response with new metadata
8. `frontend/types/merchant.ts` - Update TypeScript interfaces
9. `frontend/components/merchant/ClassificationMetadata.tsx` - Enhance UI to display new features
10. `frontend/lib/api-validation.ts` - Update Zod schemas for new fields

## Files to Create

1. `internal/classification/repository/supabase_repository_keyword_codes_test.go` - Unit tests for keyword code matching
2. `internal/classification/classifier_hybrid_test.go` - Unit tests for hybrid generation
3. `services/classification-service/test/integration/hybrid_code_generation_test.go` - Integration tests
4. `frontend/components/merchant/CodeSourceBadge.tsx` - New component for code source display
5. `frontend/__tests__/components/merchant/CodeSourceBadge.test.tsx` - Unit tests for new component