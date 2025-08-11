# KYB Platform - Rollback Procedures

## Overview

The KYB Platform implements comprehensive rollback procedures designed to provide rapid, reliable, and safe rollback capabilities across all environments. This system supports multiple rollback strategies with automated validation and safety checks.

## Rollback Architecture

### Multi-Strategy Rollback Support

The rollback system supports multiple rollback strategies:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Previous      │    │   Specific      │    │   Emergency     │
│   Rollback      │    │   Version       │    │   Rollback      │
│                 │    │   Rollback      │    │                 │
│ • Previous      │    │ • Target        │    │ • Known Stable  │
│   Version       │    │   Version       │    │   Version       │
│ • Automatic     │    │ • Manual        │    │ • Critical      │
│   Detection     │    │   Selection     │    │   Situations    │
│ • Safe Default  │    │ • Validation    │    │ • Force Rollback│
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Rollback Pipeline

```
Trigger → Validate → Rollback → Verify → Test → Notify → Record
   ↓         ↓          ↓         ↓       ↓       ↓        ↓
Manual/   Version    ECS/K8s   Health   API     Slack/   Audit
Auto     Check      Update    Check    Tests   Email    Trail
```

## GitHub Actions Integration

### Dedicated Rollback Workflow

The rollback procedures are integrated with GitHub Actions through the `.github/workflows/rollback.yml` workflow:

**Triggers:**
- Manual workflow dispatch with comprehensive options
- Automated rollback triggers based on health check failures
- Emergency rollback procedures

**Workflow Inputs:**
```yaml
workflow_dispatch:
  inputs:
    environment:
      description: "Environment to rollback"
      required: true
      default: "staging"
      type: choice
      options:
        - staging
        - production
    rollback_type:
      description: "Type of rollback to perform"
      required: true
      default: "previous"
      type: choice
      options:
        - previous
        - specific_version
        - emergency
    target_version:
      description: "Specific version to rollback to"
      required: false
      type: string
    reason:
      description: "Reason for rollback"
      required: true
      type: string
    force_rollback:
      description: "Force rollback even if health checks fail"
      required: false
      default: false
      type: boolean
```

### Workflow Jobs

1. **Pre-rollback Validation**
   - Validates rollback conditions and parameters
   - Determines current and target versions
   - Checks environment and version compatibility
   - Validates rollback version existence

2. **Environment-Specific Rollback**
   - Staging rollback with health verification
   - Production rollback with comprehensive testing
   - Environment-specific validation and safety checks

3. **Post-rollback Verification**
   - Health check validation
   - API functionality testing
   - Version confirmation
   - Rollback record creation

4. **Rollback Summary**
   - Comprehensive rollback reporting
   - Issue creation for tracking
   - Notification distribution
   - Audit trail maintenance

## Rollback Scripts

### Main Rollback Script

The `scripts/rollback.sh` script provides comprehensive rollback capabilities:

**Features:**
- Multiple rollback strategies (previous, specific, emergency)
- Multi-environment support (staging, production)
- AWS ECS and Kubernetes rollback support
- Comprehensive validation and safety checks
- Health verification and post-rollback testing
- Rollback records and notifications
- Dry-run mode for testing

**Usage Examples:**

```bash
# Rollback to previous version in staging
./scripts/rollback.sh -e staging -t previous -r "Performance issues detected"

# Rollback to specific version in production
./scripts/rollback.sh -e production -t specific -v v1.2.3 -r "Critical bug in v1.3.0"

# Emergency rollback to stable version
./scripts/rollback.sh -e production -t emergency -r "Service unavailable" -f

# Dry run for testing rollback procedure
./scripts/rollback.sh -e staging --dry-run -t previous -r "Testing rollback procedure"
```

**Script Options:**

```bash
Options:
    -e, --environment ENV     Environment to rollback (staging|production)
    -t, --type TYPE          Rollback type (previous|specific|emergency)
    -v, --version VERSION    Specific version to rollback to (required for specific type)
    -r, --reason REASON      Reason for rollback (required)
    -f, --force             Force rollback even if health checks fail
    -d, --dry-run           Show what would be rolled back without actually rolling back
    -h, --help              Show this help message
```

## Rollback Strategies

### 1. Previous Version Rollback

**Use Case:** Standard rollback to the immediately previous version
**Trigger:** Performance issues, minor bugs, or deployment problems

**Process:**
```bash
# Get current version
CURRENT_VERSION=$(aws ecs describe-services --cluster kyb-platform-production --services kyb-platform-api --query 'services[0].taskDefinition' --output text)

# Calculate previous version
PREVIOUS_VERSION=$((CURRENT_VERSION - 1))

# Perform rollback
aws ecs update-service \
  --cluster kyb-platform-production \
  --service kyb-platform-api \
  --task-definition "kyb-platform-api:$PREVIOUS_VERSION" \
  --force-new-deployment
```

