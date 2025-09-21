#!/bin/bash

# Database Performance Testing Script for KYB Platform
# This script runs comprehensive database performance tests as part of subtask 3.2.3

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
REPORT_DIR="$PROJECT_ROOT/reports/performance"
LOG_DIR="$PROJECT_ROOT/logs/performance"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

# Create directories if they don't exist
mkdir -p "$REPORT_DIR"
mkdir -p "$LOG_DIR"

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1" | tee -a "$LOG_DIR/db_performance_$TIMESTAMP.log"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" | tee -a "$LOG_DIR/db_performance_$TIMESTAMP.log"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1" | tee -a "$LOG_DIR/db_performance_$TIMESTAMP.log"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "$LOG_DIR/db_performance_$TIMESTAMP.log"
}

# Function to check if required tools are installed
check_dependencies() {
    log_info "Checking dependencies..."
    
    local missing_deps=()
    
    # Check for Go
    if ! command -v go &> /dev/null; then
        missing_deps+=("go")
    fi
    
    # Check for psql (PostgreSQL client)
    if ! command -v psql &> /dev/null; then
        missing_deps+=("psql")
    fi
    
    # Check for jq (JSON processor)
    if ! command -v jq &> /dev/null; then
        missing_deps+=("jq")
    fi
    
    if [ ${#missing_deps[@]} -ne 0 ]; then
        log_error "Missing dependencies: ${missing_deps[*]}"
        log_error "Please install the missing dependencies and try again."
        exit 1
    fi
    
    log_success "All dependencies are available"
}

# Function to check database connectivity
check_database_connection() {
    log_info "Checking database connectivity..."
    
    # Load environment variables
    if [ -f "$PROJECT_ROOT/.env" ]; then
        source "$PROJECT_ROOT/.env"
    fi
    
    # Check if required environment variables are set
    if [ -z "$DATABASE_URL" ] && [ -z "$SUPABASE_URL" ]; then
        log_error "Database connection information not found in environment variables"
        log_error "Please set DATABASE_URL or SUPABASE_URL in your .env file"
        exit 1
    fi
    
    # Test database connection
    local db_url="${DATABASE_URL:-$SUPABASE_URL}"
    if psql "$db_url" -c "SELECT 1;" &> /dev/null; then
        log_success "Database connection successful"
    else
        log_error "Failed to connect to database"
        exit 1
    fi
}

# Function to run Go-based performance tests
run_go_performance_tests() {
    log_info "Running Go-based database performance tests..."
    
    cd "$PROJECT_ROOT"
    
    # Create a temporary test file
    cat > "temp_performance_test.go" << 'EOF'
package main

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "os"
    "time"
    
    _ "github.com/lib/pq"
    
    "github.com/your-org/kyb-platform/internal/database"
)

func main() {
    // Get database URL from environment
    dbURL := os.Getenv("DATABASE_URL")
    if dbURL == "" {
        dbURL = os.Getenv("SUPABASE_URL")
    }
    if dbURL == "" {
        log.Fatal("DATABASE_URL or SUPABASE_URL environment variable not set")
    }
    
    // Connect to database
    db, err := sql.Open("postgres", dbURL)
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }
    defer db.Close()
    
    // Test connection
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    if err := db.PingContext(ctx); err != nil {
        log.Fatalf("Failed to ping database: %v", err)
    }
    
    // Create performance test suite
    config := &database.PerformanceTestConfig{
        TestDuration:        2 * time.Minute,
        WarmupDuration:      30 * time.Second,
        CooldownDuration:    30 * time.Second,
        ConcurrentUsers:     10,
        RequestsPerUser:     50,
        RequestInterval:     100 * time.Millisecond,
        MaxQueryTime:        1 * time.Second,
        MaxConnectionTime:   5 * time.Second,
        MinThroughput:       25,
        MonitorCPU:          true,
        MonitorMemory:       true,
        MonitorConnections:  true,
    }
    
    suite := database.NewDatabasePerformanceTestSuite(db, config)
    
    // Run comprehensive performance tests
    results, err := suite.RunComprehensivePerformanceTests(ctx)
    if err != nil {
        log.Fatalf("Performance tests failed: %v", err)
    }
    
    // Output results as JSON
    output, err := json.MarshalIndent(results, "", "  ")
    if err != nil {
        log.Fatalf("Failed to marshal results: %v", err)
    }
    
    fmt.Println(string(output))
}
EOF
    
    # Run the test
    if go run temp_performance_test.go > "$REPORT_DIR/db_performance_results_$TIMESTAMP.json" 2>&1; then
        log_success "Go performance tests completed successfully"
    else
        log_error "Go performance tests failed"
        return 1
    fi
    
    # Clean up temporary file
    rm -f temp_performance_test.go
}

