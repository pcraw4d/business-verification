-- =============================================================================
-- TASK 5.1.2: ADDITIONAL INDUSTRY KEYWORDS EXPANSION
-- =============================================================================
-- This script adds additional keywords for key industries to reach 2000+ total.
-- Focus: Healthcare, Financial Services, Retail, Legal Services, Manufacturing
-- =============================================================================

BEGIN;

-- =============================================================================
-- HEALTHCARE INDUSTRY - ADDITIONAL KEYWORDS (50+ keywords)
-- =============================================================================

INSERT INTO industry_keywords (industry_id, keyword, weight, context, is_primary, created_at, updated_at)
SELECT i.id, k.keyword, k.weight, k.context, k.is_primary, NOW(), NOW()
FROM industries i
CROSS JOIN (VALUES
    -- Advanced Healthcare Terms
    ('telemedicine', 0.95, 'technical', true),
    ('telehealth', 0.90, 'technical', true),
    ('electronic health records', 0.90, 'technical', true),
    ('ehr', 0.85, 'technical', false),
    ('electronic medical records', 0.85, 'technical', false),
    ('emr', 0.80, 'technical', false),
    ('health information systems', 0.85, 'technical', false),
    ('medical informatics', 0.80, 'technical', false),
    ('clinical decision support', 0.85, 'technical', false),
    ('medical imaging', 0.90, 'technical', true),
    ('radiology', 0.90, 'business', true),
    ('ultrasound', 0.85, 'technical', false),
    ('mri', 0.85, 'technical', false),
    ('ct scan', 0.85, 'technical', false),
    ('x-ray', 0.80, 'technical', false),
    ('mammography', 0.80, 'technical', false),
    ('pathology', 0.90, 'business', true),
    ('laboratory', 0.90, 'business', true),
    ('clinical laboratory', 0.85, 'business', false),
    ('blood work', 0.80, 'technical', false),
    ('diagnostics', 0.90, 'technical', true),
    ('biomarkers', 0.80, 'technical', false),
    ('genetic testing', 0.85, 'technical', false),
    ('genomics', 0.80, 'technical', false),
    ('precision medicine', 0.85, 'technical', false),
    ('personalized medicine', 0.85, 'technical', false),
    ('clinical trials', 0.85, 'technical', false),
    ('research', 0.80, 'business', false),
    ('drug development', 0.80, 'technical', false),
    ('pharmaceutical research', 0.80, 'technical', false),
    ('biotechnology', 0.85, 'technical', false),
    ('biotech', 0.80, 'technical', false),
    ('medical devices', 0.90, 'technical', true),
    ('surgical instruments', 0.85, 'technical', false),
    ('implants', 0.80, 'technical', false),
    ('prosthetics', 0.80, 'technical', false),
    ('rehabilitation', 0.85, 'business', false),
    ('physical therapy', 0.90, 'business', true),
    ('occupational therapy', 0.85, 'business', false),
    ('speech therapy', 0.80, 'business', false),
    ('mental health', 0.95, 'business', true),
    ('psychiatry', 0.90, 'business', true),
    ('psychology', 0.85, 'business', false),
    ('counseling', 0.85, 'business', false),
    ('behavioral health', 0.85, 'business', false),
    ('addiction treatment', 0.80, 'business', false),
    ('substance abuse', 0.75, 'business', false),
    ('wellness', 0.80, 'business', false),
    ('preventive care', 0.80, 'business', false),
    ('public health', 0.85, 'business', false),
    ('epidemiology', 0.75, 'technical', false)
) AS k(keyword, weight, context, is_primary)
WHERE i.name = 'Healthcare'
ON CONFLICT (industry_id, keyword) DO NOTHING;

-- =============================================================================
-- FINANCIAL SERVICES - ADDITIONAL KEYWORDS (50+ keywords)
-- =============================================================================

