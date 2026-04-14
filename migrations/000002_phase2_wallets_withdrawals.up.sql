CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id TEXT NOT NULL,
    asset TEXT NOT NULL,
    network TEXT NOT NULL,
    address TEXT NOT NULL UNIQUE,
    provider TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_wallets_tenant_created_at ON wallets(tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_wallets_tenant_asset_network ON wallets(tenant_id, asset, network);

CREATE TABLE IF NOT EXISTS wallet_deposits (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    tenant_id TEXT NOT NULL,
    asset TEXT NOT NULL,
    network TEXT NOT NULL,
    tx_hash TEXT NOT NULL,
    amount_minor BIGINT NOT NULL,
    confirmations INT NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'pending',
    detected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(wallet_id, tx_hash)
);

CREATE INDEX IF NOT EXISTS idx_wallet_deposits_wallet_detected ON wallet_deposits(wallet_id, detected_at DESC);
CREATE INDEX IF NOT EXISTS idx_wallet_deposits_tenant_detected ON wallet_deposits(tenant_id, detected_at DESC);

CREATE TABLE IF NOT EXISTS withdrawals (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id TEXT NOT NULL,
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE RESTRICT,
    destination TEXT NOT NULL,
    asset TEXT NOT NULL,
    network TEXT NOT NULL,
    amount_minor BIGINT NOT NULL,
    status TEXT NOT NULL,
    required_approvals INT NOT NULL DEFAULT 0,
    approved_count INT NOT NULL DEFAULT 0,
    policy_reason TEXT NOT NULL DEFAULT '',
    risk_level TEXT NOT NULL DEFAULT 'low',
    provider_tx_hash TEXT NOT NULL DEFAULT '',
    requested_by TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_withdrawals_tenant_created ON withdrawals(tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_withdrawals_wallet_created ON withdrawals(wallet_id, created_at DESC);

CREATE TABLE IF NOT EXISTS withdrawal_approvals (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    withdrawal_id UUID NOT NULL REFERENCES withdrawals(id) ON DELETE CASCADE,
    approved_by TEXT NOT NULL,
    decision TEXT NOT NULL,
    reason TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(withdrawal_id, approved_by)
);

CREATE INDEX IF NOT EXISTS idx_withdrawal_approvals_withdrawal ON withdrawal_approvals(withdrawal_id, created_at DESC);

CREATE TABLE IF NOT EXISTS policy_rules (
    tenant_id TEXT PRIMARY KEY,
    whitelist_destinations JSONB NOT NULL DEFAULT '[]'::jsonb,
    per_tx_limit_minor BIGINT NOT NULL,
    daily_limit_minor BIGINT NOT NULL,
    velocity_count INT NOT NULL,
    velocity_window_minutes INT NOT NULL,
    approval_tier TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
