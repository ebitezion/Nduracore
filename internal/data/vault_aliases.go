package data

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/lib/pq"
)

type VaultAlias struct {
	ID              string    `json:"id"`
	TenantID        string    `json:"tenant_id"`
	VaultID         string    `json:"vault_id"`
	Provider        string    `json:"provider"`
	ExternalVaultID string    `json:"external_vault_id"`
	ExternalName    string    `json:"external_name"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (m TreasuryModel) CreateVaultAlias(ctx context.Context, alias *VaultAlias) error {
	ctx, cancel := NewTimeoutContext(ctx)
	defer cancel()

	stmt := `INSERT INTO vault_aliases (tenant_id, vault_id, provider, external_vault_id, external_name)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`

	err := m.DB.QueryRowContext(ctx, stmt,
		strings.TrimSpace(alias.TenantID),
		strings.TrimSpace(alias.VaultID),
		strings.ToLower(strings.TrimSpace(alias.Provider)),
		strings.TrimSpace(alias.ExternalVaultID),
		strings.TrimSpace(alias.ExternalName),
	).Scan(&alias.ID, &alias.CreatedAt, &alias.UpdatedAt)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return ErrDuplicateRecord
		}
		if errors.As(err, &pqErr) && pqErr.Code == "23503" {
			return ErrRecordNotFound
		}
		return err
	}
	return nil
}

func (m TreasuryModel) ListVaultAliases(ctx context.Context, tenantID, vaultID string, filters Filters) ([]VaultAlias, Metadata, error) {
	ctx, cancel := NewTimeoutContext(ctx)
	defer cancel()

	stmt := `SELECT count(*) OVER(), id, tenant_id, vault_id::text, provider, external_vault_id, external_name, created_at, updated_at
		FROM vault_aliases
		WHERE tenant_id = $1 AND vault_id = $2::uuid
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4`

	rows, err := m.DB.QueryContext(ctx, stmt,
		strings.TrimSpace(tenantID),
		strings.TrimSpace(vaultID),
		filters.PageSize,
		(filters.Page-1)*filters.PageSize,
	)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	items := make([]VaultAlias, 0)
	total := 0
	for rows.Next() {
		var item VaultAlias
		if err := rows.Scan(&total, &item.ID, &item.TenantID, &item.VaultID, &item.Provider, &item.ExternalVaultID, &item.ExternalName, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, Metadata{}, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	return items, CalculateMetadata(total, filters.Page, filters.PageSize), nil
}

func (m TreasuryModel) ResolveVaultByAlias(ctx context.Context, tenantID, provider, externalVaultID string) (*VaultAlias, error) {
	ctx, cancel := NewTimeoutContext(ctx)
	defer cancel()

	stmt := `SELECT id, tenant_id, vault_id::text, provider, external_vault_id, external_name, created_at, updated_at
		FROM vault_aliases
		WHERE tenant_id = $1 AND provider = $2 AND external_vault_id = $3`

	var alias VaultAlias
	err := m.DB.QueryRowContext(ctx, stmt,
		strings.TrimSpace(tenantID),
		strings.ToLower(strings.TrimSpace(provider)),
		strings.TrimSpace(externalVaultID),
	).Scan(&alias.ID, &alias.TenantID, &alias.VaultID, &alias.Provider, &alias.ExternalVaultID, &alias.ExternalName, &alias.CreatedAt, &alias.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &alias, nil
}
