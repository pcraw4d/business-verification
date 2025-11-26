<!-- 5eb6e4b7-b4aa-4192-8012-c430f713b689 d2e64be5-aa47-4b12-b5da-c39ea1a48bc2 -->
# Classification Accuracy Improvement Plan - Target: 95%

## Overview

This plan addresses critical database issues, implements word segmentation and advanced NLP capabilities, and dramatically improves classification accuracy to 95% by ensuring comprehensive code coverage and keyword matching.

## Current Issues

1. **Database Issues**:

   - Missing SIC codes for Food & Beverage (industry_id=10)
   - Incorrect NAICS codes (hotel codes instead of food/beverage codes)
   - Incomplete code coverage across all industries
   - Insufficient keyword matching in `code_keywords` table

2. **Keyword Extraction Limitations**:

   - Compound domain names not properly segmented (e.g., "thegreenegrape")
   - Limited entity recognition from website content
   - No topic modeling for industry classification

3. **Classification Accuracy**:

   - Current accuracy below target (likely 60-70%)
   - Need to reach 95% accuracy

## Phase 1: Database Fixes and Code Population

### 1.1 Fix Incorrect Codes for Food & Beverage (Industry ID 10)

**File**: `supabase-migrations/XXX_fix_food_beverage_codes.sql`

**Tasks**:

- Disable incorrect hotel NAICS codes (721110, 721120, 721191)
- Add correct Food & Beverage NAICS codes:
  - 722410 - Drinking Places (Alcoholic Beverages)
  - 445310 - Beer, Wine, and Liquor Stores
  - 722511 - Full-Service Restaurants
  - 722513 - Limited-Service Restaurants
  - 445110 - Supermarkets and Grocery Stores
  - 311111 - Dog and Cat Food Manufacturing
- Add missing SIC codes:
  - 5812 - Eating Places
  - 5921 - Package Stores (Beer, Wine, Liquor)
  - 5499 - Miscellaneous Food Stores
  - 5813 - Drinking Places

**Implementation**:

```sql
-- Disable incorrect codes
UPDATE classification_codes 
SET is_active = false 
WHERE industry_id = 10 
  AND code_type = 'NAICS' 
  AND code IN ('721110', '721120', '721191');

-- Add correct NAICS codes
INSERT INTO classification_codes (industry_id, code_type, code, description, is_active, is_primary, confidence)
VALUES 
  (10, 'NAICS', '722410', 'Drinking Places (Alcoholic Beverages)', true, true, 0.95),
  (10, 'NAICS', '445310', 'Beer, Wine, and Liquor Stores', true, true, 0.95),
  (10, 'NAICS', '722511', 'Full-Service Restaurants', true, false, 0.90),
  (10, 'NAICS', '722513', 'Limited-Service Restaurants', true, false, 0.90),
  (10, 'NAICS', '445110', 'Supermarkets and Grocery Stores', true, false, 0.85),
  (10, 'NAICS', '311111', 'Dog and Cat Food Manufacturing', true, false, 0.75)
ON CONFLICT (code_type, code) DO UPDATE 
SET industry_id = 10, description = EXCLUDED.description, is_active = true;

-- Add SIC codes
INSERT INTO classification_codes (industry_id, code_type, code, description, is_active, is_primary, confidence)
VALUES 
  (10, 'SIC', '5812', 'Eating Places', true, true, 0.95),
  (10, 'SIC', '5921', 'Package Stores (Beer, Wine, Liquor)', true, true, 0.95),
  (10, 'SIC', '5499', 'Miscellaneous Food Stores', true, false, 0.85),
  (10, 'SIC', '5813', 'Drinking Places', true, false, 0.90)
ON CONFLICT (code_type, code) DO UPDATE 
SET industry_id = 10, description = EXCLUDED.description, is_active = true;
```

### 1.2 Comprehensive Code Population Script

**File**: `scripts/populate_all_classification_codes.sql`

**Tasks**:

- Create script to populate ALL MCC, SIC, and NAICS codes for all industries
- Source data from official code lists:
  - MCC: Complete list from payment processors
  - SIC: Complete SIC code list
  - NAICS: Complete NAICS 2022 code list
- Map codes to industries based on:
  - Official code descriptions
  - Industry alignment
  - Crosswalk tables (NAICS to SIC, etc.)

