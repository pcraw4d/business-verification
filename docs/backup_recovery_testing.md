# Backup and Recovery Testing Documentation

## Overview

This document describes the comprehensive backup and recovery testing framework implemented for the Supabase Table Improvement project. The testing framework ensures that our enhanced classification system, risk keywords, and ML model data can be safely backed up and recovered.

## Architecture

The backup and recovery testing framework follows a modular design with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────────┐
│                Backup Recovery Test Runner                  │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────┐  ┌──────────────┐ │
│  │   Backup        │  │   Recovery      │  │   Point-in-  │ │
│  │   Procedures    │  │   Scenarios     │  │   Time       │ │
│  │   Testing       │  │   Testing       │  │   Recovery   │ │
│  └─────────────────┘  └─────────────────┘  └──────────────┘ │
├─────────────────────────────────────────────────────────────┤
│                Data Restoration Validation                  │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────┐  ┌──────────────┐ │
│  │   Data          │  │   Foreign Key   │  │   Index      │ │
│  │   Integrity     │  │   Constraints   │  │   Validation │ │
│  │   Validation    │  │   Validation    │  │              │ │
│  └─────────────────┘  └─────────────────┘  └──────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Components

### 1. BackupRecoveryTester

The main testing component that orchestrates all backup and recovery operations.

**Key Features:**
- Full database backup testing
- Incremental backup testing
- Schema-only backup testing
- Data-only backup testing
- Complete database recovery testing
- Partial table recovery testing
- Schema recovery testing
- Data integrity validation
- Point-in-time recovery testing

### 2. BackupRecoveryTestRunner

Orchestrates the complete testing process and generates comprehensive reports.

**Key Features:**
- Runs all test suites
- Generates detailed reports
- Provides recommendations
- Creates human-readable summaries
- Validates overall test results

### 3. Configuration Management

Flexible configuration system supporting different environments.

**Configuration Types:**
- Development configuration
- Production configuration
- Custom configuration

## Test Categories

### 1. Backup Procedures Testing

Tests various backup strategies to ensure data can be safely preserved.

**Tests Included:**
- **Full Database Backup**: Complete database backup using pg_dump
- **Incremental Backup**: Table-specific backups for frequently changing data
- **Schema-Only Backup**: Database structure backup
- **Data-Only Backup**: Data content backup

**Validation:**
- Backup file creation
- File size validation
- Content verification

### 2. Recovery Scenarios Testing

Tests different recovery scenarios to ensure data can be restored.

**Tests Included:**
- **Complete Database Recovery**: Full database restoration
- **Partial Table Recovery**: Individual table restoration
- **Schema Recovery**: Database structure restoration

**Validation:**
- Recovery process completion
- Data accessibility
- System functionality

### 3. Data Restoration Validation

Comprehensive validation of restored data integrity.

**Validation Areas:**
- **Data Integrity**: Row counts, NULL value checks
- **Foreign Key Constraints**: Orphaned record detection
- **Index Validation**: Missing index detection
- **Classification System**: Critical table validation

**Metrics:**
- Integrity score calculation
- Constraint violation detection
- Index completeness validation

### 4. Point-in-Time Recovery Testing

Tests the ability to recover data to specific timestamps.

**Test Process:**
1. Create timestamped test data
2. Recover to specific timestamps
3. Validate recovered data state
4. Verify data consistency

**Validation:**
- Timestamp accuracy
- Data state correctness
- Consistency verification

## Configuration

### Environment Variables

```bash
# Database connections
SUPABASE_URL=postgresql://user:password@host:port/database
TEST_DATABASE_URL=postgresql://user:password@host:port/test_database

# Backup configuration
BACKUP_DIRECTORY=/path/to/backup/directory
TEST_DATA_SIZE=1000
RECOVERY_TIMEOUT=10m
VALIDATION_RETRIES=3
```

### Configuration Examples

#### Development Configuration
```go
config := DevelopmentBackupTestConfig()
// Uses local databases with smaller test data
```

#### Production Configuration
```go
config := ProductionBackupTestConfig()
// Uses production databases with larger test data
```

## Usage

### Running Individual Tests

```bash
# Run only backup procedures test
go test -run TestBackupProceduresOnly ./internal/testing

# Run only recovery scenarios test
go test -run TestRecoveryScenariosOnly ./internal/testing

# Run only data restoration validation
go test -run TestDataRestorationOnly ./internal/testing

# Run only point-in-time recovery test
go test -run TestPointInTimeRecoveryOnly ./internal/testing
```

### Running Complete Test Suite

```bash
# Run all backup and recovery tests
go test -run TestBackupRecoveryIntegration ./internal/testing

# Run with verbose output
go test -v -run TestBackupRecoveryIntegration ./internal/testing
```

### Running Benchmarks

