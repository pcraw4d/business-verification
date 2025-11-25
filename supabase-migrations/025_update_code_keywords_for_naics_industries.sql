-- Migration: Update code_keywords for NAICS-Aligned Industries
-- This migration re-runs the keyword population for newly classified codes
-- to ensure all codes have appropriate keyword mappings

-- =====================================================
-- Step 1: Remove existing code_keywords for reclassified codes
-- =====================================================

-- This ensures we get fresh keyword mappings based on new industry assignments
-- We'll keep keywords for codes that haven't changed industries

-- Optional: Clear keywords for codes that were reclassified
-- Uncomment if you want to start fresh for reclassified codes
/*
DELETE FROM code_keywords
WHERE code_id IN (
    SELECT cc.id 
    FROM classification_codes cc
    INNER JOIN industries i ON i.id = cc.industry_id
    WHERE i.name IN (
        'Agriculture, Forestry, Fishing and Hunting',
        'Mining, Quarrying, and Oil and Gas Extraction',
        'Utilities',
        'Wholesale Trade',
        'Real Estate and Rental and Leasing',
        'Professional, Scientific, and Technical Services',
        'Management of Companies and Enterprises',
        'Administrative and Support Services',
        'Arts, Entertainment, and Recreation',
        'Other Services',
        'Public Administration'
    )
);
*/

-- =====================================================
-- Step 2: Add keywords from newly classified code descriptions
-- =====================================================

-- Extract keywords from code descriptions for newly classified codes
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT
    cc.id as code_id,
    trim(lower(keyword)) as keyword,
    0.85 as relevance_score,
    'exact' as match_type
FROM classification_codes cc,
LATERAL unnest(
    string_to_array(
        regexp_replace(
            lower(cc.description),
            '[^a-z0-9\s]',
            ' ',
            'g'
        ),
        ' '
    )
) as keyword
WHERE trim(keyword) != ''
  AND length(trim(keyword)) >= 3
  AND trim(keyword) NOT IN (
    'the', 'a', 'an', 'and', 'or', 'but', 'in', 'on', 'at', 'to', 'for',
    'of', 'with', 'by', 'from', 'as', 'is', 'was', 'are', 'were', 'be',
    'been', 'being', 'have', 'has', 'had', 'do', 'does', 'did', 'will',
    'would', 'should', 'could', 'may', 'might', 'must', 'can', 'this',
    'that', 'these', 'those', 'it', 'its', 'they', 'them', 'their', 'our',
    'your', 'my', 'his', 'her', 'he', 'she', 'we', 'you', 'i', 'me', 'us',
    'services', 'service', 'products', 'product', 'business', 'company',
    'companies', 'other', 'etc', 'including', 'excluding', 'related', 'not'
  )
  AND trim(keyword) !~ '^\d+$'
  AND cc.industry_id IS NOT NULL
  AND cc.industry_id IN (
    SELECT id FROM industries WHERE name IN (
        'Agriculture, Forestry, Fishing and Hunting',
        'Mining, Quarrying, and Oil and Gas Extraction',
        'Utilities',
        'Wholesale Trade',
        'Real Estate and Rental and Leasing',
        'Professional, Scientific, and Technical Services',
        'Management of Companies and Enterprises',
        'Administrative and Support Services',
        'Arts, Entertainment, and Recreation',
        'Other Services',
        'Public Administration'
    )
  )
ON CONFLICT (code_id, keyword) DO NOTHING;

-- =====================================================
-- Step 3: Link industry keywords to newly classified codes
-- =====================================================

-- Link industry keywords to codes in the same industry
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT
    cc.id as code_id,
    ik.keyword,
    LEAST(0.95, (ik.weight * 0.8 + 0.15)) as relevance_score,
    CASE 
        WHEN ik.weight >= 0.90 THEN 'exact' 
        ELSE 'partial' 
    END as match_type
FROM classification_codes cc
INNER JOIN industry_keywords ik ON ik.industry_id = cc.industry_id
WHERE cc.industry_id IS NOT NULL
  AND lower(cc.description) LIKE '%' || lower(ik.keyword) || '%'
  AND cc.industry_id IN (
    SELECT id FROM industries WHERE name IN (
        'Agriculture, Forestry, Fishing and Hunting',
        'Mining, Quarrying, and Oil and Gas Extraction',
        'Utilities',
        'Wholesale Trade',
        'Real Estate and Rental and Leasing',
        'Professional, Scientific, and Technical Services',
        'Management of Companies and Enterprises',
        'Administrative and Support Services',
        'Arts, Entertainment, and Recreation',
        'Other Services',
        'Public Administration'
    )
  )
