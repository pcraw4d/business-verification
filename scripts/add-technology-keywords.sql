-- =============================================================================
-- TECHNOLOGY KEYWORDS COMPREHENSIVE SCRIPT
-- Task 3.2.3: Add technology keywords (50+ technology-specific keywords with base weights 0.5-1.0)
-- =============================================================================
-- This script adds comprehensive technology keywords across all technology-related
-- industries to achieve >85% classification accuracy for technology businesses.
-- 
-- Technology Industries Covered:
-- 1. Technology (general)
-- 2. Software Development
-- 3. Cloud Computing
-- 4. Artificial Intelligence
-- 5. Technology Services
-- 6. Digital Services
-- 7. EdTech
-- 8. Industrial Technology
-- 9. Food Technology
-- 10. Healthcare Technology
-- 11. Fintech
--
-- Total: 200+ comprehensive keywords across 11 technology industries
-- =============================================================================

-- Start transaction for atomic operation
BEGIN;

-- =============================================================================
-- 1. TECHNOLOGY (GENERAL) KEYWORDS
-- =============================================================================
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active, created_at, updated_at)
SELECT i.id, k.keyword, k.weight, true, NOW(), NOW()
FROM industries i, (VALUES
    -- Core Technology Keywords
    ('technology', 1.0000),
    ('tech', 0.9500),
    ('digital', 0.9000),
    ('innovation', 0.8500),
    ('software', 0.8000),
    ('hardware', 0.8000),
    ('computer', 0.7500),
    ('electronic', 0.7500),
    ('automation', 0.7000),
    ('platform', 0.7000),
    ('system', 0.7000),
    ('solution', 0.6500),
    ('service', 0.6500),
    ('development', 0.6000),
    ('engineering', 0.6000),
    ('data', 0.6000),
    ('network', 0.5500),
    ('security', 0.5500),
    ('integration', 0.5500),
    ('optimization', 0.5000)
) AS k(keyword, weight)
WHERE i.name = 'Technology' AND i.is_active = true
ON CONFLICT (industry_id, keyword) DO UPDATE SET
    weight = EXCLUDED.weight,
    updated_at = NOW();

-- =============================================================================
-- 2. SOFTWARE DEVELOPMENT KEYWORDS
-- =============================================================================
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active, created_at, updated_at)
SELECT i.id, k.keyword, k.weight, true, NOW(), NOW()
FROM industries i, (VALUES
    -- Core Software Development Keywords
    ('software', 1.0000),
    ('development', 0.9500),
    ('programming', 0.9000),
    ('coding', 0.9000),
    ('application', 0.8500),
    ('app', 0.8000),
    ('web', 0.8000),
    ('mobile', 0.8000),
    ('desktop', 0.7500),
    ('api', 0.7500),
    ('framework', 0.7000),
    ('library', 0.7000),
    ('database', 0.7000),
    ('backend', 0.7000),
    ('frontend', 0.7000),
    ('fullstack', 0.6500),
    ('microservice', 0.6500),
    ('devops', 0.6500),
    ('deployment', 0.6000),
    ('testing', 0.6000),
    ('debugging', 0.6000),
    ('version control', 0.6000),
    ('git', 0.5500),
    ('agile', 0.5500),
    ('scrum', 0.5500),
    ('code review', 0.5500),
    ('refactoring', 0.5000),
    ('architecture', 0.5000),
    ('algorithm', 0.5000),
    ('performance', 0.5000)
) AS k(keyword, weight)
WHERE i.name = 'Software Development' AND i.is_active = true
ON CONFLICT (industry_id, keyword) DO UPDATE SET
    weight = EXCLUDED.weight,
    updated_at = NOW();

-- =============================================================================
-- 3. CLOUD COMPUTING KEYWORDS
-- =============================================================================
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active, created_at, updated_at)
SELECT i.id, k.keyword, k.weight, true, NOW(), NOW()
FROM industries i, (VALUES
    -- Core Cloud Computing Keywords
    ('cloud', 1.0000),
    ('aws', 0.9000),
    ('azure', 0.9000),
    ('google cloud', 0.9000),
    ('infrastructure', 0.8500),
    ('saas', 0.8500),
    ('paas', 0.8000),
    ('iaas', 0.8000),
    ('hosting', 0.7500),
    ('server', 0.7500),
    ('virtualization', 0.7000),
    ('container', 0.7000),
    ('docker', 0.7000),
    ('kubernetes', 0.7000),
    ('scalability', 0.6500),
    ('elasticity', 0.6500),
    ('load balancing', 0.6500),
    ('cdn', 0.6000),
    ('storage', 0.6000),
    ('backup', 0.6000),
    ('disaster recovery', 0.6000),
    ('monitoring', 0.5500),
    ('logging', 0.5500),
    ('security', 0.5500),
    ('compliance', 0.5500),
    ('migration', 0.5000),
    ('hybrid cloud', 0.5000),
    ('multi cloud', 0.5000),
    ('serverless', 0.5000),
    ('edge computing', 0.5000)
) AS k(keyword, weight)
WHERE i.name = 'Cloud Computing' AND i.is_active = true
ON CONFLICT (industry_id, keyword) DO UPDATE SET
    weight = EXCLUDED.weight,
    updated_at = NOW();

