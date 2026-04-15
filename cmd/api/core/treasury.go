package core

import (
	"errors"
	"net/http"
	"strings"

	"github.com/ebitezion/Nduracore/internal/data"
	"github.com/ebitezion/Nduracore/internal/validator"
)

func (app *application) createTreasury(w http.ResponseWriter, r *http.Request) {
	tenantID := app.tenantIDFromRequest(r)
	if tenantID == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "X-Tenant-ID header is required")
		return
	}

	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Status      string `json:"status"`
	}
	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if strings.TrimSpace(input.Status) == "" {
		input.Status = "active"
	}

	v := validator.New()
	v.Check(strings.TrimSpace(input.Name) != "", "name", "must be provided")
	v.Check(len(strings.TrimSpace(input.Name)) <= 120, "name", "must not exceed 120 chars")
	v.Check(len(strings.TrimSpace(input.Description)) <= 500, "description", "must not exceed 500 chars")
	v.Check(validator.In(strings.TrimSpace(input.Status), "active", "archived"), "status", "must be active or archived")
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	treasury := data.Treasury{
		TenantID:    tenantID,
		Name:        strings.TrimSpace(input.Name),
		Description: strings.TrimSpace(input.Description),
		Status:      strings.TrimSpace(input.Status),
	}
	if err := app.model.Treasury.CreateTreasury(r.Context(), &treasury); err != nil {
		if errors.Is(err, data.ErrDuplicateRecord) {
			app.errorResponse(w, r, http.StatusConflict, "treasury with same name already exists")
			return
		}
		app.serverErrorResponse(w, r)
		return
	}

	app.logAuditEvent(r, "treasury.create", "success", map[string]interface{}{"tenant_id": tenantID, "treasury_id": treasury.ID})
	_ = app.writeJSON(w, http.StatusCreated, envelope{"treasury": treasury}, nil)
}

func (app *application) listTreasuries(w http.ResponseWriter, r *http.Request) {
	tenantID := app.tenantIDFromRequest(r)
	if tenantID == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "X-Tenant-ID header is required")
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

	items, metadata, err := app.model.Treasury.ListTreasuries(r.Context(), tenantID, filters)
	if err != nil {
		app.serverErrorResponse(w, r)
		return
	}

	_ = app.writeJSON(w, http.StatusOK, envelope{"treasuries": items, "metadata": metadata}, nil)
}

func (app *application) getTreasury(w http.ResponseWriter, r *http.Request) {
	tenantID := app.tenantIDFromRequest(r)
	if tenantID == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "X-Tenant-ID header is required")
		return
	}

	treasuryID := app.pathParam(r, "id")
	if treasuryID == "" {
		app.notFoundErrorResponse(w, r)
		return
	}

	item, err := app.model.Treasury.GetTreasury(r.Context(), tenantID, treasuryID)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			app.notFoundErrorResponse(w, r)
			return
		}
		app.serverErrorResponse(w, r)
		return
	}

	_ = app.writeJSON(w, http.StatusOK, envelope{"treasury": item}, nil)
}

func (app *application) createVault(w http.ResponseWriter, r *http.Request) {
	tenantID := app.tenantIDFromRequest(r)
	if tenantID == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "X-Tenant-ID header is required")
		return
	}

	treasuryID := app.pathParam(r, "id")
	if treasuryID == "" {
		app.notFoundErrorResponse(w, r)
		return
	}

	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Status      string `json:"status"`
	}
	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if strings.TrimSpace(input.Status) == "" {
		input.Status = "active"
	}

	v := validator.New()
	v.Check(strings.TrimSpace(input.Name) != "", "name", "must be provided")
	v.Check(len(strings.TrimSpace(input.Name)) <= 120, "name", "must not exceed 120 chars")
	v.Check(len(strings.TrimSpace(input.Description)) <= 500, "description", "must not exceed 500 chars")
	v.Check(validator.In(strings.TrimSpace(input.Status), "active", "archived"), "status", "must be active or archived")
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	item := data.Vault{
		TreasuryID:  treasuryID,
		TenantID:    tenantID,
		Name:        strings.TrimSpace(input.Name),
		Description: strings.TrimSpace(input.Description),
		Status:      strings.TrimSpace(input.Status),
	}
	if err := app.model.Treasury.CreateVault(r.Context(), &item); err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateRecord):
			app.errorResponse(w, r, http.StatusConflict, "vault with same name already exists in treasury")
			return
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundErrorResponse(w, r)
			return
		default:
			app.serverErrorResponse(w, r)
			return
		}
	}

	app.logAuditEvent(r, "vault.create", "success", map[string]interface{}{"tenant_id": tenantID, "treasury_id": treasuryID, "vault_id": item.ID})
	_ = app.writeJSON(w, http.StatusCreated, envelope{"vault": item}, nil)
}

