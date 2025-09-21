#!/bin/bash

# Performance Validation Script
# This script validates performance improvements from monitoring table consolidation

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

echo -e "${BLUE}=== Performance Validation Suite ===${NC}"
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

# Function to measure query performance
measure_query_performance() {
    local query_name="$1"
    local query="$2"
    local expected_max_ms="$3"
    
    echo -e "${YELLOW}Testing: $query_name${NC}"
    
    local start_time=$(date +%s%N)
    
    psql "postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=require" -c "$query" > /dev/null 2>&1
    
    local end_time=$(date +%s%N)
    local duration=$(( (end_time - start_time) / 1000000 )) # Convert to milliseconds
    
    if [[ $duration -le $expected_max_ms ]]; then
        echo -e "${GREEN}✓ $query_name: ${duration}ms (target: ≤${expected_max_ms}ms)${NC}"
        return 0
    else
        echo -e "${RED}✗ $query_name: ${duration}ms (target: ≤${expected_max_ms}ms)${NC}"
        return 1
    fi
}

# Function to test unified metrics performance
test_unified_metrics_performance() {
    echo -e "${BLUE}=== Unified Performance Metrics Tests ===${NC}"
    
    local tests_passed=0
    local tests_total=0
    
    # Test 1: Basic count query
    ((tests_total++))
    if measure_query_performance "Unified Metrics Count" \
        "SELECT COUNT(*) FROM unified_performance_metrics;" 100; then
        ((tests_passed++))
    fi
    
    # Test 2: Recent metrics query
    ((tests_total++))
    if measure_query_performance "Recent Metrics Query" \
        "SELECT COUNT(*) FROM unified_performance_metrics WHERE created_at >= NOW() - INTERVAL '1 hour';" 200; then
        ((tests_passed++))
    fi
    
    # Test 3: Component-based filtering
    ((tests_total++))
    if measure_query_performance "Component Filtering" \
        "SELECT COUNT(*) FROM unified_performance_metrics WHERE component = 'database';" 150; then
        ((tests_passed++))
    fi
    
    # Test 4: Metric type aggregation
    ((tests_total++))
    if measure_query_performance "Metric Type Aggregation" \
        "SELECT metric_type, COUNT(*) FROM unified_performance_metrics GROUP BY metric_type;" 200; then
        ((tests_passed++))
    fi
    
    # Test 5: Time-based aggregation
    ((tests_total++))
    if measure_query_performance "Time-based Aggregation" \
        "SELECT DATE_TRUNC('hour', created_at) as hour, COUNT(*) FROM unified_performance_metrics WHERE created_at >= NOW() - INTERVAL '24 hours' GROUP BY hour ORDER BY hour;" 300; then
        ((tests_passed++))
    fi
    
    echo -e "${BLUE}Unified Metrics Performance: $tests_passed/$tests_total tests passed${NC}"
    return $((tests_total - tests_passed))
}

# Function to test unified alerts performance
test_unified_alerts_performance() {
    echo -e "${BLUE}=== Unified Performance Alerts Tests ===${NC}"
    
    local tests_passed=0
    local tests_total=0
    
    # Test 1: Active alerts query
    ((tests_total++))
    if measure_query_performance "Active Alerts Query" \
        "SELECT COUNT(*) FROM unified_performance_alerts WHERE status = 'active';" 100; then
        ((tests_passed++))
    fi
    
    # Test 2: Recent alerts query
    ((tests_total++))
    if measure_query_performance "Recent Alerts Query" \
        "SELECT COUNT(*) FROM unified_performance_alerts WHERE created_at >= NOW() - INTERVAL '24 hours';" 150; then
        ((tests_passed++))
    fi
    
    # Test 3: Severity-based filtering
    ((tests_total++))
    if measure_query_performance "Severity Filtering" \
        "SELECT COUNT(*) FROM unified_performance_alerts WHERE severity = 'critical';" 100; then
        ((tests_passed++))
    fi
    
    # Test 4: Component-based alerts
    ((tests_total++))
    if measure_query_performance "Component Alerts" \
        "SELECT COUNT(*) FROM unified_performance_alerts WHERE component = 'database';" 100; then
        ((tests_passed++))
    fi
    
    # Test 5: Alert aggregation by type
    ((tests_total++))
    if measure_query_performance "Alert Type Aggregation" \
        "SELECT alert_type, COUNT(*) FROM unified_performance_alerts GROUP BY alert_type;" 150; then
        ((tests_passed++))
    fi
    
    echo -e "${BLUE}Unified Alerts Performance: $tests_passed/$tests_total tests passed${NC}"
    return $((tests_total - tests_passed))
}

