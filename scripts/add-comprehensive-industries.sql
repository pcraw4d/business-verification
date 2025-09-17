-- =============================================================================
-- COMPREHENSIVE INDUSTRY EXPANSION SCRIPT
-- =============================================================================
-- This script adds 25+ industries across 7 major categories to achieve
-- comprehensive business classification coverage for the KYB Platform.
-- 
-- Categories:
-- 1. Legal Services (4 industries)
-- 2. Healthcare (4 industries)
-- 3. Financial Services (4 industries)
-- 4. Retail & E-commerce (4 industries)
-- 5. Manufacturing (4 industries)
-- 6. Agriculture & Energy (4 industries)
-- 7. Technology (3 industries)
--
-- Total: 27 new industries + existing 12 restaurant industries = 39 total
-- =============================================================================

-- Start transaction for atomic operation
BEGIN;

-- =============================================================================
-- 1. LEGAL SERVICES INDUSTRIES (4 industries)
-- =============================================================================
INSERT INTO industries (name, description, category, confidence_threshold, is_active, created_at, updated_at) VALUES
-- Primary Legal Categories
('Law Firms', 'Full-service law firms providing comprehensive legal services across multiple practice areas', 'traditional', 0.80, true, NOW(), NOW()),
('Legal Consulting', 'Legal consulting services, advisory, and specialized legal expertise', 'traditional', 0.75, true, NOW(), NOW()),
('Legal Services', 'General legal services including paralegal, legal research, and support services', 'traditional', 0.70, true, NOW(), NOW()),
('Intellectual Property', 'Intellectual property law, patent attorneys, trademark services, and IP consulting', 'traditional', 0.85, true, NOW(), NOW())

-- =============================================================================
-- 2. HEALTHCARE INDUSTRIES (4 industries)
-- =============================================================================
,
('Medical Practices', 'Medical practices including family medicine, specialists, and clinical services', 'traditional', 0.80, true, NOW(), NOW()),
('Healthcare Services', 'Healthcare services including hospitals, clinics, and medical facilities', 'traditional', 0.75, true, NOW(), NOW()),
('Mental Health', 'Mental health services, counseling, therapy, and psychological services', 'traditional', 0.80, true, NOW(), NOW()),
('Healthcare Technology', 'Healthcare technology, medical devices, health IT, and digital health solutions', 'emerging', 0.75, true, NOW(), NOW())

-- =============================================================================
-- 3. FINANCIAL SERVICES INDUSTRIES (4 industries)
-- =============================================================================
,
('Banking', 'Commercial banking, retail banking, and financial institutions', 'traditional', 0.80, true, NOW(), NOW()),
('Insurance', 'Insurance services including life, health, property, and casualty insurance', 'traditional', 0.75, true, NOW(), NOW()),
('Investment Services', 'Investment advisory, wealth management, and financial planning services', 'traditional', 0.80, true, NOW(), NOW()),
('Fintech', 'Financial technology, digital payments, blockchain, and financial innovation', 'emerging', 0.75, true, NOW(), NOW())

-- =============================================================================
-- 4. RETAIL & E-COMMERCE INDUSTRIES (4 industries)
-- =============================================================================
,
('Retail', 'Traditional retail stores, brick-and-mortar retail, and consumer goods sales', 'traditional', 0.70, true, NOW(), NOW()),
('E-commerce', 'Online retail, e-commerce platforms, and digital commerce solutions', 'emerging', 0.75, true, NOW(), NOW()),
('Wholesale', 'Wholesale trade, distribution, and B2B sales of goods and products', 'traditional', 0.70, true, NOW(), NOW()),
('Consumer Goods', 'Consumer goods manufacturing, retail, and distribution', 'traditional', 0.70, true, NOW(), NOW())

-- =============================================================================
-- 5. MANUFACTURING INDUSTRIES (4 industries)
-- =============================================================================
,
('Manufacturing', 'General manufacturing, production, and industrial manufacturing', 'traditional', 0.75, true, NOW(), NOW()),
('Industrial Manufacturing', 'Heavy industry, machinery, equipment, and industrial production', 'traditional', 0.80, true, NOW(), NOW()),
('Consumer Manufacturing', 'Consumer goods manufacturing, electronics, and consumer products', 'traditional', 0.75, true, NOW(), NOW()),
('Advanced Manufacturing', 'Advanced manufacturing, automation, robotics, and smart manufacturing', 'emerging', 0.80, true, NOW(), NOW())

-- =============================================================================
-- 6. AGRICULTURE & ENERGY INDUSTRIES (4 industries)
-- =============================================================================
,
('Agriculture', 'Farming, crop production, livestock, and agricultural services', 'traditional', 0.75, true, NOW(), NOW()),
('Food Production', 'Food processing, packaging, and food manufacturing', 'traditional', 0.70, true, NOW(), NOW()),
('Energy Services', 'Traditional energy services, oil, gas, and conventional energy', 'traditional', 0.75, true, NOW(), NOW()),
('Renewable Energy', 'Renewable energy, solar, wind, and sustainable energy solutions', 'emerging', 0.80, true, NOW(), NOW())

