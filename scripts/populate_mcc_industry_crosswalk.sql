-- MCC to Industry Crosswalk Mapping Population Script
-- This script populates the crosswalk_mappings table with MCC to Industry mappings

-- Enable necessary extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Insert comprehensive MCC to Industry mappings
INSERT INTO crosswalk_mappings (
    id, source_code, source_system, target_code, target_system,
    confidence_score, validation_rules, is_valid, metadata, created_at, updated_at
) VALUES
-- Technology Industry Mappings
('mcc_5734_tech', '5734', 'MCC', '1', 'INDUSTRY', 0.95, 
 '[{"rule_type": "format", "rule_name": "mcc_format_validation", "rule_value": "4-digit numeric", "description": "MCC code must be 4-digit numeric format"}]'::jsonb,
 true, '{"mcc_description": "Computer Software Stores", "industry_name": "Technology", "industry_category": "traditional", "mapping_method": "keyword_similarity"}'::jsonb,
 NOW(), NOW()),

('mcc_7372_tech', '7372', 'MCC', '1', 'INDUSTRY', 0.92,
 '[{"rule_type": "format", "rule_name": "mcc_format_validation", "rule_value": "4-digit numeric", "description": "MCC code must be 4-digit numeric format"}]'::jsonb,
 true, '{"mcc_description": "Prepackaged Software", "industry_name": "Technology", "industry_category": "traditional", "mapping_method": "keyword_similarity"}'::jsonb,
 NOW(), NOW()),

('mcc_7373_tech', '7373', 'MCC', '1', 'INDUSTRY', 0.90,
 '[{"rule_type": "format", "rule_name": "mcc_format_validation", "rule_value": "4-digit numeric", "description": "MCC code must be 4-digit numeric format"}]'::jsonb,
 true, '{"mcc_description": "Computer Integrated Systems Design", "industry_name": "Technology", "industry_category": "traditional", "mapping_method": "keyword_similarity"}'::jsonb,
 NOW(), NOW()),

-- Healthcare Industry Mappings
('mcc_8062_health', '8062', 'MCC', '2', 'INDUSTRY', 0.98,
 '[{"rule_type": "format", "rule_name": "mcc_format_validation", "rule_value": "4-digit numeric", "description": "MCC code must be 4-digit numeric format"}]'::jsonb,
 true, '{"mcc_description": "Hospitals", "industry_name": "Healthcare", "industry_category": "traditional", "mapping_method": "keyword_similarity"}'::jsonb,
 NOW(), NOW()),

('mcc_5047_health', '5047', 'MCC', '2', 'INDUSTRY', 0.95,
 '[{"rule_type": "format", "rule_name": "mcc_format_validation", "rule_value": "4-digit numeric", "description": "MCC code must be 4-digit numeric format"}]'::jsonb,
 true, '{"mcc_description": "Medical, Dental, Ophthalmic, and Hospital Equipment and Supplies", "industry_name": "Healthcare", "industry_category": "traditional", "mapping_method": "keyword_similarity"}'::jsonb,
 NOW(), NOW()),

('mcc_5122_health', '5122', 'MCC', '2', 'INDUSTRY', 0.93,
 '[{"rule_type": "format", "rule_name": "mcc_format_validation", "rule_value": "4-digit numeric", "description": "MCC code must be 4-digit numeric format"}]'::jsonb,
 true, '{"mcc_description": "Drugs, Drug Proprietaries, and Druggist Sundries", "industry_name": "Healthcare", "industry_category": "traditional", "mapping_method": "keyword_similarity"}'::jsonb,
 NOW(), NOW()),

-- Financial Services Industry Mappings
('mcc_6010_finance', '6010', 'MCC', '3', 'INDUSTRY', 0.97,
 '[{"rule_type": "format", "rule_name": "mcc_format_validation", "rule_value": "4-digit numeric", "description": "MCC code must be 4-digit numeric format"}]'::jsonb,
 true, '{"mcc_description": "Financial Institutions - Merchandise, Services", "industry_name": "Financial Services", "industry_category": "traditional", "mapping_method": "keyword_similarity"}'::jsonb,
 NOW(), NOW()),

('mcc_6011_finance', '6011', 'MCC', '3', 'INDUSTRY', 0.96,
 '[{"rule_type": "format", "rule_name": "mcc_format_validation", "rule_value": "4-digit numeric", "description": "MCC code must be 4-digit numeric format"}]'::jsonb,
 true, '{"mcc_description": "ATM Transactions", "industry_name": "Financial Services", "industry_category": "traditional", "mapping_method": "keyword_similarity"}'::jsonb,
 NOW(), NOW()),

