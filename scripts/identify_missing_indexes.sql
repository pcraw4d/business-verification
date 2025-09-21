-- =============================================================================
-- MISSING INDEX IDENTIFICATION SCRIPT
-- Subtask 3.2.1: Identify Missing Indexes for New Classification and Risk Tables
-- =============================================================================
-- This script identifies missing indexes based on query patterns and table relationships

-- =============================================================================
-- 1. MISSING INDEXES FOR RISK KEYWORDS TABLE
-- =============================================================================

-- Risk Keywords Table Analysis
-- Based on query patterns: keyword lookups, risk category filtering, severity filtering

-- Missing indexes for risk_keywords table:
CREATE INDEX IF NOT EXISTS idx_risk_keywords_keyword ON risk_keywords(keyword);
CREATE INDEX IF NOT EXISTS idx_risk_keywords_risk_category ON risk_keywords(risk_category);
CREATE INDEX IF NOT EXISTS idx_risk_keywords_risk_severity ON risk_keywords(risk_severity);
CREATE INDEX IF NOT EXISTS idx_risk_keywords_is_active ON risk_keywords(is_active);
CREATE INDEX IF NOT EXISTS idx_risk_keywords_created_at ON risk_keywords(created_at);

-- Composite indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_risk_keywords_category_severity ON risk_keywords(risk_category, risk_severity);
CREATE INDEX IF NOT EXISTS idx_risk_keywords_active_category ON risk_keywords(is_active, risk_category);
CREATE INDEX IF NOT EXISTS idx_risk_keywords_keyword_active ON risk_keywords(keyword, is_active);

-- Partial indexes for high-selectivity queries
CREATE INDEX IF NOT EXISTS idx_risk_keywords_critical_risks ON risk_keywords(keyword, risk_category) 
    WHERE risk_severity = 'critical' AND is_active = true;
CREATE INDEX IF NOT EXISTS idx_risk_keywords_illegal_activities ON risk_keywords(keyword, risk_severity) 
    WHERE risk_category = 'illegal' AND is_active = true;

-- =============================================================================
-- 2. MISSING INDEXES FOR INDUSTRY CODE CROSSWALKS TABLE
-- =============================================================================

-- Industry Code Crosswalks Table Analysis
-- Based on query patterns: industry lookups, code type filtering, primary code identification

-- Missing indexes for industry_code_crosswalks table:
CREATE INDEX IF NOT EXISTS idx_industry_code_crosswalks_industry_id ON industry_code_crosswalks(industry_id);
CREATE INDEX IF NOT EXISTS idx_industry_code_crosswalks_mcc_code ON industry_code_crosswalks(mcc_code);
CREATE INDEX IF NOT EXISTS idx_industry_code_crosswalks_naics_code ON industry_code_crosswalks(naics_code);
CREATE INDEX IF NOT EXISTS idx_industry_code_crosswalks_sic_code ON industry_code_crosswalks(sic_code);
CREATE INDEX IF NOT EXISTS idx_industry_code_crosswalks_is_primary ON industry_code_crosswalks(is_primary);
CREATE INDEX IF NOT EXISTS idx_industry_code_crosswalks_is_active ON industry_code_crosswalks(is_active);
CREATE INDEX IF NOT EXISTS idx_industry_code_crosswalks_confidence_score ON industry_code_crosswalks(confidence_score);

-- Composite indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_industry_code_crosswalks_industry_active ON industry_code_crosswalks(industry_id, is_active);
CREATE INDEX IF NOT EXISTS idx_industry_code_crosswalks_industry_primary ON industry_code_crosswalks(industry_id, is_primary);
CREATE INDEX IF NOT EXISTS idx_industry_code_crosswalks_codes_lookup ON industry_code_crosswalks(mcc_code, naics_code, sic_code);

-- Partial indexes for high-selectivity queries
CREATE INDEX IF NOT EXISTS idx_industry_code_crosswalks_primary_codes ON industry_code_crosswalks(industry_id, mcc_code, naics_code, sic_code) 
    WHERE is_primary = true AND is_active = true;
CREATE INDEX IF NOT EXISTS idx_industry_code_crosswalks_high_confidence ON industry_code_crosswalks(industry_id, confidence_score) 
    WHERE confidence_score >= 0.80 AND is_active = true;

-- =============================================================================
-- 3. MISSING INDEXES FOR BUSINESS RISK ASSESSMENTS TABLE
-- =============================================================================

