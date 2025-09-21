# Foreign Key Constraint Testing Guide

## Overview

This guide provides comprehensive instructions for testing foreign key constraints in the KYB Platform database. The testing system includes both SQL-based and Go-based validation tools to ensure data integrity across all table relationships.

## 🎯 Purpose

Foreign key constraint testing ensures:
- **Data Integrity**: All foreign key relationships are valid
- **Referential Integrity**: No orphaned records exist
- **Data Consistency**: Related data across tables is synchronized
- **Performance**: Foreign key columns are properly indexed

## 🛠️ Testing Tools

### 1. Go-Based Testing Tool (`test-foreign-keys.go`)

**Features:**
- Comprehensive foreign key constraint discovery
- Automated orphaned record detection
- Performance timing for each test
- Detailed reporting with pass/fail status
- Command-line interface with environment variable support

**Usage:**
```bash
# Using environment variable
export DATABASE_URL="postgresql://user:pass@localhost:5432/dbname"
go run scripts/test-foreign-keys.go

# Using command line argument
go run scripts/test-foreign-keys.go "postgresql://user:pass@localhost:5432/dbname"
```

### 2. SQL-Based Testing Tool (`test-foreign-keys.sql`)

**Features:**
- Static SQL queries for foreign key analysis
- Orphaned record detection queries
- Data type consistency checks
- Missing index analysis
- Comprehensive constraint reporting

**Usage:**
```bash
psql "postgresql://user:pass@localhost:5432/dbname" -f scripts/test-foreign-keys.sql
```

### 3. Automated Test Runner (`run-foreign-key-tests.sh`)

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
./scripts/run-foreign-key-tests.sh

# Using command line argument
./scripts/run-foreign-key-tests.sh "postgresql://user:pass@localhost:5432/dbname"
```

## 📊 Test Results

### Output Files

The testing system generates several output files:

1. **`foreign_key_sql_test_YYYYMMDD_HHMMSS.log`** - SQL test results
2. **`foreign_key_go_test_YYYYMMDD_HHMMSS.log`** - Go test results  
3. **`foreign_key_test_summary_YYYYMMDD_HHMMSS.md`** - Summary report

### Test Status Codes

- **✅ PASS** - No orphaned records found
- **❌ FAIL** - Orphaned records detected
- **🚨 ERROR** - Test execution error

### Sample Output

```
🔍 Starting Foreign Key Constraint Testing...
============================================================
Found 15 foreign key constraints to test

[1/15] Testing merchants.user_id -> users.id
  ✅ PASS - No orphaned records found
  ⏱️  Execution time: 45ms

[2/15] Testing business_verifications.merchant_id -> merchants.id
  ❌ FAIL - Found 3 orphaned records out of 1,247 total
  ⏱️  Execution time: 67ms

============================================================
📊 FOREIGN KEY CONSTRAINT TEST SUMMARY
============================================================
Total Tests: 15
✅ Passed: 13
❌ Failed: 2
🚨 Errors: 0
Success Rate: 86.7%
```

## 🔍 What Gets Tested

### 1. Foreign Key Constraint Discovery

The system automatically discovers all foreign key constraints by querying:
- `information_schema.table_constraints`
- `information_schema.key_column_usage`
- `information_schema.constraint_column_usage`

### 2. Orphaned Record Detection

For each foreign key constraint, the system checks for:
- Records in child tables that reference non-existent parent records
- NULL values in foreign key columns (where appropriate)
- Data type mismatches between foreign key and referenced columns

### 3. Performance Analysis

The system analyzes:
- Missing indexes on foreign key columns
- Query execution times for constraint checks
- Data volume impact on constraint validation

## 🚨 Common Issues and Solutions

### Issue 1: Orphaned Records

**Problem:** Child table records reference non-existent parent records

**Example:**
```sql
-- merchants table has user_id = 999, but users table has no id = 999
SELECT COUNT(*) FROM merchants m
LEFT JOIN users u ON m.user_id = u.id
WHERE m.user_id IS NOT NULL AND u.id IS NULL;
```

**Solution:**
1. Identify the orphaned records
2. Either delete the orphaned records or create the missing parent records
3. Re-run the test to verify the fix

### Issue 2: Missing Indexes

**Problem:** Foreign key columns lack indexes, causing performance issues

**Example:**
```sql
-- Check for missing indexes
SELECT tc.table_name, kcu.column_name
FROM information_schema.table_constraints tc
JOIN information_schema.key_column_usage kcu
    ON tc.constraint_name = kcu.constraint_name
