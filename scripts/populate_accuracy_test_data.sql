-- Populate accuracy test data for crosswalk testing
-- This script creates test cases and expected results for accuracy validation

-- Create accuracy_test_cases table if it doesn't exist
CREATE TABLE IF NOT EXISTS accuracy_test_cases (
    id SERIAL PRIMARY KEY,
    test_case_id VARCHAR(100) UNIQUE NOT NULL,
    test_name VARCHAR(255) NOT NULL,
    test_type VARCHAR(100) NOT NULL,
    input_data JSONB NOT NULL,
    expected_data JSONB NOT NULL,
    description TEXT,
    weight DECIMAL(3,2) DEFAULT 1.0,
    category VARCHAR(100),
    tags TEXT[],
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create accuracy_test_results table if it doesn't exist
CREATE TABLE IF NOT EXISTS accuracy_test_results (
    id SERIAL PRIMARY KEY,
    suite_name VARCHAR(255) NOT NULL,
    test_name VARCHAR(255) NOT NULL,
    test_type VARCHAR(100) NOT NULL,
    total_tests INTEGER NOT NULL,
    passed_tests INTEGER NOT NULL,
    failed_tests INTEGER NOT NULL,
    accuracy_score DECIMAL(5,4) NOT NULL,
    confidence_score DECIMAL(5,4) NOT NULL,
    summary TEXT,
    test_details JSONB,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert MCC mapping accuracy test cases
INSERT INTO accuracy_test_cases (test_case_id, test_name, test_type, input_data, expected_data, description, weight, category, tags) VALUES
('mcc_001', 'MCC 5411 - Grocery Stores', 'mcc_mapping_accuracy', 
 '{"mcc_code": "5411"}', 
 '{"industry_id": 1, "confidence_score": 0.95, "description": "Grocery Stores"}',
 'Test MCC 5411 maps to Grocery Stores industry', 1.0, 'retail', ARRAY['grocery', 'retail', 'food']),

('mcc_002', 'MCC 5999 - Miscellaneous Retail', 'mcc_mapping_accuracy',
 '{"mcc_code": "5999"}',
 '{"industry_id": 2, "confidence_score": 0.85, "description": "Miscellaneous Retail"}',
 'Test MCC 5999 maps to Miscellaneous Retail industry', 1.0, 'retail', ARRAY['miscellaneous', 'retail']),

('mcc_003', 'MCC 5812 - Eating Places', 'mcc_mapping_accuracy',
 '{"mcc_code": "5812"}',
 '{"industry_id": 3, "confidence_score": 0.90, "description": "Restaurants and Food Services"}',
 'Test MCC 5812 maps to Restaurants industry', 1.0, 'food_service', ARRAY['restaurant', 'food', 'service']),

('mcc_004', 'MCC 7011 - Hotels and Motels', 'mcc_mapping_accuracy',
 '{"mcc_code": "7011"}',
 '{"industry_id": 4, "confidence_score": 0.88, "description": "Accommodation Services"}',
 'Test MCC 7011 maps to Hotels industry', 1.0, 'hospitality', ARRAY['hotel', 'accommodation', 'travel']),

('mcc_005', 'MCC 7372 - Computer Programming Services', 'mcc_mapping_accuracy',
 '{"mcc_code": "7372"}',
 '{"industry_id": 5, "confidence_score": 0.92, "description": "Software Development"}',
 'Test MCC 7372 maps to Software Development industry', 1.0, 'technology', ARRAY['software', 'programming', 'technology']);

-- Insert NAICS mapping accuracy test cases
INSERT INTO accuracy_test_cases (test_case_id, test_name, test_type, input_data, expected_data, description, weight, category, tags) VALUES
('naics_001', 'NAICS 445110 - Supermarkets and Grocery Stores', 'naics_mapping_accuracy',
 '{"naics_code": "445110"}',
 '{"industry_id": 1, "confidence_score": 0.95, "description": "Grocery Stores"}',
 'Test NAICS 445110 maps to Grocery Stores industry', 1.0, 'retail', ARRAY['grocery', 'retail', 'food']),

('naics_002', 'NAICS 722511 - Full-Service Restaurants', 'naics_mapping_accuracy',
 '{"naics_code": "722511"}',
 '{"industry_id": 3, "confidence_score": 0.90, "description": "Restaurants and Food Services"}',
 'Test NAICS 722511 maps to Restaurants industry', 1.0, 'food_service', ARRAY['restaurant', 'food', 'service']),

('naics_003', 'NAICS 721110 - Hotels and Motels', 'naics_mapping_accuracy',
 '{"naics_code": "721110"}',
 '{"industry_id": 4, "confidence_score": 0.88, "description": "Accommodation Services"}',
 'Test NAICS 721110 maps to Hotels industry', 1.0, 'hospitality', ARRAY['hotel', 'accommodation', 'travel']),

('naics_004', 'NAICS 541511 - Custom Computer Programming Services', 'naics_mapping_accuracy',
 '{"naics_code": "541511"}',
 '{"industry_id": 5, "confidence_score": 0.92, "description": "Software Development"}',
 'Test NAICS 541511 maps to Software Development industry', 1.0, 'technology', ARRAY['software', 'programming', 'technology']),

('naics_005', 'NAICS 621111 - Offices of Physicians', 'naics_mapping_accuracy',
 '{"naics_code": "621111"}',
 '{"industry_id": 6, "confidence_score": 0.94, "description": "Healthcare Services"}',
 'Test NAICS 621111 maps to Healthcare industry', 1.0, 'healthcare', ARRAY['healthcare', 'medical', 'physician']);

-- Insert SIC mapping accuracy test cases
INSERT INTO accuracy_test_cases (test_case_id, test_name, test_type, input_data, expected_data, description, weight, category, tags) VALUES
('sic_001', 'SIC 5411 - Grocery Stores', 'sic_mapping_accuracy',
 '{"sic_code": "5411"}',
 '{"industry_id": 1, "confidence_score": 0.95, "description": "Grocery Stores"}',
 'Test SIC 5411 maps to Grocery Stores industry', 1.0, 'retail', ARRAY['grocery', 'retail', 'food']),

('sic_002', 'SIC 5812 - Eating Places', 'sic_mapping_accuracy',
 '{"sic_code": "5812"}',
 '{"industry_id": 3, "confidence_score": 0.90, "description": "Restaurants and Food Services"}',
 'Test SIC 5812 maps to Restaurants industry', 1.0, 'food_service', ARRAY['restaurant', 'food', 'service']),

('sic_003', 'SIC 7011 - Hotels and Motels', 'sic_mapping_accuracy',
 '{"sic_code": "7011"}',
 '{"industry_id": 4, "confidence_score": 0.88, "description": "Accommodation Services"}',
 'Test SIC 7011 maps to Hotels industry', 1.0, 'hospitality', ARRAY['hotel', 'accommodation', 'travel']),

('sic_004', 'SIC 7372 - Computer Programming Services', 'sic_mapping_accuracy',
 '{"sic_code": "7372"}',
 '{"industry_id": 5, "confidence_score": 0.92, "description": "Software Development"}',
 'Test SIC 7372 maps to Software Development industry', 1.0, 'technology', ARRAY['software', 'programming', 'technology']),

('sic_005', 'SIC 8011 - Offices and Clinics of Doctors of Medicine', 'sic_mapping_accuracy',
 '{"sic_code": "8011"}',
 '{"industry_id": 6, "confidence_score": 0.94, "description": "Healthcare Services"}',
 'Test SIC 8011 maps to Healthcare industry', 1.0, 'healthcare', ARRAY['healthcare', 'medical', 'physician']);

-- Insert confidence scoring accuracy test cases
INSERT INTO accuracy_test_cases (test_case_id, test_name, test_type, input_data, expected_data, description, weight, category, tags) VALUES
('confidence_001', 'High Confidence Grocery Store', 'confidence_scoring_accuracy',
 '{"mcc_code": "5411", "naics_code": "445110", "sic_code": "5411"}',
 '{"confidence_score": 0.95}',
 'Test high confidence scoring for grocery store codes', 1.0, 'confidence', ARRAY['high_confidence', 'grocery']),

('confidence_002', 'Medium Confidence Restaurant', 'confidence_scoring_accuracy',
 '{"mcc_code": "5812", "naics_code": "722511", "sic_code": "5812"}',
 '{"confidence_score": 0.90}',
 'Test medium confidence scoring for restaurant codes', 1.0, 'confidence', ARRAY['medium_confidence', 'restaurant']),

('confidence_003', 'Low Confidence Mixed Codes', 'confidence_scoring_accuracy',
 '{"mcc_code": "5999", "naics_code": "445110", "sic_code": "5411"}',
 '{"confidence_score": 0.70}',
 'Test low confidence scoring for mixed/inconsistent codes', 1.0, 'confidence', ARRAY['low_confidence', 'mixed']),

('confidence_004', 'Single Code Confidence', 'confidence_scoring_accuracy',
 '{"mcc_code": "5411"}',
 '{"confidence_score": 0.40}',
 'Test confidence scoring with single code', 1.0, 'confidence', ARRAY['single_code', 'grocery']),

('confidence_005', 'No Code Confidence', 'confidence_scoring_accuracy',
 '{}',
 '{"confidence_score": 0.0}',
 'Test confidence scoring with no codes', 1.0, 'confidence', ARRAY['no_codes', 'zero_confidence']);

-- Insert validation rules accuracy test cases
INSERT INTO accuracy_test_cases (test_case_id, test_name, test_type, input_data, expected_data, description, weight, category, tags) VALUES
('validation_001', 'Format Validation Rules', 'validation_rules_accuracy',
 '{"test_type": "format_validation"}',
 '{"expected_passed": 5}',
 'Test format validation rules pass', 1.0, 'validation', ARRAY['format', 'validation']),

('validation_002', 'Consistency Validation Rules', 'validation_rules_accuracy',
 '{"test_type": "consistency_validation"}',
 '{"expected_passed": 3}',
 'Test consistency validation rules pass', 1.0, 'validation', ARRAY['consistency', 'validation']),

('validation_003', 'Business Logic Validation Rules', 'validation_rules_accuracy',
 '{"test_type": "business_logic_validation"}',
 '{"expected_passed": 4}',
 'Test business logic validation rules pass', 1.0, 'validation', ARRAY['business_logic', 'validation']),

('validation_004', 'Cross Reference Validation Rules', 'validation_rules_accuracy',
 '{"test_type": "cross_reference_validation"}',
 '{"expected_passed": 2}',
 'Test cross reference validation rules pass', 1.0, 'validation', ARRAY['cross_reference', 'validation']);

-- Insert crosswalk consistency accuracy test cases
INSERT INTO accuracy_test_cases (test_case_id, test_name, test_type, input_data, expected_data, description, weight, category, tags) VALUES
('consistency_001', 'Grocery Store Consistency', 'crosswalk_consistency_accuracy',
 '{"industry_id": 1}',
 '{"expected_consistency": 0.9}',
 'Test consistency of grocery store mappings', 1.0, 'consistency', ARRAY['grocery', 'consistency']),

('consistency_002', 'Restaurant Consistency', 'crosswalk_consistency_accuracy',
 '{"industry_id": 3}',
 '{"expected_consistency": 0.85}',
 'Test consistency of restaurant mappings', 1.0, 'consistency', ARRAY['restaurant', 'consistency']),

('consistency_003', 'Hotel Consistency', 'crosswalk_consistency_accuracy',
 '{"industry_id": 4}',
 '{"expected_consistency": 0.88}',
 'Test consistency of hotel mappings', 1.0, 'consistency', ARRAY['hotel', 'consistency']),

('consistency_004', 'Software Development Consistency', 'crosswalk_consistency_accuracy',
 '{"industry_id": 5}',
 '{"expected_consistency": 0.92}',
 'Test consistency of software development mappings', 1.0, 'consistency', ARRAY['software', 'consistency']),

('consistency_005', 'Healthcare Consistency', 'crosswalk_consistency_accuracy',
 '{"industry_id": 6}',
 '{"expected_consistency": 0.94}',
 'Test consistency of healthcare mappings', 1.0, 'consistency', ARRAY['healthcare', 'consistency']);

-- Insert industry alignment accuracy test cases
INSERT INTO accuracy_test_cases (test_case_id, test_name, test_type, input_data, expected_data, description, weight, category, tags) VALUES
('alignment_001', 'Overall Industry Alignment', 'industry_alignment_accuracy',
 '{"test_type": "overall_alignment"}',
 '{"expected_conflicts": 5}',
 'Test overall industry alignment has minimal conflicts', 1.0, 'alignment', ARRAY['alignment', 'conflicts']),

('alignment_002', 'MCC Alignment', 'industry_alignment_accuracy',
 '{"test_type": "mcc_alignment"}',
 '{"expected_conflicts": 2}',
 'Test MCC alignment has minimal conflicts', 1.0, 'alignment', ARRAY['mcc', 'alignment']),

('alignment_003', 'NAICS Alignment', 'industry_alignment_accuracy',
 '{"test_type": "naics_alignment"}',
 '{"expected_conflicts": 1}',
 'Test NAICS alignment has minimal conflicts', 1.0, 'alignment', ARRAY['naics', 'alignment']),

('alignment_004', 'SIC Alignment', 'industry_alignment_accuracy',
 '{"test_type": "sic_alignment"}',
 '{"expected_conflicts": 2}',
 'Test SIC alignment has minimal conflicts', 1.0, 'alignment', ARRAY['sic', 'alignment']);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_accuracy_test_cases_test_type ON accuracy_test_cases(test_type);
CREATE INDEX IF NOT EXISTS idx_accuracy_test_cases_category ON accuracy_test_cases(category);
CREATE INDEX IF NOT EXISTS idx_accuracy_test_cases_tags ON accuracy_test_cases USING GIN(tags);
CREATE INDEX IF NOT EXISTS idx_accuracy_test_results_suite_name ON accuracy_test_results(suite_name);
CREATE INDEX IF NOT EXISTS idx_accuracy_test_results_test_type ON accuracy_test_results(test_type);
CREATE INDEX IF NOT EXISTS idx_accuracy_test_results_created_at ON accuracy_test_results(created_at);

-- Insert sample accuracy test results for demonstration
INSERT INTO accuracy_test_results (suite_name, test_name, test_type, total_tests, passed_tests, failed_tests, accuracy_score, confidence_score, summary, test_details, metadata) VALUES
('Crosswalk Accuracy Test Suite', 'MCC Mapping Accuracy Test', 'mcc_mapping_accuracy', 5, 5, 0, 1.0, 0.90, 'MCC mapping accuracy: 100.00% (5/5 tests passed)', '[]', '{"version": "1.0.0", "environment": "test"}'),

('Crosswalk Accuracy Test Suite', 'NAICS Mapping Accuracy Test', 'naics_mapping_accuracy', 5, 4, 1, 0.8, 0.88, 'NAICS mapping accuracy: 80.00% (4/5 tests passed)', '[]', '{"version": "1.0.0", "environment": "test"}'),

('Crosswalk Accuracy Test Suite', 'SIC Mapping Accuracy Test', 'sic_mapping_accuracy', 5, 5, 0, 1.0, 0.92, 'SIC mapping accuracy: 100.00% (5/5 tests passed)', '[]', '{"version": "1.0.0", "environment": "test"}'),

('Crosswalk Accuracy Test Suite', 'Confidence Scoring Accuracy Test', 'confidence_scoring_accuracy', 5, 4, 1, 0.8, 0.75, 'Confidence scoring accuracy: 80.00% (4/5 tests passed)', '[]', '{"version": "1.0.0", "environment": "test"}'),

('Crosswalk Accuracy Test Suite', 'Validation Rules Accuracy Test', 'validation_rules_accuracy', 4, 3, 1, 0.75, 0.82, 'Validation rules accuracy: 75.00% (3/4 tests passed)', '[]', '{"version": "1.0.0", "environment": "test"}'),

('Crosswalk Accuracy Test Suite', 'Crosswalk Consistency Accuracy Test', 'crosswalk_consistency_accuracy', 5, 4, 1, 0.8, 0.90, 'Crosswalk consistency accuracy: 80.00% (4/5 tests passed)', '[]', '{"version": "1.0.0", "environment": "test"}'),

('Crosswalk Accuracy Test Suite', 'Industry Alignment Accuracy Test', 'industry_alignment_accuracy', 4, 3, 1, 0.75, 0.85, 'Industry alignment accuracy: 75.00% (3/4 tests passed)', '[]', '{"version": "1.0.0", "environment": "test"}');

-- Display summary of inserted data
SELECT 
    'accuracy_test_cases' as table_name,
    COUNT(*) as record_count,
    COUNT(DISTINCT test_type) as test_types,
    COUNT(DISTINCT category) as categories
FROM accuracy_test_cases

UNION ALL

SELECT 
    'accuracy_test_results' as table_name,
    COUNT(*) as record_count,
    COUNT(DISTINCT test_type) as test_types,
    COUNT(DISTINCT suite_name) as suites
FROM accuracy_test_results;
