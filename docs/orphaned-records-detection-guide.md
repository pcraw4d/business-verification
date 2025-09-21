# Orphaned Records Detection Guide

## Overview

This guide provides comprehensive instructions for detecting orphaned records in the KYB Platform database. Orphaned records are child table records that reference non-existent parent records, violating referential integrity and potentially causing application errors.

## 🎯 Purpose

Orphaned records detection ensures:
- **Referential Integrity**: All foreign key relationships are valid
- **Data Consistency**: Related data across tables is synchronized
- **Application Stability**: Prevents errors from invalid references
- **Data Quality**: Maintains high-quality, consistent data

## 🛠️ Detection Tools

### 1. Go-Based Detection Tool (`check-orphaned-records.go`)

**Features:**
- Comprehensive relationship discovery (foreign keys and logical relationships)
- Automated orphaned record detection using LEFT JOIN analysis
- Performance timing for each detection
- Detailed reporting with pass/fail status
- Sample orphaned record values for debugging
- Support for both foreign key and logical business relationships

**Usage:**
```bash
# Using environment variable
export DATABASE_URL="postgresql://user:pass@localhost:5432/dbname"
go run scripts/check-orphaned-records.go

# Using command line argument
go run scripts/check-orphaned-records.go "postgresql://user:pass@localhost:5432/dbname"
```

### 2. SQL-Based Detection Tool (`check-orphaned-records.sql`)

**Features:**
- Static SQL queries for comprehensive orphaned record analysis
- Foreign key relationship detection
- Logical business relationship validation
- Orphaned record statistics and percentages
- Cleanup recommendations with sample queries
- Impact analysis across all relationships

**Usage:**
```bash
psql "postgresql://user:pass@localhost:5432/dbname" -f scripts/check-orphaned-records.sql
```

### 3. Automated Test Runner (`run-orphaned-records-tests.sh`)

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
./scripts/run-orphaned-records-tests.sh

# Using command line argument
./scripts/run-orphaned-records-tests.sh "postgresql://user:pass@localhost:5432/dbname"
```

## 🔍 Detection Coverage

### Foreign Key Relationships

The system automatically detects and validates all foreign key constraints:

1. **merchants.user_id → users.id**
   - Merchants referencing non-existent users
   - Critical for user authentication and authorization

2. **business_verifications.merchant_id → merchants.id**
   - Verifications referencing non-existent merchants
   - Essential for business verification workflows

3. **classification_results.merchant_id → merchants.id**
   - Classification results referencing non-existent merchants
   - Important for business classification accuracy

4. **risk_assessments.merchant_id → merchants.id**
   - Risk assessments referencing non-existent merchants
   - Critical for risk management and compliance

5. **audit_logs.user_id → users.id**
   - Audit logs referencing non-existent users
   - Important for audit trail integrity

### Logical Business Relationships

The system also checks logical business relationships that should exist:

1. **business_verifications.user_id → users.id**
   - Verifications should reference existing users

2. **merchant_audit_logs.merchant_id → merchants.id**
   - Merchant audit logs should reference existing merchants

3. **industry_keywords.industry_id → industries.id**
   - Industry keywords should reference existing industries

4. **business_risk_assessments.business_id → merchants.id**
   - Risk assessments should reference existing merchants

5. **business_risk_assessments.risk_keyword_id → risk_keywords.id**
   - Risk assessments should reference existing risk keywords

## 📊 Detection Results

### Test Status Codes

- ✅ **PASS**: No orphaned records found
- ❌ **FAIL**: Orphaned records detected with counts and samples
- 🚨 **ERROR**: Detection execution errors

### Sample Output

```
🔍 Starting Orphaned Records Detection...
============================================================
Found 12 relationships to check for orphaned records

[1/12] Checking merchants.user_id -> users.id (foreign_key)
  ✅ PASS - No orphaned records found
  ⏱️  Execution time: 45ms

[2/12] Checking business_verifications.merchant_id -> merchants.id (foreign_key)
  ❌ FAIL - Found 3 orphaned records out of 1,247 total
  📝 Sample orphaned values: 550e8400-e29b-41d4-a716-446655440001, 550e8400-e29b-41d4-a716-446655440002
  ⏱️  Execution time: 67ms

============================================================
📊 ORPHANED RECORDS DETECTION SUMMARY
============================================================
Total Tests: 12
✅ Passed: 10
❌ Failed: 2
🚨 Errors: 0
Success Rate: 83.3%
```

## 🚨 Common Orphaned Record Scenarios

### 1. Data Migration Issues

**Problem:** Records created before proper foreign key constraints were in place

**Example:**
```sql
-- Find merchants with invalid user references
SELECT m.id, m.user_id 
FROM merchants m
LEFT JOIN users u ON m.user_id = u.id
WHERE m.user_id IS NOT NULL AND u.id IS NULL;
```

**Solution:**
1. Identify the source of invalid references
2. Either create missing parent records or remove invalid child records
3. Implement proper foreign key constraints
4. Re-run detection to verify fixes

### 2. Cascade Delete Issues

**Problem:** Parent records deleted without proper cascade handling

**Example:**
```sql
-- Find verifications with invalid merchant references
SELECT bv.id, bv.merchant_id 
FROM business_verifications bv
LEFT JOIN merchants m ON bv.merchant_id = m.id
WHERE bv.merchant_id IS NOT NULL AND m.id IS NULL;
```

**Solution:**
```sql
-- Option 1: Delete orphaned verifications
DELETE FROM business_verifications 
WHERE merchant_id NOT IN (SELECT id FROM merchants);