ON CONFLICT (code_id, keyword) 
DO UPDATE SET
    relevance_score = GREATEST(code_keywords.relevance_score, EXCLUDED.relevance_score),
    match_type = CASE WHEN EXCLUDED.match_type = 'exact' THEN 'exact' ELSE code_keywords.match_type END;

-- =====================================================
-- Step 4: Add NAICS-specific keywords for new industries
-- =====================================================

-- Agriculture, Forestry, Fishing and Hunting
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.90, 'exact'
FROM classification_codes cc
INNER JOIN industries i ON cc.industry_id = i.id
CROSS JOIN (VALUES
    ('agriculture'), ('farming'), ('farm'), ('farmer'), ('crop'), ('livestock'),
    ('forestry'), ('forest'), ('logging'), ('timber'), ('fishing'), ('fishery'),
    ('aquaculture'), ('hunting'), ('trapping'), ('harvest'), ('cultivation')
) AS keywords(keyword)
WHERE cc.industry_id IS NOT NULL
  AND i.name = 'Agriculture, Forestry, Fishing and Hunting'
  AND lower(cc.description) LIKE '%' || keywords.keyword || '%'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- Mining
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.90, 'exact'
FROM classification_codes cc
INNER JOIN industries i ON cc.industry_id = i.id
CROSS JOIN (VALUES
    ('mining'), ('extraction'), ('quarry'), ('quarrying'), ('drilling'),
    ('oil'), ('gas'), ('petroleum'), ('coal'), ('metals'), ('mineral'), ('ore')
) AS keywords(keyword)
WHERE i.name = 'Mining, Quarrying, and Oil and Gas Extraction'
  AND lower(cc.description) LIKE '%' || keywords.keyword || '%'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- Utilities
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.90, 'exact'
FROM classification_codes cc
INNER JOIN industries i ON cc.industry_id = i.id
CROSS JOIN (VALUES
    ('utilities'), ('utility'), ('electric'), ('electricity'), ('power'),
    ('natural gas'), ('water'), ('sewage'), ('wastewater'), ('sanitation'),
    ('grid'), ('transmission'), ('distribution')
) AS keywords(keyword)
WHERE i.name = 'Utilities'
  AND lower(cc.description) LIKE '%' || keywords.keyword || '%'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- Wholesale Trade
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.90, 'exact'
FROM classification_codes cc
INNER JOIN industries i ON cc.industry_id = i.id
CROSS JOIN (VALUES
    ('wholesale'), ('wholesaler'), ('distributor'), ('distribution'),
    ('b2b'), ('business to business'), ('trade'), ('import'), ('export'),
    ('supplier'), ('vendor'), ('bulk')
) AS keywords(keyword)
WHERE i.name = 'Wholesale Trade'
  AND lower(cc.description) LIKE '%' || keywords.keyword || '%'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- Real Estate and Rental and Leasing
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.90, 'exact'
FROM classification_codes cc
INNER JOIN industries i ON cc.industry_id = i.id
CROSS JOIN (VALUES
    ('real estate'), ('realty'), ('property'), ('properties'), ('realtor'),
    ('broker'), ('agent'), ('property management'), ('landlord'), ('leasing'),
    ('rental'), ('commercial real estate'), ('residential real estate'),
    ('appraisal'), ('valuation')
) AS keywords(keyword)
WHERE i.name = 'Real Estate and Rental and Leasing'
  AND lower(cc.description) LIKE '%' || keywords.keyword || '%'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- Professional, Scientific, and Technical Services
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.90, 'exact'
FROM classification_codes cc
INNER JOIN industries i ON cc.industry_id = i.id
CROSS JOIN (VALUES
    ('professional services'), ('consulting'), ('consultant'), ('legal'),
    ('attorney'), ('lawyer'), ('law firm'), ('accounting'), ('accountant'),
    ('cpa'), ('audit'), ('tax'), ('engineering'), ('engineer'), ('architecture'),
    ('architect'), ('scientific'), ('research'), ('technical services')
) AS keywords(keyword)
WHERE i.name = 'Professional, Scientific, and Technical Services'
  AND lower(cc.description) LIKE '%' || keywords.keyword || '%'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- Management of Companies and Enterprises
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.90, 'exact'
FROM classification_codes cc
INNER JOIN industries i ON cc.industry_id = i.id
CROSS JOIN (VALUES
    ('holding company'), ('management company'), ('corporate management'),
    ('enterprise management'), ('corporate'), ('enterprise')
) AS keywords(keyword)
WHERE i.name = 'Management of Companies and Enterprises'
  AND lower(cc.description) LIKE '%' || keywords.keyword || '%'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- Administrative and Support Services
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.90, 'exact'
FROM classification_codes cc
INNER JOIN industries i ON cc.industry_id = i.id
CROSS JOIN (VALUES
    ('administrative services'), ('administrative'), ('administration'),
    ('support services'), ('facilities management'), ('office management'),
    ('employment services'), ('staffing'), ('temp'), ('temporary'),
    ('call center'), ('customer service'), ('waste management'), ('cleaning'),
    ('janitorial')
) AS keywords(keyword)
WHERE i.name = 'Administrative and Support Services'
  AND lower(cc.description) LIKE '%' || keywords.keyword || '%'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- Arts, Entertainment, and Recreation
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.90, 'exact'
FROM classification_codes cc
INNER JOIN industries i ON cc.industry_id = i.id
CROSS JOIN (VALUES
    ('arts'), ('art'), ('entertainment'), ('recreation'), ('recreational'),
    ('gallery'), ('museum'), ('theater'), ('theatre'), ('sports'), ('fitness'),
    ('gym'), ('music'), ('musician'), ('film'), ('movie'), ('cinema'),
    ('amusement park'), ('gaming')
) AS keywords(keyword)
WHERE i.name = 'Arts, Entertainment, and Recreation'
  AND lower(cc.description) LIKE '%' || keywords.keyword || '%'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- Other Services
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.90, 'exact'
FROM classification_codes cc
INNER JOIN industries i ON cc.industry_id = i.id
CROSS JOIN (VALUES
    ('repair'), ('maintenance'), ('personal services'), ('laundry'),
    ('dry cleaning'), ('funeral'), ('pet services'), ('religious'),
    ('church'), ('ministry'), ('grantmaking')
) AS keywords(keyword)
WHERE i.name = 'Other Services'
  AND lower(cc.description) LIKE '%' || keywords.keyword || '%'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- Public Administration
