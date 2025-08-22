# Success Rate Benchmarking API Documentation

## Overview

The Success Rate Benchmarking API provides comprehensive tools for measuring, validating, and comparing success rates across different business verification scenarios. This system enables statistical validation of performance improvements and baseline comparison to ensure the effectiveness of optimizations.

## Features

- **Benchmark Suite Management**: Create and manage test suites for different business scenarios
- **Statistical Validation**: Perform statistical analysis of success rates with confidence intervals
- **Baseline Comparison**: Compare current performance against historical baselines
- **Trend Analysis**: Analyze success rate trends over time
- **Comprehensive Reporting**: Generate detailed reports with validation metrics
- **Configuration Management**: Manage benchmarking parameters and thresholds

## API Endpoints

### Base URL
```
https://api.kyb-platform.com/api/v3/benchmarking
```

### Authentication
All endpoints require API key authentication via the `Authorization` header:
```
Authorization: Bearer YOUR_API_KEY
```

## Benchmark Suite Management

### Create Benchmark Suite

**POST** `/suites`

Creates a new benchmark suite with defined test cases and parameters.

**Request Body:**
```json
{
  "id": "business-verification-suite",
  "name": "Business Verification Benchmark Suite",
  "description": "Comprehensive benchmark suite for business verification scenarios",
  "category": "business_verification",
  "test_cases": [
    {
      "id": "basic-verification",
      "name": "Basic Business Verification",
      "description": "Test basic business verification functionality",
      "input": {
        "business_name": "Test Corporation",
        "address": "123 Test Street, Test City, ST 12345"
      },
      "expected_success_rate": 0.95,
      "max_duration": 5000
    },
    {
      "id": "complex-verification",
      "name": "Complex Business Verification",
      "description": "Test complex verification with multiple data sources",
      "input": {
        "business_name": "Complex Corporation",
        "address": "456 Complex Avenue, Complex City, ST 67890",
        "industry": "technology",
        "registration_number": "123456789"
      },
      "expected_success_rate": 0.90,
      "max_duration": 10000
    }
  ],
  "sample_size": 100,
  "max_iterations": 3
}
```

**Response:**
```json
{
  "success": true,
  "suite": {
    "id": "business-verification-suite",
    "name": "Business Verification Benchmark Suite",
    "description": "Comprehensive benchmark suite for business verification scenarios",
    "category": "business_verification",
    "test_cases": [...],
    "sample_size": 100,
    "max_iterations": 3,
    "created_at": "2024-01-15T10:30:00Z"
  },
  "message": "Benchmark suite created successfully"
}
```

### Execute Benchmark

**POST** `/suites/{suiteId}/execute`

Executes a benchmark suite and returns detailed results.

**Path Parameters:**
- `suiteId` (string, required): The ID of the benchmark suite to execute

**Response:**
```json
{
  "success": true,
  "result": {
    "suite_id": "business-verification-suite",
    "execution_id": "exec-123456789",
    "success_rate": 0.94,
    "sample_size": 100,
    "duration": 2500,
    "timestamp": "2024-01-15T10:35:00Z",
    "test_case_results": [
      {
        "test_case_id": "basic-verification",
        "success_rate": 0.96,
        "sample_size": 50,
        "duration": 1200,
        "passed": true
      },
      {
        "test_case_id": "complex-verification",
        "success_rate": 0.92,
        "sample_size": 50,
        "duration": 1300,
        "passed": true
      }
    ],
    "validation": {
      "is_statistically_significant": true,
      "confidence_interval": 0.03,
      "p_value": 0.001,
      "meets_target": true
    }
  },
  "message": "Benchmark executed successfully"
}
```

### Get Benchmark Results

**GET** `/suites/{suiteId}/results`

Retrieves historical benchmark results for a specific suite.

**Path Parameters:**
- `suiteId` (string, required): The ID of the benchmark suite

**Query Parameters:**
- `limit` (integer, optional): Maximum number of results to return (default: all)

**Response:**
```json
{
  "success": true,
  "results": [
    {
      "suite_id": "business-verification-suite",
      "execution_id": "exec-123456789",
      "success_rate": 0.94,
      "sample_size": 100,
      "duration": 2500,
      "timestamp": "2024-01-15T10:35:00Z",
      "validation": {
        "is_statistically_significant": true,
        "confidence_interval": 0.03,
        "p_value": 0.001,
        "meets_target": true
      }
    }
  ],
  "count": 1,
  "suite_id": "business-verification-suite"
}
```

### Generate Benchmark Report

**GET** `/suites/{suiteId}/report`

Generates a comprehensive benchmark report with trend analysis and validation metrics.

**Path Parameters:**
- `suiteId` (string, required): The ID of the benchmark suite