-- Business Risk Assessments Table Analysis
-- Based on query patterns: business lookups, risk level filtering, assessment date filtering

-- Missing indexes for business_risk_assessments table:
CREATE INDEX IF NOT EXISTS idx_business_risk_assessments_business_id ON business_risk_assessments(business_id);
CREATE INDEX IF NOT EXISTS idx_business_risk_assessments_risk_keyword_id ON business_risk_assessments(risk_keyword_id);
CREATE INDEX IF NOT EXISTS idx_business_risk_assessments_risk_level ON business_risk_assessments(risk_level);
CREATE INDEX IF NOT EXISTS idx_business_risk_assessments_risk_score ON business_risk_assessments(risk_score);
CREATE INDEX IF NOT EXISTS idx_business_risk_assessments_assessment_date ON business_risk_assessments(assessment_date);
CREATE INDEX IF NOT EXISTS idx_business_risk_assessments_created_at ON business_risk_assessments(created_at);
CREATE INDEX IF NOT EXISTS idx_business_risk_assessments_expires_at ON business_risk_assessments(expires_at);

-- Composite indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_business_risk_assessments_business_risk ON business_risk_assessments(business_id, risk_level);
CREATE INDEX IF NOT EXISTS idx_business_risk_assessments_business_date ON business_risk_assessments(business_id, assessment_date);
CREATE INDEX IF NOT EXISTS idx_business_risk_assessments_risk_level_score ON business_risk_assessments(risk_level, risk_score);
CREATE INDEX IF NOT EXISTS idx_business_risk_assessments_date_risk ON business_risk_assessments(assessment_date, risk_level);

-- Partial indexes for high-selectivity queries
CREATE INDEX IF NOT EXISTS idx_business_risk_assessments_high_risk ON business_risk_assessments(business_id, assessment_date) 
    WHERE risk_level IN ('high', 'critical');
CREATE INDEX IF NOT EXISTS idx_business_risk_assessments_active_assessments ON business_risk_assessments(business_id, risk_level) 
    WHERE expires_at IS NULL OR expires_at > NOW();

-- =============================================================================
-- 4. MISSING INDEXES FOR ENHANCED CLASSIFICATION TABLES
-- =============================================================================

-- Industries Table - Additional indexes needed
CREATE INDEX IF NOT EXISTS idx_industries_category ON industries(category);
CREATE INDEX IF NOT EXISTS idx_industries_is_active ON industries(is_active);
CREATE INDEX IF NOT EXISTS idx_industries_confidence_threshold ON industries(confidence_threshold);

-- Composite indexes for industries
CREATE INDEX IF NOT EXISTS idx_industries_active_category ON industries(is_active, category);
CREATE INDEX IF NOT EXISTS idx_industries_name_active ON industries(name, is_active);

-- Industry Keywords Table - Additional indexes needed
CREATE INDEX IF NOT EXISTS idx_industry_keywords_weight ON industry_keywords(weight);
CREATE INDEX IF NOT EXISTS idx_industry_keywords_is_primary ON industry_keywords(is_primary);
CREATE INDEX IF NOT EXISTS idx_industry_keywords_context ON industry_keywords(context);

-- Composite indexes for industry_keywords
CREATE INDEX IF NOT EXISTS idx_industry_keywords_industry_weight ON industry_keywords(industry_id, weight);
CREATE INDEX IF NOT EXISTS idx_industry_keywords_industry_primary ON industry_keywords(industry_id, is_primary);
CREATE INDEX IF NOT EXISTS idx_industry_keywords_keyword_weight ON industry_keywords(keyword, weight);

-- Classification Codes Table - Additional indexes needed
CREATE INDEX IF NOT EXISTS idx_classification_codes_confidence ON classification_codes(confidence);
CREATE INDEX IF NOT EXISTS idx_classification_codes_is_primary ON classification_codes(is_primary);

-- Composite indexes for classification_codes
CREATE INDEX IF NOT EXISTS idx_classification_codes_type_code ON classification_codes(code_type, code);
CREATE INDEX IF NOT EXISTS idx_classification_codes_industry_type ON classification_codes(industry_id, code_type);
CREATE INDEX IF NOT EXISTS idx_classification_codes_industry_primary ON classification_codes(industry_id, is_primary);

-- =============================================================================
-- 5. MISSING INDEXES FOR EXISTING BUSINESS TABLES
-- =============================================================================