-- =============================================================================
-- 4. ARTIFICIAL INTELLIGENCE KEYWORDS
-- =============================================================================
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active, created_at, updated_at)
SELECT i.id, k.keyword, k.weight, true, NOW(), NOW()
FROM industries i, (VALUES
    -- Core AI Keywords
    ('artificial intelligence', 1.0000),
    ('ai', 1.0000),
    ('machine learning', 0.9500),
    ('ml', 0.9500),
    ('deep learning', 0.9000),
    ('neural network', 0.9000),
    ('algorithm', 0.8500),
    ('data science', 0.8500),
    ('predictive analytics', 0.8000),
    ('natural language processing', 0.8000),
    ('nlp', 0.8000),
    ('computer vision', 0.8000),
    ('chatbot', 0.7500),
    ('automation', 0.7500),
    ('intelligent', 0.7000),
    ('smart', 0.7000),
    ('cognitive', 0.7000),
    ('robotics', 0.7000),
    ('pattern recognition', 0.6500),
    ('classification', 0.6500),
    ('regression', 0.6500),
    ('clustering', 0.6500),
    ('reinforcement learning', 0.6000),
    ('supervised learning', 0.6000),
    ('unsupervised learning', 0.6000),
    ('feature engineering', 0.6000),
    ('model training', 0.6000),
    ('inference', 0.5500),
    ('optimization', 0.5500),
    ('tensorflow', 0.5000),
    ('pytorch', 0.5000)
) AS k(keyword, weight)
WHERE i.name = 'Artificial Intelligence' AND i.is_active = true
ON CONFLICT (industry_id, keyword) DO UPDATE SET
    weight = EXCLUDED.weight,
    updated_at = NOW();

-- =============================================================================
-- 5. TECHNOLOGY SERVICES KEYWORDS
-- =============================================================================
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active, created_at, updated_at)
SELECT i.id, k.keyword, k.weight, true, NOW(), NOW()
FROM industries i, (VALUES
    -- Core Technology Services Keywords
    ('it services', 1.0000),
    ('technology consulting', 0.9500),
    ('technical support', 0.9000),
    ('help desk', 0.8500),
    ('system administration', 0.8500),
    ('network administration', 0.8000),
    ('database administration', 0.8000),
    ('security services', 0.8000),
    ('managed services', 0.7500),
    ('outsourcing', 0.7500),
    ('maintenance', 0.7000),
    ('upgrade', 0.7000),
    ('migration', 0.7000),
    ('integration', 0.7000),
    ('implementation', 0.7000),
    ('consulting', 0.6500),
    ('advisory', 0.6500),
    ('strategy', 0.6500),
    ('planning', 0.6500),
    ('architecture', 0.6500),
    ('design', 0.6000),
    ('deployment', 0.6000),
    ('training', 0.6000),
    ('documentation', 0.6000),
    ('troubleshooting', 0.6000),
    ('monitoring', 0.5500),
    ('backup', 0.5500),
    ('recovery', 0.5500),
    ('compliance', 0.5500),
    ('governance', 0.5000)
) AS k(keyword, weight)
WHERE i.name = 'Technology Services' AND i.is_active = true
ON CONFLICT (industry_id, keyword) DO UPDATE SET
    weight = EXCLUDED.weight,
    updated_at = NOW();

