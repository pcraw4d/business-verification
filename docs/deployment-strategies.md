# Deployment Strategies for KYB Platform

## Overview

This document outlines comprehensive deployment strategies for the KYB Platform, ensuring reliable, scalable, and secure deployments across different environments.

## Deployment Patterns

### 1. Blue-Green Deployment

#### Implementation
```yaml
# Blue-Green deployment configuration
apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  name: kyb-platform
spec:
  replicas: 5
  strategy:
    blueGreen:
      activeService: kyb-platform-active
      previewService: kyb-platform-preview
      autoPromotionEnabled: false
      scaleDownDelaySeconds: 30
      prePromotionAnalysis:
        templates:
        - templateName: success-rate
        args:
        - name: service-name
          value: kyb-platform-preview
      postPromotionAnalysis:
        templates:
        - templateName: success-rate
        args:
        - name: service-name
          value: kyb-platform-active
```

### 2. Canary Deployment

#### Configuration
```yaml
# Canary deployment with traffic splitting
apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  name: kyb-platform-canary
spec:
  replicas: 10
  strategy:
    canary:
      steps:
      - setWeight: 10
      - pause: {duration: 5m}
      - setWeight: 20
      - pause: {duration: 10m}
      - setWeight: 50
      - pause: {duration: 15m}
      - setWeight: 100
      analysis:
        templates:
        - templateName: error-rate
        args:
        - name: service-name
          value: kyb-platform
```

### 3. Rolling Deployment

#### Kubernetes Configuration
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kyb-platform
spec:
  replicas: 5
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxUnavailable: 1
      maxSurge: 2
  template:
    spec:
      containers:
      - name: kyb-platform
        image: kyb-platform:latest
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 5
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
```

## CI/CD Pipeline Configuration

### GitHub Actions Workflow
```yaml
name: Deploy KYB Platform
on:
  push:
    branches: [main, develop]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v4
    
    - name: Setup Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.22'
    
    - name: Run tests
      run: |
        go test -v -race -coverprofile=coverage.out ./...
        
    - name: Code quality check
      run: |
        ./bin/validate-quality --project . --alerts --severity critical
    
    - name: Security scan
      run: |
        go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
        gosec ./...

  build:
    needs: test
    runs-on: ubuntu-latest
    outputs:
      image: ${{ steps.image.outputs.image }}
      digest: ${{ steps.build.outputs.digest }}
    
    steps:
    - name: Checkout
      uses: actions/checkout@v4
    
    - name: Login to Container Registry
      uses: docker/login-action@v3
      with:
        registry: ${{ env.REGISTRY }}
        username: ${{ github.actor }}
        password: ${{ secrets.GITHUB_TOKEN }}
    
    - name: Extract metadata
      id: meta
      uses: docker/metadata-action@v5
      with:
        images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
        tags: |
          type=ref,event=branch
          type=ref,event=pr
          type=sha,prefix={{branch}}-
    
    - name: Build and push
      id: build
      uses: docker/build-push-action@v5
      with:
        context: .
        push: true
        tags: ${{ steps.meta.outputs.tags }}
        labels: ${{ steps.meta.outputs.labels }}

  deploy-staging:
    if: github.ref == 'refs/heads/develop'
    needs: build
    runs-on: ubuntu-latest
    environment: staging
    
    steps:
    - name: Deploy to staging
      uses: azure/k8s-deploy@v1
      with:
        manifests: |
          k8s/staging/deployment.yaml
          k8s/staging/service.yaml
        images: ${{ needs.build.outputs.image }}
        
    - name: Run integration tests
      run: |
        kubectl wait --for=condition=available deployment/kyb-platform-staging
        ./scripts/integration-tests.sh staging

  deploy-production:
    if: github.ref == 'refs/heads/main'
    needs: [build, deploy-staging]
    runs-on: ubuntu-latest
    environment: production
    
    steps:
    - name: Deploy to production (Blue-Green)
      uses: azure/k8s-deploy@v1
      with:
        strategy: blue-green
        manifests: |
          k8s/production/deployment.yaml
          k8s/production/service.yaml
        images: ${{ needs.build.outputs.image }}
        
    - name: Monitor deployment
      run: |
        ./scripts/monitor-deployment.sh production 300
```

## Environment Configuration

### Development Environment
```yaml
# dev/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: kyb-dev

resources:
- ../base

patchesStrategicMerge:
- deployment-patch.yaml
- service-patch.yaml

configMapGenerator:
- name: kyb-config
  files:
  - config.yaml