**Validation:**
- Previous version existence check
- Health check verification
- API functionality testing
- Version confirmation

### 2. Specific Version Rollback

**Use Case:** Rollback to a known good version
**Trigger:** Specific version issues or targeted rollback requirements

**Process:**
```bash
# Validate target version exists
aws ecs describe-task-definition --task-definition "kyb-platform-api:v1.2.3"

# Perform rollback to specific version
aws ecs update-service \
  --cluster kyb-platform-production \
  --service kyb-platform-api \
  --task-definition "kyb-platform-api:v1.2.3" \
  --force-new-deployment
```

**Validation:**
- Target version existence verification
- Version compatibility check
- Comprehensive health validation
- Full API testing suite

### 3. Emergency Rollback

**Use Case:** Critical service issues requiring immediate rollback
**Trigger:** Service unavailability, critical bugs, or security issues

**Process:**
```bash
# Use predefined stable version
STABLE_VERSION="kyb-platform-api:stable"

# Force rollback to stable version
aws ecs update-service \
  --cluster kyb-platform-production \
  --service kyb-platform-api \
  --task-definition "$STABLE_VERSION" \
  --force-new-deployment
```

**Validation:**
- Bypass normal health checks (with force flag)
- Basic service availability verification
- Emergency notification procedures
- Post-rollback investigation initiation

## AWS ECS Rollback

### ECS Service Rollback

**Service Update Process:**
```bash
# Update ECS service with rollback version
aws ecs update-service \
  --cluster kyb-platform-production \
  --service kyb-platform-api \
  --task-definition "$ROLLBACK_VERSION" \
  --force-new-deployment

# Wait for rollback to complete
aws ecs wait services-stable \
  --cluster kyb-platform-production \
  --services kyb-platform-api
```

**Rollback Validation:**
- Service stability verification
- Task definition validation
- Load balancer health check
- Auto-scaling group status

### ECS Task Definition Management

**Version Tracking:**
```bash
# List task definition revisions
aws ecs list-task-definitions --family-prefix kyb-platform-api

# Get specific task definition
aws ecs describe-task-definition --task-definition kyb-platform-api:123

# Compare task definitions
aws ecs describe-task-definition --task-definition kyb-platform-api:123 > current.json
aws ecs describe-task-definition --task-definition kyb-platform-api:122 > previous.json
diff current.json previous.json
```

## Kubernetes Rollback

### Kubernetes Deployment Rollback

**Rollback Process:**
```bash
# Rollback to previous revision
kubectl rollout undo deployment/kyb-platform-api -n kyb-platform

# Check rollback status
kubectl rollout status deployment/kyb-platform-api -n kyb-platform

# View rollback history
kubectl rollout history deployment/kyb-platform-api -n kyb-platform
```

**Rollback to Specific Revision:**
```bash
# Rollback to specific revision
kubectl rollout undo deployment/kyb-platform-api --to-revision=2 -n kyb-platform

# Check revision details
kubectl rollout history deployment/kyb-platform-api --revision=2 -n kyb-platform
```

### Kubernetes Rollback Validation

**Health Check Verification:**
```bash
# Check pod status
kubectl get pods -n kyb-platform -l app=kyb-platform-api

# Check service endpoints
kubectl get endpoints kyb-platform-api-service -n kyb-platform

# Verify ingress configuration
kubectl describe ingress kyb-platform-ingress -n kyb-platform
```

## Rollback Validation

### Health Check Procedures

**Pre-rollback Health Check:**
```bash
# Check current service health
curl -f https://api.kybplatform.com/health

# Check service status
curl -f https://api.kybplatform.com/status

# Verify metrics endpoint
curl -f https://api.kybplatform.com/metrics
```

**Post-rollback Health Check:**
```bash
# Wait for service readiness
sleep 60

# Comprehensive health validation
for endpoint in health status metrics; do
  curl -f "https://api.kybplatform.com/$endpoint" || {
    echo "❌ $endpoint check failed"
    exit 1
  }
done
```

### API Functionality Testing

**Post-rollback API Tests:**
```bash
# Test classification endpoint
curl -f https://api.kybplatform.com/v1/classify \
  -H "Content-Type: application/json" \
  -d '{"business_name": "Test Corp"}'

# Test authentication endpoint
curl -f https://api.kybplatform.com/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "password": "testpass"}'

# Test risk assessment endpoint
curl -f https://api.kybplatform.com/v1/risk/assess \
  -H "Content-Type: application/json" \
  -d '{"business_id": "test-123"}'
```

## Rollback Safety Measures

### Validation and Safety Checks

**Pre-rollback Validation:**
- Environment validation
- Version existence verification
- Rollback necessity check
- Prerequisites validation

**Rollback Safety:**
- Force rollback override capability
- Health check bypass options
- Emergency rollback procedures
- Rollback confirmation prompts

**Post-rollback Safety:**
- Comprehensive health verification
- API functionality testing
- Version confirmation
- Rollback record creation

