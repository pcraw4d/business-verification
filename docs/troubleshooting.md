# KYB Platform - Troubleshooting Guide

## Table of Contents

1. [Quick Diagnostics](#quick-diagnostics)
2. [Application Issues](#application-issues)
3. [Database Issues](#database-issues)
4. [Performance Issues](#performance-issues)
5. [Security Issues](#security-issues)
6. [Deployment Issues](#deployment-issues)
7. [Monitoring Issues](#monitoring-issues)
8. [API Issues](#api-issues)
9. [Common Error Messages](#common-error-messages)
10. [Emergency Procedures](#emergency-procedures)

## Quick Diagnostics

### Health Check Commands

```bash
# Application health
curl -f http://localhost:8080/health

# Database connectivity
psql -h localhost -p 5432 -U kyb_user -d kyb_platform -c "SELECT 1;"

# Redis connectivity
redis-cli ping

# System resources
free -h && df -h && top -n 1

# Network connectivity
netstat -tlnp | grep :8080
```

### Log Locations

```bash
# Application logs
journalctl -u kyb-platform -f
tail -f /var/log/kyb-platform/app.log

# Database logs
tail -f /var/log/postgresql/postgresql-*.log

# System logs
tail -f /var/log/syslog
dmesg | tail -20
```

## Application Issues

### Application Won't Start

**Symptoms**: Application fails to start or crashes immediately

**Diagnostic Steps**:
```bash
# Check configuration
./kyb-platform --config-check

# Check dependencies
systemctl status postgresql redis

# Check ports
netstat -tlnp | grep :8080

# Check permissions
ls -la /opt/kyb-platform/
```

**Common Solutions**:

1. **Port Already in Use**
   ```bash
   # Find process using port 8080
   lsof -i :8080
   
   # Kill process
   sudo kill -9 <PID>
   ```

2. **Configuration Error**
   ```bash
   # Validate environment variables
   env | grep KYB_
   
   # Check config file
   cat configs/production.env
   ```

3. **Permission Issues**
   ```bash
   # Fix permissions
   sudo chown -R kyb:kyb /opt/kyb-platform/
   sudo chmod +x /opt/kyb-platform/kyb-platform
   ```

### Application Crashes

**Symptoms**: Application starts but crashes periodically

**Diagnostic Steps**:
```bash
# Check memory usage
free -h
ps aux --sort=-%mem | head -10

# Check for memory leaks
go tool pprof http://localhost:8080/debug/pprof/heap

# Check system limits
ulimit -a
```

**Common Solutions**:

1. **Memory Issues**
   ```bash
   # Increase memory limits
   echo 'vm.max_map_count=262144' >> /etc/sysctl.conf
   sysctl -p
   
   # Restart with more memory
   docker run --memory=2g kyb-platform
   ```

2. **File Descriptor Limits**
   ```bash
   # Increase file descriptor limits
   echo '* soft nofile 65536' >> /etc/security/limits.conf
   echo '* hard nofile 65536' >> /etc/security/limits.conf
   ```

### High CPU Usage

**Symptoms**: Application consuming excessive CPU

**Diagnostic Steps**:
```bash
# Check CPU usage
top -p $(pgrep kyb-platform)

# Profile CPU usage
go tool pprof http://localhost:8080/debug/pprof/profile

# Check for goroutine leaks
curl http://localhost:8080/debug/pprof/goroutine
```

**Common Solutions**:

1. **Optimize Database Queries**
   ```sql
   -- Check slow queries
   SELECT query, mean_time, calls
   FROM pg_stat_statements
   ORDER BY mean_time DESC
   LIMIT 10;
   ```

2. **Reduce Concurrency**
   ```bash
   # Limit concurrent requests
   export KYB_MAX_CONCURRENT_REQUESTS=100
   ```

## Database Issues

### Connection Failures

**Symptoms**: Database connection errors

**Diagnostic Steps**:
```bash
# Test connection
psql -h localhost -p 5432 -U kyb_user -d kyb_platform

# Check PostgreSQL status
systemctl status postgresql

# Check connection pool
psql -h localhost -p 5432 -U kyb_user -d kyb_platform -c "
SELECT count(*) FROM pg_stat_activity;
"
```

**Common Solutions**:

1. **PostgreSQL Not Running**
   ```bash
   # Start PostgreSQL
   sudo systemctl start postgresql
   sudo systemctl enable postgresql
   ```

2. **Connection Pool Exhausted**
   ```sql
   -- Increase max connections
   ALTER SYSTEM SET max_connections = 200;
   SELECT pg_reload_conf();
   ```

3. **Authentication Issues**
   ```bash
   # Check pg_hba.conf
   sudo cat /etc/postgresql/*/main/pg_hba.conf
   
   # Reset password
   sudo -u postgres psql -c "ALTER USER kyb_user PASSWORD 'new_password';"
   ```

### Slow Queries

**Symptoms**: Database queries taking too long

**Diagnostic Steps**:
```sql
-- Check slow queries
SELECT query, mean_time, calls, total_time
FROM pg_stat_statements
ORDER BY mean_time DESC
LIMIT 10;

-- Check table statistics
SELECT schemaname, tablename, n_tup_ins, n_tup_upd, n_tup_del
FROM pg_stat_user_tables;
```

**Common Solutions**:

1. **Missing Indexes**
   ```sql
   -- Create indexes
   CREATE INDEX CONCURRENTLY idx_businesses_name ON businesses(name);
   CREATE INDEX CONCURRENTLY idx_classifications_business_id ON classifications(business_id);
   ```

2. **Update Statistics**
   ```sql
   -- Update table statistics
   ANALYZE businesses;
   ANALYZE classifications;
   ```

3. **Optimize Queries**
   ```sql
   -- Use EXPLAIN to analyze queries
   EXPLAIN (ANALYZE, BUFFERS) SELECT * FROM businesses WHERE name ILIKE '%acme%';
   ```

### Database Corruption

**Symptoms**: Data inconsistencies or corruption errors

**Diagnostic Steps**:
```bash
# Check database integrity
pg_dump -h localhost -p 5432 -U kyb_user -d kyb_platform --verbose > /dev/null

# Check for corrupted tables
psql -h localhost -p 5432 -U kyb_user -d kyb_platform -c "
SELECT schemaname, tablename, n_dead_tup, n_live_tup
FROM pg_stat_user_tables
WHERE n_dead_tup > n_live_tup * 0.1;
"
```

**Common Solutions**:

1. **Vacuum Database**
   ```sql
   -- Full vacuum
   VACUUM FULL businesses;
   VACUUM FULL classifications;
   ```

2. **Restore from Backup**
   ```bash
   # Restore from latest backup
   ./scripts/restore.sh /backups/kyb_platform_$(date +%Y%m%d).sql.gz
   ```

## Performance Issues

### Slow Response Times

**Symptoms**: API responses taking longer than expected

**Diagnostic Steps**:
```bash
# Check response times
curl -w "@curl-format.txt" -o /dev/null -s http://localhost:8080/health

# Check application metrics
curl -s http://localhost:8080/metrics | grep kyb_http_request_duration

# Check database performance
psql -h localhost -p 5432 -U kyb_user -d kyb_platform -c "
SELECT query, mean_time, calls
FROM pg_stat_statements
ORDER BY mean_time DESC
LIMIT 5;
"
```

**Common Solutions**:

1. **Optimize Database Queries**
   ```sql
   -- Add missing indexes
   CREATE INDEX CONCURRENTLY idx_businesses_created_at ON businesses(created_at);
   
   -- Optimize slow queries
   EXPLAIN (ANALYZE) SELECT * FROM businesses WHERE created_at > NOW() - INTERVAL '1 day';
   ```

2. **Enable Caching**
   ```bash
   # Check Redis cache hit ratio
   redis-cli info stats | grep keyspace_hits
   
   # Increase cache size
   redis-cli config set maxmemory 2gb
   ```

3. **Scale Application**
   ```bash
   # Increase replicas
   kubectl scale deployment kyb-platform --replicas=5
   ```

### High Memory Usage

**Symptoms**: Application consuming excessive memory

**Diagnostic Steps**:
```bash
# Check memory usage
free -h
ps aux --sort=-%mem | head -10

# Check Go memory stats
curl -s http://localhost:8080/debug/vars | jq '.memstats'

# Check for memory leaks
go tool pprof http://localhost:8080/debug/pprof/heap
```

**Common Solutions**:

1. **Optimize Memory Usage**
   ```go
   // Enable memory profiling
   import _ "net/http/pprof"
   
   // Add to main.go
   go func() {
       log.Println(http.ListenAndServe("localhost:6060", nil))
   }()
   ```

2. **Increase Memory Limits**
   ```bash
   # Docker
   docker run --memory=2g kyb-platform
   
   # Kubernetes
   kubectl patch deployment kyb-platform -p '{"spec":{"template":{"spec":{"containers":[{"name":"kyb-platform","resources":{"limits":{"memory":"2Gi"}}}]}}}}'
   ```

### High CPU Usage

**Symptoms**: Application consuming excessive CPU

**Diagnostic Steps**:
```bash
# Check CPU usage
top -p $(pgrep kyb-platform)

# Profile CPU usage
go tool pprof http://localhost:8080/debug/pprof/profile

# Check goroutines
curl http://localhost:8080/debug/pprof/goroutine
```

**Common Solutions**:

1. **Optimize Algorithms**
   ```go
   // Use connection pooling
   db.SetMaxOpenConns(25)
   db.SetMaxIdleConns(25)
   db.SetConnMaxLifetime(5 * time.Minute)
   ```

2. **Reduce Concurrency**
   ```bash
   # Limit concurrent requests
   export KYB_MAX_CONCURRENT_REQUESTS=50
   ```

## Security Issues

### Authentication Failures

**Symptoms**: Users unable to authenticate

**Diagnostic Steps**:
```bash
# Check JWT configuration
echo $JWT_SECRET
echo $JWT_EXPIRY

# Check user table
psql -h localhost -p 5432 -U kyb_user -d kyb_platform -c "
SELECT id, email, role, created_at FROM users LIMIT 5;
"

# Check authentication logs
tail -f /var/log/kyb-platform/auth.log
```

**Common Solutions**:

1. **Reset User Password**
   ```sql
   -- Reset password
   UPDATE users SET password_hash = crypt('new_password', gen_salt('bf')) WHERE email = 'user@example.com';
   ```

2. **Regenerate JWT Secret**
   ```bash
   # Generate new JWT secret
   openssl rand -base64 32
   
   # Update environment variable
   export JWT_SECRET="new_secret_key"
   ```

### Rate Limiting Issues

**Symptoms**: Legitimate requests being rate limited

**Diagnostic Steps**:
```bash
# Check rate limit configuration
echo $KYB_RATE_LIMIT_REQUESTS
echo $KYB_RATE_LIMIT_WINDOW

# Check Redis rate limit counters
redis-cli keys "*rate_limit*"

# Check rate limit logs
tail -f /var/log/kyb-platform/rate_limit.log
```

**Common Solutions**:

1. **Adjust Rate Limits**
   ```bash
   # Increase rate limits
   export KYB_RATE_LIMIT_REQUESTS=1000
   export KYB_RATE_LIMIT_WINDOW=1m
   ```

2. **Whitelist IP Addresses**
   ```bash
   # Add IP to whitelist
   redis-cli sadd "rate_limit_whitelist" "192.168.1.100"
   ```

## Deployment Issues

### Docker Issues

**Symptoms**: Docker containers failing to start or run

**Diagnostic Steps**:
```bash
# Check container status
docker ps -a

# Check container logs
docker logs kyb-platform

# Check Docker daemon
systemctl status docker

# Check disk space
df -h
```

**Common Solutions**:

1. **Container Won't Start**
   ```bash
   # Check image
   docker images kyb-platform
   
   # Rebuild image
   docker build -t kyb-platform .
   
   # Run with debug
   docker run --rm -it kyb-platform /bin/sh
   ```

2. **Port Conflicts**
   ```bash
   # Check port usage
   netstat -tlnp | grep :8080
   
   # Use different port
   docker run -p 8081:8080 kyb-platform
   ```

### Kubernetes Issues

**Symptoms**: Pods failing to start or services not accessible

**Diagnostic Steps**:
```bash
# Check pod status
kubectl get pods -n production

# Check pod logs
kubectl logs -f deployment/kyb-platform -n production

# Check service status
kubectl get svc -n production

# Check events
kubectl get events -n production --sort-by='.lastTimestamp'
```

**Common Solutions**:

1. **Pod CrashLoopBackOff**
   ```bash
   # Check pod description
   kubectl describe pod <pod-name> -n production
   
   # Check resource limits
   kubectl get pod <pod-name> -n production -o yaml | grep -A 10 resources
   ```

2. **Service Not Accessible**
   ```bash
   # Check service endpoints
   kubectl get endpoints kyb-platform -n production
   
   # Check ingress
   kubectl get ingress -n production
   ```

## Monitoring Issues

### Prometheus Issues

**Symptoms**: Metrics not being collected or displayed

**Diagnostic Steps**:
```bash
# Check Prometheus status
systemctl status prometheus

# Check targets
curl -s http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | {job: .labels.job, health: .health}'

# Check metrics endpoint
curl -s http://localhost:8080/metrics | head -20
```

**Common Solutions**:

1. **Targets Down**
   ```bash
   # Check network connectivity
   curl -f http://kyb-platform:8080/metrics
   
   # Update Prometheus config
   sudo systemctl reload prometheus
   ```

2. **Metrics Not Available**
   ```bash
   # Check application metrics
   curl -s http://localhost:8080/metrics | grep kyb_
   
   # Restart application
   systemctl restart kyb-platform
   ```

### Grafana Issues

**Symptoms**: Dashboards not loading or data not displayed

**Diagnostic Steps**:
```bash
# Check Grafana status
systemctl status grafana-server

# Check data sources
curl -s http://admin:admin@localhost:3000/api/datasources

# Check dashboard
curl -s http://admin:admin@localhost:3000/api/dashboards
```

**Common Solutions**:

1. **Data Source Issues**
   ```bash
   # Test Prometheus connection
   curl -f http://localhost:9090/api/v1/query?query=up
   
   # Update data source configuration
   # Access Grafana UI and update Prometheus data source
   ```

2. **Dashboard Not Loading**
   ```bash
   # Check dashboard permissions
   # Access Grafana UI and verify dashboard access
   
   # Import dashboard
   curl -X POST http://admin:admin@localhost:3000/api/dashboards/db \
     -H "Content-Type: application/json" \
     -d @kyb-platform-dashboard.json
   ```

## API Issues

### 500 Internal Server Errors

**Symptoms**: API returning 500 errors

**Diagnostic Steps**:
```bash
# Check application logs
tail -f /var/log/kyb-platform/app.log | grep ERROR

# Check error rate
curl -s http://localhost:8080/metrics | grep kyb_http_requests_total | grep status="500"

# Check database connectivity
psql -h localhost -p 5432 -U kyb_user -d kyb_platform -c "SELECT 1;"
```

**Common Solutions**:

1. **Database Connection Issues**
   ```bash
   # Restart database
   sudo systemctl restart postgresql
   
   # Check connection pool
   psql -h localhost -p 5432 -U kyb_user -d kyb_platform -c "
   SELECT count(*) FROM pg_stat_activity;
   "
   ```

2. **Memory Issues**
   ```bash
   # Check memory usage
   free -h
   
   # Restart application
   systemctl restart kyb-platform
   ```

### 404 Not Found Errors

**Symptoms**: API endpoints returning 404 errors

**Diagnostic Steps**:
```bash
# Check routing
curl -v http://localhost:8080/v1/classify

# Check application routes
curl -s http://localhost:8080/docs

# Check API documentation
curl -s http://localhost:8080/docs/openapi.yaml
```

**Common Solutions**:

1. **Incorrect URL**
   ```bash
   # Verify correct endpoint
   curl -v http://localhost:8080/v1/classify
   
   # Check API version
   curl -s http://localhost:8080/docs/openapi.yaml | grep -A 5 "/v1/"
   ```

2. **Route Not Registered**
   ```bash
   # Check application startup logs
   journalctl -u kyb-platform --since "10 minutes ago"
   
   # Restart application
   systemctl restart kyb-platform
   ```

## Common Error Messages

### Database Errors

**"connection refused"**
```bash
# PostgreSQL not running
sudo systemctl start postgresql
```

**"authentication failed"**
```bash
# Check credentials
psql -h localhost -p 5432 -U kyb_user -d kyb_platform
```

**"relation does not exist"**
```bash
# Run migrations
go run cmd/migrate/main.go up
```

### Application Errors

**"port already in use"**
```bash
# Find and kill process
lsof -i :8080
sudo kill -9 <PID>
```

**"permission denied"**
```bash
# Fix permissions
sudo chown -R kyb:kyb /opt/kyb-platform/
```

**"no such file or directory"**
```bash
# Check file exists
ls -la /opt/kyb-platform/kyb-platform
```

## Emergency Procedures

### Application Down

**Immediate Actions**:
```bash
# 1. Check application status
systemctl status kyb-platform

# 2. Check logs for errors
journalctl -u kyb-platform -f

# 3. Restart application
systemctl restart kyb-platform

# 4. Verify health
curl -f http://localhost:8080/health
```

### Database Down

**Immediate Actions**:
```bash
# 1. Check database status
systemctl status postgresql

# 2. Check disk space
df -h

# 3. Restart database
sudo systemctl restart postgresql

# 4. Verify connectivity
psql -h localhost -p 5432 -U kyb_user -d kyb_platform -c "SELECT 1;"
```

### Security Breach

**Immediate Actions**:
```bash
# 1. Stop application
systemctl stop kyb-platform

# 2. Check logs for suspicious activity
tail -f /var/log/kyb-platform/auth.log

# 3. Rotate secrets
openssl rand -base64 32 > new_jwt_secret

# 4. Restart with new secrets
export JWT_SECRET=$(cat new_jwt_secret)
systemctl start kyb-platform
```

### Data Loss

**Immediate Actions**:
```bash
# 1. Stop application
systemctl stop kyb-platform

# 2. Check backup availability
ls -la /backups/

# 3. Restore from backup
./scripts/restore.sh /backups/kyb_platform_latest.sql.gz

# 4. Verify data integrity
psql -h localhost -p 5432 -U kyb_user -d kyb_platform -c "
SELECT count(*) FROM businesses;
"

# 5. Restart application
systemctl start kyb-platform
```

---

## Conclusion

This troubleshooting guide provides comprehensive procedures for diagnosing and resolving common issues with the KYB Platform. Key points to remember:

- **Always check logs first** for error messages and clues
- **Use systematic diagnostic procedures** to isolate the root cause
- **Test solutions in staging** before applying to production
- **Document any custom solutions** for future reference
- **Have emergency procedures ready** for critical issues

For additional support or complex issues, please contact the development team with:
- Detailed error messages
- System logs and metrics
- Steps to reproduce the issue
- Environment details (OS, versions, configuration)
