

### 🚀 **Nduracore — Defining Prompt**

**Nduracore** is a modular, custodial wallet and crypto-native neo-core banking framework designed to power the next generation of fintech platforms across emerging and global markets.

It functions as the **core banking infrastructure layer**—similar to what Paystack and Flutterwave provide for fiat payments—but built **crypto-first with seamless fiat interoperability**.

---

### 🧠 **Core Description**

Nduracore provides a **secure, extendible, and API-driven financial engine** that enables businesses to:

* Create and manage **custodial crypto wallets** at scale
* Bridge **fiat ↔ crypto** transactions effortlessly
* Handle **payments, settlements, and treasury operations** across both rails
* Build **financial products (apps, cards, savings, lending)** on top of a unified core

---

### ⚙️ **Key Capabilities**

* **Custodial Wallet Infrastructure**
  Multi-asset wallet creation, key management, and transaction orchestration

* **Fiat Bridge Layer**
  On-ramps and off-ramps connecting bank rails, cards, and mobile money to crypto liquidity

* **Core Banking Engine**
  Ledgering, balance management, transaction states, compliance hooks

* **API-First Architecture**
  Developer-friendly APIs and SDKs for rapid fintech integration

* **Extensible Modules**
  Plug-and-play support for:

  * Cards & payments
  * FX & stablecoins
  * Lending & savings
  * Compliance (KYC/AML)

---

### 🌍 **Positioning Statement**

> Nduracore is the **crypto-native core banking layer for fintechs**, enabling any company to build, launch, and scale financial products that unify digital assets and traditional money—securely and seamlessly.

---

### 🪽 **Brand Alignment**

Rooted in “Ndu” (life), **Nduracore** represents:

* Financial life infrastructure
* Freedom to move value globally
* The foundation that gives users and businesses **“wings to fly”**



---

## API Walkthrough (Implemented So Far)

Base URL:

```bash
http://localhost:4000
```

Common headers:

- Protected routes: `Authorization: Bearer <token>`
- Tenant-scoped routes: `X-Tenant-ID: <tenant-id>`

### 1) System Endpoints

#### `GET /healthcheck`
```bash
curl http://localhost:4000/healthcheck
```

#### `GET /liveness`
```bash
curl http://localhost:4000/liveness
```

#### `GET /readiness`
```bash
curl http://localhost:4000/readiness
```

#### `GET /metrics`
```bash
curl http://localhost:4000/metrics
```

### 2) Auth

#### `POST /v1/auth/token`
```bash
curl -X POST http://localhost:4000/v1/auth/token \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@vein.dev","password":"VeinPass#2026!"}'
```

### 3) Users

#### `POST /v1/users` (admin only)
```bash
curl -X POST http://localhost:4000/v1/users \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"first_name":"Jane","last_name":"Doe","email":"jane@example.com","phone":"+10000000010","password":"StrongPass#2026","role":"manager","status":"active","email_verified":true}'
```

#### `GET /v1/users`
```bash
curl "http://localhost:4000/v1/users?page=1&page_size=20&sort=created_at" \
  -H "Authorization: Bearer <token>"
```

### 4) Treasury / Vault / Assets

#### `POST /v1/treasuries`
```bash
curl -X POST http://localhost:4000/v1/treasuries \
  -H "Authorization: Bearer <token>" \
  -H "X-Tenant-ID: tenant-1" \
  -H "Content-Type: application/json" \
  -d '{"name":"Main Treasury","description":"Primary treasury","status":"active"}'
```

#### `GET /v1/treasuries`
```bash
curl "http://localhost:4000/v1/treasuries?page=1&page_size=20&sort=-created_at" \
  -H "Authorization: Bearer <token>" \
  -H "X-Tenant-ID: tenant-1"
```

#### `GET /v1/treasuries/:id`
```bash
curl http://localhost:4000/v1/treasuries/<treasury_id> \
  -H "Authorization: Bearer <token>" \
  -H "X-Tenant-ID: tenant-1"
```

#### `POST /v1/treasuries/:id/vaults`
```bash
curl -X POST http://localhost:4000/v1/treasuries/<treasury_id>/vaults \
  -H "Authorization: Bearer <token>" \
  -H "X-Tenant-ID: tenant-1" \
  -H "Content-Type: application/json" \
  -d '{"name":"Operations Vault","description":"Ops balances","status":"active"}'
```

#### `GET /v1/treasuries/:id/vaults`
```bash
curl "http://localhost:4000/v1/treasuries/<treasury_id>/vaults?page=1&page_size=20&sort=-created_at" \
  -H "Authorization: Bearer <token>" \
  -H "X-Tenant-ID: tenant-1"
```

#### `GET /v1/vaults/:id`
```bash
curl http://localhost:4000/v1/vaults/<vault_id> \
  -H "Authorization: Bearer <token>" \
  -H "X-Tenant-ID: tenant-1"
```