-- =============================================================================
-- 6. DIGITAL SERVICES KEYWORDS
-- =============================================================================
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active, created_at, updated_at)
SELECT i.id, k.keyword, k.weight, true, NOW(), NOW()
FROM industries i, (VALUES
    -- Core Digital Services Keywords
    ('digital marketing', 1.0000),
    ('web development', 0.9500),
    ('digital transformation', 0.9000),
    ('online', 0.8500),
    ('website', 0.8500),
    ('ecommerce', 0.8000),
    ('seo', 0.8000),
    ('sem', 0.7500),
    ('social media', 0.7500),
    ('content marketing', 0.7500),
    ('email marketing', 0.7000),
    ('ppc', 0.7000),
    ('analytics', 0.7000),
    ('conversion', 0.7000),
    ('user experience', 0.7000),
    ('ux', 0.7000),
    ('ui', 0.7000),
    ('responsive', 0.6500),
    ('mobile first', 0.6500),
    ('performance', 0.6500),
    ('optimization', 0.6500),
    ('automation', 0.6000),
    ('personalization', 0.6000),
    ('crm', 0.6000),
    ('lead generation', 0.6000),
    ('branding', 0.5500),
    ('design', 0.5500),
    ('creative', 0.5500),
    ('strategy', 0.5500),
    ('campaign', 0.5000)
) AS k(keyword, weight)
WHERE i.name = 'Digital Services' AND i.is_active = true
ON CONFLICT (industry_id, keyword) DO UPDATE SET
    weight = EXCLUDED.weight,
    updated_at = NOW();

-- =============================================================================
-- 7. EDTECH KEYWORDS
-- =============================================================================
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active, created_at, updated_at)
SELECT i.id, k.keyword, k.weight, true, NOW(), NOW()
FROM industries i, (VALUES
    -- Core EdTech Keywords
    ('educational technology', 1.0000),
    ('edtech', 1.0000),
    ('online learning', 0.9500),
    ('e-learning', 0.9500),
    ('learning management system', 0.9000),
    ('lms', 0.9000),
    ('virtual classroom', 0.8500),
    ('distance learning', 0.8500),
    ('remote learning', 0.8000),
    ('digital education', 0.8000),
    ('interactive', 0.7500),
    ('multimedia', 0.7500),
    ('gamification', 0.7000),
    ('adaptive learning', 0.7000),
    ('personalized learning', 0.7000),
    ('assessment', 0.7000),
    ('quiz', 0.6500),
    ('certification', 0.6500),
    ('training', 0.6500),
    ('course', 0.6500),
    ('curriculum', 0.6000),
    ('pedagogy', 0.6000),
    ('instructional design', 0.6000),
    ('student engagement', 0.6000),
    ('progress tracking', 0.6000),
    ('analytics', 0.5500),
    ('reporting', 0.5500),
    ('collaboration', 0.5500),
    ('communication', 0.5500),
    ('accessibility', 0.5000)
) AS k(keyword, weight)
WHERE i.name = 'EdTech' AND i.is_active = true
ON CONFLICT (industry_id, keyword) DO UPDATE SET
    weight = EXCLUDED.weight,
    updated_at = NOW();

-- =============================================================================
-- 8. INDUSTRIAL TECHNOLOGY KEYWORDS
-- =============================================================================
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active, created_at, updated_at)
SELECT i.id, k.keyword, k.weight, true, NOW(), NOW()
FROM industries i, (VALUES
    -- Core Industrial Technology Keywords
    ('industrial automation', 1.0000),
    ('iot', 0.9500),
    ('internet of things', 0.9500),
    ('smart manufacturing', 0.9000),
    ('industry 4.0', 0.9000),
    ('robotics', 0.8500),
    ('automation', 0.8500),
    ('sensors', 0.8000),
    ('actuators', 0.8000),
    ('plc', 0.7500),
    ('scada', 0.7500),
    ('hmi', 0.7000),
    ('m2m', 0.7000),
    ('machine to machine', 0.7000),
    ('predictive maintenance', 0.7000),
    ('condition monitoring', 0.6500),
    ('digital twin', 0.6500),
    ('cyber physical', 0.6500),
    ('connected devices', 0.6500),
    ('edge computing', 0.6000),
    ('fog computing', 0.6000),
    ('data analytics', 0.6000),
    ('real time', 0.6000),
    ('control systems', 0.6000),
    ('process optimization', 0.5500),
    ('quality control', 0.5500),
    ('supply chain', 0.5500),
    ('logistics', 0.5500),
    ('warehouse', 0.5000),
    ('inventory', 0.5000)
) AS k(keyword, weight)
WHERE i.name = 'Industrial Technology' AND i.is_active = true
ON CONFLICT (industry_id, keyword) DO UPDATE SET
    weight = EXCLUDED.weight,
    updated_at = NOW();