INSERT INTO industry_keywords (industry_id, keyword, weight, context, is_primary, created_at, updated_at)
SELECT i.id, k.keyword, k.weight, k.context, k.is_primary, NOW(), NOW()
FROM industries i
CROSS JOIN (VALUES
    -- Advanced Financial Terms
    ('fintech', 0.95, 'technical', true),
    ('financial technology', 0.90, 'technical', true),
    ('digital banking', 0.90, 'technical', true),
    ('mobile banking', 0.85, 'technical', false),
    ('online banking', 0.85, 'technical', false),
    ('neobank', 0.80, 'technical', false),
    ('challenger bank', 0.75, 'technical', false),
    ('payment processing', 0.90, 'technical', true),
    ('payment gateway', 0.85, 'technical', false),
    ('merchant services', 0.85, 'business', false),
    ('point of sale', 0.80, 'business', false),
    ('pos', 0.75, 'business', false),
    ('contactless payment', 0.80, 'technical', false),
    ('mobile payment', 0.85, 'technical', false),
    ('digital wallet', 0.85, 'technical', false),
    ('cryptocurrency', 0.90, 'technical', true),
    ('bitcoin', 0.85, 'technical', false),
    ('ethereum', 0.80, 'technical', false),
    ('blockchain', 0.90, 'technical', true),
    ('defi', 0.85, 'technical', false),
    ('decentralized finance', 0.80, 'technical', false),
    ('smart contracts', 0.80, 'technical', false),
    ('digital assets', 0.85, 'technical', false),
    ('asset management', 0.90, 'business', true),
    ('wealth management', 0.90, 'business', true),
    ('portfolio management', 0.85, 'business', false),
    ('investment advisory', 0.85, 'business', false),
    ('financial planning', 0.85, 'business', false),
    ('retirement planning', 0.80, 'business', false),
    ('estate planning', 0.80, 'business', false),
    ('tax planning', 0.80, 'business', false),
    ('robo advisor', 0.80, 'technical', false),
    ('algorithmic trading', 0.85, 'technical', false),
    ('high frequency trading', 0.80, 'technical', false),
    ('quantitative analysis', 0.80, 'technical', false),
    ('risk management', 0.90, 'business', true),
    ('credit risk', 0.85, 'technical', false),
    ('market risk', 0.85, 'technical', false),
    ('operational risk', 0.80, 'technical', false),
    ('compliance', 0.90, 'business', true),
    ('regulatory', 0.85, 'business', false),
    ('aml', 0.80, 'technical', false),
    ('anti-money laundering', 0.80, 'technical', false),
    ('kyc', 0.80, 'technical', false),
    ('know your customer', 0.75, 'technical', false),
    ('fraud detection', 0.85, 'technical', false),
    ('fraud prevention', 0.85, 'technical', false),
    ('identity verification', 0.80, 'technical', false),
    ('credit scoring', 0.85, 'technical', false),
    ('underwriting', 0.85, 'business', false)
) AS k(keyword, weight, context, is_primary)
WHERE i.name = 'Financial Services'
ON CONFLICT (industry_id, keyword) DO NOTHING;

-- =============================================================================
-- RETAIL & E-COMMERCE - ADDITIONAL KEYWORDS (50+ keywords)
-- =============================================================================

INSERT INTO industry_keywords (industry_id, keyword, weight, context, is_primary, created_at, updated_at)
SELECT i.id, k.keyword, k.weight, k.context, k.is_primary, NOW(), NOW()
FROM industries i
CROSS JOIN (VALUES
    -- Advanced Retail Terms
    ('ecommerce', 0.95, 'technical', true),
    ('e-commerce', 0.95, 'technical', true),
    ('online retail', 0.90, 'business', true),
    ('digital commerce', 0.85, 'technical', false),
    ('omnichannel', 0.90, 'business', true),
    ('multichannel', 0.85, 'business', false),
    ('marketplace', 0.90, 'business', true),
    ('online marketplace', 0.85, 'business', false),
    ('third-party seller', 0.75, 'business', false),
    ('fulfillment', 0.85, 'business', false),
    ('order fulfillment', 0.80, 'business', false),
    ('warehouse management', 0.80, 'technical', false),
    ('inventory management', 0.85, 'business', false),
    ('supply chain', 0.85, 'business', false),
    ('logistics', 0.85, 'business', false),
    ('distribution', 0.80, 'business', false),
    ('last mile delivery', 0.75, 'business', false),
    ('shipping', 0.80, 'business', false),
    ('delivery', 0.80, 'business', false),
    ('returns', 0.75, 'business', false),
    ('reverse logistics', 0.70, 'business', false),
    ('customer experience', 0.90, 'business', true),
    ('customer service', 0.85, 'business', false),
    ('customer support', 0.85, 'business', false),
    ('live chat', 0.80, 'technical', false),
    ('chatbot', 0.80, 'technical', false),
    ('personalization', 0.85, 'technical', false),
    ('recommendation engine', 0.80, 'technical', false),
    ('product recommendations', 0.80, 'business', false),
    ('search', 0.80, 'technical', false),
    ('product search', 0.75, 'technical', false),
    ('product catalog', 0.80, 'business', false),
    ('product information management', 0.75, 'technical', false),
    ('pim', 0.70, 'technical', false),
    ('merchandising', 0.85, 'business', false),
    ('visual merchandising', 0.80, 'business', false),
    ('category management', 0.80, 'business', false),
    ('pricing', 0.85, 'business', false),
    ('dynamic pricing', 0.80, 'technical', false),
    ('price optimization', 0.75, 'technical', false),
    ('promotion', 0.80, 'business', false),
    ('discount', 0.75, 'business', false),
    ('coupon', 0.75, 'business', false),
    ('loyalty program', 0.80, 'business', false),
    ('rewards', 0.75, 'business', false),
    ('membership', 0.75, 'business', false),
    ('subscription', 0.80, 'business', false),
    ('recurring billing', 0.75, 'technical', false),
    ('subscription commerce', 0.75, 'business', false),
    ('social commerce', 0.80, 'business', false)
) AS k(keyword, weight, context, is_primary)
WHERE i.name IN ('Retail', 'E-commerce')
ON CONFLICT (industry_id, keyword) DO NOTHING;

