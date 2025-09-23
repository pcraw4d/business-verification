-- =====================================================
-- Enhanced Classification Migration Script (SAFE VERSION)
-- KYB Platform - Task 1.5.1
-- =====================================================
-- 
-- This script creates an enhanced classification system that builds upon
-- existing classification tables and adds comprehensive risk management,
-- code crosswalks, and advanced indexing for optimal performance.
--
-- Author: KYB Platform Development Team
-- Date: January 19, 2025
-- Version: 1.1 (Safe)
-- 
-- Dependencies:
-- - Existing classification tables (industries, industry_keywords, etc.)
-- =====================================================

-- Enable necessary extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
CREATE EXTENSION IF NOT EXISTS "btree_gin";

-- =====================================================
-- 1. ENHANCED RISK KEYWORDS TABLE
-- =====================================================

CREATE TABLE IF NOT EXISTS risk_keywords (
    id SERIAL PRIMARY KEY,
    keyword VARCHAR(255) NOT NULL,
    risk_category VARCHAR(50) NOT NULL CHECK (risk_category IN (
        'illegal', 'prohibited', 'high_risk', 'tbml', 'sanctions', 'fraud', 'regulatory'
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
    risk_score_weight DECIMAL(3,2) DEFAULT 1.00 CHECK (risk_score_weight >= 0.00 AND risk_score_weight <= 2.00),
    detection_confidence DECIMAL(3,2) DEFAULT 0.80 CHECK (detection_confidence >= 0.00 AND detection_confidence <= 1.00),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(keyword, risk_category)
);

-- =====================================================
-- 2. INDUSTRY CODE CROSSWALKS TABLE
-- =====================================================

CREATE TABLE IF NOT EXISTS industry_code_crosswalks (
    id SERIAL PRIMARY KEY,
    source_system VARCHAR(20) NOT NULL CHECK (source_system IN ('NAICS', 'SIC', 'MCC', 'ISIC', 'NACE')),
    source_code VARCHAR(20) NOT NULL,
    target_system VARCHAR(20) NOT NULL CHECK (target_system IN ('NAICS', 'SIC', 'MCC', 'ISIC', 'NACE')),
    target_code VARCHAR(20) NOT NULL,
    confidence_score DECIMAL(3,2) DEFAULT 0.80 CHECK (confidence_score >= 0.00 AND confidence_score <= 1.00),
    mapping_type VARCHAR(20) DEFAULT 'direct' CHECK (mapping_type IN ('direct', 'approximate', 'hierarchical')),
    notes TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(source_system, source_code, target_system, target_code)
);

-- =====================================================
-- 3. BUSINESS RISK ASSESSMENTS TABLE
-- =====================================================

CREATE TABLE IF NOT EXISTS business_risk_assessments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    business_id VARCHAR(255) NOT NULL,
    business_name VARCHAR(500) NOT NULL,
    risk_score DECIMAL(3,2) NOT NULL CHECK (risk_score >= 0.00 AND risk_score <= 1.00),
    risk_level VARCHAR(20) NOT NULL CHECK (risk_level IN ('low', 'medium', 'high', 'critical')),
    risk_factors JSONB NOT NULL DEFAULT '{}',
    prohibited_keywords_found TEXT[],
    sanctions_matches TEXT[],
    regulatory_concerns TEXT[],
    geographic_risk_factors TEXT[],
    industry_risk_factors TEXT[],
    assessment_methodology VARCHAR(50) DEFAULT 'automated' CHECK (assessment_methodology IN ('automated', 'manual', 'hybrid')),
    assessor_id VARCHAR(255),
    assessment_notes TEXT,
    review_required BOOLEAN DEFAULT false,
    review_date TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(business_id)
);

-- =====================================================
-- 4. RISK KEYWORD RELATIONSHIPS TABLE
-- =====================================================

CREATE TABLE IF NOT EXISTS risk_keyword_relationships (
    id SERIAL PRIMARY KEY,
    primary_keyword_id INTEGER NOT NULL REFERENCES risk_keywords(id) ON DELETE CASCADE,
    related_keyword_id INTEGER NOT NULL REFERENCES risk_keywords(id) ON DELETE CASCADE,
    relationship_type VARCHAR(20) NOT NULL CHECK (relationship_type IN ('synonym', 'related', 'opposite', 'hierarchical')),
    strength DECIMAL(3,2) DEFAULT 0.50 CHECK (strength >= 0.00 AND strength <= 1.00),
    context TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(primary_keyword_id, related_keyword_id, relationship_type)
);

-- =====================================================
-- 5. CLASSIFICATION PERFORMANCE METRICS TABLE
-- =====================================================

CREATE TABLE IF NOT EXISTS classification_performance_metrics (
    id SERIAL PRIMARY KEY,
    metric_date DATE NOT NULL DEFAULT CURRENT_DATE,
    total_classifications INTEGER DEFAULT 0,
    successful_classifications INTEGER DEFAULT 0,
    failed_classifications INTEGER DEFAULT 0,
    accuracy_percentage DECIMAL(5,2) DEFAULT 0.00,
    average_confidence_score DECIMAL(3,2) DEFAULT 0.00,
    processing_time_avg_ms INTEGER DEFAULT 0,
    risk_assessments_completed INTEGER DEFAULT 0,
    high_risk_businesses_detected INTEGER DEFAULT 0,
    false_positive_rate DECIMAL(5,2) DEFAULT 0.00,
    false_negative_rate DECIMAL(5,2) DEFAULT 0.00,
    industry_breakdown JSONB DEFAULT '{}',
    risk_level_breakdown JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(metric_date)
);

-- =====================================================
-- INDEXES FOR PERFORMANCE
-- =====================================================

-- Risk keywords indexes
CREATE INDEX IF NOT EXISTS idx_risk_keywords_keyword ON risk_keywords(keyword);
CREATE INDEX IF NOT EXISTS idx_risk_keywords_risk_category ON risk_keywords(risk_category);
CREATE INDEX IF NOT EXISTS idx_risk_keywords_risk_severity ON risk_keywords(risk_severity);
CREATE INDEX IF NOT EXISTS idx_risk_keywords_active ON risk_keywords(is_active);
CREATE INDEX IF NOT EXISTS idx_risk_keywords_gin ON risk_keywords USING gin(synonyms);

-- Industry code crosswalks indexes
CREATE INDEX IF NOT EXISTS idx_industry_code_crosswalks_source ON industry_code_crosswalks(source_system, source_code);
CREATE INDEX IF NOT EXISTS idx_industry_code_crosswalks_target ON industry_code_crosswalks(target_system, target_code);
CREATE INDEX IF NOT EXISTS idx_industry_code_crosswalks_active ON industry_code_crosswalks(is_active);

-- Business risk assessments indexes
CREATE INDEX IF NOT EXISTS idx_business_risk_assessments_business_id ON business_risk_assessments(business_id);
CREATE INDEX IF NOT EXISTS idx_business_risk_assessments_risk_level ON business_risk_assessments(risk_level);
CREATE INDEX IF NOT EXISTS idx_business_risk_assessments_risk_score ON business_risk_assessments(risk_score);
CREATE INDEX IF NOT EXISTS idx_business_risk_assessments_active ON business_risk_assessments(is_active);
CREATE INDEX IF NOT EXISTS idx_business_risk_assessments_gin ON business_risk_assessments USING gin(risk_factors);

-- Risk keyword relationships indexes
CREATE INDEX IF NOT EXISTS idx_risk_keyword_relationships_primary ON risk_keyword_relationships(primary_keyword_id);
CREATE INDEX IF NOT EXISTS idx_risk_keyword_relationships_related ON risk_keyword_relationships(related_keyword_id);
CREATE INDEX IF NOT EXISTS idx_risk_keyword_relationships_type ON risk_keyword_relationships(relationship_type);
CREATE INDEX IF NOT EXISTS idx_risk_keyword_relationships_active ON risk_keyword_relationships(is_active);

-- Classification performance metrics indexes
CREATE INDEX IF NOT EXISTS idx_classification_performance_metrics_date ON classification_performance_metrics(metric_date);
CREATE INDEX IF NOT EXISTS idx_classification_performance_metrics_accuracy ON classification_performance_metrics(accuracy_percentage);

-- =====================================================
-- SAMPLE DATA INSERTION (SAFE)
-- =====================================================

-- Insert sample risk keywords
INSERT INTO risk_keywords (keyword, risk_category, risk_severity, description, risk_score_weight, detection_confidence) VALUES
-- Illegal/Prohibited keywords
('gambling', 'prohibited', 'high', 'Gambling and betting services', 1.5, 0.95),
('casino', 'prohibited', 'high', 'Casino and gaming operations', 1.5, 0.95),
('lottery', 'prohibited', 'medium', 'Lottery and sweepstakes', 1.2, 0.90),
('betting', 'prohibited', 'high', 'Sports betting and wagering', 1.5, 0.95),

-- High-risk keywords
('cryptocurrency', 'high_risk', 'medium', 'Cryptocurrency and digital assets', 1.3, 0.85),
('bitcoin', 'high_risk', 'medium', 'Bitcoin and cryptocurrency trading', 1.3, 0.85),
('forex', 'high_risk', 'medium', 'Foreign exchange trading', 1.2, 0.80),
('trading', 'high_risk', 'low', 'General trading activities', 1.1, 0.75),

-- TBML keywords
('money transfer', 'tbml', 'high', 'Money transfer and remittance services', 1.4, 0.90),
('wire transfer', 'tbml', 'high', 'Wire transfer services', 1.4, 0.90),
('cash advance', 'tbml', 'medium', 'Cash advance services', 1.2, 0.85),

-- Sanctions keywords
('iran', 'sanctions', 'critical', 'Iran-related business activities', 2.0, 0.99),
('north korea', 'sanctions', 'critical', 'North Korea-related business activities', 2.0, 0.99),
('cuba', 'sanctions', 'critical', 'Cuba-related business activities', 2.0, 0.99),

-- Fraud keywords
('phishing', 'fraud', 'critical', 'Phishing and identity theft', 1.8, 0.95),
('scam', 'fraud', 'high', 'Scam and fraudulent activities', 1.6, 0.90),
('identity theft', 'fraud', 'critical', 'Identity theft services', 1.8, 0.95)
ON CONFLICT (keyword, risk_category) DO NOTHING;

-- Insert sample industry code crosswalks
INSERT INTO industry_code_crosswalks (source_system, source_code, target_system, target_code, confidence_score, mapping_type) VALUES
-- NAICS to SIC mappings
('NAICS', '541511', 'SIC', '7372', 0.95, 'direct'),
('NAICS', '454110', 'SIC', '5961', 0.90, 'direct'),
('NAICS', '621111', 'SIC', '8011', 0.95, 'direct'),
('NAICS', '522110', 'SIC', '6021', 0.90, 'direct'),

-- SIC to MCC mappings
('SIC', '7372', 'MCC', '7372', 0.95, 'direct'),
('SIC', '5961', 'MCC', '5311', 0.85, 'approximate'),
('SIC', '8011', 'MCC', '8062', 0.90, 'direct'),
('SIC', '6021', 'MCC', '6010', 0.90, 'direct'),

-- NAICS to MCC mappings
('NAICS', '541511', 'MCC', '7372', 0.90, 'direct'),
('NAICS', '454110', 'MCC', '5311', 0.80, 'approximate'),
('NAICS', '621111', 'MCC', '8062', 0.85, 'direct'),
('NAICS', '522110', 'MCC', '6010', 0.85, 'direct')
ON CONFLICT (source_system, source_code, target_system, target_code) DO NOTHING;

-- Insert sample business risk assessments
INSERT INTO business_risk_assessments (business_id, business_name, risk_score, risk_level, risk_factors, assessment_methodology) VALUES
('biz_001', 'Safe Technology Corp', 0.15, 'low', '{"industry": "technology", "geographic": "low_risk", "regulatory": "compliant"}', 'automated'),
('biz_002', 'High Risk Trading LLC', 0.85, 'high', '{"industry": "trading", "geographic": "medium_risk", "regulatory": "requires_review"}', 'automated'),
('biz_003', 'Prohibited Casino Inc', 0.95, 'critical', '{"industry": "gambling", "geographic": "high_risk", "regulatory": "prohibited"}', 'automated')
ON CONFLICT (business_id) DO NOTHING;

-- Insert sample risk keyword relationships
INSERT INTO risk_keyword_relationships (primary_keyword_id, related_keyword_id, relationship_type, strength) VALUES
(1, 2, 'related', 0.80), -- gambling related to casino
(1, 3, 'related', 0.70), -- gambling related to lottery
(2, 3, 'related', 0.60), -- casino related to lottery
(5, 6, 'synonym', 0.90), -- cryptocurrency synonym to bitcoin
(9, 10, 'synonym', 0.85), -- money transfer synonym to wire transfer
(13, 14, 'related', 0.75), -- iran related to north korea
(15, 16, 'related', 0.70) -- phishing related to scam
ON CONFLICT (primary_keyword_id, related_keyword_id, relationship_type) DO NOTHING;

-- Insert sample classification performance metrics
INSERT INTO classification_performance_metrics (metric_date, total_classifications, successful_classifications, accuracy_percentage, average_confidence_score) VALUES
(CURRENT_DATE, 100, 95, 95.00, 0.87),
(CURRENT_DATE - INTERVAL '1 day', 85, 82, 96.47, 0.89),
(CURRENT_DATE - INTERVAL '2 days', 120, 115, 95.83, 0.85)
ON CONFLICT (metric_date) DO NOTHING;

-- =====================================================
-- ROW LEVEL SECURITY (RLS) POLICIES
-- =====================================================

-- Enable RLS on all tables
ALTER TABLE risk_keywords ENABLE ROW LEVEL SECURITY;
ALTER TABLE industry_code_crosswalks ENABLE ROW LEVEL SECURITY;
ALTER TABLE business_risk_assessments ENABLE ROW LEVEL SECURITY;
ALTER TABLE risk_keyword_relationships ENABLE ROW LEVEL SECURITY;
ALTER TABLE classification_performance_metrics ENABLE ROW LEVEL SECURITY;

-- Create policies for public read access
DROP POLICY IF EXISTS "Enable read access for all users" ON risk_keywords;
CREATE POLICY "Enable read access for all users" ON risk_keywords FOR SELECT USING (true);

DROP POLICY IF EXISTS "Enable read access for all users" ON industry_code_crosswalks;
CREATE POLICY "Enable read access for all users" ON industry_code_crosswalks FOR SELECT USING (true);

DROP POLICY IF EXISTS "Enable read access for all users" ON business_risk_assessments;
CREATE POLICY "Enable read access for all users" ON business_risk_assessments FOR SELECT USING (true);

DROP POLICY IF EXISTS "Enable read access for all users" ON risk_keyword_relationships;
CREATE POLICY "Enable read access for all users" ON risk_keyword_relationships FOR SELECT USING (true);

DROP POLICY IF EXISTS "Enable read access for all users" ON classification_performance_metrics;
CREATE POLICY "Enable read access for all users" ON classification_performance_metrics FOR SELECT USING (true);

-- =====================================================
-- TRIGGERS FOR UPDATED_AT TIMESTAMPS
-- =====================================================

-- Create triggers for updated_at (safe to run multiple times)
DROP TRIGGER IF EXISTS update_risk_keywords_updated_at ON risk_keywords;
CREATE TRIGGER update_risk_keywords_updated_at BEFORE UPDATE ON risk_keywords FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_industry_code_crosswalks_updated_at ON industry_code_crosswalks;
CREATE TRIGGER update_industry_code_crosswalks_updated_at BEFORE UPDATE ON industry_code_crosswalks FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_business_risk_assessments_updated_at ON business_risk_assessments;
CREATE TRIGGER update_business_risk_assessments_updated_at BEFORE UPDATE ON business_risk_assessments FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_risk_keyword_relationships_updated_at ON risk_keyword_relationships;
CREATE TRIGGER update_risk_keyword_relationships_updated_at BEFORE UPDATE ON risk_keyword_relationships FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

DROP TRIGGER IF EXISTS update_classification_performance_metrics_updated_at ON classification_performance_metrics;
CREATE TRIGGER update_classification_performance_metrics_updated_at BEFORE UPDATE ON classification_performance_metrics FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- =====================================================
-- VERIFICATION QUERIES
-- =====================================================

-- Verify tables were created
SELECT 'risk_keywords' as table_name, COUNT(*) as row_count FROM risk_keywords
UNION ALL
SELECT 'industry_code_crosswalks', COUNT(*) FROM industry_code_crosswalks
UNION ALL
SELECT 'business_risk_assessments', COUNT(*) FROM business_risk_assessments
UNION ALL
SELECT 'risk_keyword_relationships', COUNT(*) FROM risk_keyword_relationships
UNION ALL
SELECT 'classification_performance_metrics', COUNT(*) FROM classification_performance_metrics;

-- Show sample data
SELECT 'Sample Risk Keywords:' as info;
SELECT keyword, risk_category, risk_severity, risk_score_weight FROM risk_keywords LIMIT 10;

SELECT 'Sample Business Risk Assessments:' as info;
SELECT business_name, risk_level, risk_score FROM business_risk_assessments LIMIT 5;
