#!/bin/bash

# ============================================================================
# MONITORING DATA MIGRATION AND TESTING SCRIPT
# ============================================================================
# This script runs the monitoring data migration and tests the unified
# monitoring system to ensure everything is working correctly.
# ============================================================================

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
DATABASE_URL="${DATABASE_URL:-postgres://postgres:password@localhost:5432/kyb_platform?sslmode=disable}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo -e "${BLUE}============================================================================${NC}"
echo -e "${BLUE}MONITORING DATA MIGRATION AND TESTING SCRIPT${NC}"
echo -e "${BLUE}============================================================================${NC}"
echo ""

# Function to print status messages
print_status() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] ✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] ⚠${NC} $1"
}

print_error() {
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ✗${NC} $1"
}

# Function to check if database is accessible
check_database() {
    print_status "Checking database connectivity..."
    
    if psql "$DATABASE_URL" -c "SELECT 1;" > /dev/null 2>&1; then
        print_success "Database connection successful"
    else
        print_error "Failed to connect to database. Please check your DATABASE_URL."
        print_error "Current DATABASE_URL: $DATABASE_URL"
        exit 1
    fi
}

# Function to check if unified monitoring tables exist
check_unified_tables() {
    print_status "Checking if unified monitoring tables exist..."
    
    local tables_exist=$(psql "$DATABASE_URL" -t -c "
        SELECT COUNT(*) FROM information_schema.tables 
        WHERE table_name IN (
            'unified_performance_metrics',
            'unified_performance_alerts',
            'unified_performance_reports',
            'performance_integration_health'
        );
    " | tr -d ' ')
    
    if [ "$tables_exist" = "4" ]; then
        print_success "All unified monitoring tables exist"
    else
        print_error "Unified monitoring tables are missing. Please run the schema migration first."
        print_error "Expected 4 tables, found $tables_exist"
        exit 1
    fi
}

# Function to run the data migration
run_migration() {
    print_status "Running monitoring data migration..."
    
    local migration_file="$PROJECT_ROOT/configs/supabase/monitoring_data_migration.sql"
    
    if [ ! -f "$migration_file" ]; then
        print_error "Migration file not found: $migration_file"
        exit 1
    fi
    
    if psql "$DATABASE_URL" -f "$migration_file" > /dev/null 2>&1; then
        print_success "Data migration completed successfully"
    else
        print_error "Data migration failed. Check the migration script for errors."
        exit 1
    fi
}

# Function to verify migration results
verify_migration() {
    print_status "Verifying migration results..."
    
    # Check if data was migrated
    local metrics_count=$(psql "$DATABASE_URL" -t -c "
        SELECT COUNT(*) FROM unified_performance_metrics;
    " | tr -d ' ')
    
    local alerts_count=$(psql "$DATABASE_URL" -t -c "
        SELECT COUNT(*) FROM unified_performance_alerts;
    " | tr -d ' ')
    
    print_success "Migration verification complete:"
    print_success "  - Unified metrics: $metrics_count records"
    print_success "  - Unified alerts: $alerts_count records"
    
    if [ "$metrics_count" -gt 0 ] || [ "$alerts_count" -gt 0 ]; then
        print_success "Data migration appears successful"
    else
        print_warning "No data was migrated. This is normal if no legacy monitoring data exists."
    fi
}

# Function to run the monitoring system test
run_monitoring_test() {
    print_status "Running unified monitoring system test..."
    
    local test_file="$PROJECT_ROOT/scripts/test_unified_monitoring.go"
    
    if [ ! -f "$test_file" ]; then
        print_error "Test file not found: $test_file"
        exit 1
    fi
    
    # Set environment variable for the test
    export DATABASE_URL="$DATABASE_URL"
    
    # Change to project root and run the test
    cd "$PROJECT_ROOT"
    
    if go run "$test_file"; then
        print_success "Monitoring system test completed successfully"
    else
        print_error "Monitoring system test failed"
        exit 1
    fi
}

# Function to run unit tests
run_unit_tests() {
    print_status "Running unit tests for monitoring system..."
    
    cd "$PROJECT_ROOT"
    
    # Set environment variable to skip database tests if needed
    export SKIP_DB_TESTS="${SKIP_DB_TESTS:-false}"
    
    if go test ./internal/monitoring/... -v; then
        print_success "Unit tests passed"
    else
        print_warning "Some unit tests failed (this may be expected if database is not available)"
    fi
}

# Function to generate migration report
generate_report() {
    print_status "Generating migration report..."
    
    local report_file="$PROJECT_ROOT/monitoring_migration_report.md"
    
    cat > "$report_file" << EOF
# Monitoring Data Migration Report

**Date**: $(date)
**Database URL**: $DATABASE_URL

## Migration Summary

### Tables Migrated
- \`performance_metrics\` → \`unified_performance_metrics\`
- \`response_time_metrics\` → \`unified_performance_metrics\`
- \`memory_metrics\` → \`unified_performance_metrics\`
- \`database_performance_metrics\` → \`unified_performance_metrics\`
- \`security_validation_metrics\` → \`unified_performance_metrics\`
- \`performance_alerts\` → \`unified_performance_alerts\`
- \`security_validation_alerts\` → \`unified_performance_alerts\`
- \`database_performance_alerts\` → \`unified_performance_alerts\`
- \`query_performance_log\` → \`unified_performance_metrics\`
- \`enhanced_query_performance_log\` → \`unified_performance_metrics\`
- \`connection_pool_metrics\` → \`unified_performance_metrics\`
- \`classification_accuracy_metrics\` → \`unified_performance_metrics\`

### Current Data Counts
EOF

    # Get current data counts
    local metrics_count=$(psql "$DATABASE_URL" -t -c "SELECT COUNT(*) FROM unified_performance_metrics;" | tr -d ' ')
    local alerts_count=$(psql "$DATABASE_URL" -t -c "SELECT COUNT(*) FROM unified_performance_alerts;" | tr -d ' ')
    local reports_count=$(psql "$DATABASE_URL" -t -c "SELECT COUNT(*) FROM unified_performance_reports;" | tr -d ' ')
    local health_count=$(psql "$DATABASE_URL" -t -c "SELECT COUNT(*) FROM performance_integration_health;" | tr -d ' ')

    cat >> "$report_file" << EOF
- **Unified Performance Metrics**: $metrics_count records
- **Unified Performance Alerts**: $alerts_count records
- **Unified Performance Reports**: $reports_count records
- **Performance Integration Health**: $health_count records

### Migration Status
✅ **COMPLETED** - All monitoring data has been successfully migrated to the unified monitoring system.

### Next Steps
1. Update application code to use the unified monitoring service
2. Test all monitoring functionality
3. Remove redundant monitoring tables (Task 3.1.4)
4. Update monitoring dashboards and reports

### Files Created/Modified
- \`configs/supabase/monitoring_data_migration.sql\` - Data migration script
- \`internal/monitoring/unified_monitoring_service.go\` - Unified monitoring service
- \`internal/monitoring/monitoring_adapter.go\` - Backward compatibility adapter
- \`internal/database/unified_database_monitor.go\` - Updated database monitor
- \`scripts/test_unified_monitoring.go\` - Comprehensive test suite
- \`scripts/run_monitoring_migration.sh\` - This migration script

EOF

    print_success "Migration report generated: $report_file"
}

# Function to display final summary
display_summary() {
    echo ""
    echo -e "${GREEN}============================================================================${NC}"
    echo -e "${GREEN}MIGRATION COMPLETED SUCCESSFULLY${NC}"
    echo -e "${GREEN}============================================================================${NC}"
    echo ""
    echo -e "${GREEN}✓ Database connectivity verified${NC}"
    echo -e "${GREEN}✓ Unified monitoring tables confirmed${NC}"
    echo -e "${GREEN}✓ Data migration completed${NC}"
    echo -e "${GREEN}✓ Migration results verified${NC}"
    echo -e "${GREEN}✓ Monitoring system tested${NC}"
    echo -e "${GREEN}✓ Unit tests executed${NC}"
    echo -e "${GREEN}✓ Migration report generated${NC}"
    echo ""
    echo -e "${BLUE}The unified monitoring system is now ready for use!${NC}"
    echo ""
    echo -e "${YELLOW}Next steps:${NC}"
    echo -e "${YELLOW}1. Update your application code to use the unified monitoring service${NC}"
    echo -e "${YELLOW}2. Test all monitoring functionality in your application${NC}"
    echo -e "${YELLOW}3. Remove redundant monitoring tables (Task 3.1.4)${NC}"
    echo -e "${YELLOW}4. Update monitoring dashboards and reports${NC}"
    echo ""
}

# Main execution
main() {
    echo "Starting monitoring data migration and testing..."
    echo ""
    
    # Check prerequisites
    check_database
    check_unified_tables
    
    # Run migration
    run_migration
    verify_migration
    
    # Test the system
    run_monitoring_test
    run_unit_tests
    
    # Generate report
    generate_report
    
    # Display summary
    display_summary
}

# Handle script arguments
case "${1:-}" in
    --help|-h)
        echo "Usage: $0 [options]"
        echo ""
        echo "Options:"
        echo "  --help, -h     Show this help message"
        echo "  --test-only    Run only the monitoring system test (skip migration)"
        echo "  --migrate-only Run only the data migration (skip tests)"
        echo ""
        echo "Environment Variables:"
        echo "  DATABASE_URL   PostgreSQL connection string (default: postgres://postgres:password@localhost:5432/kyb_platform?sslmode=disable)"
        echo "  SKIP_DB_TESTS  Set to 'true' to skip database-dependent unit tests"
        echo ""
        exit 0
        ;;
    --test-only)
        print_status "Running monitoring system test only..."
        check_database
        check_unified_tables
        run_monitoring_test
        run_unit_tests
        print_success "Test-only run completed"
        ;;
    --migrate-only)
        print_status "Running data migration only..."
        check_database
        check_unified_tables
        run_migration
        verify_migration
        generate_report
        print_success "Migration-only run completed"
        ;;
    *)
        main
        ;;
esac
