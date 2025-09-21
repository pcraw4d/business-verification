#!/bin/bash

# Enhanced Classification System Test Execution Script
# This script executes comprehensive tests for subtask 1.5.4

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
TEST_DIR="$PROJECT_ROOT/test"
REPORT_DIR="$PROJECT_ROOT/test_reports"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
REPORT_FILE="$REPORT_DIR/enhanced_classification_test_report_${TIMESTAMP}.md"
LOG_FILE="$REPORT_DIR/test_execution_${TIMESTAMP}.log"

# Create report directory
mkdir -p "$REPORT_DIR"

# Function to log messages
log_message() {
    local level="$1"
    local message="$2"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    echo "[$timestamp] [$level] $message" | tee -a "$LOG_FILE"
}

# Function to print section headers
print_section() {
    echo -e "\n${CYAN}📋 $1${NC}"
    echo -e "${CYAN}$(printf '=%.0s' {1..60})${NC}"
    log_message "INFO" "Starting section: $1"
}

# Function to print test results
print_test_result() {
    local test_name="$1"
    local status="$2"
    local duration="$3"
    local details="$4"
    
    if [ "$status" = "PASS" ]; then
        echo -e "✅ ${GREEN}$test_name${NC} - ${GREEN}PASSED${NC} (${duration})"
        log_message "PASS" "$test_name - PASSED (${duration})"
    else
        echo -e "❌ ${RED}$test_name${NC} - ${RED}FAILED${NC} (${duration})"
        log_message "FAIL" "$test_name - FAILED (${duration})"
        if [ -n "$details" ]; then
            echo -e "   ${YELLOW}Details: $details${NC}"
            log_message "ERROR" "Details: $details"
        fi
    fi
}

