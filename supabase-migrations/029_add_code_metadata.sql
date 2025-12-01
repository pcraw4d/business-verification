-- =====================================================
-- Migration: Add Code Metadata Table
-- Purpose: Store additional code information (descriptions, mappings, crosswalks, hierarchies)
-- Date: 2025-01-XX
-- OPTIMIZATION #6.2: Database Schema Enhancements
-- =====================================================

-- Step 1: Create code_metadata table
CREATE TABLE IF NOT EXISTS code_metadata (
    id BIGSERIAL PRIMARY KEY,
    code_type VARCHAR(10) NOT NULL CHECK (code_type IN ('MCC', 'SIC', 'NAICS')),
    code VARCHAR(20) NOT NULL,
    
    -- Official information
    official_description TEXT,
    official_name VARCHAR(255),
    official_category VARCHAR(100),
    
    -- Industry mappings (JSONB for flexibility)
    industry_mappings JSONB DEFAULT '{}'::jsonb,
    -- Example: {"primary_industry": "Technology", "secondary_industries": ["Software", "IT Services"]}
    
    -- Crosswalk data (links to other code types)
    crosswalk_data JSONB DEFAULT '{}'::jsonb,
    -- Example: {"naics": ["541511", "541512"], "sic": ["7371", "7372"], "mcc": ["5734", "5735"]}
    
    -- Code hierarchy (parent/child relationships)
    hierarchy JSONB DEFAULT '{}'::jsonb,
    -- Example: {"parent_code": "54", "parent_type": "NAICS", "child_codes": ["541511", "541512"]}
    
    -- Additional metadata
    metadata JSONB DEFAULT '{}'::jsonb,
    -- Example: {"source": "official", "last_updated": "2024-01-01", "notes": "..."}
    
    -- Status and timestamps
    is_active BOOLEAN DEFAULT true,
    is_official BOOLEAN DEFAULT false, -- Whether this is from an official source
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Unique constraint: one metadata record per code
    UNIQUE(code_type, code)
);

-- Step 2: Create indexes for efficient queries

-- Primary lookup index (most common query)
CREATE INDEX IF NOT EXISTS idx_code_metadata_type_code 
    ON code_metadata(code_type, code);

-- Index for filtering by is_active
CREATE INDEX IF NOT EXISTS idx_code_metadata_active 
    ON code_metadata(is_active) 
    WHERE is_active = true;

-- Index for filtering by is_official
CREATE INDEX IF NOT EXISTS idx_code_metadata_official 
    ON code_metadata(is_official) 
    WHERE is_official = true;

-- GIN index for JSONB queries on industry_mappings
CREATE INDEX IF NOT EXISTS idx_code_metadata_industry_mappings 
    ON code_metadata USING gin(industry_mappings);

-- GIN index for JSONB queries on crosswalk_data
CREATE INDEX IF NOT EXISTS idx_code_metadata_crosswalk_data 
    ON code_metadata USING gin(crosswalk_data);

-- GIN index for JSONB queries on hierarchy
CREATE INDEX IF NOT EXISTS idx_code_metadata_hierarchy 
    ON code_metadata USING gin(hierarchy);

-- GIN index for JSONB queries on metadata
CREATE INDEX IF NOT EXISTS idx_code_metadata_metadata 
    ON code_metadata USING gin(metadata);

-- Full-text search index on official_description
CREATE INDEX IF NOT EXISTS idx_code_metadata_description_fts 
    ON code_metadata USING gin(to_tsvector('english', COALESCE(official_description, '')));

-- Trigram index for fuzzy matching on official_name
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX IF NOT EXISTS idx_code_metadata_name_trgm 
    ON code_metadata USING gin(official_name gin_trgm_ops);

-- Index on updated_at for time-based queries
CREATE INDEX IF NOT EXISTS idx_code_metadata_updated_at 
    ON code_metadata(updated_at DESC);

-- Index on created_at for time-based queries
CREATE INDEX IF NOT EXISTS idx_code_metadata_created_at 
    ON code_metadata(created_at DESC);

-- Step 3: Create function to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_code_metadata_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Step 4: Create trigger to update updated_at on changes
CREATE TRIGGER update_code_metadata_updated_at
    BEFORE UPDATE ON code_metadata
    FOR EACH ROW
    EXECUTE FUNCTION update_code_metadata_updated_at();

