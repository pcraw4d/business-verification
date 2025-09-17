-- =============================================================================
-- FINANCIAL SERVICES KEYWORDS COMPREHENSIVE SCRIPT
-- Task 3.2.6: Add financial services keywords (50+ financial-specific keywords with base weights 0.5-1.0)
-- =============================================================================
-- This script adds comprehensive financial services keywords across all financial
-- services industries to achieve >85% classification accuracy for financial businesses.
-- 
-- Financial Services Industries Covered:
-- 1. Banking (confidence_threshold: 0.80)
-- 2. Insurance (confidence_threshold: 0.75)
-- 3. Investment Services (confidence_threshold: 0.80)
-- Note: Fintech keywords are already covered in technology keywords script
--
-- Total: 200+ comprehensive keywords across 3 financial services industries
-- =============================================================================

-- Start transaction for atomic operation
BEGIN;

-- =============================================================================
-- 1. BANKING KEYWORDS (70+ keywords)
-- =============================================================================
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active, created_at, updated_at)
SELECT i.id, k.keyword, k.weight, true, NOW(), NOW()
FROM industries i, (VALUES
    -- Core Banking Keywords (highest weight)
    ('banking', 1.0000),
    ('bank', 1.0000),
    ('financial institution', 0.9500),
    ('commercial bank', 0.9500),
    ('retail bank', 0.9000),
    ('credit union', 0.9000),
    ('savings bank', 0.8500),
    ('community bank', 0.8500),
    ('investment bank', 0.8000),
    ('merchant bank', 0.8000),
    
    -- Banking Services (high weight)
    ('deposits', 0.9000),
    ('loans', 0.9000),
    ('mortgage', 0.9000),
    ('credit', 0.8500),
    ('lending', 0.8500),
    ('checking', 0.8000),
    ('savings', 0.8000),
    ('cd', 0.7500),
    ('certificate of deposit', 0.7500),
    ('line of credit', 0.7500),
    ('personal loan', 0.7500),
    ('business loan', 0.7500),
    ('auto loan', 0.7000),
    ('home equity', 0.7000),
    ('refinancing', 0.7000),
    
    -- Banking Operations (medium-high weight)
    ('atm', 0.8000),
    ('automated teller', 0.8000),
    ('online banking', 0.8000),
    ('mobile banking', 0.7500),
    ('digital banking', 0.7500),
    ('wire transfer', 0.7000),
    ('ach', 0.7000),
    ('direct deposit', 0.7000),
    ('bill pay', 0.6500),
    ('debit card', 0.6500),
    ('credit card', 0.6500),
    ('cash management', 0.6500),
    ('treasury', 0.6000),
    ('commercial lending', 0.6000),
    ('small business', 0.6000),
    
    -- Banking Products (medium weight)
    ('checking account', 0.7500),
    ('savings account', 0.7500),
    ('money market', 0.7000),
    ('ira', 0.7000),
    ('retirement', 0.6500),
    ('wealth management', 0.6500),
    ('trust services', 0.6000),
    ('estate planning', 0.6000),
    ('financial planning', 0.6000),
    ('investment advisory', 0.6000),
    
    -- Banking Terms (medium weight)
    ('interest rate', 0.7000),
    ('apr', 0.7000),
    ('annual percentage rate', 0.7000),
    ('principal', 0.6500),
    ('collateral', 0.6500),
    ('underwriting', 0.6000),
    ('credit score', 0.6000),
    ('fico', 0.6000),
    ('credit report', 0.6000),
    ('default', 0.5500),
    ('foreclosure', 0.5500),
    ('bankruptcy', 0.5500),
    
    -- Banking Compliance (medium weight)
    ('fdic', 0.7000),
    ('federal deposit insurance', 0.7000),
    ('compliance', 0.6500),
    ('regulatory', 0.6500),
    ('audit', 0.6000),
    ('risk management', 0.6000),
    ('aml', 0.6000),
    ('anti money laundering', 0.6000),
    ('kyc', 0.6000),
    ('know your customer', 0.6000),
    ('bsa', 0.5500),
    ('bank secrecy act', 0.5500),
    
    -- Banking Technology (lower weight)
    ('core banking', 0.6000),
    ('fiserv', 0.5000),
    ('jack henry', 0.5000),
    ('fis', 0.5000),
    ('teller', 0.5500),
    ('branch', 0.5500),
    ('customer service', 0.5000)
) AS k(keyword, weight)
WHERE i.name = 'Banking' AND i.is_active = true
ON CONFLICT (industry_id, keyword) DO UPDATE SET
    weight = EXCLUDED.weight,
    updated_at = NOW();

