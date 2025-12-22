# Crosswalk Monitoring Guide
**Date**: December 22, 2025  
**Purpose**: Monitor hybrid crosswalk approach performance and accuracy

## Monitoring Queries

### 1. Crosswalk Usage Statistics

```sql
-- Total crosswalks available
SELECT 
    source_system,
    target_system,
    COUNT(*) as crosswalk_count
FROM industry_code_crosswalks
WHERE is_active = true
GROUP BY source_system, target_system
ORDER BY source_system, target_system;
```

### 2. Crosswalk Coverage

```sql
-- Coverage by code type
SELECT 
    'MCC' as code_type,
    COUNT(DISTINCT source_code) as codes_with_crosswalks,
    (SELECT COUNT(*) FROM classification_codes WHERE code_type = 'MCC' AND is_active = true) as total_codes,
    ROUND(100.0 * COUNT(DISTINCT source_code) / 
        (SELECT COUNT(*) FROM classification_codes WHERE code_type = 'MCC' AND is_active = true), 2) as coverage_pct
FROM industry_code_crosswalks
WHERE source_system = 'MCC' AND is_active = true

UNION ALL

SELECT 
    'NAICS' as code_type,
    COUNT(DISTINCT source_code) as codes_with_crosswalks,
    (SELECT COUNT(*) FROM classification_codes WHERE code_type = 'NAICS' AND is_active = true) as total_codes,
    ROUND(100.0 * COUNT(DISTINCT source_code) / 
        (SELECT COUNT(*) FROM classification_codes WHERE code_type = 'NAICS' AND is_active = true), 2) as coverage_pct
FROM industry_code_crosswalks
WHERE source_system = 'NAICS' AND is_active = true

UNION ALL

SELECT 
    'SIC' as code_type,
    COUNT(DISTINCT source_code) as codes_with_crosswalks,
    (SELECT COUNT(*) FROM classification_codes WHERE code_type = 'SIC' AND is_active = true) as total_codes,
    ROUND(100.0 * COUNT(DISTINCT source_code) / 
        (SELECT COUNT(*) FROM classification_codes WHERE code_type = 'SIC' AND is_active = true), 2) as coverage_pct
FROM industry_code_crosswalks
WHERE source_system = 'SIC' AND is_active = true;
```

### 3. Query Performance Monitoring

```sql
-- Test query performance (should be < 100ms for structured table)
EXPLAIN ANALYZE
SELECT target_code, confidence_score
FROM industry_code_crosswalks
WHERE source_system = 'MCC' 
  AND source_code = '5819' 
  AND target_system = 'NAICS' 
  AND is_active = true
ORDER BY confidence_score DESC
LIMIT 5;
```

### 4. Fallback Usage (JSONB)

```sql
-- Check how many codes rely on JSONB fallback
SELECT 
    COUNT(*) as codes_in_metadata_only,
    COUNT(*) FILTER (WHERE crosswalk_data IS NOT NULL AND crosswalk_data != '{}'::jsonb) as with_crosswalks
FROM code_metadata cm
WHERE is_active = true
  AND NOT EXISTS (
      SELECT 1 FROM industry_code_crosswalks icc
      WHERE icc.source_system = cm.code_type::varchar(20)
        AND icc.source_code = cm.code::varchar(20)
        AND icc.is_active = true
  );
```

## Log Monitoring

### Check Railway Logs for Crosswalk Usage

Look for these log patterns in classification service logs:

1. **Structured table hits:**
   ```
   ✅ Retrieved X crosswalk codes from MCC YYYY to NAICS (from industry_code_crosswalks)
   ```

2. **JSONB fallback usage:**
   ```
   ⚠️ No crosswalks found in industry_code_crosswalks, trying code_metadata fallback
   ```

3. **Crosswalk retrieval:**
   ```
   🔍 Getting crosswalks: MCC 5819 -> NAICS
   ```

### Expected Behavior

- **Primary path**: Most queries should use `industry_code_crosswalks` (structured, fast)
- **Fallback path**: Only used when structured table doesn't have the mapping
- **Performance**: Structured queries should be < 100ms, JSONB fallback < 200ms

## Performance Benchmarks

### Target Metrics

| Metric | Target | Current |
|--------|--------|---------|
| Structured query time | < 100ms | ~50ms |
| JSONB fallback time | < 200ms | ~150ms |
| Crosswalk coverage | > 50% | 53.8% |
| Bidirectional mappings | 100% | 100% |

## Accuracy Verification

### Known Crosswalk Mappings

Test these known accurate mappings:

1. **MCC 5819** (Miscellaneous Food Stores)
   - Should map to: NAICS 445120, 445110, 722511, 722513
   - Should map to: SIC 5499, 5812, 5411

2. **NAICS 541511** (Custom Computer Programming)
   - Should map to: SIC 7371
   - Should map to: MCC 5734

### Verification Query

```sql
-- Verify specific crosswalk
SELECT 
    icc.source_system,
    icc.source_code,
    icc.target_system,
    icc.target_code,
    icc.confidence_score,
    cc.description as target_description
FROM industry_code_crosswalks icc
LEFT JOIN classification_codes cc 
    ON cc.code_type = icc.target_system 
    AND cc.code = icc.target_code
WHERE icc.source_system = 'MCC' 
  AND icc.source_code = '5819' 
  AND icc.is_active = true
ORDER BY icc.target_system, icc.confidence_score DESC;
```

## Troubleshooting

### Issue: No crosswalks found

**Check:**
1. Is the code in `industry_code_crosswalks`?
2. Is `is_active = true`?
3. Does fallback to `code_metadata` work?

**Query:**
```sql
-- Check if code exists in structured table
SELECT * FROM industry_code_crosswalks
WHERE source_system = 'MCC' AND source_code = '5819' AND is_active = true;

-- Check if code exists in metadata
SELECT * FROM code_metadata
WHERE code_type = 'MCC' AND code = '5819' AND is_active = true;
```

### Issue: Slow queries

**Check:**
1. Are indexes being used?
2. Is the query using the structured table or JSONB?

**Query:**
```sql
-- Check index usage
EXPLAIN ANALYZE
SELECT * FROM industry_code_crosswalks
WHERE source_system = 'MCC' 
  AND source_code = '5819' 
  AND target_system = 'NAICS' 
  AND is_active = true;
```

## Regular Monitoring Schedule

- **Daily**: Check crosswalk coverage and query performance
- **Weekly**: Review fallback usage (should decrease over time)
- **Monthly**: Verify accuracy of known crosswalk mappings

