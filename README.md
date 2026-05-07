# configuration-service

Lightweight Kubernetes-native configuration bridge for KIVP mobile clients. Reads a single ConfigMap from the cluster and exposes its key-value pairs as a JSON object over HTTP. No business logic — one ConfigMap in, one JSON object out.

## Architecture position

```
Mobile App  ──► GET /configuration  ──► Kubernetes ConfigMap
                                               (name + namespace from env)
```

The mobile app calls `/configuration` on startup to discover runtime settings such as `DISCOVERY_API_URL`. This keeps environment-specific config out of the app binary.

## API

### GET /configuration

Returns all key-value pairs from the configured Kubernetes ConfigMap as a flat JSON object.

**Response `200 OK`**
```json
{
  "DISCOVERY_API_URL": "https://discovery.example.local",
  "RELEASE_LABEL":     "2026-05"
}
```

### GET /health

Returns `{"status":"ok", "time":"..."}` with current UTC timestamp.

### GET /isAlive

Legacy liveness endpoint. Returns `200 OK` (no body).

### GET /metrics

Prometheus metrics.

## Configuration

| Variable | Required | Description |
|----------|----------|-------------|
| `PORT` | Yes | HTTP listen port |
| `CONFIGMAP_NAME` | Yes | Name of the Kubernetes ConfigMap to serve |
| `CONFIGMAP_NAMESPACE` | Yes | Namespace of the ConfigMap |

The service requires a Kubernetes ServiceAccount with `get` permission on ConfigMaps in the target namespace.

## Run locally

```bash
go mod tidy
go run .
```

With Docker Compose (includes Prometheus and Grafana):

```bash
docker compose up -d --build
```

## Observability

Custom Prometheus metrics:

| Metric | Description |
|--------|-------------|
| `configuration_service_http_requests_total{method,route,status}` | HTTP request counter |
| `configuration_service_http_request_duration_seconds{method,route,status}` | HTTP latency histogram |
| `configuration_service_http_in_flight_requests` | Current in-flight requests gauge |

Alert rules: `alerts/configuration-service-alerts.yml`  
Grafana dashboard: `grafana/provisioning/dashboards/configuration-service-dashboard.json`

## Mobile compatibility tests

Contract and live compatibility tests for the mobile client journey are maintained in the discovery-api repository:

```bash
cd ../discovery-api
npm run test:mobile
```

Tests model multiple Android/iOS app cohorts (including older versions) and verify that `/configuration` and `/isAlive` remain backward-compatible.

## Kubernetes RBAC

The service uses in-cluster Kubernetes client authentication. The Helm chart creates the required ServiceAccount, ClusterRole, and ClusterRoleBinding automatically.

## AI Governance

This service uses shared AI governance from [ai-governance](https://github.com/CivicAIPortal/ai-governance). Local config: `.ai/config.yaml`.

```bash
go run ./.ai validate
go run ./.ai ensure-sync
```
