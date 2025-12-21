# Code Generation Manual Test Plan

**Date**: December 21, 2025  
**Investigation Track**: Track 4.2 - NAICS/SIC Code Generation Investigation  
**Status**: Pending Manual Testing

## Executive Summary

This document outlines the manual testing plan for NAICS/SIC code generation to verify that codes are being generated correctly and identify why accuracy is 0%.

---

## Test Results Summary

### Current Status

| Metric | Value | Target | Status |
|--------|-------|--------|--------|
| **NAICS Accuracy** | 0% | ≥70% | ❌ **Critical** |
| **SIC Accuracy** | 0% | ≥70% | ❌ **Critical** |
| **MCC Top 1 Accuracy** | 10% | ≥60% | ❌ **Low** |
| **Code Generation Rate** | 23.1% | ≥90% | ❌ **Low** (Fixed: threshold lowered) |

---

## Potential Root Causes

### 1. Database Function Missing

**Issue**: `get_codes_by_trigram_similarity` RPC function may not exist in Supabase

**Location**: `internal/classification/repository/supabase_repository.go:2033`

**Code**:
```go
url := fmt.Sprintf("%s/rest/v1/rpc/get_codes_by_trigram_similarity", r.client.GetURL())
```

**Evidence**: 
- Plan mentions "GetCodesByTrigramSimilarity returned status 404"
- Function may not be created in database migrations

**Fix Required**:
- Check if function exists in Supabase
- Create function if missing
- Verify function signature matches code expectations

### 2. Database Data Missing

**Issue**: NAICS/SIC code data may not be populated in database

**Tables to Check**:
- `classification_codes` (where code_type = 'NAICS' or 'SIC')
- `code_keywords` (keywords for NAICS/SIC codes)
- `naics_codes` (if separate table exists)
- `sic_codes` (if separate table exists)

**Fix Required**:
- Verify data exists: `SELECT COUNT(*) FROM classification_codes WHERE code_type IN ('NAICS', 'SIC')`
- Populate data if missing
- Verify code_keywords have entries for NAICS/SIC codes

### 3. Code Generation Logic Issue

**Issue**: Code generation may not be calling NAICS/SIC generation correctly

**Location**: `internal/classification/classifier.go:256-320`

**Fix Required**:
- Verify `generateCodesInParallel` includes NAICS/SIC
- Check if NAICS/SIC generation is being skipped
- Verify code generation threshold fix (0.15) applies to NAICS/SIC

### 4. Code Matching Algorithm Issue

**Issue**: Code matching algorithm may not be working for NAICS/SIC

**Location**: `internal/classification/classifier.go:339-375`

**Fix Required**:
- Verify keyword matching works for NAICS/SIC
- Check if minRelevance threshold is too high
- Verify industry matching works for NAICS/SIC

---

## Manual Test Plan

### Test 1: Verify Database Function Exists

**Steps**:
1. Connect to Supabase database
2. Check if function exists:
   ```sql
   SELECT routine_name, routine_type 
   FROM information_schema.routines 
   WHERE routine_schema = 'public' 
   AND routine_name = 'get_codes_by_trigram_similarity';
   ```
3. If missing, check for similar functions:
   ```sql
   SELECT routine_name 
   FROM information_schema.routines 
   WHERE routine_schema = 'public' 
   AND routine_name LIKE '%trigram%';
   ```

**Expected Result**: Function exists and is callable

**If Missing**: Create function or update code to use existing function

---

### Test 2: Verify Database Data Exists

**Steps**:
1. Check NAICS codes:
   ```sql
   SELECT COUNT(*) as naics_count 
   FROM classification_codes 
   WHERE code_type = 'NAICS';
   ```
2. Check SIC codes:
   ```sql
   SELECT COUNT(*) as sic_count 
   FROM classification_codes 
   WHERE code_type = 'SIC';
   ```
3. Check code_keywords for NAICS:
   ```sql
   SELECT COUNT(*) as naics_keywords 
   FROM code_keywords 
   WHERE code_type = 'NAICS';
   ```
4. Check code_keywords for SIC:
   ```sql
   SELECT COUNT(*) as sic_keywords 
   FROM code_keywords 
   WHERE code_type = 'SIC';
   ```

**Expected Result**: 
- NAICS codes: > 1000
- SIC codes: > 1000
- NAICS keywords: > 5000
- SIC keywords: > 5000

**If Missing**: Populate data using migration scripts

---

### Test 3: Test Code Generation Manually

**Steps**:
1. Create test request with known industry (e.g., "Technology")
2. Call code generation function directly
3. Verify codes are generated:
   - Check if MCC codes are generated
   - Check if NAICS codes are generated
   - Check if SIC codes are generated
4. Verify code content:
   - Check if codes have descriptions
   - Check if codes have confidence scores
   - Check if codes are ranked correctly

