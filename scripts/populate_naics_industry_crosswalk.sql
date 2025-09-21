-- NAICS to Industry Crosswalk Mapping Population Script
-- This script populates the crosswalk_mappings table with NAICS to Industry mappings

-- Enable necessary extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Insert comprehensive NAICS to Industry mappings
INSERT INTO crosswalk_mappings (
    id, source_code, source_system, target_code, target_system,
    confidence_score, validation_rules, is_valid, metadata, created_at, updated_at
) VALUES
-- Technology Industry Mappings (Sector 51 - Information)
('naics_511210_tech', '511210', 'NAICS', '1', 'INDUSTRY', 0.95,
 '[{"rule_type": "format", "rule_name": "naics_format_validation", "rule_value": "6-digit numeric", "description": "NAICS code must be 6-digit numeric format"}, {"rule_type": "hierarchy", "rule_name": "naics_hierarchy_validation", "rule_value": "valid_sector", "description": "NAICS code must have valid sector code"}]'::jsonb,
 true, '{"naics_description": "Software Publishers", "industry_name": "Technology", "industry_category": "traditional", "mapping_method": "keyword_similarity", "sector": "51"}'::jsonb,
 NOW(), NOW()),

('naics_518210_tech', '518210', 'NAICS', '1', 'INDUSTRY', 0.92,
 '[{"rule_type": "format", "rule_name": "naics_format_validation", "rule_value": "6-digit numeric", "description": "NAICS code must be 6-digit numeric format"}, {"rule_type": "hierarchy", "rule_name": "naics_hierarchy_validation", "rule_value": "valid_sector", "description": "NAICS code must have valid sector code"}]'::jsonb,
 true, '{"naics_description": "Data Processing, Hosting, and Related Services", "industry_name": "Technology", "industry_category": "traditional", "mapping_method": "keyword_similarity", "sector": "51"}'::jsonb,
 NOW(), NOW()),

('naics_541511_tech', '541511', 'NAICS', '1', 'INDUSTRY', 0.90,
 '[{"rule_type": "format", "rule_name": "naics_format_validation", "rule_value": "6-digit numeric", "description": "NAICS code must be 6-digit numeric format"}, {"rule_type": "hierarchy", "rule_name": "naics_hierarchy_validation", "rule_value": "valid_sector", "description": "NAICS code must have valid sector code"}]'::jsonb,
 true, '{"naics_description": "Custom Computer Programming Services", "industry_name": "Technology", "industry_category": "traditional", "mapping_method": "keyword_similarity", "sector": "54"}'::jsonb,
 NOW(), NOW()),

-- Healthcare Industry Mappings (Sector 62 - Health Care and Social Assistance)
('naics_622110_health', '622110', 'NAICS', '2', 'INDUSTRY', 0.98,
 '[{"rule_type": "format", "rule_name": "naics_format_validation", "rule_value": "6-digit numeric", "description": "NAICS code must be 6-digit numeric format"}, {"rule_type": "hierarchy", "rule_name": "naics_hierarchy_validation", "rule_value": "valid_sector", "description": "NAICS code must have valid sector code"}]'::jsonb,
 true, '{"naics_description": "General Medical and Surgical Hospitals", "industry_name": "Healthcare", "industry_category": "traditional", "mapping_method": "keyword_similarity", "sector": "62"}'::jsonb,
 NOW(), NOW()),

('naics_621111_health', '621111', 'NAICS', '2', 'INDUSTRY', 0.95,
 '[{"rule_type": "format", "rule_name": "naics_format_validation", "rule_value": "6-digit numeric", "description": "NAICS code must be 6-digit numeric format"}, {"rule_type": "hierarchy", "rule_name": "naics_hierarchy_validation", "rule_value": "valid_sector", "description": "NAICS code must have valid sector code"}]'::jsonb,
 true, '{"naics_description": "Offices of Physicians (except Mental Health Specialists)", "industry_name": "Healthcare", "industry_category": "traditional", "mapping_method": "keyword_similarity", "sector": "62"}'::jsonb,
 NOW(), NOW()),

