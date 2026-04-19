# Simple Update Notes (April 19, 2026)

This release adds the core groundwork to make Nduracore much closer to "green" for provider migration and enterprise operations.

## What changed in plain language

1. Webhooks are now more generic.
- We added a provider-agnostic webhook endpoint (`/v1/webhooks/:provider`).
- Nduracore now stores raw provider events and normalized events.
- Failed events are tracked and can be replayed from DLQ.

2. Withdrawal states are more structured.
- We introduced canonical state handling and transition history (`withdrawal_state_history`).
- This improves traceability and makes provider differences easier to map.

3. External vault IDs can be mapped.
- You can now attach aliases from external systems to internal vault UUIDs.
- New alias endpoints help resolve provider vault IDs into Nduracore vault IDs.

4. Fee planning is now first-class.
- We added tenant fee policies by asset/network.
- Withdrawal request flow now calculates fee and stores fee charge records.
- Simulation now returns fee preview when a fee policy exists.

5. Reconciliation and DLQ operations were added.
- New reconciliation endpoint checks withdrawal state drift vs provider/chain observations.
- New DLQ replay endpoint lets ops replay failed webhook events safely.

## New tables added

- `vault_aliases`
- `fee_policies`
- `fee_charges`
- `withdrawal_state_history`
- `provider_events`
- `normalized_events`
- `event_failures`
- `event_dlq`
- `reconciliation_runs`
- `reconciliation_items`

## What this means

- Better migration support from external providers.
- Better auditability and reliability for transaction lifecycle.
- Better operational tools for handling event failures.
- Better fee control and predictable net transfer behavior.