secretGenerator:
- name: kyb-secrets
  files:
  - secrets.yaml

images:
- name: kyb-platform
  newTag: dev-latest
```

### Staging Environment
```yaml
# staging/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: kyb-staging

resources:
- ../base

patchesStrategicMerge:
- deployment-patch.yaml
- hpa-patch.yaml

replicas:
- name: kyb-platform
  count: 3

images:
- name: kyb-platform
  newTag: staging-v2.1.0
```

### Production Environment
```yaml
# production/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: kyb-production

resources:
- ../base
- pdb.yaml
- networkpolicy.yaml

patchesStrategicMerge:
- deployment-patch.yaml
- service-patch.yaml
- hpa-patch.yaml

replicas:
- name: kyb-platform
  count: 10

images:
- name: kyb-platform
  newTag: v2.1.0
```

## Monitoring and Alerting

### Deployment Monitoring
```yaml
# Prometheus alerts for deployments
groups:
- name: deployment-alerts
  rules:
  - alert: DeploymentFailed
    expr: kube_deployment_status_replicas_unavailable > 0
    for: 5m
    labels:
      severity: critical
    annotations:
      summary: "Deployment has unavailable replicas"
      description: "{{ $labels.deployment }} has {{ $value }} unavailable replicas"

  - alert: HighErrorRate
    expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
    for: 2m
    labels:
      severity: warning
    annotations:
      summary: "High error rate detected"
      description: "Error rate is {{ $value }} requests/second"

  - alert: HighLatency
    expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 0.5
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "High latency detected"
      description: "95th percentile latency is {{ $value }}s"
```

### Health Check Configuration
```go
// Health check implementation
func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    status := &HealthStatus{
        Status:    "healthy",
        Timestamp: time.Now(),
        Version:   h.version,
        Checks:    make(map[string]CheckResult),
    }
    
    // Database health check
    if err := h.checkDatabase(ctx); err != nil {
        status.Status = "unhealthy"
        status.Checks["database"] = CheckResult{
            Status:  "unhealthy",
            Message: err.Error(),
        }
    } else {
        status.Checks["database"] = CheckResult{
            Status: "healthy",
        }
    }
    
    // External services health check
    if err := h.checkExternalServices(ctx); err != nil {
        status.Status = "unhealthy"
        status.Checks["external_services"] = CheckResult{
            Status:  "unhealthy",
            Message: err.Error(),
        }
    } else {
        status.Checks["external_services"] = CheckResult{
            Status: "healthy",
        }
    }
    
    httpStatus := http.StatusOK
    if status.Status == "unhealthy" {
        httpStatus = http.StatusServiceUnavailable
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(httpStatus)
    json.NewEncoder(w).Encode(status)
}
```

## Rollback Procedures

### Automatic Rollback
```yaml
# ArgoCD rollback configuration
apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  name: kyb-platform
spec:
  strategy:
    canary:
      analysis:
        templates:
        - templateName: error-rate-analysis
        args:
        - name: service-name
          value: kyb-platform
        - name: error-threshold
          value: "0.05"
      abortScaleDownDelaySeconds: 30
      maxUnavailable: 0
```

### Manual Rollback Script
```bash
#!/bin/bash
# rollback.sh - Manual rollback script

set -euo pipefail

NAMESPACE=${1:-kyb-production}
PREVIOUS_VERSION=${2:-}

echo "Starting rollback for namespace: $NAMESPACE"

if [ -z "$PREVIOUS_VERSION" ]; then
    # Get previous successful deployment
    PREVIOUS_VERSION=$(kubectl rollout history deployment/kyb-platform -n $NAMESPACE | tail -2 | head -1 | awk '{print $1}')
fi

echo "Rolling back to version: $PREVIOUS_VERSION"

# Perform rollback
kubectl rollout undo deployment/kyb-platform --to-revision=$PREVIOUS_VERSION -n $NAMESPACE

# Wait for rollback to complete
kubectl rollout status deployment/kyb-platform -n $NAMESPACE --timeout=300s

# Verify health
echo "Verifying health after rollback..."
sleep 30
kubectl get pods -n $NAMESPACE -l app=kyb-platform

# Run smoke tests
./scripts/smoke-tests.sh $NAMESPACE

echo "Rollback completed successfully"
```

## Security Considerations

### Image Security
```dockerfile
# Dockerfile with security best practices
FROM golang:1.22-alpine AS builder

# Add security updates
RUN apk update && apk add --no-cache ca-certificates git

# Create non-root user
RUN adduser -D -s /bin/sh appuser

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM scratch

