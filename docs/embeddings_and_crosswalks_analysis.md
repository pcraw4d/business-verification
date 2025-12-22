# Embeddings and Crosswalks Analysis & Action Plan
**Date**: December 22, 2025  
**Status**: Analysis Complete - Ready for Implementation

## 1. Code Embeddings Analysis

### Current State ✅

**Data Population:**
- ✅ **3,138 total embeddings** populated
- ✅ **950 MCC** codes with embeddings
- ✅ **1,160 NAICS** codes with embeddings  
- ✅ **1,028 SIC** codes with embeddings
- ✅ **2,740 unique codes** covered
- ✅ Vector index exists (`idx_code_embeddings_vector`)

**Status**: ✅ **Embeddings are fully populated and ready to use**

### Usage Status

**Currently Active:**
- ✅ Layer 2 classification uses embeddings when Layer 1 confidence < 80%
- ✅ `match_code_embeddings` RPC function exists and works
- ✅ Embeddings are being used in production flow

**Recommendation**: ✅ **No action needed** - embeddings are working correctly

### Maintenance

**Script Available**: `scripts/precompute_embeddings.py`
- Can be run to update embeddings for new codes
- Uses `sentence-transformers/all-MiniLM-L6-v2` model
- Batch processing for efficiency

**Action Items:**
- ⏭️ Set up periodic embedding updates for new codes (optional)
- ⏭️ Monitor Layer 2 usage and performance

---

## 2. Crosswalks Analysis

### Current State

#### `code_metadata.crosswalk_data` (JSONB) ✅ **Currently Used**
- **Total entries**: 545
- **With crosswalks**: 293 (53.8%)
- **Breakdown**:
  - MCC: 85/154 (55.2%)
  - NAICS: 108/236 (45.8%)
  - SIC: 100/155 (64.5%)

**Status**: ✅ Active and working, but coverage could be improved

#### `industry_code_crosswalks` (Structured Table) ❌ **Not Used**
- **Schema**: Has `mcc_code`, `naics_code`, `sic_code` columns
- **Status**: Table exists but code doesn't query it
- **Data**: Need to verify population status

### Issues Identified

1. **Low Crosswalk Coverage**: Only 53.8% of codes have crosswalk data
2. **Unused Table**: `industry_code_crosswalks` exists but isn't being used
3. **No Fallback**: Code only queries `code_metadata`, no fallback implemented

### Recommendation: Hybrid Approach (Option C)

**Implementation Plan:**

1. **Update `GetCrosswalks()` to try both sources:**
   - First: Query `industry_code_crosswalks` (structured, faster)
   - Fallback: Query `code_metadata.crosswalk_data` (flexible, backward compatible)

2. **Populate `industry_code_crosswalks` from `code_metadata`:**
   - Migrate existing crosswalk data
   - Maintain both during transition

3. **Improve Crosswalk Coverage:**
   - Add missing crosswalks for common codes
   - Use official crosswalk mappings where available

---

## 3. Action Plan

### Priority 1: Implement Hybrid Crosswalk Approach

**Steps:**
1. ✅ Create migration to populate `industry_code_crosswalks` from `code_metadata`
2. ✅ Update `GetCrosswalks()` to query both sources
3. ✅ Test hybrid approach
4. ⏭️ Monitor performance improvements

### Priority 2: Improve Crosswalk Coverage

**Steps:**
1. ⏭️ Identify codes missing crosswalks
2. ⏭️ Add official crosswalk mappings
3. ⏭️ Update both `code_metadata` and `industry_code_crosswalks`

### Priority 3: Embeddings Maintenance (Optional)

**Steps:**
1. ⏭️ Set up periodic embedding updates
2. ⏭️ Monitor Layer 2 usage metrics
3. ⏭️ Optimize vector index if needed

---

## 4. Expected Benefits

### Hybrid Crosswalk Approach
- ✅ **Better Performance**: Structured queries are faster than JSONB
- ✅ **Backward Compatible**: Falls back to JSONB if structured data missing
- ✅ **Data Integrity**: Foreign keys ensure consistency
- ✅ **Easier Maintenance**: Structured data easier to update

### Improved Coverage
- ✅ **More Accurate**: More codes will have crosswalk mappings
- ✅ **Better Code Generation**: Crosswalks help find related codes
- ✅ **Higher Accuracy**: More complete data improves classification

---

## Next Steps

1. ✅ Create migration to populate `industry_code_crosswalks`
2. ✅ Update `GetCrosswalks()` function
3. ✅ Test and verify improvements
4. ⏭️ Monitor performance and accuracy gains

