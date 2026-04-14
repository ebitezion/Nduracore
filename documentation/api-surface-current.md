# Nduracore API Surface (Implemented So Far)

This document lists all API endpoints currently implemented in the backend.

## Base
- API prefix for business endpoints: `/v1`
- Health/ops endpoints are at root.

## Authentication Model
- Bearer token is required for protected routes:
  - Header: `Authorization: Bearer <token>`
- Role-based authorization currently used:
  - `admin`
  - `manager`
- Tenant-scoped endpoints require:
  - Header: `X-Tenant-ID`

## Public / System Endpoints
1. `GET /healthcheck`
- Purpose: service health summary.
- Auth: none.

2. `GET /liveness`
- Purpose: liveness probe.
- Auth: none.

3. `GET /readiness`
- Purpose: readiness probe.
- Auth: none.

4. `GET /metrics`
- Purpose: request/latency metrics snapshot.
- Auth: none.

## Auth Endpoints
1. `POST /v1/auth/token`
- Purpose: issue JWT access token.
- Auth: none.
- Typical input: email + password.

## User Endpoints
1. `POST /v1/users`
- Purpose: create a user.
- Auth: bearer token required.
- Roles: `admin` only.

2. `GET /v1/users`
- Purpose: list users with pagination/sorting.
- Auth: bearer token required.
- Roles: `admin`, `manager`.

## Ops / Internal Job Endpoints
1. `POST /v1/jobs/audit`
- Purpose: enqueue audit job.
- Auth: bearer token required.
- Roles: `admin`, `manager`.

## Wallet Endpoints
1. `POST /v1/wallets`
- Purpose: create custodial wallet.
- Auth: bearer token required.
- Roles: `admin`, `manager`.
- Tenant header: required (`X-Tenant-ID`).

2. `GET /v1/wallets`
- Purpose: list wallets.
- Auth: bearer token required.
- Roles: `admin`, `manager`.
- Tenant header: required.

3. `GET /v1/wallets/:id`
- Purpose: get wallet by id.
- Auth: bearer token required.
- Roles: `admin`, `manager`.
- Tenant header: required.

4. `GET /v1/wallets/:id/deposits`
- Purpose: list wallet deposits.
- Auth: bearer token required.
- Roles: `admin`, `manager`.
- Tenant header: required.

5. `POST /v1/wallets/:id/deposits/webhook`
- Purpose: ingest deposit webhook events.
- Auth: no bearer token.
- Security: webhook signature validation via `X-Alchemy-Signature` when configured.
- Tenant header: required.

## Withdrawal Endpoints
1. `POST /v1/withdrawals`
- Purpose: request withdrawal.
- Auth: bearer token required.
- Roles: `admin`, `manager`.
- Tenant header: required.

2. `GET /v1/withdrawals/:id`
- Purpose: get withdrawal by id.
- Auth: bearer token required.
- Roles: `admin`, `manager`.
- Tenant header: required.

3. `POST /v1/withdrawals/:id/approve`
- Purpose: approve a pending withdrawal.
- Auth: bearer token required.
- Roles: `admin`, `manager`.
- Tenant header: required.

4. `POST /v1/withdrawals/:id/reject`
- Purpose: reject a pending withdrawal.
- Auth: bearer token required.
- Roles: `admin`, `manager`.
- Tenant header: required.

## Related Documentation
- Webhook contract and examples:
  - `documentation/runbooks/alchemy-webhook-integration.md`
- Postman collection and environment:
  - `documentation/postman/Nduracore-Phase2.postman_collection.json`
  - `documentation/postman/Nduracore-Phase2.postman_environment.json`
