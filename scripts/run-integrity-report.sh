#!/bin/bash

# Comprehensive Data Integrity Report Generation Script
# This script generates comprehensive data integrity reports

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

# Function to run SQL report generation
run_sql_report() {
    local db_url=$1
    local output_file="$OUTPUT_DIR/integrity_report_sql_$TIMESTAMP.log"
    
    print_status $BLUE "📊 Generating SQL-based data integrity report..."
    
    if command_exists psql; then
        print_status $BLUE "   Executing SQL report generation script..."
        if psql "$db_url" -f "$SCRIPT_DIR/generate-integrity-report.sql" > "$output_file" 2>&1; then
            print_status $GREEN "✅ SQL report generated successfully"
            print_status $BLUE "   Results saved to: $output_file"
            
            # Check for any issues in the report
            if grep -q "FAIL\|ERROR\|POOR" "$output_file"; then
                print_status $YELLOW "⚠️  Some data integrity issues detected - check report"
            else
                print_status $GREEN "✅ No data integrity issues detected"
            fi
            
            return 0
        else
            print_status $RED "❌ SQL report generation failed"
            print_status $RED "   Check error log: $output_file"
            return 1
        fi
    else
        print_status $YELLOW "⚠️  psql not found, skipping SQL report generation"
        return 0
    fi
}

# Function to run Go report generation
run_go_report() {
    local db_url=$1
    local output_file="$OUTPUT_DIR/integrity_report_go_$TIMESTAMP.log"
    
    print_status $BLUE "🔧 Generating Go-based data integrity report..."
    
    if command_exists go; then
        print_status $BLUE "   Compiling and running Go report generator..."
        
        # Change to script directory
        cd "$SCRIPT_DIR"
        
        # Compile the Go program
        if go build -o generate-integrity-report generate-integrity-report.go; then
            print_status $GREEN "✅ Go report generator compiled successfully"
            
            # Run the report generator
            if ./generate-integrity-report "$db_url" > "$output_file" 2>&1; then
                print_status $GREEN "✅ Go report generated successfully"
                print_status $BLUE "   Results saved to: $output_file"
                
                # Check for generated report files
                if ls data_integrity_report_*.html data_integrity_report_*.json data_integrity_report_*.md >/dev/null 2>&1; then
                    print_status $GREEN "✅ Report files generated successfully"
                    print_status $BLUE "   HTML, JSON, and Markdown reports created"
                fi
                
                return 0
            else
                print_status $RED "❌ Go report generation failed"
                print_status $RED "   Check error log: $output_file"
                return 1
            fi
        else
            print_status $RED "❌ Failed to compile Go report generator"
            return 1
        fi
    else
        print_status $YELLOW "⚠️  Go not found, skipping Go report generation"
        return 0
    fi
}

