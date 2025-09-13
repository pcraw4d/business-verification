# KYB Platform Monitoring Setup Guide

This guide provides comprehensive instructions for setting up and configuring the monitoring stack for the KYB Platform.

## Overview

The KYB Platform monitoring stack includes:

- **Prometheus** - Metrics collection and storage
- **Grafana** - Metrics visualization and dashboards
- **AlertManager** - Alert management and notification
- **Node Exporter** - System metrics collection
- **Blackbox Exporter** - HTTP health checks
- **Custom KYB Monitoring** - Application-specific monitoring

## Quick Start

### Prerequisites

- Docker and Docker Compose installed
- Go 1.22+ (for building the application)
- curl (for testing endpoints)

### 1. Setup Monitoring Stack

```bash
# Run the monitoring setup script
./scripts/setup-monitoring.sh setup
```

This will:
- Create necessary directories
- Start the monitoring stack
- Import Grafana dashboards
- Configure health checks

### 2. Access Monitoring Interfaces

- **Grafana Dashboard**: http://localhost:3000 (admin/admin123)
- **Prometheus**: http://localhost:9090
- **AlertManager**: http://localhost:9093
- **KYB Application**: http://localhost:8080

### 3. Verify Setup

```bash
# Check if all services are running
./scripts/setup-monitoring.sh status

# View logs
./scripts/setup-monitoring.sh logs
```

## Detailed Configuration

### Monitoring Configuration

The monitoring configuration is defined in `configs/monitoring.yml`:

```yaml
monitoring:
  enabled: true
  environment: "development"
  service_name: "kyb-platform"
  version: "1.0.0"
  
  collection:
    interval: "30s"
    timeout: "10s"
    retry_attempts: 3
```

### Health Checks

Health checks are configured for:

- **API Server** - Basic API responsiveness
- **Database** - Database connectivity and performance
- **Cache System** - Cache availability and performance
- **External APIs** - Third-party service availability
- **File System** - Disk space and permissions
- **Memory** - Memory usage and garbage collection

Access health checks at:
- `/health` - Overall system health
- `/health/detailed` - Detailed health information
- `/health/live` - Kubernetes liveness probe
- `/health/ready` - Kubernetes readiness probe
- `/health/startup` - Kubernetes startup probe

### Alerting Rules

Default alerting rules include:

- **High Error Rate** (>5% for 2 minutes) - Critical
- **High Response Time** (>2s for 5 minutes) - Warning
- **High Memory Usage** (>80% for 5 minutes) - Warning
- **High CPU Usage** (>80% for 5 minutes) - Warning
- **Service Down** (no response for 1 minute) - Critical

### Metrics Collection

The following metrics are collected:

#### Application Metrics
- Request rate (requests/second)
- Response time (percentiles: 50th, 95th, 99th)
- Error rate (4xx and 5xx responses)
- Active users
- Business verification metrics

#### System Metrics
- Memory usage (heap, system, GC)
- CPU usage
- Goroutine count
- Garbage collection statistics

#### Database Metrics
- Connection pool status
- Query duration
- Query count
- Error count

#### External API Metrics
- Response time
- Error rate
- Rate limit status

## Dashboard Overview

### KYB Platform Dashboard

The main dashboard includes:

1. **Key Metrics Cards**
   - Request Rate
   - Response Time
   - Error Rate
   - Active Users
   - Memory Usage
   - CPU Usage

2. **Performance Charts**
   - Request Rate Over Time
   - Response Time Distribution
   - Memory Usage Trends
   - CPU Usage Trends

3. **Health Status**
   - System Health Checks
   - Service Status Indicators
   - Dependency Health

4. **Active Alerts**
   - Current Alerts
   - Alert History
   - Alert Statistics

### Custom Dashboards

You can create additional dashboards for:

- **Business Metrics** - Verification success rates, processing times
- **Security Metrics** - Authentication failures, rate limiting
- **Performance Metrics** - Detailed performance analysis
- **Infrastructure Metrics** - System resource utilization

## Alerting Configuration

### Notification Channels

Configured notification channels:

1. **Email Notifications**
   - SMTP configuration
   - Recipient lists
   - Severity filtering

2. **Slack Notifications**
   - Webhook integration
   - Channel routing
   - Message formatting

3. **Webhook Notifications**
   - Custom endpoint integration
   - Authentication
   - Payload customization

### Alert Suppression

Alert suppression rules:

