ALTER TABLE vault_assets
    DROP CONSTRAINT IF EXISTS vault_assets_vault_id_asset_code_network_key;

ALTER TABLE vault_assets
    ADD CONSTRAINT vault_assets_vault_id_asset_code_key UNIQUE (vault_id, asset_code);
