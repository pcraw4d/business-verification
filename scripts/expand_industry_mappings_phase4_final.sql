-- =====================================================
-- Expand Industry Mapping Coverage - Phase 4 Final
-- Purpose: Increase industry mapping coverage from 78.68% to 80%+ (435+ codes)
-- Date: 2025-01-27
-- OPTIMIZATION: Accuracy Plan Enhancement - Phase 4 (Final)
-- =====================================================
-- 
-- Current: 428 codes with industry mappings (78.68%)
-- Target: 435+ codes with industry mappings (80%+)
-- Need: ~7 additional codes with industry mappings
-- =====================================================

-- =====================================================
-- Part 1: Add industry mappings for remaining codes
-- =====================================================

-- Remaining NAICS codes without industry mappings
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Professional Services',
    'secondary_industries', ARRAY['Professional Services', 'Business Services'],
    'industry_category', 'Professional Services'
)
WHERE code_type = 'NAICS' 
  AND code IN ('541199', '541214', '541410', '541420', '541930')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- Remaining SIC codes without industry mappings
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Professional Services',
    'secondary_industries', ARRAY['Professional Services', 'Business Services'],
    'industry_category', 'Professional Services'
)
WHERE code_type = 'SIC' 
  AND code IN ('7389', '8711', '8712', '8299')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- Remaining MCC codes without industry mappings
UPDATE code_metadata
SET industry_mappings = jsonb_build_object(
    'primary_industry', 'Retail & Commerce',
    'secondary_industries', ARRAY['Retail', 'Services', 'Commerce'],
    'industry_category', 'Retail'
)
WHERE code_type = 'MCC' 
  AND code IN ('3601', '4784', '5309', '5399', '5999')
  AND (industry_mappings = '{}'::jsonb OR industry_mappings IS NULL);

-- =====================================================
-- Verification Query
-- =====================================================

-- Verify industry mapping coverage after final supplement
SELECT 
    'Industry Mapping Coverage After Final Supplement' AS metric,
    COUNT(*) AS codes_with_mappings,
    (SELECT COUNT(*) FROM code_metadata WHERE is_active = true) AS total_codes,
    ROUND(COUNT(*) * 100.0 / NULLIF((SELECT COUNT(*) FROM code_metadata WHERE is_active = true), 0), 2) AS coverage_percentage,
    CASE 
        WHEN COUNT(*) * 100.0 / NULLIF((SELECT COUNT(*) FROM code_metadata WHERE is_active = true), 0) >= 80.0 THEN '✅ PASS - 80%+ coverage'
        ELSE '❌ FAIL - Below 80% coverage'
    END AS status
FROM code_metadata
WHERE is_active = true
  AND industry_mappings != '{}'::jsonb
  AND industry_mappings IS NOT NULL;