('naics_325412_health', '325412', 'NAICS', '2', 'INDUSTRY', 0.93,
 '[{"rule_type": "format", "rule_name": "naics_format_validation", "rule_value": "6-digit numeric", "description": "NAICS code must be 6-digit numeric format"}, {"rule_type": "hierarchy", "rule_name": "naics_hierarchy_validation", "rule_value": "valid_sector", "description": "NAICS code must have valid sector code"}]'::jsonb,
 true, '{"naics_description": "Pharmaceutical Preparation Manufacturing", "industry_name": "Healthcare", "industry_category": "traditional", "mapping_method": "keyword_similarity", "sector": "32"}'::jsonb,
 NOW(), NOW()),

-- Financial Services Industry Mappings (Sector 52 - Finance and Insurance)
('naics_522110_finance', '522110', 'NAICS', '3', 'INDUSTRY', 0.97,
 '[{"rule_type": "format", "rule_name": "naics_format_validation", "rule_value": "6-digit numeric", "description": "NAICS code must be 6-digit numeric format"}, {"rule_type": "hierarchy", "rule_name": "naics_hierarchy_validation", "rule_value": "valid_sector", "description": "NAICS code must have valid sector code"}]'::jsonb,
 true, '{"naics_description": "Commercial Banking", "industry_name": "Financial Services", "industry_category": "traditional", "mapping_method": "keyword_similarity", "sector": "52"}'::jsonb,
 NOW(), NOW()),

('naics_523110_finance', '523110', 'NAICS', '3', 'INDUSTRY', 0.96,
 '[{"rule_type": "format", "rule_name": "naics_format_validation", "rule_value": "6-digit numeric", "description": "NAICS code must be 6-digit numeric format"}, {"rule_type": "hierarchy", "rule_name": "naics_hierarchy_validation", "rule_value": "valid_sector", "description": "NAICS code must have valid sector code"}]'::jsonb,
 true, '{"naics_description": "Investment Banking and Securities Dealing", "industry_name": "Financial Services", "industry_category": "traditional", "mapping_method": "keyword_similarity", "sector": "52"}'::jsonb,
 NOW(), NOW()),

('naics_524113_finance', '524113', 'NAICS', '3', 'INDUSTRY', 0.94,
 '[{"rule_type": "format", "rule_name": "naics_format_validation", "rule_value": "6-digit numeric", "description": "NAICS code must be 6-digit numeric format"}, {"rule_type": "hierarchy", "rule_name": "naics_hierarchy_validation", "rule_value": "valid_sector", "description": "NAICS code must have valid sector code"}]'::jsonb,
 true, '{"naics_description": "Direct Life Insurance Carriers", "industry_name": "Financial Services", "industry_category": "traditional", "mapping_method": "keyword_similarity", "sector": "52"}'::jsonb,
 NOW(), NOW()),

-- Retail Industry Mappings (Sectors 44-45 - Retail Trade)
('naics_441110_retail', '441110', 'NAICS', '4', 'INDUSTRY', 0.95,
 '[{"rule_type": "format", "rule_name": "naics_format_validation", "rule_value": "6-digit numeric", "description": "NAICS code must be 6-digit numeric format"}, {"rule_type": "hierarchy", "rule_name": "naics_hierarchy_validation", "rule_value": "valid_sector", "description": "NAICS code must have valid sector code"}]'::jsonb,
 true, '{"naics_description": "New Car Dealers", "industry_name": "Retail", "industry_category": "traditional", "mapping_method": "keyword_similarity", "sector": "44"}'::jsonb,
 NOW(), NOW()),

('naics_452111_retail', '452111', 'NAICS', '4', 'INDUSTRY', 0.93,
 '[{"rule_type": "format", "rule_name": "naics_format_validation", "rule_value": "6-digit numeric", "description": "NAICS code must be 6-digit numeric format"}, {"rule_type": "hierarchy", "rule_name": "naics_hierarchy_validation", "rule_value": "valid_sector", "description": "NAICS code must have valid sector code"}]'::jsonb,
 true, '{"naics_description": "Department Stores", "industry_name": "Retail", "industry_category": "traditional", "mapping_method": "keyword_similarity", "sector": "45"}'::jsonb,
 NOW(), NOW()),

