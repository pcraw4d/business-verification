-- =====================================================
-- Migration: Populate Technology Keywords
-- Purpose: Add missing technology keywords (cloud, computing, software) to industry_keywords table
-- Date: 2025-01-XX
-- Phase 2: Cloud Services Keyword Extraction Enhancement
-- =====================================================

-- This migration populates technology-related keywords that are missing from the
-- industry_keywords table, specifically for "cloud", "computing", "software", etc.

-- =====================================================
-- Part 1: Find or Create Technology Industry
-- =====================================================

-- Ensure "Technology" industry exists (ID may vary, so we'll use a subquery)
DO $$
DECLARE
    tech_industry_id INTEGER;
BEGIN
    -- Find Technology industry
    SELECT id INTO tech_industry_id
    FROM industries
    WHERE LOWER(name) IN ('technology', 'software', 'it services', 'information technology')
    LIMIT 1;
    
    -- If not found, create it
    IF tech_industry_id IS NULL THEN
        INSERT INTO industries (name, description, category, is_active)
        VALUES ('Technology', 'Technology and software services industry', 'emerging', true)
        RETURNING id INTO tech_industry_id;
    END IF;
    
    -- =====================================================
    -- Part 2: Populate Technology Keywords
    -- =====================================================
    
    -- Cloud computing keywords
    -- Note: industry_keywords table schema: id, industry_id, keyword, weight, is_active, created_at, updated_at
    -- No context or is_primary columns in this schema
    INSERT INTO industry_keywords (industry_id, keyword, weight, is_active)
    VALUES
    (tech_industry_id, 'cloud', 0.95, true),
    (tech_industry_id, 'cloud computing', 0.95, true),
    (tech_industry_id, 'cloud services', 0.95, true),
    (tech_industry_id, 'computing', 0.90, true),
    (tech_industry_id, 'software', 0.90, true),
    (tech_industry_id, 'saas', 0.90, true),
    (tech_industry_id, 'platform', 0.85, true),
    (tech_industry_id, 'technology', 0.85, true),
    (tech_industry_id, 'tech', 0.80, true),
    (tech_industry_id, 'it services', 0.85, true),
    (tech_industry_id, 'information technology', 0.85, true),
    (tech_industry_id, 'solutions', 0.75, true),
    (tech_industry_id, 'services', 0.70, true)
    ON CONFLICT (industry_id, keyword) DO UPDATE SET
        weight = GREATEST(industry_keywords.weight, EXCLUDED.weight),
        is_active = true,
        updated_at = NOW();
END $$;

-- =====================================================
-- Part 3: Populate Additional Technology Keywords for Other Industries
-- =====================================================

-- Add cloud keywords to "Software" industry if it exists separately
DO $$
DECLARE
    software_industry_id INTEGER;
BEGIN
    SELECT id INTO software_industry_id
    FROM industries
    WHERE LOWER(name) IN ('software', 'software development', 'software services')
    LIMIT 1;
    
    IF software_industry_id IS NOT NULL THEN
        INSERT INTO industry_keywords (industry_id, keyword, weight, is_active)
        VALUES
        (software_industry_id, 'cloud', 0.90, true),
        (software_industry_id, 'cloud computing', 0.90, true),
        (software_industry_id, 'computing', 0.85, true),
        (software_industry_id, 'saas', 0.90, true),
        (software_industry_id, 'platform', 0.85, true)
        ON CONFLICT (industry_id, keyword) DO UPDATE SET
            weight = GREATEST(industry_keywords.weight, EXCLUDED.weight),
            is_active = true,
            updated_at = NOW();
    END IF;
END $$;

-- =====================================================
-- Part 4: Verification Queries
-- =====================================================

-- Verify keywords were populated
-- SELECT i.name, ik.keyword, ik.weight, ik.is_primary
-- FROM industry_keywords ik
-- JOIN industries i ON i.id = ik.industry_id
-- WHERE ik.keyword IN ('cloud', 'computing', 'software', 'saas', 'platform')
-- ORDER BY i.name, ik.weight DESC;
