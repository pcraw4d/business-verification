#!/bin/bash

# Integration Testing Script
echo "🔗 Running Compliance Integration Tests"
echo "======================================"

# Set test environment variables
export TEST_ENV="integration"
export LOG_LEVEL="info"

# Test 1: API Integration
echo "🌐 Testing API Integration..."
go test -run TestComplianceAPIIntegration -v ./test/compliance/integration_test.go

# Test 2: Service Integration
echo "⚙️  Testing Service Integration..."
go test -run TestComplianceServiceIntegration -v ./test/compliance/integration_test.go

# Test 3: Component Integration
echo "🧩 Testing Component Integration..."
go test -run TestComplianceComponentIntegration -v ./test/compliance/integration_test.go

# Test 4: End-to-End Integration
echo "🔄 Testing End-to-End Integration..."
go test -run TestComplianceEndToEndIntegration -v ./test/compliance/integration_test.go

# Test 5: Run all integration tests together
echo "🎯 Running All Integration Tests..."
go test -v ./test/compliance/integration_test.go

echo "✅ All Integration Tests Completed!"
echo ""
echo "📊 Test Summary:"
echo "- API Integration: Validates integration between compliance APIs and handlers"
echo "- Service Integration: Validates integration between compliance services"
echo "- Component Integration: Validates integration between compliance components"
echo "- End-to-End Integration: Validates complete end-to-end workflows"
echo "- Multi-Framework Integration: Validates integration across multiple frameworks"
echo "- Data Consistency Integration: Validates data consistency across components"
echo "- Error Handling Integration: Validates error handling across components"
echo "- Cross-Component Data Flow: Validates data flow between components"
echo ""
echo "🔍 Check test output above for detailed results"
