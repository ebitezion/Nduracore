# Nduracore API Surface (Implemented So Far)

This document lists all API endpoints currently implemented in the backend.

## Base
- API prefix for business endpoints: `/v1`
- Health/ops endpoints are at root.

## Authentication and Access Model
- Bearer token is required for protected routes:
  - Header: `Authorization: Bearer <token>`
- Roles currently implemented:
  - `user`
  - `admin`
  - `manager`
  - `super_admin`
- Tenant-scoped routes require:
  - Header: `X-Tenant-ID`
- Tenant access is enforced for non-`super_admin` users on scoped routes.
- `super_admin` bypasses tenant-access checks and can operate across tenants.

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
- Purpose: issue JWT access token for active users.
- Auth: none.
- Typical input: email + password.

2. `POST /v1/auth/register`
- Purpose: self-service registration.
- Auth: none.
- Behavior: creates a `pending` user for super-admin approval.

## User Endpoints
1. `POST /v1/users`
- Purpose: create a user directly.
- Auth: bearer token required.
- Roles: `admin` (and `super_admin` by override).

2. `GET /v1/users`
- Purpose: list users with pagination/sorting.
- Auth: bearer token required.
- Roles: `admin`, `manager` (and `super_admin` by override).

## Admin User Approval Endpoints
1. `GET /v1/admin/users/pending`
- Purpose: list pending registrations.
- Auth: bearer token required.
- Roles: `super_admin`.

2. `POST /v1/admin/users/:id/approve`
- Purpose: approve a pending user, set role, and grant tenant access.
- Auth: bearer token required.
- Roles: `super_admin`.
- Typical input: `tenant_id`, optional `role`.

3. `POST /v1/admin/users/:id/reject`
- Purpose: reject a pending user.
- Auth: bearer token required.
- Roles: `super_admin`.

## Treasury Endpoints
1. `POST /v1/treasuries`
- Purpose: create treasury.
- Auth: bearer token required.
- Roles: `admin`, `manager` (and `super_admin` by override).
- Tenant header: required.

2. `GET /v1/treasuries`
- Purpose: list treasuries.
- Auth: bearer token required.
- Roles: `admin`, `manager` (and `super_admin` by override).
- Tenant header: required.

3. `GET /v1/treasuries/:id`
- Purpose: get treasury by id.
- Auth: bearer token required.
- Roles: `admin`, `manager` (and `super_admin` by override).
- Tenant header: required.

4. `POST /v1/treasuries/:id/vaults`
- Purpose: create vault under treasury.
- Auth: bearer token required.
- Roles: `admin`, `manager` (and `super_admin` by override).
- Tenant header: required.

5. `GET /v1/treasuries/:id/vaults`
- Purpose: list vaults under treasury.
- Auth: bearer token required.
- Roles: `admin`, `manager` (and `super_admin` by override).
- Tenant header: required.

6. `GET /v1/vaults/:id`
- Purpose: get vault by id.
- Auth: bearer token required.
- Roles: `admin`, `manager` (and `super_admin` by override).
- Tenant header: required.

7. `POST /v1/vaults/:id/assets`
- Purpose: create vault asset (asset is unique per vault by asset_code + network).
- Auth: bearer token required.
- Roles: `admin`, `manager` (and `super_admin` by override).
- Tenant header: required.

8. `GET /v1/vaults/:id/assets`
- Purpose: list assets in vault.
- Auth: bearer token required.
- Roles: `admin`, `manager` (and `super_admin` by override).
- Tenant header: required.

9. `POST /v1/vaults/:id/aliases`
- Purpose: map an external provider vault identifier to an internal Nduracore vault UUID.
- Auth: bearer token required.
- Roles: `admin`, `manager` (and `super_admin` by override).
- Tenant header: required.

10. `GET /v1/vaults/:id/aliases`
- Purpose: list aliases registered for a vault.
- Auth: bearer token required.
- Roles: `admin`, `manager` (and `super_admin` by override).
- Tenant header: required.

