# Staging Operations Runbook

## Purpose
Operate and verify Nduracore safely in staging before production release.

## Prerequisites
- PostgreSQL available and reachable from staging runtime.
- Optional Redis available for shared cache/queue/rate limiting.
- Environment variables configured from `.env.example` baseline.

## Startup Checklist
1. Confirm `MY_ENV=staging`.
2. Confirm `DB_DSN` points to staging DB.
3. Confirm `TOKEN_SECRET` is non-default and strong.
4. Confirm `CORS_TRUSTED_ORIGINS` only includes approved staging clients.
5. Confirm tracing setting:
   - `OTEL_ENABLED=true`
   - `OTEL_EXPORTER_OTLP_ENDPOINT` set to collector endpoint.

## Deployment Validation
1. Run CI pipeline on commit.
2. Run staging workflow (`.github/workflows/staging.yml`) or push to `main`.
3. Apply migrations.
4. Deploy API.
5. Verify endpoints:
   - `GET /healthcheck`
   - `GET /liveness`
   - `GET /readiness`
   - `GET /metrics`
6. Perform smoke test:
   - Issue auth token via `POST /v1/auth/token`.
   - Access `GET /v1/users` with bearer token.

## Audit and Observability Checks
1. Confirm structured request logs include `request_id`.
2. Confirm audit events are emitted for:
   - `auth.token.issue`
   - `jobs.audit.enqueue`
3. Confirm tracing spans are exported when OTEL is enabled.

## Rollback Procedure
1. Pause new traffic (or scale to zero).
2. Roll back deployment to previous known good image/commit.
3. Re-run health checks.
4. If DB migration caused issue, apply pre-tested rollback migration.

## Incident Trigger Thresholds
- Readiness endpoint fails for more than 5 minutes.
- 5xx ratio exceeds 2% over a 15 minute window.
- Token issuance failure spike above normal baseline for 10 minutes.