-- =============================================================================
-- 2. INSURANCE KEYWORDS (70+ keywords)
-- =============================================================================
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active, created_at, updated_at)
SELECT i.id, k.keyword, k.weight, true, NOW(), NOW()
FROM industries i, (VALUES
    -- Core Insurance Keywords (highest weight)
    ('insurance', 1.0000),
    ('insurer', 0.9500),
    ('insurance company', 0.9500),
    ('insurance agency', 0.9000),
    ('insurance broker', 0.9000),
    ('underwriter', 0.8500),
    ('actuary', 0.8000),
    ('claims', 0.8000),
    ('policy', 0.8000),
    ('premium', 0.8000),
    
    -- Life Insurance (high weight)
    ('life insurance', 0.9500),
    ('term life', 0.9000),
    ('whole life', 0.9000),
    ('universal life', 0.8500),
    ('variable life', 0.8000),
    ('annuity', 0.8000),
    ('death benefit', 0.7500),
    ('beneficiary', 0.7000),
    ('cash value', 0.7000),
    ('surrender value', 0.6500),
    
    -- Health Insurance (high weight)
    ('health insurance', 0.9500),
    ('medical insurance', 0.9000),
    ('health plan', 0.9000),
    ('hmo', 0.8500),
    ('ppo', 0.8500),
    ('health maintenance organization', 0.8000),
    ('preferred provider', 0.8000),
    ('deductible', 0.8000),
    ('copay', 0.7500),
    ('copayment', 0.7500),
    ('coinsurance', 0.7000),
    ('out of pocket', 0.7000),
    ('medicare', 0.7500),
    ('medicaid', 0.7500),
    ('aca', 0.7000),
    ('affordable care act', 0.7000),
    
    -- Property & Casualty (high weight)
    ('property insurance', 0.9000),
    ('casualty insurance', 0.9000),
    ('homeowners insurance', 0.9000),
    ('auto insurance', 0.9000),
    ('car insurance', 0.9000),
    ('liability insurance', 0.8500),
    ('commercial insurance', 0.8500),
    ('business insurance', 0.8500),
    ('workers compensation', 0.8000),
    ('workers comp', 0.8000),
    ('general liability', 0.8000),
    ('professional liability', 0.7500),
    ('errors omissions', 0.7500),
    ('e&o', 0.7500),
    ('directors officers', 0.7000),
    ('d&o', 0.7000),
    
    -- Insurance Operations (medium-high weight)
    ('underwriting', 0.8000),
    ('actuarial', 0.7500),
    ('risk assessment', 0.7500),
    ('loss ratio', 0.7000),
    ('combined ratio', 0.7000),
    ('reserves', 0.7000),
    ('reinsurance', 0.7000),
    ('catastrophe', 0.6500),
    ('cat', 0.6500),
    ('adjuster', 0.6500),
    ('claims adjuster', 0.6500),
    ('settlement', 0.6000),
    ('subrogation', 0.6000),
    
    -- Insurance Products (medium weight)
    ('umbrella policy', 0.7000),
    ('flood insurance', 0.7000),
    ('earthquake insurance', 0.6500),
    ('disability insurance', 0.7000),
    ('long term care', 0.7000),
    ('travel insurance', 0.6000),
    ('pet insurance', 0.6000),
    ('cyber insurance', 0.6500),
    ('data breach', 0.6000),
    
    -- Insurance Terms (medium weight)
    ('coverage', 0.8000),
    ('limit', 0.7500),
    ('exclusion', 0.7000),
    ('endorsement', 0.6500),
    ('rider', 0.6500),
    ('grace period', 0.6000),
    ('lapse', 0.6000),
    ('reinstatement', 0.5500),
    ('contestability', 0.5500),
    
    -- Insurance Technology (lower weight)
    ('policy management', 0.6000),
    ('claims processing', 0.6000),
    ('billing', 0.5500),
    ('customer portal', 0.5000),
    ('mobile app', 0.5000)
) AS k(keyword, weight)
WHERE i.name = 'Insurance' AND i.is_active = true
ON CONFLICT (industry_id, keyword) DO UPDATE SET
    weight = EXCLUDED.weight,
    updated_at = NOW();

