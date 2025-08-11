# KYB Platform - Blue-Green Deployments

## Overview

The KYB Platform implements comprehensive blue-green deployment capabilities designed to provide zero-downtime deployments with automatic traffic switching, health monitoring, and rollback capabilities. This system ensures continuous service availability during deployments.

## Blue-Green Deployment Architecture

### Dual Environment Setup

The blue-green deployment system maintains two identical environments:

```
┌─────────────────┐    ┌─────────────────┐
│   Blue Environment │    │  Green Environment │
│                 │    │                 │
│ • Active Traffic │    │ • Standby Ready │
│ • Production     │    │ • Pre-deployed  │
│ • Current Version│    │ • New Version   │
│ • Load Balancer  │    │ • Health Checks │
└─────────────────┘    └─────────────────┘
         │                       │
         └───────────────────────┘
                    │
         ┌─────────────────┐
         │  Load Balancer  │
         │                 │
         │ • Traffic Switch│
         │ • Health Monitor│
         │ • Auto Failover │
         └─────────────────┘
```

### Deployment Pipeline

```
Code Commit → Build → Deploy to Inactive → Health Check → Traffic Switch → Verify → Cleanup
     ↓           ↓           ↓                ↓              ↓            ↓         ↓
   Git Push   Docker    ECS Service      Health/API     Load Balancer  Tests    Scale Down
              Build     Update          Validation      Switch         Pass     Old Env
```

## GitHub Actions Integration

### Blue-Green Deployment Workflow

The blue-green deployment automation is integrated with GitHub Actions through the `.github/workflows/blue-green-deployment.yml` workflow:

**Triggers:**
- Push to `main` branch → Production blue-green deployment
- Push to `develop` branch → Staging blue-green deployment
- Manual workflow dispatch → Custom environment deployment

**Workflow Inputs:**
```yaml
workflow_dispatch:
  inputs:
    environment:
      description: "Environment to deploy to"
      required: true
      default: "staging"
      type: choice
      options:
        - staging
        - production
    deployment_strategy:
      description: "Blue-green deployment strategy"
      required: true
      default: "automatic"
      type: choice
      options:
        - automatic
        - manual_switch
        - canary
    health_check_timeout:
      description: "Health check timeout in minutes"
      required: false
      default: "10"
      type: string
    auto_switch_traffic:
      description: "Automatically switch traffic after health checks"
      required: false
      default: true
      type: boolean
```

### Workflow Jobs

1. **Pre-deployment Validation**
   - Validates blue-green deployment conditions
   - Determines current active environment (blue/green)
   - Validates deployment strategy and parameters
   - Checks environment readiness

2. **Build and Push Image**
   - Multi-platform Docker builds (amd64, arm64)
   - Container registry integration (GHCR.io)
   - Image tagging and metadata management

3. **Deploy to Inactive Environment**
   - Deploys to the currently inactive environment
   - Waits for deployment completion
   - Runs health checks and smoke tests
   - Validates new environment readiness

4. **Traffic Switch**
   - Automatically switches traffic to new environment
   - Updates load balancer configuration
   - Verifies traffic switch success
   - Monitors post-switch health

5. **Post-Switch Validation**
   - Comprehensive API testing
   - Performance validation
   - Endpoint verification
   - Error rate monitoring

6. **Cleanup Old Environment**
   - Scales down old environment
   - Resource cleanup
   - Cost optimization
   - Environment preparation for next deployment

## Blue-Green Deployment Scripts

### Main Blue-Green Deployment Script

The `scripts/blue-green-deploy.sh` script provides comprehensive blue-green deployment capabilities:

**Features:**
- Multiple deployment strategies (automatic, manual_switch, canary)
- Multi-environment support (staging, production)
- Automatic environment detection (blue/green)
- Comprehensive health checks and validation
- Traffic switching automation
- Post-deployment testing
- Old environment cleanup
- Deployment records and notifications

**Usage Examples:**

```bash
# Automatic blue-green deployment to staging
./scripts/blue-green-deploy.sh -e staging -s automatic

# Manual switch deployment to production
./scripts/blue-green-deploy.sh -e production -s manual_switch

# Dry run for staging deployment
./scripts/blue-green-deploy.sh -e staging --dry-run

# Custom health check timeout
./scripts/blue-green-deploy.sh -e production -t 15
```

**Script Options:**

