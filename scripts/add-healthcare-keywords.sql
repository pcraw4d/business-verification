-- =============================================================================
-- COMPREHENSIVE HEALTHCARE KEYWORDS SCRIPT
-- =============================================================================
-- This script adds comprehensive healthcare keywords for all 4 healthcare
-- industries to achieve >85% classification accuracy for healthcare businesses.
-- 
-- Healthcare Industries Covered:
-- 1. Medical Practices (50+ keywords)
-- 2. Healthcare Services (50+ keywords) 
-- 3. Mental Health (50+ keywords)
-- 4. Healthcare Technology (50+ keywords)
--
-- Total: 200+ healthcare-specific keywords with base weights 0.50-1.00
-- =============================================================================

-- Start transaction for atomic operation
BEGIN;

-- =============================================================================
-- 1. MEDICAL PRACTICES KEYWORDS (50+ keywords)
-- =============================================================================
-- Base weights: 0.50-1.00 (higher weights for more specific medical terms)
-- Coverage: Family medicine, specialists, clinical services, medical procedures

INSERT INTO keyword_weights (industry_id, keyword, base_weight, is_active, created_at, updated_at)
SELECT 
    i.id,
    keyword_data.keyword,
    keyword_data.base_weight,
    true,
    NOW(),
    NOW()
FROM industries i
CROSS JOIN (
    VALUES
    -- Core Medical Practice Terms (Weight: 0.90-1.00)
    ('medical practice', 1.00),
    ('family medicine', 0.95),
    ('primary care', 0.95),
    ('internal medicine', 0.90),
    ('pediatrics', 0.90),
    ('obstetrics', 0.90),
    ('gynecology', 0.90),
    ('cardiology', 0.90),
    ('dermatology', 0.90),
    ('orthopedics', 0.90),
    ('neurology', 0.90),
    ('oncology', 0.90),
    ('urology', 0.90),
    ('ophthalmology', 0.90),
    ('otolaryngology', 0.90),
    ('psychiatry', 0.90),
    ('anesthesiology', 0.90),
    ('radiology', 0.90),
    ('pathology', 0.90),
    ('emergency medicine', 0.90),
    
    -- Medical Services & Procedures (Weight: 0.80-0.90)
    ('medical consultation', 0.85),
    ('physical examination', 0.85),
    ('diagnostic testing', 0.85),
    ('medical diagnosis', 0.85),
    ('treatment planning', 0.85),
    ('preventive care', 0.85),
    ('chronic disease management', 0.85),
    ('vaccination', 0.80),
    ('immunization', 0.80),
    ('health screening', 0.80),
    ('medical checkup', 0.80),
    ('annual physical', 0.80),
    ('wellness visit', 0.80),
    ('sports physical', 0.80),
    ('pre-employment physical', 0.80),
    
    -- Medical Professionals (Weight: 0.70-0.85)
    ('physician', 0.85),
    ('doctor', 0.85),
    ('medical doctor', 0.85),
    ('family physician', 0.85),
    ('specialist', 0.80),
    ('medical specialist', 0.80),
    ('surgeon', 0.80),
    ('nurse practitioner', 0.75),
    ('physician assistant', 0.75),
    ('medical assistant', 0.70),
    ('nurse', 0.70),
    ('registered nurse', 0.70),
    
    -- Medical Facilities & Equipment (Weight: 0.60-0.80)
    ('medical office', 0.80),
    ('clinic', 0.80),
    ('medical clinic', 0.80),
    ('family practice', 0.80),
    ('medical center', 0.75),
    ('health center', 0.75),
    ('medical equipment', 0.70),
    ('diagnostic equipment', 0.70),
    ('medical instruments', 0.70),
    ('examination room', 0.65),
    ('treatment room', 0.65),
    ('waiting room', 0.60),
    
    -- Medical Conditions & Treatments (Weight: 0.50-0.75)
    ('hypertension', 0.75),
    ('diabetes', 0.75),
    ('asthma', 0.70),
    ('arthritis', 0.70),
    ('depression', 0.70),
    ('anxiety', 0.70),
    ('high blood pressure', 0.70),
    ('heart disease', 0.70),
    ('cancer screening', 0.70),
    ('mammography', 0.70),
    ('colonoscopy', 0.70),
    ('blood test', 0.65),
    ('urine test', 0.65),
    ('x-ray', 0.65),
    ('ultrasound', 0.65),
    ('mri', 0.65),
    ('ct scan', 0.65),
    ('prescription', 0.60),
    ('medication', 0.60),
    ('medical treatment', 0.60),
    ('therapy', 0.60),
    ('rehabilitation', 0.60),
    ('medical care', 0.55),
    ('healthcare', 0.55),
    ('patient care', 0.55),
    ('medical services', 0.50)
) AS keyword_data(keyword, base_weight)
WHERE i.name = 'Medical Practices'
ON CONFLICT (industry_id, keyword) DO UPDATE SET
    base_weight = EXCLUDED.base_weight,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();

