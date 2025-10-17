"""
Locust performance testing for Risk Assessment Service

This file contains comprehensive load testing scenarios for the risk assessment service
including various risk assessment workflows, batch processing, and stress testing.
"""

import json
import random
import time
from locust import HttpUser, task, between, events
from locust.exception import StopUser


class RiskAssessmentUser(HttpUser):
    """Base user class for risk assessment testing"""
    
    wait_time = between(1, 3)  # Wait 1-3 seconds between requests
    
    def on_start(self):
        """Called when a user starts"""
        self.client.verify = False  # Disable SSL verification for testing
        self.test_data = self.generate_test_data()
    
    def generate_test_data(self):
        """Generate test data for risk assessments"""
        industries = [
            "technology", "financial_services", "healthcare", "retail",
            "manufacturing", "energy", "real_estate", "transportation"
        ]
        
        countries = ["US", "CA", "GB", "DE", "FR", "AU", "JP", "SG"]
        
        return {
            "industries": industries,
            "countries": countries,
            "business_names": [
                "Acme Corporation", "Global Tech Solutions", "Innovative Systems Inc",
                "Premier Financial Group", "Advanced Healthcare Corp", "Retail Excellence Ltd",
                "Manufacturing Dynamics", "Energy Solutions Inc", "Real Estate Partners",
                "Transportation Logistics"
            ],
            "addresses": [
                "123 Main Street, New York, NY 10001",
                "456 Business Avenue, Los Angeles, CA 90210",
                "789 Corporate Drive, Chicago, IL 60601",
                "321 Enterprise Boulevard, Houston, TX 77001",
                "654 Innovation Lane, Phoenix, AZ 85001"
            ]
        }
    
    def create_risk_assessment_request(self):
        """Create a random risk assessment request"""
        return {
            "business_name": random.choice(self.test_data["business_names"]),
            "business_address": random.choice(self.test_data["addresses"]),
            "industry": random.choice(self.test_data["industries"]),
            "country": random.choice(self.test_data["countries"]),
            "prediction_horizon": random.randint(1, 12),
            "model_type": random.choice(["xgboost", "lstm", "ensemble"])
        }
    
    @task(10)
    def single_risk_assessment(self):
        """Test single risk assessment endpoint"""
        request_data = self.create_risk_assessment_request()
        
        with self.client.post(
            "/api/v1/assess",
            json=request_data,
            catch_response=True,
            name="single_risk_assessment"
        ) as response:
            if response.status_code == 200:
                try:
                    data = response.json()
                    if "id" in data and "risk_score" in data:
                        response.success()
                    else:
                        response.failure("Invalid response format")
                except json.JSONDecodeError:
                    response.failure("Invalid JSON response")
            elif response.status_code == 429:
                response.failure("Rate limited")
            else:
                response.failure(f"Unexpected status code: {response.status_code}")
    
    @task(5)
    def batch_risk_assessment(self):
        """Test batch risk assessment endpoint"""
        batch_size = random.randint(5, 20)
        batch_requests = []
        
        for _ in range(batch_size):
            batch_requests.append(self.create_risk_assessment_request())
        
        with self.client.post(
            "/api/v1/assess/batch",
            json={"requests": batch_requests},
            catch_response=True,
            name="batch_risk_assessment"
        ) as response:
            if response.status_code == 200:
                try:
                    data = response.json()
                    if "batch_id" in data and "results" in data:
                        response.success()
                    else:
                        response.failure("Invalid batch response format")
                except json.JSONDecodeError:
                    response.failure("Invalid JSON response")
            elif response.status_code == 429:
                response.failure("Rate limited")
            else:
                response.failure(f"Unexpected status code: {response.status_code}")
    
    @task(3)
    def get_risk_assessment(self):
        """Test getting a risk assessment by ID"""
        # First create a risk assessment
        request_data = self.create_risk_assessment_request()
        
        with self.client.post(
            "/api/v1/assess",
            json=request_data,
            catch_response=True,
            name="create_for_get"
        ) as response:
            if response.status_code == 200:
                try:
                    data = response.json()
                    assessment_id = data.get("id")
                    
                    if assessment_id:
                        # Now get the assessment
                        with self.client.get(
                            f"/api/v1/assess/{assessment_id}",
                            catch_response=True,
                            name="get_risk_assessment"
                        ) as get_response:
                            if get_response.status_code == 200:
                                get_response.success()
                            else:
                                get_response.failure(f"Failed to get assessment: {get_response.status_code}")
                    else:
                        response.failure("No assessment ID in response")
                except json.JSONDecodeError:
                    response.failure("Invalid JSON response")
            else:
                response.failure(f"Failed to create assessment: {response.status_code}")
    
    @task(2)
    def list_risk_assessments(self):
        """Test listing risk assessments"""
        params = {
            "limit": random.randint(10, 50),
            "offset": random.randint(0, 100),
            "status": random.choice(["pending", "completed", "failed", ""])
        }
        
        with self.client.get(
            "/api/v1/assess",
            params=params,
            catch_response=True,
            name="list_risk_assessments"
        ) as response:
            if response.status_code == 200:
                try:
                    data = response.json()
                    if "assessments" in data and "total" in data:
                        response.success()
                    else:
                        response.failure("Invalid list response format")
                except json.JSONDecodeError:
                    response.failure("Invalid JSON response")
            elif response.status_code == 429:
                response.failure("Rate limited")
            else:
                response.failure(f"Unexpected status code: {response.status_code}")
    
    @task(1)
    def health_check(self):
        """Test health check endpoint"""
        with self.client.get(
            "/health",
            catch_response=True,
            name="health_check"
        ) as response:
            if response.status_code == 200:
                response.success()
            else:
                response.failure(f"Health check failed: {response.status_code}")
    
    @task(1)
    def metrics_endpoint(self):
        """Test metrics endpoint"""
        with self.client.get(
            "/metrics",
            catch_response=True,
            name="metrics_endpoint"
        ) as response:
            if response.status_code == 200:
                response.success()
            else:
                response.failure(f"Metrics endpoint failed: {response.status_code}")


