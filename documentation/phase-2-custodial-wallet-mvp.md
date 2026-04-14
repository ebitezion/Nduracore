# Phase 2: Custodial Wallet MVP (Alchemy + Minimal Fireblocks Controls)

## Objective
Deliver a production-ready custodial wallet MVP that supports secure wallet creation, deposit tracking, and controlled withdrawals using Alchemy as blockchain infrastructure, with policy-driven approval controls inspired by minimal Fireblocks workflows.

## Duration
Weeks 7-12 (6 weeks).

## Scope
- Multi-tenant custodial wallet lifecycle.
- Provider abstraction with Alchemy as first provider.
- Deposit monitoring and confirmation tracking.
- Withdrawal flow with policy checks and approvals.
- Full audit trail for wallet and transfer actions.

## Out of Scope (Phase 2)
- Fiat on-ramp/off-ramp settlement execution.
- Advanced treasury automation.
- Full multi-provider failover.
- Cross-chain smart-order routing.

## Architecture Targets
1. `internal/wallet`
- Wallet service orchestration and domain logic.
- Asset/network validation and tenant isolation.
- Address metadata and lifecycle state transitions.

2. `internal/integrations`
- `alchemy` adapter implementing provider interface.
- Chain RPC/event subscription wrapper.
- Transaction submission and status retrieval.

3. `internal/compliance`
- Hook points for transaction screening calls before broadcast.
- Blocking and approval-required decisions.

4. `internal/ledger` (light integration only)
- Event emission contract so deposits/withdrawals can post to ledger in Phase 3.

## Phase 2 Workstreams

### Workstream A: Wallet Orchestration Core
- Implement wallet creation use case per tenant and network.
- Add wallet repository schema and persistence layer.
- Add wallet status model (`active`, `frozen`, `closed`).

Deliverables:
- Wallet service implementation.
- Migration for wallet tables.
- API endpoints for wallet creation and retrieval.

### Workstream B: Alchemy Provider Integration
- Implement Alchemy-backed provider interface in `internal/integrations/alchemy`.
- Support EVM network configuration and wallet/address derivation strategy.
- Implement transaction submit + status polling methods.

Deliverables:
- Alchemy adapter and config contract.
- Health checks for provider connectivity.
- Integration tests with mocked provider responses.

### Workstream C: Deposit Monitoring
- Build deposit watcher by tracked addresses.
- Persist deposit events with confirmation counts.
- Trigger internal domain event when deposit reaches required confirmations.

Deliverables:
- Background job/worker for deposits.
- API endpoint to query deposit status.
- Webhook ingestion endpoint for provider-pushed deposit events.
- Audit logs for every observed inbound transfer.

### Workstream D: Controlled Withdrawals (Minimal Fireblocks Style)
- Build policy engine with rules:
  - Whitelist-only destination enforcement.
  - Amount limits per tenant and per period.
  - Velocity thresholds.
  - Approval tiers (auto, one approver, two approvers).
- Add withdrawal states:
  - `requested`, `policy_pending`, `approved`, `broadcasted`, `confirmed`, `failed`, `rejected`.
- Add pre-broadcast validation/simulation hook.

Deliverables:
- Policy evaluation module.
- Withdrawal approval workflow APIs.
- Broadcast executor with retry rules and idempotency protection.

### Workstream E: Security, Audit, and Operations
- Extend structured audit events for all wallet and withdrawal actions.
- Add access control matrix for wallet operators vs approvers.
- Add operational dashboards for deposits, pending approvals, and failed broadcasts.

Deliverables:
- Audit event catalog.
- Runbook updates for stuck withdrawals and chain incident handling.
- Alert thresholds for deposit delays and broadcast failures.

## Suggested API Surface (Phase 2)
- `POST /v1/wallets`
- `GET /v1/wallets/:id`
- `GET /v1/wallets`
- `GET /v1/wallets/:id/deposits`
- `POST /v1/wallets/:id/deposits/webhook`
- `POST /v1/withdrawals`
- `GET /v1/withdrawals/:id`
- `POST /v1/withdrawals/:id/approve`
- `POST /v1/withdrawals/:id/reject`

## Data Model Additions (Initial)
- `wallets`
- `wallet_addresses`
- `wallet_deposits`
- `withdrawals`
- `withdrawal_approvals`
- `policy_rules`
- `audit_events` (or extend existing log/event store)

## Security and Policy Baseline
- Enforce idempotency key for withdrawal creation.
- Require role-based authorization for request/approve actions.
- Require tenant scoping on all wallet and withdrawal reads/writes.
- Deny broadcast if policy evaluation result is not `approved`.
- Require immutable audit log record before and after each state transition.

## Test Strategy
- Unit tests:
  - policy engine rule evaluation.
  - wallet state transitions.
  - approval flow state machine.
- Integration tests:
  - Alchemy adapter behavior under normal/error conditions.
  - withdrawal lifecycle end-to-end with mocked chain responses.
- Non-functional tests:
  - retry/idempotency verification.
  - race/concurrency tests for approval actions.

## Exit Criteria
- Wallet creation and retrieval works for supported networks.
- Deposit detection and confirmation flow is reliable.
- Withdrawals cannot be broadcast without policy approval.
- Every wallet and withdrawal action produces auditable events.
- Staging demo completes end-to-end:
  - create wallet
  - detect deposit
  - request withdrawal
  - approve withdrawal
  - broadcast and confirm

## Dependencies
- Staging environment from Phase 1 is operational.
- Alchemy project credentials and network scopes are provisioned.
- Alchemy RPC and webhook secrets are provisioned in environment config.
- Security roles for operator/approver/admin are available.

## Risks and Mitigations
1. Provider API instability
- Mitigation: retries, circuit breaker, and fallback job replay queue.

2. Confirmation delays on chain
- Mitigation: configurable confirmation thresholds and pending-state alerts.

3. Policy misconfiguration
- Mitigation: safe defaults (`deny`), policy simulation endpoint, and change audit logs.