-- =============================================================================
-- 9. FOOD TECHNOLOGY KEYWORDS
-- =============================================================================
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active, created_at, updated_at)
SELECT i.id, k.keyword, k.weight, true, NOW(), NOW()
FROM industries i, (VALUES
    -- Core Food Technology Keywords
    ('food technology', 1.0000),
    ('food innovation', 0.9500),
    ('alternative protein', 0.9000),
    ('plant based', 0.9000),
    ('lab grown', 0.8500),
    ('cultured meat', 0.8500),
    ('food science', 0.8000),
    ('nutrition', 0.8000),
    ('food safety', 0.8000),
    ('sustainability', 0.7500),
    ('food waste', 0.7500),
    ('packaging', 0.7000),
    ('preservation', 0.7000),
    ('processing', 0.7000),
    ('fermentation', 0.7000),
    ('biotechnology', 0.6500),
    ('enzymes', 0.6500),
    ('probiotics', 0.6500),
    ('functional food', 0.6500),
    ('supplements', 0.6000),
    ('ingredients', 0.6000),
    ('formulation', 0.6000),
    ('testing', 0.6000),
    ('quality control', 0.6000),
    ('traceability', 0.5500),
    ('blockchain', 0.5500),
    ('supply chain', 0.5500),
    ('logistics', 0.5500),
    ('cold chain', 0.5000),
    ('shelf life', 0.5000)
) AS k(keyword, weight)
WHERE i.name = 'Food Technology' AND i.is_active = true
ON CONFLICT (industry_id, keyword) DO UPDATE SET
    weight = EXCLUDED.weight,
    updated_at = NOW();

-- =============================================================================
-- 10. HEALTHCARE TECHNOLOGY KEYWORDS
-- =============================================================================
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active, created_at, updated_at)
SELECT i.id, k.keyword, k.weight, true, NOW(), NOW()
FROM industries i, (VALUES
    -- Core Healthcare Technology Keywords
    ('healthcare technology', 1.0000),
    ('medical devices', 0.9500),
    ('health it', 0.9000),
    ('digital health', 0.9000),
    ('telemedicine', 0.8500),
    ('telehealth', 0.8500),
    ('electronic health records', 0.8000),
    ('ehr', 0.8000),
    ('electronic medical records', 0.8000),
    ('emr', 0.8000),
    ('health information system', 0.7500),
    ('his', 0.7500),
    ('patient monitoring', 0.7500),
    ('diagnostic', 0.7000),
    ('imaging', 0.7000),
    ('radiology', 0.7000),
    ('laboratory', 0.7000),
    ('pharmacy', 0.7000),
    ('clinical decision support', 0.6500),
    ('cds', 0.6500),
    ('health analytics', 0.6500),
    ('population health', 0.6500),
    ('interoperability', 0.6000),
    ('hl7', 0.6000),
    ('fhir', 0.6000),
    ('compliance', 0.6000),
    ('hipaa', 0.6000),
    ('security', 0.5500),
    ('privacy', 0.5500),
    ('workflow', 0.5000)
) AS k(keyword, weight)
WHERE i.name = 'Healthcare Technology' AND i.is_active = true
ON CONFLICT (industry_id, keyword) DO UPDATE SET
    weight = EXCLUDED.weight,
    updated_at = NOW();

-- =============================================================================
-- 11. FINTECH KEYWORDS
-- =============================================================================
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active, created_at, updated_at)
SELECT i.id, k.keyword, k.weight, true, NOW(), NOW()
FROM industries i, (VALUES
    -- Core Fintech Keywords
    ('fintech', 1.0000),
    ('financial technology', 1.0000),
    ('digital banking', 0.9500),
    ('mobile banking', 0.9000),
    ('online banking', 0.9000),
    ('payment', 0.8500),
    ('digital payment', 0.8500),
    ('mobile payment', 0.8000),
    ('cryptocurrency', 0.8000),
    ('blockchain', 0.8000),
    ('digital wallet', 0.7500),
    ('p2p', 0.7500),
    ('peer to peer', 0.7500),
    ('lending', 0.7000),
    ('digital lending', 0.7000),
    ('robo advisor', 0.7000),
    ('wealth management', 0.7000),
    ('investment', 0.7000),
    ('trading', 0.7000),
    ('insurtech', 0.6500),
    ('insurance technology', 0.6500),
    ('regtech', 0.6500),
    ('regulatory technology', 0.6500),
    ('compliance', 0.6500),
    ('kyc', 0.6000),
    ('aml', 0.6000),
    ('fraud detection', 0.6000),
    ('risk management', 0.6000),
    ('api', 0.5500),
    ('open banking', 0.5000)
) AS k(keyword, weight)
WHERE i.name = 'Fintech' AND i.is_active = true
ON CONFLICT (industry_id, keyword) DO UPDATE SET
    weight = EXCLUDED.weight,
    updated_at = NOW();

