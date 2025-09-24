#!/bin/bash

# Security Testing Script for Subtask 4.2.4
# This script runs comprehensive security tests for the KYB Platform

set -e

echo "🔐 Starting Security Testing for Subtask 4.2.4..."
echo "=================================================="

# Create reports directory
mkdir -p test/reports/security

# Run security tests
echo "Running comprehensive security tests..."
go test ./test/security/... -v -timeout 30s

# Check if tests passed
if [ $? -eq 0 ]; then
    echo ""
    echo "✅ Security tests completed successfully!"
    echo ""
    echo "📊 Test Reports Generated:"
    echo "  - test/reports/security/security_test_results.json"
    echo "  - test/reports/security/security_test_report.md"
    echo "  - test/reports/security/security_summary.md"
    echo ""
    echo "📋 Summary:"
    if [ -f "test/reports/security/security_summary.md" ]; then
        cat test/reports/security/security_summary.md
    fi
else
    echo ""
    echo "❌ Security tests failed!"
    echo "Please review the test output and fix any issues."
    exit 1
fi

echo ""
echo "🎯 Security Testing for Subtask 4.2.4 Complete!"
echo "=================================================="
