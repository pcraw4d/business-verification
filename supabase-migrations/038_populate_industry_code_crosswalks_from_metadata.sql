-- Migration: Populate industry_code_crosswalks from code_metadata
-- Created: 2025-12-22
-- Purpose: Migrate crosswalk data from code_metadata.crosswalk_data (JSONB) to industry_code_crosswalks (structured table)
-- This enables the hybrid crosswalk approach for better performance

-- =====================================================
-- PART 1: Populate industry_code_crosswalks from code_metadata
-- =====================================================

-- Extract crosswalk data from code_metadata and insert into industry_code_crosswalks
-- This creates structured crosswalk entries from the JSONB data
-- Note: industry_code_crosswalks uses source_system/source_code -> target_system/target_code format

INSERT INTO industry_code_crosswalks (
    source_system,
    source_code,
    target_system,
    target_code,
    confidence_score,
    mapping_type,
    is_active
)
SELECT DISTINCT
    cm.code_type::varchar(20) as source_system,
    cm.code::varchar(20) as source_code,
    target_type::varchar(20) as target_system,
    target_code::varchar(20) as target_code,
    0.80::numeric(3,2) as confidence_score,  -- Default confidence
    'direct'::varchar(20) as mapping_type,
    true as is_active
FROM code_metadata cm
CROSS JOIN LATERAL (
    -- Extract NAICS codes
    SELECT 'NAICS' as target_type, jsonb_array_elements_text(cm.crosswalk_data->'naics') as target_code
    WHERE cm.crosswalk_data->'naics' IS NOT NULL
    AND jsonb_typeof(cm.crosswalk_data->'naics') = 'array'
    AND jsonb_array_length(cm.crosswalk_data->'naics') > 0
    
    UNION ALL
    
    -- Extract SIC codes
    SELECT 'SIC' as target_type, jsonb_array_elements_text(cm.crosswalk_data->'sic') as target_code
    WHERE cm.crosswalk_data->'sic' IS NOT NULL
    AND jsonb_typeof(cm.crosswalk_data->'sic') = 'array'
    AND jsonb_array_length(cm.crosswalk_data->'sic') > 0
    
    UNION ALL
    
    -- Extract MCC codes
    SELECT 'MCC' as target_type, jsonb_array_elements_text(cm.crosswalk_data->'mcc') as target_code
    WHERE cm.crosswalk_data->'mcc' IS NOT NULL
    AND jsonb_typeof(cm.crosswalk_data->'mcc') = 'array'
    AND jsonb_array_length(cm.crosswalk_data->'mcc') > 0
) extracted
WHERE cm.is_active = true
  AND cm.crosswalk_data IS NOT NULL
  AND cm.crosswalk_data != '{}'::jsonb
  AND target_code IS NOT NULL
  AND target_code != ''
  AND target_code != 'null'  -- Filter out null strings
  -- Avoid duplicates (using unique constraint columns)
  AND NOT EXISTS (
      SELECT 1 FROM industry_code_crosswalks icc
      WHERE icc.source_system = cm.code_type::varchar(20)
        AND icc.source_code = cm.code::varchar(20)
        AND icc.target_system = target_type::varchar(20)
        AND icc.target_code = target_code::varchar(20)
  )
ON CONFLICT (source_system, source_code, target_system, target_code) DO NOTHING;

-- =====================================================
-- PART 2: Create reverse mappings (bidirectional crosswalks)
-- =====================================================

-- Create reverse mappings: if MCC -> NAICS exists, also create NAICS -> MCC
INSERT INTO industry_code_crosswalks (
    source_system,
    source_code,
    target_system,
    target_code,
    confidence_score,
    mapping_type,
    is_active
)
SELECT DISTINCT
    target_system as source_system,
    target_code as source_code,
    source_system as target_system,
    source_code as target_code,
    confidence_score,
    mapping_type,
    is_active
FROM industry_code_crosswalks
WHERE is_active = true
  -- Avoid duplicates
  AND NOT EXISTS (
      SELECT 1 FROM industry_code_crosswalks icc
      WHERE icc.source_system = industry_code_crosswalks.target_system
        AND icc.source_code = industry_code_crosswalks.target_code
        AND icc.target_system = industry_code_crosswalks.source_system
        AND icc.target_code = industry_code_crosswalks.source_code
  )
ON CONFLICT (source_system, source_code, target_system, target_code) DO NOTHING;

-- =====================================================
-- PART 3: Verification and Statistics
-- =====================================================

DO $$
DECLARE
    total_crosswalks int;
    mcc_to_naics int;
    mcc_to_sic int;
    naics_to_sic int;
    naics_to_mcc int;
    sic_to_naics int;
    sic_to_mcc int;
BEGIN
    RAISE NOTICE '========================================';
    RAISE NOTICE 'CROSSWALK MIGRATION SUMMARY';
    RAISE NOTICE '========================================';
    
    SELECT COUNT(*) INTO total_crosswalks
    FROM industry_code_crosswalks
    WHERE is_active = true;
    
    SELECT COUNT(*) INTO mcc_to_naics
    FROM industry_code_crosswalks
    WHERE source_system = 'MCC' AND target_system = 'NAICS' AND is_active = true;
    
    SELECT COUNT(*) INTO mcc_to_sic
    FROM industry_code_crosswalks
    WHERE source_system = 'MCC' AND target_system = 'SIC' AND is_active = true;
    
    SELECT COUNT(*) INTO naics_to_sic
    FROM industry_code_crosswalks
    WHERE source_system = 'NAICS' AND target_system = 'SIC' AND is_active = true;
    
    SELECT COUNT(*) INTO naics_to_mcc
    FROM industry_code_crosswalks
    WHERE source_system = 'NAICS' AND target_system = 'MCC' AND is_active = true;
    
    SELECT COUNT(*) INTO sic_to_naics
    FROM industry_code_crosswalks
    WHERE source_system = 'SIC' AND target_system = 'NAICS' AND is_active = true;
    
    SELECT COUNT(*) INTO sic_to_mcc
    FROM industry_code_crosswalks
    WHERE source_system = 'SIC' AND target_system = 'MCC' AND is_active = true;
    
    RAISE NOTICE 'Total crosswalks: %', total_crosswalks;
    RAISE NOTICE '';
    RAISE NOTICE 'MCC -> NAICS: %', mcc_to_naics;
    RAISE NOTICE 'MCC -> SIC: %', mcc_to_sic;
    RAISE NOTICE 'NAICS -> SIC: %', naics_to_sic;
    RAISE NOTICE 'NAICS -> MCC: %', naics_to_mcc;
    RAISE NOTICE 'SIC -> NAICS: %', sic_to_naics;
    RAISE NOTICE 'SIC -> MCC: %', sic_to_mcc;
    RAISE NOTICE '';
    RAISE NOTICE '✅ Crosswalk migration complete';
END $$;

COMMENT ON TABLE industry_code_crosswalks IS 
    'Structured crosswalk mappings between classification code systems (MCC, SIC, NAICS).
     Populated from code_metadata.crosswalk_data for better query performance.
     Supports bidirectional mappings for flexible code lookups.';