### Rollback Triggers

**Automatic Rollback Triggers:**
- Health check failures
- Performance degradation
- Error rate thresholds
- Response time violations

**Manual Rollback Triggers:**
- Critical bug reports
- Security vulnerability detection
- Service unavailability
- Performance issues

## Rollback Monitoring and Alerting

### Rollback Metrics

**Key Metrics:**
- Rollback frequency and success rate
- Rollback duration and time to recovery
- Rollback trigger analysis
- Post-rollback stability metrics

**Monitoring Dashboard:**
```yaml
# Example Prometheus metrics
kyb_rollback_total{environment="production",type="previous"}
kyb_rollback_duration_seconds{environment="production"}
kyb_rollback_success_rate{environment="production"}
kyb_post_rollback_health_status{environment="production"}
```

### Rollback Alerting

**Alert Rules:**
- Rollback frequency alerts
- Rollback failure notifications
- Post-rollback health issues
- Emergency rollback triggers

**Notification Channels:**
- Slack notifications
- Email alerts
- PagerDuty integration
- SMS notifications

## Rollback Records and Audit Trail

### Rollback Record Creation

**Record Structure:**
```json
{
  "rollback_id": "20241201-143022",
  "environment": "production",
  "reason": "Critical bug in v1.3.0",
  "rollback_type": "specific",
  "from_version": "kyb-platform-api:123",
  "to_version": "kyb-platform-api:122",
  "triggered_by": "john.doe",
  "timestamp": "2024-12-01T14:30:22Z",
  "status": "completed",
  "force_rollback": false,
  "health_check_status": "passed",
  "api_test_status": "passed"
}
```

**Record Storage:**
- Local file system
- S3 bucket storage
- Database records
- Audit log integration

### Audit Trail

**Audit Information:**
- Rollback initiation details
- Approval and authorization
- Execution timeline
- Verification results
- Post-rollback actions

**Compliance Requirements:**
- SOC 2 audit trail
- Change management records
- Incident response documentation
- Post-incident analysis

## Rollback Best Practices

### Rollback Strategy Best Practices

1. **Version Management**
   - Maintain clear version history
   - Tag stable versions appropriately
   - Document version compatibility
   - Regular version cleanup

2. **Testing and Validation**
   - Comprehensive pre-rollback testing
   - Automated health check validation
   - API functionality verification
   - Performance impact assessment

3. **Communication and Coordination**
   - Clear rollback notification procedures
   - Stakeholder communication plans
   - Post-rollback status updates
   - Incident response coordination

4. **Documentation and Training**
   - Rollback procedure documentation
   - Team training and drills
   - Emergency contact procedures
   - Post-incident review processes

### Security Best Practices

1. **Access Control**
   - Role-based rollback permissions
   - Multi-factor authentication
   - Audit trail maintenance
   - Emergency access procedures

2. **Data Protection**
   - Database backup verification
   - Configuration backup procedures
   - Secret management during rollback
   - Data integrity validation

3. **Compliance**
   - Regulatory compliance maintenance
   - Audit trail requirements
   - Change management procedures
   - Incident reporting requirements

## Troubleshooting

### Common Rollback Issues

1. **Version Not Found**
   - Check version existence
   - Verify version naming convention
   - Validate version compatibility
   - Check version history

2. **Health Check Failures**
   - Verify service configuration
   - Check network connectivity
   - Validate endpoint availability
   - Review application logs

3. **Rollback Timeout**
   - Check resource availability
   - Verify auto-scaling configuration
   - Review deployment strategy
   - Monitor resource utilization

### Debugging Commands

```bash
# Check ECS service status
aws ecs describe-services --cluster kyb-platform-production --services kyb-platform-api

# View container logs
aws logs get-log-events --log-group-name /ecs/kyb-platform-api --log-stream-name <stream-name>

# Check Kubernetes deployment status
kubectl get deployments -n kyb-platform
kubectl describe deployment kyb-platform-api -n kyb-platform

# Test service connectivity
curl -v https://api.kybplatform.com/health
telnet api.kybplatform.com 443
```

## Future Enhancements

### Planned Improvements

1. **Advanced Rollback Strategies**
   - Blue-green rollback implementation
   - Canary rollback procedures
   - A/B testing rollback support
   - Feature flag rollback integration

2. **Intelligent Rollback**
   - Machine learning-based rollback decisions
   - Predictive rollback triggers
   - Automated rollback optimization
   - Smart version selection

3. **Enhanced Monitoring**
   - Real-time rollback monitoring
   - Predictive failure detection
   - Automated rollback recommendations
   - Performance impact analysis

4. **Integration Enhancements**
   - Multi-cloud rollback support
   - Hybrid deployment rollback
   - Third-party service integration
   - Advanced notification systems

---

This documentation provides a comprehensive overview of the KYB Platform's rollback procedures. For specific implementation details, refer to the rollback scripts and workflow files referenced throughout this document.
