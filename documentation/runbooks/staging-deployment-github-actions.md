# Staging Deployment via GitHub Actions

## Workflow
File: `.github/workflows/staging.yml`

Triggers:
- Push to `main`
- Manual run (`workflow_dispatch`)

Jobs:
1. `verify`: formatting, migration checks, and test suite.
2. `package`: builds Linux `amd64` binary and uploads artifact.
3. `deploy`: runs staging deploy command and optional health check.

## Required GitHub Environment
Create a GitHub environment named `staging` and configure:
- Approval rules (recommended).
- Environment secrets below.

## Staging Secrets
- `STAGING_DEPLOY_COMMAND`:
  - Shell command executed by deploy job.
  - Example: command to copy artifact and restart service in staging.
- `STAGING_HEALTHCHECK_URL` (optional):
  - URL checked after deploy to verify service health.

## Recommended Deploy Command Contract
`STAGING_DEPLOY_COMMAND` should:
1. Pull or receive artifact generated in CI.
2. Replace running binary/service.
3. Restart process manager (`systemd`, container runtime, etc.).
4. Exit non-zero on failure.

## Rollback Guidance
- Redeploy previous known-good artifact/tag.
- Re-run health check.
- Open incident note if rollback was required.
