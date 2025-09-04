-- =============================================================================
-- SUPABASE KEYWORD MANAGEMENT QUERIES
-- =============================================================================
-- This file contains practical SQL queries for managing keywords and 
-- classification codes using the Supabase table editor and SQL editor.
-- 
-- Usage: Copy and paste these queries into the Supabase SQL Editor
-- =============================================================================

-- =============================================================================
-- 1. DATA OVERVIEW AND ANALYSIS QUERIES
-- =============================================================================

-- View all industries with keyword counts
SELECT 
    i.id,
    i.name,
    i.description,
    i.category,
    COUNT(ik.id) as keyword_count,
    AVG(ik.weight) as avg_keyword_weight,
    i.is_active,
    i.created_at
FROM industries i
LEFT JOIN industry_keywords ik ON i.id = ik.industry_id
GROUP BY i.id, i.name, i.description, i.category, i.is_active, i.created_at
ORDER BY keyword_count DESC;

-- View all classification codes with their industries
SELECT 
    cc.id,
    cc.code,
    cc.description,
    cc.code_type,
    i.name as industry_name,
    COUNT(ck.id) as keyword_count,
    cc.is_active
FROM classification_codes cc
JOIN industries i ON cc.industry_id = i.id
LEFT JOIN code_keywords ck ON cc.id = ck.code_id
GROUP BY cc.id, cc.code, cc.description, cc.code_type, i.name, cc.is_active
ORDER BY cc.code_type, cc.code;

-- Find keywords used across multiple industries
SELECT 
    ik.keyword,
    COUNT(DISTINCT ik.industry_id) as industry_count,
    STRING_AGG(i.name, ', ') as industries,
    AVG(ik.weight) as avg_weight
FROM industry_keywords ik
JOIN industries i ON ik.industry_id = i.id
GROUP BY ik.keyword
HAVING COUNT(DISTINCT ik.industry_id) > 1
ORDER BY industry_count DESC, avg_weight DESC;

-- =============================================================================
-- 2. KEYWORD MANAGEMENT QUERIES
-- =============================================================================

-- Add new keywords for an industry (replace industry_id and keywords as needed)
INSERT INTO industry_keywords (industry_id, keyword, weight, keyword_type, is_active, created_at)
VALUES 
    (1, 'artificial intelligence', 0.9, 'primary', true, NOW()),
    (1, 'machine learning', 0.8, 'primary', true, NOW()),
    (1, 'deep learning', 0.7, 'secondary', true, NOW()),
    (1, 'neural networks', 0.6, 'secondary', true, NOW()),
    (1, 'data science', 0.8, 'primary', true, NOW());

-- Update keyword weights based on performance
UPDATE industry_keywords 
SET 
    weight = CASE 
        WHEN weight < 0.5 THEN weight * 1.2
        WHEN weight >= 0.5 AND weight < 0.8 THEN weight * 1.1
        ELSE weight
    END,
    updated_at = NOW()
WHERE keyword IN (
    SELECT DISTINCT unnest(keywords_used) as keyword
    FROM classification_history 
    WHERE confidence_score > 0.8
    AND created_at >= NOW() - INTERVAL '30 days'
);

-- Deactivate low-performing keywords
UPDATE industry_keywords 
SET 
    is_active = false,
    updated_at = NOW()
WHERE weight < 0.3
AND keyword NOT IN (
    SELECT DISTINCT unnest(keywords_used) as keyword
    FROM classification_history 
    WHERE confidence_score > 0.7
    AND created_at >= NOW() - INTERVAL '7 days'
);

-- Find and remove duplicate keywords within the same industry
WITH duplicates AS (
    SELECT 
        industry_id,
        keyword,
        MIN(id) as keep_id,
        ARRAY_AGG(id) as all_ids
    FROM industry_keywords
    GROUP BY industry_id, keyword
    HAVING COUNT(*) > 1
)
DELETE FROM industry_keywords 
WHERE id IN (
    SELECT unnest(all_ids[2:]) -- Keep first, delete rest
    FROM duplicates
);