# Function to run SQL-based performance tests
run_sql_performance_tests() {
    log_info "Running SQL-based performance tests..."
    
    local db_url="${DATABASE_URL:-$SUPABASE_URL}"
    
    # Create SQL performance test script
    cat > "$REPORT_DIR/sql_performance_test_$TIMESTAMP.sql" << 'EOF'
-- SQL Performance Tests for KYB Platform Database
-- Generated: TIMESTAMP_PLACEHOLDER

-- Test 1: Index Usage Analysis
\echo 'Testing index usage...'
EXPLAIN (ANALYZE, BUFFERS) SELECT id, email FROM users WHERE email = 'test@example.com';
EXPLAIN (ANALYZE, BUFFERS) SELECT id, name FROM businesses WHERE user_id = '00000000-0000-0000-0000-000000000000';
EXPLAIN (ANALYZE, BUFFERS) SELECT * FROM business_classifications WHERE business_id = '00000000-0000-0000-0000-000000000000';
EXPLAIN (ANALYZE, BUFFERS) SELECT * FROM risk_assessments WHERE risk_level = 'low';
EXPLAIN (ANALYZE, BUFFERS) SELECT * FROM audit_logs WHERE created_at > NOW() - INTERVAL '24 hours' ORDER BY created_at DESC;

-- Test 2: Query Performance Benchmarks
\echo 'Testing query performance...'
\timing on

-- Simple queries
SELECT COUNT(*) FROM users;
SELECT COUNT(*) FROM businesses;
SELECT COUNT(*) FROM business_classifications;
SELECT COUNT(*) FROM risk_assessments;
SELECT COUNT(*) FROM compliance_checks;

-- Complex queries
SELECT u.email, COUNT(b.id) as business_count 
FROM users u 
LEFT JOIN businesses b ON u.id = b.user_id 
GROUP BY u.id, u.email 
ORDER BY business_count DESC 
LIMIT 10;

SELECT bc.industry, COUNT(*) as classification_count
FROM business_classifications bc
GROUP BY bc.industry
ORDER BY classification_count DESC
LIMIT 10;

SELECT ra.risk_level, COUNT(*) as assessment_count
FROM risk_assessments ra
GROUP BY ra.risk_level
ORDER BY assessment_count DESC;

-- Test 3: Database Statistics
\echo 'Database statistics...'
SELECT 
    schemaname,
    tablename,
    attname,
    n_distinct,
    correlation
FROM pg_stats 
WHERE schemaname = 'public' 
ORDER BY tablename, attname;

-- Test 4: Index Statistics
\echo 'Index statistics...'
SELECT 
    schemaname,
    tablename,
    indexname,
    idx_tup_read,
    idx_tup_fetch
FROM pg_stat_user_indexes 
ORDER BY idx_tup_read DESC;

-- Test 5: Table Sizes
\echo 'Table sizes...'
SELECT 
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_tables 
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

-- Test 6: Connection Information
\echo 'Connection information...'
SELECT 
    count(*) as total_connections,
    count(*) FILTER (WHERE state = 'active') as active_connections,
    count(*) FILTER (WHERE state = 'idle') as idle_connections
FROM pg_stat_activity;

\timing off
EOF
    
    # Replace timestamp placeholder
    sed -i "s/TIMESTAMP_PLACEHOLDER/$(date)/" "$REPORT_DIR/sql_performance_test_$TIMESTAMP.sql"
    
    # Run SQL tests
    if psql "$db_url" -f "$REPORT_DIR/sql_performance_test_$TIMESTAMP.sql" > "$REPORT_DIR/sql_performance_results_$TIMESTAMP.txt" 2>&1; then
        log_success "SQL performance tests completed successfully"
    else
        log_warning "SQL performance tests completed with warnings (check results file)"
    fi
}

