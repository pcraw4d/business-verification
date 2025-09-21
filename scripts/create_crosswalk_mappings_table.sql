-- Create crosswalk_mappings table for classification alignment
-- This table stores mappings between different classification systems and industries

-- Enable necessary extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create crosswalk_mappings table
CREATE TABLE IF NOT EXISTS crosswalk_mappings (
    id SERIAL PRIMARY KEY,
    industry_id INTEGER NOT NULL REFERENCES industries(id) ON DELETE CASCADE,
    source_code VARCHAR(20) NOT NULL,
    source_system VARCHAR(10) NOT NULL CHECK (source_system IN ('MCC', 'NAICS', 'SIC')),
    target_code VARCHAR(20),
    target_system VARCHAR(20) DEFAULT 'INDUSTRY',
    mcc_code VARCHAR(4),
    naics_code VARCHAR(6),
    sic_code VARCHAR(4),
    description TEXT,
    confidence_score DECIMAL(3,2) DEFAULT 0.80 CHECK (confidence_score >= 0.0 AND confidence_score <= 1.0),
    validation_rules JSONB,
    is_valid BOOLEAN DEFAULT true,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(source_code, source_system, industry_id)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_crosswalk_mappings_industry_id ON crosswalk_mappings(industry_id);
CREATE INDEX IF NOT EXISTS idx_crosswalk_mappings_source ON crosswalk_mappings(source_code, source_system);
CREATE INDEX IF NOT EXISTS idx_crosswalk_mappings_target ON crosswalk_mappings(target_code, target_system);
CREATE INDEX IF NOT EXISTS idx_crosswalk_mappings_mcc_code ON crosswalk_mappings(mcc_code);
CREATE INDEX IF NOT EXISTS idx_crosswalk_mappings_naics_code ON crosswalk_mappings(naics_code);
CREATE INDEX IF NOT EXISTS idx_crosswalk_mappings_sic_code ON crosswalk_mappings(sic_code);
CREATE INDEX IF NOT EXISTS idx_crosswalk_mappings_confidence_score ON crosswalk_mappings(confidence_score);
CREATE INDEX IF NOT EXISTS idx_crosswalk_mappings_valid ON crosswalk_mappings(is_valid);
CREATE INDEX IF NOT EXISTS idx_crosswalk_mappings_metadata ON crosswalk_mappings USING gin(metadata);

-- Create trigger for updated_at
CREATE TRIGGER update_crosswalk_mappings_updated_at 
    BEFORE UPDATE ON crosswalk_mappings 
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create views for easier querying
CREATE OR REPLACE VIEW crosswalk_alignment_summary AS
SELECT 
    i.id as industry_id,
    i.name as industry_name,
    COUNT(DISTINCT CASE WHEN cm.mcc_code IS NOT NULL THEN cm.mcc_code END) as mcc_count,
    COUNT(DISTINCT CASE WHEN cm.naics_code IS NOT NULL THEN cm.naics_code END) as naics_count,
    COUNT(DISTINCT CASE WHEN cm.sic_code IS NOT NULL THEN cm.sic_code END) as sic_count,
    ROUND(AVG(cm.confidence_score), 3) as avg_confidence_score,
    COUNT(CASE WHEN cm.is_valid = true THEN 1 END) as valid_mappings,
    COUNT(CASE WHEN cm.is_valid = false THEN 1 END) as invalid_mappings,
    COUNT(cm.id) as total_mappings
FROM industries i
LEFT JOIN crosswalk_mappings cm ON i.id = cm.industry_id
WHERE i.is_active = true
GROUP BY i.id, i.name
ORDER BY i.name;

CREATE OR REPLACE VIEW classification_coverage_analysis AS
SELECT 
    classification_system,
    industry_name,
    code_count,
    avg_confidence,
    alignment_quality
FROM (
    SELECT 
        'MCC' as classification_system,
        i.name as industry_name,
        COUNT(DISTINCT cm.mcc_code) as code_count,
        ROUND(AVG(cm.confidence_score), 3) as avg_confidence,
        CASE 
            WHEN AVG(cm.confidence_score) >= 0.9 THEN 'Excellent'
            WHEN AVG(cm.confidence_score) >= 0.8 THEN 'Good'
            WHEN AVG(cm.confidence_score) >= 0.7 THEN 'Fair'
            ELSE 'Poor'
        END as alignment_quality
    FROM industries i
    LEFT JOIN crosswalk_mappings cm ON i.id = cm.industry_id AND cm.mcc_code IS NOT NULL
    WHERE i.is_active = true
    GROUP BY i.name
    
    UNION ALL
    
    SELECT 
        'NAICS' as classification_system,
        i.name as industry_name,
        COUNT(DISTINCT cm.naics_code) as code_count,
        ROUND(AVG(cm.confidence_score), 3) as avg_confidence,
        CASE 
            WHEN AVG(cm.confidence_score) >= 0.9 THEN 'Excellent'
            WHEN AVG(cm.confidence_score) >= 0.8 THEN 'Good'
            WHEN AVG(cm.confidence_score) >= 0.7 THEN 'Fair'
            ELSE 'Poor'
        END as alignment_quality
    FROM industries i
    LEFT JOIN crosswalk_mappings cm ON i.id = cm.industry_id AND cm.naics_code IS NOT NULL
    WHERE i.is_active = true
    GROUP BY i.name
    
    UNION ALL
    
    SELECT 
        'SIC' as classification_system,
        i.name as industry_name,
        COUNT(DISTINCT cm.sic_code) as code_count,
        ROUND(AVG(cm.confidence_score), 3) as avg_confidence,
        CASE 
            WHEN AVG(cm.confidence_score) >= 0.9 THEN 'Excellent'
            WHEN AVG(cm.confidence_score) >= 0.8 THEN 'Good'
            WHEN AVG(cm.confidence_score) >= 0.7 THEN 'Fair'
            ELSE 'Poor'
        END as alignment_quality
    FROM industries i
    LEFT JOIN crosswalk_mappings cm ON i.id = cm.industry_id AND cm.sic_code IS NOT NULL
    WHERE i.is_active = true
    GROUP BY i.name
) AS combined_analysis
ORDER BY classification_system, industry_name;

-- Comments for documentation
COMMENT ON TABLE crosswalk_mappings IS 'Mappings between classification systems (MCC, NAICS, SIC) and industries';
COMMENT ON COLUMN crosswalk_mappings.industry_id IS 'Reference to the industries table';
COMMENT ON COLUMN crosswalk_mappings.source_code IS 'The original classification code';
COMMENT ON COLUMN crosswalk_mappings.source_system IS 'The classification system (MCC, NAICS, SIC)';
COMMENT ON COLUMN crosswalk_mappings.target_code IS 'The target code (usually industry ID)';
COMMENT ON COLUMN crosswalk_mappings.target_system IS 'The target system (usually INDUSTRY)';
COMMENT ON COLUMN crosswalk_mappings.mcc_code IS 'Merchant Category Code if applicable';
COMMENT ON COLUMN crosswalk_mappings.naics_code IS 'North American Industry Classification System code if applicable';
COMMENT ON COLUMN crosswalk_mappings.sic_code IS 'Standard Industrial Classification code if applicable';
COMMENT ON COLUMN crosswalk_mappings.confidence_score IS 'Confidence score for the mapping (0.0 to 1.0)';
COMMENT ON COLUMN crosswalk_mappings.validation_rules IS 'JSON array of validation rules applied to this mapping';
COMMENT ON COLUMN crosswalk_mappings.is_valid IS 'Whether this mapping has passed validation';
COMMENT ON COLUMN crosswalk_mappings.metadata IS 'Additional metadata for the mapping';