# Function to check prerequisites
check_prerequisites() {
    print_section "Prerequisites Check"
    
    local missing_deps=()
    
    # Check Go
    if ! command -v go &> /dev/null; then
        missing_deps+=("go")
        echo -e "❌ ${RED}Go not found${NC}"
    else
        echo -e "✅ ${GREEN}Go $(go version | cut -d' ' -f3)${NC}"
        log_message "INFO" "Go version: $(go version)"
    fi
    
    # Check required Go packages
    local required_packages=(
        "github.com/stretchr/testify"
        "github.com/lib/pq"
    )
    
    for package in "${required_packages[@]}"; do
        if go list -m "$package" &> /dev/null; then
            echo -e "✅ ${GREEN}$package${NC}"
            log_message "INFO" "Package $package is available"
        else
            echo -e "❌ ${RED}$package (missing)${NC}"
            missing_deps+=("$package")
            log_message "WARN" "Package $package is missing"
        fi
    done
    
    # Check database connectivity
    if [ -z "$DATABASE_URL" ]; then
        echo -e "⚠️  ${YELLOW}DATABASE_URL not set. Some tests may be skipped.${NC}"
        log_message "WARN" "DATABASE_URL not set"
    else
        echo -e "✅ ${GREEN}DATABASE_URL is set${NC}"
        log_message "INFO" "DATABASE_URL is configured"
    fi
    
    # Install missing dependencies
    if [ ${#missing_deps[@]} -gt 0 ]; then
        echo -e "\n${YELLOW}Installing missing dependencies...${NC}"
        for dep in "${missing_deps[@]}"; do
            if [[ "$dep" == go ]]; then
                echo -e "${RED}Please install Go from https://golang.org/dl/${NC}"
                log_message "ERROR" "Go installation required"
                exit 1
            else
                echo -e "Installing $dep..."
                go get "$dep"
                log_message "INFO" "Installed $dep"
            fi
        done
    fi
    
    echo -e "\n${GREEN}Prerequisites check completed${NC}"
}

# Function to setup test data
setup_test_data() {
    print_section "Test Data Setup"
    
    if [ -z "$DATABASE_URL" ]; then
        echo -e "⚠️  ${YELLOW}Skipping test data setup - DATABASE_URL not set${NC}"
        log_message "WARN" "Skipping test data setup"
        return 0
    fi
    
    echo "Setting up test data in database..."
    log_message "INFO" "Setting up test data"
    
    # Run test data setup script
    if psql "$DATABASE_URL" -f "$SCRIPT_DIR/setup_test_data.sql" > "$REPORT_DIR/test_data_setup.log" 2>&1; then
        echo -e "✅ ${GREEN}Test data setup completed successfully${NC}"
        log_message "INFO" "Test data setup completed"
    else
        echo -e "❌ ${RED}Test data setup failed${NC}"
        log_message "ERROR" "Test data setup failed"
        echo "Check $REPORT_DIR/test_data_setup.log for details"
        return 1
    fi
}

# Function to run individual test categories
run_test_category() {
    local category="$1"
    local test_pattern="$2"
    local description="$3"
    
    print_section "Testing $description"
    
    local start_time=$(date +%s)
    local test_output_file="$REPORT_DIR/${category}_test_output.log"
    
    # Run tests with verbose output
    if go test -v -run "$test_pattern" ./test/ > "$test_output_file" 2>&1; then
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        print_test_result "$description" "PASS" "${duration}s"
        
        # Extract test details
        local test_count=$(grep -c "=== RUN" "$test_output_file" || echo "0")
        local pass_count=$(grep -c "--- PASS:" "$test_output_file" || echo "0")
        local fail_count=$(grep -c "--- FAIL:" "$test_output_file" || echo "0")
        
        echo -e "   📊 Tests: $test_count, Passed: $pass_count, Failed: $fail_count"
        log_message "INFO" "$description: $test_count tests, $pass_count passed, $fail_count failed"
        
        return 0
    else
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        
        # Extract error details
        local error_details=$(tail -5 "$test_output_file" | tr '\n' ' ')
        print_test_result "$description" "FAIL" "${duration}s" "$error_details"
        
        # Log detailed error information
        log_message "ERROR" "$description failed: $error_details"
        
        return 1
    fi
}

# Function to run performance benchmarks
run_performance_benchmarks() {
    print_section "Performance Benchmarks"
    
    local benchmark_output_file="$REPORT_DIR/performance_benchmarks.log"
    
    echo "Running performance benchmarks..."
    log_message "INFO" "Starting performance benchmarks"
    
    # Run benchmarks
    if go test -bench=. -benchmem ./test/ > "$benchmark_output_file" 2>&1; then
        echo -e "✅ ${GREEN}Performance benchmarks completed${NC}"
        log_message "INFO" "Performance benchmarks completed"
        
        # Extract benchmark results
        if [ -f "$benchmark_output_file" ]; then
            echo -e "\n${BLUE}Benchmark Results:${NC}"
            grep -E "Benchmark|PASS|FAIL" "$benchmark_output_file" | head -20
        fi
    else
        echo -e "❌ ${RED}Performance benchmarks failed${NC}"
        log_message "ERROR" "Performance benchmarks failed"
    fi
}

# Function to run test coverage analysis
run_test_coverage() {
    print_section "Test Coverage Analysis"
    
    local coverage_output_file="$REPORT_DIR/coverage_report.html"
    local coverage_log_file="$REPORT_DIR/coverage.log"
    
    echo "Running test coverage analysis..."
    log_message "INFO" "Starting test coverage analysis"
    
    # Run coverage analysis
    if go test -coverprofile="$REPORT_DIR/coverage.out" -covermode=atomic ./test/ > "$coverage_log_file" 2>&1; then
        # Generate HTML coverage report
        go tool cover -html="$REPORT_DIR/coverage.out" -o "$coverage_output_file"
        
        # Extract coverage percentage
        local coverage_percent=$(go tool cover -func="$REPORT_DIR/coverage.out" | tail -1 | awk '{print $3}')
        
        echo -e "✅ ${GREEN}Test coverage analysis completed${NC}"
        echo -e "   📊 Coverage: $coverage_percent"
        log_message "INFO" "Test coverage: $coverage_percent"
        
        # Check if coverage meets requirements
        local coverage_num=$(echo "$coverage_percent" | sed 's/%//')
        if (( $(echo "$coverage_num >= 80" | bc -l) )); then
            echo -e "   ✅ ${GREEN}Coverage meets requirements (>= 80%)${NC}"
        else
            echo -e "   ⚠️  ${YELLOW}Coverage below requirements (< 80%)${NC}"
            log_message "WARN" "Test coverage below requirements: $coverage_percent"
        fi
    else
        echo -e "❌ ${RED}Test coverage analysis failed${NC}"
        log_message "ERROR" "Test coverage analysis failed"
    fi
}

# Function to generate comprehensive test report
generate_test_report() {
    print_section "Generating Test Report"
    
    log_message "INFO" "Generating comprehensive test report"
    
    cat > "$REPORT_FILE" << EOF
# Enhanced Classification System Test Report

**Generated**: $(date)
**Test Suite**: Enhanced Classification System (Subtask 1.5.4)
**Environment**: $(uname -s) $(uname -r)
**Go Version**: $(go version 2>/dev/null || echo "Not available")

## 🎯 Test Objectives

This test suite validates the enhanced classification system implementation including:

1. **Risk Keyword Detection** - Comprehensive testing of risk keyword matching, synonym detection, and pattern recognition
2. **Code Crosswalk Functionality** - Validation of MCC/NAICS/SIC code mapping and crosswalk validation
3. **Business Risk Assessment Workflow** - End-to-end testing of the complete risk assessment process
4. **UI Integration Points** - Verification of data formats and display compatibility
5. **Performance Testing** - Large dataset and concurrent processing validation

## 📊 Test Execution Summary

**Test Execution Time**: $(date)
**Total Test Categories**: 5
**Test Environment**: $(uname -s) $(uname -r)
**Database**: $(if [ -n "$DATABASE_URL" ]; then echo "Connected"; else echo "Not configured"; fi)

## 🔍 Test Results

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
- [ ] Test coverage > 80%
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

## 📁 Test Artifacts

- **Test Log**: $LOG_FILE
- **Coverage Report**: $REPORT_DIR/coverage_report.html
- **Performance Benchmarks**: $REPORT_DIR/performance_benchmarks.log
- **Test Data Setup Log**: $REPORT_DIR/test_data_setup.log

---

**Test Report Generated**: $(date)
**Test Environment**: $(uname -s) $(uname -r)
**Go Version**: $(go version 2>/dev/null || echo "Not available")

EOF

    echo -e "✅ ${GREEN}Test report generated: $REPORT_FILE${NC}"
    log_message "INFO" "Test report generated: $REPORT_FILE"
}

# Function to cleanup test data
cleanup_test_data() {
    print_section "Test Data Cleanup"
    
    if [ -z "$DATABASE_URL" ]; then
        echo -e "⚠️  ${YELLOW}Skipping test data cleanup - DATABASE_URL not set${NC}"
        log_message "WARN" "Skipping test data cleanup"
        return 0
    fi
    
    echo "Cleaning up test data..."
    log_message "INFO" "Cleaning up test data"
    
    # Run cleanup function
    if psql "$DATABASE_URL" -c "SELECT cleanup_test_data();" > "$REPORT_DIR/test_data_cleanup.log" 2>&1; then
        echo -e "✅ ${GREEN}Test data cleanup completed${NC}"
        log_message "INFO" "Test data cleanup completed"
    else
        echo -e "❌ ${RED}Test data cleanup failed${NC}"
        log_message "ERROR" "Test data cleanup failed"
    fi
}

# Main execution function
main() {
    echo -e "${PURPLE}🧪 Enhanced Classification System Test Execution${NC}"
    echo -e "${PURPLE}================================================${NC}"
    echo ""
    
    log_message "INFO" "Starting Enhanced Classification System Test Execution"
    
    # Track overall test results
    local total_tests=0
    local passed_tests=0
    local failed_tests=0
    
    # Check prerequisites
    check_prerequisites
    echo ""
    
    # Setup test data
    setup_test_data
    echo ""
    
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
    
    # Run test coverage analysis
    run_test_coverage
    echo ""
    
    # Generate comprehensive test report
    generate_test_report
    echo ""
    
    # Cleanup test data
    cleanup_test_data
    echo ""
    
    # Print final summary
    print_section "Test Execution Summary"
    echo -e "📊 ${BLUE}Total Test Categories:${NC} $total_tests"
    echo -e "✅ ${GREEN}Passed:${NC} $passed_tests"
    echo -e "❌ ${RED}Failed:${NC} $failed_tests"
    echo -e "📈 ${BLUE}Success Rate:${NC} $(( (passed_tests * 100) / total_tests ))%"
    echo ""
    echo -e "📄 ${BLUE}Detailed Report:${NC} $REPORT_FILE"
    echo -e "📝 ${BLUE}Execution Log:${NC} $LOG_FILE"
    echo -e "📊 ${BLUE}Coverage Report:${NC} $REPORT_DIR/coverage_report.html"
    
    log_message "INFO" "Test execution completed: $passed_tests/$total_tests passed"
    
    # Determine exit code
    if [ $failed_tests -eq 0 ]; then
        echo -e "\n🎉 ${GREEN}All tests passed! Enhanced classification system is ready.${NC}"
        log_message "INFO" "All tests passed - Enhanced classification system is ready"
        exit 0
    else
        echo -e "\n⚠️  ${YELLOW}Some tests failed. Please review the report and fix issues.${NC}"
        log_message "WARN" "Some tests failed - Review required"
        exit 1
    fi
}

# Run main function
main "$@"
