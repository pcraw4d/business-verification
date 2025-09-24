# Comprehensive API Endpoint Testing Guide

## Overview

This document provides comprehensive guidance for testing all API endpoints in the KYB Platform as specified in **Subtask 4.2.1** of the Supabase Table Improvement Implementation Plan. The testing suite ensures that all business-related, classification, user management, and monitoring endpoints function correctly and meet performance requirements.

## Table of Contents

1. [Test Objectives](#test-objectives)
2. [Test Categories](#test-categories)
3. [Test Infrastructure](#test-infrastructure)
4. [Running Tests](#running-tests)
5. [Test Results Analysis](#test-results-analysis)
6. [Performance Testing](#performance-testing)
7. [Error Handling Testing](#error-handling-testing)
8. [Security Testing](#security-testing)
9. [Troubleshooting](#troubleshooting)
10. [Best Practices](#best-practices)

## Test Objectives

### Primary Objectives

1. **Comprehensive Coverage**: Test all API endpoints across all categories
2. **Functional Validation**: Ensure all endpoints return expected responses
3. **Performance Validation**: Verify endpoints meet performance requirements
4. **Error Handling**: Validate proper error responses and status codes
5. **Security Validation**: Ensure endpoints are properly secured
6. **Integration Validation**: Verify endpoints work together correctly

### Success Criteria

- ✅ All business-related endpoints return correct responses
- ✅ All classification endpoints function with ML models
- ✅ All user management endpoints handle authentication properly
- ✅ All monitoring endpoints provide accurate system information
- ✅ Response times meet performance thresholds
- ✅ Error handling is comprehensive and informative
- ✅ Security measures are properly implemented

## Test Categories

### 1. Business-Related Endpoints

#### Classification Endpoints
- **POST /v1/classify**: Single business classification
- **POST /v1/classify/batch**: Batch business classification
- **GET /v1/classify/{business_id}**: Retrieve classification results
- **GET /v1/classify/history**: Get classification history

#### Merchant Management Endpoints
- **POST /v1/merchants**: Create new merchant
- **GET /v1/merchants/{merchant_id}**: Retrieve merchant information
- **PUT /v1/merchants/{merchant_id}**: Update merchant information
- **DELETE /v1/merchants/{merchant_id}**: Delete merchant

#### Risk Assessment Endpoints
- **POST /v1/risk/assess**: Perform risk assessment
- **GET /v1/risk/{business_id}**: Get risk assessment results
- **GET /v1/risk/history/{business_id}**: Get risk assessment history
- **POST /v1/risk/enhanced/assess**: Enhanced risk assessment
- **GET /v1/risk/alerts**: Get risk alerts

### 2. Classification Endpoints

#### Enhanced Classification
- **POST /v2/classify**: Enhanced classification with ML models
- **POST /v2/classify/batch**: Enhanced batch classification

#### Classification Monitoring
- **GET /v1/monitoring/accuracy/metrics**: Get accuracy metrics
- **POST /v1/monitoring/accuracy/track**: Track classification accuracy
- **GET /v1/monitoring/misclassifications**: Get misclassification data
- **GET /v1/monitoring/patterns**: Get error patterns
- **GET /v1/monitoring/statistics**: Get error statistics

### 3. User Management Endpoints

#### Authentication
- **POST /v1/auth/register**: User registration
- **POST /v1/auth/login**: User login
- **POST /v1/auth/logout**: User logout
- **POST /v1/auth/refresh**: Token refresh

#### Profile Management
- **GET /v1/users/profile**: Get user profile
- **PUT /v1/users/profile**: Update user profile

#### API Key Management
- **GET /v1/users/api-keys**: Get API keys
- **POST /v1/users/api-keys**: Create API key
- **DELETE /v1/users/api-keys/{key_id}**: Delete API key

### 4. Monitoring Endpoints

#### Health and Status
- **GET /health**: System health check
- **GET /v1/status**: Detailed system status

#### Metrics and Analytics
- **GET /v1/metrics**: System performance metrics
- **GET /v1/analytics/classification**: Classification analytics
- **GET /v1/analytics/performance**: Performance analytics

#### Monitoring and Alerting
- **GET /v1/monitoring/alerts**: Get active alerts
- **GET /v1/monitoring/alerts/history**: Get alert history
- **POST /v1/monitoring/alerts/{alert_id}/resolve**: Resolve alert

#### Compliance
- **POST /v1/compliance/check**: Check compliance
- **GET /v1/compliance/status/{business_id}**: Get compliance status
- **GET /v1/compliance/reports**: Get compliance reports

## Test Infrastructure

### Test Files

```
test/
├── integration/
│   └── comprehensive_api_endpoint_test.go    # Main test suite
├── scripts/
│   └── run_comprehensive_api_tests.sh        # Test runner script
├── config/
│   └── comprehensive_api_test_config.yaml    # Test configuration
├── docs/
│   └── comprehensive_api_endpoint_testing_guide.md  # This guide
└── reports/
    ├── comprehensive_test_results.txt        # Test results
    ├── performance_test_results.txt          # Performance results
    ├── error_handling_test_results.txt       # Error handling results
    └── comprehensive_api_test_report.md      # Comprehensive report
```

### Test Configuration

The test configuration is defined in `test/config/comprehensive_api_test_config.yaml` and includes:

- **Test Categories**: Configuration for each endpoint category
- **Test Data**: Sample data for testing
- **Performance Thresholds**: Response time requirements
- **Error Handling**: Expected error responses
- **Security Tests**: Security validation scenarios
- **Mock Services**: Mock service configurations

### Mock Services

The test suite uses comprehensive mock services to simulate real API behavior:

- **MockClassificationService**: Simulates classification operations
- **MockAuthService**: Simulates authentication operations
- **MockRiskService**: Simulates risk assessment operations
- **MockMonitoringService**: Simulates monitoring operations

## Running Tests

### Prerequisites

1. **Go Environment**: Go 1.22 or later
2. **Test Dependencies**: All required packages installed
3. **Environment Variables**: Proper test environment configuration
4. **Mock Services**: Mock services properly configured

### Quick Start

Run all comprehensive API endpoint tests:

```bash
# Run all tests
./test/scripts/run_comprehensive_api_tests.sh

# Run with specific options
./test/scripts/run_comprehensive_api_tests.sh --integration-only
./test/scripts/run_comprehensive_api_tests.sh --performance-only
./test/scripts/run_comprehensive_api_tests.sh --verbose
```

### Environment Variables

```bash
# Enable integration tests
export INTEGRATION_TESTS=true

# Enable performance tests
export PERFORMANCE_TESTS=true

# Enable coverage reporting
export COVERAGE=true

# Enable verbose output
export VERBOSE=true
```

### Individual Test Categories

```bash
# Run only business-related endpoint tests
go test -run TestComprehensiveAPIEndpoints/BusinessRelatedEndpoints ./test/integration/

# Run only classification endpoint tests
go test -run TestComprehensiveAPIEndpoints/ClassificationEndpoints ./test/integration/

# Run only user management endpoint tests
go test -run TestComprehensiveAPIEndpoints/UserManagementEndpoints ./test/integration/

# Run only monitoring endpoint tests
go test -run TestComprehensiveAPIEndpoints/MonitoringEndpoints ./test/integration/
```

## Test Results Analysis

### Test Output

The test suite generates comprehensive reports including:

1. **Test Results**: Detailed results for each endpoint
2. **Performance Metrics**: Response times and throughput
3. **Coverage Reports**: Code coverage analysis
4. **Error Analysis**: Error handling validation
5. **Security Reports**: Security test results

### Success Metrics

- **Response Time**: < 2 seconds for classification, < 100ms for health checks
- **Success Rate**: > 99% for all endpoints
- **Error Handling**: Proper HTTP status codes and error messages
- **Security**: All security tests pass
- **Coverage**: > 90% code coverage

### Failure Analysis

When tests fail, the suite provides:

1. **Detailed Error Messages**: Specific failure reasons
2. **Response Analysis**: Actual vs expected responses
3. **Performance Analysis**: Response time breakdowns
4. **Debugging Information**: Request/response details

## Performance Testing

### Performance Thresholds

| Endpoint Category | Max Response Time | Target Success Rate |
|------------------|-------------------|-------------------|
| Classification | 2 seconds | 99.9% |
| Health Checks | 100ms | 99.9% |
| Metrics | 500ms | 99.9% |
| Batch Operations | 5 seconds | 99.5% |
| Risk Assessment | 3 seconds | 99.9% |

### Load Testing

The performance tests include:

1. **Concurrent Requests**: Multiple simultaneous requests
2. **Sustained Load**: Continuous load over time
3. **Peak Load**: Maximum expected load
4. **Stress Testing**: Beyond normal capacity

### Performance Monitoring

Performance tests monitor:

- **Response Time**: End-to-end response time
- **Throughput**: Requests per second
- **Resource Usage**: CPU, memory, database connections
- **Error Rate**: Failed requests percentage

## Error Handling Testing

### Error Scenarios

The test suite validates error handling for:

1. **Invalid Input**: Malformed JSON, missing fields
2. **Authentication Errors**: Invalid tokens, expired sessions
3. **Authorization Errors**: Insufficient permissions
4. **Resource Errors**: Non-existent resources
5. **System Errors**: Database failures, service unavailability

### Expected Error Responses

| Error Type | HTTP Status | Response Format |
|------------|-------------|-----------------|
| Bad Request | 400 | JSON error message |
| Unauthorized | 401 | Authentication error |
| Forbidden | 403 | Authorization error |
| Not Found | 404 | Resource not found |
| Internal Error | 500 | System error message |

### Error Message Validation

Error messages must include:

- **Error Code**: Unique error identifier
- **Message**: Human-readable error description
- **Details**: Additional error context
- **Timestamp**: Error occurrence time
- **Request ID**: Unique request identifier

## Security Testing

### Security Test Categories

1. **Authentication**: Token validation, session management
2. **Authorization**: Permission checking, role validation
3. **Input Validation**: SQL injection, XSS prevention
4. **Rate Limiting**: Request throttling, abuse prevention
5. **Data Protection**: Sensitive data handling

### Security Test Cases

```yaml
# SQL Injection Test
- name: "SQL injection attempt"
  method: "POST"
  path: "/v1/classify"
  body:
    business_name: "'; DROP TABLE users; --"
  expected_status: 400

# XSS Test
- name: "XSS attempt"
  method: "POST"
  path: "/v1/classify"
  body:
    business_name: "<script>alert('xss')</script>"
  expected_status: 400

# Rate Limiting Test
- name: "Rate limiting test"
  method: "POST"
  path: "/v1/classify"
  requests_per_minute: 100
  expected_status: 429
```

## Troubleshooting

### Common Issues

#### Test Failures

1. **Mock Service Issues**: Ensure mock services are properly configured
2. **Environment Variables**: Verify all required environment variables are set
3. **Dependencies**: Ensure all required packages are installed
4. **Network Issues**: Check network connectivity for external dependencies

#### Performance Issues

1. **Slow Response Times**: Check system resources and database performance
2. **High Error Rates**: Verify service health and configuration
3. **Resource Exhaustion**: Monitor CPU, memory, and database connections

#### Security Issues

1. **Authentication Failures**: Verify token generation and validation
2. **Authorization Errors**: Check permission configurations
3. **Input Validation**: Ensure proper input sanitization

### Debug Mode

Enable debug mode for detailed troubleshooting:

```bash
export VERBOSE=true
export DEBUG=true
./test/scripts/run_comprehensive_api_tests.sh
```

### Log Analysis

Review test logs for detailed information:

```bash
# View test logs
tail -f test/logs/comprehensive_api_tests.log

# Search for specific errors
grep "ERROR" test/logs/comprehensive_api_tests.log

# Analyze performance issues
grep "response_time" test/logs/comprehensive_api_tests.log
```

## Best Practices

### Test Development

1. **Comprehensive Coverage**: Test all endpoints and scenarios
2. **Realistic Data**: Use realistic test data that mirrors production
3. **Error Scenarios**: Test both success and failure cases
4. **Performance Validation**: Include performance requirements
5. **Security Testing**: Validate security measures

### Test Maintenance

1. **Regular Updates**: Keep tests updated with API changes
2. **Documentation**: Maintain comprehensive test documentation
3. **Monitoring**: Monitor test results and performance trends
4. **Automation**: Integrate tests into CI/CD pipeline
5. **Review**: Regular review of test effectiveness

### Test Execution

1. **Environment Isolation**: Use isolated test environments
2. **Data Cleanup**: Clean up test data after execution
3. **Parallel Execution**: Run tests in parallel when possible
4. **Resource Management**: Monitor and manage test resources
5. **Reporting**: Generate comprehensive test reports

## Integration with Implementation Plan

This comprehensive API endpoint testing directly supports **Subtask 4.2.1** of the Supabase Table Improvement Implementation Plan by:

1. **Validating All Endpoints**: Ensures all API endpoints function correctly
2. **Performance Validation**: Verifies endpoints meet performance requirements
3. **Error Handling**: Validates comprehensive error handling
4. **Security Validation**: Ensures proper security measures
5. **Integration Testing**: Verifies endpoints work together correctly

The test results provide critical validation that the enhanced database schema and improved classification system are working correctly through all API endpoints, supporting the overall goal of creating a best-in-class merchant risk and verification product.

## Next Steps

After completing the comprehensive API endpoint testing:

1. **Review Results**: Analyze all test results and address any failures
2. **Performance Optimization**: Optimize any endpoints that don't meet performance requirements
3. **Error Handling Enhancement**: Improve error handling based on test findings
4. **Security Hardening**: Address any security issues identified
5. **Documentation Updates**: Update API documentation based on test results
6. **Continuous Monitoring**: Implement continuous monitoring for API endpoints

This testing foundation ensures that the KYB Platform's API endpoints are robust, performant, secure, and ready for production use.
