# KYB Platform - Deployment Automation

## Overview

The KYB Platform implements a comprehensive deployment automation system designed to provide reliable, secure, and efficient deployments across multiple environments. This system supports both AWS ECS and Kubernetes deployments with full CI/CD integration.

## Deployment Architecture

### Multi-Environment Support

The deployment system supports multiple environments:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Development   │    │     Staging     │    │   Production    │
│                 │    │                 │    │                 │
│ • Local Docker  │    │ • AWS ECS       │    │ • AWS ECS       │
│ • Hot Reload    │    │ • Auto-scaling  │    │ • Auto-scaling  │
│ • Debug Mode    │    │ • Load Testing  │    │ • High Availability │
│ • Unit Tests    │    │ • Integration   │    │ • Monitoring    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Deployment Pipeline

```
Code Commit → Build → Test → Security Scan → Deploy → Validate → Monitor
     ↓           ↓       ↓         ↓           ↓         ↓         ↓
   Git Push   Docker   Unit/Int   Trivy/Snyk  ECS/K8s  Health    Metrics
```

## GitHub Actions Integration

### Deployment Workflow

The deployment automation is integrated with GitHub Actions through the `.github/workflows/deployment.yml` workflow:

**Triggers:**
- Push to `main` branch → Production deployment
- Push to `develop` branch → Staging deployment
- Manual workflow dispatch → Custom environment deployment

**Workflow Jobs:**

1. **Pre-deployment Validation**
   - Validates deployment conditions
   - Determines deployment version
   - Checks branch and environment compatibility

2. **Build and Push Image**
   - Multi-platform Docker builds (amd64, arm64)
   - Container registry integration (GHCR.io)
   - Image tagging and metadata

3. **Deploy to Staging**
   - AWS ECS service update
   - Health checks and smoke tests
   - Deployment notifications

4. **Deploy to Production**
   - Production ECS deployment
   - Comprehensive health checks
   - Release creation and notifications

5. **Rollback Capability**
   - Automated rollback procedures
   - Previous version restoration
   - Rollback verification

### Workflow Configuration

```yaml
# Example workflow dispatch inputs
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
    force_deploy:
      description: "Force deployment even if tests fail"
      required: false
      default: false
      type: boolean
    version:
      description: "Specific version to deploy (optional)"
      required: false
      type: string
```

## Deployment Scripts

### Main Deployment Script

The `scripts/deploy.sh` script provides comprehensive deployment capabilities:

**Features:**
- Multi-environment support (staging, production)
- Docker image building and pushing
- AWS ECS and Kubernetes deployment
- Health checks and smoke tests
- Deployment records and notifications
- Dry-run mode for testing

**Usage Examples:**

```bash
# Deploy to staging
./scripts/deploy.sh -e staging

# Deploy specific version to production
./scripts/deploy.sh -e production -v v1.2.3

# Force deploy to staging
./scripts/deploy.sh -e staging -f

# Dry run for production
./scripts/deploy.sh -e production --dry-run

# Skip tests and deploy
./scripts/deploy.sh -e staging -s
```

**Script Options:**

```bash
Options:
    -e, --environment ENV     Environment to deploy to (staging|production)
    -v, --version VERSION     Specific version to deploy (optional)
    -f, --force              Force deployment even if tests fail
    -s, --skip-tests         Skip running tests before deployment
    -d, --dry-run            Show what would be deployed without actually deploying
    -h, --help               Show this help message
```

### Deployment Process

1. **Prerequisites Check**
   - Docker installation verification
   - AWS CLI configuration
   - Kubernetes tools (optional)

2. **Version Management**
   - Git commit hash as default version
   - Custom version specification
   - Version validation

3. **Pre-deployment Testing**
   - Unit test execution
   - Integration test validation
   - Force deployment override

4. **Image Building**
   - Multi-stage Docker builds
   - Build argument injection
   - Image tagging strategy

5. **Image Publishing**
   - Container registry authentication
   - Multi-tag pushing
   - Build cache optimization

6. **Environment Deployment**
   - AWS ECS service updates
   - Kubernetes deployment (if available)
   - Service health monitoring

7. **Post-deployment Validation**
   - Health check execution
   - Smoke test running
   - Endpoint validation

8. **Record Keeping**
   - Deployment metadata
   - Audit trail creation
   - Notification sending

## AWS ECS Deployment

### ECS Task Definition

The deployment system uses ECS task definitions for container orchestration:

```json
{
  "family": "kyb-platform-api",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "512",
  "memory": "1024",
  "executionRoleArn": "arn:aws:iam::ACCOUNT_ID:role/ecsTaskExecutionRole",
  "taskRoleArn": "arn:aws:iam::ACCOUNT_ID:role/kyb-platform-task-role",
  "containerDefinitions": [
    {
      "name": "kyb-platform-api",
      "image": "ghcr.io/REPOSITORY/kyb-platform:latest",
      "essential": true,
      "portMappings": [
        {
          "containerPort": 8080,
          "protocol": "tcp"
        }
      ],
      "environment": [...],
      "secrets": [...],
      "logConfiguration": {...},
      "healthCheck": {...}
    }
  ]
}
```

