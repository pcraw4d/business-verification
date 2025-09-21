# Database Backup Procedures

## 🎯 **Overview**

This document outlines the comprehensive backup procedures for the Supabase database as part of the Table Improvement Implementation Plan. These procedures ensure data safety before making any schema changes.

## 📋 **Backup Requirements**

### **Critical Requirements**
- ✅ Complete database backup before any schema changes
- ✅ Backup integrity verification
- ✅ Secure storage of backup files
- ✅ Documented recovery procedures
- ✅ Automated backup validation

### **Backup Scope**
- All user tables in the `public` schema
- Table structures and constraints
- Data integrity and relationships
- Metadata and configuration

## 🛠️ **Backup Infrastructure**

### **Components**

#### **1. Backup Manager (`internal/database/backup/supabase_backup.go`)**
- **Purpose**: Core backup functionality
- **Features**:
  - Complete table backup
  - Integrity verification
  - Checksum calculation
  - Metadata management
  - Retention policy enforcement

#### **2. Backup CLI Tool (`cmd/backup/main.go`)**
- **Purpose**: Command-line interface for backup operations
- **Features**:
  - Create backups
  - List available backups
  - Cleanup old backups
  - Verify backup integrity

#### **3. Backup Script (`scripts/backup-database.sh`)**
- **Purpose**: Automated backup execution
- **Features**:
  - Environment validation
  - Error handling
  - Progress reporting
  - Backup verification

## 🚀 **Backup Procedures**

### **Step 1: Pre-Backup Validation**

#### **Environment Check**
```bash
# Verify required environment variables
echo "SUPABASE_URL: ${SUPABASE_URL:-NOT_SET}"
echo "SUPABASE_API_KEY: ${SUPABASE_API_KEY:-NOT_SET}"
echo "SUPABASE_SERVICE_ROLE_KEY: ${SUPABASE_SERVICE_ROLE_KEY:-NOT_SET}"
```

#### **Connection Test**
```bash
# Test Supabase connection
go run test-supabase-connection.go
```

### **Step 2: Create Full Backup**

#### **Using the Backup Script (Recommended)**
```bash
# Create a complete backup
./scripts/backup-database.sh backup
```

#### **Using the CLI Tool Directly**
```bash
# Create backup with custom settings
go run cmd/backup/main.go \
  -output ./backups \
  -retention 30 \
  -verify \
  -timeout 30m
```

#### **Manual Backup Process**
```bash
# 1. Create backup directory
mkdir -p backups/backup_$(date +%Y%m%d_%H%M%S)

# 2. Run backup tool
go run cmd/backup/main.go -output ./backups

# 3. Verify backup integrity
./scripts/backup-database.sh list
```

### **Step 3: Backup Verification**

#### **Automatic Verification**
The backup system automatically verifies:
- ✅ All table files exist
- ✅ File checksums match
- ✅ Metadata is complete
- ✅ Backup structure is valid

#### **Manual Verification**
```bash
# List all backups
./scripts/backup-database.sh list

# Check specific backup
ls -la backups/backup_YYYYMMDD_HHMMSS/
cat backups/backup_YYYYMMDD_HHMMSS/backup_metadata.json
```

### **Step 4: Backup Storage**

#### **Local Storage**
- **Location**: `./backups/backup_YYYYMMDD_HHMMSS/`
- **Structure**:
  ```
  backup_20250119_143022/
  ├── backup_metadata.json
  ├── users.json
  ├── businesses.json
  ├── merchants.json
  ├── business_classifications.json
  └── ... (other table files)
  ```

#### **Backup Metadata**
```json
{
  "backup_id": "backup_20250119_143022",
  "timestamp": "2025-01-19T14:30:22Z",
  "database_url": "https://qpqhuqqmkjxsltzshfam.supabase.co",
  "tables": [
    {
      "name": "users",
      "records": 150,
      "size": 45678,
      "checksum": "sha256:abc123..."
    }
  ],
  "total_records": 1250,
  "backup_size": 2048576,
  "checksum": "sha256:def456...",
  "status": "completed",
  "environment": "production",
  "version": "1.0.0"
}
```

## 🔄 **Recovery Procedures**

### **Full Database Recovery**

#### **Step 1: Identify Backup**
```bash
# List available backups
./scripts/backup-database.sh list

# Select the appropriate backup
BACKUP_ID="backup_20250119_143022"
```

#### **Step 2: Verify Backup Integrity**
```bash
# Check backup metadata
cat backups/$BACKUP_ID/backup_metadata.json

# Verify file checksums
# (This would be implemented in the recovery tool)
```

