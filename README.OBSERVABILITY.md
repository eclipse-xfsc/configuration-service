# Local Development Observability Stack

This directory includes a complete local observability stack for configuration-service development and testing.

## Quick Start

1. **Start the stack:**
   ```bash
   docker-compose up -d
   ```

2. **Access services:**
   - Configuration Service: http://localhost:8080
   - Prometheus: http://localhost:9090
   - Grafana: http://localhost:3000 (default admin:admin)

3. **Stop the stack:**
   ```bash
   docker-compose down
   ```

## Service Endpoints

### Configuration Service
- `GET http://localhost:8080/health` — JSON health status
- `GET http://localhost:8080/isAlive` — Simple 200 OK (backward compatible)
- `GET http://localhost:8080/configuration` — ConfigMap content
- `GET http://localhost:8080/metrics` — Prometheus metrics

### Prometheus
- UI: http://localhost:9090
- Metrics endpoint: http://localhost:9090/metrics
- Query examples:
  - `configuration_service_http_requests_total` — total request count by route/method/status
  - `configuration_service_http_request_duration_seconds` — request latency histogram
  - `configuration_service_http_in_flight_requests` — current in-flight requests
  - `probe_success{job="service-health-checks"}` — health probe status

### Grafana
- UI: http://localhost:3000
- Pre-configured datasource: Prometheus (http://prometheus:9090)
- Pre-configured dashboard: Configuration Service dashboard
- Credentials: admin / admin

### Blackbox Exporter
- Endpoint: http://localhost:9115
- Probes HTTP/2xx and TCP connectivity
- Used by Prometheus for external health checks

## Manual Testing

### Test the health endpoint:
```bash
curl -s http://localhost:8080/health | jq .
```

### Test metrics endpoint:
```bash
curl -s http://localhost:8080/metrics | grep configuration_service
```

### Query Prometheus:
```bash
curl -s 'http://localhost:9090/api/v1/query?query=configuration_service_http_requests_total'
```

### Trigger alerts (optional):
Prometheus alert rules evaluate every 15 seconds. Generated alerts appear in Grafana if Prometheus alerts datasource is configured.

## Troubleshooting

**Prometheus not scraping configuration-service?**
- Verify service is running: `docker ps | grep configuration-service`
- Check Prometheus UI http://localhost:9090/targets — service should show as "UP"
- Review logs: `docker-compose logs prometheus`

**Grafana dashboard not showing metrics?**
- Wait 1–2 minutes for Prometheus to scrape and store metrics
- Verify datasource is configured: Grafana Settings → Data Sources → Prometheus (should show green checkmark)
- Check dashboard UID: "kivp-configuration-service"

**Port already in use?**
Edit `docker-compose.yml` and change ports, e.g., `8081:8080` for configuration-service.

## Files

- `docker-compose.yml` — orchestration definition
- `prometheus.yml` — Prometheus scrape config and rules
- `blackbox-exporter.yml` — Blackbox HTTP/TCP probe modules
- `alerts/` — Prometheus alert rules
- `grafana/provisioning/` — Grafana datasources and dashboards