11. `GET /v1/vault-aliases/resolve`
- Purpose: resolve `provider + external_vault_id` to internal vault mapping.
- Auth: bearer token required.
- Roles: `admin`, `manager` (and `super_admin` by override).
- Tenant header: required.
- Query params: `provider`, `external_vault_id`.

## Ops / Internal Job Endpoints
1. `POST /v1/jobs/audit`
- Purpose: enqueue audit job.
- Auth: bearer token required.
- Roles: `admin`, `manager` (and `super_admin` by override).

## Wallet Endpoints
1. `POST /v1/wallets`
- Purpose: create custodial wallet.
- Auth: bearer token required.
- Roles: `admin`, `manager` (and `super_admin` by override).
- Tenant header: required (`X-Tenant-ID`).
- Notes:
  - `vault_id` is required.
  - Wallet must be linked to an existing tenant vault.

2. `GET /v1/wallets`
- Purpose: list wallets.
- Auth: bearer token required.
- Roles: `admin`, `manager` (and `super_admin` by override).
- Tenant header: required.
- Query params (optional):
  - `vault_id` (UUID)
  - `asset`
  - `network`

3. `GET /v1/wallets/:id`
- Purpose: get wallet by id.
- Auth: bearer token required.
- Roles: `admin`, `manager` (and `super_admin` by override).
- Tenant header: required.

4. `GET /v1/wallets/:id/deposits`
- Purpose: list wallet deposits.
- Auth: bearer token required.
- Roles: `admin`, `manager` (and `super_admin` by override).
- Tenant header: required.

5. `GET /v1/wallets/:id/balance`
- Purpose: return on-chain balance for wallet/network (optional `asset` query override).
- Auth: bearer token required.
- Roles: `admin`, `manager` (and `super_admin` by override).
- Tenant header: required.

6. `GET /v1/wallets/:id/transactions`
- Purpose: aggregate wallet transaction history from recorded deposits + withdrawals.
- Auth: bearer token required.
- Roles: `admin`, `manager` (and `super_admin` by override).
- Tenant header: required.

7. `GET /v1/wallets/:id/gas-estimate`
- Purpose: estimate transfer gas/fee for destination + amount.
- Auth: bearer token required.
- Roles: `admin`, `manager` (and `super_admin` by override).
- Tenant header: required.
- Query params: `destination`, `amount_minor`, optional `asset`.

8. `POST /v1/wallets/:id/withdrawals/simulate`
- Purpose: dry-run withdrawal risk simulation and fee preview.
- Auth: bearer token required.
- Roles: `admin`, `manager` (and `super_admin` by override).
- Tenant header: required.

9. `GET /v1/wallets/:id/token-allowances`
- Purpose: check token allowance for wallet owner to spender.
- Auth: bearer token required.
- Roles: `admin`, `manager` (and `super_admin` by override).
- Tenant header: required.
- Query params: `spender`, optional `asset`.

10. `GET /v1/wallets/:id/risk-score`
- Purpose: return internal heuristic risk score for wallet activity.
- Auth: bearer token required.
- Roles: `admin`, `manager` (and `super_admin` by override).
- Tenant header: required.

11. `GET /v1/wallets/:id/nonces`
- Purpose: return nonce state (`latest`, `pending`) for supported networks.
- Auth: bearer token required.
- Roles: `admin`, `manager` (and `super_admin` by override).
- Tenant header: required.

12. `POST /v1/wallets/:id/deposits/webhook`
- Purpose: ingest deposit webhook events.
- Auth: no bearer token.
- Security: webhook signature validation via `X-Alchemy-Signature` when configured.
- Tenant header: required.

13. `POST /v1/webhooks/:provider`
- Purpose: provider-agnostic webhook ingest endpoint that normalizes events (`N2`).
- Auth: no bearer token.
- Tenant header: required.
- Query params: `wallet_id` required for `provider=alchemy`.
- Notes:
  - stores raw provider event + normalized event records.
  - failed normalization goes to event failure + DLQ records.

