package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

type WalletModel struct {
	DB *sql.DB
}

type Wallet struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	VaultID   string    `json:"vault_id,omitempty"`
	Asset     string    `json:"asset"`
	Network   string    `json:"network"`
	Address   string    `json:"address"`
	Provider  string    `json:"provider"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type WalletDeposit struct {
	ID            string    `json:"id"`
	WalletID      string    `json:"wallet_id"`
	TenantID      string    `json:"tenant_id"`
	Asset         string    `json:"asset"`
	Network       string    `json:"network"`
	TxHash        string    `json:"tx_hash"`
	AmountMinor   int64     `json:"amount_minor"`
	Confirmations int       `json:"confirmations"`
	Status        string    `json:"status"`
	DetectedAt    time.Time `json:"detected_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type WalletBalance struct {
	WalletID     string `json:"wallet_id"`
	TenantID     string `json:"tenant_id"`
	VaultID      string `json:"vault_id,omitempty"`
	Address      string `json:"address"`
	Asset        string `json:"asset"`
	Network      string `json:"network"`
	BalanceMinor string `json:"balance_minor"`
}

type Withdrawal struct {
	ID                string    `json:"id"`
	TenantID          string    `json:"tenant_id"`
	WalletID          string    `json:"wallet_id"`
	VaultID           string    `json:"vault_id,omitempty"`
	Destination       string    `json:"destination"`
	Asset             string    `json:"asset"`
	Network           string    `json:"network"`
	AmountMinor       int64     `json:"amount_minor"`
	Status            string    `json:"status"`
	RequiredApprovals int       `json:"required_approvals"`
	ApprovedCount     int       `json:"approved_count"`
	PolicyReason      string    `json:"policy_reason"`
	RiskLevel         string    `json:"risk_level"`
	ProviderTxHash    string    `json:"provider_tx_hash"`
	RequestedBy       string    `json:"requested_by"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type WithdrawalApproval struct {
	ID           string    `json:"id"`
	WithdrawalID string    `json:"withdrawal_id"`
	ApprovedBy   string    `json:"approved_by"`
	Decision     string    `json:"decision"`
	Reason       string    `json:"reason"`
	CreatedAt    time.Time `json:"created_at"`
}

type PolicyRule struct {
	TenantID              string    `json:"tenant_id"`
	WhitelistDestinations []string  `json:"whitelist_destinations"`
	PerTxLimitMinor       int64     `json:"per_tx_limit_minor"`
	DailyLimitMinor       int64     `json:"daily_limit_minor"`
	VelocityCount         int       `json:"velocity_count"`
	VelocityWindowMinutes int       `json:"velocity_window_minutes"`
	ApprovalTier          string    `json:"approval_tier"`
	UpdatedAt             time.Time `json:"updated_at"`
}

func (m WalletModel) CreateWallet(ctx context.Context, wallet *Wallet) error {
	ctx, cancel := NewTimeoutContext(ctx)
	defer cancel()

	stmt := `INSERT INTO wallets (tenant_id, vault_id, asset, network, address, provider, status)
		VALUES ($1, NULLIF($2, '')::uuid, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`

	return m.DB.QueryRowContext(ctx, stmt,
		wallet.TenantID,
		wallet.VaultID,
		wallet.Asset,
		wallet.Network,
		wallet.Address,
		wallet.Provider,
		wallet.Status,
	).Scan(&wallet.ID, &wallet.CreatedAt, &wallet.UpdatedAt)
}

func (m WalletModel) GetWallet(ctx context.Context, tenantID, walletID string) (*Wallet, error) {
	ctx, cancel := NewTimeoutContext(ctx)
	defer cancel()

	stmt := `SELECT id, tenant_id, COALESCE(vault_id::text, ''), asset, network, address, provider, status, created_at, updated_at
		FROM wallets
		WHERE tenant_id = $1 AND id = $2`

	var wallet Wallet
	err := m.DB.QueryRowContext(ctx, stmt, tenantID, walletID).Scan(
		&wallet.ID,
		&wallet.TenantID,
		&wallet.VaultID,
		&wallet.Asset,
		&wallet.Network,
		&wallet.Address,
		&wallet.Provider,
		&wallet.Status,
		&wallet.CreatedAt,
		&wallet.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &wallet, nil
}

func (m WalletModel) GetWalletByVaultAssetNetwork(ctx context.Context, tenantID, vaultID, asset, network string) (*Wallet, error) {
	ctx, cancel := NewTimeoutContext(ctx)
	defer cancel()

	stmt := `SELECT id, tenant_id, COALESCE(vault_id::text, ''), asset, network, address, provider, status, created_at, updated_at
		FROM wallets
		WHERE tenant_id = $1 AND vault_id = $2::uuid AND asset = $3 AND network = $4 AND status = 'active'
		ORDER BY created_at DESC
		LIMIT 1`

	var wallet Wallet
	err := m.DB.QueryRowContext(ctx, stmt, tenantID, vaultID, asset, network).Scan(
		&wallet.ID,
		&wallet.TenantID,
		&wallet.VaultID,
		&wallet.Asset,
		&wallet.Network,
		&wallet.Address,
		&wallet.Provider,
		&wallet.Status,
		&wallet.CreatedAt,
		&wallet.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &wallet, nil
}

func (m WalletModel) ListWallets(ctx context.Context, tenantID, vaultID, asset, network string, filters Filters) ([]Wallet, Metadata, error) {
	ctx, cancel := NewTimeoutContext(ctx)
	defer cancel()

	stmt := `SELECT count(*) OVER(), id, tenant_id, COALESCE(vault_id::text, ''), asset, network, address, provider, status, created_at, updated_at
		FROM wallets
		WHERE tenant_id = $1`
	args := []interface{}{tenantID}
	argPosition := 2

	if strings.TrimSpace(vaultID) != "" {
		stmt += fmt.Sprintf(" AND vault_id = $%d::uuid", argPosition)
		args = append(args, vaultID)
		argPosition++
	}
	if strings.TrimSpace(asset) != "" {
		stmt += fmt.Sprintf(" AND asset = $%d", argPosition)
		args = append(args, asset)
		argPosition++
	}
	if strings.TrimSpace(network) != "" {
		stmt += fmt.Sprintf(" AND network = $%d", argPosition)
		args = append(args, network)
		argPosition++
	}

	stmt += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argPosition, argPosition+1)
	args = append(args, filters.PageSize, (filters.Page-1)*filters.PageSize)

	rows, err := m.DB.QueryContext(ctx, stmt, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	wallets := make([]Wallet, 0)
	total := 0
	for rows.Next() {
		var wallet Wallet
		if err := rows.Scan(
			&total,
			&wallet.ID,
			&wallet.TenantID,
			&wallet.VaultID,
			&wallet.Asset,
			&wallet.Network,
			&wallet.Address,
			&wallet.Provider,
			&wallet.Status,
			&wallet.CreatedAt,
			&wallet.UpdatedAt,
		); err != nil {
			return nil, Metadata{}, err
		}
		wallets = append(wallets, wallet)
	}
	if err := rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	return wallets, CalculateMetadata(total, filters.Page, filters.PageSize), nil
}

func (m WalletModel) UpsertDeposit(ctx context.Context, dep *WalletDeposit) error {
	ctx, cancel := NewTimeoutContext(ctx)
	defer cancel()

	stmt := `INSERT INTO wallet_deposits (wallet_id, tenant_id, asset, network, tx_hash, amount_minor, confirmations, status, detected_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		ON CONFLICT (wallet_id, tx_hash)
		DO UPDATE SET confirmations = EXCLUDED.confirmations, status = EXCLUDED.status, updated_at = NOW()
		RETURNING id, detected_at, updated_at`

	return m.DB.QueryRowContext(ctx, stmt,
		dep.WalletID,
		dep.TenantID,
		dep.Asset,
		dep.Network,
		dep.TxHash,
		dep.AmountMinor,
		dep.Confirmations,
		dep.Status,
	).Scan(&dep.ID, &dep.DetectedAt, &dep.UpdatedAt)
}

func (m WalletModel) ListDeposits(ctx context.Context, tenantID, walletID string, filters Filters) ([]WalletDeposit, Metadata, error) {
	ctx, cancel := NewTimeoutContext(ctx)
	defer cancel()

	stmt := `SELECT count(*) OVER(), id, wallet_id, tenant_id, asset, network, tx_hash, amount_minor, confirmations, status, detected_at, updated_at
		FROM wallet_deposits
		WHERE tenant_id = $1 AND wallet_id = $2
		ORDER BY detected_at DESC
		LIMIT $3 OFFSET $4`

	rows, err := m.DB.QueryContext(ctx, stmt, tenantID, walletID, filters.PageSize, (filters.Page-1)*filters.PageSize)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	deposits := make([]WalletDeposit, 0)
	total := 0
	for rows.Next() {
		var dep WalletDeposit
		if err := rows.Scan(
			&total,
			&dep.ID,
			&dep.WalletID,
			&dep.TenantID,
			&dep.Asset,
			&dep.Network,
			&dep.TxHash,
			&dep.AmountMinor,
			&dep.Confirmations,
			&dep.Status,
			&dep.DetectedAt,
			&dep.UpdatedAt,
		); err != nil {
			return nil, Metadata{}, err
		}
		deposits = append(deposits, dep)
	}
	if err := rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	return deposits, CalculateMetadata(total, filters.Page, filters.PageSize), nil
}

func (m WalletModel) CreateWithdrawal(ctx context.Context, withdrawal *Withdrawal) error {
	ctx, cancel := NewTimeoutContext(ctx)
	defer cancel()

	stmt := `INSERT INTO withdrawals (tenant_id, wallet_id, vault_id, destination, asset, network, amount_minor, status, required_approvals, approved_count, policy_reason, risk_level, provider_tx_hash, requested_by)
		VALUES ($1, $2, NULLIF($3, '')::uuid, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, created_at, updated_at`

	return m.DB.QueryRowContext(ctx, stmt,
		withdrawal.TenantID,
		withdrawal.WalletID,
		withdrawal.VaultID,
		withdrawal.Destination,
		withdrawal.Asset,
		withdrawal.Network,
		withdrawal.AmountMinor,
		withdrawal.Status,
		withdrawal.RequiredApprovals,
		withdrawal.ApprovedCount,
		withdrawal.PolicyReason,
		withdrawal.RiskLevel,
		withdrawal.ProviderTxHash,
		withdrawal.RequestedBy,
	).Scan(&withdrawal.ID, &withdrawal.CreatedAt, &withdrawal.UpdatedAt)
}

func (m WalletModel) GetWithdrawal(ctx context.Context, tenantID, withdrawalID string) (*Withdrawal, error) {
	ctx, cancel := NewTimeoutContext(ctx)
	defer cancel()

	stmt := `SELECT id, tenant_id, wallet_id, COALESCE(vault_id::text, ''), destination, asset, network, amount_minor, status, required_approvals, approved_count, policy_reason, risk_level, provider_tx_hash, requested_by, created_at, updated_at
		FROM withdrawals
		WHERE tenant_id = $1 AND id = $2`

	var wd Withdrawal
	err := m.DB.QueryRowContext(ctx, stmt, tenantID, withdrawalID).Scan(
		&wd.ID,
		&wd.TenantID,
		&wd.WalletID,
		&wd.VaultID,
		&wd.Destination,
		&wd.Asset,
		&wd.Network,
		&wd.AmountMinor,
		&wd.Status,
		&wd.RequiredApprovals,
		&wd.ApprovedCount,
		&wd.PolicyReason,
		&wd.RiskLevel,
		&wd.ProviderTxHash,
		&wd.RequestedBy,
		&wd.CreatedAt,
		&wd.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &wd, nil
}

func (m WalletModel) ListRecentWithdrawals(ctx context.Context, tenantID string, since time.Time) ([]Withdrawal, error) {
	ctx, cancel := NewTimeoutContext(ctx)
	defer cancel()

	stmt := `SELECT id, tenant_id, wallet_id, COALESCE(vault_id::text, ''), destination, asset, network, amount_minor, status, required_approvals, approved_count, policy_reason, risk_level, provider_tx_hash, requested_by, created_at, updated_at
		FROM withdrawals
		WHERE tenant_id = $1 AND created_at >= $2`

	rows, err := m.DB.QueryContext(ctx, stmt, tenantID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]Withdrawal, 0)
	for rows.Next() {
		var wd Withdrawal
		if err := rows.Scan(
			&wd.ID,
			&wd.TenantID,
			&wd.WalletID,
			&wd.VaultID,
			&wd.Destination,
			&wd.Asset,
			&wd.Network,
			&wd.AmountMinor,
			&wd.Status,
			&wd.RequiredApprovals,
			&wd.ApprovedCount,
			&wd.PolicyReason,
			&wd.RiskLevel,
			&wd.ProviderTxHash,
			&wd.RequestedBy,
			&wd.CreatedAt,
			&wd.UpdatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, wd)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (m WalletModel) ListWithdrawals(ctx context.Context, tenantID, status string, filters Filters) ([]Withdrawal, Metadata, error) {
	ctx, cancel := NewTimeoutContext(ctx)
	defer cancel()

	stmt := `SELECT count(*) OVER(), id, tenant_id, wallet_id, COALESCE(vault_id::text, ''), destination, asset, network, amount_minor, status, required_approvals, approved_count, policy_reason, risk_level, provider_tx_hash, requested_by, created_at, updated_at
		FROM withdrawals
		WHERE tenant_id = $1`
	args := []interface{}{tenantID}
	argPosition := 2

	if strings.TrimSpace(status) != "" {
		stmt += fmt.Sprintf(" AND status = $%d", argPosition)
		args = append(args, status)
		argPosition++
	}

	stmt += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argPosition, argPosition+1)
	args = append(args, filters.PageSize, (filters.Page-1)*filters.PageSize)

	rows, err := m.DB.QueryContext(ctx, stmt, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	items := make([]Withdrawal, 0)
	total := 0
	for rows.Next() {
		var wd Withdrawal
		if err := rows.Scan(
			&total,
			&wd.ID,
			&wd.TenantID,
			&wd.WalletID,
			&wd.VaultID,
			&wd.Destination,
			&wd.Asset,
			&wd.Network,
			&wd.AmountMinor,
			&wd.Status,
			&wd.RequiredApprovals,
			&wd.ApprovedCount,
			&wd.PolicyReason,
			&wd.RiskLevel,
			&wd.ProviderTxHash,
			&wd.RequestedBy,
			&wd.CreatedAt,
			&wd.UpdatedAt,
		); err != nil {
			return nil, Metadata{}, err
		}
		items = append(items, wd)
	}
	if err := rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	return items, CalculateMetadata(total, filters.Page, filters.PageSize), nil
}

func (m WalletModel) ListWithdrawalsByWallet(ctx context.Context, tenantID, walletID string, filters Filters) ([]Withdrawal, Metadata, error) {
	ctx, cancel := NewTimeoutContext(ctx)
	defer cancel()

	stmt := `SELECT count(*) OVER(), id, tenant_id, wallet_id, COALESCE(vault_id::text, ''), destination, asset, network, amount_minor, status, required_approvals, approved_count, policy_reason, risk_level, provider_tx_hash, requested_by, created_at, updated_at
		FROM withdrawals
		WHERE tenant_id = $1 AND wallet_id = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`

	rows, err := m.DB.QueryContext(ctx, stmt, tenantID, walletID, filters.PageSize, (filters.Page-1)*filters.PageSize)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	items := make([]Withdrawal, 0)
	total := 0
	for rows.Next() {
		var wd Withdrawal
		if err := rows.Scan(
			&total,
			&wd.ID,
			&wd.TenantID,
			&wd.WalletID,
			&wd.VaultID,
			&wd.Destination,
			&wd.Asset,
			&wd.Network,
			&wd.AmountMinor,
			&wd.Status,
			&wd.RequiredApprovals,
			&wd.ApprovedCount,
			&wd.PolicyReason,
			&wd.RiskLevel,
			&wd.ProviderTxHash,
			&wd.RequestedBy,
			&wd.CreatedAt,
			&wd.UpdatedAt,
		); err != nil {
			return nil, Metadata{}, err
		}
		items = append(items, wd)
	}
	if err := rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	return items, CalculateMetadata(total, filters.Page, filters.PageSize), nil
}

func (m WalletModel) AddApproval(ctx context.Context, approval *WithdrawalApproval) error {
	ctx, cancel := NewTimeoutContext(ctx)
	defer cancel()

	stmt := `INSERT INTO withdrawal_approvals (withdrawal_id, approved_by, decision, reason)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at`

	return m.DB.QueryRowContext(ctx, stmt,
		approval.WithdrawalID,
		approval.ApprovedBy,
		approval.Decision,
		approval.Reason,
	).Scan(&approval.ID, &approval.CreatedAt)
}

func (m WalletModel) CountApprovals(ctx context.Context, withdrawalID string) (int, error) {
	ctx, cancel := NewTimeoutContext(ctx)
	defer cancel()

	stmt := `SELECT count(*) FROM withdrawal_approvals WHERE withdrawal_id = $1 AND decision = 'approved'`
	var count int
	if err := m.DB.QueryRowContext(ctx, stmt, withdrawalID).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (m WalletModel) UpdateWithdrawalStatus(ctx context.Context, tenantID, withdrawalID, status, txHash, reason string, approvedCount int) error {
	ctx, cancel := NewTimeoutContext(ctx)
	defer cancel()

	stmt := `UPDATE withdrawals
		SET status = $1,
			provider_tx_hash = CASE WHEN $2 = '' THEN provider_tx_hash ELSE $2 END,
			policy_reason = CASE WHEN $3 = '' THEN policy_reason ELSE $3 END,
			approved_count = $4,
			updated_at = NOW()
		WHERE tenant_id = $5 AND id = $6`

	result, err := m.DB.ExecContext(ctx, stmt, status, txHash, reason, approvedCount, tenantID, withdrawalID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrRecordNotFound
	}
	return nil
}

func (m WalletModel) GetPolicyRule(ctx context.Context, tenantID string) (PolicyRule, error) {
	ctx, cancel := NewTimeoutContext(ctx)
	defer cancel()

	stmt := `SELECT tenant_id, whitelist_destinations, per_tx_limit_minor, daily_limit_minor, velocity_count, velocity_window_minutes, approval_tier, updated_at
		FROM policy_rules
		WHERE tenant_id = $1`

	var (
		rule          PolicyRule
		whitelistJSON []byte
	)

	err := m.DB.QueryRowContext(ctx, stmt, tenantID).Scan(
		&rule.TenantID,
		&whitelistJSON,
		&rule.PerTxLimitMinor,
		&rule.DailyLimitMinor,
		&rule.VelocityCount,
		&rule.VelocityWindowMinutes,
		&rule.ApprovalTier,
		&rule.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return PolicyRule{}, ErrRecordNotFound
		}
		return PolicyRule{}, err
	}

	if len(whitelistJSON) > 0 {
		if err := json.Unmarshal(whitelistJSON, &rule.WhitelistDestinations); err != nil {
			return PolicyRule{}, err
		}
	}

	return rule, nil
}

func (m WalletModel) UpsertPolicyRule(ctx context.Context, rule PolicyRule) error {
	ctx, cancel := NewTimeoutContext(ctx)
	defer cancel()

	whitelistJSON, err := json.Marshal(rule.WhitelistDestinations)
	if err != nil {
		return err
	}

	stmt := `INSERT INTO policy_rules (tenant_id, whitelist_destinations, per_tx_limit_minor, daily_limit_minor, velocity_count, velocity_window_minutes, approval_tier)
		VALUES ($1, $2::jsonb, $3, $4, $5, $6, $7)
		ON CONFLICT (tenant_id)
		DO UPDATE SET whitelist_destinations = EXCLUDED.whitelist_destinations,
			per_tx_limit_minor = EXCLUDED.per_tx_limit_minor,
			daily_limit_minor = EXCLUDED.daily_limit_minor,
			velocity_count = EXCLUDED.velocity_count,
			velocity_window_minutes = EXCLUDED.velocity_window_minutes,
			approval_tier = EXCLUDED.approval_tier,
			updated_at = NOW()`

	_, err = m.DB.ExecContext(ctx, stmt,
		rule.TenantID,
		string(whitelistJSON),
		rule.PerTxLimitMinor,
		rule.DailyLimitMinor,
		rule.VelocityCount,
		rule.VelocityWindowMinutes,
		rule.ApprovalTier,
	)
	return err
}