INSERT INTO code_keywords (code_id, keyword, relevance_score, match_type)
SELECT DISTINCT cc.id, keyword, 0.90, 'exact'
FROM classification_codes cc
INNER JOIN industries i ON cc.industry_id = i.id
CROSS JOIN (VALUES
    ('public administration'), ('government'), ('federal'), ('state'),
    ('local'), ('municipal'), ('regulatory'), ('public works'),
    ('infrastructure'), ('public service')
) AS keywords(keyword)
WHERE i.name = 'Public Administration'
  AND lower(cc.description) LIKE '%' || keywords.keyword || '%'
ON CONFLICT (code_id, keyword) DO NOTHING;

-- =====================================================
-- Step 5: Display Summary
-- =====================================================

SELECT 
    'Code Keywords Updated for NAICS Industries' as status,
    COUNT(*) as total_keywords,
    COUNT(DISTINCT code_id) as codes_with_keywords,
    ROUND(COUNT(*)::NUMERIC / NULLIF(COUNT(DISTINCT code_id), 0), 2) as avg_keywords_per_code
FROM code_keywords
WHERE code_id IN (
    SELECT cc.id 
    FROM classification_codes cc
    INNER JOIN industries i ON i.id = cc.industry_id
    WHERE i.name IN (
        'Agriculture, Forestry, Fishing and Hunting',
        'Mining, Quarrying, and Oil and Gas Extraction',
        'Utilities',
        'Wholesale Trade',
        'Real Estate and Rental and Leasing',
        'Professional, Scientific, and Technical Services',
        'Management of Companies and Enterprises',
        'Administrative and Support Services',
        'Arts, Entertainment, and Recreation',
        'Other Services',
        'Public Administration'
    )
);

-- Breakdown by new industry
SELECT 
    i.name as industry,
    COUNT(DISTINCT ck.code_id) as codes_with_keywords,
    COUNT(ck.id) as total_keywords,
    ROUND(AVG(ck.relevance_score), 2) as avg_relevance
FROM code_keywords ck
INNER JOIN classification_codes cc ON cc.id = ck.code_id
INNER JOIN industries i ON i.id = cc.industry_id
WHERE i.name IN (
    'Agriculture, Forestry, Fishing and Hunting',
    'Mining, Quarrying, and Oil and Gas Extraction',
    'Utilities',
    'Wholesale Trade',
    'Real Estate and Rental and Leasing',
    'Professional, Scientific, and Technical Services',
    'Management of Companies and Enterprises',
    'Administrative and Support Services',
    'Arts, Entertainment, and Recreation',
    'Other Services',
    'Public Administration'
)
GROUP BY i.name
ORDER BY i.name;

