#!/bin/bash

# Orphaned Records Detection Script
# This script runs comprehensive orphaned records detection tests

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
OUTPUT_DIR="$PROJECT_ROOT/test-results"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Function to print colored output
print_status() {
    local color=$1
    local message=$2
    echo -e "${color}${message}${NC}"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to validate database connection
validate_db_connection() {
    local db_url=$1
    print_status $BLUE "🔍 Validating database connection..."
    
    if command_exists psql; then
        if psql "$db_url" -c "SELECT 1;" >/dev/null 2>&1; then
            print_status $GREEN "✅ Database connection successful"
            return 0
        else
            print_status $RED "❌ Database connection failed"
            return 1
        fi
    else
        print_status $YELLOW "⚠️  psql not found, skipping connection validation"
        return 0
    fi
}

# Function to run SQL tests
run_sql_tests() {
    local db_url=$1
    local output_file="$OUTPUT_DIR/orphaned_records_sql_test_$TIMESTAMP.log"
    
    print_status $BLUE "📊 Running SQL-based orphaned records detection tests..."
    
    if command_exists psql; then
        print_status $BLUE "   Executing SQL test script..."
        if psql "$db_url" -f "$SCRIPT_DIR/check-orphaned-records.sql" > "$output_file" 2>&1; then
            print_status $GREEN "✅ SQL tests completed successfully"
            print_status $BLUE "   Results saved to: $output_file"
            
            # Check for any orphaned records in the output
            if grep -q "orphaned.*[1-9]" "$output_file"; then
                print_status $YELLOW "⚠️  Some orphaned records detected - check results"
            fi
            
            return 0
        else
            print_status $RED "❌ SQL tests failed"
            print_status $RED "   Check error log: $output_file"
            return 1
        fi
    else
        print_status $YELLOW "⚠️  psql not found, skipping SQL tests"
        return 0
    fi
}

# Function to run Go tests
run_go_tests() {
    local db_url=$1
    local output_file="$OUTPUT_DIR/orphaned_records_go_test_$TIMESTAMP.log"
    
    print_status $BLUE "🔧 Running Go-based orphaned records detection tests..."
    
    if command_exists go; then
        print_status $BLUE "   Compiling and running Go test program..."
        
        # Change to script directory
        cd "$SCRIPT_DIR"
        
        # Compile the Go program
        if go build -o check-orphaned-records check-orphaned-records.go; then
            print_status $GREEN "✅ Go program compiled successfully"
            
            # Run the test
            if ./check-orphaned-records "$db_url" > "$output_file" 2>&1; then
                print_status $GREEN "✅ Go tests completed successfully"
                print_status $BLUE "   Results saved to: $output_file"
                return 0
            else
                print_status $RED "❌ Go tests failed"
                print_status $RED "   Check error log: $output_file"
                return 1
            fi
        else
            print_status $RED "❌ Failed to compile Go program"
            return 1
        fi
    else
        print_status $YELLOW "⚠️  Go not found, skipping Go tests"
        return 0
    fi
}

# Function to generate summary report
generate_summary_report() {
    local output_file="$OUTPUT_DIR/orphaned_records_test_summary_$TIMESTAMP.md"
    
    print_status $BLUE "📋 Generating summary report..."
    
    cat > "$output_file" << EOF
# Orphaned Records Detection Test Summary

**Test Date:** $(date)
**Timestamp:** $TIMESTAMP

## Test Results

### SQL Tests
EOF

    if [ -f "$OUTPUT_DIR/orphaned_records_sql_test_$TIMESTAMP.log" ]; then
        echo "✅ SQL tests completed" >> "$output_file"
        echo "📁 Results: \`orphaned_records_sql_test_$TIMESTAMP.log\`" >> "$output_file"
    else
        echo "⚠️ SQL tests skipped (psql not available)" >> "$output_file"
    fi

    cat >> "$output_file" << EOF

### Go Tests
EOF

    if [ -f "$OUTPUT_DIR/orphaned_records_go_test_$TIMESTAMP.log" ]; then
        echo "✅ Go tests completed" >> "$output_file"
        echo "📁 Results: \`orphaned_records_go_test_$TIMESTAMP.log\`" >> "$output_file"
    else
        echo "⚠️ Go tests skipped (Go not available)" >> "$output_file"
    fi

    cat >> "$output_file" << EOF

## What Was Tested

### Foreign Key Relationships
- **merchants.user_id → users.id**: Merchants referencing non-existent users
- **business_verifications.merchant_id → merchants.id**: Verifications referencing non-existent merchants
- **classification_results.merchant_id → merchants.id**: Classification results referencing non-existent merchants
- **risk_assessments.merchant_id → merchants.id**: Risk assessments referencing non-existent merchants
- **audit_logs.user_id → users.id**: Audit logs referencing non-existent users

### Logical Business Relationships
- **business_verifications.user_id → users.id**: Verifications referencing non-existent users
- **merchant_audit_logs.merchant_id → merchants.id**: Merchant audit logs referencing non-existent merchants
- **industry_keywords.industry_id → industries.id**: Industry keywords referencing non-existent industries
- **business_risk_assessments.business_id → merchants.id**: Risk assessments referencing non-existent merchants
- **business_risk_assessments.risk_keyword_id → risk_keywords.id**: Risk assessments referencing non-existent risk keywords

### Detection Methods
- **LEFT JOIN Analysis**: Uses LEFT JOIN to identify records with NULL parent references
- **Comprehensive Coverage**: Checks both foreign key constraints and logical business relationships
- **Sample Data**: Provides sample orphaned record values for debugging
- **Impact Analysis**: Calculates percentage of orphaned records per relationship

## Common Orphaned Record Scenarios

### 1. Data Migration Issues
- Records created before proper foreign key constraints were in place
- Data imported from external systems without proper validation
- Manual data entry errors

### 2. Cascade Delete Issues
- Parent records deleted without proper cascade handling
- Application-level deletions that bypass database constraints
- Bulk operations that don't maintain referential integrity

### 3. Application Bugs
- Race conditions in concurrent operations
- Transaction rollback issues
- API endpoints that don't validate relationships

## Cleanup Recommendations

### Safe Cleanup Approaches
1. **Identify Root Cause**: Determine why orphaned records exist
2. **Backup Data**: Always backup before cleanup operations
3. **Gradual Cleanup**: Process orphaned records in batches
4. **Validate Fixes**: Re-run tests after cleanup

### Cleanup Queries (Use with Caution)
\`\`\`sql
-- Example: Remove merchants with invalid user references
DELETE FROM merchants 
WHERE user_id NOT IN (SELECT id FROM users);

-- Example: Remove verifications with invalid merchant references
DELETE FROM business_verifications 
WHERE merchant_id NOT IN (SELECT id FROM merchants);
\`\`\`

## Next Steps

1. Review the test results in the log files
2. Identify the root cause of any orphaned records
3. Implement proper cleanup procedures
4. Add application-level validation to prevent future orphaned records
5. Re-run tests to verify cleanup

## Files Generated

- \`orphaned_records_sql_test_$TIMESTAMP.log\` - SQL test results
- \`orphaned_records_go_test_$TIMESTAMP.log\` - Go test results
- \`orphaned_records_test_summary_$TIMESTAMP.md\` - This summary report

EOF

    print_status $GREEN "✅ Summary report generated: $output_file"
}

# Function to display usage
usage() {
    echo "Usage: $0 <database_url>"
    echo ""
    echo "Arguments:"
    echo "  database_url    PostgreSQL connection string"
    echo ""
    echo "Examples:"
    echo "  $0 'postgresql://user:pass@localhost:5432/dbname'"
    echo "  $0 'postgresql://user:pass@localhost:5432/dbname?sslmode=require'"
    echo ""
    echo "Environment Variables:"
    echo "  DATABASE_URL    Alternative way to specify database URL"
}

# Main execution
main() {
    print_status $BLUE "🚀 Starting Orphaned Records Detection Testing"
    print_status $BLUE "=============================================="
    
    # Get database URL
    local db_url=""
    if [ $# -eq 1 ]; then
        db_url="$1"
    elif [ -n "$DATABASE_URL" ]; then
        db_url="$DATABASE_URL"
    else
        print_status $RED "❌ Database URL required"
        usage
        exit 1
    fi
    
    print_status $BLUE "📊 Database URL: ${db_url%%:*}:****"
    print_status $BLUE "📁 Output directory: $OUTPUT_DIR"
    print_status $BLUE "🕐 Timestamp: $TIMESTAMP"
    echo ""
    
    # Validate database connection
    if ! validate_db_connection "$db_url"; then
        exit 1
    fi
    echo ""
    
    # Run tests
    local sql_result=0
    local go_result=0
    
    run_sql_tests "$db_url" || sql_result=1
    echo ""
    
    run_go_tests "$db_url" || go_result=1
    echo ""
    
    # Generate summary
    generate_summary_report
    echo ""
    
    # Final status
    if [ $sql_result -eq 0 ] && [ $go_result -eq 0 ]; then
        print_status $GREEN "🎉 All orphaned records detection tests completed successfully!"
        print_status $BLUE "📁 Check the test results in: $OUTPUT_DIR"
        exit 0
    else
        print_status $YELLOW "⚠️  Some tests failed - check the logs for details"
        print_status $BLUE "📁 Check the test results in: $OUTPUT_DIR"
        exit 1
    fi
}

# Run main function
main "$@"