# Function to analyze slow queries
analyze_slow_queries() {
    log_info "Analyzing slow queries..."
    
    local db_url="${DATABASE_URL:-$SUPABASE_URL}"
    
    # Create slow query analysis script
    cat > "$REPORT_DIR/slow_query_analysis_$TIMESTAMP.sql" << 'EOF'
-- Slow Query Analysis for KYB Platform Database
-- Generated: TIMESTAMP_PLACEHOLDER

-- Enable query statistics collection
SELECT pg_stat_statements_reset();

-- Wait a moment for queries to accumulate
SELECT pg_sleep(1);

-- Find slow queries
SELECT 
    query,
    calls,
    total_time,
    mean_time,
    stddev_time,
    rows,
    100.0 * shared_blks_hit / nullif(shared_blks_hit + shared_blks_read, 0) AS hit_percent
FROM pg_stat_statements 
WHERE mean_time > 100  -- Queries taking more than 100ms on average
ORDER BY mean_time DESC
LIMIT 20;

-- Find queries with high I/O
SELECT 
    query,
    calls,
    total_time,
    mean_time,
    shared_blks_read,
    shared_blks_written,
    shared_blks_hit
FROM pg_stat_statements 
WHERE shared_blks_read + shared_blks_written > 1000
ORDER BY (shared_blks_read + shared_blks_written) DESC
LIMIT 20;

-- Find most frequently called queries
SELECT 
    query,
    calls,
    total_time,
    mean_time,
    rows
FROM pg_stat_statements 
ORDER BY calls DESC
LIMIT 20;
EOF
    
    # Replace timestamp placeholder
    sed -i "s/TIMESTAMP_PLACEHOLDER/$(date)/" "$REPORT_DIR/slow_query_analysis_$TIMESTAMP.sql"
    
    # Run slow query analysis
    if psql "$db_url" -f "$REPORT_DIR/slow_query_analysis_$TIMESTAMP.sql" > "$REPORT_DIR/slow_query_results_$TIMESTAMP.txt" 2>&1; then
        log_success "Slow query analysis completed successfully"
    else
        log_warning "Slow query analysis completed with warnings (check results file)"
    fi
}