-- =============================================================================
-- VERIFICATION QUERIES
-- =============================================================================

-- Verify all technology keywords were added
DO $$
DECLARE
    keyword_count INTEGER;
    expected_count INTEGER := 200; -- 200+ keywords across all technology industries
BEGIN
    SELECT COUNT(*) INTO keyword_count
    FROM industry_keywords ik
    JOIN industries i ON ik.industry_id = i.id
    WHERE i.is_active = true
    AND i.name IN (
        'Technology', 'Software Development', 'Cloud Computing', 'Artificial Intelligence',
        'Technology Services', 'Digital Services', 'EdTech', 'Industrial Technology',
        'Food Technology', 'Healthcare Technology', 'Fintech'
    )
    AND ik.is_active = true;
    
    IF keyword_count >= expected_count THEN
        RAISE NOTICE 'SUCCESS: % technology keywords added successfully', keyword_count;
    ELSE
        RAISE NOTICE 'WARNING: Expected %+ keywords, but found %', expected_count, keyword_count;
    END IF;
END $$;

-- Display technology keyword summary by industry
SELECT 
    'TECHNOLOGY KEYWORDS SUMMARY' as summary_type,
    '' as spacer;

SELECT 
    i.name as industry_name,
    COUNT(ik.keyword) as keyword_count,
    ROUND(AVG(ik.weight), 4) as avg_weight,
    ROUND(MIN(ik.weight), 4) as min_weight,
    ROUND(MAX(ik.weight), 4) as max_weight
FROM industries i
LEFT JOIN industry_keywords ik ON i.id = ik.industry_id AND ik.is_active = true
WHERE i.is_active = true
AND i.name IN (
    'Technology', 'Software Development', 'Cloud Computing', 'Artificial Intelligence',
    'Technology Services', 'Digital Services', 'EdTech', 'Industrial Technology',
    'Food Technology', 'Healthcare Technology', 'Fintech'
)
GROUP BY i.name
ORDER BY keyword_count DESC;

-- Display sample keywords for each technology industry
SELECT 
    'SAMPLE TECHNOLOGY KEYWORDS' as summary_type,
    '' as spacer;

SELECT 
    i.name as industry_name,
    ik.keyword,
    ik.weight
FROM industries i
JOIN industry_keywords ik ON i.id = ik.industry_id
WHERE i.is_active = true
AND i.name IN (
    'Technology', 'Software Development', 'Cloud Computing', 'Artificial Intelligence',
    'Technology Services', 'Digital Services', 'EdTech', 'Industrial Technology',
    'Food Technology', 'Healthcare Technology', 'Fintech'
)
AND ik.is_active = true
ORDER BY i.name, ik.weight DESC
LIMIT 50;

-- Final verification
SELECT 
    'FINAL VERIFICATION' as summary_type,
    '' as spacer;

SELECT 
    COUNT(DISTINCT i.name) as technology_industries,
    COUNT(ik.keyword) as total_technology_keywords,
    ROUND(AVG(ik.weight), 4) as avg_keyword_weight
FROM industries i
JOIN industry_keywords ik ON i.id = ik.industry_id
WHERE i.is_active = true
AND i.name IN (
    'Technology', 'Software Development', 'Cloud Computing', 'Artificial Intelligence',
    'Technology Services', 'Digital Services', 'EdTech', 'Industrial Technology',
    'Food Technology', 'Healthcare Technology', 'Fintech'
)
AND ik.is_active = true;

COMMIT;

-- =============================================================================
-- COMPLETION MESSAGE
-- =============================================================================
DO $$
BEGIN
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'TECHNOLOGY KEYWORDS COMPREHENSIVE SCRIPT COMPLETED';
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'Technology industries covered: 11';
    RAISE NOTICE 'Total keywords added: 200+';
    RAISE NOTICE 'Keyword weight range: 0.5000-1.0000';
    RAISE NOTICE 'Status: Ready for technology classification testing';
    RAISE NOTICE 'Next: Test technology classification accuracy (Task 3.2.3 testing)';
    RAISE NOTICE '=============================================================================';
END $$;