-- =============================================================================
-- 2. HEALTHCARE SERVICES KEYWORDS (50+ keywords)
-- =============================================================================
-- Base weights: 0.50-1.00 (higher weights for hospital and facility-specific terms)
-- Coverage: Hospitals, clinics, medical facilities, healthcare systems

INSERT INTO keyword_weights (industry_id, keyword, base_weight, is_active, created_at, updated_at)
SELECT 
    i.id,
    keyword_data.keyword,
    keyword_data.base_weight,
    true,
    NOW(),
    NOW()
FROM industries i
CROSS JOIN (
    VALUES
    -- Core Healthcare Services Terms (Weight: 0.90-1.00)
    ('healthcare services', 1.00),
    ('hospital', 0.95),
    ('medical center', 0.95),
    ('healthcare facility', 0.95),
    ('medical facility', 0.95),
    ('healthcare system', 0.90),
    ('hospital system', 0.90),
    ('health network', 0.90),
    ('medical network', 0.90),
    ('healthcare organization', 0.90),
    
    -- Hospital Types & Departments (Weight: 0.80-0.90)
    ('general hospital', 0.90),
    ('community hospital', 0.90),
    ('teaching hospital', 0.90),
    ('regional medical center', 0.90),
    ('trauma center', 0.90),
    ('emergency department', 0.85),
    ('intensive care unit', 0.85),
    ('icu', 0.85),
    ('surgery center', 0.85),
    ('outpatient center', 0.85),
    ('ambulatory care', 0.85),
    ('urgent care', 0.85),
    ('emergency care', 0.85),
    ('critical care', 0.85),
    ('cardiac care', 0.85),
    ('cancer center', 0.85),
    ('oncology center', 0.85),
    ('rehabilitation center', 0.85),
    ('physical therapy', 0.85),
    ('occupational therapy', 0.85),
    
    -- Healthcare Administration (Weight: 0.70-0.85)
    ('healthcare administration', 0.85),
    ('hospital administration', 0.85),
    ('medical administration', 0.85),
    ('healthcare management', 0.80),
    ('hospital management', 0.80),
    ('healthcare operations', 0.80),
    ('patient services', 0.80),
    ('patient care services', 0.80),
    ('healthcare delivery', 0.80),
    ('medical services delivery', 0.80),
    ('healthcare coordination', 0.75),
    ('care coordination', 0.75),
    ('case management', 0.75),
    ('discharge planning', 0.75),
    ('patient navigation', 0.75),
    
    -- Healthcare Staff & Personnel (Weight: 0.60-0.80)
    ('healthcare staff', 0.80),
    ('medical staff', 0.80),
    ('hospital staff', 0.80),
    ('healthcare professionals', 0.80),
    ('medical professionals', 0.80),
    ('healthcare workers', 0.75),
    ('medical workers', 0.75),
    ('healthcare team', 0.75),
    ('medical team', 0.75),
    ('clinical staff', 0.75),
    ('nursing staff', 0.75),
    ('support staff', 0.70),
    ('administrative staff', 0.70),
    ('healthcare personnel', 0.70),
    ('medical personnel', 0.70),
    
    -- Healthcare Infrastructure (Weight: 0.50-0.75)
    ('healthcare infrastructure', 0.75),
    ('medical infrastructure', 0.75),
    ('hospital infrastructure', 0.75),
    ('healthcare technology', 0.70),
    ('medical technology', 0.70),
    ('health information systems', 0.70),
    ('electronic health records', 0.70),
    ('ehr', 0.70),
    ('healthcare data', 0.65),
    ('medical data', 0.65),
    ('patient data', 0.65),
    ('healthcare analytics', 0.65),
    ('medical analytics', 0.65),
    ('healthcare quality', 0.60),
    ('medical quality', 0.60),
    ('patient safety', 0.60),
    ('healthcare safety', 0.60),
    ('medical safety', 0.60),
    ('healthcare compliance', 0.60),
    ('medical compliance', 0.60),
    ('healthcare accreditation', 0.60),
    ('medical accreditation', 0.60),
    ('healthcare standards', 0.55),
    ('medical standards', 0.55),
    ('healthcare regulations', 0.55),
    ('medical regulations', 0.55),
    ('healthcare policy', 0.50),
    ('medical policy', 0.50)
) AS keyword_data(keyword, base_weight)
WHERE i.name = 'Healthcare Services'
ON CONFLICT (industry_id, keyword) DO UPDATE SET
    base_weight = EXCLUDED.base_weight,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();

