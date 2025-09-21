-- SIC to Industry Crosswalk Mapping Population Script
-- This script populates the crosswalk_mappings table with SIC to Industry mappings

-- Enable necessary extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Insert comprehensive SIC to Industry mappings
INSERT INTO crosswalk_mappings (
    id, source_code, source_system, target_code, target_system,
    confidence_score, validation_rules, is_valid, metadata, created_at, updated_at
) VALUES
-- Technology Industry Mappings (Division I - Services)
('sic_7372_tech', '7372', 'SIC', '1', 'INDUSTRY', 0.95,
 '[{"rule_type": "format", "rule_name": "sic_format_validation", "rule_value": "4-digit numeric", "description": "SIC code must be 4-digit numeric format"}, {"rule_type": "division", "rule_name": "sic_division_validation", "rule_value": "valid_division", "description": "SIC code must have valid division code"}]'::jsonb,
 true, '{"sic_description": "Prepackaged Software", "industry_name": "Technology", "industry_category": "traditional", "mapping_method": "keyword_similarity", "division": "I"}'::jsonb,
 NOW(), NOW()),

('sic_7373_tech', '7373', 'SIC', '1', 'INDUSTRY', 0.92,
 '[{"rule_type": "format", "rule_name": "sic_format_validation", "rule_value": "4-digit numeric", "description": "SIC code must be 4-digit numeric format"}, {"rule_type": "division", "rule_name": "sic_division_validation", "rule_value": "valid_division", "description": "SIC code must have valid division code"}]'::jsonb,
 true, '{"sic_description": "Computer Integrated Systems Design", "industry_name": "Technology", "industry_category": "traditional", "mapping_method": "keyword_similarity", "division": "I"}'::jsonb,
 NOW(), NOW()),

('sic_7374_tech', '7374', 'SIC', '1', 'INDUSTRY', 0.90,
 '[{"rule_type": "format", "rule_name": "sic_format_validation", "rule_value": "4-digit numeric", "description": "SIC code must be 4-digit numeric format"}, {"rule_type": "division", "rule_name": "sic_division_validation", "rule_value": "valid_division", "description": "SIC code must have valid division code"}]'::jsonb,
 true, '{"sic_description": "Computer Processing and Data Preparation and Processing Services", "industry_name": "Technology", "industry_category": "traditional", "mapping_method": "keyword_similarity", "division": "I"}'::jsonb,
 NOW(), NOW()),

-- Healthcare Industry Mappings (Division I - Services)
('sic_8062_health', '8062', 'SIC', '2', 'INDUSTRY', 0.98,
 '[{"rule_type": "format", "rule_name": "sic_format_validation", "rule_value": "4-digit numeric", "description": "SIC code must be 4-digit numeric format"}, {"rule_type": "division", "rule_name": "sic_division_validation", "rule_value": "valid_division", "description": "SIC code must have valid division code"}]'::jsonb,
 true, '{"sic_description": "General Medical and Surgical Hospitals", "industry_name": "Healthcare", "industry_category": "traditional", "mapping_method": "keyword_similarity", "division": "I"}'::jsonb,
 NOW(), NOW()),

('sic_8011_health', '8011', 'SIC', '2', 'INDUSTRY', 0.95,
 '[{"rule_type": "format", "rule_name": "sic_format_validation", "rule_value": "4-digit numeric", "description": "SIC code must be 4-digit numeric format"}, {"rule_type": "division", "rule_name": "sic_division_validation", "rule_value": "valid_division", "description": "SIC code must have valid division code"}]'::jsonb,
 true, '{"sic_description": "Offices and Clinics of Doctors of Medicine", "industry_name": "Healthcare", "industry_category": "traditional", "mapping_method": "keyword_similarity", "division": "I"}'::jsonb,
 NOW(), NOW()),

('sic_2834_health', '2834', 'SIC', '2', 'INDUSTRY', 0.93,
 '[{"rule_type": "format", "rule_name": "sic_format_validation", "rule_value": "4-digit numeric", "description": "SIC code must be 4-digit numeric format"}, {"rule_type": "division", "rule_name": "sic_division_validation", "rule_value": "valid_division", "description": "SIC code must have valid division code"}]'::jsonb,
 true, '{"sic_description": "Pharmaceutical Preparations", "industry_name": "Healthcare", "industry_category": "traditional", "mapping_method": "keyword_similarity", "division": "D"}'::jsonb,
 NOW(), NOW()),

-- Financial Services Industry Mappings (Division H - Finance, Insurance, and Real Estate)
('sic_6021_finance', '6021', 'SIC', '3', 'INDUSTRY', 0.97,
 '[{"rule_type": "format", "rule_name": "sic_format_validation", "rule_value": "4-digit numeric", "description": "SIC code must be 4-digit numeric format"}, {"rule_type": "division", "rule_name": "sic_division_validation", "rule_value": "valid_division", "description": "SIC code must have valid division code"}]'::jsonb,
 true, '{"sic_description": "National Commercial Banks", "industry_name": "Financial Services", "industry_category": "traditional", "mapping_method": "keyword_similarity", "division": "H"}'::jsonb,
 NOW(), NOW()),

('sic_6211_finance', '6211', 'SIC', '3', 'INDUSTRY', 0.96,
 '[{"rule_type": "format", "rule_name": "sic_format_validation", "rule_value": "4-digit numeric", "description": "SIC code must be 4-digit numeric format"}, {"rule_type": "division", "rule_name": "sic_division_validation", "rule_value": "valid_division", "description": "SIC code must have valid division code"}]'::jsonb,
 true, '{"sic_description": "Security Brokers, Dealers, and Flotation Companies", "industry_name": "Financial Services", "industry_category": "traditional", "mapping_method": "keyword_similarity", "division": "H"}'::jsonb,
 NOW(), NOW()),

