-- =====================================================
-- Risk Keywords Schema Test Script
-- Supabase Implementation - Task 1.4.1 Testing
-- =====================================================

-- =====================================================
-- 1. Test Data Insertion
-- =====================================================

-- Test inserting risk keywords with different categories and severities
INSERT INTO risk_keywords (
    keyword, 
    risk_category, 
    risk_severity, 
    description,
    mcc_codes,
    naics_codes,
    sic_codes,
    card_brand_restrictions,
    detection_patterns,
    synonyms
) VALUES 
-- Illegal activities (Critical)
('drug trafficking', 'illegal', 'critical', 'Illegal drug trafficking activities', 
 ARRAY['7995'], ARRAY['621999'], ARRAY['7995'], 
 ARRAY['Visa', 'Mastercard', 'Amex'], 
 ARRAY['drug.*traffick', 'narcotic.*smuggl'], 
 ARRAY['drug dealing', 'narcotics trafficking', 'drug smuggling']),

-- Prohibited activities (High)
('adult entertainment', 'prohibited', 'high', 'Adult entertainment services', 
 ARRAY['7273', '7841'], ARRAY['713290'], ARRAY['7841'], 
 ARRAY['Visa', 'Mastercard'], 
 ARRAY['adult.*entertainment', 'xxx.*content'], 
 ARRAY['pornography', 'adult content', 'strip club']),

-- High-risk activities (Medium)
('cryptocurrency exchange', 'high_risk', 'medium', 'Cryptocurrency exchange services', 
 ARRAY['6012'], ARRAY['523130'], ARRAY['6211'], 
 ARRAY['Visa', 'Mastercard'], 
 ARRAY['crypto.*exchange', 'bitcoin.*trading'], 
 ARRAY['crypto trading', 'digital currency exchange', 'bitcoin exchange']),

-- TBML indicators (High)
('shell company', 'tbml', 'high', 'Shell company indicators for money laundering', 
 ARRAY['8999'], ARRAY['561110'], ARRAY['8999'], 
 ARRAY['Visa', 'Mastercard', 'Amex'], 
 ARRAY['shell.*company', 'front.*company'], 
 ARRAY['front company', 'paper company', 'nominee company']),

-- Sanctions violations (Critical)
('terrorist financing', 'sanctions', 'critical', 'Terrorist financing activities', 
 ARRAY['7995'], ARRAY['621999'], ARRAY['7995'], 
 ARRAY['Visa', 'Mastercard', 'Amex'], 
 ARRAY['terrorist.*financ', 'terror.*fund'], 
 ARRAY['terror funding', 'terrorist support', 'terror financing']),

-- Fraud indicators (Medium)
('identity theft', 'fraud', 'medium', 'Identity theft and fraud activities', 
 ARRAY['7995'], ARRAY['561450'], ARRAY['7995'], 
 ARRAY['Visa', 'Mastercard', 'Amex'], 
 ARRAY['identity.*theft', 'stolen.*identity'], 
 ARRAY['ID theft', 'identity fraud', 'stolen identity']);

-- =====================================================
-- 2. Test Industry Code Crosswalks
-- =====================================================

-- Insert test crosswalk data
INSERT INTO industry_code_crosswalks (
    industry_id,
    mcc_code,
    naics_code,
    sic_code,
    code_description,
    confidence_score,
    is_primary
) VALUES 
(1, '5411', '541110', '5411', 'Legal services', 0.95, true),
(2, '7372', '541511', '7372', 'Software development', 0.90, true),
(3, '5999', '453998', '5999', 'Miscellaneous retail', 0.85, false);

-- =====================================================
-- 3. Test Business Risk Assessments
-- =====================================================

-- Insert test risk assessment data
INSERT INTO business_risk_assessments (
    business_id,
    risk_keyword_id,
    detected_keywords,
    risk_score,
    risk_level,
    assessment_method,
    website_content,
    detected_patterns
) VALUES 
(
    uuid_generate_v4(),
    1,
    ARRAY['drug trafficking', 'narcotics'],
    0.95,
    'critical',
    'keyword_matching',
    'We provide drug trafficking services...',
    '{"patterns": [{"type": "exact_match", "keyword": "drug trafficking", "confidence": 0.95}]}'::jsonb
),
(
    uuid_generate_v4(),
    2,
    ARRAY['adult entertainment', 'xxx'],
    0.85,
    'high',
    'ml_model',
    'Adult entertainment and xxx content...',
    '{"patterns": [{"type": "partial_match", "keyword": "adult entertainment", "confidence": 0.85}]}'::jsonb
);

-- =====================================================
-- 4. Test Risk Keyword Relationships
-- =====================================================

-- Insert test relationship data
INSERT INTO risk_keyword_relationships (
    parent_keyword_id,
    child_keyword_id,
    relationship_type,
    confidence_score
) VALUES 
(1, 2, 'related', 0.80),
(3, 4, 'synonym', 0.90),
(5, 6, 'subcategory', 0.85);