-- =============================================================================
-- 3. MENTAL HEALTH KEYWORDS (50+ keywords)
-- =============================================================================
-- Base weights: 0.50-1.00 (higher weights for mental health-specific terms)
-- Coverage: Counseling, therapy, psychological services, mental wellness

INSERT INTO keyword_weights (industry_id, keyword, base_weight, is_active, created_at, updated_at)
SELECT 
    i.id,
    keyword_data.keyword,
    keyword_data.base_weight,
    true,
    NOW(),
    NOW()
FROM industries i
CROSS JOIN (
    VALUES
    -- Core Mental Health Terms (Weight: 0.90-1.00)
    ('mental health', 1.00),
    ('mental health services', 0.95),
    ('psychological services', 0.95),
    ('counseling services', 0.95),
    ('therapy services', 0.95),
    ('psychiatric services', 0.90),
    ('behavioral health', 0.90),
    ('mental wellness', 0.90),
    ('emotional wellness', 0.90),
    ('psychological wellness', 0.90),
    
    -- Mental Health Professionals (Weight: 0.80-0.90)
    ('psychologist', 0.90),
    ('psychiatrist', 0.90),
    ('therapist', 0.90),
    ('counselor', 0.90),
    ('mental health counselor', 0.90),
    ('licensed counselor', 0.90),
    ('clinical psychologist', 0.90),
    ('clinical social worker', 0.85),
    ('psychiatric nurse', 0.85),
    ('mental health nurse', 0.85),
    ('behavioral therapist', 0.85),
    ('cognitive therapist', 0.85),
    ('family therapist', 0.85),
    ('marriage counselor', 0.85),
    ('addiction counselor', 0.85),
    ('substance abuse counselor', 0.85),
    ('trauma therapist', 0.85),
    ('grief counselor', 0.85),
    ('child psychologist', 0.85),
    ('adolescent therapist', 0.85),
    
    -- Therapy Types & Approaches (Weight: 0.70-0.85)
    ('individual therapy', 0.85),
    ('group therapy', 0.85),
    ('family therapy', 0.85),
    ('couples therapy', 0.85),
    ('marriage therapy', 0.85),
    ('cognitive behavioral therapy', 0.85),
    ('cbt', 0.85),
    ('dialectical behavior therapy', 0.85),
    ('dbt', 0.85),
    ('psychotherapy', 0.80),
    ('talk therapy', 0.80),
    ('behavioral therapy', 0.80),
    ('cognitive therapy', 0.80),
    ('humanistic therapy', 0.80),
    ('psychodynamic therapy', 0.80),
    ('art therapy', 0.75),
    ('music therapy', 0.75),
    ('play therapy', 0.75),
    ('occupational therapy', 0.75),
    ('recreational therapy', 0.75),
    
    -- Mental Health Conditions (Weight: 0.60-0.80)
    ('depression', 0.80),
    ('anxiety', 0.80),
    ('anxiety disorder', 0.80),
    ('panic disorder', 0.80),
    ('bipolar disorder', 0.80),
    ('ptsd', 0.80),
    ('post-traumatic stress disorder', 0.80),
    ('trauma', 0.80),
    ('stress', 0.75),
    ('stress management', 0.75),
    ('anger management', 0.75),
    ('substance abuse', 0.75),
    ('addiction', 0.75),
    ('eating disorder', 0.75),
    ('adhd', 0.75),
    ('attention deficit disorder', 0.75),
    ('autism', 0.75),
    ('autism spectrum', 0.75),
    ('schizophrenia', 0.75),
    ('personality disorder', 0.75),
    ('mood disorder', 0.70),
    ('sleep disorder', 0.70),
    ('grief', 0.70),
    ('bereavement', 0.70),
    ('relationship issues', 0.70),
    ('family issues', 0.70),
    ('workplace stress', 0.70),
    ('career counseling', 0.70),
    ('life coaching', 0.70),
    ('personal development', 0.70),
    
    -- Mental Health Facilities & Programs (Weight: 0.50-0.75)
    ('mental health clinic', 0.75),
    ('counseling center', 0.75),
    ('therapy center', 0.75),
    ('psychological center', 0.75),
    ('behavioral health center', 0.75),
    ('mental health program', 0.70),
    ('counseling program', 0.70),
    ('therapy program', 0.70),
    ('outpatient program', 0.70),
    ('intensive outpatient', 0.70),
    ('partial hospitalization', 0.70),
    ('day treatment', 0.70),
    ('residential treatment', 0.70),
    ('inpatient treatment', 0.70),
    ('crisis intervention', 0.70),
    ('emergency services', 0.70),
    ('24-hour crisis line', 0.70),
    ('suicide prevention', 0.70),
    ('mental health support', 0.65),
    ('peer support', 0.65),
    ('support group', 0.65),
    ('mental health education', 0.60),
    ('mental health awareness', 0.60),
    ('mental health advocacy', 0.60),
    ('mental health resources', 0.60),
    ('mental health information', 0.55),
    ('mental health screening', 0.55),
    ('mental health assessment', 0.55),
    ('psychological evaluation', 0.55),
    ('mental health treatment', 0.50)
) AS keyword_data(keyword, base_weight)
WHERE i.name = 'Mental Health'
ON CONFLICT (industry_id, keyword) DO UPDATE SET
    base_weight = EXCLUDED.base_weight,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();

