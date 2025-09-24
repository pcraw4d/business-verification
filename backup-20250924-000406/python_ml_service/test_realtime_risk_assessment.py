#!/usr/bin/env python3
"""
Test Suite for Real-time Risk Assessment Capabilities

This script tests the real-time risk assessment features including:
- Starting assessments
- Updating assessments with new data
- WebSocket connections
- Real-time monitoring
"""

import asyncio
import json
import time
import logging
import websockets
import requests
from typing import Dict, Any, List
from datetime import datetime

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

class RealTimeRiskAssessmentTester:
    """Tester for real-time risk assessment capabilities"""
    
    def __init__(self, base_url: str = "http://localhost:8000"):
        self.base_url = base_url
        self.ws_url = base_url.replace("http", "ws")
        self.test_results = {}
        
    async def run_all_tests(self) -> Dict[str, Any]:
        """Run all real-time risk assessment tests"""
        logger.info("🚀 Starting real-time risk assessment tests...")
        
        # Test 1: Start Real-time Assessment
        await self.test_start_assessment()
        
        # Test 2: Update Assessment
        await self.test_update_assessment()
        
        # Test 3: Get Assessment Status
        await self.test_get_assessment_status()
        
        # Test 4: WebSocket Connection
        await self.test_websocket_connection()
        
        # Test 5: Complete Assessment
        await self.test_complete_assessment()
        
        # Test 6: Active Assessments List
        await self.test_get_active_assessments()
        
        # Test 7: Real-time Monitoring
        await self.test_real_time_monitoring()
        
        # Generate test report
        self.generate_test_report()
        
        return self.test_results
    
    async def test_start_assessment(self):
        """Test starting a real-time risk assessment"""
        logger.info("🧪 Testing start real-time assessment...")
        
        try:
            # Test data
            test_data = {
                "business_name": "Test Company for Real-time Assessment",
                "description": "A test company for real-time risk assessment",
                "website_url": "https://testcompany.com",
                "website_content": "Test website content"
            }
            
            # Start assessment
            response = requests.post(
                f"{self.base_url}/real-time/start-assessment",
                json=test_data,
                timeout=10
            )
            
            if response.status_code == 200:
                result = response.json()
                assessment_id = result.get('assessment_id')
                
                if assessment_id:
                    self.test_results['start_assessment'] = {
                        'status': 'PASSED',
                        'message': f'Assessment started with ID: {assessment_id}',
                        'assessment_id': assessment_id,
                        'details': result
                    }
                    logger.info(f"✅ Start assessment test PASSED: {assessment_id}")
                else:
                    self.test_results['start_assessment'] = {
                        'status': 'FAILED',
                        'message': 'No assessment ID returned',
                        'error': 'Missing assessment_id in response'
                    }
                    logger.error("❌ Start assessment test FAILED: No assessment ID")
            else:
                self.test_results['start_assessment'] = {
                    'status': 'FAILED',
                    'message': f'HTTP {response.status_code}',
                    'error': response.text
                }
                logger.error(f"❌ Start assessment test FAILED: {response.status_code}")
                
        except Exception as e:
            self.test_results['start_assessment'] = {
                'status': 'FAILED',
                'message': f'Start assessment failed: {str(e)}',
                'error': str(e)
            }
            logger.error(f"❌ Start assessment test FAILED: {e}")
    
    async def test_update_assessment(self):
        """Test updating a real-time risk assessment"""
        logger.info("🧪 Testing update real-time assessment...")
        
        try:
            # Get assessment ID from previous test
            assessment_id = self.test_results.get('start_assessment', {}).get('assessment_id')
            
            if not assessment_id:
                self.test_results['update_assessment'] = {
                    'status': 'SKIPPED',
                    'message': 'No assessment ID available from start test'
                }
                logger.warning("⚠️ Update assessment test SKIPPED: No assessment ID")
                return
            
            # Update data
            update_data = {
                "description": "Updated description with additional risk factors",
                "website_content": "Updated website content with suspicious patterns"
            }
            
            # Update assessment
            response = requests.post(
                f"{self.base_url}/real-time/update-assessment/{assessment_id}",
                json=update_data,
                timeout=10
            )
            
            if response.status_code == 200:
                result = response.json()
                self.test_results['update_assessment'] = {
                    'status': 'PASSED',
                    'message': 'Assessment updated successfully',
                    'details': result
                }
                logger.info("✅ Update assessment test PASSED")
            else:
                self.test_results['update_assessment'] = {
                    'status': 'FAILED',
                    'message': f'HTTP {response.status_code}',
                    'error': response.text
                }
                logger.error(f"❌ Update assessment test FAILED: {response.status_code}")
                
        except Exception as e:
            self.test_results['update_assessment'] = {
                'status': 'FAILED',
                'message': f'Update assessment failed: {str(e)}',
                'error': str(e)
            }
            logger.error(f"❌ Update assessment test FAILED: {e}")
    
    async def test_get_assessment_status(self):
        """Test getting assessment status"""
        logger.info("🧪 Testing get assessment status...")
        
        try:
            # Get assessment ID from previous test
            assessment_id = self.test_results.get('start_assessment', {}).get('assessment_id')
            
            if not assessment_id:
                self.test_results['get_assessment_status'] = {
                    'status': 'SKIPPED',
                    'message': 'No assessment ID available'
                }
                logger.warning("⚠️ Get assessment status test SKIPPED: No assessment ID")
                return
            
            # Get status
            response = requests.get(
                f"{self.base_url}/real-time/assessment-status/{assessment_id}",
                timeout=10
            )
            
            if response.status_code == 200:
                result = response.json()
                self.test_results['get_assessment_status'] = {
                    'status': 'PASSED',
                    'message': 'Assessment status retrieved successfully',
                    'details': result
                }
                logger.info("✅ Get assessment status test PASSED")
            else:
                self.test_results['get_assessment_status'] = {
                    'status': 'FAILED',
                    'message': f'HTTP {response.status_code}',
                    'error': response.text
                }
                logger.error(f"❌ Get assessment status test FAILED: {response.status_code}")
                
        except Exception as e:
            self.test_results['get_assessment_status'] = {
                'status': 'FAILED',
                'message': f'Get assessment status failed: {str(e)}',
                'error': str(e)
            }
            logger.error(f"❌ Get assessment status test FAILED: {e}")
    
    async def test_websocket_connection(self):
        """Test WebSocket connection for real-time updates"""
        logger.info("🧪 Testing WebSocket connection...")
        
        try:
            # Get assessment ID from previous test
            assessment_id = self.test_results.get('start_assessment', {}).get('assessment_id')
            
            if not assessment_id:
                self.test_results['websocket_connection'] = {
                    'status': 'SKIPPED',
                    'message': 'No assessment ID available'
                }
                logger.warning("⚠️ WebSocket connection test SKIPPED: No assessment ID")
                return
            
            # Connect to WebSocket
            ws_url = f"{self.ws_url}/ws/risk-assessment/{assessment_id}"
            
            async with websockets.connect(ws_url) as websocket:
                # Wait for initial status message
                initial_message = await asyncio.wait_for(websocket.recv(), timeout=5.0)
                initial_data = json.loads(initial_message)
                
                if initial_data.get('type') == 'initial_status':
                    # Send ping message
                    ping_message = json.dumps({'type': 'ping'})
                    await websocket.send(ping_message)
                    
                    # Wait for pong response
                    pong_message = await asyncio.wait_for(websocket.recv(), timeout=5.0)
                    pong_data = json.loads(pong_message)
                    
                    if pong_data.get('type') == 'pong':
                        self.test_results['websocket_connection'] = {
                            'status': 'PASSED',
                            'message': 'WebSocket connection and ping/pong working',
                            'details': {
                                'initial_message': initial_data,
                                'pong_message': pong_data
                            }
                        }
                        logger.info("✅ WebSocket connection test PASSED")
                    else:
                        self.test_results['websocket_connection'] = {
                            'status': 'FAILED',
                            'message': 'Invalid pong response',
                            'error': 'Expected pong message type'
                        }
                        logger.error("❌ WebSocket connection test FAILED: Invalid pong")
                else:
                    self.test_results['websocket_connection'] = {
                        'status': 'FAILED',
                        'message': 'Invalid initial message',
                        'error': 'Expected initial_status message type'
                    }
                    logger.error("❌ WebSocket connection test FAILED: Invalid initial message")
                    
        except asyncio.TimeoutError:
            self.test_results['websocket_connection'] = {
                'status': 'FAILED',
                'message': 'WebSocket connection timeout',
                'error': 'Connection or message timeout'
            }
            logger.error("❌ WebSocket connection test FAILED: Timeout")
        except Exception as e:
            self.test_results['websocket_connection'] = {
                'status': 'FAILED',
                'message': f'WebSocket connection failed: {str(e)}',
                'error': str(e)
            }
            logger.error(f"❌ WebSocket connection test FAILED: {e}")
    
    async def test_complete_assessment(self):
        """Test completing a real-time risk assessment"""
        logger.info("🧪 Testing complete real-time assessment...")
        
        try:
            # Get assessment ID from previous test
            assessment_id = self.test_results.get('start_assessment', {}).get('assessment_id')
            
            if not assessment_id:
                self.test_results['complete_assessment'] = {
                    'status': 'SKIPPED',
                    'message': 'No assessment ID available'
                }
                logger.warning("⚠️ Complete assessment test SKIPPED: No assessment ID")
                return
            
            # Complete assessment
            response = requests.post(
                f"{self.base_url}/real-time/complete-assessment/{assessment_id}",
                timeout=10
            )
            
            if response.status_code == 200:
                result = response.json()
                self.test_results['complete_assessment'] = {
                    'status': 'PASSED',
                    'message': 'Assessment completed successfully',
                    'details': result
                }
                logger.info("✅ Complete assessment test PASSED")
            else:
                self.test_results['complete_assessment'] = {
                    'status': 'FAILED',
                    'message': f'HTTP {response.status_code}',
                    'error': response.text
                }
                logger.error(f"❌ Complete assessment test FAILED: {response.status_code}")
                
        except Exception as e:
            self.test_results['complete_assessment'] = {
                'status': 'FAILED',
                'message': f'Complete assessment failed: {str(e)}',
                'error': str(e)
            }
            logger.error(f"❌ Complete assessment test FAILED: {e}")
    
    async def test_get_active_assessments(self):
        """Test getting list of active assessments"""
        logger.info("🧪 Testing get active assessments...")
        
        try:
            # Get active assessments
            response = requests.get(
                f"{self.base_url}/real-time/active-assessments",
                timeout=10
            )
            
            if response.status_code == 200:
                result = response.json()
                self.test_results['get_active_assessments'] = {
                    'status': 'PASSED',
                    'message': 'Active assessments retrieved successfully',
                    'details': result
                }
                logger.info("✅ Get active assessments test PASSED")
            else:
                self.test_results['get_active_assessments'] = {
                    'status': 'FAILED',
                    'message': f'HTTP {response.status_code}',
                    'error': response.text
                }
                logger.error(f"❌ Get active assessments test FAILED: {response.status_code}")
                
        except Exception as e:
            self.test_results['get_active_assessments'] = {
                'status': 'FAILED',
                'message': f'Get active assessments failed: {str(e)}',
                'error': str(e)
            }
            logger.error(f"❌ Get active assessments test FAILED: {e}")
    
    async def test_real_time_monitoring(self):
        """Test real-time monitoring endpoint"""
        logger.info("🧪 Testing real-time monitoring...")
        
        try:
            # Get monitoring data
            response = requests.get(
                f"{self.base_url}/real-time/monitoring",
                timeout=10
            )
            
            if response.status_code == 200:
                result = response.json()
                self.test_results['real_time_monitoring'] = {
                    'status': 'PASSED',
                    'message': 'Real-time monitoring data retrieved successfully',
                    'details': result
                }
                logger.info("✅ Real-time monitoring test PASSED")
            else:
                self.test_results['real_time_monitoring'] = {
                    'status': 'FAILED',
                    'message': f'HTTP {response.status_code}',
                    'error': response.text
                }
                logger.error(f"❌ Real-time monitoring test FAILED: {response.status_code}")
                
        except Exception as e:
            self.test_results['real_time_monitoring'] = {
                'status': 'FAILED',
                'message': f'Real-time monitoring failed: {str(e)}',
                'error': str(e)
            }
            logger.error(f"❌ Real-time monitoring test FAILED: {e}")
    
    def generate_test_report(self):
        """Generate comprehensive test report"""
        logger.info("📊 Generating real-time test report...")
        
        # Calculate overall results
        total_tests = len(self.test_results)
        passed_tests = sum(1 for result in self.test_results.values() if result['status'] == 'PASSED')
        failed_tests = sum(1 for result in self.test_results.values() if result['status'] == 'FAILED')
        skipped_tests = sum(1 for result in self.test_results.values() if result['status'] == 'SKIPPED')
        
        # Create summary
        summary = {
            'overall_status': 'PASSED' if failed_tests == 0 else 'FAILED',
            'total_tests': total_tests,
            'passed_tests': passed_tests,
            'failed_tests': failed_tests,
            'skipped_tests': skipped_tests,
            'success_rate': passed_tests / total_tests if total_tests > 0 else 0,
            'timestamp': time.strftime('%Y-%m-%d %H:%M:%S'),
            'test_results': self.test_results
        }
        
        # Save report
        from pathlib import Path
        report_path = Path("test_reports/realtime_risk_assessment_test_report.json")
        report_path.parent.mkdir(exist_ok=True)
        
        with open(report_path, 'w') as f:
            json.dump(summary, f, indent=2)
        
        # Print summary
        logger.info("=" * 60)
        logger.info("🧪 REAL-TIME RISK ASSESSMENT TEST SUMMARY")
        logger.info("=" * 60)
        logger.info(f"Overall Status: {summary['overall_status']}")
        logger.info(f"Total Tests: {total_tests}")
        logger.info(f"Passed: {passed_tests}")
        logger.info(f"Failed: {failed_tests}")
        logger.info(f"Skipped: {skipped_tests}")
        logger.info(f"Success Rate: {summary['success_rate']:.2%}")
        logger.info("=" * 60)
        
        for test_name, result in self.test_results.items():
            status_icon = "✅" if result['status'] == 'PASSED' else "❌" if result['status'] == 'FAILED' else "⚠️"
            logger.info(f"{status_icon} {test_name}: {result['status']} - {result['message']}")
        
        logger.info("=" * 60)
        logger.info(f"📄 Detailed report saved to: {report_path}")
        
        return summary

async def main():
    """Run real-time risk assessment tests"""
    logger.info("🚀 Starting real-time risk assessment tests...")
    
    # Create tester
    tester = RealTimeRiskAssessmentTester()
    
    # Run all tests
    results = await tester.run_all_tests()
    
    # Check if all tests passed
    if results['overall_status'] == 'PASSED':
        logger.info("🎉 All real-time risk assessment tests PASSED!")
        return 0
    else:
        logger.error("💥 Some real-time risk assessment tests FAILED!")
        return 1

if __name__ == "__main__":
    exit(asyncio.run(main()))