#### `POST /v1/vaults/:id/assets`
```bash
curl -X POST http://localhost:4000/v1/vaults/<vault_id>/assets \
  -H "Authorization: Bearer <token>" \
  -H "X-Tenant-ID: tenant-1" \
  -H "Content-Type: application/json" \
  -d '{"asset_code":"USDC","network":"eth-sepolia","metadata":{"category":"stablecoin"}}'
```

#### `GET /v1/vaults/:id/assets`
```bash
curl "http://localhost:4000/v1/vaults/<vault_id>/assets?page=1&page_size=20&sort=-created_at" \
  -H "Authorization: Bearer <token>" \
  -H "X-Tenant-ID: tenant-1"
```

### 5) Jobs

#### `POST /v1/jobs/audit`
```bash
curl -X POST http://localhost:4000/v1/jobs/audit \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"action":"manual_audit"}'
```

### 6) Wallets

#### `POST /v1/wallets`
```bash
curl -X POST http://localhost:4000/v1/wallets \
  -H "Authorization: Bearer <token>" \
  -H "X-Tenant-ID: tenant-1" \
  -H "Content-Type: application/json" \
  -d '{"vault_id":"<vault_id>","asset":"USDC","network":"eth-sepolia"}'
```

#### `GET /v1/wallets`
```bash
curl "http://localhost:4000/v1/wallets?page=1&page_size=20&sort=-created_at" \
  -H "Authorization: Bearer <token>" \
  -H "X-Tenant-ID: tenant-1"
```

Optional filters: `vault_id` (UUID), `asset`, `network`.
```bash
curl "http://localhost:4000/v1/wallets?vault_id=<vault_id>&asset=USDC&network=eth-sepolia&page=1&page_size=20&sort=-created_at" \
  -H "Authorization: Bearer <token>" \
  -H "X-Tenant-ID: tenant-1"
```

#### `GET /v1/wallets/:id`
```bash
curl http://localhost:4000/v1/wallets/<wallet_id> \
  -H "Authorization: Bearer <token>" \
  -H "X-Tenant-ID: tenant-1"
```

#### `GET /v1/wallets/:id/deposits`
```bash
curl "http://localhost:4000/v1/wallets/<wallet_id>/deposits?page=1&page_size=20&sort=-created_at" \
  -H "Authorization: Bearer <token>" \
  -H "X-Tenant-ID: tenant-1"
```

#### `POST /v1/wallets/:id/deposits/webhook`
```bash
curl -X POST http://localhost:4000/v1/wallets/<wallet_id>/deposits/webhook \
  -H "X-Tenant-ID: tenant-1" \
  -H "X-Alchemy-Signature: <signature>" \
  -H "Content-Type: application/json" \
  -d '{"tx_hash":"0xabc123","amount_minor":250000,"confirmations":12,"status":"confirmed"}'
```

### 7) Withdrawals

#### `POST /v1/withdrawals`
```bash
curl -X POST http://localhost:4000/v1/withdrawals \
  -H "Authorization: Bearer <token>" \
  -H "X-Tenant-ID: tenant-1" \
  -H "Idempotency-Key: withdrawal-req-001" \
  -H "Content-Type: application/json" \
  -d '{"wallet_id":"<wallet_id>","destination":"0x1111111111111111111111111111111111111111","amount_minor":100000}'
```

Vault mode (mutually exclusive with wallet mode):
```bash
curl -X POST http://localhost:4000/v1/withdrawals \
  -H "Authorization: Bearer <token>" \
  -H "X-Tenant-ID: tenant-1" \
  -H "Idempotency-Key: withdrawal-req-002" \
  -H "Content-Type: application/json" \
  -d '{"vault_id":"<vault_id>","asset":"USDC","network":"eth-sepolia","destination":"0x1111111111111111111111111111111111111111","amount_minor":100000}'
```

#### `GET /v1/withdrawals/:id`
```bash
curl http://localhost:4000/v1/withdrawals/<withdrawal_id> \
  -H "Authorization: Bearer <token>" \
  -H "X-Tenant-ID: tenant-1"
```

#### `GET /v1/withdrawals`
```bash
curl "http://localhost:4000/v1/withdrawals?page=1&page_size=20&sort=-created_at" \
  -H "Authorization: Bearer <token>" \
  -H "X-Tenant-ID: tenant-1"
```

### 8) Chain Analysis APIs

#### `GET /v1/wallets/:id/balance`
```bash
curl "http://localhost:4000/v1/wallets/<wallet_id>/balance?asset=USDC" \
  -H "Authorization: Bearer <token>" \
  -H "X-Tenant-ID: tenant-1"
```

#### `GET /v1/wallets/:id/transactions`
```bash
curl "http://localhost:4000/v1/wallets/<wallet_id>/transactions?page=1&page_size=20" \
  -H "Authorization: Bearer <token>" \
  -H "X-Tenant-ID: tenant-1"
```