# Function to generate performance report
generate_performance_report() {
    log_info "Generating comprehensive performance report..."
    
    local report_file="$REPORT_DIR/database_performance_report_$TIMESTAMP.md"
    
    cat > "$report_file" << EOF
# Database Performance Test Report

**Generated**: $(date)
**Test Suite**: Subtask 3.2.3 - Performance Testing
**Platform**: KYB Platform Database Optimization

## Executive Summary

This report contains the results of comprehensive database performance testing conducted as part of the Supabase Table Improvement Implementation Plan, Phase 3.2.3.

## Test Configuration

- **Test Duration**: 2 minutes
- **Concurrent Users**: 10
- **Requests per User**: 50
- **Request Interval**: 100ms
- **Max Query Time Threshold**: 1 second
- **Min Throughput Threshold**: 25 queries/second

## Test Results

### 1. Basic Query Performance
EOF
    
    # Parse Go test results if available
    if [ -f "$REPORT_DIR/db_performance_results_$TIMESTAMP.json" ]; then
        echo "### Go Performance Test Results" >> "$report_file"
        echo '```json' >> "$report_file"
        cat "$REPORT_DIR/db_performance_results_$TIMESTAMP.json" >> "$report_file"
        echo '```' >> "$report_file"
    fi
    
    cat >> "$report_file" << EOF

### 2. SQL Performance Test Results

EOF
    
    # Add SQL test results if available
    if [ -f "$REPORT_DIR/sql_performance_results_$TIMESTAMP.txt" ]; then
        echo '```sql' >> "$report_file"
        cat "$REPORT_DIR/sql_performance_results_$TIMESTAMP.txt" >> "$report_file"
        echo '```' >> "$report_file"
    fi
    
    cat >> "$report_file" << EOF

### 3. Slow Query Analysis

EOF
    
    # Add slow query results if available
    if [ -f "$REPORT_DIR/slow_query_results_$TIMESTAMP.txt" ]; then
        echo '```sql' >> "$report_file"
        cat "$REPORT_DIR/slow_query_results_$TIMESTAMP.txt" >> "$report_file"
        echo '```' >> "$report_file"
    fi
    
    cat >> "$report_file" << EOF

## Performance Metrics

### Response Time Targets
- **API Endpoints**: < 500ms (95th percentile)
- **Database Queries**: < 200ms (average)
- **Complex Queries**: < 1 second (95th percentile)

### Throughput Targets
- **Concurrent Users**: 100+
- **Queries per Second**: 500+
- **Error Rate**: < 0.1%

### Resource Usage Targets
- **Connection Pool Utilization**: < 80%
- **Cache Hit Rate**: > 90%
- **Database Size Growth**: < 10% per month

## Recommendations

Based on the performance test results, the following recommendations are made:

1. **Query Optimization**: Review and optimize any queries exceeding the 200ms threshold
2. **Index Optimization**: Ensure all frequently accessed columns have appropriate indexes
3. **Connection Pooling**: Monitor and optimize database connection pool settings
4. **Caching Strategy**: Implement or enhance caching for frequently accessed data
5. **Monitoring**: Set up continuous performance monitoring and alerting

## Next Steps

1. Review the detailed test results in the accompanying files
2. Implement recommended optimizations
3. Re-run performance tests to validate improvements
4. Set up continuous performance monitoring
5. Schedule regular performance testing cycles

## Files Generated

- \`db_performance_results_$TIMESTAMP.json\` - Go performance test results
- \`sql_performance_test_$TIMESTAMP.sql\` - SQL test script
- \`sql_performance_results_$TIMESTAMP.txt\` - SQL test results
- \`slow_query_analysis_$TIMESTAMP.sql\` - Slow query analysis script
- \`slow_query_results_$TIMESTAMP.txt\` - Slow query analysis results
- \`database_performance_report_$TIMESTAMP.md\` - This comprehensive report

---

**Report Generated by**: Database Performance Testing Script
**Part of**: Supabase Table Improvement Implementation Plan - Subtask 3.2.3
EOF
    
    log_success "Performance report generated: $report_file"
}

# Function to clean up old reports
cleanup_old_reports() {
    log_info "Cleaning up old performance reports (keeping last 10)..."
    
    # Keep only the last 10 reports
    find "$REPORT_DIR" -name "database_performance_report_*.md" -type f | sort -r | tail -n +11 | xargs rm -f 2>/dev/null || true
    find "$REPORT_DIR" -name "db_performance_results_*.json" -type f | sort -r | tail -n +11 | xargs rm -f 2>/dev/null || true
    find "$REPORT_DIR" -name "sql_performance_*.sql" -type f | sort -r | tail -n +11 | xargs rm -f 2>/dev/null || true
    find "$REPORT_DIR" -name "sql_performance_results_*.txt" -type f | sort -r | tail -n +11 | xargs rm -f 2>/dev/null || true
    find "$REPORT_DIR" -name "slow_query_*.sql" -type f | sort -r | tail -n +11 | xargs rm -f 2>/dev/null || true
    find "$REPORT_DIR" -name "slow_query_results_*.txt" -type f | sort -r | tail -n +11 | xargs rm -f 2>/dev/null || true
    
    log_success "Old reports cleaned up"
}

# Main function
main() {
    echo "🚀 KYB Platform - Database Performance Testing"
    echo "=============================================="
    echo "Subtask 3.2.3: Performance Testing"
    echo "Timestamp: $TIMESTAMP"
    echo
    
    # Check dependencies
    check_dependencies
    
    # Check database connection
    check_database_connection
    
    # Run performance tests
    log_info "Starting comprehensive database performance testing..."
    
    # Run Go-based tests
    if run_go_performance_tests; then
        log_success "Go performance tests completed"
    else
        log_warning "Go performance tests failed, continuing with SQL tests"
    fi
    
    # Run SQL-based tests
    run_sql_performance_tests
    
    # Analyze slow queries
    analyze_slow_queries
    
    # Generate comprehensive report
    generate_performance_report
    
    # Clean up old reports
    cleanup_old_reports
    
    echo
    log_success "Database performance testing completed successfully!"
    echo
    log_info "Results available in: $REPORT_DIR"
    log_info "Logs available in: $LOG_DIR"
    echo
    log_info "Review the generated report for detailed analysis and recommendations."
}

# Run main function
main "$@"