-- =============================================================================
-- 4. HEALTHCARE TECHNOLOGY KEYWORDS (50+ keywords)
-- =============================================================================
-- Base weights: 0.50-1.00 (higher weights for health tech-specific terms)
-- Coverage: Medical devices, health IT, digital health, health innovation

INSERT INTO keyword_weights (industry_id, keyword, base_weight, is_active, created_at, updated_at)
SELECT 
    i.id,
    keyword_data.keyword,
    keyword_data.base_weight,
    true,
    NOW(),
    NOW()
FROM industries i
CROSS JOIN (
    VALUES
    -- Core Healthcare Technology Terms (Weight: 0.90-1.00)
    ('healthcare technology', 1.00),
    ('health technology', 0.95),
    ('medical technology', 0.95),
    ('health tech', 0.95),
    ('medtech', 0.95),
    ('digital health', 0.90),
    ('healthcare innovation', 0.90),
    ('medical innovation', 0.90),
    ('healthcare digitalization', 0.90),
    ('healthcare automation', 0.90),
    
    -- Medical Devices & Equipment (Weight: 0.80-0.90)
    ('medical devices', 0.90),
    ('medical equipment', 0.90),
    ('diagnostic devices', 0.90),
    ('therapeutic devices', 0.90),
    ('monitoring devices', 0.90),
    ('wearable devices', 0.85),
    ('health wearables', 0.85),
    ('fitness trackers', 0.85),
    ('smart health devices', 0.85),
    ('iot health devices', 0.85),
    ('connected health', 0.85),
    ('telemedicine devices', 0.85),
    ('remote monitoring', 0.85),
    ('home health monitoring', 0.85),
    ('patient monitoring', 0.85),
    ('vital signs monitoring', 0.85),
    ('continuous monitoring', 0.85),
    ('real-time monitoring', 0.85),
    ('health sensors', 0.80),
    ('biometric sensors', 0.80),
    ('health data collection', 0.80),
    ('medical data collection', 0.80),
    
    -- Health Information Technology (Weight: 0.70-0.85)
    ('health information technology', 0.85),
    ('health it', 0.85),
    ('healthcare it', 0.85),
    ('medical it', 0.85),
    ('electronic health records', 0.85),
    ('ehr', 0.85),
    ('electronic medical records', 0.85),
    ('emr', 0.85),
    ('health information systems', 0.85),
    ('medical information systems', 0.85),
    ('healthcare data systems', 0.85),
    ('medical data systems', 0.85),
    ('health data management', 0.80),
    ('medical data management', 0.80),
    ('healthcare analytics', 0.80),
    ('medical analytics', 0.80),
    ('health data analytics', 0.80),
    ('medical data analytics', 0.80),
    ('clinical decision support', 0.80),
    ('healthcare decision support', 0.80),
    ('medical decision support', 0.80),
    ('healthcare workflow', 0.75),
    ('medical workflow', 0.75),
    ('clinical workflow', 0.75),
    ('healthcare automation', 0.75),
    ('medical automation', 0.75),
    ('clinical automation', 0.75),
    
    -- Digital Health & Telehealth (Weight: 0.60-0.80)
    ('telehealth', 0.80),
    ('telemedicine', 0.80),
    ('virtual care', 0.80),
    ('remote care', 0.80),
    ('digital care', 0.80),
    ('online health', 0.80),
    ('mobile health', 0.80),
    ('mhealth', 0.80),
    ('health apps', 0.75),
    ('medical apps', 0.75),
    ('healthcare apps', 0.75),
    ('wellness apps', 0.75),
    ('fitness apps', 0.75),
    ('health monitoring apps', 0.75),
    ('patient engagement', 0.75),
    ('health engagement', 0.75),
    ('digital patient engagement', 0.75),
    ('health portals', 0.70),
    ('patient portals', 0.70),
    ('healthcare portals', 0.70),
    ('medical portals', 0.70),
    ('online health services', 0.70),
    ('digital health services', 0.70),
    ('virtual health services', 0.70),
    ('remote health services', 0.70),
    ('healthcare communication', 0.70),
    ('medical communication', 0.70),
    ('patient communication', 0.70),
    ('healthcare messaging', 0.70),
    ('medical messaging', 0.70),
    
    -- Health Data & AI (Weight: 0.50-0.75)
    ('health data', 0.75),
    ('medical data', 0.75),
    ('healthcare data', 0.75),
    ('clinical data', 0.75),
    ('patient data', 0.75),
    ('health big data', 0.70),
    ('medical big data', 0.70),
    ('healthcare big data', 0.70),
    ('health data science', 0.70),
    ('medical data science', 0.70),
    ('healthcare data science', 0.70),
    ('health machine learning', 0.70),
    ('medical machine learning', 0.70),
    ('healthcare machine learning', 0.70),
    ('health ai', 0.70),
    ('medical ai', 0.70),
    ('healthcare ai', 0.70),
    ('artificial intelligence health', 0.70),
    ('ai healthcare', 0.70),
    ('health predictive analytics', 0.65),
    ('medical predictive analytics', 0.65),
    ('healthcare predictive analytics', 0.65),
    ('health insights', 0.65),
    ('medical insights', 0.65),
    ('healthcare insights', 0.65),
    ('health intelligence', 0.60),
    ('medical intelligence', 0.60),
    ('healthcare intelligence', 0.60),
    ('health innovation', 0.60),
    ('medical innovation', 0.60),
    ('healthcare innovation', 0.60),
    ('health technology solutions', 0.55),
    ('medical technology solutions', 0.55),
    ('healthcare technology solutions', 0.55),
    ('health tech solutions', 0.55),
    ('medical tech solutions', 0.55),
    ('healthcare tech solutions', 0.55),
    ('digital health solutions', 0.50),
    ('healthcare digital solutions', 0.50),
    ('medical digital solutions', 0.50)
) AS keyword_data(keyword, base_weight)
WHERE i.name = 'Healthcare Technology'
ON CONFLICT (industry_id, keyword) DO UPDATE SET
    base_weight = EXCLUDED.base_weight,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();

