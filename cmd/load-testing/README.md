# KYB Platform Load Testing

This tool performs comprehensive load testing on the KYB platform to identify performance bottlenecks and optimization opportunities.

## Features

- **Multi-endpoint Testing**: Tests health, metrics, analytics, and classification endpoints
- **Concurrent User Simulation**: Simulates multiple concurrent users
- **Stress Testing**: Gradually increases load to find breaking points
- **Performance Metrics**: Measures response times, throughput, and error rates
- **Real-time Results**: Provides detailed performance analysis

## Usage

```bash
cd cmd/load-testing
go run main.go
```

## Test Configuration

- **Concurrent Users**: 50 (configurable)
- **Requests per User**: 20 (configurable)
- **Test Duration**: 30 seconds
- **Stress Test Levels**: 10, 25, 50, 100, 200 concurrent users

## Performance Targets

- **Error Rate**: < 5%
- **Response Time**: < 500ms
- **Throughput**: > 10 requests/second
- **Success Rate**: > 95%

## Results Interpretation

- ✅ **Good Performance**: All metrics within targets
- ⚠️ **Needs Optimization**: One or more metrics outside targets
- 🔥 **Critical Issues**: Multiple metrics significantly outside targets