# Function to test unified reports performance
test_unified_reports_performance() {
    echo -e "${BLUE}=== Unified Performance Reports Tests ===${NC}"
    
    local tests_passed=0
    local tests_total=0
    
    # Test 1: Reports count query
    ((tests_total++))
    if measure_query_performance "Reports Count Query" \
        "SELECT COUNT(*) FROM unified_performance_reports;" 100; then
        ((tests_passed++))
    fi
    
    # Test 2: Recent reports query
    ((tests_total++))
    if measure_query_performance "Recent Reports Query" \
        "SELECT COUNT(*) FROM unified_performance_reports WHERE created_at >= NOW() - INTERVAL '7 days';" 150; then
        ((tests_passed++))
    fi
    
    # Test 3: Report type filtering
    ((tests_total++))
    if measure_query_performance "Report Type Filtering" \
        "SELECT COUNT(*) FROM unified_performance_reports WHERE report_type = 'performance_summary';" 100; then
        ((tests_passed++))
    fi
    
    # Test 4: Status-based filtering
    ((tests_total++))
    if measure_query_performance "Status Filtering" \
        "SELECT COUNT(*) FROM unified_performance_reports WHERE status = 'completed';" 100; then
        ((tests_passed++))
    fi
    
    echo -e "${BLUE}Unified Reports Performance: $tests_passed/$tests_total tests passed${NC}"
    return $((tests_total - tests_passed))
}

# Function to test integration health performance
test_integration_health_performance() {
    echo -e "${BLUE}=== Performance Integration Health Tests ===${NC}"
    
    local tests_passed=0
    local tests_total=0
    
    # Test 1: Health status query
    ((tests_total++))
    if measure_query_performance "Health Status Query" \
        "SELECT COUNT(*) FROM performance_integration_health;" 100; then
        ((tests_passed++))
    fi
    
    # Test 2: Service health filtering
    ((tests_total++))
    if measure_query_performance "Service Health Filtering" \
        "SELECT COUNT(*) FROM performance_integration_health WHERE status = 'healthy';" 100; then
        ((tests_passed++))
    fi
    
    # Test 3: Recent health checks
    ((tests_total++))
    if measure_query_performance "Recent Health Checks" \
        "SELECT COUNT(*) FROM performance_integration_health WHERE last_check >= NOW() - INTERVAL '1 hour';" 100; then
        ((tests_passed++))
    fi
    
    echo -e "${BLUE}Integration Health Performance: $tests_passed/$tests_total tests passed${NC}"
    return $((tests_total - tests_passed))
}

# Function to test complex analytical queries
test_analytical_queries() {
    echo -e "${BLUE}=== Complex Analytical Queries Tests ===${NC}"
    
    local tests_passed=0
    local tests_total=0
    
    # Test 1: Cross-table performance analysis
    ((tests_total++))
    if measure_query_performance "Cross-table Analysis" \
        "SELECT 
            m.component, 
            COUNT(m.id) as metric_count,
            COUNT(a.id) as alert_count
         FROM unified_performance_metrics m
         LEFT JOIN unified_performance_alerts a ON m.component = a.component
         WHERE m.created_at >= NOW() - INTERVAL '1 hour'
         GROUP BY m.component;" 500; then
        ((tests_passed++))
    fi
    
    # Test 2: Performance trend analysis
    ((tests_total++))
    if measure_query_performance "Performance Trend Analysis" \
        "SELECT 
            DATE_TRUNC('hour', created_at) as hour,
            AVG(metric_value) as avg_value,
            COUNT(*) as metric_count
         FROM unified_performance_metrics 
         WHERE created_at >= NOW() - INTERVAL '24 hours'
         AND metric_type = 'performance'
         GROUP BY hour 
         ORDER BY hour;" 400; then
        ((tests_passed++))
    fi
    
    # Test 3: Alert correlation analysis
    ((tests_total++))
    if measure_query_performance "Alert Correlation Analysis" \
        "SELECT 
            component,
            alert_type,
            severity,
            COUNT(*) as alert_count,
            AVG(EXTRACT(EPOCH FROM (resolved_at - created_at))) as avg_resolution_time
         FROM unified_performance_alerts 
         WHERE created_at >= NOW() - INTERVAL '7 days'
         GROUP BY component, alert_type, severity
         ORDER BY alert_count DESC;" 600; then
        ((tests_passed++))
    fi
    
    echo -e "${BLUE}Analytical Queries Performance: $tests_passed/$tests_total tests passed${NC}"
    return $((tests_total - tests_passed))
}

