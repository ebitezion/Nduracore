package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type AssetRegistryModel struct {
	DB *sql.DB
}

type AssetDefinition struct {
	ID              string                 `json:"id"`
	Network         string                 `json:"network"`
	AssetCode       string                 `json:"asset_code"`
	ChainFamily     string                 `json:"chain_family"`
	AssetType       string                 `json:"asset_type"`
	ContractAddress string                 `json:"contract_address"`
	Decimals        int                    `json:"decimals"`
	IsActive        bool                   `json:"is_active"`
	Metadata        map[string]interface{} `json:"metadata"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

func (m AssetRegistryModel) GetActiveAsset(ctx context.Context, network, assetCode string) (*AssetDefinition, error) {
	ctx, cancel := NewTimeoutContext(ctx)
	defer cancel()

	stmt := `SELECT id, network, asset_code, chain_family, asset_type, contract_address, decimals, is_active, metadata, created_at, updated_at
		FROM asset_registry
		WHERE network = $1 AND asset_code = $2 AND is_active = true`

	var (
		item         AssetDefinition
		metadataJSON []byte
	)
	err := m.DB.QueryRowContext(
		ctx,
		stmt,
		strings.ToLower(strings.TrimSpace(network)),
		strings.ToUpper(strings.TrimSpace(assetCode)),
	).Scan(
		&item.ID,
		&item.Network,
		&item.AssetCode,
		&item.ChainFamily,
		&item.AssetType,
		&item.ContractAddress,
		&item.Decimals,
		&item.IsActive,
		&metadataJSON,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	if len(metadataJSON) > 0 {
		_ = json.Unmarshal(metadataJSON, &item.Metadata)
	}
	if item.Metadata == nil {
		item.Metadata = map[string]interface{}{}
	}
	return &item, nil
}

func (m AssetRegistryModel) UpsertAsset(ctx context.Context, item *AssetDefinition) error {
	ctx, cancel := NewTimeoutContext(ctx)
	defer cancel()

	metadataJSON, err := json.Marshal(item.Metadata)
	if err != nil {
		return err
	}

	stmt := `INSERT INTO asset_registry (network, asset_code, chain_family, asset_type, contract_address, decimals, is_active, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8::jsonb)
		ON CONFLICT (network, asset_code)
		DO UPDATE SET
			chain_family = EXCLUDED.chain_family,
			asset_type = EXCLUDED.asset_type,
			contract_address = EXCLUDED.contract_address,
			decimals = EXCLUDED.decimals,
			is_active = EXCLUDED.is_active,
			metadata = EXCLUDED.metadata,
			updated_at = NOW()
		RETURNING id, created_at, updated_at`

	return m.DB.QueryRowContext(
		ctx,
		stmt,
		strings.ToLower(strings.TrimSpace(item.Network)),
		strings.ToUpper(strings.TrimSpace(item.AssetCode)),
		strings.ToLower(strings.TrimSpace(item.ChainFamily)),
		strings.ToLower(strings.TrimSpace(item.AssetType)),
		strings.TrimSpace(item.ContractAddress),
		item.Decimals,
		item.IsActive,
		string(metadataJSON),
	).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
}
