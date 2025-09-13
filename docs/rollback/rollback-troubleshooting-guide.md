# Rollback Troubleshooting Guide
## KYB Platform - Merchant-Centric UI Implementation

**Document Version**: 1.0.0  
**Created**: January 2025  
**Status**: Production Ready  
**Target**: Comprehensive Troubleshooting for Rollback Operations

---

## Table of Contents

1. [Quick Reference](#quick-reference)
2. [Database Rollback Issues](#database-rollback-issues)
3. [Application Rollback Issues](#application-rollback-issues)
4. [Configuration Rollback Issues](#configuration-rollback-issues)
5. [System-Level Issues](#system-level-issues)
6. [Performance Issues](#performance-issues)
7. [Security Issues](#security-issues)
8. [Emergency Recovery](#emergency-recovery)
9. [Prevention Strategies](#prevention-strategies)
10. [Contact Information](#contact-information)

---

## Quick Reference

### Emergency Commands

```bash
# Emergency stop all rollback operations
pkill -f "rollback"

# Emergency database connection test
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT 1;"

# Emergency application status check
ps aux | grep kyb-platform

# Emergency log check
tail -f logs/*rollback*.log
```

### Common Error Codes

| Code | Issue | Quick Fix |
|------|-------|-----------|
| 1 | General error | Check logs, verify environment |
| 2 | Invalid arguments | Use `--help` for correct syntax |
| 3 | Database connection failed | Check DB credentials and connectivity |
| 4 | Backup file not found | Create backup or verify path |
| 5 | Permission denied | Check file/database permissions |
| 6 | Configuration validation failed | Validate config files |

---

## Database Rollback Issues

### Issue: Database Connection Failed

**Symptoms**:
- Error: "Failed to connect to database"
- Script exits with code 3
- Connection timeout errors

**Diagnosis**:
```bash
# Test database connectivity
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT 1;"

# Check environment variables
echo "DB_HOST: $DB_HOST"
echo "DB_PORT: $DB_PORT"
echo "DB_NAME: $DB_NAME"
echo "DB_USER: $DB_USER"

# Check network connectivity
ping $DB_HOST
telnet $DB_HOST $DB_PORT
```

**Solutions**:

1. **Verify Database Credentials**:
   ```bash
   # Check .env file
   cat .env | grep DB_
   
   # Verify credentials manually
   psql -h $DB_HOST -U $DB_USER -d $DB_NAME
   ```

2. **Check Database Service**:
   ```bash
   # Check if PostgreSQL is running
   sudo systemctl status postgresql
   
   # Start PostgreSQL if stopped
   sudo systemctl start postgresql
   ```

3. **Verify Network Access**:
   ```bash
   # Check firewall rules
   sudo ufw status
   
   # Check if port is open
   netstat -tlnp | grep $DB_PORT
   ```

### Issue: Backup File Not Found

**Symptoms**:
- Error: "Backup file not found"
- Script exits with code 4
- Missing backup files in expected location

**Diagnosis**:
```bash
# List available backups
ls -la backups/database/

# Check backup directory permissions
ls -ld backups/database/

# Search for backup files
find . -name "*backup*" -type f
```

**Solutions**:

1. **Create Missing Backup**:
   ```bash
   # Create database backup
   ./scripts/rollback/database-rollback.sh --backup schema
   
   # Manual backup creation
   pg_dump -h $DB_HOST -U $DB_USER -d $DB_NAME > backups/database/manual-backup.sql
   ```

2. **Verify Backup Path**:
   ```bash
   # Check backup directory structure
   tree backups/
   
   # Create backup directory if missing
   mkdir -p backups/database/
   ```

3. **Restore from Alternative Location**:
   ```bash
   # Copy backup from alternative location
   cp /path/to/alternative/backup.sql backups/database/
   
   # Update script to use correct path
   export BACKUP_DIR="/path/to/alternative/backups"
   ```

### Issue: Schema Rollback Failed

**Symptoms**:
- Error: "Schema rollback failed"
- Migration errors
- Database constraint violations

**Diagnosis**:
```bash
# Check current database schema
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "\dt"

# Check migration history
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "SELECT * FROM schema_migrations;"

# Check for constraint violations
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "SELECT * FROM pg_constraint WHERE NOT convalidated;"
```

**Solutions**:

1. **Manual Schema Rollback**:
   ```bash
   # Connect to database
   psql -h $DB_HOST -U $DB_USER -d $DB_NAME
   
   # Drop problematic tables/columns
   DROP TABLE IF EXISTS problematic_table;
   
   # Recreate from backup
   \i backups/database/schema-backup.sql
   ```

2. **Fix Constraint Issues**:
   ```bash
   # Disable constraints temporarily
   SET session_replication_role = replica;
   
   # Perform rollback
   \i backups/database/data-backup.sql
   
   # Re-enable constraints
   SET session_replication_role = DEFAULT;
   ```

### Issue: Data Rollback Failed

**Symptoms**:
- Error: "Data rollback failed"
- Data integrity issues
- Foreign key constraint violations

**Diagnosis**:
```bash
# Check data integrity
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "SELECT COUNT(*) FROM merchants;"

# Check foreign key constraints
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "SELECT * FROM information_schema.table_constraints WHERE constraint_type = 'FOREIGN KEY';"

# Check for orphaned records
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "SELECT * FROM audit_logs WHERE merchant_id NOT IN (SELECT id FROM merchants);"
```

**Solutions**:

1. **Clean Data Rollback**:
   ```bash
   # Truncate tables in correct order
   psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "TRUNCATE audit_logs, compliance_records, merchants CASCADE;"
   
   # Restore data from backup
   psql -h $DB_HOST -U $DB_USER -d $DB_NAME < backups/database/data-backup.sql
   ```

2. **Incremental Data Rollback**:
   ```bash
   # Restore data table by table
   psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "DELETE FROM audit_logs;"
   psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "DELETE FROM compliance_records;"
   psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "DELETE FROM merchants;"
   
   # Restore each table
   psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "\copy merchants FROM 'backups/database/merchants.csv' CSV HEADER;"
   ```

---

## Application Rollback Issues

### Issue: Application Binary Not Found

**Symptoms**:
- Error: "Application binary not found"
- Script exits during binary rollback
- Missing executable files

**Diagnosis**:
```bash
# Check for application binaries
find . -name "kyb-platform*" -type f

# Check backup directory
ls -la backups/application/

# Check if binary is executable
ls -la kyb-platform
```

**Solutions**:

1. **Rebuild Application**:
   ```bash
   # Build application from source
   go build -o kyb-platform ./cmd/server
   
   # Make executable
   chmod +x kyb-platform
   ```

2. **Restore from Backup**:
   ```bash
   # Extract application backup
   tar -xzf backups/application/app-backup-v1.2.3.tar.gz
   
   # Make executable
   chmod +x kyb-platform
   ```

3. **Download from Repository**:
   ```bash
   # Download specific version
   wget https://releases.kyb-platform.com/v1.2.3/kyb-platform
   
   # Make executable
   chmod +x kyb-platform
   ```

### Issue: Application Won't Start

**Symptoms**:
- Error: "Application failed to start"
- Process exits immediately
- Port already in use errors

**Diagnosis**:
```bash
# Check if port is in use
netstat -tlnp | grep 8080

# Check application logs
tail -f logs/app.log

# Test application manually
./kyb-platform --help
```

**Solutions**:

1. **Kill Existing Process**:
   ```bash
   # Find and kill existing process
   pkill -f kyb-platform
   
   # Kill process on specific port
   lsof -ti:8080 | xargs kill -9
   ```

2. **Check Configuration**:
   ```bash
   # Validate configuration files
   yq eval '.' configs/database.yaml
   yq eval '.' configs/api.yaml
   
   # Test with minimal configuration
   ./kyb-platform --config configs/minimal.yaml
   ```

3. **Check Dependencies**:
   ```bash
   # Check if database is accessible
   psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "SELECT 1;"
   
   # Check if required files exist
   ls -la configs/
   ```

### Issue: Docker Rollback Failed

**Symptoms**:
- Error: "Docker rollback failed"
- Container won't start
- Image pull failures

**Diagnosis**:
```bash
# Check Docker status
docker --version
docker ps -a

# Check available images
docker images | grep kyb-platform

# Check Docker Compose
docker-compose --version
```

**Solutions**:

1. **Fix Docker Issues**:
   ```bash
   # Restart Docker service
   sudo systemctl restart docker
   
   # Clean up Docker resources
   docker system prune -f
   ```

2. **Rebuild Docker Image**:
   ```bash
   # Build Docker image
   docker build -t kyb-platform:v1.2.3 .
   
   # Tag for rollback
   docker tag kyb-platform:v1.2.3 kyb-platform:rollback
   ```

3. **Fix Docker Compose**:
   ```bash
   # Stop all containers
   docker-compose down
   
   # Start with specific version
   docker-compose -f docker-compose.production.yml up -d
   ```

---

## Configuration Rollback Issues

### Issue: Configuration File Invalid

**Symptoms**:
- Error: "Configuration validation failed"
- YAML/JSON parsing errors
- Script exits with code 6

**Diagnosis**:
```bash
# Validate YAML files
yq eval '.' configs/database.yaml
yq eval '.' configs/api.yaml
yq eval '.' configs/security.yaml

# Validate JSON files
jq empty configs/features.json

# Check file permissions
ls -la configs/
```

**Solutions**:

1. **Fix YAML Syntax**:
   ```bash
   # Use yq to format YAML
   yq eval '.' configs/database.yaml > configs/database.yaml.tmp
   mv configs/database.yaml.tmp configs/database.yaml
   ```

2. **Fix JSON Syntax**:
   ```bash
   # Use jq to format JSON
   jq '.' configs/features.json > configs/features.json.tmp
   mv configs/features.json.tmp configs/features.json
   ```

3. **Restore from Backup**:
   ```bash
   # Extract configuration backup
   tar -xzf backups/configuration/config-backup-v1.2.3.tar.gz
   
   # Verify restored files
   ls -la configs/
   ```

### Issue: Environment Variables Missing

**Symptoms**:
- Error: "Environment variable not set"
- Configuration loading failures
- Default values being used

**Diagnosis**:
```bash
# Check environment variables
env | grep -E "(DB_|API_|LOG_)"

# Check .env file
cat .env

# Check shell environment
echo $SHELL
```

**Solutions**:

1. **Set Environment Variables**:
   ```bash
   # Source .env file
   source .env
   
   # Set variables manually
   export DB_HOST=localhost
   export DB_PORT=5432
   export DB_NAME=kyb_platform
   ```

2. **Fix .env File**:
   ```bash
   # Create .env file
   cat > .env << EOF
   DB_HOST=localhost
   DB_PORT=5432
   DB_NAME=kyb_platform
   DB_USER=postgres
   DB_PASSWORD=password
   EOF
   ```

3. **Use Environment-Specific Config**:
   ```bash
   # Copy environment-specific config
   cp configs/production.env .env
   
   # Source the file
   source .env
   ```

### Issue: Feature Flags Not Applied

**Symptoms**:
- Features not working after rollback
- Configuration not taking effect
- Application using old feature flags

**Diagnosis**:
```bash
# Check feature flags file
cat configs/features.json

# Check application logs for feature flag loading
grep -i "feature" logs/app.log

# Test feature flag API
curl -s http://localhost:8080/api/features | jq .
```

**Solutions**:

1. **Restart Application**:
   ```bash
   # Restart application to reload configuration
   sudo systemctl restart kyb-platform
   
   # Or kill and restart manually
   pkill -f kyb-platform
   nohup ./kyb-platform > logs/app.log 2>&1 &
   ```

2. **Verify Feature Flags**:
   ```bash
   # Check feature flags are loaded
   curl -s http://localhost:8080/api/features
   
   # Update feature flags
   jq '.features.merchant_portfolio = true' configs/features.json > configs/features.json.tmp
   mv configs/features.json.tmp configs/features.json
   ```

3. **Clear Application Cache**:
   ```bash
   # Clear any application caches
   rm -rf cache/
   
   # Restart application
   sudo systemctl restart kyb-platform
   ```

---

## System-Level Issues

### Issue: Insufficient Disk Space

**Symptoms**:
- Error: "No space left on device"
- Backup creation fails
- Rollback operations fail

**Diagnosis**:
```bash
# Check disk usage
df -h

# Check directory sizes
du -sh backups/ logs/ configs/

# Check for large files
find . -type f -size +100M
```

**Solutions**:

1. **Clean Up Old Files**:
   ```bash
   # Remove old log files
   find logs/ -name "*.log" -mtime +30 -delete
   
   # Remove old backups
   find backups/ -name "*.sql" -mtime +7 -delete
   ```

2. **Compress Files**:
   ```bash
   # Compress old backups
   gzip backups/database/*.sql
   
   # Compress old logs
   gzip logs/*.log
   ```

3. **Move to External Storage**:
   ```bash
   # Move backups to external storage
   rsync -av backups/ /external/backups/
   
   # Create symlink
   ln -sf /external/backups backups
   ```

### Issue: Permission Denied

**Symptoms**:
- Error: "Permission denied"
- Script exits with code 5
- Cannot access files or directories

**Diagnosis**:
```bash
# Check file permissions
ls -la scripts/rollback/

# Check directory permissions
ls -ld backups/ logs/ configs/

# Check user and group
id
groups
```

**Solutions**:

1. **Fix File Permissions**:
   ```bash
   # Make scripts executable
   chmod +x scripts/rollback/*.sh
   
   # Fix directory permissions
   chmod 755 backups/ logs/ configs/
   ```

2. **Fix Ownership**:
   ```bash
   # Change ownership
   sudo chown -R $USER:$USER scripts/ backups/ logs/ configs/
   
   # Or change to specific user
   sudo chown -R kyb:kyb scripts/ backups/ logs/ configs/
   ```

3. **Use Sudo**:
   ```bash
   # Run with sudo if necessary
   sudo ./scripts/rollback/database-rollback.sh --dry-run schema
   ```

### Issue: Network Connectivity Problems

**Symptoms**:
- Error: "Connection refused"
- Timeout errors
- Cannot reach external services

**Diagnosis**:
```bash
# Test network connectivity
ping google.com

# Test specific host
ping $DB_HOST

# Test port connectivity
telnet $DB_HOST $DB_PORT

# Check DNS resolution
nslookup $DB_HOST
```

**Solutions**:

1. **Check Network Configuration**:
   ```bash
   # Check network interfaces
   ip addr show
   
   # Check routing table
   ip route show
   
   # Check DNS configuration
   cat /etc/resolv.conf
   ```

2. **Fix Firewall Rules**:
   ```bash
   # Check firewall status
   sudo ufw status
   
   # Allow specific ports
   sudo ufw allow 5432
   sudo ufw allow 8080
   ```

3. **Use Alternative Network**:
   ```bash
   # Use different network interface
   export DB_HOST=192.168.1.100
   
   # Use VPN if available
   sudo openvpn --config /path/to/vpn.conf
   ```

---

## Performance Issues

### Issue: Slow Rollback Operations

**Symptoms**:
- Rollback takes too long
- Timeout errors
- System becomes unresponsive

**Diagnosis**:
```bash
# Check system resources
top
htop

# Check disk I/O
iostat -x 1

# Check memory usage
free -h

# Check rollback logs for timing
grep -i "time\|duration" logs/*rollback*.log
```

**Solutions**:

1. **Optimize Database Operations**:
   ```bash
   # Use parallel operations
   pg_dump -j 4 -h $DB_HOST -U $DB_USER -d $DB_NAME > backup.sql
   
   # Use compression
   pg_dump -Z 9 -h $DB_HOST -U $DB_USER -d $DB_NAME > backup.sql.gz
   ```

2. **Optimize File Operations**:
   ```bash
   # Use faster compression
   tar --use-compress-program=pigz -czf backup.tar.gz files/
   
   # Use parallel compression
   pigz -p 4 backup.sql
   ```

3. **Increase System Resources**:
   ```bash
   # Increase memory limits
   ulimit -m unlimited
   
   # Increase file descriptor limits
   ulimit -n 65536
   ```

### Issue: High Memory Usage

**Symptoms**:
- Out of memory errors
- System swapping
- Rollback operations fail

**Diagnosis**:
```bash
# Check memory usage
free -h

# Check swap usage
swapon -s

# Check process memory usage
ps aux --sort=-%mem | head -10
```

**Solutions**:

1. **Optimize Memory Usage**:
   ```bash
   # Use streaming operations
   pg_dump -h $DB_HOST -U $DB_USER -d $DB_NAME | gzip > backup.sql.gz
   
   # Process data in chunks
   psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "COPY (SELECT * FROM merchants LIMIT 1000) TO STDOUT;"
   ```

2. **Increase Swap Space**:
   ```bash
   # Create swap file
   sudo fallocate -l 2G /swapfile
   sudo chmod 600 /swapfile
   sudo mkswap /swapfile
   sudo swapon /swapfile
   ```

3. **Kill Memory-Intensive Processes**:
   ```bash
   # Find memory-intensive processes
   ps aux --sort=-%mem | head -10
   
   # Kill specific processes
   kill -9 <PID>
   ```

---

## Security Issues

### Issue: Unauthorized Access

**Symptoms**:
- Permission denied errors
- Authentication failures
- Access control violations

**Diagnosis**:
```bash
# Check file permissions
ls -la scripts/rollback/

# Check user permissions
id
groups

# Check database permissions
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "\du"
```

**Solutions**:

1. **Fix File Permissions**:
   ```bash
   # Set secure permissions
   chmod 750 scripts/rollback/
   chmod 640 configs/*.yaml
   chmod 600 .env
   ```

2. **Fix Database Permissions**:
   ```bash
   # Grant necessary permissions
   psql -h $DB_HOST -U postgres -c "GRANT CONNECT ON DATABASE kyb_platform TO kyb_user;"
   psql -h $DB_HOST -U postgres -c "GRANT USAGE ON SCHEMA public TO kyb_user;"
   ```

3. **Use Service Account**:
   ```bash
   # Create service account
   sudo useradd -r -s /bin/false kyb-service
   
   # Run as service account
   sudo -u kyb-service ./scripts/rollback/database-rollback.sh --dry-run schema
   ```

### Issue: Sensitive Data Exposure

**Symptoms**:
- Sensitive data in logs
- Unencrypted backups
- Credentials in process list

**Diagnosis**:
```bash
# Check for sensitive data in logs
grep -i "password\|secret\|key" logs/*.log

# Check backup file permissions
ls -la backups/

# Check process environment
ps aux | grep -E "(password|secret|key)"
```

**Solutions**:

1. **Encrypt Backups**:
   ```bash
   # Encrypt backup files
   gpg --symmetric --cipher-algo AES256 backup.sql
   
   # Use encrypted storage
   mount -t ecryptfs /path/to/encrypted/backups /backups
   ```

2. **Sanitize Logs**:
   ```bash
   # Remove sensitive data from logs
   sed -i 's/password=[^[:space:]]*/password=***/g' logs/*.log
   
   # Use log sanitization in scripts
   export LOG_SANITIZE=true
   ```

3. **Secure Environment Variables**:
   ```bash
   # Use secure environment file
   chmod 600 .env
   
   # Use environment variable encryption
   export ENCRYPTED_DB_PASSWORD=$(echo "password" | openssl enc -aes-256-cbc -base64)
   ```

---

## Emergency Recovery

### Critical System Failure

**Symptoms**:
- Complete system failure
- Database corruption
- Application won't start

**Emergency Steps**:

1. **Assess Situation**:
   ```bash
   # Check system status
   systemctl status postgresql
   systemctl status kyb-platform
   
   # Check disk space
   df -h
   
   # Check memory
   free -h
   ```

2. **Stop All Services**:
   ```bash
   # Stop application
   sudo systemctl stop kyb-platform
   
   # Stop database
   sudo systemctl stop postgresql
   ```

3. **Emergency Backup**:
   ```bash
   # Create emergency backup
   sudo -u postgres pg_dump kyb_platform > emergency-backup-$(date +%Y%m%d-%H%M%S).sql
   
   # Copy to safe location
   cp emergency-backup-*.sql /external/backups/
   ```

4. **Emergency Rollback**:
   ```bash
   # Emergency database rollback
   sudo -u postgres psql kyb_platform < backups/database/stable-backup.sql
   
   # Emergency application rollback
   cp backups/application/stable-kyb-platform kyb-platform
   chmod +x kyb-platform
   ```

5. **Restart Services**:
   ```bash
   # Start database
   sudo systemctl start postgresql
   
   # Start application
   sudo systemctl start kyb-platform
   ```

### Data Corruption

**Symptoms**:
- Database errors
- Data inconsistency
- Application crashes

**Recovery Steps**:

1. **Stop Application**:
   ```bash
   sudo systemctl stop kyb-platform
   ```

2. **Check Database Integrity**:
   ```bash
   # Check database integrity
   sudo -u postgres psql kyb_platform -c "VACUUM ANALYZE;"
   
   # Check for corruption
   sudo -u postgres psql kyb_platform -c "SELECT * FROM pg_stat_database WHERE datname = 'kyb_platform';"
   ```

3. **Restore from Backup**:
   ```bash
   # Drop and recreate database
   sudo -u postgres dropdb kyb_platform
   sudo -u postgres createdb kyb_platform
   
   # Restore from backup
   sudo -u postgres psql kyb_platform < backups/database/latest-backup.sql
   ```

4. **Verify Data**:
   ```bash
   # Check data integrity
   sudo -u postgres psql kyb_platform -c "SELECT COUNT(*) FROM merchants;"
   sudo -u postgres psql kyb_platform -c "SELECT COUNT(*) FROM audit_logs;"
   ```

5. **Restart Application**:
   ```bash
   sudo systemctl start kyb-platform
   ```

---

## Prevention Strategies

### Regular Maintenance

1. **Daily Checks**:
   ```bash
   # Check system health
   ./scripts/health-check.sh
   
   # Check backup status
   ./scripts/backup-status.sh
   
   # Check log files
   ./scripts/log-rotation.sh
   ```

2. **Weekly Maintenance**:
   ```bash
   # Test rollback procedures
   ./scripts/rollback/database-rollback.sh --dry-run schema
   
   # Clean up old files
   ./scripts/cleanup.sh
   
   # Update documentation
   ./scripts/update-docs.sh
   ```

3. **Monthly Reviews**:
   ```bash
   # Review rollback procedures
   ./scripts/review-procedures.sh
   
   # Test disaster recovery
   ./scripts/disaster-recovery-test.sh
   
   # Update backup strategies
   ./scripts/update-backup-strategy.sh
   ```

### Monitoring and Alerting

1. **Set Up Monitoring**:
   ```bash
   # Monitor disk space
   df -h | awk '$5 > 80 {print $0}'
   
   # Monitor memory usage
   free | awk 'NR==2{printf "%.2f%%\n", $3*100/$2}'
   
   # Monitor database connections
   psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "SELECT count(*) FROM pg_stat_activity;"
   ```

2. **Set Up Alerting**:
   ```bash
   # Disk space alerts
   if [ $(df / | awk 'NR==2 {print $5}' | sed 's/%//') -gt 80 ]; then
       echo "Disk space warning" | mail -s "Disk Space Alert" admin@company.com
   fi
   
   # Memory usage alerts
   if [ $(free | awk 'NR==2{printf "%.0f", $3*100/$2}') -gt 80 ]; then
       echo "Memory usage warning" | mail -s "Memory Alert" admin@company.com
   fi
   ```

### Documentation and Training

1. **Keep Documentation Updated**:
   - Update rollback procedures regularly
   - Document new issues and solutions
   - Maintain troubleshooting guides

2. **Regular Training**:
   - Train team members on rollback procedures
   - Conduct disaster recovery drills
   - Review and update emergency contacts

---

## Contact Information

### Primary Contacts

- **DevOps Team Lead**: devops-lead@company.com
- **Platform Engineering**: platform-eng@company.com
- **Database Administrator**: dba@company.com

### Emergency Contacts

- **On-Call Engineer**: +1-555-ONCALL
- **Engineering Manager**: +1-555-ENG-MGR
- **CTO**: +1-555-CTO

### External Support

- **PostgreSQL Support**: https://www.postgresql.org/support/
- **Docker Support**: https://www.docker.com/support/
- **Cloud Provider Support**: https://cloud-provider.com/support

---

**Document Version**: 1.0.0  
**Last Updated**: January 19, 2025  
**Next Review**: April 19, 2025