-- =============================================================================
-- 3. INVESTMENT SERVICES KEYWORDS (70+ keywords)
-- =============================================================================
INSERT INTO industry_keywords (industry_id, keyword, weight, is_active, created_at, updated_at)
SELECT i.id, k.keyword, k.weight, true, NOW(), NOW()
FROM industries i, (VALUES
    -- Core Investment Keywords (highest weight)
    ('investment', 1.0000),
    ('investment services', 0.9500),
    ('investment advisory', 0.9500),
    ('wealth management', 0.9500),
    ('asset management', 0.9000),
    ('portfolio management', 0.9000),
    ('financial advisor', 0.9000),
    ('financial planner', 0.9000),
    ('investment advisor', 0.9000),
    ('broker', 0.8500),
    ('brokerage', 0.8500),
    
    -- Investment Products (high weight)
    ('stocks', 0.9000),
    ('bonds', 0.9000),
    ('mutual funds', 0.9000),
    ('etf', 0.8500),
    ('exchange traded fund', 0.8500),
    ('index fund', 0.8000),
    ('hedge fund', 0.8000),
    ('private equity', 0.8000),
    ('venture capital', 0.8000),
    ('real estate investment', 0.7500),
    ('reit', 0.7500),
    ('real estate investment trust', 0.7500),
    ('commodities', 0.7000),
    ('futures', 0.7000),
    ('options', 0.7000),
    ('derivatives', 0.7000),
    
    -- Investment Services (high weight)
    ('financial planning', 0.9000),
    ('retirement planning', 0.8500),
    ('estate planning', 0.8500),
    ('tax planning', 0.8000),
    ('college planning', 0.7500),
    ('529 plan', 0.7500),
    ('ira', 0.8000),
    ('roth ira', 0.8000),
    ('401k', 0.8000),
    ('403b', 0.7500),
    ('pension', 0.7500),
    ('rollover', 0.7000),
    
    -- Investment Analysis (medium-high weight)
    ('research', 0.8000),
    ('analysis', 0.8000),
    ('fundamental analysis', 0.7500),
    ('technical analysis', 0.7500),
    ('valuation', 0.7500),
    ('due diligence', 0.7000),
    ('risk assessment', 0.7000),
    ('diversification', 0.7000),
    ('asset allocation', 0.7000),
    ('rebalancing', 0.6500),
    ('tax loss harvesting', 0.6500),
    
    -- Investment Terms (medium weight)
    ('return', 0.8000),
    ('yield', 0.7500),
    ('dividend', 0.7500),
    ('capital gains', 0.7500),
    ('volatility', 0.7000),
    ('beta', 0.6500),
    ('alpha', 0.6500),
    ('sharpe ratio', 0.6000),
    ('expense ratio', 0.6500),
    ('load', 0.6000),
    ('no load', 0.6000),
    ('front end load', 0.5500),
    ('back end load', 0.5500),
    
    -- Investment Regulations (medium weight)
    ('sec', 0.7000),
    ('securities exchange commission', 0.7000),
    ('finra', 0.7000),
    ('financial industry regulatory authority', 0.7000),
    ('fiduciary', 0.7500),
    ('suitability', 0.7000),
    ('disclosure', 0.6500),
    ('prospectus', 0.6500),
    ('advisory agreement', 0.6000),
    ('custody', 0.6000),
    
    -- Investment Technology (lower weight)
    ('robo advisor', 0.7000),
    ('automated investing', 0.6500),
    ('trading platform', 0.6500),
    ('portfolio tracker', 0.6000),
    ('financial software', 0.5500),
    ('client portal', 0.5000),
    ('mobile trading', 0.5000)
) AS k(keyword, weight)
WHERE i.name = 'Investment Services' AND i.is_active = true
ON CONFLICT (industry_id, keyword) DO UPDATE SET
    weight = EXCLUDED.weight,
    updated_at = NOW();

-- =============================================================================
-- VERIFICATION QUERIES
-- =============================================================================

-- Verify all financial services keywords were added
DO $$
DECLARE
    banking_count INTEGER;
    insurance_count INTEGER;
    investment_count INTEGER;
    expected_count INTEGER := 200; -- 200+ keywords across all financial services industries