class HighVolumeUser(RiskAssessmentUser):
    """User class for high-volume testing"""
    
    wait_time = between(0.1, 0.5)  # Faster requests for high volume
    
    @task(20)
    def rapid_risk_assessment(self):
        """Test rapid risk assessment requests"""
        request_data = self.create_risk_assessment_request()
        
        with self.client.post(
            "/api/v1/assess",
            json=request_data,
            catch_response=True,
            name="rapid_risk_assessment"
        ) as response:
            if response.status_code == 200:
                response.success()
            elif response.status_code == 429:
                response.failure("Rate limited")
            else:
                response.failure(f"Unexpected status code: {response.status_code}")


class BatchProcessingUser(RiskAssessmentUser):
    """User class for batch processing testing"""
    
    wait_time = between(5, 10)  # Longer wait for batch processing
    
    @task(10)
    def large_batch_assessment(self):
        """Test large batch risk assessment"""
        batch_size = random.randint(50, 100)
        batch_requests = []
        
        for _ in range(batch_size):
            batch_requests.append(self.create_risk_assessment_request())
        
        with self.client.post(
            "/api/v1/assess/batch",
            json={"requests": batch_requests},
            catch_response=True,
            name="large_batch_assessment"
        ) as response:
            if response.status_code == 200:
                try:
                    data = response.json()
                    if "batch_id" in data:
                        response.success()
                    else:
                        response.failure("Invalid batch response format")
                except json.JSONDecodeError:
                    response.failure("Invalid JSON response")
            elif response.status_code == 429:
                response.failure("Rate limited")
            else:
                response.failure(f"Unexpected status code: {response.status_code}")


class StressTestUser(RiskAssessmentUser):
    """User class for stress testing"""
    
    wait_time = between(0.01, 0.1)  # Very fast requests for stress testing
    
    @task(50)
    def stress_risk_assessment(self):
        """Test stress risk assessment requests"""
        request_data = self.create_risk_assessment_request()
        
        with self.client.post(
            "/api/v1/assess",
            json=request_data,
            catch_response=True,
            name="stress_risk_assessment"
        ) as response:
            if response.status_code == 200:
                response.success()
            elif response.status_code == 429:
                response.failure("Rate limited")
            elif response.status_code == 503:
                response.failure("Service unavailable")
            else:
                response.failure(f"Unexpected status code: {response.status_code}")


class ScenarioAnalysisUser(RiskAssessmentUser):
    """User class for scenario analysis testing"""
    
    wait_time = between(2, 5)
    
    @task(5)
    def scenario_analysis(self):
        """Test scenario analysis endpoint"""
        base_request = self.create_risk_assessment_request()
        scenarios = [
            {"name": "optimistic", "multiplier": 0.8},
            {"name": "realistic", "multiplier": 1.0},
            {"name": "pessimistic", "multiplier": 1.2}
        ]
        
        request_data = {
            "base_request": base_request,
            "scenarios": scenarios
        }
        
        with self.client.post(
            "/api/v1/assess/scenarios",
            json=request_data,
            catch_response=True,
            name="scenario_analysis"
        ) as response:
            if response.status_code == 200:
                try:
                    data = response.json()
                    if "scenarios" in data and len(data["scenarios"]) == 3:
                        response.success()
                    else:
                        response.failure("Invalid scenario response format")
                except json.JSONDecodeError:
                    response.failure("Invalid JSON response")
            elif response.status_code == 429:
                response.failure("Rate limited")
            else:
                response.failure(f"Unexpected status code: {response.status_code}")