### ECS Service Configuration

**Service Properties:**
- **Cluster**: Environment-specific ECS clusters
- **Service Type**: Fargate with auto-scaling
- **Load Balancer**: Application Load Balancer integration
- **Health Checks**: Container and ALB health monitoring
- **Auto Scaling**: CPU and memory-based scaling policies

**Deployment Strategy:**
- **Rolling Update**: Zero-downtime deployments
- **Health Check Grace Period**: 60 seconds
- **Maximum Percent**: 200% (allows 2x capacity during deployment)
- **Minimum Healthy Percent**: 100% (ensures availability)

### ECS Deployment Commands

```bash
# Update ECS service
aws ecs update-service \
  --cluster kyb-platform-production \
  --service kyb-platform-api \
  --force-new-deployment

# Wait for deployment completion
aws ecs wait services-stable \
  --cluster kyb-platform-production \
  --services kyb-platform-api

# Check service status
aws ecs describe-services \
  --cluster kyb-platform-production \
  --services kyb-platform-api
```

## Kubernetes Deployment

### Deployment Manifests

The system includes comprehensive Kubernetes manifests:

**Deployment:**
- Rolling update strategy
- Resource limits and requests
- Health checks (liveness and readiness probes)
- Security context configuration

**Service:**
- ClusterIP service type
- Port mapping configuration
- Load balancer integration

**Ingress:**
- SSL/TLS termination
- Domain routing
- Rate limiting configuration

**Horizontal Pod Autoscaler:**
- CPU and memory-based scaling
- Scaling policies and behavior
- Minimum and maximum replica configuration

### Kubernetes Deployment Commands

```bash
# Apply deployment
kubectl apply -f deployments/kubernetes/deployment.yaml

# Update image
kubectl set image deployment/kyb-platform-api \
  kyb-platform-api=ghcr.io/REPOSITORY/kyb-platform:v1.2.3 \
  -n kyb-platform

# Check rollout status
kubectl rollout status deployment/kyb-platform-api -n kyb-platform

# Rollback deployment
kubectl rollout undo deployment/kyb-platform-api -n kyb-platform
```

## Environment Configuration

### Environment Variables

**Common Variables:**
```bash
ENVIRONMENT=production
LOG_LEVEL=info
DB_HOST=kyb-platform-db.cluster-xyz.us-east-1.rds.amazonaws.com
DB_PORT=5432
DB_NAME=kyb_platform
REDIS_HOST=kyb-platform-redis.xyz.cache.amazonaws.com
REDIS_PORT=6379
```

**Secret Management:**
- AWS Secrets Manager integration
- Kubernetes secrets
- Environment-specific secret rotation

### Configuration Management

**ECS Configuration:**
- Task definition environment variables
- Secrets Manager integration
- Parameter Store for configuration

**Kubernetes Configuration:**
- ConfigMaps for non-sensitive data
- Secrets for sensitive information
- Namespace-based configuration isolation

## Health Checks and Monitoring

### Health Check Endpoints

**Application Health:**
- `/health` - Basic health status
- `/status` - Detailed service status
- `/metrics` - Prometheus metrics
- `/ready` - Readiness probe endpoint

**Health Check Configuration:**
```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 60
  periodSeconds: 30
  timeoutSeconds: 5
  failureThreshold: 3

readinessProbe:
  httpGet:
    path: /ready
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3
```

### Smoke Tests

**Post-deployment Validation:**
```bash
# Health endpoint test
curl -f https://api.kybplatform.com/health

# Status endpoint test
curl -f https://api.kybplatform.com/status

# Metrics endpoint test
curl -f https://api.kybplatform.com/metrics

# API functionality test
curl -f https://api.kybplatform.com/v1/classify \
  -H "Content-Type: application/json" \
  -d '{"business_name": "Test Corp"}'
```

## Security and Compliance

### Container Security

**Security Context:**
- Non-root user execution
- Read-only root filesystem
- Dropped capabilities
- Security context configuration

**Image Security:**
- Multi-stage builds for minimal attack surface
- Base image vulnerability scanning
- Regular security updates
- Image signing and verification

### Access Control

**AWS IAM Roles:**
- ECS task execution role
- ECS task role for application permissions
- Least privilege principle
- Role-based access control

**Kubernetes RBAC:**
- Service account configuration
- Role and role binding definitions
- Namespace isolation
- Pod security policies

## Rollback Procedures

### Automated Rollback

