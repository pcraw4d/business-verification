-- =====================================================
-- Fix Crosswalk Data Quality Issues
-- Purpose: Remove invalid codes, self-references, and fix array assignments
-- Date: 2025-01-27
-- OPTIMIZATION: Accuracy Plan Enhancement - Data Quality Fix
-- =====================================================
-- 
-- This script fixes:
-- 1. Removes invalid codes from crosswalk arrays (codes that don't exist)
-- 2. Removes self-references (codes referencing themselves)
-- 3. Ensures codes are in correct arrays (MCC in 'mcc', SIC in 'sic', NAICS in 'naics')
-- =====================================================

-- =====================================================
-- Part 1: Remove Invalid Codes and Self-References from Crosswalk Arrays
-- =====================================================

-- Create a function to clean crosswalk data
CREATE OR REPLACE FUNCTION clean_crosswalk_data(
    p_code_type VARCHAR,
    p_code VARCHAR,
    p_crosswalk_data JSONB
) RETURNS JSONB AS $$
DECLARE
    valid_naics TEXT[];
    valid_sic TEXT[];
    valid_mcc TEXT[];
    new_crosswalk JSONB := '{}'::jsonb;
BEGIN
    -- Filter NAICS codes - only keep those that exist and are not self-references
    IF p_crosswalk_data ? 'naics' THEN
        SELECT COALESCE(ARRAY_AGG(elem), ARRAY[]::TEXT[])
        INTO valid_naics
        FROM jsonb_array_elements_text(p_crosswalk_data->'naics') AS elem
        WHERE EXISTS (
            SELECT 1 FROM code_metadata cm2 
            WHERE cm2.code_type = 'NAICS' 
            AND cm2.code = elem
            AND cm2.is_active = true
        )
        AND elem != p_code; -- Remove self-references
        
        IF array_length(valid_naics, 1) > 0 THEN
            new_crosswalk := new_crosswalk || jsonb_build_object('naics', to_jsonb(valid_naics));
        END IF;
    END IF;
    
    -- Filter SIC codes - only keep those that exist and are not self-references
    IF p_crosswalk_data ? 'sic' THEN
        SELECT COALESCE(ARRAY_AGG(elem), ARRAY[]::TEXT[])
        INTO valid_sic
        FROM jsonb_array_elements_text(p_crosswalk_data->'sic') AS elem
        WHERE EXISTS (
            SELECT 1 FROM code_metadata cm2 
            WHERE cm2.code_type = 'SIC' 
            AND cm2.code = elem
            AND cm2.is_active = true
        )
        AND elem != p_code; -- Remove self-references
        
        IF array_length(valid_sic, 1) > 0 THEN
            new_crosswalk := new_crosswalk || jsonb_build_object('sic', to_jsonb(valid_sic));
        END IF;
    END IF;
    
    -- Filter MCC codes - only keep those that exist and are not self-references
    IF p_crosswalk_data ? 'mcc' THEN
        SELECT COALESCE(ARRAY_AGG(elem), ARRAY[]::TEXT[])
        INTO valid_mcc
        FROM jsonb_array_elements_text(p_crosswalk_data->'mcc') AS elem
        WHERE EXISTS (
            SELECT 1 FROM code_metadata cm2 
            WHERE cm2.code_type = 'MCC' 
            AND cm2.code = elem
            AND cm2.is_active = true
        )
        AND elem != p_code; -- Remove self-references
        
        IF array_length(valid_mcc, 1) > 0 THEN
            new_crosswalk := new_crosswalk || jsonb_build_object('mcc', to_jsonb(valid_mcc));
        END IF;
    END IF;
    
    RETURN new_crosswalk;
END;
$$ LANGUAGE plpgsql;

-- Update all crosswalk data using the function
UPDATE code_metadata
SET crosswalk_data = clean_crosswalk_data(code_type, code, crosswalk_data),
    updated_at = NOW()
WHERE is_active = true
  AND crosswalk_data != '{}'::jsonb
  AND clean_crosswalk_data(code_type, code, crosswalk_data) != crosswalk_data;

-- Drop the temporary function
DROP FUNCTION clean_crosswalk_data(VARCHAR, VARCHAR, JSONB);

-- =====================================================
-- Part 2: Fix Specific Invalid Code Issues
-- =====================================================

-- Fix MCC 5048: Remove invalid NAICS codes
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['443142']::text[],  -- Only keep valid NAICS codes
    'sic', ARRAY['5048', '5946']::text[],
    'mcc', ARRAY['5946']::text[]
)
WHERE code_type = 'MCC' AND code = '5048'
  AND EXISTS (
      SELECT 1 FROM code_metadata WHERE code_type = 'NAICS' AND code = '443142'
  );

