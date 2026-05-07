# ADR-002: In-cluster Kubernetes client with ServiceAccount RBAC

> **Status**: Accepted  
> **Date**: 2026-05-06  
> **Related Tickets**: [KIVP-331](https://jira.example.local/browse/KIVP-331)

#### Context

configuration-service needs to read ConfigMaps from the Kubernetes API. Access can be granted in several ways:

1. **In-cluster ServiceAccount** — pod uses the mounted token; RBAC grants `get` on ConfigMaps in the target namespace.
2. **KubeConfig file** — external credentials mounted as a Secret; works outside the cluster too.
3. **Environment-variable token** — manual token injection.

#### Decision

Use the **in-cluster ServiceAccount** pattern exclusively. The Helm chart creates a dedicated ServiceAccount, ClusterRole (scoped to `configmaps` resource with `get` verb), and ClusterRoleBinding. The `client-go` library picks up the mounted token automatically when running inside the cluster.

For local Docker Compose development, ConfigMap reading is disabled and a mock config is returned, or the developer points the service at a real cluster with a kubeconfig override.

#### Consequences

| Positive | Negative |
|---------|----------|
| ✅ Zero secrets to manage — token is automatically rotated by Kubernetes | ❌ Service cannot run outside the cluster without additional kubeconfig setup |
| ✅ Least-privilege — only `get configmaps` in one namespace | ❌ ServiceAccount binding must be updated if ConfigMap moves to a different namespace |
| ✅ Standard Kubernetes pattern — works with any RBAC-compliant cluster | |

#### Next Steps

- [x] ServiceAccount and ClusterRoleBinding in Helm chart
- [x] Verified access in local Minikube
- [ ] Scope RoleBinding to a specific namespace (ClusterRoleBinding is overly broad)