('mcc_6300_finance', '6300', 'MCC', '3', 'INDUSTRY', 0.94,
 '[{"rule_type": "format", "rule_name": "mcc_format_validation", "rule_value": "4-digit numeric", "description": "MCC code must be 4-digit numeric format"}]'::jsonb,
 true, '{"mcc_description": "Insurance Sales, Underwriting, and Premiums", "industry_name": "Financial Services", "industry_category": "traditional", "mapping_method": "keyword_similarity"}'::jsonb,
 NOW(), NOW()),

-- Retail Industry Mappings
('mcc_5310_retail', '5310', 'MCC', '4', 'INDUSTRY', 0.95,
 '[{"rule_type": "format", "rule_name": "mcc_format_validation", "rule_value": "4-digit numeric", "description": "MCC code must be 4-digit numeric format"}]'::jsonb,
 true, '{"mcc_description": "Discount Stores", "industry_name": "Retail", "industry_category": "traditional", "mapping_method": "keyword_similarity"}'::jsonb,
 NOW(), NOW()),

('mcc_5311_retail', '5311', 'MCC', '4', 'INDUSTRY', 0.93,
 '[{"rule_type": "format", "rule_name": "mcc_format_validation", "rule_value": "4-digit numeric", "description": "MCC code must be 4-digit numeric format"}]'::jsonb,
 true, '{"mcc_description": "Department Stores", "industry_name": "Retail", "industry_category": "traditional", "mapping_method": "keyword_similarity"}'::jsonb,
 NOW(), NOW()),

('mcc_5312_retail', '5312', 'MCC', '4', 'INDUSTRY', 0.91,
 '[{"rule_type": "format", "rule_name": "mcc_format_validation", "rule_value": "4-digit numeric", "description": "MCC code must be 4-digit numeric format"}]'::jsonb,
 true, '{"mcc_description": "Variety Stores", "industry_name": "Retail", "industry_category": "traditional", "mapping_method": "keyword_similarity"}'::jsonb,
 NOW(), NOW()),

-- Manufacturing Industry Mappings
('mcc_5085_manufacturing', '5085', 'MCC', '5', 'INDUSTRY', 0.94,
 '[{"rule_type": "format", "rule_name": "mcc_format_validation", "rule_value": "4-digit numeric", "description": "MCC code must be 4-digit numeric format"}]'::jsonb,
 true, '{"mcc_description": "Industrial Supplies", "industry_name": "Manufacturing", "industry_category": "traditional", "mapping_method": "keyword_similarity"}'::jsonb,
 NOW(), NOW()),

('mcc_5087_manufacturing', '5087', 'MCC', '5', 'INDUSTRY', 0.92,
 '[{"rule_type": "format", "rule_name": "mcc_format_validation", "rule_value": "4-digit numeric", "description": "MCC code must be 4-digit numeric format"}]'::jsonb,
 true, '{"mcc_description": "Service Establishment Equipment and Supplies", "industry_name": "Manufacturing", "industry_category": "traditional", "mapping_method": "keyword_similarity"}'::jsonb,
 NOW(), NOW()),

('mcc_5088_manufacturing', '5088', 'MCC', '5', 'INDUSTRY', 0.90,
 '[{"rule_type": "format", "rule_name": "mcc_format_validation", "rule_value": "4-digit numeric", "description": "MCC code must be 4-digit numeric format"}]'::jsonb,
 true, '{"mcc_description": "Transportation Equipment and Supplies", "industry_name": "Manufacturing", "industry_category": "traditional", "mapping_method": "keyword_similarity"}'::jsonb,
 NOW(), NOW())

ON CONFLICT (source_code, source_system, target_code, target_system)
DO UPDATE SET
    confidence_score = EXCLUDED.confidence_score,
    validation_rules = EXCLUDED.validation_rules,
    is_valid = EXCLUDED.is_valid,
    metadata = EXCLUDED.metadata,
    updated_at = EXCLUDED.updated_at;

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_crosswalk_mappings_source ON crosswalk_mappings(source_code, source_system);
CREATE INDEX IF NOT EXISTS idx_crosswalk_mappings_target ON crosswalk_mappings(target_code, target_system);
CREATE INDEX IF NOT EXISTS idx_crosswalk_mappings_confidence ON crosswalk_mappings(confidence_score);
CREATE INDEX IF NOT EXISTS idx_crosswalk_mappings_valid ON crosswalk_mappings(is_valid);

-- Verify the mappings were inserted
SELECT 
    source_code,
    source_system,
    target_code,
    target_system,
    confidence_score,
    is_valid,
    metadata->>'industry_name' as industry_name
FROM crosswalk_mappings 
WHERE source_system = 'MCC' 
ORDER BY confidence_score DESC;
