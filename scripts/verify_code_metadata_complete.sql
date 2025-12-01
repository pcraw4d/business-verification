-- =====================================================
-- Comprehensive Code Metadata Verification Query
-- Purpose: Verify all code metadata has been successfully added
-- =====================================================

-- =====================================================
-- Part 1: Basic Counts and Statistics
-- =====================================================

SELECT 
    '=== BASIC STATISTICS ===' AS section;

-- Total records
SELECT 
    'Total Records' AS metric,
    COUNT(*) AS count,
    'Total codes in code_metadata table' AS description
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

-- Count by official status
SELECT 
    'Records by Official Status' AS metric,
    is_official,
    COUNT(*) AS count,
    ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER (), 2) AS percentage
FROM code_metadata
GROUP BY is_official
ORDER BY is_official DESC;

-- Count by active status
SELECT 
    'Records by Active Status' AS metric,
    is_active,
    COUNT(*) AS count,
    ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER (), 2) AS percentage
FROM code_metadata
GROUP BY is_active
ORDER BY is_active DESC;

-- =====================================================
-- Part 2: Data Completeness Checks
-- =====================================================

SELECT 
    '=== DATA COMPLETENESS ===' AS section;

-- Records with crosswalk data
SELECT 
    'Records with Crosswalk Data' AS metric,
    COUNT(*) AS count,
    ROUND(COUNT(*) * 100.0 / (SELECT COUNT(*) FROM code_metadata), 2) AS percentage
FROM code_metadata
WHERE crosswalk_data != '{}'::jsonb AND crosswalk_data IS NOT NULL;

-- Records with hierarchy data
SELECT 
    'Records with Hierarchy Data' AS metric,
    COUNT(*) AS count,
    ROUND(COUNT(*) * 100.0 / (SELECT COUNT(*) FROM code_metadata), 2) AS percentage
FROM code_metadata
WHERE hierarchy != '{}'::jsonb AND hierarchy IS NOT NULL;

-- Records with industry mappings
SELECT 
    'Records with Industry Mappings' AS metric,
    COUNT(*) AS count,
    ROUND(COUNT(*) * 100.0 / (SELECT COUNT(*) FROM code_metadata), 2) AS percentage
FROM code_metadata
WHERE industry_mappings != '{}'::jsonb AND industry_mappings IS NOT NULL;

-- Records with official descriptions
SELECT 
    'Records with Official Descriptions' AS metric,
    COUNT(*) AS count,
    ROUND(COUNT(*) * 100.0 / (SELECT COUNT(*) FROM code_metadata), 2) AS percentage
FROM code_metadata
WHERE official_description IS NOT NULL AND official_description != '';

-- Records with metadata JSONB (tags could be stored here)
SELECT 
    'Records with Additional Metadata' AS metric,
    COUNT(*) AS count,
    ROUND(COUNT(*) * 100.0 / (SELECT COUNT(*) FROM code_metadata), 2) AS percentage
FROM code_metadata
WHERE metadata != '{}'::jsonb AND metadata IS NOT NULL;

-- =====================================================
-- Part 3: Crosswalk Data Analysis
-- =====================================================

SELECT 
    '=== CROSSWALK DATA ANALYSIS ===' AS section;

-- Crosswalk coverage by code type
SELECT 
    'Crosswalk Coverage by Code Type' AS metric,
    code_type,
    COUNT(*) AS total_codes,
    COUNT(CASE WHEN crosswalk_data != '{}'::jsonb THEN 1 END) AS codes_with_crosswalk,
    ROUND(COUNT(CASE WHEN crosswalk_data != '{}'::jsonb THEN 1 END) * 100.0 / COUNT(*), 2) AS coverage_percentage
FROM code_metadata
GROUP BY code_type
ORDER BY code_type;

-- Average crosswalk links per code
SELECT 
    'Average Crosswalk Links' AS metric,
    code_type,
    ROUND(AVG(
        COALESCE(jsonb_array_length(crosswalk_data->'naics'), 0) +
        COALESCE(jsonb_array_length(crosswalk_data->'sic'), 0) +
        COALESCE(jsonb_array_length(crosswalk_data->'mcc'), 0)
    ), 2) AS avg_links_per_code
FROM code_metadata
WHERE crosswalk_data != '{}'::jsonb
GROUP BY code_type
ORDER BY code_type;

