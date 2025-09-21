#!/bin/bash

# Enhanced Classification System Test Runner
# This script runs comprehensive tests for the enhanced classification system
# and generates detailed reports for subtask 1.5.4

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
TEST_DIR="test"
REPORT_DIR="test_reports"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
REPORT_FILE="${REPORT_DIR}/enhanced_classification_test_report_${TIMESTAMP}.md"

# Create report directory
mkdir -p "$REPORT_DIR"

echo -e "${BLUE}🧪 Enhanced Classification System Test Runner${NC}"
echo -e "${BLUE}================================================${NC}"
echo ""

# Function to print section headers
print_section() {
    echo -e "\n${YELLOW}📋 $1${NC}"
    echo -e "${YELLOW}$(printf '=%.0s' {1..50})${NC}"
}

# Function to print test results
print_test_result() {
    local test_name="$1"
    local status="$2"
    local duration="$3"
    
    if [ "$status" = "PASS" ]; then
        echo -e "✅ ${GREEN}$test_name${NC} - ${GREEN}PASSED${NC} (${duration})"
    else
        echo -e "❌ ${RED}$test_name${NC} - ${RED}FAILED${NC} (${duration})"
    fi
}

# Function to generate test report
generate_report() {
    local report_file="$1"
    
    cat > "$report_file" << EOF
# Enhanced Classification System Test Report

**Generated**: $(date)
**Test Suite**: Enhanced Classification System (Subtask 1.5.4)
**Environment**: $(uname -s) $(uname -r)

## 🎯 Test Objectives

This test suite validates the enhanced classification system implementation including:

1. **Risk Keyword Detection** - Comprehensive testing of risk keyword matching, synonym detection, and pattern recognition
2. **Code Crosswalk Functionality** - Validation of MCC/NAICS/SIC code mapping and crosswalk validation
3. **Business Risk Assessment Workflow** - End-to-end testing of the complete risk assessment process
4. **UI Integration Points** - Verification of data formats and display compatibility
5. **Performance Testing** - Large dataset and concurrent processing validation

## 📊 Test Results Summary

EOF
}

# Function to run individual test categories
run_test_category() {
    local category="$1"
    local test_pattern="$2"
    local description="$3"
    
    print_section "Testing $description"
    
    local start_time=$(date +%s)
    local test_output=$(go test -v -run "$test_pattern" ./test/ 2>&1)
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    if echo "$test_output" | grep -q "PASS"; then
        print_test_result "$description" "PASS" "${duration}s"
        echo "$test_output" | grep "PASS\|FAIL" | while read line; do
            echo "  $line"
        done
        return 0
    else
        print_test_result "$description" "FAIL" "${duration}s"
        echo "$test_output" | tail -20
        return 1
    fi
}

# Function to run performance benchmarks
run_performance_benchmarks() {
    print_section "Performance Benchmarks"
    
    echo "Running performance benchmarks..."
    
    # Risk keyword detection benchmark
    echo "🔍 Risk Keyword Detection Performance:"
    go test -bench=BenchmarkRiskKeywordDetection -benchmem ./test/ 2>/dev/null || echo "  Benchmark not implemented yet"
    
    # Crosswalk query benchmark
    echo "🔗 Crosswalk Query Performance:"
    go test -bench=BenchmarkCrosswalkQueries -benchmem ./test/ 2>/dev/null || echo "  Benchmark not implemented yet"
    
    # Risk assessment benchmark
    echo "⚖️ Risk Assessment Performance:"
    go test -bench=BenchmarkRiskAssessment -benchmem ./test/ 2>/dev/null || echo "  Benchmark not implemented yet"
}

# Function to validate database connectivity
validate_database() {
    print_section "Database Connectivity Validation"
    
    if [ -z "$DATABASE_URL" ]; then
        echo -e "${YELLOW}⚠️  DATABASE_URL not set. Some tests may be skipped.${NC}"
        return 1
    fi
    
    echo "Testing database connection..."
    if go run -c "package main; import \"database/sql\"; import _ \"github.com/lib/pq\"; func main() { db, err := sql.Open(\"postgres\", \"$DATABASE_URL\"); if err != nil { panic(err) }; defer db.Close(); if err := db.Ping(); err != nil { panic(err) }; println(\"Database connection successful\") }" 2>/dev/null; then
        echo -e "✅ ${GREEN}Database connection successful${NC}"
        return 0
    else
        echo -e "❌ ${RED}Database connection failed${NC}"
        return 1
    fi
}

