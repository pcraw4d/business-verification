-- =====================================================
-- Code Metadata Coverage Verification
-- Purpose: Verify code metadata coverage meets Phase 1 goals
-- Date: 2025-01-27
-- OPTIMIZATION: Accuracy Plan Enhancement - Phase 1, Task 1.1
-- =====================================================

-- =====================================================
-- Part 1: Basic Statistics
-- =====================================================

-- Total records
SELECT 
    'Total Records' AS metric,
    COUNT(*) AS count,
    CASE 
        WHEN COUNT(*) >= 500 THEN '✅ PASS - 500+ codes'
        ELSE '❌ FAIL - Below 500 codes'
    END AS status
FROM code_metadata;

-- Count by code type
SELECT 
    'Records by Code Type' AS metric,
    code_type,
    COUNT(*) AS count,
    ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER (), 2) AS percentage
FROM code_metadata
GROUP BY code_type
ORDER BY code_type;

-- =====================================================
-- Part 2: Official Description Coverage
-- =====================================================

-- Records with official descriptions
SELECT 
    'Records with Official Descriptions' AS metric,
    COUNT(*) AS count,
    ROUND(COUNT(*) * 100.0 / (SELECT COUNT(*) FROM code_metadata), 2) AS percentage,
    CASE 
        WHEN COUNT(*) = (SELECT COUNT(*) FROM code_metadata) THEN '✅ PASS - 100% coverage'
        ELSE '❌ FAIL - Not all codes have descriptions'
    END AS status
FROM code_metadata
WHERE official_description IS NOT NULL AND official_description != '';

-- Records with official names
SELECT 
    'Records with Official Names' AS metric,
    COUNT(*) AS count,
    ROUND(COUNT(*) * 100.0 / (SELECT COUNT(*) FROM code_metadata), 2) AS percentage
FROM code_metadata
WHERE official_name IS NOT NULL AND official_name != '';

-- =====================================================
-- Part 3: Industry Coverage
-- =====================================================

-- Count by industry (from industry_mappings)
SELECT 
    'Industry Coverage' AS metric,
    industry_mappings->>'primary_industry' AS industry,
    COUNT(*) AS count
FROM code_metadata
WHERE industry_mappings != '{}'::jsonb 
  AND industry_mappings->>'primary_industry' IS NOT NULL
GROUP BY industry_mappings->>'primary_industry'
ORDER BY count DESC;

-- Total industries covered
SELECT 
    'Total Industries Covered' AS metric,
    COUNT(DISTINCT industry_mappings->>'primary_industry') AS count,
    CASE 
        WHEN COUNT(DISTINCT industry_mappings->>'primary_industry') >= 10 THEN '✅ PASS - 10+ industries'
        ELSE '❌ FAIL - Below 10 industries'
    END AS status
FROM code_metadata
WHERE industry_mappings != '{}'::jsonb 
  AND industry_mappings->>'primary_industry' IS NOT NULL;

-- =====================================================
-- Part 4: Code Distribution by Industry and Type
-- =====================================================

-- Codes per industry per code type
SELECT 
    'Codes per Industry per Code Type' AS metric,
    industry_mappings->>'primary_industry' AS industry,
    code_type,
    COUNT(*) AS count
FROM code_metadata
WHERE industry_mappings != '{}'::jsonb 
  AND industry_mappings->>'primary_industry' IS NOT NULL
GROUP BY industry_mappings->>'primary_industry', code_type
ORDER BY industry_mappings->>'primary_industry', code_type;

-- Verify 10+ codes per code type per major industry
SELECT 
    'Codes per Industry per Code Type (10+ Target)' AS metric,
    industry_mappings->>'primary_industry' AS industry,
    code_type,
    COUNT(*) AS count,
    CASE 
        WHEN COUNT(*) >= 10 THEN '✅ PASS'
        ELSE '⚠️ WARNING - Below 10 codes'
    END AS status
FROM code_metadata
WHERE industry_mappings != '{}'::jsonb 
  AND industry_mappings->>'primary_industry' IS NOT NULL
GROUP BY industry_mappings->>'primary_industry', code_type
HAVING COUNT(*) < 10
ORDER BY industry_mappings->>'primary_industry', code_type;

-- =====================================================
-- Part 5: Data Quality Checks
-- =====================================================

-- Records marked as official
SELECT 
    'Records Marked as Official' AS metric,
    COUNT(*) AS count,
    ROUND(COUNT(*) * 100.0 / (SELECT COUNT(*) FROM code_metadata), 2) AS percentage
FROM code_metadata
WHERE is_official = true;

-- Records marked as active
SELECT 
    'Records Marked as Active' AS metric,
    COUNT(*) AS count,
    ROUND(COUNT(*) * 100.0 / (SELECT COUNT(*) FROM code_metadata), 2) AS percentage,
    CASE 
        WHEN COUNT(*) = (SELECT COUNT(*) FROM code_metadata) THEN '✅ PASS - All active'
        ELSE '❌ FAIL - Some inactive codes'
    END AS status
FROM code_metadata
WHERE is_active = true;

-- =====================================================
-- Part 6: Summary Report
-- =====================================================

SELECT 
    '=== SUMMARY REPORT ===' AS section;

SELECT 
    'Total Codes' AS metric,
    COUNT(*) AS value,
    CASE 
        WHEN COUNT(*) >= 500 THEN '✅ PASS'
        ELSE '❌ FAIL'
    END AS status
FROM code_metadata

UNION ALL

SELECT 
    'Codes with Official Descriptions' AS metric,
    COUNT(*) AS value,
    CASE 
        WHEN COUNT(*) = (SELECT COUNT(*) FROM code_metadata) THEN '✅ PASS'
        ELSE '❌ FAIL'
    END AS status
FROM code_metadata
WHERE official_description IS NOT NULL AND official_description != ''

UNION ALL

SELECT 
    'Industries Covered' AS metric,
    COUNT(DISTINCT industry_mappings->>'primary_industry') AS value,
    CASE 
        WHEN COUNT(DISTINCT industry_mappings->>'primary_industry') >= 10 THEN '✅ PASS'
        ELSE '❌ FAIL'
    END AS status
FROM code_metadata
WHERE industry_mappings != '{}'::jsonb 
  AND industry_mappings->>'primary_industry' IS NOT NULL

UNION ALL

SELECT 
    'NAICS Codes' AS metric,
    COUNT(*) AS value,
    NULL AS status
FROM code_metadata
WHERE code_type = 'NAICS'

UNION ALL

SELECT 
    'SIC Codes' AS metric,
    COUNT(*) AS value,
    NULL AS status
FROM code_metadata
WHERE code_type = 'SIC'

UNION ALL

SELECT 
    'MCC Codes' AS metric,
    COUNT(*) AS value,
    NULL AS status
FROM code_metadata
WHERE code_type = 'MCC';