-- =============================================================================
-- LEGAL SERVICES - ADDITIONAL KEYWORDS (50+ keywords)
-- =============================================================================

INSERT INTO industry_keywords (industry_id, keyword, weight, context, is_primary, created_at, updated_at)
SELECT i.id, k.keyword, k.weight, k.context, k.is_primary, NOW(), NOW()
FROM industries i
CROSS JOIN (VALUES
    -- Advanced Legal Terms
    ('legal technology', 0.95, 'technical', true),
    ('legaltech', 0.90, 'technical', true),
    ('legal tech', 0.90, 'technical', true),
    ('contract management', 0.90, 'business', true),
    ('document review', 0.85, 'business', false),
    ('legal research', 0.85, 'business', false),
    ('case management', 0.85, 'business', false),
    ('practice management', 0.85, 'business', false),
    ('legal billing', 0.80, 'business', false),
    ('time tracking', 0.75, 'business', false),
    ('legal analytics', 0.80, 'technical', false),
    ('predictive analytics', 0.75, 'technical', false),
    ('artificial intelligence', 0.85, 'technical', false),
    ('machine learning', 0.80, 'technical', false),
    ('natural language processing', 0.80, 'technical', false),
    ('document automation', 0.85, 'technical', false),
    ('legal document automation', 0.80, 'technical', false),
    ('contract automation', 0.80, 'technical', false),
    ('e-discovery', 0.85, 'technical', false),
    ('electronic discovery', 0.80, 'technical', false),
    ('litigation support', 0.85, 'business', false),
    ('due diligence', 0.85, 'business', false),
    ('regulatory compliance', 0.85, 'business', false),
    ('compliance management', 0.80, 'business', false),
    ('risk assessment', 0.80, 'business', false),
    ('legal risk', 0.75, 'business', false),
    ('intellectual property', 0.90, 'business', true),
    ('patent', 0.85, 'business', false),
    ('trademark', 0.85, 'business', false),
    ('copyright', 0.80, 'business', false),
    ('trade secret', 0.75, 'business', false),
    ('ip', 0.75, 'business', false),
    ('patent prosecution', 0.80, 'business', false),
    ('patent litigation', 0.80, 'business', false),
    ('trademark registration', 0.75, 'business', false),
    ('copyright infringement', 0.75, 'business', false),
    ('corporate law', 0.90, 'business', true),
    ('business law', 0.85, 'business', false),
    ('commercial law', 0.85, 'business', false),
    ('securities law', 0.80, 'business', false),
    ('mergers and acquisitions', 0.85, 'business', false),
    ('m&a', 0.80, 'business', false),
    ('private equity', 0.80, 'business', false),
    ('venture capital', 0.75, 'business', false),
    ('employment law', 0.85, 'business', false),
    ('labor law', 0.80, 'business', false),
    ('immigration law', 0.80, 'business', false),
    ('family law', 0.80, 'business', false),
    ('criminal law', 0.85, 'business', false),
    ('personal injury', 0.80, 'business', false)
) AS k(keyword, weight, context, is_primary)
WHERE i.name = 'Legal Services'
ON CONFLICT (industry_id, keyword) DO NOTHING;

-- =============================================================================
-- MANUFACTURING - ADDITIONAL KEYWORDS (50+ keywords)
-- =============================================================================

