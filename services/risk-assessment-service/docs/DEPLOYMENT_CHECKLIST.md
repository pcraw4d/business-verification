# Railway Deployment Checklist

This checklist ensures a successful deployment of the Risk Assessment Service to Railway with proper configuration and monitoring.

## Pre-Deployment Checklist

### ✅ Prerequisites

- [ ] Railway CLI installed and authenticated
- [ ] Go 1.22+ installed locally
- [ ] Docker installed (for local testing)
- [ ] Railway account created and verified
- [ ] Supabase project set up with required tables
- [ ] External API keys obtained (NewsAPI, OpenCorporates - optional)

### ✅ Code Quality

- [ ] All tests pass locally (`make test`)
- [ ] Code coverage meets requirements (`make coverage`)
- [ ] Linting passes (`make lint`)
- [ ] Security scan passes (`make security`)
- [ ] Load tests pass locally (`make load-test`)

### ✅ Configuration

- [ ] `railway.json` configured correctly
- [ ] `Dockerfile` optimized for production
- [ ] Environment variables documented
- [ ] Performance targets set (1000 req/min)
- [ ] Monitoring endpoints configured

## Deployment Checklist

### ✅ Railway Setup

- [ ] Railway project created or linked
- [ ] Service added to Railway project
- [ ] Environment variables set in Railway
- [ ] Supabase credentials configured
- [ ] External API keys configured (if applicable)

### ✅ Environment Variables

#### Required Variables
- [ ] `SUPABASE_URL` - Supabase project URL
- [ ] `SUPABASE_ANON_KEY` - Supabase anonymous key
- [ ] `SUPABASE_SERVICE_ROLE_KEY` - Supabase service role key
- [ ] `ENV` - Environment (production/staging)
- [ ] `LOG_LEVEL` - Logging level (info/debug)

#### Performance Monitoring
- [ ] `PERFORMANCE_MONITORING_ENABLED=true`
- [ ] `PERFORMANCE_TARGET_RPS=16.67`
- [ ] `PERFORMANCE_TARGET_LATENCY=1s`
- [ ] `PERFORMANCE_TARGET_ERROR_RATE=0.01`
- [ ] `PERFORMANCE_TARGET_THROUGHPUT=1000`

#### Optional External APIs
- [ ] `NEWS_API_KEY` - NewsAPI key (if using)
- [ ] `OPEN_CORPORATES_API_KEY` - OpenCorporates key (if using)
- [ ] `GOVERNMENT_API_KEY` - Government API key (if using)

### ✅ Deployment Process

- [ ] Code committed to version control
- [ ] Railway deployment initiated
- [ ] Build process completed successfully
- [ ] Service started without errors
- [ ] Health checks passing

## Post-Deployment Checklist

### ✅ Service Health

- [ ] Basic health check passes (`/health`)
- [ ] Performance health check passes (`/api/v1/performance/health`)
- [ ] All monitoring endpoints responding
- [ ] No critical errors in logs
- [ ] Service responding within target latency

### ✅ Performance Validation

- [ ] Load test passes (1000 req/min target)
- [ ] Response time < 1 second (95th percentile)
- [ ] Error rate < 1%
- [ ] Memory usage < 512MB
- [ ] CPU usage < 80%

### ✅ Monitoring Setup

- [ ] Performance metrics endpoint working
- [ ] Alert system configured
- [ ] Log aggregation working
- [ ] Health monitoring active
- [ ] Performance targets being tracked

### ✅ Integration Testing

- [ ] Supabase connection working
- [ ] External API integrations working (if enabled)
- [ ] ML model loading correctly
- [ ] Cache system functioning
- [ ] Rate limiting working

## Verification Commands

### Health Checks

```bash
# Basic health check
curl https://your-service.railway.app/health

# Performance health check
curl https://your-service.railway.app/api/v1/performance/health

# Performance statistics
curl https://your-service.railway.app/api/v1/performance/stats
```