func (app *application) listVaults(w http.ResponseWriter, r *http.Request) {
	tenantID := app.tenantIDFromRequest(r)
	if tenantID == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "X-Tenant-ID header is required")
		return
	}

	treasuryID := app.pathParam(r, "id")
	if treasuryID == "" {
		app.notFoundErrorResponse(w, r)
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

	items, metadata, err := app.model.Treasury.ListVaultsByTreasury(r.Context(), tenantID, treasuryID, filters)
	if err != nil {
		app.serverErrorResponse(w, r)
		return
	}

	_ = app.writeJSON(w, http.StatusOK, envelope{"vaults": items, "metadata": metadata}, nil)
}

func (app *application) getVault(w http.ResponseWriter, r *http.Request) {
	tenantID := app.tenantIDFromRequest(r)
	if tenantID == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "X-Tenant-ID header is required")
		return
	}

	vaultID := app.pathParam(r, "id")
	if vaultID == "" {
		app.notFoundErrorResponse(w, r)
		return
	}

	item, err := app.model.Treasury.GetVault(r.Context(), tenantID, vaultID)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			app.notFoundErrorResponse(w, r)
			return
		}
		app.serverErrorResponse(w, r)
		return
	}

	_ = app.writeJSON(w, http.StatusOK, envelope{"vault": item}, nil)
}

func (app *application) createVaultAsset(w http.ResponseWriter, r *http.Request) {
	tenantID := app.tenantIDFromRequest(r)
	if tenantID == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "X-Tenant-ID header is required")
		return
	}

	vaultID := app.pathParam(r, "id")
	if vaultID == "" {
		app.notFoundErrorResponse(w, r)
		return
	}

	var input struct {
		AssetCode string            `json:"asset_code"`
		Network   string            `json:"network"`
		Metadata  map[string]string `json:"metadata"`
	}
	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	v := validator.New()
	v.Check(strings.TrimSpace(input.AssetCode) != "", "asset_code", "must be provided")
	v.Check(len(strings.TrimSpace(input.AssetCode)) <= 40, "asset_code", "must not exceed 40 chars")
	v.Check(strings.TrimSpace(input.Network) != "", "network", "must be provided")
	v.Check(len(strings.TrimSpace(input.Network)) <= 60, "network", "must not exceed 60 chars")
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	if input.Metadata == nil {
		input.Metadata = map[string]string{}
	}

	item := data.VaultAsset{
		VaultID:   vaultID,
		TenantID:  tenantID,
		AssetCode: strings.ToUpper(strings.TrimSpace(input.AssetCode)),
		Network:   strings.ToLower(strings.TrimSpace(input.Network)),
		Metadata:  input.Metadata,
	}

	if err := app.model.Treasury.CreateVaultAsset(r.Context(), &item); err != nil {
		switch {
		case errors.Is(err, data.ErrDuplicateRecord):
			app.errorResponse(w, r, http.StatusConflict, "asset already exists for this vault")
			return
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundErrorResponse(w, r)
			return
		default:
			app.serverErrorResponse(w, r)
			return
		}
	}

	app.logAuditEvent(r, "vault.asset.create", "success", map[string]interface{}{"tenant_id": tenantID, "vault_id": vaultID, "asset_id": item.ID, "asset_code": item.AssetCode})
	_ = app.writeJSON(w, http.StatusCreated, envelope{"asset": item}, nil)
}

func (app *application) listVaultAssets(w http.ResponseWriter, r *http.Request) {
	tenantID := app.tenantIDFromRequest(r)
	if tenantID == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "X-Tenant-ID header is required")
		return
	}

	vaultID := app.pathParam(r, "id")
	if vaultID == "" {
		app.notFoundErrorResponse(w, r)
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

	items, metadata, err := app.model.Treasury.ListVaultAssets(r.Context(), tenantID, vaultID, filters)
	if err != nil {
		app.serverErrorResponse(w, r)
		return
	}

	_ = app.writeJSON(w, http.StatusOK, envelope{"assets": items, "metadata": metadata}, nil)
}
