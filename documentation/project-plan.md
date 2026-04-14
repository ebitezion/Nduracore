# Nduracore Project Plan

## Vision
Nduracore is a modular, custodial wallet and crypto-native neo-core banking framework designed to power the next generation of fintech platforms across emerging and global markets.

It functions as the core banking infrastructure layer, similar to Paystack and Flutterwave for fiat payments, but built crypto-first with seamless fiat interoperability.

## Core Description
Nduracore provides a secure, extendible, and API-driven financial engine that enables businesses to:

- Create and manage custodial crypto wallets at scale.
- Bridge fiat and crypto transactions effortlessly.
- Handle payments, settlements, and treasury operations across both rails.
- Build financial products such as apps, cards, savings, and lending on top of a unified core.

## Key Capabilities
- Custodial Wallet Infrastructure: Multi-asset wallet creation, key management, and transaction orchestration.
- Fiat Bridge Layer: On-ramps and off-ramps connecting bank rails, cards, and mobile money to crypto liquidity.
- Core Banking Engine: Ledgering, balance management, transaction states, and compliance hooks.
- API-First Architecture: Developer-friendly APIs and SDKs for rapid fintech integration.
- Extensible Modules: Cards and payments, FX and stablecoins, lending and savings, and compliance (KYC/AML).

## Positioning Statement
Nduracore is the crypto-native core banking layer for fintechs, enabling any company to build, launch, and scale financial products that unify digital assets and traditional money securely and seamlessly.

## Brand Alignment
Rooted in “Ndu” (life), Nduracore represents:

- Financial life infrastructure.
- Freedom to move value globally.
- The foundation that gives users and businesses wings to fly.

## Delivery Plan by Phase

### Phase 0: Product Blueprint (Weeks 1-2)
- Define ideal customer profile and primary use cases.
- Lock MVP scope across wallet, ledger, transfers, and compliance hooks.
- Produce architecture, threat model, and delivery backlog.

**Exit criteria**
- Signed-off PRD and architecture decisions.
- Prioritized sprint backlog.

### Phase 1: Core Platform Foundation (Weeks 3-6)
- Harden the current backend into Nduracore service modules.
- Set up CI/CD, environment separation, secrets, observability, and audit logging.
- Define module boundaries (`wallet`, `ledger`, `payments`, `compliance`, `integrations`).

**Exit criteria**
- Stable staging environment.
- Baseline operational runbooks and on-call alerts.

### Phase 2: Custodial Wallet MVP (Alchemy + Minimal Fireblocks MVP) (Weeks 7-12)
- Implement wallet orchestration with provider abstraction.
- Use Alchemy for blockchain connectivity and wallet operations.
- Build minimal Fireblocks-style controls:
  - Policy engine (whitelists, limits, velocity checks, approval tiers).
  - Pre-broadcast simulation and risk checks.
  - Tenant-safe custody abstraction with strict audit trails.
- Ship deposit monitoring and controlled withdrawals.

**Exit criteria**
- Wallet creation, deposit detection, and withdrawal flow with policy approvals.
- Full transaction and action audit trail.

Detailed execution doc: `documentation/phase-2-custodial-wallet-mvp.md`

### Phase 3: Core Banking Engine (Weeks 13-18)
- Implement double-entry ledger, holds, pending and posted states, and reversals.
- Build reconciliation jobs and idempotent transaction handling.
- Add treasury and internal transfer workflows.

**Exit criteria**
- Every movement maps to ledger entries.
- Automated reconciliation passes.

### Phase 4: Fiat Bridge Layer (Weeks 19-26)
- Add adapters for bank transfer, card, and mobile money providers.
- Implement on-ramp and off-ramp workflows.
- Add FX/stablecoin quoting, fee models, and settlement states.

**Exit criteria**
- End-to-end fiat-to-crypto and crypto-to-fiat flow in sandbox and pilot.

### Phase 5: Compliance and Risk Controls (Weeks 19-26, parallel with Phase 4)
- Integrate KYC/KYB, AML, sanctions, and PEP checks via hooks.
- Build transaction monitoring rules and risk scoring events.
- Add case evidence export and compliance-ready logs.

**Exit criteria**
- Configurable compliance workflows.
- Evidence-grade auditability for key actions.

### Phase 6: API-First Productization (Weeks 27-30)
- Publish versioned APIs and webhook contracts.
- Deliver SDK starter kits and integration quickstarts.
- Add tenant-level API keys and access controls.

**Exit criteria**
- External teams can integrate without internal hand-holding.

### Phase 7: Extensible Financial Modules (Ongoing)
- Launch plug-and-play modules for cards, savings/lending, FX/stablecoins, and treasury.
- Define extension contracts and partner integration boundaries.

**Exit criteria**
- At least two production-ready modules beyond core wallet and ledger.

### Phase 8: Scale, Reliability, and Expansion (Ongoing)
- Achieve high availability goals, DR drills, and regional deployment readiness.
- Add partner onboarding standards, SLA tiers, and production certifications.

**Exit criteria**
- Production SLA targets met.
- Repeatable onboarding and deployment model across markets.

## Recommended Execution Sequence
1. Complete Phases 0, 1, 2, and 3 as the MVP backbone.
2. Run Phases 4 and 5 in parallel once ledger integrity is stable.
3. Move to Phase 6 after first pilot integrations.
4. Continue Phases 7 and 8 as scale and product expansion tracks.
