# Data Consistency Verification Guide

## Overview

This guide provides comprehensive instructions for verifying data consistency across related tables in the KYB Platform database. Data consistency ensures that related data across different tables is synchronized, accurate, and follows business rules.

## 🎯 Purpose

Data consistency verification ensures:
- **Referential Integrity**: All relationships between tables are valid
- **Business Rule Compliance**: Data follows established business logic
- **Data Quality**: High-quality, consistent data across all tables
- **Application Stability**: Prevents errors from inconsistent data

## 🛠️ Verification Tools

### 1. Go-Based Verification Tool (`verify-data-consistency.go`)

**Features:**
- Comprehensive consistency test discovery and execution
- Multiple test types: count, sum, avg, min, max, custom
- Automated test result comparison with tolerance support
- Performance timing for each verification
- Detailed reporting with pass/fail status
- Critical test identification and prioritization
- Support for complex business logic validation

**Usage:**
```bash
# Using environment variable
export DATABASE_URL="postgresql://user:pass@localhost:5432/dbname"
go run scripts/verify-data-consistency.go

# Using command line argument
go run scripts/verify-data-consistency.go "postgresql://user:pass@localhost:5432/dbname"
```

### 2. SQL-Based Verification Tool (`verify-data-consistency.sql`)

**Features:**
- Static SQL queries for comprehensive consistency analysis
- Table existence and structure verification
- Count consistency across related tables
- Business logic consistency validation
- Data integrity consistency checks
- Referential integrity verification
- Data quality consistency analysis
- Business rule consistency validation
- Performance consistency checks

**Usage:**
```bash
psql "postgresql://user:pass@localhost:5432/dbname" -f scripts/verify-data-consistency.sql
```

### 3. Automated Test Runner (`run-data-consistency-tests.sh`)

**Features:**
- Runs both Go and SQL tests automatically
- Generates comprehensive reports
- Timestamped output files
- Colored console output
- Error handling and validation

**Usage:**
```bash
# Using environment variable
export DATABASE_URL="postgresql://user:pass@localhost:5432/dbname"
./scripts/run-data-consistency-tests.sh

# Using command line argument
./scripts/run-data-consistency-tests.sh "postgresql://user:pass@localhost:5432/dbname"
```

## 🔍 Verification Coverage

### Table Existence and Structure

The system verifies that all critical tables exist and have proper structure:

1. **Core Tables**: users, merchants, business_verifications, classification_results, risk_assessments, audit_logs
2. **Supporting Tables**: industries, industry_keywords, risk_keywords, business_risk_assessments, merchant_audit_logs
3. **Table Structure**: Required columns, data types, and relationships

### Count Consistency

The system verifies count consistency across related tables:

1. **User-Merchant Consistency**: Users should have merchant records
2. **Merchant-Verification Consistency**: Merchants should have verification records
3. **Merchant-Classification Consistency**: Merchants should have classification results
4. **Merchant-Risk Assessment Consistency**: Merchants should have risk assessments

### Business Logic Consistency

The system validates business logic consistency:

1. **Verification Status**: Business verifications should have valid status values (pending, approved, rejected, in_progress, completed)
2. **Classification Confidence**: Confidence scores should be between 0 and 1
3. **Risk Assessment Levels**: Risk levels should be valid (low, medium, high, critical)
4. **Email Format**: User emails should follow valid email format
5. **Industry Classification**: Classification results should reference valid industries

### Data Integrity Consistency

The system checks data integrity consistency:

1. **Date Consistency**: Created dates should be before updated dates
2. **Timestamp Validation**: All records should have valid timestamps
3. **Assessment Dates**: Risk assessments should have valid assessment dates
4. **Audit Log Consistency**: Audit logs should have valid timestamps

### Referential Integrity Consistency

The system verifies referential integrity:

1. **Foreign Key Consistency**: All foreign key references should be valid
2. **Orphaned Records**: No orphaned records should exist
3. **Relationship Integrity**: All table relationships should be consistent

### Data Quality Consistency

The system validates data quality:

1. **Duplicate Records**: Check for duplicate records where they shouldn't exist
2. **NULL Value Consistency**: Critical fields should not be NULL
3. **Data Length Consistency**: String values should not exceed column limits

