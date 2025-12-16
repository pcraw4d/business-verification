# Phase 2 Testing Guide

## Quick Start

### Option 1: API Test Script (Recommended)

```bash
# Set API URL if different from default
export API_BASE_URL="http://localhost:8080"

# Run comprehensive tests
./test/phase2_api_test.sh
```

### Option 2: Manual API Testing

```bash
# Test Top 3 Codes
curl -X POST http://localhost:8080/classify \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Joe'\''s Pizza Restaurant",
    "description": "Family pizza restaurant serving authentic Italian cuisine"
  }' | jq '.classification | {mcc_codes: .mcc_codes | length, sic_codes: .sic_codes | length, naics_codes: .naics_codes | length}'

# Test Confidence Calibration
curl -X POST http://localhost:8080/classify \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Starbucks Coffee",
    "description": "Coffee shop and cafe"
  }' | jq '{confidence: .confidence_score, industry: .classification.industry}'

# Test Fast Path
time curl -X POST http://localhost:8080/classify \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Pizza Hut",
    "description": "Pizza restaurant"
  }'

# Test Explanations
curl -X POST http://localhost:8080/classify \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Mario'\''s Italian Restaurant",
    "description": "Authentic Italian restaurant"
  }' | jq '.classification.explanation'
```

## Test Cases

### 1. Top 3 Codes Per Type

**Expected:**

- Exactly 3 MCC codes
- Exactly 3 SIC codes
- Exactly 3 NAICS codes
- Each code has `source` field (array)

**Test Businesses:**

- "Joe's Pizza Restaurant" - Should return restaurant codes
- "Tech Startup Inc" - Should return technology codes
- "Fashion Boutique" - Should return retail codes

### 2. Confidence Calibration

**Expected:**

- Confidence scores in 70-95% range (not all ~50%)
- High-quality cases: 85-95%
- Ambiguous cases: 65-80%

**Test Businesses:**

- "Starbucks Coffee" - High confidence (85-95%)
- "ABC Services" - Medium confidence (60-80%)

### 3. Fast Path Performance

**Expected:**

- Fast path handles 60-70% of obvious cases
- Fast path latency <100ms (p95)
- Average latency <200ms

**Test Businesses (Obvious Cases):**

- "Pizza Hut" - Restaurant
- "Starbucks Coffee" - Coffee shop
- "Hilton Hotel" - Hotel
- "Chase Bank" - Bank

### 4. Structured Explanations

**Expected:**

- Every classification has `primary_reason`
- 3-5 `supporting_factors` listed
- `key_terms_found` array populated
- `method_used` field present
- `processing_path` field present

**Test Businesses:**

- "Mario's Italian Restaurant" - Should have detailed explanation
- "Cloud Services Inc" - Should explain tech classification

### 5. Generic Fallback Fix

**Expected:**

- "General Business" < 10% of results
- Specific industries preferred over generic
- Generic only used when truly appropriate

**Test Businesses (Ambiguous):**

- "ABC Corporation" - Should prefer specific industry
- "XYZ Services" - Should prefer specific industry
- "Global Enterprises" - May use generic if truly ambiguous

### 6. Performance Metrics

**Expected:**

- P50 latency < 200ms
- P90 latency < 400ms
- P95 latency < 500ms

## Success Criteria

- ✅ Top 3 codes returned per type with Source field
- ✅ Confidence scores in 70-95% range
- ✅ Fast path hit rate >= 60%
- ✅ Fast path latency < 100ms (p95)
- ✅ Structured explanations for all classifications
- ✅ "General Business" < 10% of results
- ✅ Overall P95 latency < 500ms

## Troubleshooting

### API Not Responding

```bash
# Check if service is running
curl http://localhost:8080/health

# Check service logs
# Look for Phase 2 log messages:
# - "⚡ [Phase 2] Fast path succeeded"
# - "📊 [Phase 2] Confidence calibrated"
# - "🔗 [Phase 2] Enriching codes with crosswalk validation"
```

### Missing Codes

- Verify database migration 040 was applied
- Check that indexes exist: `supabase-migrations/verify_040_migration.sql`
- Verify RPC functions exist in database

### Low Confidence Scores

- Check confidence calibration logs
- Verify code agreement calculation
- Review strategy scores

### Slow Performance

- Check fast path hit rate
- Verify database indexes are being used
- Review query execution plans