-- =============================================================================
-- 3. CLASSIFICATION CODE MANAGEMENT QUERIES
-- =============================================================================

-- Add new NAICS codes for an industry
INSERT INTO classification_codes (code, description, code_type, industry_id, is_active, created_at)
VALUES 
    ('541511', 'Custom Computer Programming Services', 'NAICS', 1, true, NOW()),
    ('541512', 'Computer Systems Design Services', 'NAICS', 1, true, NOW()),
    ('541513', 'Computer Facilities Management Services', 'NAICS', 1, true, NOW()),
    ('541519', 'Other Computer Related Services', 'NAICS', 1, true, NOW());

-- Add new MCC codes for an industry
INSERT INTO classification_codes (code, description, code_type, industry_id, is_active, created_at)
VALUES 
    ('5734', 'Computer Software Stores', 'MCC', 1, true, NOW()),
    ('7372', 'Computer Programming Services', 'MCC', 1, true, NOW()),
    ('7373', 'Computer Integrated Systems Design', 'MCC', 1, true, NOW());

-- Add new SIC codes for an industry
INSERT INTO classification_codes (code, description, code_type, industry_id, is_active, created_at)
VALUES 
    ('7371', 'Computer Programming Services', 'SIC', 1, true, NOW()),
    ('7372', 'Prepackaged Software', 'SIC', 1, true, NOW()),
    ('7373', 'Computer Integrated Systems Design', 'SIC', 1, true, NOW());

-- Link keywords to classification codes
INSERT INTO code_keywords (code_id, keyword, weight, is_active, created_at)
SELECT 
    cc.id,
    ik.keyword,
    ik.weight * 0.8, -- Slightly lower weight for code keywords
    true,
    NOW()
FROM classification_codes cc
JOIN industries i ON cc.industry_id = i.id
JOIN industry_keywords ik ON i.id = ik.industry_id
WHERE cc.code_type = 'NAICS'
AND ik.is_active = true
AND ik.weight > 0.5;

-- =============================================================================
-- 4. DATA VALIDATION AND QUALITY CHECKS
-- =============================================================================

-- Check for orphaned keywords (keywords without valid industry)
SELECT 
    ik.id,
    ik.keyword,
    ik.industry_id
FROM industry_keywords ik
LEFT JOIN industries i ON ik.industry_id = i.id
WHERE i.id IS NULL;

-- Check for orphaned code keywords (code keywords without valid code)
SELECT 
    ck.id,
    ck.keyword,
    ck.code_id
FROM code_keywords ck
LEFT JOIN classification_codes cc ON ck.code_id = cc.id
WHERE cc.id IS NULL;

-- Check for invalid weights
SELECT 
    'industry_keywords' as table_name,
    id,
    keyword,
    weight
FROM industry_keywords 
WHERE weight < 0 OR weight > 1
UNION ALL
SELECT 
    'code_keywords' as table_name,
    id,
    keyword,
    weight
FROM code_keywords 
WHERE weight < 0 OR weight > 1;

-- Check for empty or null keywords
SELECT 
    'industry_keywords' as table_name,
    id,
    keyword
FROM industry_keywords 
WHERE keyword IS NULL OR TRIM(keyword) = ''
UNION ALL
SELECT 
    'code_keywords' as table_name,
    id,
    keyword
FROM code_keywords 
WHERE keyword IS NULL OR TRIM(keyword) = '';

-- Check for missing keyword mappings
SELECT 
    cc.code,
    cc.description,
    cc.code_type,
    i.name as industry_name
FROM classification_codes cc
JOIN industries i ON cc.industry_id = i.id
LEFT JOIN code_keywords ck ON cc.id = ck.code_id
WHERE ck.id IS NULL
ORDER BY cc.code_type, cc.code;

-- =============================================================================
-- 5. PERFORMANCE MONITORING QUERIES
-- =============================================================================

-- Monitor classification performance over time
SELECT 
    DATE(created_at) as date,
    COUNT(*) as classification_count,
    AVG(confidence_score) as avg_confidence,
    MIN(confidence_score) as min_confidence,
    MAX(confidence_score) as max_confidence
