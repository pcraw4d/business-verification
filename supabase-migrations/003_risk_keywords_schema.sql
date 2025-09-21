-- =====================================================
-- Risk Keywords System Schema Migration
-- Supabase Implementation - Task 1.4.1
-- =====================================================

-- Enable necessary extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =====================================================
-- 1. risk_keywords Table
-- =====================================================
CREATE TABLE risk_keywords (
    id SERIAL PRIMARY KEY,
    keyword VARCHAR(255) NOT NULL,
    risk_category VARCHAR(50) NOT NULL CHECK (risk_category IN (
        'illegal', 'prohibited', 'high_risk', 'tbml', 'sanctions', 'fraud'
    )),
    risk_severity VARCHAR(20) NOT NULL CHECK (risk_severity IN (
        'low', 'medium', 'high', 'critical'
    )),
    description TEXT,
    mcc_codes TEXT[], -- Associated prohibited MCC codes
    naics_codes TEXT[], -- Associated prohibited NAICS codes
    sic_codes TEXT[], -- Associated prohibited SIC codes
    card_brand_restrictions TEXT[], -- Visa, Mastercard, Amex restrictions
    detection_patterns TEXT[], -- Regex patterns for detection
    synonyms TEXT[], -- Alternative terms and variations
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Ensure keyword uniqueness within active records
    UNIQUE(keyword) WHERE is_active = true
);

-- =====================================================
-- 2. Indexes for Performance Optimization
-- =====================================================

-- Primary search indexes
CREATE INDEX idx_risk_keywords_keyword ON risk_keywords(keyword);
CREATE INDEX idx_risk_keywords_category ON risk_keywords(risk_category);
CREATE INDEX idx_risk_keywords_severity ON risk_keywords(risk_severity);
CREATE INDEX idx_risk_keywords_active ON risk_keywords(is_active);

-- Composite indexes for common queries
CREATE INDEX idx_risk_keywords_category_severity ON risk_keywords(risk_category, risk_severity);
CREATE INDEX idx_risk_keywords_active_category ON risk_keywords(is_active, risk_category);
CREATE INDEX idx_risk_keywords_active_severity ON risk_keywords(is_active, risk_severity);

-- GIN indexes for array fields (PostgreSQL specific)
CREATE INDEX idx_risk_keywords_mcc_codes ON risk_keywords USING GIN(mcc_codes);
CREATE INDEX idx_risk_keywords_naics_codes ON risk_keywords USING GIN(naics_codes);
CREATE INDEX idx_risk_keywords_sic_codes ON risk_keywords USING GIN(sic_codes);
CREATE INDEX idx_risk_keywords_card_restrictions ON risk_keywords USING GIN(card_brand_restrictions);
CREATE INDEX idx_risk_keywords_detection_patterns ON risk_keywords USING GIN(detection_patterns);
CREATE INDEX idx_risk_keywords_synonyms ON risk_keywords USING GIN(synonyms);

-- Full-text search index for keyword and description
CREATE INDEX idx_risk_keywords_fulltext ON risk_keywords USING GIN(
    to_tsvector('english', keyword || ' ' || COALESCE(description, ''))
);

-- =====================================================
-- 3. industry_code_crosswalks Table
-- =====================================================
CREATE TABLE industry_code_crosswalks (
    id SERIAL PRIMARY KEY,
    industry_id INTEGER NOT NULL REFERENCES industries(id) ON DELETE CASCADE,
    mcc_code VARCHAR(10),
    naics_code VARCHAR(10),
    sic_code VARCHAR(10),
    code_description TEXT,
    confidence_score DECIMAL(3,2) DEFAULT 0.80 CHECK (confidence_score >= 0.00 AND confidence_score <= 1.00),
    is_primary BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Ensure unique combinations
    UNIQUE(industry_id, mcc_code, naics_code, sic_code)
);

-- Indexes for industry_code_crosswalks
CREATE INDEX idx_industry_code_crosswalks_industry ON industry_code_crosswalks(industry_id);
CREATE INDEX idx_industry_code_crosswalks_mcc ON industry_code_crosswalks(mcc_code);
CREATE INDEX idx_industry_code_crosswalks_naics ON industry_code_crosswalks(naics_code);
CREATE INDEX idx_industry_code_crosswalks_sic ON industry_code_crosswalks(sic_code);
CREATE INDEX idx_industry_code_crosswalks_active ON industry_code_crosswalks(is_active);
CREATE INDEX idx_industry_code_crosswalks_primary ON industry_code_crosswalks(is_primary);

-- =====================================================
-- 4. business_risk_assessments Table
-- =====================================================
CREATE TABLE business_risk_assessments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    business_id UUID NOT NULL, -- Will reference merchants table when available
    risk_keyword_id INTEGER REFERENCES risk_keywords(id) ON DELETE SET NULL,
    detected_keywords TEXT[],
    risk_score DECIMAL(3,2) NOT NULL CHECK (risk_score >= 0.00 AND risk_score <= 1.00),
    risk_level VARCHAR(20) NOT NULL CHECK (risk_level IN (
        'low', 'medium', 'high', 'critical'
    )),
    assessment_method VARCHAR(100),
    website_content TEXT,
    detected_patterns JSONB,
    assessment_date TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for business_risk_assessments
