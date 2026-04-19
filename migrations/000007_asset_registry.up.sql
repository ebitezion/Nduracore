CREATE TABLE IF NOT EXISTS asset_registry (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    network TEXT NOT NULL,
    asset_code TEXT NOT NULL,
    chain_family TEXT NOT NULL,
    asset_type TEXT NOT NULL DEFAULT 'native',
    contract_address TEXT NOT NULL DEFAULT '',
    decimals INT NOT NULL DEFAULT 18,
    is_active BOOLEAN NOT NULL DEFAULT true,
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (network, asset_code)
);

CREATE INDEX IF NOT EXISTS idx_asset_registry_network_active ON asset_registry(network, is_active);
CREATE INDEX IF NOT EXISTS idx_asset_registry_family_active ON asset_registry(chain_family, is_active);
