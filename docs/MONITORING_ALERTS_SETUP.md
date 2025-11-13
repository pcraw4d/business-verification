# Railway Monitoring & Alerts Setup Guide

## Overview

This guide covers setting up monitoring and alerts for the KYB Platform services deployed on Railway.

## 🎯 Monitoring Objectives

1. **Service Availability**: Ensure all services are running and responding
2. **Error Rate Monitoring**: Track and alert on high error rates
3. **Performance Monitoring**: Monitor response times and resource usage
4. **Resource Usage**: Track CPU, memory, and network usage
5. **Database Health**: Monitor Supabase connection and query performance
6. **Redis Health**: Monitor Redis connectivity and cache performance

## 📊 Railway Built-in Monitoring

Railway provides built-in monitoring for all services:

### Metrics Available
- **CPU Usage**: Percentage of CPU used
- **Memory Usage**: Memory consumption in MB
- **Network I/O**: Incoming and outgoing traffic
- **Request Count**: Number of HTTP requests
- **Error Rate**: Percentage of failed requests
- **Response Time**: Average response time

### Accessing Metrics
1. Go to Railway Dashboard
2. Select a service
3. Click on "Metrics" tab
4. View real-time and historical metrics

## 🔔 Setting Up Alerts

### Railway Alerts (Recommended)

Railway provides built-in alerting:

1. **Go to Railway Dashboard**
2. **Select your project**
3. **Go to "Settings" → "Alerts"**
4. **Configure alerts for:**

#### Service Down Alert
- **Trigger**: Service deployment fails or service becomes unresponsive
- **Action**: Email/Slack notification
- **Threshold**: Service unavailable for > 2 minutes

#### High Error Rate Alert
- **Trigger**: Error rate > 5% for 5 minutes
- **Action**: Email/Slack notification
- **Threshold**: 5% error rate sustained for 5 minutes

#### Resource Usage Alerts
- **CPU Alert**: CPU usage > 80% for 10 minutes
- **Memory Alert**: Memory usage > 90% for 5 minutes
- **Action**: Email/Slack notification

### Custom Alert Configuration

```yaml
# Example Railway alert configuration
alerts:
  - name: "Service Down"
    condition: "deployment_status == 'failed'"
    duration: "2m"
    action: "email,slack"
  
  - name: "High Error Rate"
    condition: "error_rate > 0.05"
    duration: "5m"
    action: "email,slack"
  
  - name: "High CPU Usage"
    condition: "cpu_usage > 0.80"
    duration: "10m"
    action: "email"
  
  - name: "High Memory Usage"
    condition: "memory_usage > 0.90"
    duration: "5m"
    action: "email"
```

## 🔍 Service-Specific Monitoring

### API Gateway Monitoring

**Key Metrics:**
- Request rate (requests/second)
- Response time (p50, p95, p99)
- Error rate (4xx, 5xx)
- Upstream service availability

**Alerts:**
- Response time > 2 seconds (p95)
- Error rate > 1%
- Upstream service unavailable

### Redis Cache Monitoring

**Key Metrics:**
- Connection count
- Memory usage
- Hit/miss ratio
- Command latency

**Alerts:**
- Memory usage > 80%
- Connection failures
- High latency (> 10ms)

### Database (Supabase) Monitoring

**Key Metrics:**
- Connection pool usage
- Query performance
- Database size
- Active connections

**Alerts:**
- Connection pool exhaustion
- Slow queries (> 1 second)
- Database size approaching limits

## 📈 External Monitoring Tools

### Option 1: Uptime Robot (Free Tier)

1. **Sign up at**: https://uptimerobot.com
2. **Add monitors for each service:**
   - API Gateway: `https://api-gateway-service-production-21fd.up.railway.app/health`
   - Classification: `https://classification-service-production.up.railway.app/health`
   - Merchant: `https://merchant-service-production.up.railway.app/health`
   - Risk Assessment: `https://risk-assessment-service-production.up.railway.app/health`
   - Frontend: `https://frontend-service-production-b225.up.railway.app/health`

3. **Configure alerts:**
   - Email notifications
   - SMS (paid)
   - Slack webhook
   - PagerDuty integration

### Option 2: Pingdom

1. **Sign up at**: https://www.pingdom.com
2. **Create uptime checks** for all health endpoints
3. **Set up alerting** via email/SMS/Slack

### Option 3: StatusCake

1. **Sign up at**: https://www.statuscake.com
2. **Add uptime tests** for all services
3. **Configure alerting** channels

## 🔗 Integration with Logging

### Railway Logs

Railway provides built-in log aggregation:

1. **Access logs**: Railway Dashboard → Service → Logs
2. **Filter logs**: By level, time range, search terms
3. **Export logs**: Download logs for analysis

### Log Monitoring

**Watch for:**
- Error patterns
- Warning messages
- Connection failures
- Performance degradation indicators

**Example log queries:**
- `level:error` - All errors
- `redis connection failed` - Redis issues
- `database connection failed` - Database issues
- `response_time > 2000` - Slow requests

## 📊 Dashboard Setup

### Railway Dashboard

Railway provides a built-in dashboard showing:
- Service status
- Resource usage
- Request metrics
- Error rates

**Access**: Railway Dashboard → Project → Overview

### Custom Dashboard (Optional)

If using external monitoring tools, create dashboards showing:
- Service health status
- Request rates
- Error rates
- Response times
- Resource usage

## 🚨 Alert Response Procedures

### Service Down Alert

1. **Check Railway Dashboard**:
   - Verify service status
   - Check deployment logs
   - Review recent changes

2. **Check Service Logs**:
   - Look for startup errors
   - Check for crash messages
   - Verify environment variables

3. **Restart Service** (if needed):
   - Railway Dashboard → Service → Deployments → Redeploy

### High Error Rate Alert

1. **Check Error Logs**:
   - Identify error patterns
   - Check for specific endpoints failing
   - Review recent code changes

2. **Check Dependencies**:
   - Verify database connectivity
   - Check Redis connectivity
   - Test upstream services

3. **Scale Service** (if needed):
   - Increase resources in Railway
   - Add service replicas

### Resource Usage Alert

1. **Check Resource Metrics**:
   - Identify which resource is constrained
   - Check for memory leaks
   - Review CPU-intensive operations

2. **Optimize or Scale**:
   - Optimize code if possible
   - Scale up service resources
   - Add service replicas

## 📝 Monitoring Checklist

- [ ] Railway alerts configured for all services
- [ ] Health check endpoints monitored
- [ ] Error rate alerts set up
- [ ] Resource usage alerts configured
- [ ] External monitoring tool set up (optional)
- [ ] Log aggregation configured
- [ ] Dashboard created for key metrics
- [ ] Alert response procedures documented
- [ ] Team notified of alert channels

## 🔄 Continuous Monitoring

### Daily Checks
- Review service health status
- Check error logs for new issues
- Monitor resource usage trends

### Weekly Reviews
- Analyze error patterns
- Review performance metrics
- Check for optimization opportunities

### Monthly Reviews
- Review alert effectiveness
- Analyze resource usage trends
- Plan capacity adjustments

## 📚 Additional Resources

- [Railway Monitoring Docs](https://docs.railway.app/develop/monitoring)
- [Railway Alerts Guide](https://docs.railway.app/develop/alerts)
- [Supabase Monitoring](https://supabase.com/docs/guides/platform/metrics)