**Implementation Strategy**:

1. Create mapping tables for code-to-industry relationships
2. Use official crosswalk data where available
3. For each industry, identify all relevant codes (top 20-30 per code type)
4. Bulk insert with conflict handling

**Estimated Codes per Industry**:

- MCC: 10-20 codes per industry
- SIC: 15-25 codes per industry  
- NAICS: 20-30 codes per industry

### 1.3 Comprehensive Keyword Population for Codes

**File**: `scripts/populate_code_keywords.sql`

**Tasks**:

- For each classification code, extract keywords from:
  - Official code description
  - Industry-specific terminology
  - Common business phrases
  - Synonyms and related terms
- Populate `code_keywords` table with:
  - Primary keywords (relevance_score = 1.0)
  - Secondary keywords (relevance_score = 0.8)
  - Related keywords (relevance_score = 0.6)
  - Synonym keywords (relevance_score = 0.7, match_type = 'synonym')

**Target**: 10-20 keywords per code minimum

**Example**:

```sql
-- For NAICS 722410 (Drinking Places)
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT id, keyword, score, match_type
FROM (VALUES
  ('drinking places', 1.0, 'exact'),
  ('bar', 1.0, 'exact'),
  ('pub', 0.9, 'exact'),
  ('tavern', 0.9, 'exact'),
  ('alcoholic beverages', 0.95, 'exact'),
  ('wine bar', 0.95, 'exact'),
  ('cocktail', 0.85, 'partial'),
  ('liquor', 0.9, 'synonym'),
  ('spirits', 0.85, 'synonym')
) AS keywords(keyword, score, match_type)
CROSS JOIN classification_codes 
WHERE code = '722410' AND code_type = 'NAICS';
```

## Phase 2: Word Segmentation Library Implementation

### 2.1 Research and Select Word Segmentation Library

**Options**:

- `github.com/kljensen/snowball` - Snowball stemmer (basic)
- Custom dictionary-based approach
- Hybrid: Dictionary + heuristics

**Decision**: Implement hybrid approach with:

- Dictionary-based segmentation for common business terms
- Heuristic fallback for unknown words
- Caching for performance

### 2.2 Implement Word Segmentation Module

**File**: `internal/classification/word_segmentation/segmenter.go`

**Implementation**:

```go
package word_segmentation

type Segmenter struct {
    dictionary map[string]bool
    cache      map[string][]string
    mutex      sync.RWMutex
}

func NewSegmenter() *Segmenter {
    return &Segmenter{
        dictionary: loadBusinessDictionary(),
        cache:      make(map[string][]string),
    }
}

func (s *Segmenter) Segment(domain string) []string {
    // Check cache
    s.mutex.RLock()
    if cached, exists := s.cache[domain]; exists {
        s.mutex.RUnlock()
        return cached
    }
    s.mutex.RUnlock()

    // Perform segmentation
    segments := s.segmentWithDictionary(domain)
    if len(segments) == 0 {
        segments = s.segmentWithHeuristics(domain)
    }

    // Cache result
    s.mutex.Lock()
    s.cache[domain] = segments
    s.mutex.Unlock()

    return segments
}
```

**Dictionary Sources**:

- Common business terms (wine, shop, store, retail, etc.)
- Industry-specific terms
- Common English words (for compound domains)

### 2.3 Integrate with Domain Name Extraction

**File**: `internal/classification/repository/supabase_repository.go`

**Changes**:

- Update `splitDomainName()` to use word segmentation
- Extract 2-3x more keywords from compound domains
- Example: "thegreenegrape" → ["the", "green", "grape", "green grape", "wine", "shop"]

**Location**: Line ~1950 in `extractKeywordsFromURLEnhanced()`

## Phase 3: Advanced NLP Implementation (Hybrid Approach)

### 3.1 Named Entity Recognition (NER)

**File**: `internal/classification/nlp/entity_recognizer.go`

**Implementation Strategy**:

1. **Pattern-based NER** (fast, high precision):

   - Business entity patterns (e.g., "wine shop", "retail store")
   - Industry indicators (e.g., "restaurant", "technology company")
   - Location entities (for regional classification)

2. **Library-based NER** (higher recall):

   - Use Go NLP library (e.g., `github.com/jdkato/prose` or `github.com/advancedlogic/go-nlp`)
   - Extract named entities from website content
   - Identify business types, services, products