**Response:**
```json
{
  "success": true,
  "report": {
    "suite_id": "business-verification-suite",
    "generated_at": "2024-01-15T10:40:00Z",
    "summary": {
      "total_results": 10,
      "average_success_rate": 0.945,
      "average_duration": 2300,
      "best_success_rate": 0.96,
      "worst_success_rate": 0.92
    },
    "trend_analysis": {
      "success_rate_trend": 0.02,
      "performance_trend": -0.05,
      "stability_score": 0.85,
      "trend_direction": "improving"
    },
    "baseline_comparison": {
      "baseline_success_rate": 0.90,
      "current_success_rate": 0.945,
      "improvement_percentage": 5.0,
      "exceeds_baseline": true,
      "statistical_significance": true
    },
    "validation_summary": {
      "statistically_significant_results": 8,
      "meets_target_results": 9,
      "confidence_level": 0.95,
      "overall_validation_status": "passed"
    }
  },
  "message": "Benchmark report generated successfully"
}
```

## Baseline Management

### Update Baseline

**POST** `/baselines`

Updates baseline metrics for a specific category.

**Request Body:**
```json
{
  "category": "business_verification",
  "success_rate": 0.90,
  "sample_count": 1000
}
```

**Response:**
```json
{
  "success": true,
  "message": "Baseline updated successfully for category: business_verification",
  "category": "business_verification",
  "success_rate": 0.90,
  "sample_count": 1000,
  "updated_at": "2024-01-15T10:45:00Z"
}
```

### Get Baseline Metrics

**GET** `/baselines/{category}`

Retrieves baseline metrics for a specific category.

**Path Parameters:**
- `category` (string, required): The category to retrieve baseline for

**Response:**
```json
{
  "success": true,
  "baseline": {
    "category": "business_verification",
    "success_rate": 0.90,
    "sample_count": 1000,
    "created_at": "2024-01-15T09:00:00Z",
    "updated_at": "2024-01-15T10:45:00Z"
  },
  "category": "business_verification"
}
```

## Configuration Management

### Get Benchmark Configuration

**GET** `/config`

Retrieves current benchmark configuration settings.

**Response:**
```json
{
  "success": true,
  "config": {
    "enable_benchmarking": true,
    "enable_statistical_validation": true,
    "target_success_rate": 0.95,
    "confidence_level": 0.95,
    "min_sample_size": 100,
    "max_sample_size": 10000,
    "validation_threshold": 0.02,
    "trend_analysis_window": 24,
    "baseline_update_frequency": 168
  },
  "retrieved_at": "2024-01-15T10:50:00Z"
}
```

### Update Benchmark Configuration

**PUT** `/config`

Updates benchmark configuration settings.

**Request Body:**
```json
{
  "enable_benchmarking": true,
  "enable_statistical_validation": true,
  "target_success_rate": 0.95,
  "confidence_level": 0.95,
  "min_sample_size": 100,
  "max_sample_size": 10000,
  "validation_threshold": 0.02,
  "trend_analysis_window": 24,
  "baseline_update_frequency": 168
}
```

**Response:**
```json
{
  "success": true,
  "message": "Benchmark configuration updated successfully",
  "updated_at": "2024-01-15T10:55:00Z"
}
```

## Data Models

### BenchmarkSuite
```json
{
  "id": "string",
  "name": "string",
  "description": "string",
  "category": "string",
  "test_cases": [
    {
      "id": "string",
      "name": "string",
      "description": "string",
      "input": "object",
      "expected_success_rate": "number",
      "max_duration": "number"
    }
  ],
  "sample_size": "number",
  "max_iterations": "number",
  "created_at": "string (ISO 8601)"
}
```

### BenchmarkResult
```json
{
  "suite_id": "string",
  "execution_id": "string",
  "success_rate": "number",
  "sample_size": "number",
  "duration": "number",
  "timestamp": "string (ISO 8601)",
  "test_case_results": [
    {
      "test_case_id": "string",
      "success_rate": "number",
      "sample_size": "number",
      "duration": "number",
      "passed": "boolean"
    }
  ],
  "validation": {
    "is_statistically_significant": "boolean",
    "confidence_interval": "number",
    "p_value": "number",
    "meets_target": "boolean"
  }
}
```

### BenchmarkReport
```json
{
  "suite_id": "string",
  "generated_at": "string (ISO 8601)",
  "summary": {
    "total_results": "number",
    "average_success_rate": "number",
    "average_duration": "number",
    "best_success_rate": "number",
    "worst_success_rate": "number"
  },
  "trend_analysis": {
    "success_rate_trend": "number",
    "performance_trend": "number",
    "stability_score": "number",
    "trend_direction": "string"
  },
  "baseline_comparison": {
    "baseline_success_rate": "number",
    "current_success_rate": "number",
    "improvement_percentage": "number",
    "exceeds_baseline": "boolean",
    "statistical_significance": "boolean"
  },
  "validation_summary": {
    "statistically_significant_results": "number",
    "meets_target_results": "number",
    "confidence_level": "number",
    "overall_validation_status": "string"
  }
}
```

