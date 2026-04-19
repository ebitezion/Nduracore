# Environments and Secrets

## Environments
Nduracore currently supports:
- `development`
- `test`
- `staging`
- `production`

Set with `MY_ENV`.

## Required Environment Variables
- `APP_NAME`
- `APP_VERSION`
- `PORT`
- `MY_ENV`
- `DB_DSN`
- `TOKEN_SECRET`
- `TOKEN_ISSUER`
- `TOKEN_AUDIENCE`
- `TOKEN_TTL`

## Optional Environment Variables
- `REDIS_ADDR`
- `REDIS_PASSWORD`
- `REDIS_DB`
- `REDIS_QUEUE_KEY`
- `OTEL_ENABLED`
- `OTEL_EXPORTER_OTLP_ENDPOINT`
- `OTEL_SAMPLE_RATIO`
- `TRUSTED_PROXIES`
- `ALCHEMY_NETWORK`
- `ALCHEMY_ENABLE_LIVE_DEPOSITS`
- `ALCHEMY_ENABLE_LIVE_BROADCASTS`
- `ALCHEMY_CONFIRMATIONS_REQUIRED`
- `ALCHEMY_RPC_URL`
- `ALCHEMY_API_KEY`
- `ALCHEMY_RPC_URLS_JSON`
- `ALCHEMY_API_KEYS_JSON`
- `ALCHEMY_PRIVATE_KEY`
- `ALCHEMY_PRIVATE_KEYS_JSON`
- `ALCHEMY_VAULT_PRIVATE_KEYS_JSON`
- `ALCHEMY_ERC20_CONTRACTS_JSON`
- `ALCHEMY_WEBHOOK_SIGNING_SECRET`

## Suggested Multi-Chain Targets (Alchemy EVM)
Use these network keys in `ALCHEMY_RPC_URLS_JSON`, `ALCHEMY_API_KEYS_JSON`, `ALCHEMY_PRIVATE_KEYS_JSON`, and `ALCHEMY_ERC20_CONTRACTS_JSON`:

- `eth-mainnet`
- `eth-sepolia`
- `polygon-mainnet`
- `polygon-amoy`
- `arbitrum-mainnet`
- `arbitrum-sepolia`
- `optimism-mainnet`
- `optimism-sepolia`
- `base-mainnet`
- `base-sepolia`
- `zksync-mainnet`
- `zksync-sepolia`
- `linea-mainnet`
- `linea-sepolia`
- `mantle-mainnet`
- `mantle-sepolia`
- `scroll-mainnet`
- `scroll-sepolia`
- `blast-mainnet`
- `worldchain-mainnet`

## Live Broadcast Validation
When `ALCHEMY_ENABLE_LIVE_BROADCASTS=true`, startup validation enforces:
1. At least one RPC endpoint is configured from `ALCHEMY_RPC_URL`, `ALCHEMY_RPC_URLS_JSON`, `ALCHEMY_API_KEY`, or `ALCHEMY_API_KEYS_JSON`.
2. Every configured broadcast network has a signer key (`ALCHEMY_VAULT_PRIVATE_KEYS_JSON`, `ALCHEMY_PRIVATE_KEYS_JSON`, or `ALCHEMY_PRIVATE_KEY` for the default `ALCHEMY_NETWORK`).
3. `ALCHEMY_ERC20_CONTRACTS_JSON` entries use flat `network:ASSET` keys and valid `0x...` addresses.

Vault-scoped signer format in `ALCHEMY_VAULT_PRIVATE_KEYS_JSON`:
1. Key: `vault_id:network`
2. Value: private key/seed for that network family
3. Example: `{"b0fdf2d7-7f11-4fca-8c23-0fd1f8133cc8:eth-sepolia":"<private-key-hex>"}`

## Live Broadcast Chain Families
Implemented live broadcast families:
1. EVM (`eth-*`, `polygon-*`, `arbitrum-*`, `optimism-*`, `base-*`, `zksync-*`, `linea-*`, `mantle-*`, `scroll-*`, `blast-*`, `worldchain-*`)
2. Solana (`solana-*`) - native `SOL` transfers
3. Stellar (`stellar-*`) - native `XLM` transfers
4. XRPL (`xrpl-*` / `xrp-*`) - native `XRP` transfers
5. Tron (`tron-*`) - native `TRX` transfers
6. Aptos (`aptos-*`) - native `APT` transfers
7. Sui (`sui-*`) - native `SUI` transfers

Signer key format in `ALCHEMY_PRIVATE_KEYS_JSON`:
1. EVM networks: hex private key (with or without `0x`)
2. Solana networks: base58 secret key (or keygen JSON array)
3. Stellar networks: seed starting with `S...`
4. XRPL networks: seed starting with `s...`
5. Tron networks: secp256k1 private key hex (with or without `0x`)
6. Aptos networks: ed25519 private key hex (32-byte seed or 64-byte key)
7. Sui networks: `suiprivkey...` secret or ed25519 private key hex

Before rollout, run:
```bash
make alchemy-smoke
```
This probes every configured network and fails fast on unreachable/misconfigured RPC endpoints (EVM: `eth_chainId`, Solana: `getHealth`, XRPL: `server_info`, Stellar/Aptos: HTTP health check, Sui: `suix_getLatestSuiSystemState`, Tron: `/wallet/getnodeinfo`).

## Secret Management Policy
1. Never commit live secrets to source control.
2. Use `*_FILE` environment support for mounted secrets where available:
   - `DB_DSN_FILE`
   - `TOKEN_SECRET_FILE`
   - `REDIS_PASSWORD_FILE`
3. Use `*_REF` support when secrets are exposed through controlled environment references:
   - `DB_DSN_REF=env:STAGING_DB_DSN`
   - `TOKEN_SECRET_REF=env:STAGING_TOKEN_SECRET`
   - `REDIS_PASSWORD_REF=env:STAGING_REDIS_PASSWORD`
   - `ALCHEMY_API_KEY_REF=env:STAGING_ALCHEMY_API_KEY`
   - `ALCHEMY_WEBHOOK_SIGNING_SECRET_REF=env:STAGING_ALCHEMY_WEBHOOK_SECRET`
   - Supported schemes: `env:` and `file:`
4. In staging and production:
   - Rotate `TOKEN_SECRET` at least quarterly.
   - Enforce 32+ char entropy requirements.
   - Restrict secret access to deployment runtime only.
5. For CI:
   - Store credentials in GitHub Actions Secrets.
   - Use scoped service credentials with least privilege.

## Naming Conventions
- App identity: `nduracore`
- Queue prefix: `nduracore:*`
- Test DB: `nduracore_test`