**Test Request**:
```go
keywords := []string{"software", "development", "technology"}
industry := "Technology"
confidence := 0.8

codes, err := codeGenerator.GenerateClassificationCodes(ctx, keywords, industry, confidence)
```

**Expected Result**: 
- MCC codes: > 0
- NAICS codes: > 0
- SIC codes: > 0

---

### Test 4: Test Database Function Call

**Steps**:
1. Call `get_codes_by_trigram_similarity` RPC function directly
2. Test with different parameters:
   - codeType: "NAICS", "SIC", "MCC"
   - industryName: "Technology", "Finance", "Healthcare"
   - threshold: 0.3, 0.5, 0.7
   - limit: 3, 5, 10
3. Verify response format matches code expectations

**Test Call**:
```bash
curl -X POST "https://<supabase-url>/rest/v1/rpc/get_codes_by_trigram_similarity" \
  -H "Content-Type: application/json" \
  -H "apikey: <api-key>" \
  -H "Authorization: Bearer <service-key>" \
  -d '{
    "p_code_type": "NAICS",
    "p_industry_name": "Technology",
    "p_threshold": 0.3,
    "p_limit": 3
  }'
```

**Expected Result**: Returns JSON array with code, description, similarity

---

### Test 5: Test Code Generation with Low Confidence

**Steps**:
1. Test with confidence = 0.15 (new threshold)
2. Test with confidence = 0.10 (below threshold)
3. Verify codes are generated for 0.15 but not for 0.10

**Expected Result**: 
- Confidence 0.15: Codes generated
- Confidence 0.10: Codes not generated (threshold working)

---

## Test Script

Create `scripts/test_code_generation.go`:

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/your-org/kyb-platform/internal/classification"
    "github.com/your-org/kyb-platform/internal/classification/repository"
)

func main() {
    ctx := context.Background()
    
    // Initialize code generator
    repo := repository.NewSupabaseKeywordRepository(...)
    codeGenerator := classification.NewClassificationCodeGenerator(repo, nil)
    
    // Test cases
    testCases := []struct {
        name       string
        keywords   []string
        industry   string
        confidence float64
    }{
        {"Technology", []string{"software", "development"}, "Technology", 0.8},
        {"Finance", []string{"banking", "financial"}, "Financial Services", 0.7},
        {"Low Confidence", []string{"business"}, "General Business", 0.15},
    }
    
    for _, tc := range testCases {
        fmt.Printf("\n=== Testing: %s ===\n", tc.name)
        codes, err := codeGenerator.GenerateClassificationCodes(
            ctx,
            tc.keywords,
            tc.industry,
            tc.confidence,
        )
        if err != nil {
            log.Printf("Error: %v", err)
            continue
        }
        
        fmt.Printf("MCC codes: %d\n", len(codes.MCC))
        fmt.Printf("NAICS codes: %d\n", len(codes.NAICS))
        fmt.Printf("SIC codes: %d\n", len(codes.SIC))
        
        if len(codes.NAICS) > 0 {
            fmt.Printf("First NAICS: %s - %s (confidence: %.2f)\n",
                codes.NAICS[0].Code, codes.NAICS[0].Description, codes.NAICS[0].Confidence)
        }
        if len(codes.SIC) > 0 {
            fmt.Printf("First SIC: %s - %s (confidence: %.2f)\n",
                codes.SIC[0].Code, codes.SIC[0].Description, codes.SIC[0].Confidence)
        }
    }
}
```

---

## Expected Findings

### If Database Function Missing

**Fix**: Create `get_codes_by_trigram_similarity` function in Supabase

**SQL**:
```sql
CREATE OR REPLACE FUNCTION get_codes_by_trigram_similarity(
    p_code_type text,
    p_industry_name text,
    p_threshold float DEFAULT 0.3,
    p_limit int DEFAULT 3
)
RETURNS TABLE (
    code text,
    description text,
    similarity float
)
LANGUAGE plpgsql
AS $$
BEGIN
    RETURN QUERY
    SELECT 
        cc.code,
        cc.description,
        similarity(cc.description, p_industry_name) as similarity
    FROM classification_codes cc
    WHERE cc.code_type = p_code_type
        AND similarity(cc.description, p_industry_name) >= p_threshold
    ORDER BY similarity DESC
    LIMIT p_limit;
END;
$$;
```

### If Database Data Missing

**Fix**: Populate NAICS/SIC code data using migration scripts

**Script**: `scripts/populate_all_classification_codes_comprehensive.sql`

---

## Next Steps

1. [ ] Run Test 1: Verify database function exists
2. [ ] Run Test 2: Verify database data exists
3. [ ] Run Test 3: Test code generation manually
4. [ ] Run Test 4: Test database function call
5. [ ] Run Test 5: Test with low confidence
6. [ ] Document findings
7. [ ] Implement fixes based on findings

---

**Document Status**: Test Plan Created  
**Next Steps**: Execute manual tests