CREATE INDEX idx_business_risk_assessments_business ON business_risk_assessments(business_id);
CREATE INDEX idx_business_risk_assessments_keyword ON business_risk_assessments(risk_keyword_id);
CREATE INDEX idx_business_risk_assessments_score ON business_risk_assessments(risk_score);
CREATE INDEX idx_business_risk_assessments_level ON business_risk_assessments(risk_level);
CREATE INDEX idx_business_risk_assessments_date ON business_risk_assessments(assessment_date);
CREATE INDEX idx_business_risk_assessments_method ON business_risk_assessments(assessment_method);

-- GIN index for detected_keywords array
CREATE INDEX idx_business_risk_assessments_keywords ON business_risk_assessments USING GIN(detected_keywords);

-- GIN index for detected_patterns JSONB
CREATE INDEX idx_business_risk_assessments_patterns ON business_risk_assessments USING GIN(detected_patterns);

-- =====================================================
-- 5. risk_keyword_relationships Table
-- =====================================================
CREATE TABLE risk_keyword_relationships (
    id SERIAL PRIMARY KEY,
    parent_keyword_id INTEGER NOT NULL REFERENCES risk_keywords(id) ON DELETE CASCADE,
    child_keyword_id INTEGER NOT NULL REFERENCES risk_keywords(id) ON DELETE CASCADE,
    relationship_type VARCHAR(50) NOT NULL CHECK (relationship_type IN (
        'synonym', 'related', 'subcategory', 'superset', 'conflict'
    )),
    confidence_score DECIMAL(3,2) DEFAULT 0.80 CHECK (confidence_score >= 0.00 AND confidence_score <= 1.00),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Prevent self-references and duplicate relationships
    CHECK (parent_keyword_id != child_keyword_id),
    UNIQUE(parent_keyword_id, child_keyword_id, relationship_type)
);

-- Indexes for risk_keyword_relationships
CREATE INDEX idx_risk_keyword_relationships_parent ON risk_keyword_relationships(parent_keyword_id);
CREATE INDEX idx_risk_keyword_relationships_child ON risk_keyword_relationships(child_keyword_id);
CREATE INDEX idx_risk_keyword_relationships_type ON risk_keyword_relationships(relationship_type);
CREATE INDEX idx_risk_keyword_relationships_active ON risk_keyword_relationships(is_active);

-- =====================================================
-- 6. Row Level Security (RLS) Policies
-- =====================================================

-- Enable RLS on all tables
ALTER TABLE risk_keywords ENABLE ROW LEVEL SECURITY;
ALTER TABLE industry_code_crosswalks ENABLE ROW LEVEL SECURITY;
ALTER TABLE business_risk_assessments ENABLE ROW LEVEL SECURITY;
ALTER TABLE risk_keyword_relationships ENABLE ROW LEVEL SECURITY;

-- Create policies for read access (public read, authenticated write)
CREATE POLICY "Allow public read access to risk_keywords" ON risk_keywords
    FOR SELECT USING (true);

CREATE POLICY "Allow public read access to industry_code_crosswalks" ON industry_code_crosswalks
    FOR SELECT USING (true);

CREATE POLICY "Allow public read access to business_risk_assessments" ON business_risk_assessments
    FOR SELECT USING (true);

CREATE POLICY "Allow public read access to risk_keyword_relationships" ON risk_keyword_relationships
    FOR SELECT USING (true);

-- Create policies for authenticated write access
CREATE POLICY "Allow authenticated users to manage risk_keywords" ON risk_keywords
    FOR ALL USING (auth.role() = 'authenticated');

CREATE POLICY "Allow authenticated users to manage industry_code_crosswalks" ON industry_code_crosswalks
    FOR ALL USING (auth.role() = 'authenticated');

CREATE POLICY "Allow authenticated users to manage business_risk_assessments" ON business_risk_assessments
    FOR ALL USING (auth.role() = 'authenticated');

CREATE POLICY "Allow authenticated users to manage risk_keyword_relationships" ON risk_keyword_relationships
    FOR ALL USING (auth.role() = 'authenticated');

-- =====================================================
-- 7. Triggers for Updated At Timestamps
-- =====================================================

-- Create triggers for updated_at columns
CREATE TRIGGER update_risk_keywords_updated_at 
    BEFORE UPDATE ON risk_keywords 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_industry_code_crosswalks_updated_at 
    BEFORE UPDATE ON industry_code_crosswalks 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_business_risk_assessments_updated_at 
    BEFORE UPDATE ON business_risk_assessments 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_risk_keyword_relationships_updated_at 
    BEFORE UPDATE ON risk_keyword_relationships 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =====================================================
-- 8. Audit Logging for Risk Keywords
-- =====================================================

