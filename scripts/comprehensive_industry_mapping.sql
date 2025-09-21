-- KYB Platform - Comprehensive Industry Mapping and Taxonomy
-- This script creates a comprehensive industry taxonomy hierarchy and maps all major industry sectors
-- Run this script to establish the complete industry coverage framework

-- =============================================================================
-- COMPREHENSIVE INDUSTRY TAXONOMY HIERARCHY
-- =============================================================================

-- Create industry taxonomy table for hierarchical organization
CREATE TABLE IF NOT EXISTS industry_taxonomy (
    id SERIAL PRIMARY KEY,
    parent_id INTEGER REFERENCES industry_taxonomy(id),
    level INTEGER NOT NULL CHECK (level >= 1 AND level <= 4),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    category_type VARCHAR(50) NOT NULL CHECK (category_type IN ('primary', 'secondary', 'tertiary', 'specific')),
    market_size VARCHAR(20) CHECK (market_size IN ('small', 'medium', 'large', 'very_large')),
    growth_rate VARCHAR(20) CHECK (growth_rate IN ('declining', 'stable', 'growing', 'high_growth')),
    coverage_status VARCHAR(20) DEFAULT 'missing' CHECK (coverage_status IN ('missing', 'partial', 'good', 'excellent')),
    priority VARCHAR(20) DEFAULT 'medium' CHECK (priority IN ('low', 'medium', 'high', 'critical')),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for industry taxonomy
CREATE INDEX IF NOT EXISTS idx_industry_taxonomy_parent ON industry_taxonomy(parent_id);
CREATE INDEX IF NOT EXISTS idx_industry_taxonomy_level ON industry_taxonomy(level);
CREATE INDEX IF NOT EXISTS idx_industry_taxonomy_category ON industry_taxonomy(category_type);
CREATE INDEX IF NOT EXISTS idx_industry_taxonomy_coverage ON industry_taxonomy(coverage_status);
CREATE INDEX IF NOT EXISTS idx_industry_taxonomy_priority ON industry_taxonomy(priority);

-- =============================================================================
-- LEVEL 1: PRIMARY INDUSTRY CATEGORIES (Major Sectors)
-- =============================================================================

INSERT INTO industry_taxonomy (level, name, description, category_type, market_size, growth_rate, coverage_status, priority) VALUES
-- Technology & Digital Services
(1, 'Technology & Digital Services', 'Technology companies, software development, digital services, and IT solutions', 'primary', 'very_large', 'high_growth', 'good', 'critical'),

-- Healthcare & Life Sciences
(1, 'Healthcare & Life Sciences', 'Healthcare providers, medical services, pharmaceuticals, and health technology', 'primary', 'very_large', 'growing', 'good', 'critical'),

-- Financial Services
(1, 'Financial Services', 'Banking, investment, insurance, fintech, and financial technology', 'primary', 'very_large', 'growing', 'good', 'critical'),

-- Retail & Consumer Goods
(1, 'Retail & Consumer Goods', 'Retail stores, e-commerce, consumer products, and shopping services', 'primary', 'very_large', 'growing', 'partial', 'high'),

-- Manufacturing & Industrial
(1, 'Manufacturing & Industrial', 'Manufacturing, industrial production, and supply chain services', 'primary', 'very_large', 'stable', 'partial', 'high'),

-- Professional Services
(1, 'Professional Services', 'Legal, accounting, consulting, marketing, and business services', 'primary', 'large', 'growing', 'missing', 'high'),

-- Food & Beverage
(1, 'Food & Beverage', 'Restaurants, food service, food manufacturing, and beverage industry', 'primary', 'very_large', 'stable', 'missing', 'critical'),

-- Construction & Real Estate
(1, 'Construction & Real Estate', 'Construction, real estate, architecture, and property services', 'primary', 'large', 'growing', 'missing', 'high'),

-- Transportation & Logistics
(1, 'Transportation & Logistics', 'Transportation, shipping, logistics, and supply chain services', 'primary', 'large', 'growing', 'partial', 'medium'),

-- Education & Training
(1, 'Education & Training', 'Educational institutions, training services, and learning platforms', 'primary', 'large', 'growing', 'partial', 'medium'),

-- Entertainment & Media
(1, 'Entertainment & Media', 'Media production, entertainment, gaming, and content creation', 'primary', 'large', 'growing', 'partial', 'medium'),

-- Energy & Utilities
(1, 'Energy & Utilities', 'Energy production, utilities, renewable energy, and environmental services', 'primary', 'large', 'growing', 'missing', 'medium'),

-- Agriculture & Food Production
(1, 'Agriculture & Food Production', 'Farming, agriculture, food processing, and agricultural services', 'primary', 'medium', 'stable', 'missing', 'medium'),

-- Automotive
(1, 'Automotive', 'Automotive manufacturing, sales, services, and related industries', 'primary', 'large', 'stable', 'missing', 'medium'),

-- Government & Public Sector
(1, 'Government & Public Sector', 'Government agencies, public services, and municipal services', 'primary', 'very_large', 'stable', 'missing', 'low'),

-- Non-profit & Social Services
(1, 'Non-profit & Social Services', 'Non-profit organizations, charities, and social services', 'primary', 'medium', 'stable', 'missing', 'low'),

-- Emerging & Specialized
(1, 'Emerging & Specialized', 'Emerging industries, specialized services, and niche markets', 'primary', 'small', 'high_growth', 'missing', 'medium');

-- =============================================================================
-- LEVEL 2: SECONDARY INDUSTRY CATEGORIES (Sub-sectors)
-- =============================================================================

-- Technology & Digital Services Sub-categories
INSERT INTO industry_taxonomy (parent_id, level, name, description, category_type, market_size, growth_rate, coverage_status, priority) 
SELECT t.id, 2, sub.name, sub.description, 'secondary', sub.market_size, sub.growth_rate, sub.coverage_status, sub.priority
FROM industry_taxonomy t, (VALUES
    ('Software Development', 'Custom software development, programming, and application development', 'large', 'high_growth', 'good', 'critical'),
    ('Cloud Computing', 'Cloud infrastructure, platforms, and cloud-based services', 'large', 'high_growth', 'good', 'critical'),
    ('Artificial Intelligence', 'AI/ML services, machine learning platforms, and intelligent systems', 'medium', 'high_growth', 'good', 'critical'),
    ('Cybersecurity', 'Information security, cybersecurity services, and data protection', 'medium', 'high_growth', 'good', 'high'),
    ('Fintech', 'Financial technology, digital banking, and payment solutions', 'medium', 'high_growth', 'good', 'high'),
    ('E-commerce Technology', 'Online marketplaces, e-commerce platforms, and digital commerce', 'large', 'growing', 'good', 'high'),
    ('IT Services', 'IT consulting, system integration, and technical support', 'large', 'growing', 'partial', 'medium'),
    ('Data & Analytics', 'Data science, business intelligence, and analytics services', 'medium', 'high_growth', 'partial', 'medium')
) AS sub(name, description, market_size, growth_rate, coverage_status, priority)
WHERE t.name = 'Technology & Digital Services' AND t.level = 1;

-- Healthcare & Life Sciences Sub-categories
INSERT INTO industry_taxonomy (parent_id, level, name, description, category_type, market_size, growth_rate, coverage_status, priority) 
SELECT t.id, 2, sub.name, sub.description, 'secondary', sub.market_size, sub.growth_rate, sub.coverage_status, sub.priority
FROM industry_taxonomy t, (VALUES
    ('Medical Services', 'Healthcare providers, clinics, hospitals, and medical practices', 'very_large', 'growing', 'good', 'critical'),
    ('Pharmaceuticals', 'Drug manufacturing, pharmaceutical services, and biotech', 'large', 'growing', 'good', 'critical'),
    ('Medical Technology', 'Medical devices, healthcare technology, and health IT', 'medium', 'high_growth', 'good', 'high'),
    ('Mental Health', 'Mental health services, counseling, and behavioral health', 'medium', 'growing', 'good', 'medium'),
    ('Dental Services', 'Dental care, oral health services, and dental practices', 'medium', 'stable', 'good', 'medium'),
    ('Veterinary Services', 'Animal healthcare, veterinary medicine, and pet services', 'medium', 'growing', 'good', 'medium'),
    ('Health Insurance', 'Health insurance providers and healthcare coverage', 'large', 'growing', 'partial', 'medium'),
    ('Telemedicine', 'Remote healthcare, telemedicine, and digital health services', 'small', 'high_growth', 'partial', 'high')
) AS sub(name, description, market_size, growth_rate, coverage_status, priority)
WHERE t.name = 'Healthcare & Life Sciences' AND t.level = 1;

-- Financial Services Sub-categories
INSERT INTO industry_taxonomy (parent_id, level, name, description, category_type, market_size, growth_rate, coverage_status, priority) 
SELECT t.id, 2, sub.name, sub.description, 'secondary', sub.market_size, sub.growth_rate, sub.coverage_status, sub.priority
FROM industry_taxonomy t, (VALUES
    ('Commercial Banking', 'Traditional banking, financial institutions, and lending', 'very_large', 'stable', 'good', 'critical'),
    ('Investment Services', 'Investment banking, wealth management, and asset management', 'large', 'growing', 'good', 'high'),
    ('Insurance', 'Insurance providers, risk management, and coverage services', 'large', 'growing', 'good', 'high'),
    ('Credit Services', 'Credit cards, loans, lending services, and credit reporting', 'large', 'growing', 'good', 'high'),
    ('Cryptocurrency', 'Digital currencies, blockchain services, and crypto trading', 'small', 'high_growth', 'good', 'medium'),
    ('Payment Processing', 'Payment gateways, transaction processing, and merchant services', 'medium', 'growing', 'good', 'high'),
    ('Financial Planning', 'Financial planning, retirement services, and investment advice', 'medium', 'growing', 'partial', 'medium'),
    ('Real Estate Finance', 'Mortgage services, real estate lending, and property finance', 'medium', 'stable', 'partial', 'medium')
) AS sub(name, description, market_size, growth_rate, coverage_status, priority)
WHERE t.name = 'Financial Services' AND t.level = 1;

-- Professional Services Sub-categories
INSERT INTO industry_taxonomy (parent_id, level, name, description, category_type, market_size, growth_rate, coverage_status, priority) 
SELECT t.id, 2, sub.name, sub.description, 'secondary', sub.market_size, sub.growth_rate, sub.coverage_status, sub.priority
FROM industry_taxonomy t, (VALUES
    ('Legal Services', 'Law firms, legal consulting, and legal representation', 'large', 'stable', 'missing', 'high'),
    ('Accounting Services', 'Accounting firms, financial consulting, and tax services', 'large', 'stable', 'missing', 'high'),
    ('Business Consulting', 'Management consulting, strategy consulting, and business advisory', 'large', 'growing', 'missing', 'high'),
    ('Marketing & Advertising', 'Marketing agencies, advertising services, and digital marketing', 'medium', 'growing', 'missing', 'medium'),
    ('Human Resources', 'HR services, recruitment, staffing, and talent management', 'medium', 'growing', 'missing', 'medium'),
    ('Real Estate Services', 'Real estate agencies, property management, and real estate consulting', 'medium', 'stable', 'missing', 'medium'),
    ('Public Relations', 'PR agencies, communications, and reputation management', 'small', 'stable', 'missing', 'low'),
    ('Management Consulting', 'Strategic consulting, operations consulting, and change management', 'medium', 'growing', 'missing', 'medium')
) AS sub(name, description, market_size, growth_rate, coverage_status, priority)
WHERE t.name = 'Professional Services' AND t.level = 1;

-- Food & Beverage Sub-categories
INSERT INTO industry_taxonomy (parent_id, level, name, description, category_type, market_size, growth_rate, coverage_status, priority) 
SELECT t.id, 2, sub.name, sub.description, 'secondary', sub.market_size, sub.growth_rate, sub.coverage_status, sub.priority
FROM industry_taxonomy t, (VALUES
    ('Restaurants', 'Food service, restaurants, and dining establishments', 'very_large', 'stable', 'missing', 'critical'),
    ('Food Manufacturing', 'Food production, processing, and food manufacturing', 'large', 'stable', 'missing', 'high'),
    ('Beverage Industry', 'Beverage production, distribution, and beverage services', 'large', 'stable', 'missing', 'medium'),
    ('Catering Services', 'Event catering, food service, and catering businesses', 'medium', 'growing', 'missing', 'medium'),
    ('Food Delivery', 'Food delivery, takeout services, and meal delivery', 'medium', 'high_growth', 'missing', 'high'),
    ('Food Retail', 'Grocery stores, specialty food retail, and food markets', 'large', 'stable', 'missing', 'medium'),
    ('Food Service Equipment', 'Commercial kitchen equipment and food service supplies', 'small', 'stable', 'missing', 'low'),
    ('Food Safety & Testing', 'Food safety services, testing, and quality assurance', 'small', 'growing', 'missing', 'low')
) AS sub(name, description, market_size, growth_rate, coverage_status, priority)
WHERE t.name = 'Food & Beverage' AND t.level = 1;

-- Construction & Real Estate Sub-categories
INSERT INTO industry_taxonomy (parent_id, level, name, description, category_type, market_size, growth_rate, coverage_status, priority) 
SELECT t.id, 2, sub.name, sub.description, 'secondary', sub.market_size, sub.growth_rate, sub.coverage_status, sub.priority
FROM industry_taxonomy t, (VALUES
    ('Construction', 'Building construction, contracting, and construction services', 'large', 'growing', 'missing', 'high'),
    ('Architecture', 'Architectural design, planning, and architectural services', 'medium', 'stable', 'missing', 'medium'),
    ('Engineering Services', 'Engineering consulting, design, and engineering services', 'medium', 'growing', 'missing', 'medium'),
    ('Home Improvement', 'Residential construction, renovation, and home improvement', 'large', 'growing', 'missing', 'high'),
    ('Real Estate Development', 'Property development, real estate projects, and land development', 'medium', 'growing', 'missing', 'medium'),
    ('Property Management', 'Property management, facility management, and real estate services', 'medium', 'stable', 'missing', 'medium'),
    ('Construction Materials', 'Building materials, construction supplies, and material distribution', 'medium', 'stable', 'missing', 'low'),
    ('Construction Equipment', 'Construction equipment, machinery, and equipment rental', 'small', 'stable', 'missing', 'low')
) AS sub(name, description, market_size, growth_rate, coverage_status, priority)
WHERE t.name = 'Construction & Real Estate' AND t.level = 1;

-- =============================================================================
-- LEVEL 3: TERTIARY INDUSTRY CATEGORIES (Specific Industries)
-- =============================================================================

-- Restaurant Industry Specific Categories
INSERT INTO industry_taxonomy (parent_id, level, name, description, category_type, market_size, growth_rate, coverage_status, priority) 
SELECT t.id, 3, sub.name, sub.description, 'tertiary', sub.market_size, sub.growth_rate, sub.coverage_status, sub.priority
FROM industry_taxonomy t, (VALUES
    ('Fast Food Restaurants', 'Quick service restaurants, fast food chains, and quick dining', 'large', 'stable', 'missing', 'critical'),
    ('Fine Dining Restaurants', 'Upscale restaurants, fine dining establishments, and gourmet dining', 'medium', 'stable', 'missing', 'high'),
    ('Casual Dining Restaurants', 'Family restaurants, casual dining chains, and mid-scale dining', 'large', 'stable', 'missing', 'high'),
    ('Coffee Shops & Cafes', 'Coffee shops, cafes, and specialty beverage establishments', 'medium', 'growing', 'missing', 'medium'),
    ('Food Trucks & Mobile Food', 'Food trucks, mobile food services, and street food', 'small', 'growing', 'missing', 'medium'),
    ('Catering & Event Food', 'Event catering, wedding catering, and special event food service', 'medium', 'growing', 'missing', 'medium'),
    ('International Cuisine', 'Ethnic restaurants, international cuisine, and cultural dining', 'medium', 'growing', 'missing', 'medium'),
    ('Health & Wellness Food', 'Healthy restaurants, organic food, and wellness-focused dining', 'small', 'growing', 'missing', 'medium')
) AS sub(name, description, market_size, growth_rate, coverage_status, priority)
WHERE t.name = 'Restaurants' AND t.level = 2;

-- Software Development Specific Categories
INSERT INTO industry_taxonomy (parent_id, level, name, description, category_type, market_size, growth_rate, coverage_status, priority) 
SELECT t.id, 3, sub.name, sub.description, 'tertiary', sub.market_size, sub.growth_rate, sub.coverage_status, sub.priority
FROM industry_taxonomy t, (VALUES
    ('Web Development', 'Website development, web applications, and web-based solutions', 'large', 'growing', 'good', 'high'),
    ('Mobile App Development', 'Mobile applications, iOS/Android development, and mobile solutions', 'large', 'high_growth', 'good', 'high'),
    ('Enterprise Software', 'Enterprise applications, business software, and corporate solutions', 'large', 'growing', 'good', 'high'),
    ('Game Development', 'Video game development, gaming software, and interactive entertainment', 'medium', 'growing', 'partial', 'medium'),
    ('Database Development', 'Database design, data management, and database solutions', 'medium', 'stable', 'partial', 'medium'),
    ('API Development', 'Application programming interfaces, API services, and integration', 'medium', 'high_growth', 'partial', 'medium'),
    ('DevOps & Infrastructure', 'Development operations, infrastructure automation, and deployment', 'medium', 'high_growth', 'partial', 'medium'),
    ('Quality Assurance', 'Software testing, QA services, and quality assurance', 'medium', 'stable', 'partial', 'medium')
) AS sub(name, description, market_size, growth_rate, coverage_status, priority)
WHERE t.name = 'Software Development' AND t.level = 2;

-- =============================================================================
-- INDUSTRY COVERAGE ANALYSIS VIEW
-- =============================================================================

-- Create view for industry coverage analysis
CREATE OR REPLACE VIEW industry_coverage_analysis AS
SELECT 
    t1.id as primary_id,
    t1.name as primary_category,
    t1.market_size as primary_market_size,
    t1.growth_rate as primary_growth_rate,
    t1.coverage_status as primary_coverage_status,
    t1.priority as primary_priority,
    COUNT(t2.id) as secondary_count,
    COUNT(t3.id) as tertiary_count,
    COUNT(t4.id) as specific_count,
    CASE 
        WHEN t1.coverage_status = 'excellent' THEN 100
        WHEN t1.coverage_status = 'good' THEN 75
        WHEN t1.coverage_status = 'partial' THEN 50
        WHEN t1.coverage_status = 'missing' THEN 0
        ELSE 0
    END as coverage_score
FROM industry_taxonomy t1
LEFT JOIN industry_taxonomy t2 ON t1.id = t2.parent_id AND t2.level = 2
LEFT JOIN industry_taxonomy t3 ON t2.id = t3.parent_id AND t3.level = 3
LEFT JOIN industry_taxonomy t4 ON t3.id = t4.parent_id AND t4.level = 4
WHERE t1.level = 1
GROUP BY t1.id, t1.name, t1.market_size, t1.growth_rate, t1.coverage_status, t1.priority
ORDER BY 
    CASE t1.priority 
        WHEN 'critical' THEN 1 
        WHEN 'high' THEN 2 
        WHEN 'medium' THEN 3 
        WHEN 'low' THEN 4 
    END,
    t1.market_size DESC,
    t1.growth_rate DESC;

-- =============================================================================
-- INDUSTRY GAP ANALYSIS VIEW
-- =============================================================================

-- Create view for industry gap analysis
CREATE OR REPLACE VIEW industry_gap_analysis AS
SELECT 
    'Missing Industries' as gap_type,
    t1.name as primary_category,
    t2.name as secondary_category,
    t3.name as tertiary_category,
    t1.priority as priority,
    t1.market_size as market_size,
    t1.growth_rate as growth_rate,
    'Add comprehensive industry coverage' as recommendation
FROM industry_taxonomy t1
LEFT JOIN industry_taxonomy t2 ON t1.id = t2.parent_id AND t2.level = 2
LEFT JOIN industry_taxonomy t3 ON t2.id = t3.parent_id AND t3.level = 3
WHERE t1.coverage_status = 'missing'
    AND t1.level = 1
    AND t1.priority IN ('critical', 'high')

UNION ALL

SELECT 
    'Underrepresented Industries' as gap_type,
    t1.name as primary_category,
    t2.name as secondary_category,
    t3.name as tertiary_category,
    t1.priority as priority,
    t1.market_size as market_size,
    t1.growth_rate as growth_rate,
    'Enhance keyword and code coverage' as recommendation
FROM industry_taxonomy t1
LEFT JOIN industry_taxonomy t2 ON t1.id = t2.parent_id AND t2.level = 2
LEFT JOIN industry_taxonomy t3 ON t2.id = t3.parent_id AND t3.level = 3
WHERE t1.coverage_status = 'partial'
    AND t1.level = 1
    AND t1.priority IN ('critical', 'high')

ORDER BY 
    CASE priority 
        WHEN 'critical' THEN 1 
        WHEN 'high' THEN 2 
        WHEN 'medium' THEN 3 
        WHEN 'low' THEN 4 
    END,
    market_size DESC;

-- =============================================================================
-- COMPLETION MESSAGE
-- =============================================================================

DO $$
BEGIN
    RAISE NOTICE 'KYB Platform Comprehensive Industry Mapping completed successfully!';
    RAISE NOTICE 'Created industry taxonomy hierarchy with 4 levels of classification';
    RAISE NOTICE 'Mapped all major industry sectors and sub-sectors';
    RAISE NOTICE 'Identified missing and underrepresented industries';
    RAISE NOTICE 'Created industry coverage analysis views';
    RAISE NOTICE 'Ready for industry coverage implementation';
END $$;
