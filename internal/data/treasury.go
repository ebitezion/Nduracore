package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/lib/pq"
)

type TreasuryModel struct {
	DB *sql.DB
}

type Treasury struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Vault struct {
	ID          string    `json:"id"`
	TreasuryID  string    `json:"treasury_id"`
	TenantID    string    `json:"tenant_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type VaultAsset struct {
	ID        string            `json:"id"`
	VaultID   string            `json:"vault_id"`
	TenantID  string            `json:"tenant_id"`
	AssetCode string            `json:"asset_code"`
	Network   string            `json:"network"`
	Metadata  map[string]string `json:"metadata"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

func (m TreasuryModel) CreateTreasury(ctx context.Context, treasury *Treasury) error {
	ctx, cancel := NewTimeoutContext(ctx)
	defer cancel()

	stmt := `INSERT INTO treasuries (tenant_id, name, description, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`

	err := m.DB.QueryRowContext(ctx, stmt,
		treasury.TenantID,
		treasury.Name,
		treasury.Description,
		treasury.Status,
	).Scan(&treasury.ID, &treasury.CreatedAt, &treasury.UpdatedAt)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return ErrDuplicateRecord
		}
		return err
	}
	return nil
}

func (m TreasuryModel) GetTreasury(ctx context.Context, tenantID, treasuryID string) (*Treasury, error) {
	ctx, cancel := NewTimeoutContext(ctx)
	defer cancel()

	stmt := `SELECT id, tenant_id, name, description, status, created_at, updated_at
		FROM treasuries
		WHERE tenant_id = $1 AND id = $2`

	var treasury Treasury
	err := m.DB.QueryRowContext(ctx, stmt, strings.TrimSpace(tenantID), strings.TrimSpace(treasuryID)).Scan(
		&treasury.ID,
		&treasury.TenantID,
		&treasury.Name,
		&treasury.Description,
		&treasury.Status,
		&treasury.CreatedAt,
		&treasury.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &treasury, nil
}

func (m TreasuryModel) ListTreasuries(ctx context.Context, tenantID string, filters Filters) ([]Treasury, Metadata, error) {
	ctx, cancel := NewTimeoutContext(ctx)
	defer cancel()

	stmt := `SELECT count(*) OVER(), id, tenant_id, name, description, status, created_at, updated_at
		FROM treasuries
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := m.DB.QueryContext(ctx, stmt, strings.TrimSpace(tenantID), filters.PageSize, (filters.Page-1)*filters.PageSize)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	items := make([]Treasury, 0)
	total := 0
	for rows.Next() {
		var item Treasury
		if err := rows.Scan(
			&total,
			&item.ID,
			&item.TenantID,
			&item.Name,
			&item.Description,
			&item.Status,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, Metadata{}, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	return items, CalculateMetadata(total, filters.Page, filters.PageSize), nil
}

func (m TreasuryModel) CreateVault(ctx context.Context, vault *Vault) error {
	ctx, cancel := NewTimeoutContext(ctx)
	defer cancel()

	stmt := `INSERT INTO vaults (treasury_id, tenant_id, name, description, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`

	err := m.DB.QueryRowContext(ctx, stmt,
		vault.TreasuryID,
		vault.TenantID,
		vault.Name,
		vault.Description,
		vault.Status,
	).Scan(&vault.ID, &vault.CreatedAt, &vault.UpdatedAt)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case "23505":
				return ErrDuplicateRecord
			case "23503":
				return ErrRecordNotFound
			}
		}
		return err
	}
	return nil
}

func (m TreasuryModel) GetVault(ctx context.Context, tenantID, vaultID string) (*Vault, error) {
	ctx, cancel := NewTimeoutContext(ctx)
	defer cancel()

	stmt := `SELECT id, treasury_id, tenant_id, name, description, status, created_at, updated_at
		FROM vaults
		WHERE tenant_id = $1 AND id = $2`

	var vault Vault
	err := m.DB.QueryRowContext(ctx, stmt, strings.TrimSpace(tenantID), strings.TrimSpace(vaultID)).Scan(
		&vault.ID,
		&vault.TreasuryID,
		&vault.TenantID,
		&vault.Name,
		&vault.Description,
		&vault.Status,
		&vault.CreatedAt,
		&vault.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &vault, nil
}

func (m TreasuryModel) ListVaultsByTreasury(ctx context.Context, tenantID, treasuryID string, filters Filters) ([]Vault, Metadata, error) {
	ctx, cancel := NewTimeoutContext(ctx)
	defer cancel()

	stmt := `SELECT count(*) OVER(), id, treasury_id, tenant_id, name, description, status, created_at, updated_at
		FROM vaults
		WHERE tenant_id = $1 AND treasury_id = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`

	rows, err := m.DB.QueryContext(ctx, stmt, strings.TrimSpace(tenantID), strings.TrimSpace(treasuryID), filters.PageSize, (filters.Page-1)*filters.PageSize)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	items := make([]Vault, 0)
	total := 0
	for rows.Next() {
		var item Vault
		if err := rows.Scan(
			&total,
			&item.ID,
			&item.TreasuryID,
			&item.TenantID,
			&item.Name,
			&item.Description,
			&item.Status,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, Metadata{}, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	return items, CalculateMetadata(total, filters.Page, filters.PageSize), nil
}

func (m TreasuryModel) CreateVaultAsset(ctx context.Context, asset *VaultAsset) error {
	ctx, cancel := NewTimeoutContext(ctx)
	defer cancel()

	metadataJSON, err := json.Marshal(asset.Metadata)
	if err != nil {
		return err
	}

	stmt := `INSERT INTO vault_assets (vault_id, tenant_id, asset_code, network, metadata)
		VALUES ($1, $2, $3, $4, $5::jsonb)
		RETURNING id, created_at, updated_at`

	err = m.DB.QueryRowContext(ctx, stmt,
		asset.VaultID,
		asset.TenantID,
		asset.AssetCode,
		asset.Network,
		string(metadataJSON),
	).Scan(&asset.ID, &asset.CreatedAt, &asset.UpdatedAt)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			switch pqErr.Code {
			case "23505":
				return ErrDuplicateRecord
			case "23503":
				return ErrRecordNotFound
			}
		}
		return err
	}
	return nil
}

func (m TreasuryModel) ListVaultAssets(ctx context.Context, tenantID, vaultID string, filters Filters) ([]VaultAsset, Metadata, error) {
	ctx, cancel := NewTimeoutContext(ctx)
	defer cancel()

	stmt := `SELECT count(*) OVER(), id, vault_id, tenant_id, asset_code, network, metadata, created_at, updated_at
		FROM vault_assets
		WHERE tenant_id = $1 AND vault_id = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`

	rows, err := m.DB.QueryContext(ctx, stmt, strings.TrimSpace(tenantID), strings.TrimSpace(vaultID), filters.PageSize, (filters.Page-1)*filters.PageSize)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	items := make([]VaultAsset, 0)
	total := 0
	for rows.Next() {
		var (
			item         VaultAsset
			metadataJSON []byte
		)
		if err := rows.Scan(
			&total,
			&item.ID,
			&item.VaultID,
			&item.TenantID,
			&item.AssetCode,
			&item.Network,
			&metadataJSON,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, Metadata{}, err
		}
		if len(metadataJSON) > 0 {
			_ = json.Unmarshal(metadataJSON, &item.Metadata)
		}
		if item.Metadata == nil {
			item.Metadata = map[string]string{}
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	return items, CalculateMetadata(total, filters.Page, filters.PageSize), nil
}
