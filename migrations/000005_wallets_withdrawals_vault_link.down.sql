DROP INDEX IF EXISTS idx_withdrawals_tenant_vault_created;

ALTER TABLE withdrawals
    DROP COLUMN IF EXISTS vault_id;

DROP INDEX IF EXISTS idx_wallets_tenant_vault_asset_network;

ALTER TABLE wallets
    DROP COLUMN IF EXISTS vault_id;
