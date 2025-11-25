-- Migration: Add NAICS-Aligned Industries
-- This migration adds all 20 NAICS 2-digit sectors as industries
-- and aligns the classification system with official NAICS structure

-- =====================================================
-- Step 1: Add Missing NAICS-Aligned Industries
-- =====================================================

-- Note: Some industries may already exist with different names
-- We'll add the missing ones and can consolidate later

INSERT INTO industries (name, description, category, confidence_threshold, is_active, created_at, updated_at) VALUES

-- NAICS 11: Agriculture, Forestry, Fishing and Hunting
('Agriculture, Forestry, Fishing and Hunting', 
 'NAICS 11: Crop production, animal production, forestry, fishing, hunting, and trapping', 
 'traditional', 0.80, true, NOW(), NOW()),

-- NAICS 21: Mining, Quarrying, and Oil and Gas Extraction
('Mining, Quarrying, and Oil and Gas Extraction', 
 'NAICS 21: Mining operations, oil and gas extraction, quarrying, and related activities', 
 'traditional', 0.80, true, NOW(), NOW()),

-- NAICS 22: Utilities (may already exist as "Energy")
('Utilities', 
 'NAICS 22: Electric power generation, transmission and distribution, natural gas distribution, water and sewage systems', 
 'traditional', 0.80, true, NOW(), NOW()),

-- NAICS 42: Wholesale Trade
('Wholesale Trade', 
 'NAICS 42: Business-to-business wholesale distribution of goods', 
 'traditional', 0.80, true, NOW(), NOW()),

-- NAICS 53: Real Estate and Rental and Leasing
('Real Estate and Rental and Leasing', 
 'NAICS 53: Real estate, rental, and leasing services including property management', 
 'traditional', 0.80, true, NOW(), NOW()),

-- NAICS 54: Professional, Scientific, and Technical Services
('Professional, Scientific, and Technical Services', 
 'NAICS 54: Legal services, accounting, consulting, engineering, architecture, scientific research, and technical services', 
 'traditional', 0.80, true, NOW(), NOW()),

-- NAICS 55: Management of Companies and Enterprises
('Management of Companies and Enterprises', 
 'NAICS 55: Holding companies and corporate management services', 
 'traditional', 0.80, true, NOW(), NOW()),

-- NAICS 56: Administrative and Support Services
('Administrative and Support Services', 
 'NAICS 56: Administrative support, facilities management, employment services, waste management, and remediation services', 
 'traditional', 0.75, true, NOW(), NOW()),

-- NAICS 71: Arts, Entertainment, and Recreation
('Arts, Entertainment, and Recreation', 
 'NAICS 71: Performing arts, sports, museums, historical sites, amusement parks, and recreation services', 
 'traditional', 0.75, true, NOW(), NOW()),

-- NAICS 81: Other Services (except Public Administration)
('Other Services', 
 'NAICS 81: Repair and maintenance, personal services, religious organizations, grantmaking, and other services', 
 'traditional', 0.75, true, NOW(), NOW()),

-- NAICS 92: Public Administration
('Public Administration', 
 'NAICS 92: Government services, public administration, regulatory services, and public works', 
 'traditional', 0.80, true, NOW(), NOW())

ON CONFLICT (name) DO NOTHING;

-- =====================================================
-- Step 2: Update Existing Industries to Match NAICS Names
-- =====================================================

-- Update existing industries to use official NAICS sector names where applicable
-- This helps with consistency and alignment

UPDATE industries 
SET description = 'NAICS 23: Building construction, heavy construction, and specialty trade contractors'
WHERE name = 'Construction';

UPDATE industries 
SET description = 'NAICS 31-33: Manufacturing of goods across all industries'
WHERE name = 'Manufacturing';

UPDATE industries 
SET description = 'NAICS 44-45: Retail stores and establishments selling goods directly to consumers'
WHERE name = 'Retail';

UPDATE industries 
SET description = 'NAICS 48-49: Transportation services, warehousing, and logistics'
WHERE name = 'Transportation';

UPDATE industries 
SET description = 'NAICS 61: Schools, universities, training programs, and educational support services'
WHERE name = 'Education';

UPDATE industries 
SET description = 'NAICS 62: Hospitals, medical practices, dental offices, and social assistance services'
WHERE name = 'Healthcare';

UPDATE industries 
SET description = 'NAICS 52: Banking, credit, insurance, investment activities, and financial services'
WHERE name = 'Finance';

UPDATE industries 
SET description = 'NAICS 72: Hotels, restaurants, food services, and accommodation'
WHERE name = 'Food & Beverage';

UPDATE industries 
SET description = 'NAICS 51: Publishing, broadcasting, telecommunications, data processing, and information services'
WHERE name = 'Technology';