FROM classification_history
WHERE created_at >= NOW() - INTERVAL '7 days'
GROUP BY DATE(created_at)
ORDER BY date DESC;

-- Monitor keyword usage and effectiveness
SELECT 
    ik.keyword,
    i.name as industry_name,
    COUNT(ch.id) as usage_count,
    AVG(ch.confidence_score) as avg_confidence,
    MAX(ch.created_at) as last_used
FROM industry_keywords ik
JOIN industries i ON ik.industry_id = i.id
LEFT JOIN classification_history ch ON ch.keywords_used @> ARRAY[ik.keyword]
WHERE ch.created_at >= NOW() - INTERVAL '30 days' OR ch.created_at IS NULL
GROUP BY ik.keyword, i.name, ik.weight
ORDER BY usage_count DESC, avg_confidence DESC;

-- Monitor system metrics
SELECT 
    metric_name,
    AVG(metric_value) as avg_value,
    MAX(metric_value) as max_value,
    MIN(metric_value) as min_value,
    COUNT(*) as sample_count
FROM system_metrics
WHERE recorded_at >= NOW() - INTERVAL '1 day'
GROUP BY metric_name
ORDER BY avg_value DESC;

-- =============================================================================
-- 6. BULK OPERATIONS AND DATA CLEANUP
-- =============================================================================

-- Clean up inactive keywords older than 90 days
DELETE FROM industry_keywords 
WHERE is_active = false 
AND updated_at < NOW() - INTERVAL '90 days';

-- Clean up old classification history (keep last 6 months)
DELETE FROM classification_history 
WHERE created_at < NOW() - INTERVAL '6 months';

-- Clean up old system metrics (keep last 3 months)
DELETE FROM system_metrics 
WHERE recorded_at < NOW() - INTERVAL '3 months';

-- Update timestamps for recently modified records
UPDATE industries 
SET updated_at = NOW()
WHERE updated_at < created_at + INTERVAL '1 day';

UPDATE industry_keywords 
SET updated_at = NOW()
WHERE updated_at < created_at + INTERVAL '1 day';

UPDATE classification_codes 
SET updated_at = NOW()
WHERE updated_at < created_at + INTERVAL '1 day';

UPDATE code_keywords 
SET updated_at = NOW()
WHERE updated_at < created_at + INTERVAL '1 day';

-- =============================================================================
-- 7. REPORTING AND ANALYTICS QUERIES
-- =============================================================================

-- Generate keyword effectiveness report
SELECT 
    i.name as industry_name,
    COUNT(ik.id) as total_keywords,
    COUNT(CASE WHEN ik.is_active THEN 1 END) as active_keywords,
    AVG(ik.weight) as avg_weight,
    COUNT(DISTINCT ch.id) as total_classifications,
    AVG(ch.confidence_score) as avg_confidence
FROM industries i
LEFT JOIN industry_keywords ik ON i.id = ik.industry_id
LEFT JOIN classification_history ch ON ch.industry_detected = i.name
WHERE ch.created_at >= NOW() - INTERVAL '30 days' OR ch.created_at IS NULL
GROUP BY i.id, i.name
ORDER BY total_classifications DESC;

-- Generate classification code coverage report
SELECT 
    cc.code_type,
    COUNT(cc.id) as total_codes,
    COUNT(CASE WHEN cc.is_active THEN 1 END) as active_codes,
    COUNT(ck.id) as codes_with_keywords,
    ROUND(
        (COUNT(ck.id)::float / COUNT(cc.id)::float) * 100, 2
    ) as keyword_coverage_percent
FROM classification_codes cc
LEFT JOIN code_keywords ck ON cc.id = ck.code_id
GROUP BY cc.code_type
ORDER BY cc.code_type;

-- Generate system health report
SELECT 
    'Total Industries' as metric,
    COUNT(*)::text as value
FROM industries
UNION ALL
SELECT 
    'Active Industries' as metric,
    COUNT(*)::text as value
FROM industries WHERE is_active = true
UNION ALL
SELECT 
    'Total Keywords' as metric,
    COUNT(*)::text as value