-- =============================================================================
-- VERIFICATION QUERIES
-- =============================================================================

-- Verify all healthcare keywords were added
DO $$
DECLARE
    medical_practices_count INTEGER;
    healthcare_services_count INTEGER;
    mental_health_count INTEGER;
    healthcare_technology_count INTEGER;
    total_healthcare_keywords INTEGER;
BEGIN
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'HEALTHCARE KEYWORDS VERIFICATION';
    RAISE NOTICE '=============================================================================';
    
    -- Count keywords for each healthcare industry
    SELECT COUNT(*) INTO medical_practices_count
    FROM keyword_weights kw
    JOIN industries i ON kw.industry_id = i.id
    WHERE i.name = 'Medical Practices' AND kw.is_active = true;
    
    SELECT COUNT(*) INTO healthcare_services_count
    FROM keyword_weights kw
    JOIN industries i ON kw.industry_id = i.id
    WHERE i.name = 'Healthcare Services' AND kw.is_active = true;
    
    SELECT COUNT(*) INTO mental_health_count
    FROM keyword_weights kw
    JOIN industries i ON kw.industry_id = i.id
    WHERE i.name = 'Mental Health' AND kw.is_active = true;
    
    SELECT COUNT(*) INTO healthcare_technology_count
    FROM keyword_weights kw
    JOIN industries i ON kw.industry_id = i.id
    WHERE i.name = 'Healthcare Technology' AND kw.is_active = true;
    
    total_healthcare_keywords := medical_practices_count + healthcare_services_count + mental_health_count + healthcare_technology_count;
    
    -- Display results
    RAISE NOTICE 'Medical Practices Keywords: %', medical_practices_count;
    RAISE NOTICE 'Healthcare Services Keywords: %', healthcare_services_count;
    RAISE NOTICE 'Mental Health Keywords: %', mental_health_count;
    RAISE NOTICE 'Healthcare Technology Keywords: %', healthcare_technology_count;
    RAISE NOTICE 'Total Healthcare Keywords: %', total_healthcare_keywords;
    
    -- Verify minimum requirements
    IF medical_practices_count >= 50 AND healthcare_services_count >= 50 AND 
       mental_health_count >= 50 AND healthcare_technology_count >= 50 THEN
        RAISE NOTICE 'SUCCESS: All healthcare industries have 50+ keywords';
    ELSE
        RAISE NOTICE 'WARNING: Some healthcare industries may not have sufficient keywords';
    END IF;
    
    IF total_healthcare_keywords >= 200 THEN
        RAISE NOTICE 'SUCCESS: Total healthcare keywords target (200+) achieved';
    ELSE
        RAISE NOTICE 'WARNING: Total healthcare keywords below target (200+)';
    END IF;