#### `GET /v1/wallets/:id/gas-estimate`
```bash
curl "http://localhost:4000/v1/wallets/<wallet_id>/gas-estimate?destination=0x1111111111111111111111111111111111111111&amount_minor=100000&asset=USDC" \
  -H "Authorization: Bearer <token>" \
  -H "X-Tenant-ID: tenant-1"
```

#### `POST /v1/wallets/:id/withdrawals/simulate`
```bash
curl -X POST "http://localhost:4000/v1/wallets/<wallet_id>/withdrawals/simulate" \
  -H "Authorization: Bearer <token>" \
  -H "X-Tenant-ID: tenant-1" \
  -H "Content-Type: application/json" \
  -d '{"destination":"0x1111111111111111111111111111111111111111","amount_minor":100000,"asset":"USDC"}'
```

#### `GET /v1/withdrawals/:id/trace`
```bash
curl "http://localhost:4000/v1/withdrawals/<withdrawal_id>/trace" \
  -H "Authorization: Bearer <token>" \
  -H "X-Tenant-ID: tenant-1"
```

#### `GET /v1/wallets/:id/token-allowances`
```bash
curl "http://localhost:4000/v1/wallets/<wallet_id>/token-allowances?asset=USDC&spender=0x1111111111111111111111111111111111111111" \
  -H "Authorization: Bearer <token>" \
  -H "X-Tenant-ID: tenant-1"
```

#### `GET /v1/networks/:network/status`
```bash
curl "http://localhost:4000/v1/networks/eth-sepolia/status" \
  -H "Authorization: Bearer <token>"
```

#### `GET /v1/wallets/:id/risk-score`
```bash
curl "http://localhost:4000/v1/wallets/<wallet_id>/risk-score" \
  -H "Authorization: Bearer <token>" \
  -H "X-Tenant-ID: tenant-1"
```

#### `POST /v1/addresses/validate`
```bash
curl -X POST "http://localhost:4000/v1/addresses/validate" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"network":"eth-sepolia","address":"0x1111111111111111111111111111111111111111"}'
```

#### `GET /v1/assets/:network/:asset/metadata`
```bash
curl "http://localhost:4000/v1/assets/eth-sepolia/USDC/metadata" \
  -H "Authorization: Bearer <token>"
```

#### `GET /v1/wallets/:id/nonces`
```bash
curl "http://localhost:4000/v1/wallets/<wallet_id>/nonces" \
  -H "Authorization: Bearer <token>" \
  -H "X-Tenant-ID: tenant-1"
```

Pending-only view for a tenant:
```bash
curl "http://localhost:4000/v1/withdrawals?status=policy_pending&page=1&page_size=20&sort=-created_at" \
  -H "Authorization: Bearer <token>" \
  -H "X-Tenant-ID: tenant-1"
```

#### `POST /v1/withdrawals/:id/approve`
```bash
curl -X POST http://localhost:4000/v1/withdrawals/<withdrawal_id>/approve \
  -H "Authorization: Bearer <token>" \
  -H "X-Tenant-ID: tenant-1" \
  -H "Content-Type: application/json" \
  -d '{"reason":"approved for payout"}'
```

#### `POST /v1/withdrawals/:id/reject`
```bash
curl -X POST http://localhost:4000/v1/withdrawals/<withdrawal_id>/reject \
  -H "Authorization: Bearer <token>" \
  -H "X-Tenant-ID: tenant-1" \
  -H "Content-Type: application/json" \
  -d '{"reason":"manual rejection"}'
```

### Related References

- API inventory: `documentation/api-surface-current.md`
- Full Postman collection: `documentation/postman/Nduracore-Current-API.postman_collection.json`
- Full Postman environment: `documentation/postman/Nduracore-Current-API.postman_environment.json`
- Infra updates (N2/N4/N5/N6/N7): `documentation/simple-updates-2026-04-19.md`
- Infra updates Postman collection: `documentation/postman/Nduracore-Infra-Updates.postman_collection.json`

### Asset Registry Sync

Use the registry sync command to bulk upsert token coverage from env-configured catalogs:

```bash
go run ./cmd/assetsync
```

Reusable cross-platform catalog command:

```bash
go run ./cmd/assetsync -catalog ./documentation/asset_catalog.nduracore.json --from-env-erc20=false
```

Generate and load a 1000-asset catalog (popular chains + live/testnet native seeds):

```bash
go run ./cmd/assetcataloggen -out ./documentation/asset_catalog.top1000.json -limit 1000
go run ./cmd/assetsync -catalog ./documentation/asset_catalog.top1000.json --from-env-erc20=false --reset-active
```

Required:
- `DB_DSN`

When syncing from env (`--from-env-erc20=true`, default):
- `ALCHEMY_ERC20_CONTRACTS_JSON` (`{"network:ASSET":"0xContractAddress"}`)

Optional:
- `ALCHEMY_ERC20_DECIMALS_JSON` (`{"network:ASSET":6}`)
