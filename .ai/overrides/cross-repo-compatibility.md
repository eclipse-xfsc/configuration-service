# Local Override: Discovery/Configuration Cross-Repo Compatibility

## Scope

- Applies to work involving mobile-facing behavior in discovery-api and configuration-service.
- Default mode is local testing only (no mandatory remote integration testing).

## Required Compatibility Contract

Treat these as a shared contract between repositories:

1. `configuration-service`:
- `GET /configuration` must return JSON object payload used by mobile startup.
- `GET /isAlive` must continue to return HTTP 200.

2. `discovery-api`:
- `GET /api/v1/resources/:id` response fields used by mobile remain stable (`url`, `etag`, `version`, `contentType`, `checksum`).
- Empty `version` query handling remains stable (HTTP 400).
- `ETag` and `If-None-Match` behavior remains stable (HTTP 304 on match).

## Local Validation Before Finalizing Changes

When making changes that can affect the shared contract, validate locally:

1. In discovery-api:
- `npm run test:mobile`

Preferred one-command helper from discovery-api root:

- `./compat-check.sh` (or `./compat-check.ps1` on PowerShell)

2. In configuration-service:
- `go test ./...`

Preferred one-command helper from configuration-service root:

- `./compat-check.sh` (or `./compat-check.ps1` on PowerShell)

## Cross-Repo Change Rule

If a change in one repository is not compatible with the shared contract in the other repository:

1. Stop and summarize the incompatibility clearly.
2. Ask the user whether to introduce counterpart changes in the other repository.
3. Only proceed with cross-repo counterpart edits after user confirmation.

## Backward Compatibility Priority

- Prefer non-breaking, additive changes for mobile-facing contracts.
- Preserve behavior for older mobile cohorts whenever possible.