-- Top codes by crosswalk links
SELECT 
    'Top Codes by Crosswalk Links' AS metric,
    code_type,
    code,
    official_name,
    (
        COALESCE(jsonb_array_length(crosswalk_data->'naics'), 0) +
        COALESCE(jsonb_array_length(crosswalk_data->'sic'), 0) +
        COALESCE(jsonb_array_length(crosswalk_data->'mcc'), 0)
    ) AS total_links
FROM code_metadata
WHERE crosswalk_data != '{}'::jsonb
ORDER BY total_links DESC
LIMIT 10;

-- =====================================================
-- Part 4: Hierarchy Data Analysis
-- =====================================================

SELECT 
    '=== HIERARCHY DATA ANALYSIS ===' AS section;

-- Hierarchy coverage by code type
SELECT 
    'Hierarchy Coverage by Code Type' AS metric,
    code_type,
    COUNT(*) AS total_codes,
    COUNT(CASE WHEN hierarchy != '{}'::jsonb THEN 1 END) AS codes_with_hierarchy,
    ROUND(COUNT(CASE WHEN hierarchy != '{}'::jsonb THEN 1 END) * 100.0 / COUNT(*), 2) AS coverage_percentage
FROM code_metadata
GROUP BY code_type
ORDER BY code_type;

-- Hierarchy parent types (for NAICS)
-- Note: The hierarchy JSONB may not have a 'level' field in all cases
SELECT 
    'NAICS Hierarchy Parent Types' AS metric,
    hierarchy->>'parent_type' AS parent_type,
    COUNT(*) AS count
FROM code_metadata
WHERE code_type = 'NAICS' AND hierarchy != '{}'::jsonb AND hierarchy->>'parent_type' IS NOT NULL
GROUP BY hierarchy->>'parent_type'
ORDER BY parent_type;

-- =====================================================
-- Part 5: Industry Mappings Analysis
-- =====================================================

SELECT 
    '=== INDUSTRY MAPPINGS ANALYSIS ===' AS section;

-- Industry mappings coverage
SELECT 
    'Industry Mappings Coverage' AS metric,
    code_type,
    COUNT(*) AS total_codes,
    COUNT(CASE WHEN industry_mappings != '{}'::jsonb THEN 1 END) AS codes_with_mappings,
    ROUND(COUNT(CASE WHEN industry_mappings != '{}'::jsonb THEN 1 END) * 100.0 / COUNT(*), 2) AS coverage_percentage
FROM code_metadata
GROUP BY code_type
ORDER BY code_type;

-- Extract and count unique primary industries
WITH industry_extract AS (
    SELECT 
        jsonb_array_elements_text(industry_mappings->'primary') AS primary_industry
    FROM code_metadata
    WHERE industry_mappings != '{}'::jsonb
)
SELECT 
    'Top Primary Industries' AS metric,
    primary_industry,
    COUNT(*) AS code_count
FROM industry_extract
GROUP BY primary_industry
ORDER BY code_count DESC
LIMIT 15;

-- =====================================================
-- Part 6: Sample Records for Quality Check
-- =====================================================

SELECT 
    '=== SAMPLE RECORDS (Quality Check) ===' AS section;

-- Sample NAICS codes
SELECT 
    'Sample NAICS Codes' AS metric,
    code,
    official_name,
    CASE WHEN crosswalk_data != '{}'::jsonb THEN 'Yes' ELSE 'No' END AS has_crosswalk,
    CASE WHEN hierarchy != '{}'::jsonb THEN 'Yes' ELSE 'No' END AS has_hierarchy,
    CASE WHEN industry_mappings != '{}'::jsonb THEN 'Yes' ELSE 'No' END AS has_industry_mapping
FROM code_metadata
WHERE code_type = 'NAICS'
ORDER BY code
LIMIT 5;

-- Sample SIC codes
SELECT 
    'Sample SIC Codes' AS metric,
    code,
    official_name,
    CASE WHEN crosswalk_data != '{}'::jsonb THEN 'Yes' ELSE 'No' END AS has_crosswalk,
    CASE WHEN hierarchy != '{}'::jsonb THEN 'Yes' ELSE 'No' END AS has_hierarchy,
    CASE WHEN industry_mappings != '{}'::jsonb THEN 'Yes' ELSE 'No' END AS has_industry_mapping
FROM code_metadata
WHERE code_type = 'SIC'
ORDER BY code
LIMIT 5;