- **Maintenance Windows** - Suppress during scheduled maintenance
- **Deployment Windows** - Suppress during deployments
- **Business Hours** - Different thresholds for business hours

## API Endpoints

### Monitoring API

The monitoring API provides programmatic access to monitoring data:

```bash
# Get current metrics
GET /api/v3/monitoring/metrics

# Get active alerts
GET /api/v3/monitoring/alerts

# Get health checks
GET /api/v3/monitoring/health

# Get performance metrics
GET /api/v3/monitoring/performance

# Get slow queries
GET /api/v3/monitoring/slow-queries

# Get resource usage
GET /api/v3/monitoring/resources
```

### Health Check API

```bash
# Overall health
GET /health

# Detailed health
GET /health/detailed

# Liveness probe
GET /health/live

# Readiness probe
GET /health/ready

# Startup probe
GET /health/startup

# Individual service health
GET /health/api
GET /health/database
GET /health/cache
GET /health/external-apis
GET /health/filesystem
GET /health/memory
```

## Troubleshooting

### Common Issues

1. **Services Not Starting**
   ```bash
   # Check Docker status
   docker ps
   
   # Check logs
   docker-compose -f docker-compose.monitoring.yml logs
   ```

2. **Grafana Dashboard Not Loading**
   ```bash
   # Check Grafana logs
   docker logs kyb-grafana
   
   # Restart Grafana
   docker-compose -f docker-compose.monitoring.yml restart grafana
   ```

3. **Prometheus Not Scraping Metrics**
   ```bash
   # Check Prometheus targets
   curl http://localhost:9090/api/v1/targets
   
   # Check Prometheus configuration
   curl http://localhost:9090/api/v1/status/config
   ```

4. **Health Checks Failing**
   ```bash
   # Test health endpoints
   curl http://localhost:8080/health
   curl http://localhost:8080/health/detailed
   ```

### Performance Issues

1. **High Memory Usage**
   - Check for memory leaks in application
   - Adjust Prometheus retention settings
   - Optimize Grafana dashboard queries

2. **Slow Dashboard Loading**
   - Reduce query time ranges
   - Optimize Prometheus queries
   - Use data source caching

3. **Alert Fatigue**
   - Adjust alert thresholds
   - Implement alert suppression rules
   - Use alert grouping

## Security Considerations

### Access Control

- Change default passwords
- Use strong authentication
- Implement role-based access control
- Enable HTTPS in production

### Data Privacy

- Anonymize sensitive data in metrics
- Use secure communication channels
- Implement data retention policies
- Regular security audits

### Network Security

- Use firewalls to restrict access
- Implement network segmentation
- Use VPN for remote access
- Monitor network traffic

## Production Deployment

### Scaling Considerations

1. **Prometheus Scaling**
   - Use Prometheus federation
   - Implement horizontal scaling
   - Use remote storage

2. **Grafana Scaling**
   - Use Grafana clustering
   - Implement load balancing
   - Use external databases

3. **AlertManager Scaling**
   - Use AlertManager clustering
   - Implement high availability
   - Use external storage

### High Availability

1. **Service Redundancy**
   - Run multiple instances
   - Use load balancers
   - Implement failover

2. **Data Backup**
   - Regular backups
   - Cross-region replication
   - Disaster recovery plans

3. **Monitoring Redundancy**
   - Multiple monitoring stacks
   - Cross-stack validation
   - Independent alerting

## Maintenance

### Regular Tasks

1. **Weekly**
   - Review alert rules
   - Check dashboard performance
   - Update documentation

2. **Monthly**
   - Analyze metrics trends
   - Optimize queries
   - Review security settings

3. **Quarterly**
   - Capacity planning
   - Performance tuning
   - Security audits

### Updates

1. **Monitoring Stack Updates**
   ```bash
   # Update Docker images
   docker-compose -f docker-compose.monitoring.yml pull
   
   # Restart services
   docker-compose -f docker-compose.monitoring.yml up -d
   ```

2. **Configuration Updates**
   - Update monitoring.yml
   - Restart affected services
   - Validate changes

## Support

For issues and questions:

1. Check the troubleshooting section
2. Review logs and metrics
3. Consult the documentation
4. Contact the development team

## Additional Resources

- [Prometheus Documentation](https://prometheus.io/docs/)
- [Grafana Documentation](https://grafana.com/docs/)
- [AlertManager Documentation](https://prometheus.io/docs/alerting/latest/alertmanager/)
- [KYB Platform Documentation](../README.md)
