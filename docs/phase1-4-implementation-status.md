# Phase 1.4 Implementation Status: Fix Co-Occurrence Strategy

## Overview
Enhanced co-occurrence strategy with relationship analysis, database-driven pattern matching, and intelligent keyword pair generation (target: 70%+ accuracy).

## Completed

### 1. Database Migration for Keyword Patterns
- ✅ Created `keyword_patterns` table with:
  - `industry_id` (FK to industries)
  - `keyword_pair` (normalized format: "keyword1|keyword2")
  - `keyword1` and `keyword2` (individual keywords)
  - `co_occurrence_score` (0.0-1.0)
  - `pattern_type` (keyword_keyword, entity_keyword, entity_entity)
  - `frequency` (usage tracking)
- ✅ Created indexes for fast queries:
  - Index on `industry_id`
  - Index on `keyword_pair`
  - Indexes on `keyword1` and `keyword2`
  - Index on `co_occurrence_score` (DESC)
  - Composite index on `(industry_id, keyword_pair)`
- ✅ Created `normalize_keyword_pair()` function:
  - Ensures consistent pair format (alphabetically sorted)
  - Immutable function for reliable normalization
- ✅ Created `find_industries_by_patterns()` function:
  - Finds industries matching keyword patterns
  - Requires at least 2 pattern matches
  - Returns sorted results by pattern matches and avg score
- ✅ Populated initial data from `keyword_weights`:
  - Generates pairs from keywords within same industry
  - Calculates co-occurrence score from average weights
  - Handles conflicts with updates
- ✅ Created `keyword_patterns_view` for easy querying
- **File**: `supabase-migrations/032_add_keyword_patterns_cooccurrence.sql`

### 2. Enhanced Co-Occurrence Classification
- ✅ Added `analyzeCoOccurrencePatterns()` method:
  - Generates keyword-keyword pairs
  - Generates entity-keyword pairs
  - Generates entity-entity pairs
  - Normalizes pairs for consistency
  - Deduplicates patterns
- ✅ Enhanced `classifyByCoOccurrence()` method:
  - Analyzes co-occurrence patterns from keywords and entities
  - Queries database for industry pattern matches
  - Calculates confidence based on pattern matches and scores
  - Applies confidence boosts for multiple pattern matches
  - Includes detailed metadata (pattern counts, scores)
- ✅ Added `classifyByCoOccurrenceFallback()` method:
  - Graceful fallback when pattern matching fails
  - Uses basic keyword classification
  - Maintains backward compatibility
- **File**: `internal/classification/multi_strategy_classifier.go`

### 3. Repository Implementation
- ✅ Added `PatternMatchResult` type:
  - Industry ID and name
  - Pattern match count
  - Average score
  - Matched patterns list
- ✅ Added `KeywordPattern` type:
  - Pattern details (pair, keywords, score, type, frequency)
- ✅ Added `FindIndustriesByPatterns()` method:
  - Calls database function via PostgREST RPC
  - Returns industries matching patterns
  - Handles errors gracefully
- ✅ Added `GetPatternMatches()` method:
  - Retrieves patterns for specific industry
  - Supports filtering by pattern list
- ✅ Added methods to `KeywordRepository` interface
- ✅ Implemented methods in `SupabaseKeywordRepository`
- **Files**: 
  - `internal/classification/repository/interface.go`
  - `internal/classification/repository/supabase_repository.go`

## Implementation Details

### Pattern Generation
```go
// Generates three types of patterns:
// 1. Keyword-keyword pairs: "wine|shop"
// 2. Entity-keyword pairs: "restaurant|food"
// 3. Entity-entity pairs: "retail|food"
```

### Confidence Calculation
```go
// Base confidence from pattern matches and scores
patternMatchRatio = patternMatches / totalPatterns
baseConfidence = (patternMatchRatio * 0.6) + (avgScore * 0.4)

// Boosts for multiple matches:
// - 3+ matches: 20% boost
// - 2 matches: 10% boost
```

### Database Function
```sql
-- find_industries_by_patterns()
-- Requires at least 2 pattern matches
-- Returns top 10 industries sorted by:
--   1. Pattern match count (DESC)
--   2. Average score (DESC)
```

## Expected Impact
- **Accuracy**: 70%+ for co-occurrence strategy (target from plan)
- **Coverage**: Better handling of complex keyword relationships
- **Performance**: Fast queries via indexes and database functions
- **Reliability**: Graceful fallback ensures classification always works

## Files Modified
- `internal/classification/multi_strategy_classifier.go` - Enhanced with pattern analysis
- `internal/classification/repository/interface.go` - Added pattern matching methods
- `internal/classification/repository/supabase_repository.go` - Implemented pattern matching
- `supabase-migrations/032_add_keyword_patterns_cooccurrence.sql` - New migration

## Pattern Types Supported

1. **Keyword-Keyword Pairs**: Relationships between extracted keywords
   - Example: "wine|shop", "retail|store"

2. **Entity-Keyword Pairs**: Relationships between entities and keywords
   - Example: "restaurant|food", "technology|software"

3. **Entity-Entity Pairs**: Relationships between multiple entities
   - Example: "retail|food", "healthcare|technology"

## Next Steps
1. Apply database migration to create `keyword_patterns` table
2. Test pattern matching with sample data
3. Measure accuracy improvement
4. Continue with Phase 1.5 (Fix Combination Logic)