-- Merchants Table - Additional indexes needed (if not already present)
CREATE INDEX IF NOT EXISTS idx_merchants_industry ON merchants(industry);
CREATE INDEX IF NOT EXISTS idx_merchants_industry_code ON merchants(industry_code);
CREATE INDEX IF NOT EXISTS idx_merchants_website_url ON merchants(website_url);
CREATE INDEX IF NOT EXISTS idx_merchants_employee_count ON merchants(employee_count);
CREATE INDEX IF NOT EXISTS idx_merchants_annual_revenue ON merchants(annual_revenue);

-- Composite indexes for merchants
CREATE INDEX IF NOT EXISTS idx_merchants_industry_code_active ON merchants(industry_code, is_active);
CREATE INDEX IF NOT EXISTS idx_merchants_country_industry ON merchants(country_code, industry);

-- Business Classifications Table - Additional indexes needed
CREATE INDEX IF NOT EXISTS idx_business_classifications_industry ON business_classifications(industry);
CREATE INDEX IF NOT EXISTS idx_business_classifications_confidence_score ON business_classifications(confidence_score);
CREATE INDEX IF NOT EXISTS idx_business_classifications_classification_method ON business_classifications(classification_method);

-- Composite indexes for business_classifications
CREATE INDEX IF NOT EXISTS idx_business_classifications_business_industry ON business_classifications(business_id, industry);
CREATE INDEX IF NOT EXISTS idx_business_classifications_business_method ON business_classifications(business_id, classification_method);
CREATE INDEX IF NOT EXISTS idx_business_classifications_industry_confidence ON business_classifications(industry, confidence_score);

-- =============================================================================
-- 6. MISSING INDEXES FOR MONITORING AND PERFORMANCE TABLES
-- =============================================================================