# Function to check test dependencies
check_dependencies() {
    print_section "Dependency Validation"
    
    local missing_deps=()
    
    # Check Go
    if ! command -v go &> /dev/null; then
        missing_deps+=("go")
    else
        echo -e "✅ ${GREEN}Go $(go version | cut -d' ' -f3)${NC}"
    fi
    
    # Check required Go packages
    local required_packages=(
        "github.com/stretchr/testify"
        "github.com/lib/pq"
    )
    
    for package in "${required_packages[@]}"; do
        if go list -m "$package" &> /dev/null; then
            echo -e "✅ ${GREEN}$package${NC}"
        else
            echo -e "❌ ${RED}$package (missing)${NC}"
            missing_deps+=("$package")
        fi
    done
    
    if [ ${#missing_deps[@]} -gt 0 ]; then
        echo -e "\n${YELLOW}Installing missing dependencies...${NC}"
        for dep in "${missing_deps[@]}"; do
            if [[ "$dep" == go ]]; then
                echo -e "${RED}Please install Go from https://golang.org/dl/${NC}"
                exit 1
            else
                go get "$dep"
            fi
        done
    fi
}

# Function to generate detailed test report
generate_detailed_report() {
    local report_file="$1"
    
    cat >> "$report_file" << EOF

## 🔍 Detailed Test Results

### Risk Keyword Detection Tests
- **Direct Keyword Matching**: Tests exact keyword detection in content
- **Synonym Matching**: Tests detection of keyword variations and synonyms  
- **Pattern Matching**: Tests regex pattern-based detection
- **Low Risk Content**: Tests handling of legitimate business content
- **Confidence Scoring**: Tests accuracy of confidence calculations

### Code Crosswalk Functionality Tests
- **MCC to Industry Mapping**: Tests mapping of MCC codes to industries
- **NAICS to Industry Mapping**: Tests mapping of NAICS codes to industries
- **SIC to Industry Mapping**: Tests mapping of SIC codes to industries
- **Crosswalk Validation**: Tests validation rules and consistency checks
- **Performance Testing**: Tests query performance with large datasets

### Business Risk Assessment Workflow Tests
- **High Risk Business Assessment**: Tests assessment of prohibited industries
- **Low Risk Business Assessment**: Tests assessment of legitimate businesses
- **Medium Risk Business Assessment**: Tests assessment of regulated industries
- **Performance Testing**: Tests assessment completion times
- **Error Handling**: Tests handling of invalid requests

### UI Integration Points Tests
- **Risk Display Data Format**: Tests data format compatibility with UI
- **Risk Level Color Mapping**: Tests risk level to color mapping
- **Risk Score Progress Bar**: Tests score conversion for progress bars

### Performance with Large Datasets Tests
- **Bulk Risk Keyword Detection**: Tests performance with large content
- **Bulk Crosswalk Queries**: Tests performance with multiple queries
- **Concurrent Risk Assessments**: Tests concurrent processing capabilities

## 📈 Performance Metrics

### Target Performance Requirements
- **Risk Keyword Detection**: < 1 second for 10KB content
- **Crosswalk Queries**: < 2 seconds for all mappings
- **Risk Assessment**: < 5 seconds per assessment
- **Concurrent Processing**: < 15 seconds for 10 concurrent assessments

### Actual Performance Results
*Performance results will be populated after test execution*

## 🎯 Success Criteria

### Technical Metrics
- [ ] All risk keyword detection tests pass
- [ ] All code crosswalk functionality tests pass  
- [ ] All business risk assessment workflow tests pass
- [ ] All UI integration point tests pass
- [ ] All performance tests meet requirements
- [ ] Database connectivity validated
- [ ] Error handling tests pass

### Quality Metrics
- [ ] Test coverage > 90%
- [ ] No critical test failures
- [ ] Performance requirements met
- [ ] Error handling validated
- [ ] UI compatibility confirmed

## 🚨 Issues and Recommendations

### Critical Issues
*Critical issues will be listed here after test execution*

### Performance Issues
*Performance issues will be listed here after test execution*

### Recommendations
*Recommendations will be generated based on test results*

## 📋 Next Steps

1. **Address Critical Issues**: Fix any critical test failures
2. **Performance Optimization**: Optimize any slow-performing components
3. **UI Integration**: Complete UI integration testing
4. **Documentation**: Update documentation based on test results
5. **Production Readiness**: Validate production readiness criteria

---

**Test Report Generated**: $(date)
**Test Environment**: $(uname -s) $(uname -r)
**Go Version**: $(go version 2>/dev/null || echo "Not available")

EOF
}

# Main execution
main() {
    echo "Starting Enhanced Classification System Testing..."
    echo ""
    
    # Initialize test report
    generate_report "$REPORT_FILE"
    
    # Check dependencies
    check_dependencies
    echo ""
    
    # Validate database connectivity
    validate_database
    echo ""
    
    # Track overall test results
    local total_tests=0
    local passed_tests=0
    local failed_tests=0
    
    # Run test categories
    local test_categories=(
        "TestRiskKeywordDetection:Risk Keyword Detection"
        "TestCodeCrosswalkFunctionality:Code Crosswalk Functionality"
        "TestBusinessRiskAssessmentWorkflow:Business Risk Assessment Workflow"
        "TestUIIntegrationPoints:UI Integration Points"
        "TestPerformanceWithLargeDatasets:Performance with Large Datasets"
    )
    
    for category in "${test_categories[@]}"; do
        IFS=':' read -r test_pattern description <<< "$category"
        total_tests=$((total_tests + 1))
        
        if run_test_category "$test_pattern" "$test_pattern" "$description"; then
            passed_tests=$((passed_tests + 1))
        else
            failed_tests=$((failed_tests + 1))
        fi
        echo ""
    done
    
    # Run performance benchmarks
    run_performance_benchmarks
    echo ""
    
    # Generate final report
    generate_detailed_report "$REPORT_FILE"
    
    # Print final summary
    print_section "Test Execution Summary"
    echo -e "📊 ${BLUE}Total Test Categories:${NC} $total_tests"
    echo -e "✅ ${GREEN}Passed:${NC} $passed_tests"
    echo -e "❌ ${RED}Failed:${NC} $failed_tests"
    echo -e "📈 ${BLUE}Success Rate:${NC} $(( (passed_tests * 100) / total_tests ))%"
    echo ""
    echo -e "📄 ${BLUE}Detailed Report:${NC} $REPORT_FILE"
    
    # Determine exit code
    if [ $failed_tests -eq 0 ]; then
        echo -e "\n🎉 ${GREEN}All tests passed! Enhanced classification system is ready.${NC}"
        exit 0
    else
        echo -e "\n⚠️  ${YELLOW}Some tests failed. Please review the report and fix issues.${NC}"
        exit 1
    fi
}

# Run main function
main "$@"