### Load Testing

```bash
# Quick load test
go run ./cmd/load_test.go \
  -url=https://your-service.railway.app \
  -duration=2m \
  -users=10 \
  -rps=16.67

# Comprehensive load test
./scripts/run_load_tests.sh
```

### Railway Commands

```bash
# Check deployment status
railway status

# View logs
railway logs --tail 100

# Check environment variables
railway variables

# Get service URL
railway domain
```

## Troubleshooting Checklist

### ❌ Build Failures

- [ ] Check Railway build logs
- [ ] Verify Go module dependencies
- [ ] Test local build
- [ ] Check Dockerfile syntax
- [ ] Verify file permissions

### ❌ Runtime Errors

- [ ] Check application logs
- [ ] Verify environment variables
- [ ] Test database connectivity
- [ ] Check external API connections
- [ ] Verify resource limits

### ❌ Performance Issues

- [ ] Check performance metrics
- [ ] Review performance alerts
- [ ] Run load tests
- [ ] Monitor resource usage
- [ ] Check rate limiting settings

### ❌ Health Check Failures

- [ ] Verify health endpoint
- [ ] Check service startup logs
- [ ] Test database connections
- [ ] Verify external dependencies
- [ ] Check network connectivity

## Success Criteria

### ✅ Deployment Success

- [ ] Service deployed without errors
- [ ] All health checks passing
- [ ] Performance targets met
- [ ] Monitoring working correctly
- [ ] No critical alerts

### ✅ Performance Success

- [ ] 1000+ requests/minute sustained
- [ ] < 1 second response time
- [ ] < 1% error rate
- [ ] Stable resource usage
- [ ] No memory leaks

### ✅ Monitoring Success

- [ ] All metrics being collected
- [ ] Alerts configured correctly
- [ ] Logs being aggregated
- [ ] Performance trends visible
- [ ] Health status accurate

## Rollback Plan

### If Deployment Fails

1. **Immediate Actions**
   - [ ] Check Railway logs for errors
   - [ ] Verify environment variables
   - [ ] Test database connectivity
   - [ ] Check external API status

2. **Rollback Steps**
   - [ ] Revert to previous deployment
   - [ ] Restore previous environment variables
   - [ ] Verify rollback success
   - [ ] Document issues for next deployment

3. **Post-Rollback**
   - [ ] Investigate root cause
   - [ ] Fix identified issues
   - [ ] Test fixes locally
   - [ ] Plan next deployment

## Maintenance Checklist

### Daily Monitoring

- [ ] Check service health status
- [ ] Review performance metrics
- [ ] Check for alerts
- [ ] Monitor error rates
- [ ] Review resource usage

### Weekly Maintenance

- [ ] Review performance trends
- [ ] Check for security updates
- [ ] Review and rotate API keys
- [ ] Update documentation
- [ ] Plan capacity scaling

### Monthly Maintenance

- [ ] Review deployment process
- [ ] Update dependencies
- [ ] Review monitoring configuration
- [ ] Conduct disaster recovery test
- [ ] Update deployment documentation

## Contact Information

### Support Channels

- **Development Team**: [team@company.com]
- **DevOps Team**: [devops@company.com]
- **Railway Support**: [Railway Documentation](https://docs.railway.app)

### Emergency Contacts

- **On-Call Engineer**: [oncall@company.com]
- **Team Lead**: [lead@company.com]
- **Manager**: [manager@company.com]

## Documentation Links

- [Railway Deployment Guide](./RAILWAY_DEPLOYMENT.md)
- [Performance Monitoring Guide](./PERFORMANCE_MONITORING.md)
- [API Documentation](./API_DOCUMENTATION.md)
- [Troubleshooting Guide](./TROUBLESHOOTING.md)

---

**Last Updated**: [Current Date]
**Version**: 1.0.0
**Next Review**: [Next Review Date]