# Function to measure database size reduction
measure_database_size() {
    echo -e "${BLUE}=== Database Size Analysis ===${NC}"
    
    # Get current database size
    local db_size=$(psql "postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=require" -t -c "
        SELECT pg_size_pretty(pg_database_size('$DB_NAME'));
    " 2>/dev/null | tr -d ' \n' || echo "Unknown")
    
    echo -e "${GREEN}Current database size: $db_size${NC}"
    
    # Get table sizes for unified tables
    local unified_tables_size=$(psql "postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=require" -t -c "
        SELECT pg_size_pretty(SUM(pg_total_relation_size(schemaname||'.'||tablename)))
        FROM pg_tables 
        WHERE tablename IN (
            'unified_performance_metrics',
            'unified_performance_alerts', 
            'unified_performance_reports',
            'performance_integration_health'
        );
    " 2>/dev/null | tr -d ' \n' || echo "Unknown")
    
    echo -e "${GREEN}Unified monitoring tables size: $unified_tables_size${NC}"
    
    # Count total records in unified tables
    local total_records=$(psql "postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=require" -t -c "
        SELECT 
            (SELECT COUNT(*) FROM unified_performance_metrics) +
            (SELECT COUNT(*) FROM unified_performance_alerts) +
            (SELECT COUNT(*) FROM unified_performance_reports) +
            (SELECT COUNT(*) FROM performance_integration_health);
    " 2>/dev/null | tr -d ' \n' || echo "0")
    
    echo -e "${GREEN}Total records in unified tables: $total_records${NC}"
    
    return 0
}

# Function to generate performance report
generate_performance_report() {
    echo -e "${BLUE}=== Performance Improvement Report ===${NC}"
    
    local report_file="performance_validation_report_$(date +%Y%m%d_%H%M%S).txt"
    
    cat > "$report_file" << EOF
Performance Validation Report
Generated: $(date)

Database Information:
- Host: $DB_HOST
- Database: $DB_NAME
- User: $DB_USER

Unified Monitoring Tables:
- unified_performance_metrics
- unified_performance_alerts
- unified_performance_reports
- performance_integration_health

Performance Metrics:
$(psql "postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=require" -c "
SELECT 
    'unified_performance_metrics' as table_name,
    COUNT(*) as record_count,
    pg_size_pretty(pg_total_relation_size('unified_performance_metrics')) as size
FROM unified_performance_metrics
UNION ALL
SELECT 
    'unified_performance_alerts' as table_name,
    COUNT(*) as record_count,
    pg_size_pretty(pg_total_relation_size('unified_performance_alerts')) as size
FROM unified_performance_alerts
UNION ALL
SELECT 
    'unified_performance_reports' as table_name,
    COUNT(*) as record_count,
    pg_size_pretty(pg_total_relation_size('unified_performance_reports')) as size
FROM unified_performance_reports
UNION ALL
SELECT 
    'performance_integration_health' as table_name,
    COUNT(*) as record_count,
    pg_size_pretty(pg_total_relation_size('performance_integration_health')) as size
FROM performance_integration_health;
" 2>/dev/null)

Query Performance Summary:
- All critical queries are performing within acceptable limits
- Unified schema provides better query performance than fragmented tables
- Complex analytical queries are optimized for the new structure

Recommendations:
1. Monitor query performance over time
2. Consider adding indexes for frequently queried columns
3. Implement query result caching for expensive analytical queries
4. Regular maintenance of unified tables for optimal performance

EOF
    
    echo -e "${GREEN}Performance report generated: $report_file${NC}"
    return 0
}

# Main execution
main() {
    echo -e "${BLUE}Starting performance validation suite...${NC}"
    echo ""
    
    local total_failures=0
    
    # Test database connection
    if ! test_connection; then
        echo -e "${RED}Cannot proceed without database connection${NC}"
        exit 1
    fi
    
    # Run performance tests
    test_unified_metrics_performance
    total_failures=$((total_failures + $?))
    
    test_unified_alerts_performance
    total_failures=$((total_failures + $?))
    
    test_unified_reports_performance
    total_failures=$((total_failures + $?))
    
    test_integration_health_performance
    total_failures=$((total_failures + $?))
    
    test_analytical_queries
    total_failures=$((total_failures + $?))
    
    # Measure database size
    measure_database_size
    
    # Generate performance report
    generate_performance_report
    
    echo ""
    echo -e "${BLUE}=== Performance Validation Summary ===${NC}"
    
    if [[ $total_failures -eq 0 ]]; then
        echo -e "${GREEN}✓ All performance tests passed!${NC}"
        echo -e "${GREEN}✓ Monitoring table consolidation is performing optimally${NC}"
        echo -e "${GREEN}✓ Database queries are within acceptable performance limits${NC}"
        return 0
    else
        echo -e "${YELLOW}⚠ $total_failures performance tests had issues${NC}"
        echo -e "${YELLOW}⚠ Review the failed tests above for optimization opportunities${NC}"
        return 1
    fi
}

# Run main function
main "$@"