BEGIN
    -- Count Banking keywords
    SELECT COUNT(*) INTO banking_count
    FROM industry_keywords ik
    JOIN industries i ON ik.industry_id = i.id
    WHERE i.name = 'Banking' AND i.is_active = true AND ik.is_active = true;
    
    -- Count Insurance keywords
    SELECT COUNT(*) INTO insurance_count
    FROM industry_keywords ik
    JOIN industries i ON ik.industry_id = i.id
    WHERE i.name = 'Insurance' AND i.is_active = true AND ik.is_active = true;
    
    -- Count Investment Services keywords
    SELECT COUNT(*) INTO investment_count
    FROM industry_keywords ik
    JOIN industries i ON ik.industry_id = i.id
    WHERE i.name = 'Investment Services' AND i.is_active = true AND ik.is_active = true;
    
    RAISE NOTICE 'Financial Services Keywords Added:';
    RAISE NOTICE 'Banking: % keywords', banking_count;
    RAISE NOTICE 'Insurance: % keywords', insurance_count;
    RAISE NOTICE 'Investment Services: % keywords', investment_count;
    RAISE NOTICE 'Total: % keywords', (banking_count + insurance_count + investment_count);
    
    IF (banking_count + insurance_count + investment_count) >= expected_count THEN
        RAISE NOTICE 'SUCCESS: Financial services keywords added successfully';
    ELSE
        RAISE NOTICE 'WARNING: Expected % keywords, but found %', expected_count, (banking_count + insurance_count + investment_count);
    END IF;
END $$;

-- Display keyword summary by industry
SELECT 
    'FINANCIAL SERVICES KEYWORD SUMMARY' as summary_type,
    '' as spacer;

SELECT 
    i.name as industry_name,
    i.confidence_threshold,
    COUNT(ik.id) as keyword_count,
    ROUND(MIN(ik.weight), 3) as min_weight,
    ROUND(MAX(ik.weight), 3) as max_weight,
    ROUND(AVG(ik.weight), 3) as avg_weight
FROM industries i
LEFT JOIN industry_keywords ik ON i.id = ik.industry_id AND ik.is_active = true
WHERE i.name IN ('Banking', 'Insurance', 'Investment Services') 
AND i.is_active = true
GROUP BY i.id, i.name, i.confidence_threshold
ORDER BY i.name;

-- Display sample keywords for each industry
SELECT 
    'SAMPLE BANKING KEYWORDS' as summary_type,
    '' as spacer;

SELECT 
    keyword,
    weight,
    CASE 
        WHEN weight >= 0.90 THEN 'Very High'
        WHEN weight >= 0.80 THEN 'High'
        WHEN weight >= 0.70 THEN 'Medium-High'
        WHEN weight >= 0.60 THEN 'Medium'
        ELSE 'Low'
    END as weight_category
FROM industry_keywords ik
JOIN industries i ON ik.industry_id = i.id
WHERE i.name = 'Banking' AND i.is_active = true AND ik.is_active = true
ORDER BY weight DESC
LIMIT 10;

SELECT 
    'SAMPLE INSURANCE KEYWORDS' as summary_type,
    '' as spacer;

SELECT 
    keyword,
    weight,
    CASE 
        WHEN weight >= 0.90 THEN 'Very High'
        WHEN weight >= 0.80 THEN 'High'
        WHEN weight >= 0.70 THEN 'Medium-High'
        WHEN weight >= 0.60 THEN 'Medium'
        ELSE 'Low'
    END as weight_category
FROM industry_keywords ik
JOIN industries i ON ik.industry_id = i.id
WHERE i.name = 'Insurance' AND i.is_active = true AND ik.is_active = true
ORDER BY weight DESC
LIMIT 10;

SELECT 
    'SAMPLE INVESTMENT SERVICES KEYWORDS' as summary_type,
    '' as spacer;

SELECT 
    keyword,
    weight,
    CASE 
        WHEN weight >= 0.90 THEN 'Very High'
        WHEN weight >= 0.80 THEN 'High'
        WHEN weight >= 0.70 THEN 'Medium-High'
        WHEN weight >= 0.60 THEN 'Medium'
        ELSE 'Low'
    END as weight_category
FROM industry_keywords ik
JOIN industries i ON ik.industry_id = i.id
WHERE i.name = 'Investment Services' AND i.is_active = true AND ik.is_active = true
ORDER BY weight DESC
LIMIT 10;

COMMIT;

-- =============================================================================
-- COMPLETION MESSAGE
-- =============================================================================
DO $$
BEGIN
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'FINANCIAL SERVICES KEYWORDS COMPREHENSIVE SCRIPT COMPLETED';
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'Task 3.2.6: Financial Services Keywords Added Successfully';
    RAISE NOTICE 'Industries covered: Banking, Insurance, Investment Services';
    RAISE NOTICE 'Total keywords added: 200+ comprehensive financial keywords';
    RAISE NOTICE 'Weight range: 0.5000-1.0000 as specified in plan';
    RAISE NOTICE 'Status: Ready for testing and validation';
    RAISE NOTICE 'Next: Test financial services classification accuracy';
    RAISE NOTICE '=============================================================================';
END $$;
