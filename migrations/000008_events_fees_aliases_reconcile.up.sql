CREATE TABLE IF NOT EXISTS vault_aliases (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id TEXT NOT NULL,
    vault_id UUID NOT NULL REFERENCES vaults(id) ON DELETE CASCADE,
    provider TEXT NOT NULL,
    external_vault_id TEXT NOT NULL,
    external_name TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (tenant_id, provider, external_vault_id)
);

CREATE INDEX IF NOT EXISTS idx_vault_aliases_tenant_vault ON vault_aliases(tenant_id, vault_id);

CREATE TABLE IF NOT EXISTS fee_policies (
    tenant_id TEXT NOT NULL,
    asset TEXT NOT NULL,
    network TEXT NOT NULL,
    fee_bps INT NOT NULL DEFAULT 0,
    min_fee_minor BIGINT NOT NULL DEFAULT 0,
    max_fee_minor BIGINT NOT NULL DEFAULT 0,
    fee_destination TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (tenant_id, asset, network)
);

CREATE TABLE IF NOT EXISTS fee_charges (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    withdrawal_id UUID NOT NULL REFERENCES withdrawals(id) ON DELETE CASCADE,
    tenant_id TEXT NOT NULL,
    gross_amount_minor BIGINT NOT NULL,
    fee_amount_minor BIGINT NOT NULL,
    net_amount_minor BIGINT NOT NULL,
    status TEXT NOT NULL DEFAULT 'planned',
    fee_tx_hash TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (withdrawal_id)
);

CREATE INDEX IF NOT EXISTS idx_fee_charges_tenant_created ON fee_charges(tenant_id, created_at DESC);

CREATE TABLE IF NOT EXISTS withdrawal_state_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    withdrawal_id UUID NOT NULL REFERENCES withdrawals(id) ON DELETE CASCADE,
    tenant_id TEXT NOT NULL,
    from_state TEXT NOT NULL,
    to_state TEXT NOT NULL,
    reason TEXT NOT NULL DEFAULT '',
    provider_status TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_withdrawal_state_history_withdrawal_created
ON withdrawal_state_history(withdrawal_id, created_at DESC);

CREATE TABLE IF NOT EXISTS provider_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    provider TEXT NOT NULL,
    tenant_id TEXT NOT NULL,
    dedupe_key TEXT NOT NULL,
    signature_valid BOOLEAN NOT NULL DEFAULT false,
    payload JSONB NOT NULL,
    headers JSONB NOT NULL DEFAULT '{}'::jsonb,
    status TEXT NOT NULL DEFAULT 'accepted',
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (provider, dedupe_key)
);

CREATE INDEX IF NOT EXISTS idx_provider_events_tenant_received ON provider_events(tenant_id, received_at DESC);

CREATE TABLE IF NOT EXISTS normalized_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    provider_event_id UUID NOT NULL REFERENCES provider_events(id) ON DELETE CASCADE,
    provider TEXT NOT NULL,
    tenant_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    wallet_id UUID,
    withdrawal_id UUID,
    tx_hash TEXT NOT NULL DEFAULT '',
    amount_minor BIGINT NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT '',
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_normalized_events_tenant_created ON normalized_events(tenant_id, created_at DESC);

CREATE TABLE IF NOT EXISTS event_failures (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    provider_event_id UUID NOT NULL REFERENCES provider_events(id) ON DELETE CASCADE,
    tenant_id TEXT NOT NULL,
    error_message TEXT NOT NULL,
    retry_count INT NOT NULL DEFAULT 0,
    next_retry_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS event_dlq (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    provider_event_id UUID NOT NULL REFERENCES provider_events(id) ON DELETE CASCADE,
    tenant_id TEXT NOT NULL,
    provider TEXT NOT NULL,
    reason TEXT NOT NULL,
    payload JSONB NOT NULL,
    retry_count INT NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'queued',
    last_error TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_event_dlq_tenant_status ON event_dlq(tenant_id, status, created_at DESC);

CREATE TABLE IF NOT EXISTS reconciliation_runs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id TEXT NOT NULL,
    provider TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'running',
    summary JSONB NOT NULL DEFAULT '{}'::jsonb,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS reconciliation_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    run_id UUID NOT NULL REFERENCES reconciliation_runs(id) ON DELETE CASCADE,
    tenant_id TEXT NOT NULL,
    object_type TEXT NOT NULL,
    object_id TEXT NOT NULL,
    expected_state TEXT NOT NULL,
    provider_state TEXT NOT NULL,
    mismatch_reason TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_reconciliation_items_run_created ON reconciliation_items(run_id, created_at DESC);
