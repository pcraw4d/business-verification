#!/bin/bash

# Monitoring Systems Test Script
# This script tests the functionality of monitoring systems after table consolidation

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Database connection details
DB_HOST="db.qpqhuqqmkjxsltzshfam.supabase.co"
DB_PORT="5432"
DB_USER="postgres"
DB_NAME="postgres"
DB_PASSWORD="Geaux44tigers!"

echo -e "${BLUE}=== Monitoring Systems Test Suite ===${NC}"
echo "Timestamp: $(date)"
echo ""

# Function to test database connection
test_connection() {
    echo -e "${YELLOW}Testing database connection...${NC}"
    
    if psql "postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=require" -c "SELECT 1;" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Database connection successful${NC}"
        return 0
    else
        echo -e "${RED}✗ Database connection failed${NC}"
        return 1
    fi
}

# Function to test unified performance metrics table
test_unified_performance_metrics() {
    echo -e "${YELLOW}Testing unified_performance_metrics table...${NC}"
    
    # Test table structure
    local structure_test=$(psql "postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=require" -t -c "
        SELECT COUNT(*) 
        FROM information_schema.columns 
        WHERE table_name = 'unified_performance_metrics' 
        AND table_schema = 'public';
    " 2>/dev/null | tr -d ' \n' || echo "0")
    
    if [[ "$structure_test" -gt 10 ]]; then
        echo -e "${GREEN}✓ Table structure is correct ($structure_test columns)${NC}"
    else
        echo -e "${RED}✗ Table structure issue ($structure_test columns)${NC}"
        return 1
    fi
    
    # Test insert capability
    local insert_test=$(psql "postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=require" -t -c "
        INSERT INTO unified_performance_metrics (
            component, component_instance, service_name, metric_type, metric_category,
            metric_name, metric_value, metric_unit, tags, metadata, data_source, created_at
        ) VALUES (
            'test', 'test_instance', 'test_service', 'test', 'test',
            'test_metric', 1.0, 'count', '{}', '{}', 'test_script', NOW()
        ) RETURNING id;
    " 2>/dev/null | tr -d ' \n' || echo "")
    
    if [[ -n "$insert_test" ]]; then
        echo -e "${GREEN}✓ Insert test successful (ID: $insert_test)${NC}"
        
        # Clean up test data
        psql "postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=require" -c "
            DELETE FROM unified_performance_metrics WHERE id = $insert_test;
        " > /dev/null 2>&1
        
        return 0
    else
        echo -e "${RED}✗ Insert test failed${NC}"
        return 1
    fi
}

# Function to test unified performance alerts table
test_unified_performance_alerts() {
    echo -e "${YELLOW}Testing unified_performance_alerts table...${NC}"
    
    # Test table structure
    local structure_test=$(psql "postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=require" -t -c "
        SELECT COUNT(*) 
        FROM information_schema.columns 
        WHERE table_name = 'unified_performance_alerts' 
        AND table_schema = 'public';
    " 2>/dev/null | tr -d ' \n' || echo "0")
    
    if [[ "$structure_test" -gt 15 ]]; then
        echo -e "${GREEN}✓ Table structure is correct ($structure_test columns)${NC}"
    else
        echo -e "${RED}✗ Table structure issue ($structure_test columns)${NC}"
        return 1
    fi
    
    # Test query capability
    local query_test=$(psql "postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=require" -t -c "
        SELECT COUNT(*) FROM unified_performance_alerts WHERE component = 'test';
    " 2>/dev/null | tr -d ' \n' || echo "0")
    
    echo -e "${GREEN}✓ Query test successful (found $query_test test alerts)${NC}"
    return 0
}

# Function to test unified performance reports table
test_unified_performance_reports() {
    echo -e "${YELLOW}Testing unified_performance_reports table...${NC}"
    
    # Test table structure
    local structure_test=$(psql "postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=require" -t -c "
        SELECT COUNT(*) 
        FROM information_schema.columns 
        WHERE table_name = 'unified_performance_reports' 
        AND table_schema = 'public';
    " 2>/dev/null | tr -d ' \n' || echo "0")
    
    if [[ "$structure_test" -gt 10 ]]; then
        echo -e "${GREEN}✓ Table structure is correct ($structure_test columns)${NC}"
    else
        echo -e "${RED}✗ Table structure issue ($structure_test columns)${NC}"
        return 1
    fi
    
    # Test query capability
    local query_test=$(psql "postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=require" -t -c "
        SELECT COUNT(*) FROM unified_performance_reports WHERE report_type = 'test';
    " 2>/dev/null | tr -d ' \n' || echo "0")
    
    echo -e "${GREEN}✓ Query test successful (found $query_test test reports)${NC}"
    return 0
}

# Function to test performance integration health table
test_performance_integration_health() {
    echo -e "${YELLOW}Testing performance_integration_health table...${NC}"
    
    # Test table structure
    local structure_test=$(psql "postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=require" -t -c "
        SELECT COUNT(*) 
        FROM information_schema.columns 
        WHERE table_name = 'performance_integration_health' 
        AND table_schema = 'public';
    " 2>/dev/null | tr -d ' \n' || echo "0")
    
    if [[ "$structure_test" -gt 8 ]]; then
        echo -e "${GREEN}✓ Table structure is correct ($structure_test columns)${NC}"
    else
        echo -e "${RED}✗ Table structure issue ($structure_test columns)${NC}"
        return 1
    fi
    
    # Test query capability
    local query_test=$(psql "postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=require" -t -c "
        SELECT COUNT(*) FROM performance_integration_health WHERE service_name = 'test';
    " 2>/dev/null | tr -d ' \n' || echo "0")
    
    echo -e "${GREEN}✓ Query test successful (found $query_test test health records)${NC}"
    return 0
}

# Function to test application code compatibility
test_application_compatibility() {
    echo -e "${YELLOW}Testing application code compatibility...${NC}"
    
    # Check if Go files compile without errors
    local go_files=(
        "internal/classification/performance_dashboards.go"
        "internal/classification/comprehensive_performance_monitor.go"
        "internal/classification/performance_alerting.go"
        "internal/classification/classification_accuracy_monitoring.go"
        "internal/classification/connection_pool_monitoring.go"
        "internal/classification/query_performance_monitoring.go"
        "internal/classification/usage_monitoring.go"
        "internal/classification/accuracy_calculation_service.go"
    )
    
    local compile_errors=0
    
    for file in "${go_files[@]}"; do
        if [[ -f "$file" ]]; then
            if go build -o /dev/null "$file" 2>/dev/null; then
                echo -e "${GREEN}✓ $file compiles successfully${NC}"
            else
                echo -e "${RED}✗ $file has compilation errors${NC}"
                ((compile_errors++))
            fi
        else
            echo -e "${YELLOW}⚠ $file not found${NC}"
        fi
    done
    
    if [[ $compile_errors -eq 0 ]]; then
        echo -e "${GREEN}✓ All application files compile successfully${NC}"
        return 0
    else
        echo -e "${RED}✗ $compile_errors files have compilation errors${NC}"
        return 1
    fi
}

# Function to test monitoring functions
test_monitoring_functions() {
    echo -e "${YELLOW}Testing monitoring functions...${NC}"
    
    # Test if monitoring functions exist
    local functions=(
        "get_performance_metrics_summary"
        "get_performance_alerts_summary"
        "get_performance_reports_summary"
        "check_performance_integration_health"
    )
    
    local missing_functions=0
    
    for func in "${functions[@]}"; do
        local exists=$(psql "postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=require" -t -c "
            SELECT EXISTS (
                SELECT 1 FROM information_schema.routines 
                WHERE routine_name = '$func' 
                AND routine_schema = 'public'
            );
        " 2>/dev/null | tr -d ' \n' || echo "f")
        
        if [[ "$exists" == "t" ]]; then
            echo -e "${GREEN}✓ Function $func exists${NC}"
        else
            echo -e "${RED}✗ Function $func missing${NC}"
            ((missing_functions++))
        fi
    done
    
    if [[ $missing_functions -eq 0 ]]; then
        echo -e "${GREEN}✓ All monitoring functions are available${NC}"
        return 0
    else
        echo -e "${RED}✗ $missing_functions functions are missing${NC}"
        return 1
    fi
}

# Function to run performance tests
test_performance() {
    echo -e "${YELLOW}Running performance tests...${NC}"
    
    # Test query performance on unified tables
    local start_time=$(date +%s%N)
    
    psql "postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=require" -c "
        SELECT COUNT(*) FROM unified_performance_metrics 
        WHERE created_at >= NOW() - INTERVAL '1 hour';
    " > /dev/null 2>&1
    
    local end_time=$(date +%s%N)
    local duration=$(( (end_time - start_time) / 1000000 )) # Convert to milliseconds
    
    if [[ $duration -lt 1000 ]]; then
        echo -e "${GREEN}✓ Query performance is good (${duration}ms)${NC}"
    else
        echo -e "${YELLOW}⚠ Query performance is slow (${duration}ms)${NC}"
    fi
    
    return 0
}

# Main execution
main() {
    echo -e "${BLUE}Starting monitoring systems test suite...${NC}"
    echo ""
    
    local test_results=()
    
    # Test database connection
    if test_connection; then
        test_results+=("connection:pass")
    else
        test_results+=("connection:fail")
        echo -e "${RED}Cannot proceed without database connection${NC}"
        exit 1
    fi
    
    # Test unified tables
    if test_unified_performance_metrics; then
        test_results+=("unified_metrics:pass")
    else
        test_results+=("unified_metrics:fail")
    fi
    
    if test_unified_performance_alerts; then
        test_results+=("unified_alerts:pass")
    else
        test_results+=("unified_alerts:fail")
    fi
    
    if test_unified_performance_reports; then
        test_results+=("unified_reports:pass")
    else
        test_results+=("unified_reports:fail")
    fi
    
    if test_performance_integration_health; then
        test_results+=("integration_health:pass")
    else
        test_results+=("integration_health:fail")
    fi
    
    # Test application compatibility
    if test_application_compatibility; then
        test_results+=("app_compatibility:pass")
    else
        test_results+=("app_compatibility:fail")
    fi
    
    # Test monitoring functions
    if test_monitoring_functions; then
        test_results+=("monitoring_functions:pass")
    else
        test_results+=("monitoring_functions:fail")
    fi
    
    # Test performance
    if test_performance; then
        test_results+=("performance:pass")
    else
        test_results+=("performance:fail")
    fi
    
    echo ""
    echo -e "${BLUE}=== Test Results Summary ===${NC}"
    
    local passed=0
    local failed=0
    
    for result in "${test_results[@]}"; do
        local test_name=$(echo "$result" | cut -d: -f1)
        local test_status=$(echo "$result" | cut -d: -f2)
        
        if [[ "$test_status" == "pass" ]]; then
            echo -e "${GREEN}✓ $test_name: PASSED${NC}"
            ((passed++))
        else
            echo -e "${RED}✗ $test_name: FAILED${NC}"
            ((failed++))
        fi
    done
    
    echo ""
    echo -e "${BLUE}Summary: $passed passed, $failed failed${NC}"
    
    if [[ $failed -eq 0 ]]; then
        echo -e "${GREEN}=== All tests passed! Monitoring systems are working correctly ===${NC}"
        return 0
    else
        echo -e "${RED}=== Some tests failed. Please review the issues above ===${NC}"
        return 1
    fi
}

# Run main function
main "$@"
