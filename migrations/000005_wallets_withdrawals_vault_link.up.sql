ALTER TABLE wallets
    ADD COLUMN IF NOT EXISTS vault_id UUID REFERENCES vaults(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_wallets_tenant_vault_asset_network
    ON wallets(tenant_id, vault_id, asset, network);

ALTER TABLE withdrawals
    ADD COLUMN IF NOT EXISTS vault_id UUID REFERENCES vaults(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_withdrawals_tenant_vault_created
    ON withdrawals(tenant_id, vault_id, created_at DESC);
