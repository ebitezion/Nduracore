package core

import (
	"errors"
	"net/http"
	"strings"

	"github.com/ebitezion/Nduracore/internal/data"
	"github.com/ebitezion/Nduracore/internal/validator"
)

func (app *application) createVaultAlias(w http.ResponseWriter, r *http.Request) {
	tenantID := app.tenantIDFromRequest(r)
	if tenantID == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "X-Tenant-ID header is required")
		return
	}

	vaultID := app.pathParam(r, "id")
	if !isValidUUID(vaultID) {
		app.errorResponse(w, r, http.StatusBadRequest, "vault id must be a valid UUID")
		return
	}

	var input struct {
		Provider        string `json:"provider"`
		ExternalVaultID string `json:"external_vault_id"`
		ExternalName    string `json:"external_name"`
	}
	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	alias := data.VaultAlias{
		TenantID:        tenantID,
		VaultID:         vaultID,
		Provider:        strings.ToLower(strings.TrimSpace(input.Provider)),
		ExternalVaultID: strings.TrimSpace(input.ExternalVaultID),
		ExternalName:    strings.TrimSpace(input.ExternalName),
	}
	if alias.Provider == "" || alias.ExternalVaultID == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "provider and external_vault_id are required")
		return
	}

	if err := app.model.Treasury.CreateVaultAlias(r.Context(), &alias); err != nil {
		if errors.Is(err, data.ErrDuplicateRecord) {
			app.errorResponse(w, r, http.StatusConflict, "alias already exists")
			return
		}
		if errors.Is(err, data.ErrRecordNotFound) {
			app.notFoundErrorResponse(w, r)
			return
		}
		app.serverErrorResponse(w, r)
		return
	}

	_ = app.writeJSON(w, http.StatusCreated, envelope{"alias": alias}, nil)
}

func (app *application) listVaultAliases(w http.ResponseWriter, r *http.Request) {
	tenantID := app.tenantIDFromRequest(r)
	if tenantID == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "X-Tenant-ID header is required")
		return
	}

	vaultID := app.pathParam(r, "id")
	if !isValidUUID(vaultID) {
		app.errorResponse(w, r, http.StatusBadRequest, "vault id must be a valid UUID")
		return
	}

	qs := r.URL.Query()
	v := validator.New()
	filters := data.Filters{
		Page:         app.readInt(qs, "page", 1, v),
		PageSize:     app.readInt(qs, "page_size", 20, v),
		Sort:         app.readString(qs, "sort", "-created_at"),
		SortSafelist: []string{"created_at", "-created_at"},
	}
	data.ValidateFilters(v, filters)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	items, metadata, err := app.model.Treasury.ListVaultAliases(r.Context(), tenantID, vaultID, filters)
	if err != nil {
		app.serverErrorResponse(w, r)
		return
	}

	_ = app.writeJSON(w, http.StatusOK, envelope{"aliases": items, "metadata": metadata}, nil)
}

func (app *application) resolveVaultAlias(w http.ResponseWriter, r *http.Request) {
	tenantID := app.tenantIDFromRequest(r)
	if tenantID == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "X-Tenant-ID header is required")
		return
	}

	provider := strings.TrimSpace(app.readString(r.URL.Query(), "provider", ""))
	externalVaultID := strings.TrimSpace(app.readString(r.URL.Query(), "external_vault_id", ""))
	if provider == "" || externalVaultID == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "provider and external_vault_id query params are required")
		return
	}

	alias, err := app.model.Treasury.ResolveVaultByAlias(r.Context(), tenantID, provider, externalVaultID)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			app.notFoundErrorResponse(w, r)
			return
		}
		app.serverErrorResponse(w, r)
		return
	}

	_ = app.writeJSON(w, http.StatusOK, envelope{"alias": alias}, nil)
}

func (app *application) upsertFeePolicy(w http.ResponseWriter, r *http.Request) {
	tenantID := app.tenantIDFromRequest(r)
	if tenantID == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "X-Tenant-ID header is required")
		return
	}

	network := strings.TrimSpace(app.pathParam(r, "network"))
	asset := strings.TrimSpace(app.pathParam(r, "asset"))
	if network == "" || asset == "" {
		app.notFoundErrorResponse(w, r)
		return
	}

	var input struct {
		FeeBps         int    `json:"fee_bps"`
		MinFeeMinor    int64  `json:"min_fee_minor"`
		MaxFeeMinor    int64  `json:"max_fee_minor"`
		FeeDestination string `json:"fee_destination"`
	}
	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if input.FeeBps < 0 || input.FeeBps > 10_000 {
		app.errorResponse(w, r, http.StatusBadRequest, "fee_bps must be between 0 and 10000")
		return
	}

	err := app.model.Wallets.UpsertFeePolicy(r.Context(), data.FeePolicy{
		TenantID:       tenantID,
		Asset:          strings.ToUpper(asset),
		Network:        strings.ToLower(network),
		FeeBps:         input.FeeBps,
		MinFeeMinor:    input.MinFeeMinor,
		MaxFeeMinor:    input.MaxFeeMinor,
		FeeDestination: strings.TrimSpace(input.FeeDestination),
	})
	if err != nil {
		app.serverErrorResponse(w, r)
		return
	}

	policy, err := app.model.Wallets.GetFeePolicy(r.Context(), tenantID, asset, network)
	if err != nil {
		app.serverErrorResponse(w, r)
		return
	}

	_ = app.writeJSON(w, http.StatusOK, envelope{"fee_policy": policy}, nil)
}

func (app *application) getFeePolicy(w http.ResponseWriter, r *http.Request) {
	tenantID := app.tenantIDFromRequest(r)
	if tenantID == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "X-Tenant-ID header is required")
		return
	}

	network := strings.TrimSpace(app.pathParam(r, "network"))
	asset := strings.TrimSpace(app.pathParam(r, "asset"))
	if network == "" || asset == "" {
		app.notFoundErrorResponse(w, r)
		return
	}

	policy, err := app.model.Wallets.GetFeePolicy(r.Context(), tenantID, asset, network)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			app.notFoundErrorResponse(w, r)
			return
		}
		app.serverErrorResponse(w, r)
		return
	}
	_ = app.writeJSON(w, http.StatusOK, envelope{"fee_policy": policy}, nil)
}
