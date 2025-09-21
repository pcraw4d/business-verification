#!/bin/bash

# Data Consistency Verification Script
# This script runs comprehensive data consistency verification tests

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
    local output_file="$OUTPUT_DIR/data_consistency_sql_test_$TIMESTAMP.log"
    
    print_status $BLUE "📊 Running SQL-based data consistency verification tests..."
    
    if command_exists psql; then
        print_status $BLUE "   Executing SQL test script..."
        if psql "$db_url" -f "$SCRIPT_DIR/verify-data-consistency.sql" > "$output_file" 2>&1; then
            print_status $GREEN "✅ SQL tests completed successfully"
            print_status $BLUE "   Results saved to: $output_file"
            
            # Check for any consistency issues in the output
            if grep -q "inconsistent_count.*[1-9]" "$output_file"; then
                print_status $YELLOW "⚠️  Some data consistency issues detected - check results"
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
    local output_file="$OUTPUT_DIR/data_consistency_go_test_$TIMESTAMP.log"
    
    print_status $BLUE "🔧 Running Go-based data consistency verification tests..."
    
    if command_exists go; then
        print_status $BLUE "   Compiling and running Go test program..."
        
        # Change to script directory
        cd "$SCRIPT_DIR"
        
        # Compile the Go program
        if go build -o verify-data-consistency verify-data-consistency.go; then
            print_status $GREEN "✅ Go program compiled successfully"
            
            # Run the test
            if ./verify-data-consistency "$db_url" > "$output_file" 2>&1; then
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
    local output_file="$OUTPUT_DIR/data_consistency_test_summary_$TIMESTAMP.md"
    
    print_status $BLUE "📋 Generating summary report..."
    
    cat > "$output_file" << EOF
# Data Consistency Verification Test Summary

**Test Date:** $(date)
**Timestamp:** $TIMESTAMP

## Test Results

### SQL Tests
EOF

    if [ -f "$OUTPUT_DIR/data_consistency_sql_test_$TIMESTAMP.log" ]; then
        echo "✅ SQL tests completed" >> "$output_file"
        echo "📁 Results: \`data_consistency_sql_test_$TIMESTAMP.log\`" >> "$output_file"
    else
        echo "⚠️ SQL tests skipped (psql not available)" >> "$output_file"
    fi

    cat >> "$output_file" << EOF

### Go Tests
EOF

    if [ -f "$OUTPUT_DIR/data_consistency_go_test_$TIMESTAMP.log" ]; then
        echo "✅ Go tests completed" >> "$output_file"
        echo "📁 Results: \`data_consistency_go_test_$TIMESTAMP.log\`" >> "$output_file"
    else
        echo "⚠️ Go tests skipped (Go not available)" >> "$output_file"
    fi

    cat >> "$output_file" << EOF

## What Was Tested

### Table Existence and Structure
- **Core Table Existence**: Verify all critical tables exist
- **Table Structure**: Validate table schemas and relationships
- **Column Definitions**: Check for required columns and data types

### Count Consistency
- **User-Merchant Consistency**: Users should have merchant records
- **Merchant-Verification Consistency**: Merchants should have verification records
- **Merchant-Classification Consistency**: Merchants should have classification results
- **Merchant-Risk Assessment Consistency**: Merchants should have risk assessments

### Business Logic Consistency
- **Verification Status**: Business verifications should have valid status values
- **Classification Confidence**: Confidence scores should be between 0 and 1
- **Risk Assessment Levels**: Risk levels should be valid (low, medium, high, critical)
- **Email Format**: User emails should follow valid email format
- **Industry Classification**: Classification results should reference valid industries

### Data Integrity Consistency
- **Date Consistency**: Created dates should be before updated dates
- **Timestamp Validation**: All records should have valid timestamps
- **Assessment Dates**: Risk assessments should have valid assessment dates
- **Audit Log Consistency**: Audit logs should have valid timestamps

### Referential Integrity Consistency
- **Foreign Key Consistency**: All foreign key references should be valid
- **Orphaned Records**: No orphaned records should exist
- **Relationship Integrity**: All table relationships should be consistent

### Data Quality Consistency
- **Duplicate Records**: Check for duplicate records where they shouldn't exist
- **NULL Value Consistency**: Critical fields should not be NULL
- **Data Length Consistency**: String values should not exceed column limits

### Business Rule Consistency
- **Verification Workflow**: Verifications should follow proper workflow states
- **Risk Assessment Rules**: High-risk merchants should have risk assessments
- **Classification Rules**: All merchants should have primary classifications

### Performance Consistency
- **Index Consistency**: Foreign key columns should be indexed
- **Query Performance**: Critical queries should perform well

## Common Data Consistency Issues

### 1. Count Inconsistencies
- **Problem**: Related tables have mismatched record counts
- **Example**: Users without merchant records, merchants without verifications
- **Solution**: Implement proper data creation workflows and validation

### 2. Business Logic Violations
- **Problem**: Data doesn't follow business rules
- **Example**: Invalid status values, confidence scores outside 0-1 range
- **Solution**: Add application-level validation and database constraints

### 3. Date Inconsistencies
- **Problem**: Dates don't follow logical order
- **Example**: Created dates after updated dates
- **Solution**: Implement proper timestamp management

### 4. Referential Integrity Issues
- **Problem**: Foreign key references point to non-existent records
- **Example**: Orphaned records, invalid relationships
- **Solution**: Implement proper foreign key constraints and cleanup procedures

## Consistency Improvement Recommendations

### 1. Application-Level Validation
- Add validation rules in application code
- Implement proper error handling for data inconsistencies
- Use database transactions to maintain consistency

### 2. Database Constraints
- Add foreign key constraints where missing
- Implement check constraints for business rules
- Use triggers for complex consistency rules

### 3. Data Quality Monitoring
- Set up regular consistency checks
- Implement automated alerts for consistency issues
- Track consistency metrics over time

### 4. Cleanup Procedures
- Develop procedures for fixing consistency issues
- Implement data migration scripts for bulk fixes
- Create rollback procedures for failed consistency fixes

## Next Steps

1. Review the test results in the log files
2. Identify the root cause of any consistency issues
3. Implement fixes for critical consistency problems
4. Add preventive measures to avoid future issues
5. Re-run tests to verify fixes

## Files Generated

- \`data_consistency_sql_test_$TIMESTAMP.log\` - SQL test results
- \`data_consistency_go_test_$TIMESTAMP.log\` - Go test results
- \`data_consistency_test_summary_$TIMESTAMP.md\` - This summary report

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
    print_status $BLUE "🚀 Starting Data Consistency Verification Testing"
    print_status $BLUE "==============================================="
    
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
        print_status $GREEN "🎉 All data consistency verification tests completed successfully!"
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
