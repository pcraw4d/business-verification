# Risk Assessment Service - Kubernetes Deployment

This directory contains Kubernetes manifests for deploying the Risk Assessment Service in a cloud-native environment with auto-scaling capabilities.

## 🏗️ Architecture Overview

The deployment includes:
- **Deployment**: Main application deployment with 3 replicas
- **Service**: LoadBalancer and ClusterIP services for external and internal access
- **HPA**: Horizontal Pod Autoscaler for automatic scaling based on CPU, memory, and custom metrics
- **Ingress**: NGINX ingress controller for external access with SSL termination
- **ConfigMap**: Environment configuration and model metadata
- **Secrets**: Sensitive configuration (database URLs, API keys, etc.)
- **Network Policies**: Security policies for network traffic
- **RBAC**: Role-based access control for service accounts
- **PDB**: Pod Disruption Budget for high availability

## 📋 Prerequisites

- Kubernetes cluster (v1.20+)
- kubectl configured to access the cluster
- NGINX Ingress Controller installed
- cert-manager for SSL certificates (optional)
- Prometheus and Grafana for monitoring (optional)

## 🚀 Quick Start

### 1. Deploy the Service

```bash
# Deploy with default settings
./deploy.sh --deploy

# Deploy with specific image tag
./deploy.sh --deploy --tag v1.0.0

# Deploy to specific namespace
./deploy.sh --deploy --namespace my-namespace
```

### 2. Check Deployment Status

```bash
# Check overall status
./deploy.sh --status

# Check pods
kubectl get pods -n kyb-platform -l app=risk-assessment-service

# Check services
kubectl get services -n kyb-platform -l app=risk-assessment-service

# Check HPA
kubectl get hpa -n kyb-platform
```

### 3. Access the Service

```bash
# Get service endpoints
kubectl get ingress -n kyb-platform

# Port forward for local testing
kubectl port-forward -n kyb-platform service/risk-assessment-service 8080:80
```

## 🔧 Configuration

### Environment Variables

The service is configured through ConfigMap and Secrets:

- **ConfigMap**: Non-sensitive configuration (ports, timeouts, feature flags)
- **Secrets**: Sensitive data (database URLs, API keys, certificates)

### Scaling Configuration

The HPA is configured to scale based on:
- **CPU**: Target 70% utilization
- **Memory**: Target 80% utilization
- **Custom Metrics**: HTTP requests per second, risk assessments per second

Scaling limits:
- **Min Replicas**: 3
- **Max Replicas**: 20

### Resource Limits

- **CPU Request**: 250m
- **CPU Limit**: 1000m
- **Memory Request**: 512Mi
- **Memory Limit**: 2Gi

## 📊 Monitoring

### Health Checks

- **Liveness Probe**: `/health` endpoint
- **Readiness Probe**: `/ready` endpoint
- **Startup Probe**: `/health` endpoint with longer timeout

### Metrics

- **Prometheus Metrics**: Available at `/metrics` on port 9090
- **Custom Metrics**: Risk assessment metrics for HPA scaling
- **ServiceMonitor**: For Prometheus scraping

### Logging

- **Log Level**: Configurable via ConfigMap
- **Log Format**: JSON structured logging
- **Log Output**: stdout (collected by cluster logging)

## 🔒 Security

### Network Policies

- Ingress traffic allowed from:
  - NGINX Ingress Controller
  - Monitoring namespace
  - Other services in the same namespace
- Egress traffic allowed to:
  - Database services
  - Redis cache
  - External APIs (HTTPS)
  - DNS resolution

### RBAC

- **Service Account**: `risk-assessment-service`
- **Role**: Limited permissions for ConfigMaps, Secrets, and Pods
- **Cluster Role**: Access to nodes, namespaces, and metrics

### Pod Security

- **Non-root user**: Runs as user 1000
- **Read-only root filesystem**: Security best practice
- **Security Context**: Restricted capabilities