-- Extend audit_logs table to include risk-related tables
INSERT INTO audit_logs (table_name, record_id, action, old_values, new_values, user_id, timestamp)
SELECT 
    'risk_keywords' as table_name,
    0 as record_id,
    'SCHEMA_CREATED' as action,
    NULL as old_values,
    '{"tables_created": ["risk_keywords", "industry_code_crosswalks", "business_risk_assessments", "risk_keyword_relationships"]}'::jsonb as new_values,
    'system' as user_id,
    NOW() as timestamp;

-- =====================================================
-- 9. Comments for Documentation
-- =====================================================

COMMENT ON TABLE risk_keywords IS 'Core risk keywords for detecting prohibited, illegal, and high-risk business activities';
COMMENT ON TABLE industry_code_crosswalks IS 'Crosswalk mapping between industries and classification codes (MCC, NAICS, SIC)';
COMMENT ON TABLE business_risk_assessments IS 'Risk assessment results for businesses based on keyword detection';
COMMENT ON TABLE risk_keyword_relationships IS 'Relationships between risk keywords for enhanced detection';

-- Column comments for risk_keywords
COMMENT ON COLUMN risk_keywords.keyword IS 'The primary keyword or phrase to detect';
COMMENT ON COLUMN risk_keywords.risk_category IS 'Category of risk: illegal, prohibited, high_risk, tbml, sanctions, fraud';
COMMENT ON COLUMN risk_keywords.risk_severity IS 'Severity level: low, medium, high, critical';
COMMENT ON COLUMN risk_keywords.mcc_codes IS 'Associated prohibited MCC codes';
COMMENT ON COLUMN risk_keywords.naics_codes IS 'Associated prohibited NAICS codes';
COMMENT ON COLUMN risk_keywords.sic_codes IS 'Associated prohibited SIC codes';
COMMENT ON COLUMN risk_keywords.card_brand_restrictions IS 'Card brand restrictions (Visa, Mastercard, Amex)';
COMMENT ON COLUMN risk_keywords.detection_patterns IS 'Regex patterns for advanced detection';
COMMENT ON COLUMN risk_keywords.synonyms IS 'Alternative terms and variations';

-- Column comments for business_risk_assessments
COMMENT ON COLUMN business_risk_assessments.risk_score IS 'Calculated risk score between 0.00 and 1.00';
COMMENT ON COLUMN business_risk_assessments.risk_level IS 'Risk level: low, medium, high, critical';
COMMENT ON COLUMN business_risk_assessments.assessment_method IS 'Method used for assessment (keyword_matching, ml_model, etc.)';
COMMENT ON COLUMN business_risk_assessments.detected_patterns IS 'JSONB containing detected patterns and their details';

-- =====================================================
-- 10. Validation Functions
-- =====================================================

-- Function to validate risk keyword data
CREATE OR REPLACE FUNCTION validate_risk_keyword()
RETURNS TRIGGER AS $$
BEGIN
    -- Validate keyword is not empty
    IF NEW.keyword IS NULL OR TRIM(NEW.keyword) = '' THEN
        RAISE EXCEPTION 'Keyword cannot be empty';
    END IF;
    
    -- Validate risk score is within bounds
    IF NEW.risk_severity = 'critical' AND NEW.risk_category NOT IN ('illegal', 'prohibited') THEN
        RAISE WARNING 'Critical severity typically associated with illegal or prohibited categories';
    END IF;
    
    -- Validate MCC codes format if provided
    IF NEW.mcc_codes IS NOT NULL THEN
        FOR i IN 1..array_length(NEW.mcc_codes, 1) LOOP
            IF NEW.mcc_codes[i] !~ '^[0-9]{4}$' THEN
                RAISE EXCEPTION 'Invalid MCC code format: %', NEW.mcc_codes[i];
            END IF;
        END LOOP;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger for risk keyword validation
CREATE TRIGGER validate_risk_keyword_trigger
    BEFORE INSERT OR UPDATE ON risk_keywords
    FOR EACH ROW EXECUTE FUNCTION validate_risk_keyword();

-- =====================================================
-- 11. Migration Complete Verification
-- =====================================================

-- Verify schema creation
SELECT 
    table_name, 
    table_type,
    CASE 
        WHEN table_name IN ('risk_keywords', 'industry_code_crosswalks', 'business_risk_assessments', 'risk_keyword_relationships') 
        THEN '✅ Created'
        ELSE '❌ Missing'
    END as status
FROM information_schema.tables 
WHERE table_schema = 'public' 
    AND table_name IN (
        'risk_keywords', 
        'industry_code_crosswalks', 
        'business_risk_assessments', 
        'risk_keyword_relationships'
    )
ORDER BY table_name;

-- Verify indexes creation
SELECT 
    schemaname,
    tablename,
    indexname,
    indexdef
FROM pg_indexes 
WHERE schemaname = 'public' 
    AND tablename IN ('risk_keywords', 'industry_code_crosswalks', 'business_risk_assessments', 'risk_keyword_relationships')
ORDER BY tablename, indexname;

-- =====================================================
-- Migration Complete
-- =====================================================