-- Sample MCC codes
SELECT 
    'Sample MCC Codes' AS metric,
    code,
    official_name,
    CASE WHEN crosswalk_data != '{}'::jsonb THEN 'Yes' ELSE 'No' END AS has_crosswalk,
    CASE WHEN hierarchy != '{}'::jsonb THEN 'Yes' ELSE 'No' END AS has_hierarchy,
    CASE WHEN industry_mappings != '{}'::jsonb THEN 'Yes' ELSE 'No' END AS has_industry_mapping
FROM code_metadata
WHERE code_type = 'MCC'
ORDER BY code
LIMIT 5;

-- =====================================================
-- Part 7: Expected Code Verification
-- =====================================================

SELECT 
    '=== EXPECTED CODE VERIFICATION ===' AS section;

-- Check for specific expected codes (from populate_code_metadata.sql)
SELECT 
    'Expected Codes from populate_code_metadata.sql' AS metric,
    code_type,
    code,
    official_name,
    CASE WHEN EXISTS (
        SELECT 1 FROM code_metadata cm 
        WHERE cm.code_type = expected.code_type 
        AND cm.code = expected.code
    ) THEN '✅ Found' ELSE '❌ Missing' END AS status
FROM (VALUES
    ('NAICS', '541511', 'Custom Computer Programming Services'),
    ('NAICS', '522110', 'Commercial Banking'),
    ('NAICS', '621111', 'Offices of Physicians'),
    ('SIC', '7371', 'Computer Programming Services'),
    ('SIC', '6021', 'National Commercial Banks'),
    ('SIC', '8011', 'Offices and Clinics of Doctors of Medicine'),
    ('MCC', '5734', 'Computer Software Stores'),
    ('MCC', '6010', 'Financial Institutions - Manual Cash Disbursements'),
    ('MCC', '8011', 'Doctors and Physicians')
) AS expected(code_type, code, official_name)
ORDER BY code_type, code;

-- Check for specific expected codes (from expand_code_metadata_clean.sql)
SELECT 
    'Expected Codes from expand_code_metadata_clean.sql' AS metric,
    code_type,
    code,
    official_name,
    CASE WHEN EXISTS (
        SELECT 1 FROM code_metadata cm 
        WHERE cm.code_type = expected.code_type 
        AND cm.code = expected.code
    ) THEN '✅ Found' ELSE '❌ Missing' END AS status
FROM (VALUES
    ('NAICS', '541330', 'Engineering Services'),
    ('NAICS', '541110', 'Offices of Lawyers'),
    ('NAICS', '611110', 'Elementary and Secondary Schools'),
    ('SIC', '7374', 'Computer Processing and Data Preparation Services'),
    ('SIC', '8111', 'Legal Services'),
    ('SIC', '8211', 'Elementary and Secondary Schools'),
    ('MCC', '5045', 'Computers, Computer Peripheral Equipment, Software'),
    ('MCC', '8111', 'Legal Services'),
    ('MCC', '8211', 'Elementary and Secondary Schools')
) AS expected(code_type, code, official_name)
ORDER BY code_type, code;

-- =====================================================
-- Part 8: Data Quality Checks
-- =====================================================

SELECT 
    '=== DATA QUALITY CHECKS ===' AS section;

-- Codes missing official descriptions
SELECT 
    'Codes Missing Official Descriptions' AS metric,
    code_type,
    COUNT(*) AS count
FROM code_metadata
WHERE official_description IS NULL OR official_description = ''
GROUP BY code_type
ORDER BY code_type;

-- Codes missing official names
SELECT 
    'Codes Missing Official Names' AS metric,
    code_type,
    COUNT(*) AS count
FROM code_metadata
WHERE official_name IS NULL OR official_name = ''
GROUP BY code_type
ORDER BY code_type;

-- Duplicate code check (should be 0)
SELECT 
    'Duplicate Codes Check' AS metric,
    code_type,
    code,
    COUNT(*) AS occurrence_count
FROM code_metadata
GROUP BY code_type, code
HAVING COUNT(*) > 1
ORDER BY code_type, code;

-- =====================================================
-- Part 9: Summary Report
-- =====================================================

SELECT 
    '=== SUMMARY REPORT ===' AS section;