END $$;

-- Display keyword weight distribution
SELECT 
    'HEALTHCARE KEYWORD WEIGHT DISTRIBUTION' as summary_type,
    '' as spacer;

SELECT 
    i.name as industry_name,
    COUNT(kw.keyword) as keyword_count,
    MIN(kw.base_weight) as min_weight,
    MAX(kw.base_weight) as max_weight,
    ROUND(AVG(kw.base_weight), 3) as avg_weight,
    COUNT(CASE WHEN kw.base_weight >= 0.90 THEN 1 END) as high_weight_count,
    COUNT(CASE WHEN kw.base_weight >= 0.70 AND kw.base_weight < 0.90 THEN 1 END) as medium_weight_count,
    COUNT(CASE WHEN kw.base_weight < 0.70 THEN 1 END) as low_weight_count
FROM industries i
JOIN keyword_weights kw ON i.id = kw.industry_id
WHERE i.name IN ('Medical Practices', 'Healthcare Services', 'Mental Health', 'Healthcare Technology')
AND kw.is_active = true
GROUP BY i.name
ORDER BY i.name;

-- Display sample keywords for each industry
SELECT 
    'SAMPLE HEALTHCARE KEYWORDS BY INDUSTRY' as summary_type,
    '' as spacer;

SELECT 
    i.name as industry_name,
    kw.keyword,
    kw.base_weight
FROM industries i
JOIN keyword_weights kw ON i.id = kw.industry_id
WHERE i.name IN ('Medical Practices', 'Healthcare Services', 'Mental Health', 'Healthcare Technology')
AND kw.is_active = true
ORDER BY i.name, kw.base_weight DESC
LIMIT 20;

-- Commit transaction
COMMIT;

-- =============================================================================
-- SUCCESS MESSAGE
-- =============================================================================
DO $$
BEGIN
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'HEALTHCARE KEYWORDS ADDITION COMPLETED SUCCESSFULLY';
    RAISE NOTICE '=============================================================================';
    RAISE NOTICE 'Added comprehensive healthcare keywords for 4 industries:';
    RAISE NOTICE '1. Medical Practices: 50+ keywords (family medicine, specialists, clinical services)';
    RAISE NOTICE '2. Healthcare Services: 50+ keywords (hospitals, clinics, medical facilities)';
    RAISE NOTICE '3. Mental Health: 50+ keywords (counseling, therapy, psychological services)';
    RAISE NOTICE '4. Healthcare Technology: 50+ keywords (medical devices, health IT, digital health)';
    RAISE NOTICE 'Total: 200+ healthcare-specific keywords with base weights 0.50-1.00';
    RAISE NOTICE 'All keywords are active and ready for classification testing';
    RAISE NOTICE '=============================================================================';
END $$;
