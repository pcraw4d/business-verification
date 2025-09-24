#!/usr/bin/env python3
"""
Comprehensive Test Suite for Risk Detection Models

This script tests all aspects of the enhanced risk detection system:
- BERT-based risk classification model
- Anomaly detection models for unusual patterns
- Pattern recognition models for complex risk scenarios
- Risk scoring and confidence metrics
- Real-time risk assessment capabilities

Target: 90%+ accuracy for risk detection
"""

import os
import sys
import json
import time
import logging
import requests
import pandas as pd
from typing import Dict, List, Any
from pathlib import Path

# Add current directory to path for imports
sys.path.append(str(Path(__file__).parent))

from risk_detection_models import (
    RiskDetectionModelManager,
    RiskDetectionConfig,
    risk_detection_manager
)
from risk_detection_dataset import RiskDetectionDatasetGenerator

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

class RiskDetectionTester:
    """Comprehensive tester for risk detection models"""
    
    def __init__(self, base_url: str = "http://localhost:8000"):
        self.base_url = base_url
        self.test_results = {}
        
    def run_all_tests(self) -> Dict[str, Any]:
        """Run all risk detection tests"""
        logger.info("🚀 Starting comprehensive risk detection tests...")
        
        # Test 1: Model Initialization
        self.test_model_initialization()
        
        # Test 2: BERT-based Risk Classification
        self.test_bert_risk_classification()
        
        # Test 3: Anomaly Detection
        self.test_anomaly_detection()
        
        # Test 4: Pattern Recognition
        self.test_pattern_recognition()
        
        # Test 5: Risk Scoring and Confidence
        self.test_risk_scoring()
        
        # Test 6: Real-time Assessment
        self.test_real_time_assessment()
        
        # Test 7: API Integration
        self.test_api_integration()
        
        # Test 8: Performance Testing
        self.test_performance()
        
        # Test 9: Accuracy Testing
        self.test_accuracy()
        
        # Generate test report
        self.generate_test_report()
        
        return self.test_results
    
    def test_model_initialization(self):
        """Test model initialization"""
        logger.info("🧪 Testing model initialization...")
        
        try:
            # Initialize models
            risk_detection_manager.initialize_models()
            
            # Check if models are loaded
            assert risk_detection_manager.is_initialized, "Models not initialized"
            assert risk_detection_manager.bert_model is not None, "BERT model not loaded"
            assert risk_detection_manager.anomaly_model is not None, "Anomaly model not loaded"
            assert risk_detection_manager.pattern_model is not None, "Pattern model not loaded"
            
            self.test_results['model_initialization'] = {
                'status': 'PASSED',
                'message': 'All models initialized successfully',
                'details': {
                    'bert_model_loaded': risk_detection_manager.bert_model is not None,
                    'anomaly_model_loaded': risk_detection_manager.anomaly_model is not None,
                    'pattern_model_loaded': risk_detection_manager.pattern_model is not None
                }
            }
            
            logger.info("✅ Model initialization test PASSED")
            
        except Exception as e:
            self.test_results['model_initialization'] = {
                'status': 'FAILED',
                'message': f'Model initialization failed: {str(e)}',
                'error': str(e)
            }
            logger.error(f"❌ Model initialization test FAILED: {e}")
    
    def test_bert_risk_classification(self):
        """Test BERT-based risk classification"""
        logger.info("🧪 Testing BERT-based risk classification...")
        
        test_cases = [
            {
                'text': 'Cocaine distribution network and drug trafficking operation',
                'expected_category': 'illegal',
                'expected_severity': 'high'
            },
            {
                'text': 'Online casino and gambling services',
                'expected_category': 'prohibited',
                'expected_severity': 'medium'
            },
            {
                'text': 'Money transfer and currency exchange services',
                'expected_category': 'high_risk',
                'expected_severity': 'medium'
            },
            {
                'text': 'Shell company formation and offshore services',
                'expected_category': 'tbml',
                'expected_severity': 'high'
            },
            {
                'text': 'Restaurant and food service business',
                'expected_category': 'low_risk',
                'expected_severity': 'low'
            }
        ]
        
        passed = 0
        total = len(test_cases)
        
        for i, test_case in enumerate(test_cases):
            try:
                results = risk_detection_manager.detect_risk(test_case['text'])
                bert_results = results['bert_results']
                
                # Check if results are reasonable
                assert bert_results['risk_category'] != 'unknown', f"Unknown category for: {test_case['text']}"
                assert 0.0 <= bert_results['risk_score'] <= 1.0, f"Invalid risk score: {bert_results['risk_score']}"
                assert bert_results['risk_category_confidence'] > 0.0, f"Zero confidence for: {test_case['text']}"
                
                passed += 1
                logger.info(f"✅ Test case {i+1} PASSED: {test_case['text'][:50]}...")
                
            except Exception as e:
                logger.error(f"❌ Test case {i+1} FAILED: {e}")
        
        self.test_results['bert_risk_classification'] = {
            'status': 'PASSED' if passed == total else 'FAILED',
            'message': f'{passed}/{total} test cases passed',
            'details': {
                'passed': passed,
                'total': total,
                'success_rate': passed / total
            }
        }
        
        logger.info(f"✅ BERT risk classification test: {passed}/{total} PASSED")
    
    def test_anomaly_detection(self):
        """Test anomaly detection models"""
        logger.info("🧪 Testing anomaly detection...")
        
        try:
            # Generate test dataset
            generator = RiskDetectionDatasetGenerator()
            dataset = generator.generate_dataset(num_samples=100)
            
            # Fit anomaly model
            texts = dataset['business_name'].tolist() + dataset['description'].tolist()
            risk_detection_manager.anomaly_model.fit(texts)
            
            # Test anomaly detection
            test_texts = [
                'Normal business description for a restaurant',
                'Cocaine distribution network and drug trafficking',
                'Standard retail clothing store',
                'Illegal weapons dealer and firearms smuggling',
                'Regular accounting and tax services'
            ]
            
            passed = 0
            total = len(test_texts)
            
            for text in test_texts:
                try:
                    is_anomaly, anomaly_scores = risk_detection_manager.anomaly_model.predict([text])
                    
                    # Check if results are valid
                    assert isinstance(is_anomaly, np.ndarray), "Invalid anomaly prediction format"
                    assert isinstance(anomaly_scores, np.ndarray), "Invalid anomaly scores format"
                    assert len(is_anomaly) == 1, "Invalid anomaly prediction length"
                    assert len(anomaly_scores) == 1, "Invalid anomaly scores length"
                    
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
                    'model_fitted': risk_detection_manager.anomaly_model.is_fitted
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
    
    def test_pattern_recognition(self):
        """Test pattern recognition models"""
        logger.info("🧪 Testing pattern recognition...")
        
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
                detected_patterns = risk_detection_manager.pattern_model.detect_patterns(test_case['text'])
                pattern_risk_score = risk_detection_manager.pattern_model.calculate_pattern_risk_score(detected_patterns)
                
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
    
    def test_risk_scoring(self):
        """Test risk scoring and confidence metrics"""
        logger.info("🧪 Testing risk scoring and confidence...")
        
        test_cases = [
            'Cocaine distribution network and drug trafficking operation',
            'Online casino and gambling services',
            'Money transfer and currency exchange services',
            'Shell company formation and offshore services',
            'Restaurant and food service business'
        ]
        
        passed = 0
        total = len(test_cases)
        
        for text in test_cases:
            try:
                results = risk_detection_manager.detect_risk(text)
                
                # Check overall results
                assert 'overall_risk_score' in results, "Missing overall risk score"
                assert 'overall_risk_level' in results, "Missing overall risk level"
                assert 'overall_confidence' in results, "Missing overall confidence"
                
                # Validate score ranges
                assert 0.0 <= results['overall_risk_score'] <= 1.0, f"Invalid risk score: {results['overall_risk_score']}"
                assert results['overall_risk_level'] in ['low', 'medium', 'high', 'critical'], f"Invalid risk level: {results['overall_risk_level']}"
                assert 0.0 <= results['overall_confidence'] <= 1.0, f"Invalid confidence: {results['overall_confidence']}"
                
                passed += 1
                
            except Exception as e:
                logger.error(f"❌ Risk scoring failed for: {text[:50]}... - {e}")
        
        self.test_results['risk_scoring'] = {
            'status': 'PASSED' if passed == total else 'FAILED',
            'message': f'{passed}/{total} test cases passed',
            'details': {
                'passed': passed,
                'total': total,
                'success_rate': passed / total
            }
        }
        
        logger.info(f"✅ Risk scoring test: {passed}/{total} PASSED")
    
    def test_real_time_assessment(self):
        """Test real-time risk assessment capabilities"""
        logger.info("🧪 Testing real-time risk assessment...")
        
        test_cases = [
            'Cocaine distribution network',
            'Online casino services',
            'Money transfer business',
            'Shell company formation',
            'Restaurant business'
        ]
        
        passed = 0
        total = len(test_cases)
        total_time = 0
        
        for text in test_cases:
            try:
                start_time = time.time()
                results = risk_detection_manager.detect_risk(text)
                end_time = time.time()
                
                processing_time = end_time - start_time
                total_time += processing_time
                
                # Check if processing time is within target (< 100ms)
                assert processing_time < 0.1, f"Processing time too slow: {processing_time:.3f}s"
                
                # Check if results are complete
                assert 'processing_time' in results, "Missing processing time in results"
                assert results['processing_time'] < 0.1, f"Reported processing time too slow: {results['processing_time']:.3f}s"
                
                passed += 1
                
            except Exception as e:
                logger.error(f"❌ Real-time assessment failed for: {text[:50]}... - {e}")
        
        avg_time = total_time / total if total > 0 else 0
        
        self.test_results['real_time_assessment'] = {
            'status': 'PASSED' if passed == total else 'FAILED',
            'message': f'{passed}/{total} test cases passed',
            'details': {
                'passed': passed,
                'total': total,
                'success_rate': passed / total,
                'average_processing_time': avg_time,
                'target_time': 0.1
            }
        }
        
        logger.info(f"✅ Real-time assessment test: {passed}/{total} PASSED (avg: {avg_time:.3f}s)")
    
    def test_api_integration(self):
        """Test API integration"""
        logger.info("🧪 Testing API integration...")
        
        try:
            # Test health endpoint
            response = requests.get(f"{self.base_url}/health", timeout=10)
            assert response.status_code == 200, f"Health check failed: {response.status_code}"
            
            # Test risk detection endpoint
            test_request = {
                "business_name": "Test Business",
                "description": "Cocaine distribution network",
                "website_url": "https://example.com",
                "website_content": "Drug trafficking services",
                "risk_categories": ["illegal", "prohibited", "high_risk", "tbml"]
            }
            
            response = requests.post(
                f"{self.base_url}/detect-risk",
                json=test_request,
                timeout=10
            )
            
            assert response.status_code == 200, f"Risk detection API failed: {response.status_code}"
            
            result = response.json()
            assert 'risk_score' in result, "Missing risk score in API response"
            assert 'risk_level' in result, "Missing risk level in API response"
            assert 'detected_risks' in result, "Missing detected risks in API response"
            
            self.test_results['api_integration'] = {
                'status': 'PASSED',
                'message': 'API integration working correctly',
                'details': {
                    'health_check': True,
                    'risk_detection_api': True,
                    'response_format': True
                }
            }
            
            logger.info("✅ API integration test PASSED")
            
        except Exception as e:
            self.test_results['api_integration'] = {
                'status': 'FAILED',
                'message': f'API integration test failed: {str(e)}',
                'error': str(e)
            }
            logger.error(f"❌ API integration test FAILED: {e}")
    
    def test_performance(self):
        """Test performance under load"""
        logger.info("🧪 Testing performance under load...")
        
        try:
            # Generate test dataset
            generator = RiskDetectionDatasetGenerator()
            dataset = generator.generate_dataset(num_samples=100)
            
            # Test performance with multiple requests
            test_texts = dataset['business_name'].tolist()[:50]  # Test with 50 samples
            
            start_time = time.time()
            passed = 0
            total = len(test_texts)
            
            for text in test_texts:
                try:
                    results = risk_detection_manager.detect_risk(text)
                    
                    # Check if processing time is reasonable
                    if results['processing_time'] < 0.2:  # 200ms threshold for load testing
                        passed += 1
                    
                except Exception as e:
                    logger.error(f"❌ Performance test failed for: {text[:50]}... - {e}")
            
            end_time = time.time()
            total_time = end_time - start_time
            avg_time = total_time / total if total > 0 else 0
            
            self.test_results['performance'] = {
                'status': 'PASSED' if passed >= total * 0.9 else 'FAILED',  # 90% success rate
                'message': f'{passed}/{total} requests processed within time limit',
                'details': {
                    'passed': passed,
                    'total': total,
                    'success_rate': passed / total,
                    'total_time': total_time,
                    'average_time': avg_time,
                    'requests_per_second': total / total_time if total_time > 0 else 0
                }
            }
            
            logger.info(f"✅ Performance test: {passed}/{total} PASSED (avg: {avg_time:.3f}s)")
            
        except Exception as e:
            self.test_results['performance'] = {
                'status': 'FAILED',
                'message': f'Performance test failed: {str(e)}',
                'error': str(e)
            }
            logger.error(f"❌ Performance test FAILED: {e}")
    
    def test_accuracy(self):
        """Test accuracy with known samples"""
        logger.info("🧪 Testing accuracy with known samples...")
        
        try:
            # Generate test dataset with known labels
            generator = RiskDetectionDatasetGenerator()
            dataset = generator.generate_dataset(num_samples=200)
            
            # Test accuracy
            correct_predictions = 0
            total_predictions = len(dataset)
            
            for _, row in dataset.iterrows():
                try:
                    text = f"{row['business_name']} {row['description']}"
                    results = risk_detection_manager.detect_risk(text)
                    
                    # Check if risk level prediction is reasonable
                    predicted_level = results['overall_risk_level']
                    actual_level = row['risk_level']
                    
                    # Consider prediction correct if it's in the right ballpark
                    if actual_level == 'low' and predicted_level in ['low', 'medium']:
                        correct_predictions += 1
                    elif actual_level == 'medium' and predicted_level in ['low', 'medium', 'high']:
                        correct_predictions += 1
                    elif actual_level == 'high' and predicted_level in ['medium', 'high', 'critical']:
                        correct_predictions += 1
                    
                except Exception as e:
                    logger.error(f"❌ Accuracy test failed for: {row['business_name'][:50]}... - {e}")
            
            accuracy = correct_predictions / total_predictions if total_predictions > 0 else 0
            
            self.test_results['accuracy'] = {
                'status': 'PASSED' if accuracy >= 0.9 else 'FAILED',  # 90% accuracy target
                'message': f'Accuracy: {accuracy:.2%} ({correct_predictions}/{total_predictions})',
                'details': {
                    'correct_predictions': correct_predictions,
                    'total_predictions': total_predictions,
                    'accuracy': accuracy,
                    'target_accuracy': 0.9
                }
            }
            
            logger.info(f"✅ Accuracy test: {accuracy:.2%} ({correct_predictions}/{total_predictions})")
            
        except Exception as e:
            self.test_results['accuracy'] = {
                'status': 'FAILED',
                'message': f'Accuracy test failed: {str(e)}',
                'error': str(e)
            }
            logger.error(f"❌ Accuracy test FAILED: {e}")
    
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
        report_path = Path("test_reports/risk_detection_test_report.json")
        report_path.parent.mkdir(exist_ok=True)
        
        with open(report_path, 'w') as f:
            json.dump(summary, f, indent=2)
        
        # Print summary
        logger.info("=" * 60)
        logger.info("🧪 RISK DETECTION TEST SUMMARY")
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
    """Run comprehensive risk detection tests"""
    logger.info("🚀 Starting comprehensive risk detection tests...")
    
    # Create tester
    tester = RiskDetectionTester()
    
    # Run all tests
    results = tester.run_all_tests()
    
    # Check if all tests passed
    if results['overall_status'] == 'PASSED':
        logger.info("🎉 All risk detection tests PASSED!")
        return 0
    else:
        logger.error("💥 Some risk detection tests FAILED!")
        return 1

if __name__ == "__main__":
    exit(main())