-- Option 2: Set merchant_id to NULL (if column allows NULLs)
UPDATE business_verifications 
SET merchant_id = NULL 
WHERE merchant_id NOT IN (SELECT id FROM merchants);
```

### 3. Application Bugs

**Problem:** Race conditions or transaction issues creating invalid references

**Example:**
```sql
-- Find classification results with invalid merchant references
SELECT cr.id, cr.merchant_id 
FROM classification_results cr
LEFT JOIN merchants m ON cr.merchant_id = m.id
WHERE cr.merchant_id IS NOT NULL AND m.id IS NULL;
```

**Solution:**
1. Fix the application bug causing invalid references
2. Clean up existing orphaned records
3. Add application-level validation
4. Implement proper transaction handling

## 🔧 Cleanup Strategies

### 1. Safe Cleanup Approaches

**Backup First:**
```sql
-- Always backup before cleanup
CREATE TABLE merchants_backup AS SELECT * FROM merchants;
CREATE TABLE business_verifications_backup AS SELECT * FROM business_verifications;
```

**Gradual Cleanup:**
```sql
-- Process orphaned records in batches
DELETE FROM business_verifications 
WHERE merchant_id NOT IN (SELECT id FROM merchants)
AND id IN (
    SELECT id FROM business_verifications 
    WHERE merchant_id NOT IN (SELECT id FROM merchants)
    LIMIT 100
);
```

**Validation After Cleanup:**
```bash
# Re-run orphaned records detection
./scripts/run-orphaned-records-tests.sh "$DATABASE_URL"
```

### 2. Cleanup Queries (Use with Caution)

**Remove Orphaned Merchants:**
```sql
DELETE FROM merchants 
WHERE user_id NOT IN (SELECT id FROM users);
```

**Remove Orphaned Verifications:**
```sql
DELETE FROM business_verifications 
WHERE merchant_id NOT IN (SELECT id FROM merchants);
```

**Remove Orphaned Classification Results:**
```sql
DELETE FROM classification_results 
WHERE merchant_id NOT IN (SELECT id FROM merchants);
```

**Remove Orphaned Risk Assessments:**
```sql
DELETE FROM risk_assessments 
WHERE merchant_id NOT IN (SELECT id FROM merchants);
```

### 3. Prevention Strategies

**Application-Level Validation:**
```go
// Validate foreign key before inserting
func CreateBusinessVerification(merchantID string) error {
    // Check if merchant exists
    exists, err := checkMerchantExists(merchantID)
    if err != nil {
        return err
    }
    if !exists {
        return errors.New("merchant does not exist")
    }
    
    // Proceed with creation
    return createVerification(merchantID)
}
```

**Database Constraints:**
```sql
-- Ensure foreign key constraints are in place
ALTER TABLE business_verifications 
ADD CONSTRAINT fk_business_verifications_merchant_id 
FOREIGN KEY (merchant_id) REFERENCES merchants(id);
```

## 📈 Monitoring and Alerting

### Integration with Monitoring Systems

The detection results can be integrated with monitoring systems:

```bash
# Run detection and check exit code
./scripts/run-orphaned-records-tests.sh "$DATABASE_URL"
if [ $? -ne 0 ]; then
    # Send alert to monitoring system
    curl -X POST "https://monitoring.example.com/alerts" \
        -d '{"type": "orphaned_records", "severity": "high"}'
fi
```

### Automated Detection

Set up automated detection in CI/CD pipelines:

```yaml
# .github/workflows/orphaned-records-detection.yml
name: Orphaned Records Detection
on:
  schedule:
    - cron: '0 4 * * *'  # Daily at 4 AM
  push:
    branches: [main]

jobs:
  detect:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Run Orphaned Records Detection
        run: ./scripts/run-orphaned-records-tests.sh "${{ secrets.DATABASE_URL }}"
```

## 🎯 Best Practices

### 1. Regular Detection

- Run detection daily in production
- Run detection after any data migrations
- Run detection before major deployments

### 2. Proactive Prevention

- Implement foreign key constraints
- Add application-level validation
- Use proper transaction handling
- Implement cascade delete strategies

### 3. Data Quality Monitoring

- Monitor orphaned record counts over time
- Set up alerts for orphaned record detection
- Track data quality metrics

### 4. Cleanup Procedures

- Always backup before cleanup
- Process orphaned records in batches
- Validate fixes after cleanup
- Document cleanup procedures

## 🔗 Related Documentation

- [Foreign Key Constraint Testing Guide](../docs/foreign-key-testing-guide.md)
- [Data Type Validation Guide](../docs/data-type-validation-guide.md)
- [Database Schema Documentation](../docs/database-schema.md)
- [Data Integrity Validation Guide](../docs/data-integrity-validation.md)

## 📞 Support

For issues or questions regarding orphaned records detection:

1. Check the detection output logs for specific error messages
2. Review this guide for common solutions
3. Consult the database schema documentation
4. Contact the development team for complex issues

---

**Last Updated:** January 19, 2025  
**Version:** 1.0  
**Next Review:** February 19, 2025