WITH stats AS (
    SELECT 
        COUNT(*) AS total_codes,
        COUNT(DISTINCT code_type) AS code_types,
        COUNT(CASE WHEN crosswalk_data != '{}'::jsonb THEN 1 END) AS codes_with_crosswalk,
        COUNT(CASE WHEN hierarchy != '{}'::jsonb THEN 1 END) AS codes_with_hierarchy,
        COUNT(CASE WHEN industry_mappings != '{}'::jsonb THEN 1 END) AS codes_with_industry_mapping,
        COUNT(CASE WHEN is_official = true THEN 1 END) AS official_codes,
        COUNT(CASE WHEN is_active = true THEN 1 END) AS active_codes
    FROM code_metadata
)
SELECT 
    'Overall Statistics' AS metric,
    total_codes AS "Total Codes",
    code_types AS "Code Types",
    codes_with_crosswalk AS "With Crosswalk",
    codes_with_hierarchy AS "With Hierarchy",
    codes_with_industry_mapping AS "With Industry Mapping",
    official_codes AS "Official Codes",
    active_codes AS "Active Codes",
    ROUND(codes_with_crosswalk * 100.0 / NULLIF(total_codes, 0), 2) AS "Crosswalk Coverage %",
    ROUND(codes_with_hierarchy * 100.0 / NULLIF(total_codes, 0), 2) AS "Hierarchy Coverage %",
    ROUND(codes_with_industry_mapping * 100.0 / NULLIF(total_codes, 0), 2) AS "Industry Mapping Coverage %"
FROM stats;

-- =====================================================
-- Part 10: View Verification
-- =====================================================

SELECT 
    '=== VIEW VERIFICATION ===' AS section;

-- Check code_crosswalk_view
SELECT 
    'code_crosswalk_view Records' AS metric,
    COUNT(*) AS count
FROM code_crosswalk_view;

-- Sample from code_crosswalk_view
-- Note: This view expands arrays into rows, so each row represents one crosswalk link
SELECT 
    'Sample from code_crosswalk_view' AS metric,
    code_type,
    code,
    official_name,
    naics_code,
    sic_code,
    mcc_code
FROM code_crosswalk_view
LIMIT 5;

-- Check code_hierarchy_view
SELECT 
    'code_hierarchy_view Records' AS metric,
    COUNT(*) AS count
FROM code_hierarchy_view;

-- Sample from code_hierarchy_view
-- Note: This view expands child_codes into rows, so each row represents one parent-child relationship
SELECT 
    'Sample from code_hierarchy_view' AS metric,
    code_type,
    code,
    official_name,
    parent_code,
    parent_type,
    child_code
FROM code_hierarchy_view
LIMIT 5;

-- =====================================================
-- Part 11: Quick Health Check
-- =====================================================

SELECT 
    '=== QUICK HEALTH CHECK ===' AS section;

SELECT 
    CASE 
        WHEN (SELECT COUNT(*) FROM code_metadata) >= 100 THEN '✅ PASS: Sufficient codes loaded'
        ELSE '❌ FAIL: Insufficient codes loaded'
    END AS "Total Codes Check",
    
    CASE 
        WHEN (SELECT COUNT(*) FROM code_metadata WHERE code_type = 'NAICS') >= 30 THEN '✅ PASS: Sufficient NAICS codes'
        ELSE '❌ FAIL: Insufficient NAICS codes'
    END AS "NAICS Codes Check",
    
    CASE 
        WHEN (SELECT COUNT(*) FROM code_metadata WHERE code_type = 'SIC') >= 20 THEN '✅ PASS: Sufficient SIC codes'
        ELSE '❌ FAIL: Insufficient SIC codes'
    END AS "SIC Codes Check",
    
    CASE 
        WHEN (SELECT COUNT(*) FROM code_metadata WHERE code_type = 'MCC') >= 50 THEN '✅ PASS: Sufficient MCC codes'
        ELSE '❌ FAIL: Insufficient MCC codes'
    END AS "MCC Codes Check",
    
    CASE 
        WHEN (SELECT COUNT(*) FROM code_metadata WHERE crosswalk_data != '{}'::jsonb) >= 50 THEN '✅ PASS: Sufficient crosswalk data'
        ELSE '❌ FAIL: Insufficient crosswalk data'
    END AS "Crosswalk Data Check",
    
    CASE 
        WHEN (SELECT COUNT(*) FROM code_metadata WHERE industry_mappings != '{}'::jsonb) >= 50 THEN '✅ PASS: Sufficient industry mappings'
        ELSE '❌ FAIL: Insufficient industry mappings'
    END AS "Industry Mappings Check",
    
    CASE 
        WHEN (SELECT COUNT(*) FROM code_metadata WHERE code_type = 'code_type' AND code = 'code') = 0 THEN '✅ PASS: No duplicate codes'
        ELSE '❌ FAIL: Duplicate codes found'
    END AS "Duplicate Check";

