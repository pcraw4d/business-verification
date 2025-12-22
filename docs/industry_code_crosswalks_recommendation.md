# Industry Code Crosswalks - Recommendation
**Date**: December 22, 2025  
**Status**: Recommendation

## Current Situation

### Two Crosswalk Storage Mechanisms

1. **`code_metadata.crosswalk_data` (JSONB)** ✅ **Currently Used**
   - **Location**: `code_metadata` table, `crosswalk_data` JSONB column
   - **Structure**: `{"naics": ["541511", "541512"], "sic": ["7371"], "mcc": ["5734"]}`
   - **Used by**: `GetCrosswalks()` function in `supabase_repository.go` (line 2269)
   - **Status**: ✅ Active and working

2. **`industry_code_crosswalks` (Structured Table)** ❌ **Not Used**
   - **Location**: Dedicated `industry_code_crosswalks` table
   - **Structure**: Relational table with columns: `industry_id`, `mcc_code`, `naics_code`, `sic_code`, `confidence_score`, `is_primary`
   - **Used by**: ❌ Not referenced in code
   - **Status**: Table exists but unused

## Analysis

### Code Usage

The `GetCrosswalks()` function in `internal/classification/repository/supabase_repository.go`:
- **Line 2254**: Comment mentions "fall back to industry_code_crosswalks table"
- **Line 2269**: Actually queries `code_metadata` table only
- **No fallback implementation**: The fallback to `industry_code_crosswalks` is not implemented

### Pros and Cons

#### `code_metadata.crosswalk_data` (Current - JSONB)

**Pros:**
- ✅ Flexible - can store any crosswalk structure
- ✅ Single source of truth with code metadata
- ✅ Easy to update (single JSONB update)
- ✅ Already working and tested
- ✅ Supports hierarchical crosswalks (parent/child codes)

**Cons:**
- ❌ JSONB queries can be slower for complex filtering
- ❌ No foreign key constraints (data integrity risk)
- ❌ Harder to query for specific crosswalk patterns
- ❌ No direct industry_id linkage (must join through classification_codes)

#### `industry_code_crosswalks` (Unused - Structured Table)

**Pros:**
- ✅ Fast relational queries with indexes
- ✅ Foreign key constraints (data integrity)
- ✅ Direct industry_id linkage
- ✅ Better for complex crosswalk queries
- ✅ Supports confidence scores and primary flags
- ✅ Easier to maintain and audit

**Cons:**
- ❌ Less flexible (fixed schema)
- ❌ Requires separate table maintenance
- ❌ Not currently integrated in code
- ❌ Would require migration from JSONB data

## Recommendation

### **Option A: Migrate to `industry_code_crosswalks` (Recommended for Long-term)**

**Rationale:**
- Better performance for crosswalk queries
- Stronger data integrity with foreign keys
- Direct industry linkage for better classification
- Easier to maintain and extend

**Implementation Steps:**
1. Create migration to populate `industry_code_crosswalks` from `code_metadata.crosswalk_data`
2. Update `GetCrosswalks()` to query `industry_code_crosswalks` first, fallback to `code_metadata`
3. Add indexes on `industry_code_crosswalks` for performance
4. Deprecate `code_metadata.crosswalk_data` (keep for backward compatibility)
5. Update data population scripts to write to both tables initially

**Migration SQL Example:**
```sql
-- Populate industry_code_crosswalks from code_metadata
INSERT INTO industry_code_crosswalks (industry_id, mcc_code, naics_code, sic_code, confidence_score, is_primary)
SELECT 
    cc.industry_id,
    CASE WHEN cm.crosswalk_data->>'mcc' IS NOT NULL 
         THEN (cm.crosswalk_data->'mcc'->>0)::varchar(10) 
         ELSE NULL END as mcc_code,
    CASE WHEN cm.crosswalk_data->>'naics' IS NOT NULL 
         THEN (cm.crosswalk_data->'naics'->>0)::varchar(10) 
         ELSE NULL END as naics_code,
    CASE WHEN cm.crosswalk_data->>'sic' IS NOT NULL 
         THEN (cm.crosswalk_data->'sic'->>0)::varchar(10) 
         ELSE NULL END as sic_code,
    0.80 as confidence_score,  -- Default confidence
    true as is_primary
FROM code_metadata cm
JOIN classification_codes cc ON cc.code_type = cm.code_type AND cc.code = cm.code
WHERE cm.crosswalk_data IS NOT NULL
  AND cm.crosswalk_data != '{}'::jsonb
ON CONFLICT (industry_id, mcc_code, naics_code, sic_code) DO NOTHING;
```

### **Option B: Keep `code_metadata.crosswalk_data` (Recommended for Short-term)**

**Rationale:**
- Already working and tested
- No migration risk
- Flexible for future changes
- Lower implementation effort

**Implementation Steps:**
1. Keep current implementation
2. Document that `industry_code_crosswalks` is deprecated
3. Consider removing `industry_code_crosswalks` table in future cleanup
4. Optimize JSONB queries with GIN indexes (already exists)

### **Option C: Hybrid Approach (Recommended for Transition)**

**Rationale:**
- Best of both worlds
- Gradual migration path
- Backward compatible

**Implementation Steps:**
1. Update `GetCrosswalks()` to try `industry_code_crosswalks` first
2. Fallback to `code_metadata.crosswalk_data` if not found
3. Populate `industry_code_crosswalks` from `code_metadata` gradually
4. Monitor usage and performance
5. Eventually deprecate `code_metadata.crosswalk_data`

**Updated `GetCrosswalks()` Logic:**
```go
// Try industry_code_crosswalks first (structured, faster)
response, _, err := postgrestClient.
    From("industry_code_crosswalks").
    Select("mcc_code,naics_code,sic_code,code_description", "", false).
    Eq("industry_id", industryID).
    // ... filter by code type
    Execute()

if err == nil && len(results) > 0 {
    return results
}

// Fallback to code_metadata.crosswalk_data (flexible, backward compatible)
// ... existing code_metadata query ...
```

## Final Recommendation

### **Short-term (Next Sprint): Option B**
- Keep `code_metadata.crosswalk_data` as-is
- Document `industry_code_crosswalks` as deprecated/unused
- Focus on fixing critical issues (type mismatches, timeouts)

### **Medium-term (Next Quarter): Option C**
- Implement hybrid approach
- Gradually migrate data to `industry_code_crosswalks`
- Monitor performance improvements

### **Long-term (Future): Option A**
- Fully migrate to `industry_code_crosswalks`
- Remove `code_metadata.crosswalk_data` dependency
- Optimize for structured queries

## Action Items

1. ✅ **Immediate**: Document current state (this document)
2. ⏭️ **Short-term**: Keep `code_metadata.crosswalk_data` (no changes)
3. ⏭️ **Medium-term**: Implement hybrid approach (Option C)
4. ⏭️ **Long-term**: Full migration to `industry_code_crosswalks` (Option A)

## Notes

- The `industry_code_crosswalks` table was created in migration `003_risk_keywords_schema.sql`
- It has proper indexes and foreign keys
- The table structure is well-designed for crosswalk queries
- Migration would require data transformation from JSONB to structured format

