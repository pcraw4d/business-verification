#!/bin/bash

# Compliance API Testing Script
echo "🧪 Running Compliance API Tests"
echo "================================"

# Test 1: API Endpoint Testing
echo "📡 Testing API Endpoints..."
go test -run TestComplianceAPIEndpoints -v ./test/compliance/api_endpoint_test.go

# Test 2: Service Integration Testing  
echo "🔗 Testing Service Integration..."
go test -run TestComplianceServiceIntegration -v ./test/compliance/service_integration_test.go

# Test 3: Compliance Calculation Testing
echo "🧮 Testing Compliance Calculations..."
go test -run TestComplianceCalculationAccuracy -v ./test/compliance/calculation_test.go

# Test 4: Reporting Accuracy Testing
echo "📊 Testing Reporting Accuracy..."
go test -run TestReportingAccuracy -v ./test/compliance/reporting_accuracy_test.go

# Test 5: Alert System Testing
echo "🚨 Testing Alert System..."
go test -run TestAlertSystemFunctionality -v ./test/compliance/alert_system_test.go

echo "✅ All Compliance Tests Completed!"