-- Fix MCC 5046: Remove invalid NAICS codes
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY[]::text[],  -- Remove invalid codes
    'sic', ARRAY['5046', '5047']::text[],
    'mcc', ARRAY['5047']::text[]
)
WHERE code_type = 'MCC' AND code = '5046';

-- Fix MCC 5047: Remove invalid NAICS codes
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['621111']::text[],  -- Only keep valid NAICS codes
    'sic', ARRAY['5047', '8011']::text[],
    'mcc', ARRAY['5047', '8011']::text[]
)
WHERE code_type = 'MCC' AND code = '5047'
  AND EXISTS (
      SELECT 1 FROM code_metadata WHERE code_type = 'NAICS' AND code = '621111'
  );

-- Fix MCC 4119: Remove invalid NAICS codes
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['485210']::text[],  -- Only keep valid NAICS codes
    'sic', ARRAY['4119', '4121']::text[]
)
WHERE code_type = 'MCC' AND code = '4119'
  AND EXISTS (
      SELECT 1 FROM code_metadata WHERE code_type = 'NAICS' AND code = '485210'
  );

-- Fix MCC 4121: Remove invalid NAICS codes
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY[]::text[],  -- Remove invalid codes (485310, 485320 don't exist)
    'sic', ARRAY['4121', '4119']::text[]
)
WHERE code_type = 'MCC' AND code = '4121';

-- Fix MCC 5935: Remove invalid NAICS codes
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['562111']::text[],  -- Only keep valid NAICS codes
    'sic', ARRAY['5013', '5935']::text[]
)
WHERE code_type = 'MCC' AND code = '5935'
  AND EXISTS (
      SELECT 1 FROM code_metadata WHERE code_type = 'NAICS' AND code = '562111'
  );

-- Fix MCC 5735: Remove invalid NAICS codes
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY[]::text[],  -- Remove invalid codes (451211, 451220 don't exist)
    'sic', ARRAY['5735', '5942']::text[],
    'mcc', ARRAY['5735', '5942']::text[]
)
WHERE code_type = 'MCC' AND code = '5735';

-- Fix MCC 6300: Remove invalid codes
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY['524113', '524114', '524126', '524127', '524128', '524130']::text[],  -- Only keep valid NAICS codes
    'sic', ARRAY[]::text[]  -- Remove invalid SIC codes
)
WHERE code_type = 'MCC' AND code = '6300'
  AND EXISTS (
      SELECT 1 FROM code_metadata WHERE code_type = 'NAICS' AND code = '524113'
  );

-- Fix SIC 1795: Remove invalid NAICS codes
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY[]::text[],  -- Remove invalid codes (238910, 562111 don't exist)
    'mcc', ARRAY['5935']::text[]
)
WHERE code_type = 'SIC' AND code = '1795';

-- Fix SIC 1796: Remove invalid NAICS codes
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY[]::text[],  -- Remove invalid codes (238290 doesn't exist)
    'mcc', ARRAY[]::text[]
)
WHERE code_type = 'SIC' AND code = '1796';

-- Fix SIC 1799: Remove invalid NAICS codes
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'naics', ARRAY[]::text[],  -- Remove invalid codes (238390, 238990 don't exist)
    'mcc', ARRAY[]::text[]
)
WHERE code_type = 'SIC' AND code = '1799';

