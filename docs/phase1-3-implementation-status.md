# Phase 1.3 Implementation Status: Fix Topic Modeling

## Overview
Enhanced topic modeling with database-driven industry-topic mapping, TF-IDF calibration, and historical accuracy-based scoring (target: 75%+ accuracy).

## Completed

### 1. Database Migration for Industry-Topics Mapping
- ✅ Created `industry_topics` table with:
  - `industry_id` (FK to industries)
  - `topic` (keyword/phrase)
  - `relevance_score` (0.0-1.0)
  - `accuracy_score` (historical accuracy, 0.0-1.0)
  - `topic_type` (keyword/phrase/concept)
  - `usage_count` (tracking)
- ✅ Created indexes for fast queries:
  - Index on `industry_id`
  - Index on `topic`
  - Index on `relevance_score` (DESC)
  - Index on `accuracy_score` (DESC)
  - Full-text search index on `topic`
- ✅ Created `update_topic_accuracy()` function for feedback-based learning
- ✅ Populated initial data from `industry_keywords` table
- ✅ Created `industry_topics_view` for easy querying
- **File**: `supabase-migrations/031_add_industry_topics_mapping.sql`

### 2. Enhanced TopicModeler
- ✅ Added `TopicRepository` interface for database queries
- ✅ Added optional repository support (backward compatible)
- ✅ Added `mapTopicsToIndustries()` method:
  - Queries database for industry-topic relationships
  - Uses caching with TTL (1 hour default)
  - Merges database results with in-memory results
- ✅ Added `calibrateScores()` method:
  - Adjusts scores based on historical accuracy
  - Uses exponential moving average for accuracy tracking
  - Calibration factor: 0.5 + (accuracy * 0.5)
- ✅ Added `mergeTopicScores()` method:
  - Combines in-memory (40% weight) and database (60% weight) scores
  - Handles duplicate industry IDs
- ✅ Added caching for topic-industry mappings:
  - Cache key: comma-separated keywords
  - TTL: 1 hour (configurable)
  - Thread-safe with `cacheMu` mutex
- ✅ Enhanced `IdentifyTopicsWithDetails()`:
  - Now supports context for database queries
  - Falls back to in-memory if database unavailable
  - Maintains backward compatibility
- **File**: `internal/classification/nlp/topic_modeler.go`

### 3. Repository Implementation
- ✅ Added `GetIndustryTopicsByKeywords()` method:
  - Queries `industry_topics` table
  - Returns map of `industry_id -> relevance_score`
  - Weighted by accuracy score
  - Handles multiple keywords
- ✅ Added `GetTopicAccuracy()` method:
  - Retrieves accuracy score for specific topic-industry pair
  - Returns default (0.75) if not found
- ✅ Added methods to `KeywordRepository` interface
- ✅ Implemented methods in `SupabaseKeywordRepository`
- **Files**: 
  - `internal/classification/repository/interface.go`
  - `internal/classification/repository/supabase_repository.go`

### 4. Integration
- ✅ Updated `MultiStrategyClassifier` to use enhanced TopicModeler:
  - Automatically detects repository support
  - Creates adapter for TopicRepository interface
  - Uses context-aware topic identification
- ✅ Updated `classifyByTopics()` to use context
- **File**: `internal/classification/multi_strategy_classifier.go`

## Implementation Details

### Database Schema
```sql
CREATE TABLE industry_topics (
    id SERIAL PRIMARY KEY,
    industry_id INTEGER REFERENCES industries(id),
    topic VARCHAR(100) NOT NULL,
    relevance_score DECIMAL(3,2) DEFAULT 0.80,
    accuracy_score DECIMAL(3,2) DEFAULT 0.75,
    topic_type VARCHAR(50) DEFAULT 'keyword',
    usage_count INTEGER DEFAULT 0,
    UNIQUE(industry_id, topic)
);
```

### Score Calibration Formula
```go
calibrationFactor = 0.5 + (accuracyScore * 0.5)  // Range: 0.5 to 1.0
calibratedScore = baseScore * calibrationFactor
```

### Score Merging
- In-memory TF-IDF scores: 40% weight
- Database relevance scores: 60% weight
- Combined: `finalScore = (inMemory * 0.4) + (database * 0.6)`

### Caching Strategy
- Cache key: comma-separated keywords
- TTL: 1 hour (configurable)
- Cache invalidation: Time-based (TTL expiry)
- Thread-safe: Uses `sync.RWMutex`

## Expected Impact
- **Accuracy**: 75%+ for topic strategy (from current baseline)
- **Performance**: Fast queries via indexes and caching
- **Adaptability**: Self-improving via accuracy tracking

## Files Modified
- `internal/classification/nlp/topic_modeler.go` - Enhanced with database support
- `internal/classification/repository/interface.go` - Added topic mapping methods
- `internal/classification/repository/supabase_repository.go` - Implemented topic mapping
- `internal/classification/multi_strategy_classifier.go` - Integrated enhanced TopicModeler
- `supabase-migrations/031_add_industry_topics_mapping.sql` - New migration

## Next Steps
1. Apply database migration to create `industry_topics` table
2. Test topic mapping with sample data
3. Measure accuracy improvement
4. Continue with Phase 1.4 (Fix Co-Occurrence Strategy)