```bash
Options:
    -e, --environment ENV     Environment to deploy to (staging|production)
    -s, --strategy STRATEGY   Deployment strategy (automatic|manual_switch|canary)
    -t, --timeout TIMEOUT     Health check timeout in minutes [default: 10]
    -a, --auto-switch         Automatically switch traffic after health checks
    -d, --dry-run            Show what would be deployed without actually deploying
    -h, --help               Show this help message
```

## Deployment Strategies

### 1. Automatic Deployment

**Use Case:** Standard blue-green deployment with automatic traffic switching
**Process:** Deploy → Validate → Switch → Cleanup

**Process Flow:**
```bash
# 1. Determine current active environment
CURRENT_ENV=$(get_current_active_environment)

# 2. Deploy to inactive environment
TARGET_ENV=$(get_inactive_environment)
deploy_to_environment $TARGET_ENV

# 3. Run health checks
run_health_checks $TARGET_ENV

# 4. Switch traffic
switch_traffic $CURRENT_ENV $TARGET_ENV

# 5. Verify switch
verify_traffic_switch

# 6. Cleanup old environment
cleanup_environment $CURRENT_ENV
```

**Validation Steps:**
- Health check validation
- API functionality testing
- Performance monitoring
- Error rate verification

### 2. Manual Switch Deployment

**Use Case:** Deployment with manual traffic switch approval
**Process:** Deploy → Validate → Wait for Manual Switch → Cleanup

**Process Flow:**
```bash
# 1. Deploy to inactive environment
deploy_to_inactive_environment

# 2. Run comprehensive validation
run_health_checks
run_smoke_tests
run_performance_tests

# 3. Wait for manual approval
echo "Target environment ready for traffic switch"
echo "Manual intervention required"

# 4. Manual traffic switch (external process)
# 5. Verify switch and cleanup
```

**Manual Switch Commands:**
```bash
# Switch traffic manually
aws elbv2 modify-listener \
  --listener-arn $LISTENER_ARN \
  --default-actions Type=forward,TargetGroupArn=$NEW_TARGET_GROUP_ARN

# Verify switch
curl -f https://api.kybplatform.com/health
```

### 3. Canary Deployment

**Use Case:** Gradual traffic shifting for risk mitigation
**Process:** Deploy → Gradual Traffic Shift → Full Switch → Cleanup

**Process Flow:**
```bash
# 1. Deploy to inactive environment
deploy_to_inactive_environment

# 2. Initial traffic shift (5%)
shift_traffic_percentage 5

# 3. Monitor and validate
monitor_metrics 5_minutes

# 4. Increase traffic gradually
for percentage in 10 25 50 75 100; do
  shift_traffic_percentage $percentage
  monitor_metrics 5_minutes
  validate_health
done

# 5. Complete switch and cleanup
complete_traffic_switch
cleanup_old_environment
```

## AWS Infrastructure Setup

### Load Balancer Configuration

**Application Load Balancer Setup:**
```bash
# Create load balancer
aws elbv2 create-load-balancer \
  --name kyb-platform-production \
  --subnets subnet-12345678 subnet-87654321 \
  --security-groups sg-12345678

# Create target groups
aws elbv2 create-target-group \
  --name kyb-platform-production-blue \
  --protocol HTTP \
  --port 8080 \
  --vpc-id vpc-12345678 \
  --health-check-path /health

aws elbv2 create-target-group \
  --name kyb-platform-production-green \
  --protocol HTTP \
  --port 8080 \
  --vpc-id vpc-12345678 \
  --health-check-path /health

# Create listener
aws elbv2 create-listener \
  --load-balancer-arn $LB_ARN \
  --protocol HTTPS \
  --port 443 \
  --certificates CertificateArn=$CERT_ARN \
  --default-actions Type=forward,TargetGroupArn=$BLUE_TG_ARN
```

### ECS Service Configuration

**Blue Environment Service:**
```bash
# Create blue ECS service
aws ecs create-service \
  --cluster kyb-platform-production \
  --service-name kyb-platform-api-blue \
  --task-definition kyb-platform-api:latest \
  --desired-count 3 \
  --launch-type FARGATE \
  --network-configuration "awsvpcConfiguration={subnets=[subnet-12345678],securityGroups=[sg-12345678],assignPublicIp=ENABLED}" \
  --load-balancers "targetGroupArn=$BLUE_TG_ARN,containerName=kyb-platform-api,containerPort=8080"
```