# Copy certificates and user
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd

# Copy application
COPY --from=builder /app/main /main

# Use non-root user
USER appuser

EXPOSE 8080
ENTRYPOINT ["/main"]
```

### Pod Security Policy
```yaml
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: kyb-platform-psp
spec:
  privileged: false
  allowPrivilegeEscalation: false
  requiredDropCapabilities:
    - ALL
  volumes:
    - 'configMap'
    - 'emptyDir'
    - 'projected'
    - 'secret'
    - 'downwardAPI'
    - 'persistentVolumeClaim'
  runAsUser:
    rule: 'MustRunAsNonRoot'
  seLinux:
    rule: 'RunAsAny'
  fsGroup:
    rule: 'RunAsAny'
```

## Performance Optimization

### Resource Management
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kyb-platform
spec:
  template:
    spec:
      containers:
      - name: kyb-platform
        resources:
          requests:
            memory: "256Mi"
            cpu: "250m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        env:
        - name: GOMAXPROCS
          valueFrom:
            resourceFieldRef:
              resource: limits.cpu
        - name: GOMEMLIMIT
          valueFrom:
            resourceFieldRef:
              resource: limits.memory
```

### Horizontal Pod Autoscaler
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: kyb-platform-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: kyb-platform
  minReplicas: 3
  maxReplicas: 20
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  behavior:
    scaleUp:
      stabilizationWindowSeconds: 60
      policies:
      - type: Percent
        value: 100
        periodSeconds: 60
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 10
        periodSeconds: 60
```

## Disaster Recovery

### Backup Strategy
```yaml
# Velero backup configuration
apiVersion: velero.io/v1
kind: Schedule
metadata:
  name: kyb-platform-backup
spec:
  schedule: "0 2 * * *"  # Daily at 2 AM
  template:
    includedNamespaces:
    - kyb-production
    - kyb-staging
    excludedResources:
    - pods
    - replicasets
    storageLocation: default
    ttl: 720h  # 30 days
```

### Disaster Recovery Plan
```bash
#!/bin/bash
# disaster-recovery.sh

set -euo pipefail

BACKUP_DATE=${1:-latest}
TARGET_CLUSTER=${2:-disaster-recovery}

echo "Starting disaster recovery procedure..."
echo "Backup date: $BACKUP_DATE"
echo "Target cluster: $TARGET_CLUSTER"

# Switch to DR cluster
kubectl config use-context $TARGET_CLUSTER

# Restore from backup
velero restore create kyb-dr-restore-$(date +%Y%m%d-%H%M%S) \
    --from-backup kyb-platform-backup-$BACKUP_DATE \
    --wait

# Verify restoration
kubectl get pods -n kyb-production
kubectl get services -n kyb-production

# Update DNS to point to DR cluster
echo "Manual step: Update DNS records to point to DR cluster"
echo "DR cluster LoadBalancer IP: $(kubectl get service kyb-platform -n kyb-production -o jsonpath='{.status.loadBalancer.ingress[0].ip}')"

echo "Disaster recovery completed successfully"
```

## Deployment Checklist

### Pre-Deployment
```yaml
Pre-Deployment Checklist:
  Code Quality:
    - ✓ All tests passing
    - ✓ Code review completed
    - ✓ Security scan passed
    - ✓ Performance tests passed
    
  Infrastructure:
    - ✓ Infrastructure as code updated
    - ✓ Database migrations ready
    - ✓ Configuration validated
    - ✓ Secrets updated
    
  Monitoring:
    - ✓ Alerts configured
    - ✓ Dashboards updated
    - ✓ Runbooks prepared
    - ✓ Oncall schedule confirmed
    
  Communication:
    - ✓ Stakeholders notified
    - ✓ Maintenance window scheduled
    - ✓ Rollback plan reviewed
    - ✓ Team availability confirmed
```

### Post-Deployment
```yaml
Post-Deployment Checklist:
  Verification:
    - ✓ Health checks passing
    - ✓ Key functionality tested
    - ✓ Performance metrics normal
    - ✓ Error rates acceptable
    
  Monitoring:
    - ✓ Alerts functioning
    - ✓ Metrics collecting
    - ✓ Logs flowing
    - ✓ Dashboards updated
    
  Communication:
    - ✓ Deployment status communicated
    - ✓ Documentation updated
    - ✓ Lessons learned documented
    - ✓ Next deployment planned
```

---

**Document Version**: 1.0.0  
**Last Updated**: August 19, 2025  
**Next Review**: November 19, 2025