### Business Rule Consistency

The system validates business rule consistency:

1. **Verification Workflow**: Verifications should follow proper workflow states
2. **Risk Assessment Rules**: High-risk merchants should have risk assessments
3. **Classification Rules**: All merchants should have primary classifications

### Performance Consistency

The system checks performance consistency:

1. **Index Consistency**: Foreign key columns should be indexed
2. **Query Performance**: Critical queries should perform well

## 📊 Verification Results

### Test Status Codes

- ✅ **PASS**: Consistency check passed
- ❌ **FAIL**: Consistency check failed with specific details
- 🚨 **ERROR**: Verification execution errors

### Test Types

- **count**: Count-based consistency checks
- **sum**: Sum-based consistency checks
- **avg**: Average-based consistency checks
- **min**: Minimum value consistency checks
- **max**: Maximum value consistency checks
- **custom**: Custom business logic consistency checks

### Sample Output

```
🔍 Starting Data Consistency Verification...
============================================================
Found 24 consistency tests to run

[1/24] Table Exists: users (count)
  ✅ PASS - Verify that table users exists
  ⏱️  Execution time: 12ms

[2/24] Business Verification Status Consistency (custom)
  ❌ FAIL - Business verifications should have valid status values
  📊 Expected: 0, Actual: 3, Difference: 3.00
  ⏱️  Execution time: 45ms

[3/24] Classification Confidence Score Consistency (custom)
  ✅ PASS - Classification results should have confidence scores between 0 and 1
  ⏱️  Execution time: 38ms

============================================================
📊 DATA CONSISTENCY VERIFICATION SUMMARY
============================================================
Total Tests: 24
✅ Passed: 22
❌ Failed: 2
🚨 Errors: 0
Success Rate: 91.7%
```

## 🚨 Common Data Consistency Issues

### 1. Count Inconsistencies

**Problem:** Related tables have mismatched record counts

**Example:**
```sql
-- Find users without merchant records
SELECT COUNT(*) 
FROM users u 
WHERE NOT EXISTS (
    SELECT 1 FROM merchants m WHERE m.user_id = u.id
)
AND EXISTS (SELECT 1 FROM merchants LIMIT 1);
```

**Solution:**
1. Identify the root cause of missing relationships
2. Implement proper data creation workflows
3. Add validation to ensure relationships are created
4. Re-run verification to confirm fixes

### 2. Business Logic Violations

**Problem:** Data doesn't follow business rules

**Example:**
```sql
-- Find business verifications with invalid status values
SELECT COUNT(*) 
FROM business_verifications 
WHERE status NOT IN ('pending', 'approved', 'rejected', 'in_progress', 'completed');
```

**Solution:**
```sql
-- Fix invalid status values
UPDATE business_verifications 
SET status = 'pending' 
WHERE status NOT IN ('pending', 'approved', 'rejected', 'in_progress', 'completed');
```

### 3. Date Inconsistencies

**Problem:** Dates don't follow logical order

**Example:**
```sql
-- Find records where created date is after updated date
SELECT COUNT(*) 
FROM merchants 
WHERE created_at > updated_at;
```

**Solution:**
```sql
-- Fix date inconsistencies
UPDATE merchants 
SET updated_at = created_at 
WHERE created_at > updated_at;
```

### 4. Referential Integrity Issues

**Problem:** Foreign key references point to non-existent records

**Example:**
```sql
-- Find orphaned business verifications
SELECT COUNT(*) 
FROM business_verifications bv
LEFT JOIN merchants m ON bv.merchant_id = m.id
WHERE bv.merchant_id IS NOT NULL AND m.id IS NULL;
```

**Solution:**
```sql
-- Remove orphaned verifications
DELETE FROM business_verifications 
WHERE merchant_id NOT IN (SELECT id FROM merchants);
```

## 🔧 Consistency Improvement Strategies

### 1. Application-Level Validation

