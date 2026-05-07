# Configuration Service - Helm Deployment Quick Reference

## Status: ✅ Production Ready

The configuration-service is now fully configured for Helm deployment.

## Chart Location
```
deployment/helm/
├── Chart.yaml              # Chart metadata (v2 with dependencies)
├── values.yaml             # Default configuration values
├── charts/                 # Chart dependencies (library)
└── templates/
    ├── deployment.yaml     # Pod deployment with health checks
    ├── service.yaml        # Kubernetes Service (ClusterIP)
    ├── configmap.yaml      # ConfigMap for configuration data
    ├── ingress.yaml        # Ingress for external access
    ├── service-account.yaml # ServiceAccount with namespace
    ├── role.yaml           # RBAC role for ConfigMap read access
    └── role-binding.yaml   # RoleBinding connecting SA to role
```

## Quick Deploy Commands

### Build dependencies
```bash
cd deployment/helm
helm dependency build
```

### Install (default namespace)
```bash
helm install configuration-service deployment/helm \
  --create-namespace \
  --namespace configuration-service
```

### Install with custom values
```bash
helm install configuration-service deployment/helm \
  --namespace configuration-service \
  --values my-values.yaml
```

### Upgrade
```bash
helm upgrade configuration-service deployment/helm \
  --namespace configuration-service
```

### Uninstall
```bash
helm uninstall configuration-service -n configuration-service
```

## Key Features

✅ **Health Checks**: Readiness probes on `/health` endpoint
✅ **RBAC**: Full role-based access control with minimal permissions
✅ **Monitoring**: Prometheus metrics annotation configured
✅ **ConfigMaps**: Dynamic configuration injection
✅ **Ingress**: Optional ingress support for external access
✅ **Rolling Updates**: RollingUpdate strategy with zero-downtime deployment
✅ **Resource Limits**: CPU/memory requests and limits configured
✅ **Service Account**: Dedicated service account with namespace isolation

## Configuration

All configuration is in `deployment/helm/values.yaml`:
- Image repository, tag, pull policy
- Replica count
- Resource requests/limits
- Service type and port
- Ingress settings (enable/host/TLS)
- ConfigMap data
- Pod annotations (Prometheus scraping)

## CI/CD Integration

✅ **Automated Pipeline**: `.github/workflows/publish.yml` triggers on releases
- Builds Docker image
- Builds and publishes Helm chart
- Uses shared workflows from eclipse-xfsc/dev-ops

## Documentation

- **Full Guide**: [README.DEPLOYMENT.md](README.DEPLOYMENT.md)
- **Chart README**: [deployment/helm/Chart.yaml](deployment/helm/Chart.yaml)
- **Default Values**: [deployment/helm/values.yaml](deployment/helm/values.yaml)

## Verification

Chart has been validated:
```bash
✅ helm lint deployment/helm --strict
✅ helm template configuration-service deployment/helm
✅ Dependency build successful
```

## Next Steps

1. **Update Image Registry**: Edit `values.yaml` image.repository for your registry
2. **Configure Ingress**: Uncomment and update ingress settings for external access
3. **Set ConfigMap Data**: Add your configuration data to values.yaml data section
4. **Test Deployment**: Run install command in non-prod environment first
5. **Enable Monitoring**: Verify Prometheus is scraping `/metrics` endpoint

## Support

For issues or questions, refer to:
- README.DEPLOYMENT.md for detailed troubleshooting
- Chart templates in deployment/helm/templates/
- Service documentation in README.md and README.OBSERVABILITY.md