# Custom event handlers for monitoring
@events.request.add_listener
def on_request(request_type, name, response_time, response_length, exception, context, **kwargs):
    """Log request details for monitoring"""
    if exception:
        print(f"Request failed: {name} - {exception}")
    elif response_time > 5000:  # Log slow requests (>5 seconds)
        print(f"Slow request: {name} - {response_time}ms")


@events.user_error.add_listener
def on_user_error(user_instance, exception, tb, **kwargs):
    """Handle user errors"""
    print(f"User error: {exception}")


# Test configuration classes
class LoadTestConfig:
    """Configuration for load testing"""
    
    def __init__(self, host="http://localhost:8080", users=100, spawn_rate=10, run_time="5m"):
        self.host = host
        self.users = users
        self.spawn_rate = spawn_rate
        self.run_time = run_time


class StressTestConfig:
    """Configuration for stress testing"""
    
    def __init__(self, host="http://localhost:8080", users=500, spawn_rate=50, run_time="10m"):
        self.host = host
        self.users = users
        self.spawn_rate = spawn_rate
        self.run_time = run_time


class SpikeTestConfig:
    """Configuration for spike testing"""
    
    def __init__(self, host="http://localhost:8080", users=1000, spawn_rate=100, run_time="2m"):
        self.host = host
        self.users = users
        self.spawn_rate = spawn_rate
        self.run_time = run_time


# Performance thresholds
PERFORMANCE_THRESHOLDS = {
    "response_time_p95": 1000,  # 95th percentile response time < 1 second
    "response_time_p99": 2000,  # 99th percentile response time < 2 seconds
    "error_rate": 0.01,         # Error rate < 1%
    "throughput": 1000,         # Minimum throughput of 1000 requests/minute
}


def validate_performance_metrics(stats):
    """Validate performance metrics against thresholds"""
    results = {}
    
    # Check response time percentiles
    if stats.get("response_time_p95", 0) > PERFORMANCE_THRESHOLDS["response_time_p95"]:
        results["response_time_p95"] = f"FAIL: {stats['response_time_p95']}ms > {PERFORMANCE_THRESHOLDS['response_time_p95']}ms"
    else:
        results["response_time_p95"] = f"PASS: {stats['response_time_p95']}ms <= {PERFORMANCE_THRESHOLDS['response_time_p95']}ms"
    
    if stats.get("response_time_p99", 0) > PERFORMANCE_THRESHOLDS["response_time_p99"]:
        results["response_time_p99"] = f"FAIL: {stats['response_time_p99']}ms > {PERFORMANCE_THRESHOLDS['response_time_p99']}ms"
    else:
        results["response_time_p99"] = f"PASS: {stats['response_time_p99']}ms <= {PERFORMANCE_THRESHOLDS['response_time_p99']}ms"
    
    # Check error rate
    error_rate = stats.get("error_rate", 0)
    if error_rate > PERFORMANCE_THRESHOLDS["error_rate"]:
        results["error_rate"] = f"FAIL: {error_rate:.2%} > {PERFORMANCE_THRESHOLDS['error_rate']:.2%}"
    else:
        results["error_rate"] = f"PASS: {error_rate:.2%} <= {PERFORMANCE_THRESHOLDS['error_rate']:.2%}"
    
    # Check throughput
    throughput = stats.get("throughput", 0)
    if throughput < PERFORMANCE_THRESHOLDS["throughput"]:
        results["throughput"] = f"FAIL: {throughput} req/min < {PERFORMANCE_THRESHOLDS['throughput']} req/min"
    else:
        results["throughput"] = f"PASS: {throughput} req/min >= {PERFORMANCE_THRESHOLDS['throughput']} req/min"
    
    return results


if __name__ == "__main__":
    # This file can be run directly with: locust -f locustfile.py
    print("Risk Assessment Service Performance Testing")
    print("Available user classes:")
    print("- RiskAssessmentUser: Standard risk assessment testing")
    print("- HighVolumeUser: High-volume testing")
    print("- BatchProcessingUser: Batch processing testing")
    print("- StressTestUser: Stress testing")
    print("- ScenarioAnalysisUser: Scenario analysis testing")
    print("\nRun with: locust -f locustfile.py --host=http://localhost:8080")