**Add validation rules in application code:**
```go
// Validate business verification status
func ValidateVerificationStatus(status string) error {
    validStatuses := []string{"pending", "approved", "rejected", "in_progress", "completed"}
    for _, validStatus := range validStatuses {
        if status == validStatus {
            return nil
        }
    }
    return errors.New("invalid verification status")
}

// Validate classification confidence score
func ValidateConfidenceScore(score float64) error {
    if score < 0 || score > 1 {
        return errors.New("confidence score must be between 0 and 1")
    }
    return nil
}
```

### 2. Database Constraints

**Add check constraints for business rules:**
```sql
-- Add check constraint for verification status
ALTER TABLE business_verifications 
ADD CONSTRAINT chk_verification_status 
CHECK (status IN ('pending', 'approved', 'rejected', 'in_progress', 'completed'));

-- Add check constraint for confidence score
ALTER TABLE classification_results 
ADD CONSTRAINT chk_confidence_score 
CHECK (confidence_score >= 0 AND confidence_score <= 1);

-- Add check constraint for risk level
ALTER TABLE risk_assessments 
ADD CONSTRAINT chk_risk_level 
CHECK (risk_level IN ('low', 'medium', 'high', 'critical'));
```

### 3. Data Quality Monitoring

**Set up regular consistency checks:**
```bash
# Daily consistency check
0 4 * * * /path/to/scripts/run-data-consistency-tests.sh "$DATABASE_URL"

# Weekly comprehensive check
0 2 * * 0 /path/to/scripts/run-data-consistency-tests.sh "$DATABASE_URL"
```

### 4. Cleanup Procedures

**Develop procedures for fixing consistency issues:**
```sql
-- Procedure to fix date inconsistencies
CREATE OR REPLACE FUNCTION fix_date_inconsistencies()
RETURNS void AS $$
BEGIN
    UPDATE merchants 
    SET updated_at = created_at 
    WHERE created_at > updated_at;
    
    UPDATE business_verifications 
    SET updated_at = created_at 
    WHERE created_at > updated_at;
    
    UPDATE classification_results 
    SET updated_at = created_at 
    WHERE created_at > updated_at;
END;
$$ LANGUAGE plpgsql;
```

## 📈 Monitoring and Alerting

### Integration with Monitoring Systems

The verification results can be integrated with monitoring systems:

```bash
# Run verification and check exit code
./scripts/run-data-consistency-tests.sh "$DATABASE_URL"
if [ $? -ne 0 ]; then
    # Send alert to monitoring system
    curl -X POST "https://monitoring.example.com/alerts" \
        -d '{"type": "data_consistency", "severity": "high"}'
fi
```

### Automated Verification

Set up automated verification in CI/CD pipelines:

```yaml
# .github/workflows/data-consistency-verification.yml
name: Data Consistency Verification
on:
  schedule:
    - cron: '0 4 * * *'  # Daily at 4 AM
  push:
    branches: [main]

jobs:
  verify:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Run Data Consistency Verification
        run: ./scripts/run-data-consistency-tests.sh "${{ secrets.DATABASE_URL }}"
```

## 🎯 Best Practices

### 1. Regular Verification

- Run verification daily in production
- Run verification after any data migrations
- Run verification before major deployments

### 2. Proactive Prevention

- Implement database constraints
- Add application-level validation
- Use proper transaction handling
- Implement data quality checks

### 3. Data Quality Monitoring

- Monitor consistency metrics over time
- Set up alerts for consistency issues
- Track data quality trends

### 4. Consistency Improvement

- Always backup before making changes
- Process consistency fixes in batches
- Validate fixes after implementation
- Document consistency procedures

## 🔗 Related Documentation

- [Foreign Key Constraint Testing Guide](../docs/foreign-key-testing-guide.md)
- [Data Type Validation Guide](../docs/data-type-validation-guide.md)
- [Orphaned Records Detection Guide](../docs/orphaned-records-detection-guide.md)
- [Database Schema Documentation](../docs/database-schema.md)
- [Data Integrity Validation Guide](../docs/data-integrity-validation.md)

## 📞 Support

For issues or questions regarding data consistency verification:

1. Check the verification output logs for specific error messages
2. Review this guide for common solutions
3. Consult the database schema documentation
4. Contact the development team for complex issues

---

**Last Updated:** January 19, 2025  
**Version:** 1.0  
**Next Review:** February 19, 2025
