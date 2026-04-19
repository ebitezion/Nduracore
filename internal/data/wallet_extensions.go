package data

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

type FeePolicy struct {
	TenantID       string    `json:"tenant_id"`
	Asset          string    `json:"asset"`
	Network        string    `json:"network"`
	FeeBps         int       `json:"fee_bps"`
	MinFeeMinor    int64     `json:"min_fee_minor"`
	MaxFeeMinor    int64     `json:"max_fee_minor"`
	FeeDestination string    `json:"fee_destination"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type FeeCharge struct {
	ID               string    `json:"id"`
	WithdrawalID     string    `json:"withdrawal_id"`
	TenantID         string    `json:"tenant_id"`
	GrossAmountMinor int64     `json:"gross_amount_minor"`
	FeeAmountMinor   int64     `json:"fee_amount_minor"`
	NetAmountMinor   int64     `json:"net_amount_minor"`
	Status           string    `json:"status"`
	FeeTxHash        string    `json:"fee_tx_hash"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type WithdrawalStateHistory struct {
	ID             string    `json:"id"`
	WithdrawalID   string    `json:"withdrawal_id"`
	TenantID       string    `json:"tenant_id"`
	FromState      string    `json:"from_state"`
	ToState        string    `json:"to_state"`
	Reason         string    `json:"reason"`
	ProviderStatus string    `json:"provider_status"`
	CreatedAt      time.Time `json:"created_at"`
}

func (m WalletModel) GetFeePolicy(ctx context.Context, tenantID, asset, network string) (*FeePolicy, error) {
	ctx, cancel := NewTimeoutContext(ctx)
	defer cancel()

	stmt := `SELECT tenant_id, asset, network, fee_bps, min_fee_minor, max_fee_minor, fee_destination, updated_at
		FROM fee_policies
		WHERE tenant_id = $1 AND asset = $2 AND network = $3`

	var item FeePolicy
	err := m.DB.QueryRowContext(ctx, stmt,
		strings.TrimSpace(tenantID),
		strings.ToUpper(strings.TrimSpace(asset)),
		strings.ToLower(strings.TrimSpace(network)),
	).Scan(&item.TenantID, &item.Asset, &item.Network, &item.FeeBps, &item.MinFeeMinor, &item.MaxFeeMinor, &item.FeeDestination, &item.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &item, nil
}

func (m WalletModel) UpsertFeePolicy(ctx context.Context, policy FeePolicy) error {
	ctx, cancel := NewTimeoutContext(ctx)
	defer cancel()

	stmt := `INSERT INTO fee_policies (tenant_id, asset, network, fee_bps, min_fee_minor, max_fee_minor, fee_destination, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		ON CONFLICT (tenant_id, asset, network)
		DO UPDATE SET fee_bps = EXCLUDED.fee_bps,
			min_fee_minor = EXCLUDED.min_fee_minor,
			max_fee_minor = EXCLUDED.max_fee_minor,
			fee_destination = EXCLUDED.fee_destination,
			updated_at = NOW()`

	_, err := m.DB.ExecContext(ctx, stmt,
		strings.TrimSpace(policy.TenantID),
		strings.ToUpper(strings.TrimSpace(policy.Asset)),
		strings.ToLower(strings.TrimSpace(policy.Network)),
		policy.FeeBps,
		policy.MinFeeMinor,
		policy.MaxFeeMinor,
		strings.TrimSpace(policy.FeeDestination),
	)
	return err
}

func (m WalletModel) CreateFeeCharge(ctx context.Context, charge *FeeCharge) error {
	ctx, cancel := NewTimeoutContext(ctx)
	defer cancel()

	stmt := `INSERT INTO fee_charges (withdrawal_id, tenant_id, gross_amount_minor, fee_amount_minor, net_amount_minor, status, fee_tx_hash)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`

	return m.DB.QueryRowContext(ctx, stmt,
		strings.TrimSpace(charge.WithdrawalID),
		strings.TrimSpace(charge.TenantID),
		charge.GrossAmountMinor,
		charge.FeeAmountMinor,
		charge.NetAmountMinor,
		strings.TrimSpace(charge.Status),
		strings.TrimSpace(charge.FeeTxHash),
	).Scan(&charge.ID, &charge.CreatedAt, &charge.UpdatedAt)
}

func (m WalletModel) GetFeeChargeByWithdrawal(ctx context.Context, tenantID, withdrawalID string) (*FeeCharge, error) {
	ctx, cancel := NewTimeoutContext(ctx)
	defer cancel()

	stmt := `SELECT id, withdrawal_id::text, tenant_id, gross_amount_minor, fee_amount_minor, net_amount_minor, status, fee_tx_hash, created_at, updated_at
		FROM fee_charges
		WHERE tenant_id = $1 AND withdrawal_id = $2::uuid`

	var charge FeeCharge
	err := m.DB.QueryRowContext(ctx, stmt,
		strings.TrimSpace(tenantID),
		strings.TrimSpace(withdrawalID),
	).Scan(&charge.ID, &charge.WithdrawalID, &charge.TenantID, &charge.GrossAmountMinor, &charge.FeeAmountMinor, &charge.NetAmountMinor, &charge.Status, &charge.FeeTxHash, &charge.CreatedAt, &charge.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return &charge, nil
}

func (m WalletModel) AddWithdrawalStateHistory(ctx context.Context, item *WithdrawalStateHistory) error {
	ctx, cancel := NewTimeoutContext(ctx)
	defer cancel()

	stmt := `INSERT INTO withdrawal_state_history (withdrawal_id, tenant_id, from_state, to_state, reason, provider_status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at`

	return m.DB.QueryRowContext(ctx, stmt,
		strings.TrimSpace(item.WithdrawalID),
		strings.TrimSpace(item.TenantID),
		strings.ToLower(strings.TrimSpace(item.FromState)),
		strings.ToLower(strings.TrimSpace(item.ToState)),
		strings.TrimSpace(item.Reason),
		strings.TrimSpace(item.ProviderStatus),
	).Scan(&item.ID, &item.CreatedAt)
}