#### **Step 3: Restore Data**
```bash
# Restore from backup
# (This would be implemented in the recovery tool)
go run cmd/restore/main.go -backup $BACKUP_ID
```

### **Partial Recovery**

#### **Single Table Recovery**
```bash
# Restore specific table
go run cmd/restore/main.go -backup $BACKUP_ID -table users
```

#### **Data Validation**
```bash
# Verify restored data
go run cmd/validate/main.go -backup $BACKUP_ID
```

## 📊 **Backup Monitoring**

### **Backup Status Tracking**

#### **Success Metrics**
- ✅ Backup completion rate: 100%
- ✅ Integrity verification: 100%
- ✅ Recovery time: < 30 minutes
- ✅ Data consistency: 100%

#### **Monitoring Commands**
```bash
# Check backup status
./scripts/backup-database.sh list

# Monitor backup health
go run cmd/backup/main.go -list
```

### **Alerting**

#### **Backup Failures**
- Email notification on backup failure
- Slack notification for critical issues
- Dashboard alerts for backup status

#### **Storage Monitoring**
- Disk space usage
- Backup retention compliance
- File integrity checks

## 🧹 **Backup Maintenance**

### **Retention Policy**

#### **Default Settings**
- **Retention Period**: 30 days
- **Cleanup Schedule**: Daily
- **Storage Limit**: 10GB per backup

#### **Cleanup Commands**
```bash
# Manual cleanup
./scripts/backup-database.sh cleanup

# Automated cleanup
go run cmd/backup/main.go -cleanup
```

### **Storage Optimization**

#### **Compression**
```bash
# Create compressed backup
go run cmd/backup/main.go -compress
```

#### **Incremental Backups**
- Future enhancement for large databases
- Delta backup implementation
- Efficient storage utilization

## 🔒 **Security Considerations**

### **Access Control**
- Backup files are stored locally
- No external access to backup data
- Secure file permissions (644)

### **Data Protection**
- Backup files contain sensitive data
- Ensure secure storage location
- Consider encryption for production

### **Audit Trail**
- All backup operations are logged
- Metadata includes timestamps and checksums
- Recovery operations are tracked

## 📋 **Backup Checklist**

### **Before Schema Changes**
- [ ] Verify Supabase connection
- [ ] Create full database backup
- [ ] Verify backup integrity
- [ ] Document backup location
- [ ] Test recovery procedure (optional)

### **After Schema Changes**
- [ ] Verify new schema integrity
- [ ] Test application functionality
- [ ] Create post-change backup
- [ ] Update documentation

### **Regular Maintenance**
- [ ] Monitor backup storage usage
- [ ] Clean up old backups
- [ ] Verify backup integrity
- [ ] Update backup procedures

## 🚨 **Troubleshooting**

### **Common Issues**

#### **Backup Failure**
```bash
# Check environment variables
env | grep SUPABASE

# Test connection
go run test-supabase-connection.go

# Check disk space
df -h
```

#### **Integrity Verification Failure**
```bash
# Re-verify backup
./scripts/backup-database.sh list

# Check file permissions
ls -la backups/backup_*/

# Recreate backup if necessary
./scripts/backup-database.sh backup
```

#### **Recovery Issues**
```bash
# Verify backup metadata
cat backups/backup_*/backup_metadata.json

# Check table files
ls -la backups/backup_*/*.json

# Contact support if needed
```

## 📞 **Support and Escalation**

### **Backup Issues**
1. **Level 1**: Check environment and connectivity
2. **Level 2**: Verify backup integrity and storage
3. **Level 3**: Escalate to database administrator

### **Recovery Issues**
1. **Level 1**: Verify backup files and metadata
2. **Level 2**: Test recovery procedures
3. **Level 3**: Contact Supabase support

## 📚 **Additional Resources**

### **Documentation**
- [Supabase Backup Documentation](https://supabase.com/docs/guides/platform/backups)
- [PostgreSQL Backup Best Practices](https://www.postgresql.org/docs/current/backup.html)

### **Tools**
- Backup CLI: `cmd/backup/main.go`
- Backup Script: `scripts/backup-database.sh`
- Connection Test: `test-supabase-connection.go`

### **Monitoring**
- Backup Status: `./scripts/backup-database.sh list`
- System Health: `go run cmd/backup/main.go -list`

---

**Document Version**: 1.0  
**Created**: January 19, 2025  
**Last Updated**: January 19, 2025  
**Next Review**: Weekly during implementation
