# ADR-001: ConfigMap-as-a-service pattern for mobile configuration

> **Status**: Accepted  
> **Date**: 2026-05-06  
> **Related Tickets**: [KIVP-331](https://jira.example.local/browse/KIVP-331)

#### Context

Mobile apps need runtime configuration (discovery API URL, release labels, feature flags) that varies between environments (development, staging, production). Baking these values into the app binary requires a new app release for every config change.

Kubernetes ConfigMaps are the natural place to store per-environment config in the cluster. However, mobile apps cannot read ConfigMaps directly.

#### Decision

Deploy **configuration-service**: a minimal Go HTTP server that reads a single named ConfigMap via the Kubernetes API and returns it as a JSON object. The service is configured at deploy time with `CONFIGMAP_NAME` and `CONFIGMAP_NAMESPACE` — it serves exactly one ConfigMap.

Mobile apps call `GET /configuration` on startup and cache the result. All environment-specific values are managed in the cluster, not the app.

#### Consequences

| Positive | Negative |
|---------|----------|
| ✅ Zero-downtime config changes — update the ConfigMap, no app release needed | ❌ Mobile app must handle `/configuration` outages gracefully (use last cached config) |
| ✅ Single source of truth per environment | ❌ All config is public — no secrets should be stored in the ConfigMap |
| ✅ Extremely simple service — no database, no state | ❌ One service instance per environment (or use namespaced ConfigMaps) |
| ✅ Easy to audit and version-control config values via GitOps | |

#### Next Steps

- [x] Service deployed and verified in local Minikube
- [x] Mobile simulator contract tests validate `/configuration` response shape
- [ ] Add response caching with `Cache-Control` headers to reduce ConfigMap API calls
- [ ] Consider read-through cache with configurable TTL for resilience during Kubernetes API disruptions