**Rollback Triggers:**
- Health check failures
- Smoke test failures
- Performance degradation
- Manual rollback initiation

**Rollback Process:**
```bash
# Get previous task definition
PREVIOUS_TASK_DEF=$(aws ecs describe-services \
  --cluster kyb-platform-production \
  --services kyb-platform-api \
  --query 'services[0].taskDefinition' \
  --output text)

# Rollback to previous version
aws ecs update-service \
  --cluster kyb-platform-production \
  --service kyb-platform-api \
  --task-definition $PREVIOUS_TASK_DEF \
  --force-new-deployment
```

### Manual Rollback

**GitHub Actions Rollback:**
- Workflow dispatch with rollback flag
- Previous version identification
- Automated rollback execution
- Rollback verification

**Kubernetes Rollback:**
```bash
# Rollback to previous revision
kubectl rollout undo deployment/kyb-platform-api -n kyb-platform

# Check rollback status
kubectl rollout status deployment/kyb-platform-api -n kyb-platform

# View rollback history
kubectl rollout history deployment/kyb-platform-api -n kyb-platform
```

## Monitoring and Alerting

### Deployment Monitoring

**Metrics Collection:**
- Deployment success/failure rates
- Deployment duration tracking
- Rollback frequency monitoring
- Environment health status

**Alerting Rules:**
- Deployment failure notifications
- Health check failure alerts
- Performance degradation warnings
- Security vulnerability alerts

### Logging and Tracing

**Log Aggregation:**
- AWS CloudWatch integration
- Structured logging format
- Log retention policies
- Log analysis and alerting

**Distributed Tracing:**
- OpenTelemetry integration
- Request tracing across services
- Performance bottleneck identification
- Error correlation and debugging

## Best Practices

### Deployment Best Practices

1. **Blue-Green Deployments**
   - Zero-downtime deployments
   - Quick rollback capability
   - Traffic switching strategies

2. **Canary Deployments**
   - Gradual traffic shifting
   - Performance monitoring
   - Automatic rollback on issues

3. **Immutable Infrastructure**
   - Image-based deployments
   - No runtime configuration changes
   - Version-controlled infrastructure

4. **Infrastructure as Code**
   - Terraform/CloudFormation templates
   - Version-controlled configurations
   - Automated infrastructure provisioning

### Security Best Practices

1. **Secret Management**
   - AWS Secrets Manager integration
   - Kubernetes secrets encryption
   - Regular secret rotation
   - Access audit logging

2. **Network Security**
   - VPC isolation
   - Security group configuration
   - Network policies (Kubernetes)
   - SSL/TLS termination

3. **Container Security**
   - Base image scanning
   - Runtime security monitoring
   - Vulnerability management
   - Security context configuration

### Performance Best Practices

1. **Resource Optimization**
   - Right-sizing containers
   - Auto-scaling configuration
   - Resource monitoring
   - Performance benchmarking

2. **Caching Strategies**
   - Application-level caching
   - CDN integration
   - Database query optimization
   - Static asset caching

## Troubleshooting

### Common Issues

1. **Deployment Failures**
   - Check ECS service logs
   - Verify task definition
   - Validate environment variables
   - Check resource constraints

2. **Health Check Failures**
   - Verify application startup
   - Check endpoint availability
   - Validate configuration
   - Review application logs

3. **Performance Issues**
   - Monitor resource usage
   - Check auto-scaling policies
   - Analyze application metrics
   - Review database performance

### Debugging Commands

```bash
# Check ECS service status
aws ecs describe-services --cluster kyb-platform-production --services kyb-platform-api

# View container logs
aws logs get-log-events --log-group-name /ecs/kyb-platform-api --log-stream-name <stream-name>

# Check Kubernetes pod status
kubectl get pods -n kyb-platform
kubectl describe pod <pod-name> -n kyb-platform
kubectl logs <pod-name> -n kyb-platform

# Test connectivity
curl -v https://api.kybplatform.com/health
telnet api.kybplatform.com 443
```

## Future Enhancements

### Planned Improvements

1. **Advanced Deployment Strategies**
   - Blue-green deployment automation
   - Canary deployment implementation
   - A/B testing capabilities
   - Feature flag integration

2. **Multi-Cloud Support**
   - Google Cloud Platform integration
   - Azure Container Instances
   - Hybrid cloud deployments
   - Cloud-agnostic configurations

3. **Advanced Monitoring**
   - Real-time deployment monitoring
   - Predictive failure detection
   - Automated performance optimization
   - Intelligent rollback decisions

4. **Security Enhancements**
   - Automated security scanning
   - Compliance validation
   - Threat detection
   - Security policy enforcement

---

This documentation provides a comprehensive overview of the KYB Platform's deployment automation system. For specific implementation details, refer to the deployment scripts and configuration files referenced throughout this document.
