# Production Deployment Checklist

## Pre-Deployment Checklist

### 1. Code & Testing ✅
- [ ] All code changes committed and pushed
- [ ] All tests passing locally
- [ ] Integration tests completed
- [ ] No known critical bugs
- [ ] Code review completed (if applicable)

### 2. Database Setup ✅
- [ ] Database connection string configured (`DATABASE_URL`)
- [ ] Database schema verified (`risk_thresholds` table exists)
- [ ] Migration script tested
- [ ] Database connection tested (`./scripts/test_database_connection.sh`)
- [ ] Schema verification passed (`./scripts/verify_database_schema.sh`)

### 3. Environment Configuration ✅
- [ ] All required environment variables set
- [ ] Configuration validated (`./scripts/validate_config.sh`)
- [ ] Supabase credentials configured
- [ ] Optional services configured (Redis, if needed)
- [ ] Port configuration verified

### 4. Security Review ✅
- [ ] No secrets in code or committed files
- [ ] Environment variables stored securely
- [ ] API keys rotated (if needed)
- [ ] Authentication/authorization reviewed
- [ ] Rate limiting configured (if applicable)

### 5. Monitoring & Logging ✅
- [ ] Health check endpoint accessible (`/health/detailed`)
- [ ] Logging configured and tested
- [ ] Request ID tracking verified
- [ ] Error tracking configured (if applicable)
- [ ] Monitoring alerts set up (if applicable)

### 6. Documentation ✅
- [ ] API documentation updated
- [ ] Deployment procedures documented
- [ ] Environment setup guide reviewed
- [ ] Troubleshooting guide available

---

## Deployment Steps

### Step 1: Pre-Deployment Verification

Run the pre-deployment verification script:

```bash
./scripts/pre_deployment_check.sh
```

This will verify:
- ✅ Configuration is valid
- ✅ Database is accessible
- ✅ Schema is correct
- ✅ All required services are available

### Step 2: Backup Current State

**If deploying to existing production:**

```bash
# Create backup branch
git branch backup-$(date +%Y%m%d-%H%M%S)

# Create tag
git tag production-$(date +%Y%m%d-%H%M%S)

# Push backup
git push origin backup-$(date +%Y%m%d-%H%M%S)
git push origin production-$(date +%Y%m%d-%H%M%S)
```

### Step 3: Deploy to Staging (Recommended)

**Always deploy to staging first:**

1. Deploy code to staging environment
2. Configure staging environment variables
3. Run smoke tests:
   ```bash
   ./test/restoration_tests.sh
   ```
4. Monitor logs for errors
5. Verify all endpoints work
6. Test database persistence

### Step 4: Deploy to Production

**After staging verification:**

1. Deploy code to production
2. Verify environment variables are set
3. Run post-deployment verification:
   ```bash
   ./scripts/post_deployment_verify.sh
   ```
4. Monitor health endpoint:
   ```bash
   curl https://your-api.com/health/detailed
   ```
5. Run smoke tests on production:
   ```bash
   # Update API_URL in test scripts
   export API_URL="https://your-api.com"
   ./test/restoration_tests.sh
   ```

---

## Post-Deployment Verification

### Immediate Checks (First 5 minutes)

- [ ] Server starts without errors
- [ ] Health endpoint returns 200 OK
- [ ] Database connection successful
- [ ] All endpoints accessible
- [ ] No error spikes in logs

### Functional Checks (First 15 minutes)

- [ ] Create threshold endpoint works
- [ ] Get thresholds endpoint works
- [ ] Update threshold endpoint works
- [ ] Delete threshold endpoint works
- [ ] Export/import endpoints work
- [ ] Database persistence verified

### Monitoring (First 24 hours)

- [ ] Monitor error rates
- [ ] Monitor response times
- [ ] Monitor database connection pool
- [ ] Check for memory leaks
- [ ] Verify request IDs in logs
- [ ] Monitor threshold operations

---

## Rollback Procedure

If issues are detected:

### Immediate Rollback

1. **Revert to previous deployment:**
   ```bash
   # If using git tags
   git checkout production-YYYYMMDD-HHMMSS
   git push origin main --force
   ```

2. **Or restore from backup:**
   ```bash
   git checkout backup-YYYYMMDD-HHMMSS
   git push origin main --force
   ```

### Partial Rollback (Feature Flags)

If using feature flags:
1. Disable problematic feature
2. Monitor system recovery
3. Investigate issue
4. Fix and redeploy

### Database Rollback

If database issues:
1. Restore from database backup
2. Verify data integrity
3. Re-run migrations if needed

---

## Common Issues & Solutions

### Issue: "Database connection failed"

**Symptoms**: Server starts but shows database warning  
**Solutions**:
1. Verify `DATABASE_URL` is set correctly
2. Check database is accessible
3. Verify credentials
4. Check firewall rules

### Issue: "Table does not exist"

**Symptoms**: Endpoints return 500 errors  
**Solutions**:
1. Run migration: `psql $DATABASE_URL -f internal/database/migrations/012_create_risk_thresholds_table.sql`
2. Verify schema: `./scripts/verify_database_schema.sh`

### Issue: "Health check fails"

**Symptoms**: `/health/detailed` returns errors  
**Solutions**:
1. Check all required services are running
2. Verify environment variables
3. Check logs for specific errors
4. Verify database connection

### Issue: "Endpoints not accessible"

**Symptoms**: 404 or connection refused  
**Solutions**:
1. Verify server is running
2. Check port configuration
3. Verify routing is correct
4. Check firewall/security groups

---

## Success Criteria

Deployment is successful when:

- ✅ All pre-deployment checks pass
- ✅ Server starts without errors
- ✅ Health endpoint returns healthy status
- ✅ All endpoints are accessible
- ✅ Database persistence works
- ✅ No error spikes in logs
- ✅ Response times are acceptable
- ✅ Monitoring shows healthy metrics

---

## Emergency Contacts

- **On-Call Engineer**: [Contact Info]
- **Database Admin**: [Contact Info]
- **DevOps Team**: [Contact Info]

---

## Post-Deployment Tasks

After successful deployment:

1. **Document deployment**:
   - Record deployment time
   - Note any issues encountered
   - Document configuration changes

2. **Monitor closely**:
   - Watch logs for first 24 hours
   - Monitor error rates
   - Check performance metrics

3. **Gather feedback**:
   - Collect user feedback
   - Monitor API usage
   - Track error patterns

4. **Plan next steps**:
   - Address any issues found
   - Plan optimizations
   - Schedule next deployment

---

**Last Updated**: November 15, 2025  
**Version**: 1.0