('sic_6311_finance', '6311', 'SIC', '3', 'INDUSTRY', 0.94,
 '[{"rule_type": "format", "rule_name": "sic_format_validation", "rule_value": "4-digit numeric", "description": "SIC code must be 4-digit numeric format"}, {"rule_type": "division", "rule_name": "sic_division_validation", "rule_value": "valid_division", "description": "SIC code must have valid division code"}]'::jsonb,
 true, '{"sic_description": "Life Insurance", "industry_name": "Financial Services", "industry_category": "traditional", "mapping_method": "keyword_similarity", "division": "H"}'::jsonb,
 NOW(), NOW()),

-- Retail Industry Mappings (Division G - Retail Trade)
('sic_5311_retail', '5311', 'SIC', '4', 'INDUSTRY', 0.95,
 '[{"rule_type": "format", "rule_name": "sic_format_validation", "rule_value": "4-digit numeric", "description": "SIC code must be 4-digit numeric format"}, {"rule_type": "division", "rule_name": "sic_division_validation", "rule_value": "valid_division", "description": "SIC code must have valid division code"}]'::jsonb,
 true, '{"sic_description": "Department Stores", "industry_name": "Retail", "industry_category": "traditional", "mapping_method": "keyword_similarity", "division": "G"}'::jsonb,
 NOW(), NOW()),

('sic_5411_retail', '5411', 'SIC', '4', 'INDUSTRY', 0.93,
 '[{"rule_type": "format", "rule_name": "sic_format_validation", "rule_value": "4-digit numeric", "description": "SIC code must be 4-digit numeric format"}, {"rule_type": "division", "rule_name": "sic_division_validation", "rule_value": "valid_division", "description": "SIC code must have valid division code"}]'::jsonb,
 true, '{"sic_description": "Grocery Stores", "industry_name": "Retail", "industry_category": "traditional", "mapping_method": "keyword_similarity", "division": "G"}'::jsonb,
 NOW(), NOW()),

('sic_5734_retail', '5734', 'SIC', '4', 'INDUSTRY', 0.91,
 '[{"rule_type": "format", "rule_name": "sic_format_validation", "rule_value": "4-digit numeric", "description": "SIC code must be 4-digit numeric format"}, {"rule_type": "division", "rule_name": "sic_division_validation", "rule_value": "valid_division", "description": "SIC code must have valid division code"}]'::jsonb,
 true, '{"sic_description": "Computer and Computer Software Stores", "industry_name": "Retail", "industry_category": "traditional", "mapping_method": "keyword_similarity", "division": "G"}'::jsonb,
 NOW(), NOW()),

-- Manufacturing Industry Mappings (Division D - Manufacturing)
('sic_3711_manufacturing', '3711', 'SIC', '5', 'INDUSTRY', 0.94,
 '[{"rule_type": "format", "rule_name": "sic_format_validation", "rule_value": "4-digit numeric", "description": "SIC code must be 4-digit numeric format"}, {"rule_type": "division", "rule_name": "sic_division_validation", "rule_value": "valid_division", "description": "SIC code must have valid division code"}]'::jsonb,
 true, '{"sic_description": "Motor Vehicles and Passenger Car Bodies", "industry_name": "Manufacturing", "industry_category": "traditional", "mapping_method": "keyword_similarity", "division": "D"}'::jsonb,
 NOW(), NOW()),

('sic_2834_manufacturing', '2834', 'SIC', '5', 'INDUSTRY', 0.92,
 '[{"rule_type": "format", "rule_name": "sic_format_validation", "rule_value": "4-digit numeric", "description": "SIC code must be 4-digit numeric format"}, {"rule_type": "division", "rule_name": "sic_division_validation", "rule_value": "valid_division", "description": "SIC code must have valid division code"}]'::jsonb,
 true, '{"sic_description": "Pharmaceutical Preparations", "industry_name": "Manufacturing", "industry_category": "traditional", "mapping_method": "keyword_similarity", "division": "D"}'::jsonb,
 NOW(), NOW()),

('sic_3571_manufacturing', '3571', 'SIC', '5', 'INDUSTRY', 0.90,
 '[{"rule_type": "format", "rule_name": "sic_format_validation", "rule_value": "4-digit numeric", "description": "SIC code must be 4-digit numeric format"}, {"rule_type": "division", "rule_name": "sic_division_validation", "rule_value": "valid_division", "description": "SIC code must have valid division code"}]'::jsonb,
 true, '{"sic_description": "Electronic Computers", "industry_name": "Manufacturing", "industry_category": "traditional", "mapping_method": "keyword_similarity", "division": "D"}'::jsonb,
 NOW(), NOW())

ON CONFLICT (source_code, source_system, target_code, target_system)
DO UPDATE SET
    confidence_score = EXCLUDED.confidence_score,
    validation_rules = EXCLUDED.validation_rules,
    is_valid = EXCLUDED.is_valid,
    metadata = EXCLUDED.metadata,
    updated_at = EXCLUDED.updated_at;

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_crosswalk_mappings_sic_source ON crosswalk_mappings(source_code, source_system) WHERE source_system = 'SIC';
CREATE INDEX IF NOT EXISTS idx_crosswalk_mappings_sic_confidence ON crosswalk_mappings(confidence_score) WHERE source_system = 'SIC';

-- Verify the mappings were inserted
SELECT 
    source_code,
    source_system,
    target_code,
    target_system,
    confidence_score,
    is_valid,
    metadata->>'industry_name' as industry_name,
    metadata->>'division' as sic_division
FROM crosswalk_mappings 
WHERE source_system = 'SIC' 
ORDER BY confidence_score DESC;