## Error Responses

All endpoints return consistent error responses:

```json
{
  "success": false,
  "error": {
    "code": "string",
    "message": "string",
    "details": "object (optional)"
  }
}
```

### Common Error Codes

- `400` - Bad Request: Invalid request parameters or body
- `401` - Unauthorized: Missing or invalid API key
- `404` - Not Found: Resource not found
- `422` - Unprocessable Entity: Validation errors
- `429` - Too Many Requests: Rate limit exceeded
- `500` - Internal Server Error: Server error

## Rate Limits

- **Standard Plan**: 100 requests per minute
- **Professional Plan**: 500 requests per minute
- **Enterprise Plan**: 2000 requests per minute

Rate limit headers are included in all responses:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1642248600
```

## Best Practices

### 1. Benchmark Suite Design
- Create focused test cases with clear success criteria
- Use realistic input data that represents actual usage patterns
- Set appropriate sample sizes for statistical significance
- Define reasonable duration limits for test cases

### 2. Statistical Validation
- Ensure sample sizes meet minimum requirements for statistical significance
- Monitor confidence intervals to assess result reliability
- Use baseline comparisons to measure improvement over time
- Consider trend analysis for long-term performance monitoring

### 3. Configuration Management
- Set appropriate target success rates based on business requirements
- Configure confidence levels based on risk tolerance
- Adjust validation thresholds based on performance expectations
- Regular baseline updates to reflect current performance standards

### 4. Monitoring and Alerting
- Set up alerts for significant performance degradations
- Monitor trend analysis for early warning signs
- Track baseline comparisons to ensure continuous improvement
- Use comprehensive reports for stakeholder communication

## Examples

### Complete Benchmark Workflow

1. **Create Benchmark Suite**
```bash
curl -X POST https://api.kyb-platform.com/api/v3/benchmarking/suites \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d @benchmark-suite.json
```

2. **Execute Benchmark**
```bash
curl -X POST https://api.kyb-platform.com/api/v3/benchmarking/suites/business-verification-suite/execute \
  -H "Authorization: Bearer YOUR_API_KEY"
```

3. **Generate Report**
```bash
curl -X GET https://api.kyb-platform.com/api/v3/benchmarking/suites/business-verification-suite/report \
  -H "Authorization: Bearer YOUR_API_KEY"
```

4. **Update Baseline**
```bash
curl -X POST https://api.kyb-platform.com/api/v3/benchmarking/baselines \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "category": "business_verification",
    "success_rate": 0.94,
    "sample_count": 100
  }'
```

### Python SDK Example

```python
import requests

class SuccessRateBenchmarking:
    def __init__(self, api_key, base_url="https://api.kyb-platform.com"):
        self.api_key = api_key
        self.base_url = base_url
        self.headers = {
            "Authorization": f"Bearer {api_key}",
            "Content-Type": "application/json"
        }
    
    def create_benchmark_suite(self, suite_data):
        response = requests.post(
            f"{self.base_url}/api/v3/benchmarking/suites",
            headers=self.headers,
            json=suite_data
        )
        return response.json()
    
    def execute_benchmark(self, suite_id):
        response = requests.post(
            f"{self.base_url}/api/v3/benchmarking/suites/{suite_id}/execute",
            headers=self.headers
        )
        return response.json()
    
    def generate_report(self, suite_id):
        response = requests.get(
            f"{self.base_url}/api/v3/benchmarking/suites/{suite_id}/report",
            headers=self.headers
        )
        return response.json()

# Usage
benchmarking = SuccessRateBenchmarking("YOUR_API_KEY")

# Create and execute benchmark
suite = benchmarking.create_benchmark_suite({
    "id": "my-benchmark-suite",
    "name": "My Benchmark Suite",
    "category": "business_verification",
    "test_cases": [...],
    "sample_size": 100
})

result = benchmarking.execute_benchmark("my-benchmark-suite")
report = benchmarking.generate_report("my-benchmark-suite")
```

## Support

For technical support or questions about the Success Rate Benchmarking API:

- **Documentation**: [https://docs.kyb-platform.com/benchmarking](https://docs.kyb-platform.com/benchmarking)
- **API Status**: [https://status.kyb-platform.com](https://status.kyb-platform.com)
- **Support Email**: api-support@kyb-platform.com
- **Developer Community**: [https://community.kyb-platform.com](https://community.kyb-platform.com)
