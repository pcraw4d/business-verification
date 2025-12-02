# Phase 1.1 Implementation Status: Fix Keyword Strategy

## Overview
Enhancing the keyword strategy to leverage trigram indexes and full-text search for improved accuracy (target: 85%+).

## Completed

### 1. Enhanced `classifyByKeywords` Method
- ✅ Added context timeout (2 seconds) as per plan
- ✅ Improved error handling with timeout detection
- ✅ Added metadata tracking for timeout configuration
- **File**: `internal/classification/multi_strategy_classifier.go`

### 2. Database Function Migration
- ✅ Created database function `classify_business_by_keywords_trigram`
- ✅ Function uses `similarity()` for fuzzy matching with trigram indexes
- ✅ Added trigram index on `keyword_weights.keyword`
- ✅ Function supports similarity threshold parameter (default 0.3)
- **File**: `supabase-migrations/030_add_trigram_keyword_classification.sql`

## Pending

### 1. RPC Call Implementation
- ⏳ Need to add method in `SupabaseKeywordRepository` to call the database function via PostgREST RPC
- ⏳ PostgREST RPC calls are made via HTTP POST to `/rest/v1/rpc/function_name`
- ⏳ Need to implement: `ClassifyBusinessByKeywordsTrigram(ctx, keywords, businessName)`

### 2. Integration
- ⏳ Update `classifyByKeywords` to optionally use trigram function for fuzzy matching
- ⏳ Fallback to existing in-memory index for exact matches
- ⏳ Hybrid approach: exact matches from index, fuzzy matches from database

### 3. Database Migration Execution
- ⏳ Migration file needs to be applied to database
- ⏳ Verify trigram index is created
- ⏳ Test database function with sample queries

## Implementation Notes

### Database Function Details
The function `classify_business_by_keywords_trigram`:
- Takes array of keywords and optional business name
- Uses `similarity()` function from pg_trgm extension
- Returns top 10 industries with scores and matched keywords
- Leverages trigram index for fast performance

### RPC Call Pattern
PostgREST RPC calls follow this pattern:
```http
POST /rest/v1/rpc/classify_business_by_keywords_trigram
Content-Type: application/json
{
  "p_keywords": ["keyword1", "keyword2"],
  "p_business_name": "Business Name",
  "p_similarity_threshold": 0.3
}
```

### Next Steps
1. Add RPC call method to repository
2. Integrate with existing classification flow
3. Test with sample data
4. Measure accuracy improvement

## Expected Impact
- **Accuracy**: 85%+ for keyword strategy (from current baseline)
- **Performance**: Fast fuzzy matching via trigram indexes
- **Coverage**: Better handling of typos and variations