-- =============================================================================
-- 7. TECHNOLOGY INDUSTRIES (3 industries)
-- =============================================================================
,
('Software Development', 'Software development, programming, and software engineering services', 'emerging', 0.80, true, NOW(), NOW()),
('Technology Services', 'IT services, technology consulting, and technical support services', 'emerging', 0.75, true, NOW(), NOW()),
('Digital Services', 'Digital marketing, web development, and digital transformation services', 'emerging', 0.70, true, NOW(), NOW())

ON CONFLICT (name) DO UPDATE SET
    description = EXCLUDED.description,
    category = EXCLUDED.category,
    confidence_threshold = EXCLUDED.confidence_threshold,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();

-- =============================================================================
-- VERIFICATION QUERIES
-- =============================================================================

-- Verify all industries were added
DO $$
DECLARE
    industry_count INTEGER;
    expected_count INTEGER := 27; -- 27 new industries
BEGIN
    SELECT COUNT(*) INTO industry_count
    FROM industries
    WHERE is_active = true
    AND name IN (
        -- Legal Services
        'Law Firms', 'Legal Consulting', 'Legal Services', 'Intellectual Property',
        -- Healthcare
        'Medical Practices', 'Healthcare Services', 'Mental Health', 'Healthcare Technology',
        -- Financial Services
        'Banking', 'Insurance', 'Investment Services', 'Fintech',
        -- Retail & E-commerce
        'Retail', 'E-commerce', 'Wholesale', 'Consumer Goods',
        -- Manufacturing
        'Manufacturing', 'Industrial Manufacturing', 'Consumer Manufacturing', 'Advanced Manufacturing',
        -- Agriculture & Energy
        'Agriculture', 'Food Production', 'Energy Services', 'Renewable Energy',
        -- Technology
        'Software Development', 'Technology Services', 'Digital Services'
    );
    
    IF industry_count = expected_count THEN
        RAISE NOTICE 'SUCCESS: All % new industries added successfully', industry_count;
    ELSE
        RAISE NOTICE 'WARNING: Expected % industries, but found %', expected_count, industry_count;
    END IF;
END $$;

-- Display industry summary by category
SELECT 
    'INDUSTRY EXPANSION SUMMARY' as summary_type,
    '' as spacer;

SELECT 
    category,
    COUNT(*) as industry_count,
    ROUND(AVG(confidence_threshold), 2) as avg_confidence_threshold
FROM industries
WHERE is_active = true
GROUP BY category
ORDER BY industry_count DESC;

-- Display all new industries
SELECT 
    'NEW INDUSTRIES ADDED' as summary_type,
    '' as spacer;

SELECT 
    name,
    category,
    confidence_threshold,
    CASE 
        WHEN confidence_threshold >= 0.80 THEN 'High'
        WHEN confidence_threshold >= 0.75 THEN 'Medium-High'
        WHEN confidence_threshold >= 0.70 THEN 'Medium'
        ELSE 'Low'
    END as confidence_level
FROM industries
WHERE is_active = true
AND name IN (
    'Law Firms', 'Legal Consulting', 'Legal Services', 'Intellectual Property',
    'Medical Practices', 'Healthcare Services', 'Mental Health', 'Healthcare Technology',
    'Banking', 'Insurance', 'Investment Services', 'Fintech',
    'Retail', 'E-commerce', 'Wholesale', 'Consumer Goods',
    'Manufacturing', 'Industrial Manufacturing', 'Consumer Manufacturing', 'Advanced Manufacturing',
    'Agriculture', 'Food Production', 'Energy Services', 'Renewable Energy',
    'Software Development', 'Technology Services', 'Digital Services'
)
ORDER BY category, confidence_threshold DESC;

-- Final verification
SELECT 
    'FINAL VERIFICATION' as summary_type,
    '' as spacer;

SELECT 
    COUNT(*) as total_industries,
    COUNT(CASE WHEN category = 'traditional' THEN 1 END) as traditional_industries,
    COUNT(CASE WHEN category = 'emerging' THEN 1 END) as emerging_industries,
    COUNT(CASE WHEN category = 'hybrid' THEN 1 END) as hybrid_industries
FROM industries
WHERE is_active = true;

COMMIT;

-- =============================================================================
-- COMPLETION MESSAGE
-- =============================================================================
DO $$
BEGIN
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'COMPREHENSIVE INDUSTRY EXPANSION COMPLETED';
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'New industries added: 27';
    RAISE NOTICE 'Total industries: 39 (12 restaurant + 27 new)';
    RAISE NOTICE 'Categories covered: 7 major business categories';
    RAISE NOTICE 'Confidence thresholds: 0.70-0.85 range';
    RAISE NOTICE 'Status: Ready for keyword expansion (Task 3.2)';
    RAISE NOTICE '=============================================================================';
END $$;