-- Step 5: Create view for easy crosswalk queries
CREATE OR REPLACE VIEW code_crosswalk_view AS
SELECT 
    cm.code_type,
    cm.code,
    cm.official_description,
    cm.official_name,
    -- Extract NAICS codes from crosswalk
    jsonb_array_elements_text(
        COALESCE(cm.crosswalk_data->'naics', '[]'::jsonb)
    ) AS naics_code,
    -- Extract SIC codes from crosswalk
    jsonb_array_elements_text(
        COALESCE(cm.crosswalk_data->'sic', '[]'::jsonb)
    ) AS sic_code,
    -- Extract MCC codes from crosswalk
    jsonb_array_elements_text(
        COALESCE(cm.crosswalk_data->'mcc', '[]'::jsonb)
    ) AS mcc_code
FROM code_metadata cm
WHERE cm.is_active = true;

-- Step 6: Create view for code hierarchy
CREATE OR REPLACE VIEW code_hierarchy_view AS
SELECT 
    cm.code_type,
    cm.code,
    cm.official_name,
    cm.hierarchy->>'parent_code' AS parent_code,
    cm.hierarchy->>'parent_type' AS parent_type,
    jsonb_array_elements_text(
        COALESCE(cm.hierarchy->'child_codes', '[]'::jsonb)
    ) AS child_code
FROM code_metadata cm
WHERE cm.is_active = true 
AND cm.hierarchy != '{}'::jsonb;

-- Step 7: Add comments for documentation
COMMENT ON TABLE code_metadata IS 
    'Stores additional metadata for classification codes including official descriptions, industry mappings, crosswalk data, and hierarchies';

COMMENT ON COLUMN code_metadata.code_type IS 
    'Type of code: MCC, SIC, or NAICS';

COMMENT ON COLUMN code_metadata.code IS 
    'The actual code value (e.g., "541511" for NAICS)';

COMMENT ON COLUMN code_metadata.official_description IS 
    'Official description from the code standard (e.g., from Census Bureau for NAICS)';

COMMENT ON COLUMN code_metadata.official_name IS 
    'Official name/title of the code';

COMMENT ON COLUMN code_metadata.industry_mappings IS 
    'JSONB object mapping this code to industries (primary and secondary)';

COMMENT ON COLUMN code_metadata.crosswalk_data IS 
    'JSONB object containing equivalent codes in other systems (NAICS, SIC, MCC)';

COMMENT ON COLUMN code_metadata.hierarchy IS 
    'JSONB object containing parent/child relationships for hierarchical codes';

COMMENT ON COLUMN code_metadata.metadata IS 
    'Additional metadata such as source, last_updated, notes, etc.';

COMMENT ON COLUMN code_metadata.is_official IS 
    'Whether this metadata comes from an official source (e.g., Census Bureau, IRS)';

COMMENT ON INDEX idx_code_metadata_type_code IS 
    'Primary lookup index for code_type and code';

COMMENT ON INDEX idx_code_metadata_industry_mappings IS 
    'GIN index for efficient JSONB queries on industry mappings';

COMMENT ON INDEX idx_code_metadata_crosswalk_data IS 
    'GIN index for efficient JSONB queries on crosswalk data';

COMMENT ON INDEX idx_code_metadata_hierarchy IS 
    'GIN index for efficient JSONB queries on code hierarchy';

COMMENT ON VIEW code_crosswalk_view IS 
    'View for easy querying of code crosswalks (NAICS ↔ SIC ↔ MCC)';

COMMENT ON VIEW code_hierarchy_view IS 
    'View for easy querying of code hierarchies (parent/child relationships)';

-- =====================================================
-- Sample data insertion (for testing/documentation)
-- =====================================================

-- Example: NAICS code 541511 (Custom Computer Programming Services)
-- INSERT INTO code_metadata (
--     code_type,
--     code,
--     official_description,
--     official_name,
--     industry_mappings,
--     crosswalk_data,
--     hierarchy,
--     is_official
-- ) VALUES (
--     'NAICS',
--     '541511',
--     'This U.S. industry comprises establishments primarily engaged in writing, modifying, testing, and supporting software to meet the needs of a particular customer.',
--     'Custom Computer Programming Services',
--     '{"primary_industry": "Technology", "secondary_industries": ["Software", "IT Services"]}'::jsonb,
--     '{"sic": ["7371"], "mcc": ["5734"]}'::jsonb,
--     '{"parent_code": "5415", "parent_type": "NAICS", "child_codes": []}'::jsonb,
--     true
-- );

-- =====================================================
-- Verification queries (for manual checking)
-- =====================================================

-- To verify table exists:
-- SELECT table_name FROM information_schema.tables WHERE table_name = 'code_metadata';

-- To verify indexes:
-- SELECT indexname, indexdef FROM pg_indexes WHERE tablename = 'code_metadata';

-- To query crosswalk data:
-- SELECT * FROM code_crosswalk_view WHERE code_type = 'NAICS' AND code = '541511';

-- To query hierarchy:
-- SELECT * FROM code_hierarchy_view WHERE code_type = 'NAICS' AND code = '541511';

