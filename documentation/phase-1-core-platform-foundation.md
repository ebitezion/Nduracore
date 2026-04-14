# Phase 1: Core Platform Foundation

## Objective
Harden the Vein-based backend into Nduracore platform services with strong operational foundations for staging.

## Scope
- Service boundary definitions for `wallet`, `ledger`, `payments`, `compliance`, and `integrations`.
- Environment and secret management standards.
- Observability and audit log baseline.
- CI hardening baseline.
- Staging runbook and error budget policy.

## Delivered in this phase
1. Module boundaries were established in code via domain service contracts:
   - `internal/wallet/service.go`
   - `internal/ledger/service.go`
   - `internal/payments/service.go`
   - `internal/compliance/service.go`
   - `internal/integrations/service.go`
2. Environment defaults were rebranded to Nduracore:
   - `APP_NAME=nduracore`
   - `TOKEN_AUDIENCE=nduracore-clients`
   - Redis queue key uses `nduracore:jobs`
3. Structured audit events were added to critical API flows:
   - Token issuance success/failure events
   - Audit job enqueue events
4. CI test database naming was aligned to Nduracore.
5. Staging deployment automation added via GitHub Actions workflow:
   - `.github/workflows/staging.yml`
6. Operations documentation added:
   - Environment and secret strategy
   - Staging operations runbook
   - SLO and error budget policy
   - Staging deployment workflow guide

## Exit Criteria for this phase
- Staging environment configuration checklist is complete.
- Operational runbook exists and can be followed by another engineer.
- Error budget and incident trigger thresholds are documented.
- CI validates formatting, tests, static analysis, and vulnerability checks.