FROM industry_keywords
UNION ALL
SELECT 
    'Active Keywords' as metric,
    COUNT(*)::text as value
FROM industry_keywords WHERE is_active = true
UNION ALL
SELECT 
    'Total Classification Codes' as metric,
    COUNT(*)::text as value
FROM classification_codes
UNION ALL
SELECT 
    'Active Classification Codes' as metric,
    COUNT(*)::text as value
FROM classification_codes WHERE is_active = true
UNION ALL
SELECT 
    'Total Classifications (24h)' as metric,
    COUNT(*)::text as value
FROM classification_history WHERE created_at >= NOW() - INTERVAL '24 hours'
UNION ALL
SELECT 
    'Average Confidence (24h)' as metric,
    ROUND(AVG(confidence_score), 3)::text as value
FROM classification_history WHERE created_at >= NOW() - INTERVAL '24 hours';

-- =============================================================================
-- 8. UTILITY FUNCTIONS
-- =============================================================================

-- Function to get industry ID by name
CREATE OR REPLACE FUNCTION get_industry_id(industry_name TEXT)
RETURNS INTEGER AS $$
BEGIN
    RETURN (SELECT id FROM industries WHERE name = industry_name LIMIT 1);
END;
$$ LANGUAGE plpgsql;

-- Function to get classification code ID by code and type
CREATE OR REPLACE FUNCTION get_code_id(code_value TEXT, code_type_value TEXT)
RETURNS INTEGER AS $$
BEGIN
    RETURN (SELECT id FROM classification_codes WHERE code = code_value AND code_type = code_type_value LIMIT 1);
END;
$$ LANGUAGE plpgsql;

-- Function to add keyword to industry (with duplicate check)
CREATE OR REPLACE FUNCTION add_industry_keyword(
    industry_name TEXT,
    keyword_text TEXT,
    keyword_weight FLOAT DEFAULT 0.5,
    keyword_type_value TEXT DEFAULT 'secondary'
)
RETURNS BOOLEAN AS $$
DECLARE
    industry_id_val INTEGER;
BEGIN
    -- Get industry ID
    industry_id_val := get_industry_id(industry_name);
    
    IF industry_id_val IS NULL THEN
        RAISE EXCEPTION 'Industry not found: %', industry_name;
    END IF;
    
    -- Check if keyword already exists
    IF EXISTS (
        SELECT 1 FROM industry_keywords 
        WHERE industry_id = industry_id_val AND keyword = keyword_text
    ) THEN
        RETURN FALSE; -- Keyword already exists
    END IF;
    
    -- Insert new keyword
    INSERT INTO industry_keywords (industry_id, keyword, weight, keyword_type, is_active, created_at)
    VALUES (industry_id_val, keyword_text, keyword_weight, keyword_type_value, true, NOW());
    
    RETURN TRUE; -- Keyword added successfully
END;
$$ LANGUAGE plpgsql;

-- =============================================================================
-- 9. EXAMPLE USAGE
-- =============================================================================

-- Example: Add keywords to Technology industry
SELECT add_industry_keyword('Technology', 'blockchain', 0.8, 'primary');
SELECT add_industry_keyword('Technology', 'cryptocurrency', 0.7, 'secondary');
SELECT add_industry_keyword('Technology', 'web3', 0.6, 'secondary');

-- Example: Get industry ID and add multiple keywords
DO $$
DECLARE
    tech_id INTEGER;
BEGIN
    tech_id := get_industry_id('Technology');
    
    IF tech_id IS NOT NULL THEN
        INSERT INTO industry_keywords (industry_id, keyword, weight, keyword_type, is_active, created_at)
        VALUES 
            (tech_id, 'cloud computing', 0.9, 'primary', true, NOW()),
            (tech_id, 'serverless', 0.7, 'secondary', true, NOW()),
            (tech_id, 'microservices', 0.8, 'primary', true, NOW());
    END IF;
END $$;

-- =============================================================================
-- END OF KEYWORD MANAGEMENT QUERIES
-- =============================================================================