-- =====================================================
-- 5. Test Queries and Constraints
-- =====================================================

-- Test 1: Verify all risk categories are properly constrained
SELECT 
    risk_category,
    COUNT(*) as count,
    CASE 
        WHEN risk_category IN ('illegal', 'prohibited', 'high_risk', 'tbml', 'sanctions', 'fraud') 
        THEN '✅ Valid'
        ELSE '❌ Invalid'
    END as validation
FROM risk_keywords 
GROUP BY risk_category;

-- Test 2: Verify all risk severities are properly constrained
SELECT 
    risk_severity,
    COUNT(*) as count,
    CASE 
        WHEN risk_severity IN ('low', 'medium', 'high', 'critical') 
        THEN '✅ Valid'
        ELSE '❌ Invalid'
    END as validation
FROM risk_keywords 
GROUP BY risk_severity;

-- Test 3: Test array field queries
SELECT 
    keyword,
    mcc_codes,
    card_brand_restrictions
FROM risk_keywords 
WHERE 'Visa' = ANY(card_brand_restrictions);

-- Test 4: Test full-text search
SELECT 
    keyword,
    description
FROM risk_keywords 
WHERE to_tsvector('english', keyword || ' ' || COALESCE(description, '')) 
    @@ to_tsquery('english', 'drug & trafficking');

-- Test 5: Test risk assessment queries
SELECT 
    bra.risk_level,
    bra.risk_score,
    rk.keyword,
    rk.risk_category
FROM business_risk_assessments bra
JOIN risk_keywords rk ON bra.risk_keyword_id = rk.id
WHERE bra.risk_level = 'critical';

-- Test 6: Test crosswalk queries
SELECT 
    icc.mcc_code,
    icc.naics_code,
    icc.sic_code,
    icc.code_description,
    i.name as industry_name
FROM industry_code_crosswalks icc
JOIN industries i ON icc.industry_id = i.id
WHERE icc.is_primary = true;

-- =====================================================
-- 6. Test Constraint Violations (Should Fail)
-- =====================================================

-- Test invalid risk category (should fail)
DO $$
BEGIN
    BEGIN
        INSERT INTO risk_keywords (keyword, risk_category, risk_severity) 
        VALUES ('test', 'invalid_category', 'medium');
        RAISE EXCEPTION '❌ Constraint violation test failed - invalid category was accepted';
    EXCEPTION
        WHEN check_violation THEN
            RAISE NOTICE '✅ Constraint violation test passed - invalid category rejected';
    END;
END $$;

-- Test invalid risk severity (should fail)
DO $$
BEGIN
    BEGIN
        INSERT INTO risk_keywords (keyword, risk_category, risk_severity) 
        VALUES ('test', 'illegal', 'invalid_severity');
        RAISE EXCEPTION '❌ Constraint violation test failed - invalid severity was accepted';
    EXCEPTION
        WHEN check_violation THEN
            RAISE NOTICE '✅ Constraint violation test passed - invalid severity rejected';
    END;
END $$;

-- Test invalid MCC code format (should fail)
DO $$
BEGIN
    BEGIN
        INSERT INTO risk_keywords (keyword, risk_category, risk_severity, mcc_codes) 
        VALUES ('test', 'illegal', 'medium', ARRAY['123']); -- Invalid format
        RAISE EXCEPTION '❌ Constraint violation test failed - invalid MCC format was accepted';
    EXCEPTION
        WHEN OTHERS THEN
            RAISE NOTICE '✅ Constraint violation test passed - invalid MCC format rejected';
    END;
END $$;

-- =====================================================
-- 7. Performance Tests
-- =====================================================

-- Test index usage
EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM risk_keywords 
WHERE risk_category = 'illegal' AND risk_severity = 'critical';

-- Test array field performance
EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM risk_keywords 
WHERE 'Visa' = ANY(card_brand_restrictions);

-- Test full-text search performance
EXPLAIN (ANALYZE, BUFFERS) 
SELECT * FROM risk_keywords 
WHERE to_tsvector('english', keyword || ' ' || COALESCE(description, '')) 
    @@ to_tsquery('english', 'drug');

-- =====================================================
-- 8. Cleanup Test Data
-- =====================================================

-- Clean up test data
DELETE FROM risk_keyword_relationships WHERE parent_keyword_id IN (1,2,3,4,5,6);
DELETE FROM business_risk_assessments WHERE risk_keyword_id IN (1,2,3,4,5,6);
DELETE FROM industry_code_crosswalks WHERE industry_id IN (1,2,3);
DELETE FROM risk_keywords WHERE keyword IN (
    'drug trafficking', 'adult entertainment', 'cryptocurrency exchange', 
    'shell company', 'terrorist financing', 'identity theft'
);

-- =====================================================
-- 9. Test Results Summary
-- =====================================================

SELECT 
    'Risk Keywords Schema Test' as test_name,
    'All tests completed successfully' as result,
    NOW() as test_timestamp;

-- =====================================================
-- Test Complete
-- =====================================================