('naics_454111_retail', '454111', 'NAICS', '4', 'INDUSTRY', 0.91,
 '[{"rule_type": "format", "rule_name": "naics_format_validation", "rule_value": "6-digit numeric", "description": "NAICS code must be 6-digit numeric format"}, {"rule_type": "hierarchy", "rule_name": "naics_hierarchy_validation", "rule_value": "valid_sector", "description": "NAICS code must have valid sector code"}]'::jsonb,
 true, '{"naics_description": "Electronic Shopping", "industry_name": "Retail", "industry_category": "traditional", "mapping_method": "keyword_similarity", "sector": "45"}'::jsonb,
 NOW(), NOW()),

-- Manufacturing Industry Mappings (Sectors 31-33 - Manufacturing)
('naics_311111_manufacturing', '311111', 'NAICS', '5', 'INDUSTRY', 0.94,
 '[{"rule_type": "format", "rule_name": "naics_format_validation", "rule_value": "6-digit numeric", "description": "NAICS code must be 6-digit numeric format"}, {"rule_type": "hierarchy", "rule_name": "naics_hierarchy_validation", "rule_value": "valid_sector", "description": "NAICS code must have valid sector code"}]'::jsonb,
 true, '{"naics_description": "Dog and Cat Food Manufacturing", "industry_name": "Manufacturing", "industry_category": "traditional", "mapping_method": "keyword_similarity", "sector": "31"}'::jsonb,
 NOW(), NOW()),

('naics_325110_manufacturing', '325110', 'NAICS', '5', 'INDUSTRY', 0.92,
 '[{"rule_type": "format", "rule_name": "naics_format_validation", "rule_value": "6-digit numeric", "description": "NAICS code must be 6-digit numeric format"}, {"rule_type": "hierarchy", "rule_name": "naics_hierarchy_validation", "rule_value": "valid_sector", "description": "NAICS code must have valid sector code"}]'::jsonb,
 true, '{"naics_description": "Petrochemical Manufacturing", "industry_name": "Manufacturing", "industry_category": "traditional", "mapping_method": "keyword_similarity", "sector": "32"}'::jsonb,
 NOW(), NOW()),

('naics_336111_manufacturing', '336111', 'NAICS', '5', 'INDUSTRY', 0.90,
 '[{"rule_type": "format", "rule_name": "naics_format_validation", "rule_value": "6-digit numeric", "description": "NAICS code must be 6-digit numeric format"}, {"rule_type": "hierarchy", "rule_name": "naics_hierarchy_validation", "rule_value": "valid_sector", "description": "NAICS code must have valid sector code"}]'::jsonb,
 true, '{"naics_description": "Automobile Manufacturing", "industry_name": "Manufacturing", "industry_category": "traditional", "mapping_method": "keyword_similarity", "sector": "33"}'::jsonb,
 NOW(), NOW())

ON CONFLICT (source_code, source_system, target_code, target_system)
DO UPDATE SET
    confidence_score = EXCLUDED.confidence_score,
    validation_rules = EXCLUDED.validation_rules,
    is_valid = EXCLUDED.is_valid,
    metadata = EXCLUDED.metadata,
    updated_at = EXCLUDED.updated_at;

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_crosswalk_mappings_naics_source ON crosswalk_mappings(source_code, source_system) WHERE source_system = 'NAICS';
CREATE INDEX IF NOT EXISTS idx_crosswalk_mappings_naics_confidence ON crosswalk_mappings(confidence_score) WHERE source_system = 'NAICS';

-- Verify the mappings were inserted
SELECT 
    source_code,
    source_system,
    target_code,
    target_system,
    confidence_score,
    is_valid,
    metadata->>'industry_name' as industry_name,
    metadata->>'sector' as naics_sector
FROM crosswalk_mappings 
WHERE source_system = 'NAICS' 
ORDER BY confidence_score DESC;