LEFT JOIN pg_indexes i ON i.tablename = tc.table_name
    AND i.indexdef LIKE '%' || kcu.column_name || '%'
WHERE tc.constraint_type = 'FOREIGN KEY'
    AND i.indexname IS NULL;
```

**Solution:**
```sql
-- Create missing indexes
CREATE INDEX CONCURRENTLY idx_merchants_user_id ON merchants(user_id);
CREATE INDEX CONCURRENTLY idx_business_verifications_merchant_id ON business_verifications(merchant_id);
```

### Issue 3: Data Type Mismatches

**Problem:** Foreign key column and referenced column have different data types

**Example:**
```sql
-- Check for data type mismatches
SELECT 
    tc.table_name,
    kcu.column_name,
    sc.data_type as source_type,
    rc.data_type as referenced_type
FROM information_schema.table_constraints tc
JOIN information_schema.key_column_usage kcu
    ON tc.constraint_name = kcu.constraint_name
JOIN information_schema.columns sc
    ON sc.table_name = tc.table_name
    AND sc.column_name = kcu.column_name
JOIN information_schema.columns rc
    ON rc.table_name = ccu.table_name
    AND rc.column_name = ccu.column_name
WHERE tc.constraint_type = 'FOREIGN KEY'
    AND sc.data_type != rc.data_type;
```

**Solution:**
1. Alter the column to match the referenced column type
2. Update any existing data if necessary
3. Re-run the test to verify the fix

## 🔧 Configuration

### Environment Variables

- **`DATABASE_URL`** - PostgreSQL connection string
  - Format: `postgresql://user:password@host:port/database`
  - Example: `postgresql://kyb_user:password@localhost:5432/kyb_platform`

### Database Permissions

The testing tools require the following permissions:
- `SELECT` on all tables being tested
- `SELECT` on `information_schema` tables
- `SELECT` on `pg_indexes` system table

### Performance Considerations

- **Large Tables**: Tests may take longer on tables with millions of records
- **Concurrent Access**: Tests use `LEFT JOIN` queries that are generally safe for concurrent access
- **Index Usage**: Ensure foreign key columns are indexed for optimal performance

## 📈 Monitoring and Alerting

### Integration with Monitoring Systems

The test results can be integrated with monitoring systems:

```bash
# Run tests and check exit code
./scripts/run-foreign-key-tests.sh "$DATABASE_URL"
if [ $? -ne 0 ]; then
    # Send alert to monitoring system
    curl -X POST "https://monitoring.example.com/alerts" \
        -d '{"type": "foreign_key_violation", "severity": "high"}'
fi
```

### Automated Testing

Set up automated testing in CI/CD pipelines:

```yaml
# .github/workflows/foreign-key-tests.yml
name: Foreign Key Tests
on:
  schedule:
    - cron: '0 2 * * *'  # Daily at 2 AM
  push:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Run Foreign Key Tests
        run: ./scripts/run-foreign-key-tests.sh "${{ secrets.DATABASE_URL }}"
```

## 🎯 Best Practices

### 1. Regular Testing

- Run tests daily in production
- Run tests after any schema changes
- Run tests before major deployments

### 2. Test Environment

- Always test in a staging environment first
- Use production data snapshots for realistic testing
- Monitor test performance and optimize as needed

### 3. Documentation

- Document any foreign key constraint changes
- Keep test results for audit purposes
- Update this guide when adding new constraints

### 4. Performance Optimization

- Ensure all foreign key columns are indexed
- Use `CONCURRENTLY` when creating indexes on large tables
- Monitor query performance and optimize as needed

## 🔗 Related Documentation

- [Database Schema Documentation](../docs/database-schema.md)
- [Data Integrity Validation Guide](../docs/data-integrity-validation.md)
- [Performance Optimization Guide](../docs/performance-optimization.md)
- [Monitoring and Alerting Setup](../docs/monitoring-setup.md)

## 📞 Support

For issues or questions regarding foreign key constraint testing:

1. Check the test output logs for specific error messages
2. Review this guide for common solutions
3. Consult the database schema documentation
4. Contact the development team for complex issues

---

**Last Updated:** January 19, 2025  
**Version:** 1.0  
**Next Review:** February 19, 2025