## 🔄 Auto-scaling

### Horizontal Pod Autoscaler

The HPA automatically scales the deployment based on:

1. **CPU Utilization**: Scales up when CPU > 70%
2. **Memory Utilization**: Scales up when memory > 80%
3. **HTTP Requests**: Scales up when requests > 100/sec per pod
4. **Risk Assessments**: Scales up when assessments > 50/sec per pod

### Scaling Behavior

- **Scale Up**: Aggressive scaling (50% increase or 4 pods max)
- **Scale Down**: Conservative scaling (10% decrease or 2 pods max)
- **Stabilization**: 60s for scale up, 300s for scale down

## 🛠️ Maintenance

### Rolling Updates

```bash
# Update image
kubectl set image deployment/risk-assessment-service \
  risk-assessment-service=risk-assessment-service:v1.1.0 \
  -n kyb-platform

# Check rollout status
kubectl rollout status deployment/risk-assessment-service -n kyb-platform

# Rollback if needed
kubectl rollout undo deployment/risk-assessment-service -n kyb-platform
```

### Scaling

```bash
# Manual scaling
kubectl scale deployment/risk-assessment-service --replicas=5 -n kyb-platform

# Update HPA
kubectl patch hpa risk-assessment-service-hpa -n kyb-platform -p '{"spec":{"maxReplicas":30}}'
```

### Configuration Updates

```bash
# Update ConfigMap
kubectl apply -f configmap.yaml -n kyb-platform

# Restart deployment to pick up changes
kubectl rollout restart deployment/risk-assessment-service -n kyb-platform
```

## 🧹 Cleanup

```bash
# Remove all resources
./deploy.sh --cleanup

# Or manually
kubectl delete namespace kyb-platform
```

## 📝 Customization

### Custom Metrics

To add custom metrics for HPA scaling:

1. Implement metrics in the application
2. Expose metrics at `/metrics` endpoint
3. Update HPA configuration with new metrics
4. Ensure Prometheus is scraping the metrics

### Resource Requirements

To adjust resource requirements:

1. Update `deployment.yaml` resource requests/limits
2. Update `hpa.yaml` scaling thresholds
3. Update `namespace.yaml` resource quotas if needed

### Network Policies

To modify network access:

1. Update `network-policies.yaml`
2. Test connectivity after changes
3. Ensure security requirements are met

## 🐛 Troubleshooting

### Common Issues

1. **Pods not starting**:
   ```bash
   kubectl describe pod <pod-name> -n kyb-platform
   kubectl logs <pod-name> -n kyb-platform
   ```

2. **Service not accessible**:
   ```bash
   kubectl get endpoints -n kyb-platform
   kubectl describe service risk-assessment-service -n kyb-platform
   ```

3. **HPA not scaling**:
   ```bash
   kubectl describe hpa risk-assessment-service-hpa -n kyb-platform
   kubectl get --raw /apis/metrics.k8s.io/v1beta1/pods
   ```

4. **Ingress issues**:
   ```bash
   kubectl describe ingress risk-assessment-service-ingress -n kyb-platform
   kubectl logs -n ingress-nginx <ingress-controller-pod>
   ```

### Debug Commands

```bash
# Check all resources
kubectl get all -n kyb-platform

# Check events
kubectl get events -n kyb-platform --sort-by='.lastTimestamp'

# Check resource usage
kubectl top pods -n kyb-platform
kubectl top nodes

# Check network policies
kubectl get networkpolicies -n kyb-platform
```

## 📚 Additional Resources

- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [NGINX Ingress Controller](https://kubernetes.github.io/ingress-nginx/)
- [Prometheus Operator](https://prometheus-operator.dev/)
- [cert-manager](https://cert-manager.io/docs/)

## 🤝 Support

For issues or questions:
- Check the troubleshooting section above
- Review Kubernetes cluster logs
- Contact the platform team