**Green Environment Service:**
```bash
# Create green ECS service
aws ecs create-service \
  --cluster kyb-platform-production \
  --service-name kyb-platform-api-green \
  --task-definition kyb-platform-api:latest \
  --desired-count 0 \
  --launch-type FARGATE \
  --network-configuration "awsvpcConfiguration={subnets=[subnet-12345678],securityGroups=[sg-12345678],assignPublicIp=ENABLED}" \
  --load-balancers "targetGroupArn=$GREEN_TG_ARN,containerName=kyb-platform-api,containerPort=8080"
```

## Health Check and Validation

### Health Check Procedures

**Pre-Switch Health Checks:**
```bash
# Basic health check
curl -f http://green.production.kybplatform.com/health

# Comprehensive health validation
for endpoint in health status metrics; do
  curl -f "http://green.production.kybplatform.com/$endpoint" || exit 1
done

# API functionality test
curl -f http://green.production.kybplatform.com/v1/classify \
  -H "Content-Type: application/json" \
  -d '{"business_name": "Test Corp"}'
```

**Post-Switch Validation:**
```bash
# Verify traffic switch
curl -f https://api.kybplatform.com/health

# Test all endpoints
ENDPOINTS=("health" "status" "metrics" "v1/classify" "v1/auth/login" "v1/risk/assess")

for endpoint in "${ENDPOINTS[@]}"; do
  echo "Testing endpoint: $endpoint"
  curl -f "https://api.kybplatform.com/$endpoint" || {
    echo "❌ $endpoint test failed"
    exit 1
  }
done
```

### Performance Monitoring

**Key Metrics:**
- Response time (p50, p95, p99)
- Error rate (4xx, 5xx)
- Throughput (requests/second)
- Resource utilization (CPU, memory)

**Monitoring Commands:**
```bash
# Monitor response times
curl -w "@curl-format.txt" -o /dev/null -s https://api.kybplatform.com/health

# Check error rates
curl -s https://api.kybplatform.com/metrics | grep http_requests_total

# Monitor resource usage
aws cloudwatch get-metric-statistics \
  --namespace AWS/ECS \
  --metric-name CPUUtilization \
  --dimensions Name=ServiceName,Value=kyb-platform-api-green \
  --start-time $(date -d '5 minutes ago' -u +%Y-%m-%dT%H:%M:%S) \
  --end-time $(date -u +%Y-%m-%dT%H:%M:%S) \
  --period 300 \
  --statistics Average
```

## Traffic Switching

### Load Balancer Traffic Switch

**Automatic Traffic Switch:**
```bash
# Get current listener configuration
LISTENER_ARN=$(aws elbv2 describe-listeners \
  --load-balancer-arn $LB_ARN \
  --query 'Listeners[0].ListenerArn' \
  --output text)

# Switch traffic to new target group
aws elbv2 modify-listener \
  --listener-arn $LISTENER_ARN \
  --default-actions Type=forward,TargetGroupArn=$NEW_TARGET_GROUP_ARN
```

**Traffic Switch Verification:**
```bash
# Verify target group health
aws elbv2 describe-target-health \
  --target-group-arn $NEW_TARGET_GROUP_ARN \
  --query 'TargetHealthDescriptions[?TargetHealth.State==`healthy`] | length(@)'

# Test traffic flow
curl -f https://api.kybplatform.com/health

# Monitor traffic distribution
aws cloudwatch get-metric-statistics \
  --namespace AWS/ApplicationELB \
  --metric-name RequestCount \
  --dimensions Name=TargetGroup,Value=$NEW_TARGET_GROUP_NAME \
  --start-time $(date -d '1 minute ago' -u +%Y-%m-%dT%H:%M:%S) \
  --end-time $(date -u +%Y-%m-%dT%H:%M:%S) \
  --period 60 \
  --statistics Sum
```

## Environment Cleanup

### Old Environment Cleanup

**Service Scale Down:**
```bash
# Scale down old service
aws ecs update-service \
  --cluster kyb-platform-production \
  --service kyb-platform-api-blue \
  --desired-count 0

# Wait for scale down
aws ecs wait services-stable \
  --cluster kyb-platform-production \
  --services kyb-platform-api-blue
```

**Resource Cleanup:**
```bash
# Remove old target group from load balancer
aws elbv2 modify-listener \
  --listener-arn $LISTENER_ARN \
  --default-actions Type=forward,TargetGroupArn=$ACTIVE_TARGET_GROUP_ARN

# Optional: Delete old target group
aws elbv2 delete-target-group \
  --target-group-arn $OLD_TARGET_GROUP_ARN
```