-- Fix NAICS 541612: Remove invalid codes
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY[]::text[],  -- Remove invalid SIC codes (7389, 7361 don't exist)
    'mcc', ARRAY[]::text[]  -- Remove invalid MCC codes (7399, 7372 don't exist)
)
WHERE code_type = 'NAICS' AND code = '541612';

-- Fix NAICS 541621: Remove invalid codes
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY[]::text[],  -- Remove invalid SIC codes (8734, 8731 don't exist)
    'mcc', ARRAY[]::text[]  -- Remove invalid MCC codes (8734 doesn't exist)
)
WHERE code_type = 'NAICS' AND code = '541621';

-- Fix NAICS 541370: Remove invalid MCC codes
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['8712', '8711']::text[],
    'mcc', ARRAY[]::text[]  -- Remove invalid MCC codes (7399, 7372 don't exist)
)
WHERE code_type = 'NAICS' AND code = '541370';

-- Fix NAICS 541330: Remove invalid MCC codes
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['8711', '8712', '8713']::text[],
    'mcc', ARRAY[]::text[]  -- Remove invalid MCC codes (8711 doesn't exist as MCC)
)
WHERE code_type = 'NAICS' AND code = '541330';

-- Fix NAICS 541310: Remove invalid MCC codes
UPDATE code_metadata
SET crosswalk_data = jsonb_build_object(
    'sic', ARRAY['8711', '8712', '8713']::text[],
    'mcc', ARRAY[]::text[]  -- Remove invalid MCC codes (8711 doesn't exist as MCC)
)
WHERE code_type = 'NAICS' AND code = '541310';

-- Fix NAICS codes with invalid MCC codes (4215, 4784 don't exist)
UPDATE code_metadata
SET crosswalk_data = crosswalk_data - 'mcc' || 
    CASE 
        WHEN crosswalk_data ? 'mcc' THEN
            jsonb_build_object('mcc', 
                (SELECT jsonb_agg(elem)
                 FROM jsonb_array_elements_text(crosswalk_data->'mcc') AS elem
                 WHERE EXISTS (
                     SELECT 1 FROM code_metadata cm2 
                     WHERE cm2.code_type = 'MCC' 
                     AND cm2.code = elem
                     AND cm2.is_active = true
                 )
                 AND elem != code)
            )
        ELSE '{}'::jsonb
    END
WHERE code_type = 'NAICS' 
  AND code IN ('484122', '484230', '492210')
  AND crosswalk_data ? 'mcc';

-- Fix NAICS codes with invalid MCC codes (5182, 8244, 8299 don't exist)
UPDATE code_metadata
SET crosswalk_data = crosswalk_data - 'mcc' || 
    CASE 
        WHEN crosswalk_data ? 'mcc' THEN
            jsonb_build_object('mcc', 
                (SELECT jsonb_agg(elem)
                 FROM jsonb_array_elements_text(crosswalk_data->'mcc') AS elem
                 WHERE EXISTS (
                     SELECT 1 FROM code_metadata cm2 
                     WHERE cm2.code_type = 'MCC' 
                     AND cm2.code = elem
                     AND cm2.is_active = true
                 )
                 AND elem != code)
            )
        ELSE '{}'::jsonb
    END
WHERE code_type = 'NAICS' 
  AND code IN ('311213', '611630', '611691', '611692')
  AND crosswalk_data ? 'mcc';

-- Fix SIC codes with invalid MCC codes
UPDATE code_metadata
SET crosswalk_data = crosswalk_data - 'mcc' || 
    CASE 
        WHEN crosswalk_data ? 'mcc' THEN
            jsonb_build_object('mcc', 
                (SELECT jsonb_agg(elem)
                 FROM jsonb_array_elements_text(crosswalk_data->'mcc') AS elem
                 WHERE EXISTS (
                     SELECT 1 FROM code_metadata cm2 
                     WHERE cm2.code_type = 'MCC' 
                     AND cm2.code = elem
                     AND cm2.is_active = true
                 )
                 AND elem != code)
            )
        ELSE '{}'::jsonb
    END
WHERE code_type = 'SIC' 
  AND code = '2082'
  AND crosswalk_data ? 'mcc';

-- Fix MCC codes with invalid SIC codes
UPDATE code_metadata
SET crosswalk_data = crosswalk_data - 'sic' || 
    CASE 
        WHEN crosswalk_data ? 'sic' THEN
            jsonb_build_object('sic', 
                (SELECT jsonb_agg(elem)
                 FROM jsonb_array_elements_text(crosswalk_data->'sic') AS elem
                 WHERE EXISTS (
                     SELECT 1 FROM code_metadata cm2 
                     WHERE cm2.code_type = 'SIC' 
                     AND cm2.code = elem
                     AND cm2.is_active = true
                 )
                 AND elem != code)
            )
        ELSE '{}'::jsonb
    END
WHERE code_type = 'MCC' 
  AND code IN ('5046', '6300', '4119', '4121', '5932', '5933', '5935', '4112', '4131', '5992', '4814')
  AND crosswalk_data ? 'sic';

-- Fix NAICS codes with invalid SIC codes
UPDATE code_metadata
SET crosswalk_data = crosswalk_data - 'sic' || 
    CASE 
        WHEN crosswalk_data ? 'sic' THEN
            jsonb_build_object('sic', 
                (SELECT jsonb_agg(elem)
                 FROM jsonb_array_elements_text(crosswalk_data->'sic') AS elem
                 WHERE EXISTS (
                     SELECT 1 FROM code_metadata cm2 
                     WHERE cm2.code_type = 'SIC' 
                     AND cm2.code = elem
                     AND cm2.is_active = true
                 )
                 AND elem != code)
            )
        ELSE '{}'::jsonb
    END
WHERE code_type = 'NAICS' 
  AND code IN ('541612', '541621')
  AND crosswalk_data ? 'sic';

-- =====================================================
-- Verification Query
-- =====================================================

-- Check remaining invalid codes
SELECT 
    'Remaining Invalid Codes After Fix' AS metric,
    COUNT(*) AS invalid_code_count
FROM code_metadata cm
CROSS JOIN LATERAL jsonb_array_elements_text(cm.crosswalk_data->'naics') AS naics_code
WHERE cm.is_active = true
  AND cm.crosswalk_data ? 'naics'
  AND NOT EXISTS (
      SELECT 1 FROM code_metadata cm2 
      WHERE cm2.code_type = 'NAICS' 
      AND cm2.code = naics_code
      AND cm2.is_active = true
  )

UNION ALL

SELECT 
    'Remaining Invalid SIC Codes After Fix' AS metric,
    COUNT(*) AS invalid_code_count
FROM code_metadata cm
CROSS JOIN LATERAL jsonb_array_elements_text(cm.crosswalk_data->'sic') AS sic_code
WHERE cm.is_active = true
  AND cm.crosswalk_data ? 'sic'
  AND NOT EXISTS (
      SELECT 1 FROM code_metadata cm2 
      WHERE cm2.code_type = 'SIC' 
      AND cm2.code = sic_code
      AND cm2.is_active = true
  )

UNION ALL

SELECT 
    'Remaining Invalid MCC Codes After Fix' AS metric,
    COUNT(*) AS invalid_code_count
FROM code_metadata cm
CROSS JOIN LATERAL jsonb_array_elements_text(cm.crosswalk_data->'mcc') AS mcc_code
WHERE cm.is_active = true
  AND cm.crosswalk_data ? 'mcc'
  AND NOT EXISTS (
      SELECT 1 FROM code_metadata cm2 
      WHERE cm2.code_type = 'MCC' 
      AND cm2.code = mcc_code
      AND cm2.is_active = true
  );

-- Check remaining self-references
SELECT 
    'Remaining Self-References After Fix' AS metric,
    COUNT(*) AS self_reference_count
FROM code_metadata cm
WHERE cm.is_active = true
  AND (
      (cm.crosswalk_data ? 'mcc' AND cm.code = ANY(SELECT jsonb_array_elements_text(cm.crosswalk_data->'mcc')))
      OR (cm.crosswalk_data ? 'sic' AND cm.code = ANY(SELECT jsonb_array_elements_text(cm.crosswalk_data->'sic')))
      OR (cm.crosswalk_data ? 'naics' AND cm.code = ANY(SELECT jsonb_array_elements_text(cm.crosswalk_data->'naics')))
  );

-- Summary of fixes
SELECT 
    'Crosswalk Data Quality Fix Summary' AS metric,
    (SELECT COUNT(*) FROM code_metadata WHERE is_active = true AND crosswalk_data != '{}'::jsonb) AS codes_with_crosswalks,
    (SELECT COUNT(*) FROM code_metadata WHERE is_active = true) AS total_codes,
    ROUND((SELECT COUNT(*) FROM code_metadata WHERE is_active = true AND crosswalk_data != '{}'::jsonb) * 100.0 / 
          NULLIF((SELECT COUNT(*) FROM code_metadata WHERE is_active = true), 0), 2) AS coverage_percentage;