## Withdrawal Endpoints
1. `POST /v1/withdrawals`
- Purpose: request withdrawal.
- Auth: bearer token required.
- Roles: `admin`, `manager` (and `super_admin` by override).
- Tenant header: required.
- Source mode:
  - Wallet mode: `wallet_id`, `destination`, `amount_minor`.
  - Vault mode: `vault_id`, `asset`, `network`, `destination`, `amount_minor`.
  - Modes are mutually exclusive per request.

2. `GET /v1/withdrawals/:id`
- Purpose: get withdrawal by id.
- Auth: bearer token required.
- Roles: `admin`, `manager` (and `super_admin` by override).
- Tenant header: required.

3. `GET /v1/withdrawals`
- Purpose: list withdrawals for a tenant.
- Auth: bearer token required.
- Roles: `admin`, `manager` (and `super_admin` by override).
- Tenant header: required.
- Query params (optional):
  - `status` (`requested`, `policy_pending`, `approved`, `broadcasted`, `confirmed`, `rejected`, `failed`, `cancelled`)

4. `POST /v1/withdrawals/:id/approve`
- Purpose: approve a pending withdrawal.
- Auth: bearer token required.
- Roles: `admin`, `manager` (and `super_admin` by override).
- Tenant header: required.

5. `POST /v1/withdrawals/:id/reject`
- Purpose: reject a pending withdrawal.
- Auth: bearer token required.
- Roles: `admin`, `manager` (and `super_admin` by override).
- Tenant header: required.

6. `GET /v1/withdrawals/:id/trace`
- Purpose: return withdrawal lifecycle + chain receipt enrichment where available.
- Auth: bearer token required.
- Roles: `admin`, `manager` (and `super_admin` by override).
- Tenant header: required.

## Chain Analysis Endpoints
1. `GET /v1/networks/:network/status`
- Purpose: health/status probe for network RPC.

## Fee Policy Endpoints
1. `PUT /v1/fees/policies/:network/:asset`
- Purpose: create/update tenant fee policy used for fee planning during withdrawal orchestration (`N6`).
- Auth: bearer token required.
- Roles: `admin`, `manager` (and `super_admin` by override).
- Tenant header: required.

2. `GET /v1/fees/policies/:network/:asset`
- Purpose: read tenant fee policy.
- Auth: bearer token required.
- Roles: `admin`, `manager` (and `super_admin` by override).
- Tenant header: required.

## Ops Endpoints
1. `POST /v1/ops/reconcile/withdrawals/run`
- Purpose: run a withdrawal reconciliation job against provider state and store mismatch records (`N7`).
- Auth: bearer token required.
- Roles: `admin`, `manager` (and `super_admin` by override).
- Tenant header: required.

2. `POST /v1/ops/events/dlq/:id/replay`
- Purpose: replay a queued dead-letter event (`N7`).
- Auth: bearer token required.
- Roles: `admin`, `manager` (and `super_admin` by override).
- Tenant header: required.

## Canonical State Updates
- Canonical withdrawal states now include: `requested`, `policy_pending`, `approved`, `broadcasted`, `confirmed`, `failed`, `rejected`, `cancelled`.
- State transition history is persisted in `withdrawal_state_history` (`N4`).
- Auth: bearer token required.
- Roles: `admin`, `manager` (and `super_admin` by override).

2. `POST /v1/addresses/validate`
- Purpose: chain-family specific address validation.
- Auth: bearer token required.
- Roles: `admin`, `manager` (and `super_admin` by override).

3. `GET /v1/assets/:network/:asset/metadata`
- Purpose: return active asset registry metadata for network+asset.
- Auth: bearer token required.
- Roles: `admin`, `manager` (and `super_admin` by override).

## Related Documentation
- Webhook contract and examples:
  - `documentation/runbooks/alchemy-webhook-integration.md`
- Postman collection and environment:
  - `documentation/postman/Nduracore-Current-API.postman_collection.json`
  - `documentation/postman/Nduracore-Current-API.postman_environment.json`
  - `documentation/postman/Nduracore-Phase2.postman_collection.json`
  - `documentation/postman/Nduracore-Phase2.postman_environment.json`