-- Audit Logs Table - Additional indexes needed
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource_type ON audit_logs(resource_type);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource_id ON audit_logs(resource_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_event_type ON audit_logs(event_type);

-- Composite indexes for audit_logs
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_event ON audit_logs(user_id, event_type);
CREATE INDEX IF NOT EXISTS idx_audit_logs_event_date ON audit_logs(event_type, created_at);

-- External Service Calls Table - Additional indexes needed
CREATE INDEX IF NOT EXISTS idx_external_service_calls_service_name ON external_service_calls(service_name);
CREATE INDEX IF NOT EXISTS idx_external_service_calls_status ON external_service_calls(status);
CREATE INDEX IF NOT EXISTS idx_external_service_calls_response_time ON external_service_calls(response_time_ms);

-- Composite indexes for external_service_calls
CREATE INDEX IF NOT EXISTS idx_external_service_calls_service_status ON external_service_calls(service_name, status);
CREATE INDEX IF NOT EXISTS idx_external_service_calls_user_service ON external_service_calls(user_id, service_name);
CREATE INDEX IF NOT EXISTS idx_external_service_calls_service_date ON external_service_calls(service_name, created_at);

-- =============================================================================
-- 7. MISSING INDEXES FOR CACHE AND SESSION TABLES
-- =============================================================================

-- Cache Entries Table - Additional indexes needed
CREATE INDEX IF NOT EXISTS idx_cache_entries_key ON cache_entries(cache_key);
CREATE INDEX IF NOT EXISTS idx_cache_entries_created_at ON cache_entries(created_at);

-- Composite indexes for cache_entries
CREATE INDEX IF NOT EXISTS idx_cache_entries_key_expires ON cache_entries(cache_key, expires_at);
CREATE INDEX IF NOT EXISTS idx_cache_entries_type_expires ON cache_entries(cache_type, expires_at);

-- =============================================================================
-- 8. MISSING INDEXES FOR JSONB COLUMNS
-- =============================================================================

-- GIN indexes for JSONB columns to support efficient JSON queries
CREATE INDEX IF NOT EXISTS idx_users_metadata_gin ON users USING GIN (metadata);
CREATE INDEX IF NOT EXISTS idx_merchants_address_gin ON merchants USING GIN (address);
CREATE INDEX IF NOT EXISTS idx_merchants_contact_info_gin ON merchants USING GIN (contact_info);
CREATE INDEX IF NOT EXISTS idx_merchants_metadata_gin ON merchants USING GIN (metadata);
CREATE INDEX IF NOT EXISTS idx_business_risk_assessments_detected_patterns_gin ON business_risk_assessments USING GIN (detected_patterns);
CREATE INDEX IF NOT EXISTS idx_business_risk_assessments_assessment_metadata_gin ON business_risk_assessments USING GIN (assessment_metadata);

-- =============================================================================
-- 9. MISSING INDEXES FOR TEXT SEARCH
-- =============================================================================

-- Full-text search indexes for business names and descriptions
CREATE INDEX IF NOT EXISTS idx_merchants_name_fts ON merchants USING GIN (to_tsvector('english', name));
CREATE INDEX IF NOT EXISTS idx_merchants_description_fts ON merchants USING GIN (to_tsvector('english', description));
CREATE INDEX IF NOT EXISTS idx_industries_name_fts ON industries USING GIN (to_tsvector('english', name));
CREATE INDEX IF NOT EXISTS idx_industries_description_fts ON industries USING GIN (to_tsvector('english', description));

-- =============================================================================
-- 10. MISSING INDEXES FOR ARRAY COLUMNS
-- =============================================================================

-- GIN indexes for array columns
CREATE INDEX IF NOT EXISTS idx_risk_keywords_mcc_codes_gin ON risk_keywords USING GIN (mcc_codes);
CREATE INDEX IF NOT EXISTS idx_risk_keywords_naics_codes_gin ON risk_keywords USING GIN (naics_codes);
CREATE INDEX IF NOT EXISTS idx_risk_keywords_sic_codes_gin ON risk_keywords USING GIN (sic_codes);
CREATE INDEX IF NOT EXISTS idx_risk_keywords_card_brand_restrictions_gin ON risk_keywords USING GIN (card_brand_restrictions);
CREATE INDEX IF NOT EXISTS idx_risk_keywords_detection_patterns_gin ON risk_keywords USING GIN (detection_patterns);
CREATE INDEX IF NOT EXISTS idx_risk_keywords_synonyms_gin ON risk_keywords USING GIN (synonyms);
CREATE INDEX IF NOT EXISTS idx_business_risk_assessments_detected_keywords_gin ON business_risk_assessments USING GIN (detected_keywords);

-- =============================================================================
-- 11. SUMMARY OF MISSING INDEXES
-- =============================================================================

-- Create a summary view of all recommended indexes
CREATE OR REPLACE VIEW missing_indexes_summary AS
SELECT 
    'risk_keywords' as table_name,
    'Single Column Indexes' as index_type,
    ARRAY[
        'idx_risk_keywords_keyword',
        'idx_risk_keywords_risk_category', 
        'idx_risk_keywords_risk_severity',
        'idx_risk_keywords_is_active',
        'idx_risk_keywords_created_at'
    ] as index_names,
    'High priority - Core lookup columns' as priority

UNION ALL

SELECT 
    'risk_keywords' as table_name,
    'Composite Indexes' as index_type,
    ARRAY[
        'idx_risk_keywords_category_severity',
        'idx_risk_keywords_active_category',
        'idx_risk_keywords_keyword_active'
    ] as index_names,
    'Medium priority - Common query patterns' as priority

UNION ALL

SELECT 
    'industry_code_crosswalks' as table_name,
    'Single Column Indexes' as index_type,
    ARRAY[
        'idx_industry_code_crosswalks_industry_id',
        'idx_industry_code_crosswalks_mcc_code',
        'idx_industry_code_crosswalks_naics_code',
        'idx_industry_code_crosswalks_sic_code',
        'idx_industry_code_crosswalks_is_primary',
        'idx_industry_code_crosswalks_is_active'
    ] as index_names,
    'High priority - Core lookup columns' as priority

UNION ALL

SELECT 
    'business_risk_assessments' as table_name,
    'Single Column Indexes' as index_type,
    ARRAY[
        'idx_business_risk_assessments_business_id',
        'idx_business_risk_assessments_risk_level',
        'idx_business_risk_assessments_risk_score',
        'idx_business_risk_assessments_assessment_date'
    ] as index_names,
    'High priority - Core lookup columns' as priority

UNION ALL

SELECT 
    'JSONB and Array Indexes' as table_name,
    'Specialized Indexes' as index_type,
    ARRAY[
        'GIN indexes for JSONB columns',
        'GIN indexes for array columns',
        'Full-text search indexes'
    ] as index_names,
    'Medium priority - Advanced query support' as priority;

-- Query the summary
SELECT * FROM missing_indexes_summary ORDER BY priority, table_name;