-- =====================================================
-- Step 3: Add Comprehensive Keywords for Each NAICS Sector
-- =====================================================

-- NAICS 11: Agriculture, Forestry, Fishing and Hunting
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active, created_at, updated_at)
SELECT id, keyword, weight, true, NOW(), NOW()
FROM industries,
(VALUES
    ('agriculture', 0.95),
    ('farming', 0.95),
    ('farm', 0.90),
    ('farmer', 0.90),
    ('crop', 0.85),
    ('livestock', 0.85),
    ('forestry', 0.90),
    ('forest', 0.85),
    ('logging', 0.85),
    ('timber', 0.85),
    ('fishing', 0.90),
    ('fishery', 0.85),
    ('aquaculture', 0.85),
    ('hunting', 0.85),
    ('trapping', 0.80),
    ('harvest', 0.80),
    ('cultivation', 0.80)
) AS keywords(keyword, weight)
WHERE industries.name = 'Agriculture, Forestry, Fishing and Hunting'
ON CONFLICT (industry_id, keyword) DO NOTHING;

-- NAICS 21: Mining
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active, created_at, updated_at)
SELECT id, keyword, weight, true, NOW(), NOW()
FROM industries,
(VALUES
    ('mining', 0.95),
    ('extraction', 0.90),
    ('quarry', 0.85),
    ('quarrying', 0.85),
    ('drilling', 0.85),
    ('oil', 0.90),
    ('gas', 0.90),
    ('petroleum', 0.90),
    ('coal', 0.85),
    ('metals', 0.85),
    ('mineral', 0.85),
    ('ore', 0.80)
) AS keywords(keyword, weight)
WHERE industries.name = 'Mining, Quarrying, and Oil and Gas Extraction'
ON CONFLICT (industry_id, keyword) DO NOTHING;

-- NAICS 22: Utilities
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active, created_at, updated_at)
SELECT id, keyword, weight, true, NOW(), NOW()
FROM industries,
(VALUES
    ('utilities', 0.95),
    ('utility', 0.95),
    ('electric', 0.90),
    ('electricity', 0.90),
    ('power', 0.90),
    ('natural gas', 0.85),
    ('water', 0.85),
    ('sewage', 0.85),
    ('wastewater', 0.85),
    ('sanitation', 0.80),
    ('grid', 0.80),
    ('transmission', 0.80),
    ('distribution', 0.80)
) AS keywords(keyword, weight)
WHERE industries.name = 'Utilities'
ON CONFLICT (industry_id, keyword) DO NOTHING;

-- NAICS 42: Wholesale Trade
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active, created_at, updated_at)
SELECT id, keyword, weight, true, NOW(), NOW()
FROM industries,
(VALUES
    ('wholesale', 0.95),
    ('wholesaler', 0.95),
    ('distributor', 0.90),
    ('distribution', 0.90),
    ('b2b', 0.85),
    ('business to business', 0.85),
    ('trade', 0.80),
    ('import', 0.80),
    ('export', 0.80),
    ('supplier', 0.80),
    ('vendor', 0.75),
    ('bulk', 0.75)
) AS keywords(keyword, weight)
WHERE industries.name = 'Wholesale Trade'
ON CONFLICT (industry_id, keyword) DO NOTHING;

-- NAICS 53: Real Estate and Rental and Leasing
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active, created_at, updated_at)
SELECT id, keyword, weight, true, NOW(), NOW()
FROM industries,
(VALUES
    ('real estate', 0.95),
    ('realty', 0.90),
    ('property', 0.90),
    ('properties', 0.90),
    ('realtor', 0.85),
    ('broker', 0.85),
    ('agent', 0.85),
    ('property management', 0.85),
    ('landlord', 0.80),
    ('leasing', 0.90),
    ('rental', 0.90),
    ('commercial real estate', 0.85),
    ('residential real estate', 0.85),
    ('appraisal', 0.80),
    ('valuation', 0.80)
) AS keywords(keyword, weight)
WHERE industries.name = 'Real Estate and Rental and Leasing'
ON CONFLICT (industry_id, keyword) DO NOTHING;

-- NAICS 54: Professional, Scientific, and Technical Services
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active, created_at, updated_at)
SELECT id, keyword, weight, true, NOW(), NOW()
FROM industries,
(VALUES
    ('professional services', 0.95),
    ('consulting', 0.95),
    ('consultant', 0.90),
    ('legal', 0.90),
    ('attorney', 0.90),
    ('lawyer', 0.90),
    ('law firm', 0.85),
    ('accounting', 0.90),
    ('accountant', 0.90),
    ('cpa', 0.85),
    ('audit', 0.85),
    ('tax', 0.85),
    ('engineering', 0.90),
    ('engineer', 0.85),
    ('architecture', 0.85),
    ('architect', 0.85),
    ('scientific', 0.85),
    ('research', 0.85),
    ('technical services', 0.90)
) AS keywords(keyword, weight)
WHERE industries.name = 'Professional, Scientific, and Technical Services'
ON CONFLICT (industry_id, keyword) DO NOTHING;

