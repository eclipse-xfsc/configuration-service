# Configuration Service Helm Deployment Guide

This guide covers deploying the configuration-service to Kubernetes using Helm.

## Prerequisites

- Kubernetes cluster (1.20+)
- Helm 3.x installed
- kubectl configured to access your cluster
- Container image registry access (for pulling `configuration-service` image)

## Quick Start

### 1. Build Helm Dependencies

```bash
cd deployment/helm
helm dependency build
```

### 2. Install the Chart

**Using default values:**
```bash
helm install configuration-service deployment/helm \
  --namespace configuration-service \
  --create-namespace
```

**Using custom values file:**
```bash
helm install configuration-service deployment/helm \
  --namespace configuration-service \
  --create-namespace \
  --values my-values.yaml
```

### 3. Verify Installation

```bash
# Check deployment status
kubectl get deployment -n configuration-service

# Check pod status
kubectl get pods -n configuration-service

# View logs
kubectl logs -n configuration-service -f deployment/configuration-service
```

## Configuration

### Image Configuration

Edit `values.yaml` to configure the container image:

```yaml
image:
  repository: your-registry.io              # Container registry
  name: configuration-service              # Image name
  tag: v1.0.0                              # Image tag (defaults to Chart.AppVersion if empty)
  pullPolicy: IfNotPresent                 # Always, IfNotPresent, or Never
  pullSecrets: your-pull-secret            # Secret name for private registries
```

### Service Configuration

```yaml
service:
  type: ClusterIP                          # ClusterIP, NodePort, or LoadBalancer
  port: 8080                               # Service port

server:
  http:
    host: ""                               # Bind host (empty = all interfaces)
    port: 8080                             # Container port
```

### ConfigMap Data

Add configuration data in `values.yaml`:

```yaml
data:
  myconfig.json: '{"key": "value"}'
  another.yaml: |
    key: value
```

This data will be mounted in a ConfigMap named `configuration-service`.

### Ingress Configuration

To expose the service via ingress, update `values.yaml`:

```yaml
ingress:
  enabled: true
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
  hosts:
    - host: config.example.io
      paths:
        - path: /
          port: 8080
  tls:
    - secretName: config-tls
      hosts:
        - config.example.io
```

Then upgrade the release:
```bash
helm upgrade configuration-service deployment/helm \
  --namespace configuration-service \
  --values values.yaml
```

### Resource Limits

Configure CPU and memory requests/limits:

```yaml
resources:
  requests:
    cpu: 25m                               # Minimum CPU
    memory: 64Mi                           # Minimum memory
  limits:
    cpu: 150m                              # Maximum CPU
    memory: 128Mi                          # Maximum memory
```

### Replicas

```yaml
replicaCount: 1                            # Number of pod replicas
```

## RBAC (Role-Based Access Control)

The chart automatically creates:
- **ServiceAccount**: `configuration-service`
- **Role**: Allows reading ConfigMaps named `configuration-service` in the deployment namespace
- **RoleBinding**: Grants the role to the service account

The service requires read access to ConfigMaps to function properly.

## Health Checks

The deployment includes readiness probes:
- **Endpoint**: `GET /health`
- **Initial Delay**: 5 seconds
- **Period**: 5 seconds
- **Success Threshold**: 2
- **Failure Threshold**: 2

## Monitoring & Observability

### Prometheus Metrics

The service exposes metrics at `GET /metrics` (port 8080).

Pod annotations for Prometheus discovery:
```yaml
podAnnotations:
  prometheus.io/scrape: "true"
  prometheus.io/path: "/metrics"
  prometheus.io/port: "8080"
```

### Custom Metrics

- `configuration_service_http_requests_total{method,route,status}`
- `configuration_service_http_request_duration_seconds{method,route,status}`
- `configuration_service_http_in_flight_requests`

### Grafana Dashboards

Dashboards are available in `grafana/provisioning/dashboards/`.

## Common Deployment Scenarios

### Scenario 1: Development Environment

```yaml
# values-dev.yaml
replicaCount: 1
image:
  tag: latest
  pullPolicy: Always
resources:
  requests:
    cpu: 10m
    memory: 32Mi
  limits:
    cpu: 100m
    memory: 64Mi
```

Deploy:
```bash
helm install configuration-service deployment/helm \
  --namespace configuration-service \
  --create-namespace \
  --values values-dev.yaml
```

### Scenario 2: Production Environment

```yaml
# values-prod.yaml
replicaCount: 3
image:
  tag: v1.0.0
  pullPolicy: IfNotPresent
resources:
  requests:
    cpu: 50m
    memory: 128Mi
  limits:
    cpu: 200m
    memory: 256Mi
ingress:
  enabled: true
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
  hosts:
    - host: config.api.example.com
      paths:
        - path: /
          port: 8080
  tls:
    - secretName: config-tls
      hosts:
        - config.api.example.com
```

Deploy:
```bash
helm install configuration-service deployment/helm \
  --namespace configuration-service \
  --create-namespace \
  --values values-prod.yaml
```

## Upgrading

### Update Configuration

```bash
helm upgrade configuration-service deployment/helm \
  --namespace configuration-service \
  --values values.yaml
```

### Update Image

```bash
helm upgrade configuration-service deployment/helm \
  --namespace configuration-service \
  --set image.tag=v1.1.0
```

### Rollback

```bash
helm rollback configuration-service \
  --namespace configuration-service
```

## Uninstalling

```bash
helm uninstall configuration-service \
  --namespace configuration-service

# Clean up namespace
kubectl delete namespace configuration-service
```

## Troubleshooting

### Pod won't start

```bash
# Check pod logs
kubectl logs -n configuration-service deployment/configuration-service

# Describe pod for events
kubectl describe pod -n configuration-service <pod-name>

# Check events
kubectl get events -n configuration-service --sort-by='.lastTimestamp'
```

### ConfigMap not found

Ensure a ConfigMap named `configuration-service` exists in the deployment namespace:

```bash
# Create a sample ConfigMap
kubectl create configmap configuration-service \
  --from-literal=example.json='{"status":"ok"}' \
  -n configuration-service
```

### Image pull errors

```bash
# Verify image exists and is accessible
kubectl describe pod -n configuration-service <pod-name>

# Check image pull secret
kubectl get secrets -n configuration-service

# Verify registry credentials
kubectl describe secret <pull-secret-name> -n configuration-service
```

## Values File Reference

See `deployment/helm/values.yaml` for all available configuration options.

## Chart Files

- `Chart.yaml` - Chart metadata
- `values.yaml` - Default configuration values
- `templates/` - Kubernetes resource templates
  - `deployment.yaml` - Pod deployment
  - `service.yaml` - Kubernetes Service
  - `configmap.yaml` - Configuration data
  - `ingress.yaml` - Ingress rules
  - `service-account.yaml` - Service account
  - `role.yaml` - RBAC role
  - `role-binding.yaml` - RBAC role binding
