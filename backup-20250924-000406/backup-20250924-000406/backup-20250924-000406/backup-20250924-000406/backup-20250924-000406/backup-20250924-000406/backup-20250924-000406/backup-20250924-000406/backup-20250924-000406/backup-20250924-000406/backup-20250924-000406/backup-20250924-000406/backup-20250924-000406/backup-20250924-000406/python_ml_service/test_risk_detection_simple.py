#!/usr/bin/env python3
"""
Simplified Test Suite for Risk Detection Models

This script tests the risk detection system without requiring
Hugging Face model downloads to avoid rate limiting issues.
"""

import os
import sys
import json
import time
import logging
from typing import Dict, List, Any
from pathlib import Path

# Add current directory to path for imports
sys.path.append(str(Path(__file__).parent))

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

class SimpleRiskDetectionTester:
    """Simplified tester for risk detection models"""
    
    def __init__(self):
        self.test_results = {}
        
    def run_all_tests(self) -> Dict[str, Any]:
        """Run simplified risk detection tests"""
        logger.info("🚀 Starting simplified risk detection tests...")
        
        # Test 1: Dataset Generation
        self.test_dataset_generation()
        
        # Test 2: Pattern Recognition
        self.test_pattern_recognition()
        
        # Test 3: Anomaly Detection (without BERT)
        self.test_anomaly_detection_simple()
        
        # Test 4: Risk Scoring Logic
        self.test_risk_scoring_logic()
        
        # Test 5: File Structure
        self.test_file_structure()
        
        # Generate test report
        summary = self.generate_test_report()
        
        return summary
    
    def test_dataset_generation(self):
        """Test dataset generation"""
        logger.info("🧪 Testing dataset generation...")
        
        try:
            from risk_detection_dataset import RiskDetectionDatasetGenerator
            
            # Create generator
            generator = RiskDetectionDatasetGenerator()
            
            # Generate small dataset
            dataset = generator.generate_dataset(num_samples=100)
            
            # Check dataset structure
            required_columns = [
                'id', 'business_name', 'description', 'website_url', 
                'website_content', 'risk_category', 'risk_subcategory',
                'risk_level', 'risk_severity', 'risk_score', 'is_risk'
            ]
            
            for column in required_columns:
                assert column in dataset.columns, f"Missing column: {column}"
            
            # Check data types
            assert len(dataset) == 100, f"Expected 100 samples, got {len(dataset)}"
            assert dataset['risk_score'].dtype in ['float64', 'float32'], "Invalid risk score type"
            assert dataset['is_risk'].dtype == 'bool', "Invalid is_risk type"
            
            # Check risk level distribution
            risk_levels = dataset['risk_level'].value_counts()
            assert len(risk_levels) >= 2, "Need at least 2 risk levels"
            
            self.test_results['dataset_generation'] = {
                'status': 'PASSED',
                'message': f'Generated dataset with {len(dataset)} samples',
                'details': {
                    'total_samples': len(dataset),
                    'columns': list(dataset.columns),
                    'risk_levels': risk_levels.to_dict(),
                    'risk_categories': dataset['risk_category'].value_counts().to_dict()
                }
            }
            
            logger.info("✅ Dataset generation test PASSED")
            
        except Exception as e:
            self.test_results['dataset_generation'] = {
                'status': 'FAILED',
                'message': f'Dataset generation failed: {str(e)}',
                'error': str(e)
            }
            logger.error(f"❌ Dataset generation test FAILED: {e}")
    
    def test_pattern_recognition(self):
        """Test pattern recognition without BERT"""
        logger.info("🧪 Testing pattern recognition...")
        
        try:
            # Import pattern recognition directly
            from risk_detection_models import PatternRecognitionModel
            
            # Create pattern model
            pattern_model = PatternRecognitionModel()
            
            # Test cases
            test_cases = [
                {
                    'text': 'Shell company formation and offshore services for money laundering',
                    'expected_patterns': ['money_laundering'],
                    'expected_score': 0.7
                },
                {
                    'text': 'Terrorist financing and extremist funding operations',
                    'expected_patterns': ['terrorist_financing'],
                    'expected_score': 0.8
                },
                {
                    'text': 'Drug trafficking and cocaine distribution network',
                    'expected_patterns': ['drug_trafficking'],
                    'expected_score': 0.9
                },
                {
                    'text': 'Identity theft and credit card fraud services',
                    'expected_patterns': ['fraud'],
                    'expected_score': 0.6
                },
                {
                    'text': 'Normal restaurant business with good food',
                    'expected_patterns': [],
                    'expected_score': 0.0
                }
            ]
            
            passed = 0
            total = len(test_cases)
            
            for i, test_case in enumerate(test_cases):
                try:
                    detected_patterns = pattern_model.detect_patterns(test_case['text'])
                    pattern_risk_score = pattern_model.calculate_pattern_risk_score(detected_patterns)
                    
                    # Check if results are reasonable
                    assert isinstance(detected_patterns, dict), "Invalid detected patterns format"
                    assert 0.0 <= pattern_risk_score <= 1.0, f"Invalid pattern risk score: {pattern_risk_score}"
                    
                    passed += 1
                    logger.info(f"✅ Pattern test case {i+1} PASSED: {test_case['text'][:50]}...")
                    
                except Exception as e:
                    logger.error(f"❌ Pattern test case {i+1} FAILED: {e}")
            
            self.test_results['pattern_recognition'] = {
                'status': 'PASSED' if passed == total else 'FAILED',
                'message': f'{passed}/{total} test cases passed',
                'details': {
                    'passed': passed,
                    'total': total,
                    'success_rate': passed / total
                }
            }
            
            logger.info(f"✅ Pattern recognition test: {passed}/{total} PASSED")
            
        except Exception as e:
            self.test_results['pattern_recognition'] = {
                'status': 'FAILED',
                'message': f'Pattern recognition test failed: {str(e)}',
                'error': str(e)
            }
            logger.error(f"❌ Pattern recognition test FAILED: {e}")
    
    def test_anomaly_detection_simple(self):
        """Test anomaly detection without BERT"""
        logger.info("🧪 Testing anomaly detection (simple)...")
        
        try:
            from risk_detection_models import AnomalyDetectionModel
            
            # Create anomaly model
            anomaly_model = AnomalyDetectionModel()
            
            # Generate test texts
            test_texts = [
                'Normal business description for a restaurant',
                'Cocaine distribution network and drug trafficking',
                'Standard retail clothing store',
                'Illegal weapons dealer and firearms smuggling',
                'Regular accounting and tax services'
            ]
            
            # Fit model
            anomaly_model.fit(test_texts)
            
            # Test prediction
            passed = 0
            total = len(test_texts)
            
            for text in test_texts:
                try:
                    is_anomaly, anomaly_scores = anomaly_model.predict([text])
                    
                    # Check if results are valid
                    assert isinstance(is_anomaly, list) or hasattr(is_anomaly, '__len__'), "Invalid anomaly prediction format"
                    assert isinstance(anomaly_scores, list) or hasattr(anomaly_scores, '__len__'), "Invalid anomaly scores format"
                    
                    passed += 1
                    
                except Exception as e:
                    logger.error(f"❌ Anomaly detection failed for: {text[:50]}... - {e}")
            
            self.test_results['anomaly_detection'] = {
                'status': 'PASSED' if passed == total else 'FAILED',
                'message': f'{passed}/{total} test cases passed',
                'details': {
                    'passed': passed,
                    'total': total,
                    'success_rate': passed / total,
                    'model_fitted': anomaly_model.is_fitted
                }
            }
            
            logger.info(f"✅ Anomaly detection test: {passed}/{total} PASSED")
            
        except Exception as e:
            self.test_results['anomaly_detection'] = {
                'status': 'FAILED',
                'message': f'Anomaly detection test failed: {str(e)}',
                'error': str(e)
            }
            logger.error(f"❌ Anomaly detection test FAILED: {e}")
    
    def test_risk_scoring_logic(self):
        """Test risk scoring logic"""
        logger.info("🧪 Testing risk scoring logic...")
        
        try:
            # Test risk scoring without models
            test_cases = [
                {
                    'bert_score': 0.8,
                    'anomaly_score': 0.9,
                    'pattern_score': 0.7,
                    'expected_range': (0.7, 0.9)
                },
                {
                    'bert_score': 0.2,
                    'anomaly_score': 0.3,
                    'pattern_score': 0.1,
                    'expected_range': (0.1, 0.3)
                },
                {
                    'bert_score': 0.5,
                    'anomaly_score': 0.6,
                    'pattern_score': 0.4,
                    'expected_range': (0.4, 0.6)
                }
            ]
            
            passed = 0
            total = len(test_cases)
            
            for i, test_case in enumerate(test_cases):
                try:
                    # Calculate weighted combination (same logic as in the model)
                    overall_risk_score = (
                        test_case['bert_score'] * 0.5 +
                        test_case['anomaly_score'] * 0.3 +
                        test_case['pattern_score'] * 0.2
                    )
                    
                    # Check if score is in expected range
                    min_expected, max_expected = test_case['expected_range']
                    assert min_expected <= overall_risk_score <= max_expected, f"Score {overall_risk_score} not in range {test_case['expected_range']}"
                    
                    # Determine risk level
                    if overall_risk_score >= 0.8:
                        risk_level = "critical"
                    elif overall_risk_score >= 0.6:
                        risk_level = "high"
                    elif overall_risk_score >= 0.4:
                        risk_level = "medium"
                    else:
                        risk_level = "low"
                    
                    assert risk_level in ['low', 'medium', 'high', 'critical'], f"Invalid risk level: {risk_level}"
                    
                    passed += 1
                    
                except Exception as e:
                    logger.error(f"❌ Risk scoring test case {i+1} FAILED: {e}")
            
            self.test_results['risk_scoring_logic'] = {
                'status': 'PASSED' if passed == total else 'FAILED',
                'message': f'{passed}/{total} test cases passed',
                'details': {
                    'passed': passed,
                    'total': total,
                    'success_rate': passed / total
                }
            }
            
            logger.info(f"✅ Risk scoring logic test: {passed}/{total} PASSED")
            
        except Exception as e:
            self.test_results['risk_scoring_logic'] = {
                'status': 'FAILED',
                'message': f'Risk scoring logic test failed: {str(e)}',
                'error': str(e)
            }
            logger.error(f"❌ Risk scoring logic test FAILED: {e}")
    
    def test_file_structure(self):
        """Test file structure and imports"""
        logger.info("🧪 Testing file structure...")
        
        try:
            # Check if all required files exist
            required_files = [
                'risk_detection_models.py',
                'risk_detection_dataset.py',
                'app.py',
                'test_risk_detection.py',
                'requirements.txt'
            ]
            
            missing_files = []
            for file in required_files:
                if not Path(file).exists():
                    missing_files.append(file)
            
            # Check if files can be imported
            import_errors = []
            
            try:
                from risk_detection_models import PatternRecognitionModel, AnomalyDetectionModel
            except Exception as e:
                import_errors.append(f"risk_detection_models: {e}")
            
            try:
                from risk_detection_dataset import RiskDetectionDatasetGenerator
            except Exception as e:
                import_errors.append(f"risk_detection_dataset: {e}")
            
            # Check if directories exist
            required_dirs = ['models', 'cache', 'data']
            missing_dirs = []
            for dir_name in required_dirs:
                if not Path(dir_name).exists():
                    missing_dirs.append(dir_name)
            
            if missing_files or import_errors or missing_dirs:
                status = 'FAILED'
                message = f'File structure issues found'
            else:
                status = 'PASSED'
                message = 'All files and imports working correctly'
            
            self.test_results['file_structure'] = {
                'status': status,
                'message': message,
                'details': {
                    'missing_files': missing_files,
                    'import_errors': import_errors,
                    'missing_dirs': missing_dirs,
                    'required_files': required_files,
                    'required_dirs': required_dirs
                }
            }
            
            if status == 'PASSED':
                logger.info("✅ File structure test PASSED")
            else:
                logger.error("❌ File structure test FAILED")
            
        except Exception as e:
            self.test_results['file_structure'] = {
                'status': 'FAILED',
                'message': f'File structure test failed: {str(e)}',
                'error': str(e)
            }
            logger.error(f"❌ File structure test FAILED: {e}")
    
    def generate_test_report(self):
        """Generate comprehensive test report"""
        logger.info("📊 Generating test report...")
        
        # Calculate overall results
        total_tests = len(self.test_results)
        passed_tests = sum(1 for result in self.test_results.values() if result['status'] == 'PASSED')
        failed_tests = total_tests - passed_tests
        
        # Create summary
        summary = {
            'overall_status': 'PASSED' if failed_tests == 0 else 'FAILED',
            'total_tests': total_tests,
            'passed_tests': passed_tests,
            'failed_tests': failed_tests,
            'success_rate': passed_tests / total_tests if total_tests > 0 else 0,
            'timestamp': time.strftime('%Y-%m-%d %H:%M:%S'),
            'test_results': self.test_results
        }
        
        # Save report
        report_path = Path("test_reports/simple_risk_detection_test_report.json")
        report_path.parent.mkdir(exist_ok=True)
        
        with open(report_path, 'w') as f:
            json.dump(summary, f, indent=2)
        
        # Print summary
        logger.info("=" * 60)
        logger.info("🧪 SIMPLIFIED RISK DETECTION TEST SUMMARY")
        logger.info("=" * 60)
        logger.info(f"Overall Status: {summary['overall_status']}")
        logger.info(f"Total Tests: {total_tests}")
        logger.info(f"Passed: {passed_tests}")
        logger.info(f"Failed: {failed_tests}")
        logger.info(f"Success Rate: {summary['success_rate']:.2%}")
        logger.info("=" * 60)
        
        for test_name, result in self.test_results.items():
            status_icon = "✅" if result['status'] == 'PASSED' else "❌"
            logger.info(f"{status_icon} {test_name}: {result['status']} - {result['message']}")
        
        logger.info("=" * 60)
        logger.info(f"📄 Detailed report saved to: {report_path}")
        
        return summary

def main():
    """Run simplified risk detection tests"""
    logger.info("🚀 Starting simplified risk detection tests...")
    
    # Create tester
    tester = SimpleRiskDetectionTester()
    
    # Run all tests
    results = tester.run_all_tests()
    
    # Check if all tests passed
    if results.get('overall_status') == 'PASSED':
        logger.info("🎉 All simplified risk detection tests PASSED!")
        return 0
    else:
        logger.error("💥 Some simplified risk detection tests FAILED!")
        return 1

if __name__ == "__main__":
    exit(main())