-- NAICS 55: Management of Companies and Enterprises
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active, created_at, updated_at)
SELECT id, keyword, weight, true, NOW(), NOW()
FROM industries,
(VALUES
    ('holding company', 0.95),
    ('management company', 0.90),
    ('corporate management', 0.90),
    ('enterprise management', 0.85),
    ('corporate', 0.80),
    ('enterprise', 0.80)
) AS keywords(keyword, weight)
WHERE industries.name = 'Management of Companies and Enterprises'
ON CONFLICT (industry_id, keyword) DO NOTHING;

-- NAICS 56: Administrative and Support Services
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active, created_at, updated_at)
SELECT id, keyword, weight, true, NOW(), NOW()
FROM industries,
(VALUES
    ('administrative services', 0.95),
    ('administrative', 0.95),
    ('administration', 0.90),
    ('support services', 0.90),
    ('facilities management', 0.85),
    ('office management', 0.85),
    ('employment services', 0.85),
    ('staffing', 0.85),
    ('temp', 0.80),
    ('temporary', 0.80),
    ('call center', 0.80),
    ('customer service', 0.80),
    ('waste management', 0.85),
    ('cleaning', 0.80),
    ('janitorial', 0.80)
) AS keywords(keyword, weight)
WHERE industries.name = 'Administrative and Support Services'
ON CONFLICT (industry_id, keyword) DO NOTHING;

-- NAICS 71: Arts, Entertainment, and Recreation
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active, created_at, updated_at)
SELECT id, keyword, weight, true, NOW(), NOW()
FROM industries,
(VALUES
    ('arts', 0.95),
    ('art', 0.95),
    ('entertainment', 0.95),
    ('recreation', 0.95),
    ('recreational', 0.90),
    ('gallery', 0.85),
    ('museum', 0.85),
    ('theater', 0.85),
    ('theatre', 0.85),
    ('sports', 0.85),
    ('fitness', 0.85),
    ('gym', 0.80),
    ('music', 0.85),
    ('musician', 0.80),
    ('film', 0.85),
    ('movie', 0.85),
    ('cinema', 0.80),
    ('amusement park', 0.80),
    ('gaming', 0.80)
) AS keywords(keyword, weight)
WHERE industries.name = 'Arts, Entertainment, and Recreation'
ON CONFLICT (industry_id, keyword) DO NOTHING;

-- NAICS 81: Other Services
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active, created_at, updated_at)
SELECT id, keyword, weight, true, NOW(), NOW()
FROM industries,
(VALUES
    ('repair', 0.90),
    ('maintenance', 0.90),
    ('personal services', 0.85),
    ('laundry', 0.80),
    ('dry cleaning', 0.80),
    ('funeral', 0.80),
    ('pet services', 0.80),
    ('religious', 0.85),
    ('church', 0.85),
    ('ministry', 0.80),
    ('grantmaking', 0.80)
) AS keywords(keyword, weight)
WHERE industries.name = 'Other Services'
ON CONFLICT (industry_id, keyword) DO NOTHING;

-- NAICS 92: Public Administration
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active, created_at, updated_at)
SELECT id, keyword, weight, true, NOW(), NOW()
FROM industries,
(VALUES
    ('public administration', 0.95),
    ('government', 0.95),
    ('federal', 0.90),
    ('state', 0.90),
    ('local', 0.90),
    ('municipal', 0.90),
    ('regulatory', 0.85),
    ('public works', 0.85),
    ('infrastructure', 0.80),
    ('public service', 0.85)
) AS keywords(keyword, weight)
WHERE industries.name = 'Public Administration'
ON CONFLICT (industry_id, keyword) DO NOTHING;

-- =====================================================
-- Step 4: Display Summary
-- =====================================================

SELECT 
    'NAICS-Aligned Industries Added' as status,
    COUNT(*) as total_industries,
    COUNT(CASE WHEN category = 'traditional' THEN 1 END) as traditional_industries,
    COUNT(CASE WHEN category = 'emerging' THEN 1 END) as emerging_industries
FROM industries
WHERE is_active = true;

SELECT 
    i.name as industry,
    COUNT(ik.id) as keyword_count
FROM industries i
LEFT JOIN industry_keywords ik ON ik.industry_id = i.id
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