```bash
# Benchmark backup procedures
go test -bench=BenchmarkBackupProcedures ./internal/testing

# Benchmark recovery procedures
go test -bench=BenchmarkRecoveryProcedures ./internal/testing
```

## Test Results and Reporting

### Report Structure

The testing framework generates comprehensive reports including:

1. **Test Summary**
   - Total tests run
   - Pass/fail statistics
   - Duration metrics
   - Average validation scores

2. **Individual Test Results**
   - Test name and status
   - Duration and performance metrics
   - Error messages (if any)
   - Validation scores

3. **Recommendations**
   - Performance improvements
   - Configuration suggestions
   - Best practices
   - Future enhancements

### Report Files

- `backup_recovery_test_report.json`: Machine-readable JSON report
- `backup_recovery_test_summary.txt`: Human-readable summary

### Sample Report Output

```
BACKUP AND RECOVERY TEST REPORT
================================

Test Date: 2025-01-19 15:30:45
Total Tests: 4
Passed Tests: 4
Failed Tests: 0
Total Duration: 12m34s
Average Validation Score: 98.50%

TEST RESULTS
============

Test: Backup Procedures Test
Status: PASS
Duration: 3m12s
Validation Score: 100.00%

Test: Recovery Scenarios Test
Status: PASS
Duration: 4m45s
Validation Score: 97.50%

Test: Data Restoration Validation
Status: PASS
Duration: 2m18s
Validation Score: 98.75%

Test: Point-in-Time Recovery Test
Status: PASS
Duration: 2m19s
Recovery Time: 1m45s
Validation Score: 97.75%

RECOMMENDATIONS
===============

1. Implement automated backup testing in CI/CD pipeline
2. Set up monitoring and alerting for backup failures
3. Document recovery procedures and create runbooks
4. Conduct regular disaster recovery drills
5. Consider implementing backup encryption for sensitive data
```

## Integration with Existing Systems

### Classification System Integration

The backup and recovery testing specifically validates:

- **Industries Table**: Core industry classification data
- **Industry Keywords**: Keyword mapping and weights
- **Risk Keywords**: Risk detection patterns
- **Industry Code Crosswalks**: MCC/NAICS/SIC mappings
- **Business Risk Assessments**: Risk assessment results

### ML Model Data Protection

The testing framework ensures that:

- ML model training data is properly backed up
- Model parameters and configurations are preserved
- Classification results are recoverable
- Risk assessment data is maintained

### Performance Considerations

- **Backup Performance**: Optimized for large datasets
- **Recovery Performance**: Fast restoration capabilities
- **Validation Performance**: Efficient integrity checking
- **Storage Optimization**: Compressed backup files

## Best Practices

### Backup Strategy

1. **Regular Full Backups**: Daily complete database backups
2. **Incremental Backups**: Hourly incremental backups for critical tables
3. **Schema Backups**: Weekly schema-only backups
4. **Retention Policy**: 30-day backup retention

### Recovery Strategy

1. **Recovery Testing**: Weekly recovery testing
2. **Point-in-Time Recovery**: Support for 24-hour recovery window
3. **Disaster Recovery**: Complete system recovery procedures
4. **Data Validation**: Comprehensive integrity checking

### Monitoring and Alerting

1. **Backup Monitoring**: Automated backup success/failure alerts
2. **Recovery Monitoring**: Recovery time and success rate tracking
3. **Data Integrity Monitoring**: Continuous integrity validation
4. **Performance Monitoring**: Backup and recovery performance metrics

## Troubleshooting

### Common Issues

1. **Backup Failures**
   - Check database connectivity
   - Verify backup directory permissions
   - Ensure sufficient disk space

2. **Recovery Failures**
   - Validate backup file integrity
   - Check test database connectivity
   - Verify recovery permissions

3. **Data Integrity Issues**
   - Review foreign key constraints
   - Check for data corruption
   - Validate index completeness

### Debug Mode

Enable debug logging for detailed troubleshooting:

```bash
export DEBUG_BACKUP_RECOVERY=true
go test -v -run TestBackupRecoveryIntegration ./internal/testing
```

## Future Enhancements

### Planned Improvements

1. **Automated Testing**: CI/CD pipeline integration
2. **Cloud Integration**: AWS/Azure backup testing
3. **Encryption Support**: Encrypted backup testing
4. **Performance Optimization**: Parallel backup/recovery
5. **Advanced Validation**: Machine learning-based integrity checking

### Monitoring Integration

1. **Prometheus Metrics**: Backup/recovery metrics
2. **Grafana Dashboards**: Visual monitoring
3. **Alerting Rules**: Automated alerting
4. **Log Aggregation**: Centralized logging

## Conclusion

The backup and recovery testing framework provides comprehensive validation of our Supabase database improvements, ensuring that our enhanced classification system, risk keywords, and ML model data are properly protected and recoverable. The modular design allows for easy extension and customization while maintaining high reliability and performance standards.