# Function to generate summary report
generate_summary_report() {
    local output_file="$OUTPUT_DIR/integrity_report_summary_$TIMESTAMP.md"
    
    print_status $BLUE "📋 Generating summary report..."
    
    cat > "$output_file" << EOF
# Comprehensive Data Integrity Report Summary

**Generated:** $(date)
**Timestamp:** $TIMESTAMP

## Report Generation Results

### SQL Report
EOF

    if [ -f "$OUTPUT_DIR/integrity_report_sql_$TIMESTAMP.log" ]; then
        echo "✅ SQL report generated successfully" >> "$output_file"
        echo "📁 Results: \`integrity_report_sql_$TIMESTAMP.log\`" >> "$output_file"
        
        # Extract key metrics from SQL report
        if grep -q "EXECUTIVE SUMMARY" "$OUTPUT_DIR/integrity_report_sql_$TIMESTAMP.log"; then
            echo "" >> "$output_file"
            echo "### Key Metrics from SQL Report" >> "$output_file"
            echo "\`\`\`" >> "$output_file"
            grep -A 10 "EXECUTIVE SUMMARY" "$OUTPUT_DIR/integrity_report_sql_$TIMESTAMP.log" >> "$output_file"
            echo "\`\`\`" >> "$output_file"
        fi
    else
        echo "⚠️ SQL report skipped (psql not available)" >> "$output_file"
    fi

    cat >> "$output_file" << EOF

### Go Report
EOF

    if [ -f "$OUTPUT_DIR/integrity_report_go_$TIMESTAMP.log" ]; then
        echo "✅ Go report generated successfully" >> "$output_file"
        echo "📁 Results: \`integrity_report_go_$TIMESTAMP.log\`" >> "$output_file"
        
        # Check for generated report files
        if ls data_integrity_report_*.html data_integrity_report_*.json data_integrity_report_*.md >/dev/null 2>&1; then
            echo "" >> "$output_file"
            echo "### Generated Report Files" >> "$output_file"
            for file in data_integrity_report_*.html data_integrity_report_*.json data_integrity_report_*.md; do
                if [ -f "$file" ]; then
                    echo "- \`$file\`" >> "$output_file"
                fi
            done
        fi
    else
        echo "⚠️ Go report skipped (Go not available)" >> "$output_file"
    fi

    cat >> "$output_file" << EOF

## What Was Analyzed

### Database Structure
- **Table Existence**: Verification that all critical tables exist
- **Table Statistics**: Row counts, update frequencies, and activity levels
- **Schema Validation**: Column definitions and data types

### Foreign Key Integrity
- **Constraint Validation**: All foreign key constraints are properly defined
- **Referential Integrity**: No orphaned records in foreign key relationships
- **Relationship Analysis**: Comprehensive analysis of all table relationships

### Data Type and Format Validation
- **Email Format Validation**: User emails follow proper email format
- **Date Consistency**: Created dates are before updated dates
- **Status Value Validation**: Business verification statuses are valid
- **Confidence Score Validation**: Classification confidence scores are within 0-1 range
- **Risk Level Validation**: Risk assessment levels are valid

### Data Consistency Analysis
- **User-Merchant Consistency**: Users have corresponding merchant records
- **Merchant-Verification Consistency**: Merchants have verification records
- **Merchant-Classification Consistency**: Merchants have classification results
- **Cross-Table Consistency**: Data consistency across related tables

### Data Quality Metrics
- **NULL Value Analysis**: Critical fields are not NULL
- **Duplicate Record Detection**: No duplicate records where they shouldn't exist
- **Data Length Validation**: String values don't exceed column limits
- **Data Completeness**: Required fields are populated

### Performance Analysis
- **Index Analysis**: Foreign key columns are properly indexed
- **Query Performance**: Critical queries perform well
- **Database Statistics**: Table and index statistics

### Business Rule Validation
- **Workflow Consistency**: Business processes follow proper workflows
- **Risk Assessment Rules**: High-risk merchants have proper assessments
- **Classification Rules**: All merchants have primary classifications

## Report Types Generated

### 1. SQL Report
- **Format**: Text-based log file
- **Content**: Comprehensive SQL queries and results
- **Use Case**: Technical analysis and debugging

### 2. Go Report
- **HTML Report**: Interactive web-based report with styling
- **JSON Report**: Machine-readable format for integration
- **Markdown Report**: Documentation-friendly format
- **Use Case**: Executive summaries and documentation

## Key Findings

### Data Integrity Status
EOF

    # Extract status from reports
    if [ -f "$OUTPUT_DIR/integrity_report_sql_$TIMESTAMP.log" ]; then
        if grep -q "EXCELLENT\|GOOD\|FAIR\|POOR" "$OUTPUT_DIR/integrity_report_sql_$TIMESTAMP.log"; then
            status=$(grep -o "EXCELLENT\|GOOD\|FAIR\|POOR" "$OUTPUT_DIR/integrity_report_sql_$TIMESTAMP.log" | head -1)
            echo "- **Overall Health Status**: $status" >> "$output_file"
        fi
    fi

    cat >> "$output_file" << EOF

### Critical Issues
- Review the detailed reports for any critical data integrity issues
- Pay special attention to foreign key violations and data type issues
- Address any business rule violations immediately

### Recommendations
- Implement automated data integrity monitoring
- Add database constraints to prevent future issues
- Set up regular integrity checks and alerts
- Consider implementing application-level validation

## Next Steps

1. **Review Reports**: Examine all generated reports for data integrity issues
2. **Prioritize Issues**: Focus on critical issues first
3. **Implement Fixes**: Address identified data integrity problems
4. **Monitor Continuously**: Set up ongoing data integrity monitoring
5. **Document Changes**: Keep track of all integrity improvements

## Files Generated

- \`integrity_report_sql_$TIMESTAMP.log\` - SQL-based report
- \`integrity_report_go_$TIMESTAMP.log\` - Go-based report
- \`integrity_report_summary_$TIMESTAMP.md\` - This summary report
- \`data_integrity_report_*.html\` - HTML report (if Go available)
- \`data_integrity_report_*.json\` - JSON report (if Go available)
- \`data_integrity_report_*.md\` - Markdown report (if Go available)

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
    print_status $BLUE "🚀 Starting Comprehensive Data Integrity Report Generation"
    print_status $BLUE "========================================================"
    
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
    
    # Generate reports
    local sql_result=0
    local go_result=0
    
    run_sql_report "$db_url" || sql_result=1
    echo ""
    
    run_go_report "$db_url" || go_result=1
    echo ""
    
    # Generate summary
    generate_summary_report
    echo ""
    
    # Final status
    if [ $sql_result -eq 0 ] && [ $go_result -eq 0 ]; then
        print_status $GREEN "🎉 All data integrity reports generated successfully!"
        print_status $BLUE "📁 Check the reports in: $OUTPUT_DIR"
        print_status $BLUE "📄 Generated report files:"
        ls -la data_integrity_report_* 2>/dev/null || true
        exit 0
    else
        print_status $YELLOW "⚠️  Some report generation failed - check the logs for details"
        print_status $BLUE "📁 Check the reports in: $OUTPUT_DIR"
        exit 1
    fi
}

# Run main function
main "$@"
