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
- `ALCHEMY_CONFIRMATIONS_REQUIRED`
- `ALCHEMY_RPC_URL`
- `ALCHEMY_API_KEY`
- `ALCHEMY_WEBHOOK_SIGNING_SECRET`

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