## Monitoring and Alerting

### Blue-Green Deployment Metrics

**Key Metrics:**
- Deployment success/failure rate
- Traffic switch duration
- Health check pass/fail rate
- Post-deployment error rates
- Environment cleanup time

**Prometheus Metrics:**
```yaml
# Blue-green deployment metrics
kyb_blue_green_deployment_total{environment="production",strategy="automatic"}
kyb_blue_green_deployment_duration_seconds{environment="production"}
kyb_blue_green_traffic_switch_duration_seconds{environment="production"}
kyb_blue_green_health_check_status{environment="production"}
kyb_blue_green_post_deployment_error_rate{environment="production"}
```

### Alerting Rules

**Deployment Alerts:**
- Blue-green deployment failure
- Traffic switch timeout
- Health check failures
- Post-deployment error rate increase
- Environment cleanup failures

**Alert Configuration:**
```yaml
# Example alert rule
groups:
  - name: blue-green-deployment
    rules:
      - alert: BlueGreenDeploymentFailed
        expr: kyb_blue_green_deployment_total{status="failed"} > 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Blue-green deployment failed"
          description: "Blue-green deployment to {{ $labels.environment }} failed"
```

## Best Practices

### Blue-Green Deployment Best Practices

1. **Environment Preparation**
   - Maintain identical blue and green environments
   - Use infrastructure as code for consistency
   - Regular environment synchronization
   - Automated environment validation

2. **Health Check Strategy**
   - Comprehensive health check endpoints
   - Multiple validation layers
   - Performance baseline monitoring
   - Automated rollback triggers

3. **Traffic Switching**
   - Atomic traffic switch operations
   - Traffic switch verification
   - Rollback capability
   - Monitoring during switch

4. **Post-Deployment Monitoring**
   - Extended monitoring period
   - Performance comparison
   - Error rate tracking
   - User experience monitoring

### Security Best Practices

1. **Access Control**
   - Role-based deployment permissions
   - Multi-factor authentication
   - Audit trail maintenance
   - Least privilege principle

2. **Data Protection**
   - Database migration strategies
   - Data consistency validation
   - Backup verification
   - Rollback data integrity

3. **Network Security**
   - VPC isolation
   - Security group configuration
   - SSL/TLS termination
   - Network monitoring

## Troubleshooting

### Common Issues

1. **Health Check Failures**
   - Verify service configuration
   - Check network connectivity
   - Validate endpoint availability
   - Review application logs

2. **Traffic Switch Issues**
   - Verify load balancer configuration
   - Check target group health
   - Validate listener rules
   - Monitor traffic flow

3. **Environment Cleanup Problems**
   - Check service dependencies
   - Verify resource permissions
   - Monitor cleanup progress
   - Manual cleanup procedures

### Debugging Commands

```bash
# Check target group health
aws elbv2 describe-target-health --target-group-arn $TARGET_GROUP_ARN

# View ECS service events
aws ecs describe-services --cluster kyb-platform-production --services kyb-platform-api-green

# Check load balancer configuration
aws elbv2 describe-listeners --load-balancer-arn $LB_ARN

# Monitor traffic flow
aws cloudwatch get-metric-statistics \
  --namespace AWS/ApplicationELB \
  --metric-name RequestCount \
  --dimensions Name=LoadBalancer,Value=kyb-platform-production

# Test connectivity
curl -v https://api.kybplatform.com/health
telnet api.kybplatform.com 443
```

## Future Enhancements

### Planned Improvements

1. **Advanced Deployment Strategies**
   - Canary deployment implementation
   - A/B testing integration
   - Feature flag deployment
   - Progressive delivery

2. **Intelligent Traffic Management**
   - Machine learning-based traffic switching
   - Predictive failure detection
   - Automated rollback decisions
   - Performance-based routing

3. **Enhanced Monitoring**
   - Real-time deployment monitoring
   - Predictive analytics
   - Automated performance optimization
   - Intelligent alerting

4. **Multi-Cloud Support**
   - Google Cloud Platform integration
   - Azure Container Instances
   - Hybrid cloud deployments
   - Cloud-agnostic configurations

---

This documentation provides a comprehensive overview of the KYB Platform's blue-green deployment system. For specific implementation details, refer to the blue-green deployment scripts and workflow files referenced throughout this document.
