-- =====================================================
-- Quick Code Metadata Verification
-- Purpose: Fast verification that code metadata is loaded
-- =====================================================

-- Quick summary
SELECT 
    'Total Codes' AS metric,
    COUNT(*) AS count
FROM code_metadata;

-- Count by type
SELECT 
    'Codes by Type' AS metric,
    code_type,
    COUNT(*) AS count
FROM code_metadata
GROUP BY code_type
ORDER BY code_type;

-- Data completeness
SELECT 
    'Data Completeness' AS metric,
    COUNT(CASE WHEN crosswalk_data != '{}'::jsonb THEN 1 END) AS "With Crosswalk",
    COUNT(CASE WHEN hierarchy != '{}'::jsonb THEN 1 END) AS "With Hierarchy",
    COUNT(CASE WHEN industry_mappings != '{}'::jsonb THEN 1 END) AS "With Industry Mapping"
FROM code_metadata;

-- Health check
SELECT 
    CASE WHEN COUNT(*) >= 100 THEN '✅ PASS' ELSE '❌ FAIL' END AS "Total Codes (>=100)",
    CASE WHEN COUNT(*) FILTER (WHERE code_type = 'NAICS') >= 30 THEN '✅ PASS' ELSE '❌ FAIL' END AS "NAICS (>=30)",
    CASE WHEN COUNT(*) FILTER (WHERE code_type = 'SIC') >= 20 THEN '✅ PASS' ELSE '❌ FAIL' END AS "SIC (>=20)",
    CASE WHEN COUNT(*) FILTER (WHERE code_type = 'MCC') >= 50 THEN '✅ PASS' ELSE '❌ FAIL' END AS "MCC (>=50)"
FROM code_metadata;

