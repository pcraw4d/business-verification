#!/bin/bash

# Database Migration Execution Script
# This script executes the removal of redundant monitoring tables

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

echo -e "${BLUE}=== Database Migration: Remove Redundant Monitoring Tables ===${NC}"
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

# Function to check if unified tables exist
check_unified_tables() {
    echo -e "${YELLOW}Checking if unified monitoring tables exist...${NC}"
    
    local query="
    SELECT table_name 
    FROM information_schema.tables 
    WHERE table_schema = 'public' 
    AND table_name IN (
        'unified_performance_metrics',
        'unified_performance_alerts', 
        'unified_performance_reports',
        'performance_integration_health'
    )
    ORDER BY table_name;
    "
    
    local result=$(psql "postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=require" -t -c "$query" 2>/dev/null || echo "")
    
    if [[ -n "$result" ]]; then
        echo -e "${GREEN}✓ Unified monitoring tables found:${NC}"
        echo "$result" | sed 's/^/  /'
        return 0
    else
        echo -e "${RED}✗ Unified monitoring tables not found${NC}"
        return 1
    fi
}

# Function to check redundant tables
check_redundant_tables() {
    echo -e "${YELLOW}Checking for redundant monitoring tables...${NC}"
    
    local redundant_tables=(
        "performance_metrics"
        "performance_alerts" 
        "performance_reports"
        "database_performance_metrics"
        "query_performance_logs"
        "connection_pool_metrics"
        "classification_accuracy_metrics"
        "usage_monitoring_data"
        "performance_dashboard_data"
        "monitoring_health_checks"
        "performance_optimization_logs"
        "system_performance_metrics"
        "application_performance_data"
        "monitoring_alert_history"
        "performance_trend_analysis"
        "monitoring_system_status"
    )
    
    local existing_tables=""
    
    for table in "${redundant_tables[@]}"; do
        local exists=$(psql "postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=require" -t -c "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = '$table');" 2>/dev/null | tr -d ' \n' || echo "f")
        
        if [[ "$exists" == "t" ]]; then
            existing_tables="$existing_tables $table"
        fi
    done
    
    if [[ -n "$existing_tables" ]]; then
        echo -e "${YELLOW}Found redundant tables to remove:${NC}"
        echo "$existing_tables" | tr ' ' '\n' | sed 's/^/  /'
        return 0
    else
        echo -e "${GREEN}✓ No redundant monitoring tables found${NC}"
        return 1
    fi
}

# Function to create backup
create_backup() {
    echo -e "${YELLOW}Creating backup of existing tables...${NC}"
    
    local backup_file="backup_monitoring_tables_$(date +%Y%m%d_%H%M%S).sql"
    
    if pg_dump "postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=require" \
        --schema-only \
        --table="performance_metrics" \
        --table="performance_alerts" \
        --table="performance_reports" \
        --table="database_performance_metrics" \
        --table="query_performance_logs" \
        --table="connection_pool_metrics" \
        --table="classification_accuracy_metrics" \
        --table="usage_monitoring_data" \
        --table="performance_dashboard_data" \
        --table="monitoring_health_checks" \
        --table="performance_optimization_logs" \
        --table="system_performance_metrics" \
        --table="application_performance_data" \
        --table="monitoring_alert_history" \
        --table="performance_trend_analysis" \
        --table="monitoring_system_status" \
        > "backups/$backup_file" 2>/dev/null; then
        
        echo -e "${GREEN}✓ Backup created: backups/$backup_file${NC}"
        return 0
    else
        echo -e "${YELLOW}⚠ Backup creation failed (tables may not exist)${NC}"
        return 1
    fi
}

# Function to execute migration
execute_migration() {
    echo -e "${YELLOW}Executing table removal migration...${NC}"
    
    if psql "postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=require" \
        -f "configs/supabase/remove_redundant_monitoring_tables.sql"; then
        
        echo -e "${GREEN}✓ Migration executed successfully${NC}"
        return 0
    else
        echo -e "${RED}✗ Migration failed${NC}"
        return 1
    fi
}

# Function to verify migration
verify_migration() {
    echo -e "${YELLOW}Verifying migration results...${NC}"
    
    # Check that redundant tables are gone
    local redundant_count=$(psql "postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=require" -t -c "
        SELECT COUNT(*) 
        FROM information_schema.tables 
        WHERE table_schema = 'public' 
        AND table_name IN (
            'performance_metrics', 'performance_alerts', 'performance_reports',
            'database_performance_metrics', 'query_performance_logs', 'connection_pool_metrics',
            'classification_accuracy_metrics', 'usage_monitoring_data', 'performance_dashboard_data',
            'monitoring_health_checks', 'performance_optimization_logs', 'system_performance_metrics',
            'application_performance_data', 'monitoring_alert_history', 'performance_trend_analysis',
            'monitoring_system_status'
        );
    " 2>/dev/null | tr -d ' \n' || echo "0")
    
    # Check that unified tables still exist
    local unified_count=$(psql "postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=require" -t -c "
        SELECT COUNT(*) 
        FROM information_schema.tables 
        WHERE table_schema = 'public' 
        AND table_name IN (
            'unified_performance_metrics', 'unified_performance_alerts', 
            'unified_performance_reports', 'performance_integration_health'
        );
    " 2>/dev/null | tr -d ' \n' || echo "0")
    
    echo "Redundant tables remaining: $redundant_count"
    echo "Unified tables present: $unified_count"
    
    if [[ "$redundant_count" == "0" && "$unified_count" == "4" ]]; then
        echo -e "${GREEN}✓ Migration verification successful${NC}"
        return 0
    else
        echo -e "${RED}✗ Migration verification failed${NC}"
        return 1
    fi
}

# Main execution
main() {
    echo -e "${BLUE}Starting database migration process...${NC}"
    echo ""
    
    # Create backups directory
    mkdir -p backups
    
    # Test connection
    if ! test_connection; then
        echo -e "${RED}Cannot proceed without database connection${NC}"
        exit 1
    fi
    
    # Check unified tables
    if ! check_unified_tables; then
        echo -e "${RED}Cannot proceed without unified monitoring tables${NC}"
        exit 1
    fi
    
    # Check redundant tables
    if ! check_redundant_tables; then
        echo -e "${GREEN}No redundant tables to remove. Migration not needed.${NC}"
        exit 0
    fi
    
    # Create backup
    create_backup
    
    # Execute migration
    if ! execute_migration; then
        echo -e "${RED}Migration failed. Check logs for details.${NC}"
        exit 1
    fi
    
    # Verify migration
    if ! verify_migration; then
        echo -e "${RED}Migration verification failed. Check database state.${NC}"
        exit 1
    fi
    
    echo ""
    echo -e "${GREEN}=== Migration completed successfully ===${NC}"
    echo "Timestamp: $(date)"
    echo ""
    echo -e "${BLUE}Next steps:${NC}"
    echo "1. Test monitoring systems functionality"
    echo "2. Validate performance improvements"
    echo "3. Update any remaining application references"
}

# Run main function
main "$@"