INSERT INTO industry_keywords (industry_id, keyword, weight, context, is_primary, created_at, updated_at)
SELECT i.id, k.keyword, k.weight, k.context, k.is_primary, NOW(), NOW()
FROM industries i
CROSS JOIN (VALUES
    -- Advanced Manufacturing Terms
    ('smart manufacturing', 0.95, 'technical', true),
    ('industry 4.0', 0.90, 'technical', true),
    ('industrial iot', 0.90, 'technical', true),
    ('iiot', 0.85, 'technical', false),
    ('automation', 0.95, 'technical', true),
    ('robotics', 0.90, 'technical', true),
    ('industrial robotics', 0.85, 'technical', false),
    ('collaborative robots', 0.80, 'technical', false),
    ('cobots', 0.75, 'technical', false),
    ('artificial intelligence', 0.85, 'technical', false),
    ('machine learning', 0.80, 'technical', false),
    ('predictive maintenance', 0.85, 'technical', false),
    ('preventive maintenance', 0.80, 'technical', false),
    ('condition monitoring', 0.80, 'technical', false),
    ('asset management', 0.80, 'business', false),
    ('equipment management', 0.75, 'business', false),
    ('lean manufacturing', 0.85, 'business', false),
    ('six sigma', 0.80, 'business', false),
    ('continuous improvement', 0.80, 'business', false),
    ('quality control', 0.85, 'business', false),
    ('quality assurance', 0.85, 'business', false),
    ('statistical process control', 0.75, 'technical', false),
    ('spc', 0.70, 'technical', false),
    ('process optimization', 0.80, 'technical', false),
    ('operational excellence', 0.75, 'business', false),
    ('supply chain management', 0.85, 'business', false),
    ('supply chain optimization', 0.80, 'business', false),
    ('procurement', 0.80, 'business', false),
    ('sourcing', 0.75, 'business', false),
    ('vendor management', 0.75, 'business', false),
    ('supplier management', 0.75, 'business', false),
    ('inventory management', 0.85, 'business', false),
    ('warehouse management', 0.80, 'business', false),
    ('materials management', 0.75, 'business', false),
    ('production planning', 0.85, 'business', false),
    ('production scheduling', 0.80, 'business', false),
    ('capacity planning', 0.80, 'business', false),
    ('demand planning', 0.75, 'business', false),
    ('demand forecasting', 0.75, 'business', false),
    ('mrp', 0.75, 'technical', false),
    ('material requirements planning', 0.70, 'technical', false),
    ('erp', 0.80, 'technical', false),
    ('enterprise resource planning', 0.75, 'technical', false),
    ('mes', 0.75, 'technical', false),
    ('manufacturing execution system', 0.70, 'technical', false),
    ('scada', 0.75, 'technical', false),
    ('plc', 0.75, 'technical', false),
    ('programmable logic controller', 0.70, 'technical', false),
    ('hmi', 0.70, 'technical', false),
    ('human machine interface', 0.70, 'technical', false)
) AS k(keyword, weight, context, is_primary)
WHERE i.name = 'Manufacturing'
ON CONFLICT (industry_id, keyword) DO NOTHING;

COMMIT;

-- =============================================================================
-- VALIDATION QUERIES
-- =============================================================================

-- Check total keywords added across all industries
SELECT 
    i.name as industry,
    COUNT(ik.id) as keyword_count,
    COUNT(CASE WHEN ik.is_primary = true THEN 1 END) as primary_keywords,
    COUNT(CASE WHEN ik.context = 'technical' THEN 1 END) as technical_keywords,
    COUNT(CASE WHEN ik.context = 'business' THEN 1 END) as business_keywords,
    AVG(ik.weight) as avg_weight
FROM industries i
LEFT JOIN industry_keywords ik ON i.id = ik.industry_id
WHERE i.name IN ('Technology', 'Healthcare', 'Financial Services', 'Retail', 'E-commerce', 'Legal Services', 'Manufacturing')
GROUP BY i.name, i.id
ORDER BY keyword_count DESC;

-- Overall summary
SELECT 
    'TOTAL KEYWORD EXPANSION' as summary,
    COUNT(*) as total_keywords,
    COUNT(CASE WHEN is_primary = true THEN 1 END) as primary_keywords,
    COUNT(CASE WHEN context = 'technical' THEN 1 END) as technical_keywords,
    COUNT(CASE WHEN context = 'business' THEN 1 END) as business_keywords,
    AVG(weight) as avg_weight
FROM industry_keywords;