**Implementation**:

```go
package nlp

type EntityRecognizer struct {
    patterns    []EntityPattern
    nlpLibrary  NLPLibrary // Optional, for advanced extraction
}

type Entity struct {
    Text       string
    Type       EntityType // BUSINESS_TYPE, SERVICE, PRODUCT, INDUSTRY
    Confidence float64
    Source     string // "pattern" or "nlp"
}

func (er *EntityRecognizer) ExtractEntities(text string) []Entity {
    entities := []Entity{}
    
    // Pattern-based extraction (fast)
    entities = append(entities, er.extractWithPatterns(text)...)
    
    // Library-based extraction (if available)
    if er.nlpLibrary != nil {
        entities = append(entities, er.extractWithLibrary(text)...)
    }
    
    return deduplicateEntities(entities)
}
```

### 3.2 Topic Modeling

**File**: `internal/classification/nlp/topic_modeler.go`

**Implementation**:

- Use TF-IDF for keyword extraction
- Identify dominant topics in website content
- Map topics to industries
- Use topic distribution for industry confidence scoring

**Approach**:

1. Extract keywords with TF-IDF scoring
2. Group keywords into topics (e.g., "wine", "retail", "food" → Food & Beverage + Retail)
3. Calculate topic-industry alignment scores
4. Use for multi-industry classification

**Implementation**:

```go
type TopicModeler struct {
    industryTopics map[int][]string // industry_id -> topic keywords
}

func (tm *TopicModeler) IdentifyTopics(keywords []string) map[int]float64 {
    // Calculate topic scores for each industry
    topicScores := make(map[int]float64)
    
    for industryID, topicKeywords := range tm.industryTopics {
        score := calculateTopicAlignment(keywords, topicKeywords)
        if score > 0.3 {
            topicScores[industryID] = score
        }
    }
    
    return topicScores
}
```

### 3.3 Integrate NLP with Keyword Extraction

**File**: `internal/classification/smart_website_crawler.go`

**Changes**:

- Add entity recognition to `extractPageKeywords()`
- Use topic modeling for industry confidence
- Enhance keyword scoring with NLP insights

**Location**: Update `extractPageKeywords()` method (line ~782)

## Phase 4: Enhanced Keyword Matching

### 4.1 Expand Keyword Matching Strategies

**File**: `internal/classification/repository/supabase_repository.go`

**Current**: Basic exact/word-boundary matching

**Enhancement**: Add multiple matching strategies:

1. Exact match (current)
2. Word-boundary match (current)
3. Synonym matching (new)
4. Stemming-based matching (new)
5. Fuzzy matching for typos (new, low weight)

**Location**: `GetClassificationCodesByKeywords()` method (line ~1057)

**Implementation**:

```go
// Add synonym matching
synonyms := getSynonyms(searchKeyword)
for _, synonym := range synonyms {
    if rowKeywordLower == synonym {
        matches = true
        relevanceScore *= 0.9 // Slight penalty for synonym match
        break
    }
}

// Add stemming-based matching
if stemmedMatch(rowKeywordLower, searchKeyword) {
    matches = true
    relevanceScore *= 0.85
}
```

### 4.2 Increase Keyword Coverage

**Target**: 20-30 keywords per code (currently likely 5-10)

**Strategy**:

- Extract keywords from code descriptions
- Add industry-specific terminology
- Include common misspellings and variations
- Add multi-word phrases

**File**: `scripts/populate_code_keywords_comprehensive.sql`

## Phase 5: Classification Accuracy Improvements

### 5.1 Multi-Strategy Classification

**File**: `internal/classification/classifier.go`

**Enhancement**: Combine multiple classification signals:

1. Keyword-based (current)
2. Entity-based (new from NLP)
3. Topic-based (new from topic modeling)
4. Industry co-occurrence (current, enhance)

**Weighted Scoring**:

```go
finalScore = (
    keywordScore * 0.40 +
    entityScore * 0.25 +
    topicScore * 0.20 +
    coOccurrenceScore * 0.15
)
```

### 5.2 Confidence Calibration

**File**: `internal/classification/repository/supabase_repository.go`

**Enhancement**: Calibrate confidence scores to match actual accuracy:

- Track classification accuracy by confidence level
- Adjust confidence thresholds based on historical performance
- Ensure 95% accuracy target is met

**Implementation**:

- Add accuracy tracking table
- Periodically recalibrate confidence thresholds
- Use machine learning for confidence prediction (future)

### 5.3 Validation and Testing

**File**: `internal/classification/classifier_accuracy_test.go`

**Tests**:

- Test with known business websites
- Verify 95% accuracy target
- Test edge cases (compound domains, multi-industry businesses)
- Performance benchmarks

## Phase 6: Database Schema Enhancements

### 6.1 Add Missing Columns/Indexes

**File**: `supabase-migrations/XXX_enhance_classification_schema.sql`

**Additions**:

- `is_active` column to `classification_codes` (if missing)
- Full-text search indexes on descriptions
- Composite indexes for common queries
- Trigram indexes for fuzzy matching

### 6.2 Add Code Metadata Table

**File**: `supabase-migrations/XXX_add_code_metadata.sql`

**Purpose**: Store additional code information:

- Official code descriptions
- Industry mappings
- Crosswalk data (NAICS ↔ SIC ↔ MCC)
- Code hierarchies

## Implementation Timeline

### Week 1: Database Fixes

- Day 1-2: Fix Food & Beverage codes
- Day 3-4: Create comprehensive code population script
- Day 5: Test and validate code population

### Week 2: Word Segmentation

- Day 1-2: Research and implement word segmentation
- Day 3: Integrate with domain extraction
- Day 4-5: Testing and optimization

### Week 3: Advanced NLP

- Day 1-2: Implement NER (pattern-based)
- Day 3: Integrate NLP library (if needed)
- Day 4: Implement topic modeling
- Day 5: Integration testing

### Week 4: Keyword Matching & Accuracy

- Day 1-2: Enhance keyword matching strategies
- Day 3: Populate comprehensive keywords
- Day 4: Multi-strategy classification
- Day 5: Accuracy testing and calibration

## Success Criteria

1. **Database**:

   - All industries have codes (MCC, SIC, NAICS)
   - Minimum 10 codes per code type per industry
   - Minimum 15 keywords per code
   - No incorrect code mappings

2. **Word Segmentation**:

   - Extract 2-3x more keywords from compound domains
   - 90%+ accuracy on common business domain names

3. **NLP**:

   - Extract 5-10 entities per website
   - Identify 2-3 topics per website
   - Entity extraction accuracy > 85%

4. **Classification Accuracy**:

   - Overall accuracy: 95%+
   - Industry detection accuracy: 95%+
   - Code generation accuracy: 90%+ (top 3 codes)

5. **Performance**:

   - Classification completes in < 5 seconds
   - Keyword matching < 500ms
   - NLP processing < 2 seconds

## Dependencies

- Go NLP libraries (optional, for advanced features)
- Official code lists (MCC, SIC, NAICS)
- Business dictionary for word segmentation
- Testing dataset with known classifications

## Risk Mitigation

1. **Large Code Database**: Use batch inserts and indexing
2. **NLP Performance**: Start with pattern-based, add library later
3. **Accuracy Target**: Implement gradual improvements with validation
4. **Breaking Changes**: Feature flags for new functionality

### To-dos

- [ ] Fix incorrect NAICS codes and add missing SIC codes for Food & Beverage (industry_id=10)
- [ ] Create comprehensive script to populate ALL MCC, SIC, and NAICS codes for all industries
- [ ] Populate code_keywords table with 15-20 keywords per code (10-20x current coverage)
- [ ] Implement word segmentation library for compound domain names (hybrid: dictionary + heuristics)
- [ ] Integrate word segmentation with domain name extraction in extractKeywordsFromURLEnhanced()
- [ ] Implement Named Entity Recognition (hybrid: pattern-based + library-based)
- [ ] Implement topic modeling using TF-IDF for industry classification
- [ ] Integrate NER and topic modeling with keyword extraction pipeline
- [ ] Add synonym matching, stemming, and fuzzy matching to GetClassificationCodesByKeywords()
- [ ] Implement multi-strategy classification combining keywords, entities, topics, and co-occurrence
- [ ] Implement confidence calibration to ensure 95% accuracy target
- [ ] Create comprehensive accuracy tests and validate 95% target is met