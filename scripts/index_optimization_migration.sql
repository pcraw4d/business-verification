-- =============================================================================
-- INDEX OPTIMIZATION MIGRATION SCRIPT
-- Subtask 3.2.2: Implement Index Optimizations
-- Supabase Table Improvement Implementation Plan
-- =============================================================================
-- This script implements comprehensive index optimizations for the enhanced
-- classification system and risk keywords implementation

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
CREATE EXTENSION IF NOT EXISTS "btree_gin";

-- =============================================================================
-- 1. ADD MISSING INDEXES FOR NEW CLASSIFICATION TABLES
-- =============================================================================

-- Additional indexes for industries table (beyond existing ones)
CREATE INDEX IF NOT EXISTS idx_industries_category_active ON industries(category, is_active);
CREATE INDEX IF NOT EXISTS idx_industries_confidence_threshold ON industries(confidence_threshold DESC);
CREATE INDEX IF NOT EXISTS idx_industries_created_at ON industries(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_industries_updated_at ON industries(updated_at DESC);

-- Additional indexes for industry_keywords table
CREATE INDEX IF NOT EXISTS idx_industry_keywords_industry_active ON industry_keywords(industry_id, is_active);
CREATE INDEX IF NOT EXISTS idx_industry_keywords_weight_active ON industry_keywords(weight DESC, is_active);
CREATE INDEX IF NOT EXISTS idx_industry_keywords_created_at ON industry_keywords(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_industry_keywords_updated_at ON industry_keywords(updated_at DESC);

-- Additional indexes for classification_codes table
CREATE INDEX IF NOT EXISTS idx_classification_codes_industry_active ON classification_codes(industry_id, is_active);
CREATE INDEX IF NOT EXISTS idx_classification_codes_type_active ON classification_codes(code_type, is_active);
CREATE INDEX IF NOT EXISTS idx_classification_codes_created_at ON classification_codes(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_classification_codes_updated_at ON classification_codes(updated_at DESC);

-- Additional indexes for industry_patterns table
CREATE INDEX IF NOT EXISTS idx_industry_patterns_industry_active ON industry_patterns(industry_id, is_active);
CREATE INDEX IF NOT EXISTS idx_industry_patterns_type_active ON industry_patterns(pattern_type, is_active);
CREATE INDEX IF NOT EXISTS idx_industry_patterns_confidence_active ON industry_patterns(confidence_score DESC, is_active);
CREATE INDEX IF NOT EXISTS idx_industry_patterns_created_at ON industry_patterns(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_industry_patterns_updated_at ON industry_patterns(updated_at DESC);

-- Additional indexes for keyword_weights table
CREATE INDEX IF NOT EXISTS idx_keyword_weights_industry_active ON keyword_weights(industry_id, base_weight DESC);
CREATE INDEX IF NOT EXISTS idx_keyword_weights_usage_count_desc ON keyword_weights(usage_count DESC);
CREATE INDEX IF NOT EXISTS idx_keyword_weights_last_updated ON keyword_weights(last_updated DESC);
CREATE INDEX IF NOT EXISTS idx_keyword_weights_created_at ON keyword_weights(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_keyword_weights_updated_at ON keyword_weights(updated_at DESC);

-- =============================================================================
-- 2. ADD MISSING INDEXES FOR NEW RISK KEYWORDS TABLES
-- =============================================================================

-- Additional indexes for risk_keywords table (beyond existing ones)
CREATE INDEX IF NOT EXISTS idx_risk_keywords_created_at ON risk_keywords(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_risk_keywords_updated_at ON risk_keywords(updated_at DESC);

-- Additional indexes for industry_code_crosswalks table (beyond existing ones)
CREATE INDEX IF NOT EXISTS idx_industry_code_crosswalks_created_at ON industry_code_crosswalks(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_industry_code_crosswalks_updated_at ON industry_code_crosswalks(updated_at DESC);

-- Additional indexes for business_risk_assessments table (beyond existing ones)
CREATE INDEX IF NOT EXISTS idx_business_risk_assessments_created_at ON business_risk_assessments(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_business_risk_assessments_updated_at ON business_risk_assessments(updated_at DESC);

-- Additional indexes for risk_keyword_relationships table (beyond existing ones)
CREATE INDEX IF NOT EXISTS idx_risk_keyword_relationships_created_at ON risk_keyword_relationships(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_risk_keyword_relationships_updated_at ON risk_keyword_relationships(updated_at DESC);

-- Additional indexes for classification_performance_metrics table (beyond existing ones)
CREATE INDEX IF NOT EXISTS idx_classification_performance_created_at ON classification_performance_metrics(created_at DESC);

-- =============================================================================
-- 3. OPTIMIZE EXISTING INDEXES ON CORE TABLES
-- =============================================================================

-- Optimize users table indexes
CREATE INDEX IF NOT EXISTS idx_users_email_verified ON users(email_verified, email);
CREATE INDEX IF NOT EXISTS idx_users_active_created ON users(is_active, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_users_role_active ON users(role, is_active);
CREATE INDEX IF NOT EXISTS idx_users_last_login ON users(last_login_at DESC) WHERE last_login_at IS NOT NULL;

-- Optimize businesses table indexes (if exists)
CREATE INDEX IF NOT EXISTS idx_businesses_user_created ON businesses(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_businesses_country_industry ON businesses(country_code, industry);
CREATE INDEX IF NOT EXISTS idx_businesses_industry_active ON businesses(industry, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_businesses_updated_at ON businesses(updated_at DESC);

-- Optimize merchants table indexes (if exists)
CREATE INDEX IF NOT EXISTS idx_merchants_user_created ON merchants(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_merchants_country_industry ON merchants(country_code, industry);
CREATE INDEX IF NOT EXISTS idx_merchants_industry_active ON merchants(industry, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_merchants_updated_at ON merchants(updated_at DESC);

-- Optimize audit_logs table indexes
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_created ON audit_logs(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_event_created ON audit_logs(event_type, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource_created ON audit_logs(resource_type, resource_id, created_at DESC);

-- =============================================================================
-- 4. CREATE COMPOSITE INDEXES FOR COMMON QUERY PATTERNS
-- =============================================================================

-- Classification system composite indexes
CREATE INDEX IF NOT EXISTS idx_industries_category_confidence ON industries(category, confidence_threshold DESC, is_active);
CREATE INDEX IF NOT EXISTS idx_industry_keywords_weight_industry ON industry_keywords(weight DESC, industry_id, is_active);
CREATE INDEX IF NOT EXISTS idx_classification_codes_type_code ON classification_codes(code_type, code, is_active);
CREATE INDEX IF NOT EXISTS idx_industry_patterns_confidence_type ON industry_patterns(confidence_score DESC, pattern_type, is_active);

-- Risk assessment composite indexes
CREATE INDEX IF NOT EXISTS idx_risk_keywords_category_severity_weight ON risk_keywords(risk_category, risk_severity, risk_score_weight DESC, is_active);
CREATE INDEX IF NOT EXISTS idx_business_risk_assessments_business_level ON business_risk_assessments(business_id, risk_level, risk_score DESC);
CREATE INDEX IF NOT EXISTS idx_business_risk_assessments_date_level ON business_risk_assessments(assessment_date DESC, risk_level, risk_score DESC);

-- Performance monitoring composite indexes
CREATE INDEX IF NOT EXISTS idx_classification_performance_timestamp_accuracy ON classification_performance_metrics(timestamp DESC, accuracy_score DESC);
CREATE INDEX IF NOT EXISTS idx_classification_performance_method_accuracy ON classification_performance_metrics(classification_method, accuracy_score DESC, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_classification_performance_risk_accuracy ON classification_performance_metrics(risk_level, risk_score DESC, accuracy_score DESC);

-- =============================================================================
-- 5. CREATE SPECIALIZED INDEXES FOR TEXT SEARCH AND ARRAY OPERATIONS
-- =============================================================================

-- Full-text search indexes for better text matching
CREATE INDEX IF NOT EXISTS idx_industries_fulltext ON industries USING GIN(
    to_tsvector('english', name || ' ' || COALESCE(description, ''))
);

CREATE INDEX IF NOT EXISTS idx_industry_keywords_fulltext ON industry_keywords USING GIN(
    to_tsvector('english', keyword)
);

CREATE INDEX IF NOT EXISTS idx_classification_codes_fulltext ON classification_codes USING GIN(
    to_tsvector('english', code || ' ' || description)
);

-- Trigram indexes for fuzzy matching
CREATE INDEX IF NOT EXISTS idx_industries_name_trgm ON industries USING gin(name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_industry_keywords_keyword_trgm ON industry_keywords USING gin(keyword gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_classification_codes_description_trgm ON classification_codes USING gin(description gin_trgm_ops);

-- =============================================================================
-- 6. CREATE PARTIAL INDEXES FOR COMMON FILTERING CONDITIONS
-- =============================================================================

-- Partial indexes for active records only
CREATE INDEX IF NOT EXISTS idx_industries_active_recent ON industries(created_at DESC) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_industry_keywords_active_recent ON industry_keywords(created_at DESC) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_classification_codes_active_recent ON classification_codes(created_at DESC) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_industry_patterns_active_recent ON industry_patterns(created_at DESC) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_risk_keywords_active_recent ON risk_keywords(created_at DESC) WHERE is_active = true;

-- Partial indexes for high-confidence records
CREATE INDEX IF NOT EXISTS idx_industries_high_confidence ON industries(name, category) WHERE confidence_threshold >= 0.80;
CREATE INDEX IF NOT EXISTS idx_industry_patterns_high_confidence ON industry_patterns(industry_id, pattern_type) WHERE confidence_score >= 0.80;
CREATE INDEX IF NOT EXISTS idx_industry_code_crosswalks_high_confidence ON industry_code_crosswalks(industry_id, mcc_code, naics_code, sic_code) WHERE confidence_score >= 0.80;

-- Partial indexes for high-risk assessments
CREATE INDEX IF NOT EXISTS idx_business_risk_assessments_high_risk ON business_risk_assessments(business_id, assessment_date DESC) WHERE risk_level IN ('high', 'critical');
CREATE INDEX IF NOT EXISTS idx_risk_keywords_high_severity ON risk_keywords(keyword, risk_category) WHERE risk_severity IN ('high', 'critical');

-- =============================================================================
-- 7. CREATE INDEXES FOR JSONB FIELDS
-- =============================================================================

-- JSONB indexes for metadata fields
CREATE INDEX IF NOT EXISTS idx_users_metadata ON users USING GIN(metadata);
CREATE INDEX IF NOT EXISTS idx_businesses_address ON businesses USING GIN(address);
CREATE INDEX IF NOT EXISTS idx_businesses_contact_info ON businesses USING GIN(contact_info);
CREATE INDEX IF NOT EXISTS idx_businesses_metadata ON businesses USING GIN(metadata);
CREATE INDEX IF NOT EXISTS idx_merchants_address ON merchants USING GIN(address);
CREATE INDEX IF NOT EXISTS idx_merchants_contact_info ON merchants USING GIN(contact_info);
CREATE INDEX IF NOT EXISTS idx_merchants_metadata ON merchants USING GIN(metadata);

-- JSONB indexes for assessment metadata
CREATE INDEX IF NOT EXISTS idx_business_risk_assessments_metadata_gin ON business_risk_assessments USING GIN(assessment_metadata);
CREATE INDEX IF NOT EXISTS idx_business_risk_assessments_patterns_gin ON business_risk_assessments USING GIN(detected_patterns);

-- =============================================================================
-- 8. CREATE INDEXES FOR TIMESTAMP-BASED QUERIES
-- =============================================================================

-- Time-based indexes for analytics and reporting
CREATE INDEX IF NOT EXISTS idx_industries_created_month ON industries(DATE_TRUNC('month', created_at));
CREATE INDEX IF NOT EXISTS idx_industry_keywords_created_month ON industry_keywords(DATE_TRUNC('month', created_at));
CREATE INDEX IF NOT EXISTS idx_risk_keywords_created_month ON risk_keywords(DATE_TRUNC('month', created_at));
CREATE INDEX IF NOT EXISTS idx_business_risk_assessments_assessment_month ON business_risk_assessments(DATE_TRUNC('month', assessment_date));
CREATE INDEX IF NOT EXISTS idx_classification_performance_timestamp_month ON classification_performance_metrics(DATE_TRUNC('month', timestamp));

-- =============================================================================
-- 9. CREATE INDEXES FOR FOREIGN KEY RELATIONSHIPS
-- =============================================================================

-- Foreign key indexes for better join performance
CREATE INDEX IF NOT EXISTS idx_industry_keywords_industry_fk ON industry_keywords(industry_id);
CREATE INDEX IF NOT EXISTS idx_classification_codes_industry_fk ON classification_codes(industry_id);
CREATE INDEX IF NOT EXISTS idx_industry_patterns_industry_fk ON industry_patterns(industry_id);
CREATE INDEX IF NOT EXISTS idx_keyword_weights_industry_fk ON keyword_weights(industry_id);
CREATE INDEX IF NOT EXISTS idx_industry_code_crosswalks_industry_fk ON industry_code_crosswalks(industry_id);
CREATE INDEX IF NOT EXISTS idx_business_risk_assessments_keyword_fk ON business_risk_assessments(risk_keyword_id);
CREATE INDEX IF NOT EXISTS idx_risk_keyword_relationships_parent_fk ON risk_keyword_relationships(parent_keyword_id);
CREATE INDEX IF NOT EXISTS idx_risk_keyword_relationships_child_fk ON risk_keyword_relationships(child_keyword_id);

-- =============================================================================
-- 10. CREATE INDEXES FOR UNIQUE CONSTRAINTS AND BUSINESS LOGIC
-- =============================================================================

-- Indexes to support unique constraints and business logic
CREATE INDEX IF NOT EXISTS idx_industries_name_unique ON industries(name) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_industry_keywords_unique ON industry_keywords(industry_id, keyword) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_classification_codes_unique ON classification_codes(code_type, code) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_keyword_weights_unique ON keyword_weights(keyword, industry_id);
CREATE INDEX IF NOT EXISTS idx_risk_keywords_keyword_unique ON risk_keywords(keyword) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_industry_code_crosswalks_unique ON industry_code_crosswalks(industry_id, mcc_code, naics_code, sic_code);

-- =============================================================================
-- 11. CREATE INDEXES FOR PERFORMANCE MONITORING AND ANALYTICS
-- =============================================================================

-- Indexes for performance monitoring queries
CREATE INDEX IF NOT EXISTS idx_classification_performance_response_time_range ON classification_performance_metrics(response_time_ms) WHERE response_time_ms > 100;
CREATE INDEX IF NOT EXISTS idx_classification_performance_accuracy_range ON classification_performance_metrics(accuracy_score) WHERE accuracy_score < 0.80;
CREATE INDEX IF NOT EXISTS idx_classification_performance_risk_high ON classification_performance_metrics(timestamp DESC) WHERE risk_level IN ('high', 'critical');

-- Indexes for analytics and reporting
CREATE INDEX IF NOT EXISTS idx_business_risk_assessments_analytics ON business_risk_assessments(assessment_date DESC, risk_level, risk_score DESC, business_id);
CREATE INDEX IF NOT EXISTS idx_classification_performance_analytics ON classification_performance_metrics(timestamp DESC, classification_method, accuracy_score DESC, risk_level);

-- =============================================================================
-- 12. CREATE INDEXES FOR CACHING AND FREQUENT LOOKUPS
-- =============================================================================

-- Indexes for frequently accessed data
CREATE INDEX IF NOT EXISTS idx_industries_active_lookup ON industries(id, name, category) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_industry_keywords_active_lookup ON industry_keywords(industry_id, keyword, weight) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_risk_keywords_active_lookup ON risk_keywords(id, keyword, risk_category, risk_severity) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_classification_codes_active_lookup ON classification_codes(code_type, code, description) WHERE is_active = true;

-- =============================================================================
-- 13. CREATE INDEXES FOR SEARCH AND FILTERING OPERATIONS
-- =============================================================================

-- Search indexes for common filtering operations
CREATE INDEX IF NOT EXISTS idx_industries_search ON industries(name, category, is_active);
CREATE INDEX IF NOT EXISTS idx_industry_keywords_search ON industry_keywords(keyword, industry_id, is_active);
CREATE INDEX IF NOT EXISTS idx_risk_keywords_search ON risk_keywords(keyword, risk_category, risk_severity, is_active);
CREATE INDEX IF NOT EXISTS idx_classification_codes_search ON classification_codes(code_type, code, industry_id, is_active);

-- =============================================================================
-- 14. CREATE INDEXES FOR DATA INTEGRITY AND VALIDATION
-- =============================================================================

-- Indexes to support data validation and integrity checks
CREATE INDEX IF NOT EXISTS idx_industries_validation ON industries(id, name, category, confidence_threshold, is_active);
CREATE INDEX IF NOT EXISTS idx_industry_keywords_validation ON industry_keywords(industry_id, keyword, weight, is_active);
CREATE INDEX IF NOT EXISTS idx_risk_keywords_validation ON risk_keywords(id, keyword, risk_category, risk_severity, risk_score_weight, is_active);
CREATE INDEX IF NOT EXISTS idx_classification_codes_validation ON classification_codes(industry_id, code_type, code, is_active);

-- =============================================================================
-- 15. CREATE INDEXES FOR AUDIT AND COMPLIANCE TRACKING
-- =============================================================================

-- Indexes for audit trail and compliance tracking
CREATE INDEX IF NOT EXISTS idx_industries_audit ON industries(created_at, updated_at, is_active);
CREATE INDEX IF NOT EXISTS idx_industry_keywords_audit ON industry_keywords(created_at, updated_at, is_active);
CREATE INDEX IF NOT EXISTS idx_risk_keywords_audit ON risk_keywords(created_at, updated_at, is_active);
CREATE INDEX IF NOT EXISTS idx_classification_codes_audit ON classification_codes(created_at, updated_at, is_active);
CREATE INDEX IF NOT EXISTS idx_business_risk_assessments_audit ON business_risk_assessments(created_at, updated_at, assessment_date);

-- =============================================================================
-- 16. CREATE INDEXES FOR SCALABILITY AND FUTURE GROWTH
-- =============================================================================

-- Indexes designed for scalability and future growth
CREATE INDEX IF NOT EXISTS idx_industries_scalable ON industries(id, name, category, is_active, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_industry_keywords_scalable ON industry_keywords(industry_id, keyword, weight, is_active, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_risk_keywords_scalable ON risk_keywords(id, keyword, risk_category, risk_severity, is_active, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_classification_codes_scalable ON classification_codes(industry_id, code_type, code, is_active, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_business_risk_assessments_scalable ON business_risk_assessments(business_id, risk_level, risk_score DESC, assessment_date DESC);

-- =============================================================================
-- 17. CREATE INDEXES FOR MACHINE LEARNING AND ANALYTICS
-- =============================================================================

-- Indexes optimized for ML and analytics workloads
CREATE INDEX IF NOT EXISTS idx_classification_performance_ml ON classification_performance_metrics(timestamp DESC, classification_method, accuracy_score DESC, risk_score DESC);
CREATE INDEX IF NOT EXISTS idx_business_risk_assessments_ml ON business_risk_assessments(assessment_date DESC, risk_score DESC, confidence_score DESC, business_id);
CREATE INDEX IF NOT EXISTS idx_industry_keywords_ml ON industry_keywords(industry_id, weight DESC, is_active);
CREATE INDEX IF NOT EXISTS idx_risk_keywords_ml ON risk_keywords(risk_category, risk_severity, risk_score_weight DESC, detection_confidence DESC, is_active);

-- =============================================================================
-- 18. CREATE INDEXES FOR API PERFORMANCE
-- =============================================================================

-- Indexes optimized for API response times
CREATE INDEX IF NOT EXISTS idx_industries_api ON industries(id, name, category, is_active) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_industry_keywords_api ON industry_keywords(industry_id, keyword, weight) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_risk_keywords_api ON risk_keywords(id, keyword, risk_category, risk_severity) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_classification_codes_api ON classification_codes(code_type, code, description) WHERE is_active = true;
CREATE INDEX IF NOT EXISTS idx_business_risk_assessments_api ON business_risk_assessments(business_id, risk_level, risk_score DESC, assessment_date DESC);

-- =============================================================================
-- 19. CREATE INDEXES FOR REPORTING AND DASHBOARDS
-- =============================================================================

-- Indexes for reporting and dashboard queries
CREATE INDEX IF NOT EXISTS idx_classification_performance_reporting ON classification_performance_metrics(timestamp DESC, classification_method, accuracy_score DESC, risk_level);
CREATE INDEX IF NOT EXISTS idx_business_risk_assessments_reporting ON business_risk_assessments(assessment_date DESC, risk_level, risk_score DESC, business_id);
CREATE INDEX IF NOT EXISTS idx_industries_reporting ON industries(category, confidence_threshold DESC, is_active, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_risk_keywords_reporting ON risk_keywords(risk_category, risk_severity, risk_score_weight DESC, is_active, created_at DESC);

-- =============================================================================
-- 20. CREATE INDEXES FOR BACKUP AND RECOVERY
-- =============================================================================

-- Indexes to support backup and recovery operations
CREATE INDEX IF NOT EXISTS idx_industries_backup ON industries(created_at, updated_at, is_active);
CREATE INDEX IF NOT EXISTS idx_industry_keywords_backup ON industry_keywords(created_at, updated_at, is_active);
CREATE INDEX IF NOT EXISTS idx_risk_keywords_backup ON risk_keywords(created_at, updated_at, is_active);
CREATE INDEX IF NOT EXISTS idx_classification_codes_backup ON classification_codes(created_at, updated_at, is_active);
CREATE INDEX IF NOT EXISTS idx_business_risk_assessments_backup ON business_risk_assessments(created_at, updated_at, assessment_date);

-- =============================================================================
-- END OF INDEX OPTIMIZATION MIGRATION
-- =============================================================================

-- Log completion
INSERT INTO audit_logs (event_type, resource_type, resource_id, details, created_at) 
VALUES ('index_optimization', 'database', 'classification_system', 
        'Index optimization migration completed successfully', NOW())
ON CONFLICT DO NOTHING;
