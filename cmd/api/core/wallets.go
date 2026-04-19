package core

import (
	"errors"
	"net/http"
	"strings"

	"github.com/ebitezion/Nduracore/internal/data"
	"github.com/ebitezion/Nduracore/internal/validator"
	"github.com/ebitezion/Nduracore/internal/wallet"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

func (app *application) createWallet(w http.ResponseWriter, r *http.Request) {
	tenantID := app.tenantIDFromRequest(r)
	if tenantID == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "X-Tenant-ID header is required")
		return
	}

	var input struct {
		VaultID string `json:"vault_id"`
		Asset   string `json:"asset"`
		Network string `json:"network"`
	}

	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	createdWallet, err := app.walletService.CreateWallet(r.Context(), wallet.CreateWalletInput{
		TenantID: tenantID,
		VaultID:  input.VaultID,
		Asset:    input.Asset,
		Network:  input.Network,
	})
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	app.logAuditEvent(r, "wallet.create", "success", map[string]interface{}{
		"tenant_id": tenantID,
		"wallet_id": createdWallet.ID,
		"vault_id":  createdWallet.VaultID,
		"asset":     createdWallet.Asset,
		"network":   createdWallet.Network,
	})

	_ = app.writeJSON(w, http.StatusCreated, envelope{"wallet": createdWallet}, nil)
}

func (app *application) getWallet(w http.ResponseWriter, r *http.Request) {
	tenantID := app.tenantIDFromRequest(r)
	if tenantID == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "X-Tenant-ID header is required")
		return
	}

	walletID := app.pathParam(r, "id")
	if walletID == "" {
		app.notFoundErrorResponse(w, r)
		return
	}
	if !isValidUUID(walletID) {
		app.errorResponse(w, r, http.StatusBadRequest, "wallet id must be a valid UUID")
		return
	}

	result, err := app.walletService.GetWallet(r.Context(), tenantID, walletID)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			app.notFoundErrorResponse(w, r)
			return
		}
		app.serverErrorResponse(w, r)
		return
	}

	_ = app.writeJSON(w, http.StatusOK, envelope{"wallet": result}, nil)
}

func (app *application) getWalletBalance(w http.ResponseWriter, r *http.Request) {
	tenantID := app.tenantIDFromRequest(r)
	if tenantID == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "X-Tenant-ID header is required")
		return
	}

	walletID := app.pathParam(r, "id")
	if walletID == "" {
		app.notFoundErrorResponse(w, r)
		return
	}
	if !isValidUUID(walletID) {
		app.errorResponse(w, r, http.StatusBadRequest, "wallet id must be a valid UUID")
		return
	}

	asset := strings.TrimSpace(app.readString(r.URL.Query(), "asset", ""))
	result, err := app.walletService.GetWalletBalance(r.Context(), wallet.GetWalletBalanceInput{
		TenantID: tenantID,
		WalletID: walletID,
		Asset:    asset,
	})
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			app.notFoundErrorResponse(w, r)
			return
		}
		app.badRequestResponse(w, r, err)
		return
	}

	app.logAuditEvent(r, "wallet.balance.read", "success", map[string]interface{}{
		"tenant_id":     tenantID,
		"wallet_id":     result.WalletID,
		"vault_id":      result.VaultID,
		"asset":         result.Asset,
		"network":       result.Network,
		"balance_minor": result.BalanceMinor,
	})

	_ = app.writeJSON(w, http.StatusOK, envelope{"balance": result}, nil)
}

func (app *application) listWallets(w http.ResponseWriter, r *http.Request) {
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

	vaultID := strings.TrimSpace(app.readString(qs, "vault_id", ""))
	if vaultID != "" && !isValidUUID(vaultID) {
		app.errorResponse(w, r, http.StatusBadRequest, "vault_id must be a valid UUID")
		return
	}
	asset := strings.TrimSpace(app.readString(qs, "asset", ""))
	network := strings.TrimSpace(app.readString(qs, "network", ""))

	wallets, metadata, err := app.walletService.ListWallets(r.Context(), wallet.ListWalletsInput{
		TenantID: tenantID,
		VaultID:  vaultID,
		Asset:    asset,
		Network:  network,
		Filters:  filters,
	})
	if err != nil {
		app.serverErrorResponse(w, r)
		return
	}

	_ = app.writeJSON(w, http.StatusOK, envelope{"wallets": wallets, "metadata": metadata}, nil)
}

func (app *application) listWalletDeposits(w http.ResponseWriter, r *http.Request) {
	tenantID := app.tenantIDFromRequest(r)
	if tenantID == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "X-Tenant-ID header is required")
		return
	}

	walletID := app.pathParam(r, "id")
	if walletID == "" {
		app.notFoundErrorResponse(w, r)
		return
	}
	if !isValidUUID(walletID) {
		app.errorResponse(w, r, http.StatusBadRequest, "wallet id must be a valid UUID")
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

	deposits, metadata, err := app.walletService.ListDeposits(r.Context(), tenantID, walletID, filters)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			app.notFoundErrorResponse(w, r)
			return
		}
		app.serverErrorResponse(w, r)
		return
	}

	_ = app.writeJSON(w, http.StatusOK, envelope{"deposits": deposits, "metadata": metadata}, nil)
}

func (app *application) createWithdrawal(w http.ResponseWriter, r *http.Request) {
	tenantID := app.tenantIDFromRequest(r)
	if tenantID == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "X-Tenant-ID header is required")
		return
	}

	var input struct {
		WalletID    string `json:"wallet_id"`
		VaultID     string `json:"vault_id"`
		Asset       string `json:"asset"`
		Network     string `json:"network"`
		Destination string `json:"destination"`
		AmountMinor int64  `json:"amount_minor"`
	}

	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	requestedBy := app.userIDFromContext(r.Context())
	withdrawalResult, err := app.walletService.RequestWithdrawal(r.Context(), wallet.RequestWithdrawalInput{
		TenantID:    tenantID,
		WalletID:    input.WalletID,
		VaultID:     input.VaultID,
		Asset:       input.Asset,
		Network:     input.Network,
		Destination: input.Destination,
		AmountMinor: input.AmountMinor,
		RequestedBy: requestedBy,
	})
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	app.logAuditEvent(r, "withdrawal.request", withdrawalResult.Status, map[string]interface{}{
		"tenant_id":          tenantID,
		"withdrawal_id":      withdrawalResult.ID,
		"wallet_id":          withdrawalResult.WalletID,
		"vault_id":           withdrawalResult.VaultID,
		"amount_minor":       withdrawalResult.AmountMinor,
		"required_approvals": withdrawalResult.RequiredApprovals,
		"risk_level":         withdrawalResult.RiskLevel,
		"policy_reason":      withdrawalResult.PolicyReason,
	})

	_ = app.writeJSON(w, http.StatusAccepted, envelope{"withdrawal": withdrawalResult}, nil)
}

func (app *application) listWithdrawals(w http.ResponseWriter, r *http.Request) {
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

	status := strings.ToLower(strings.TrimSpace(app.readString(qs, "status", "")))
	if status != "" {
		v.Check(validator.In(status, "requested", "policy_pending", "approved", "broadcasted", "confirmed", "rejected", "failed", "cancelled"), "status", "must be a valid withdrawal status")
	}
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	withdrawals, metadata, err := app.walletService.ListWithdrawals(r.Context(), wallet.ListWithdrawalsInput{
		TenantID: tenantID,
		Status:   status,
		Filters:  filters,
	})
	if err != nil {
		app.serverErrorResponse(w, r)
		return
	}

	_ = app.writeJSON(w, http.StatusOK, envelope{"withdrawals": withdrawals, "metadata": metadata}, nil)
}

func (app *application) getWithdrawal(w http.ResponseWriter, r *http.Request) {
	tenantID := app.tenantIDFromRequest(r)
	if tenantID == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "X-Tenant-ID header is required")
		return
	}

	withdrawalID := app.pathParam(r, "id")
	if withdrawalID == "" {
		app.notFoundErrorResponse(w, r)
		return
	}
	if !isValidUUID(withdrawalID) {
		app.errorResponse(w, r, http.StatusBadRequest, "withdrawal id must be a valid UUID")
		return
	}

	result, err := app.walletService.GetWithdrawal(r.Context(), tenantID, withdrawalID)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			app.notFoundErrorResponse(w, r)
			return
		}
		app.serverErrorResponse(w, r)
		return
	}

	_ = app.writeJSON(w, http.StatusOK, envelope{"withdrawal": result}, nil)
}

func (app *application) approveWithdrawal(w http.ResponseWriter, r *http.Request) {
	tenantID := app.tenantIDFromRequest(r)
	if tenantID == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "X-Tenant-ID header is required")
		return
	}

	withdrawalID := app.pathParam(r, "id")
	if withdrawalID == "" {
		app.notFoundErrorResponse(w, r)
		return
	}
	if !isValidUUID(withdrawalID) {
		app.errorResponse(w, r, http.StatusBadRequest, "withdrawal id must be a valid UUID")
		return
	}

	var input struct {
		Reason string `json:"reason"`
	}
	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	result, err := app.walletService.ApproveWithdrawal(r.Context(), wallet.ApproveWithdrawalInput{
		TenantID:     tenantID,
		WithdrawalID: withdrawalID,
		ApprovedBy:   app.userIDFromContext(r.Context()),
		Reason:       input.Reason,
	})
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			app.notFoundErrorResponse(w, r)
			return
		}
		app.badRequestResponse(w, r, err)
		return
	}

	app.logAuditEvent(r, "withdrawal.approve", result.Status, map[string]interface{}{
		"tenant_id":          tenantID,
		"withdrawal_id":      result.ID,
		"approved_count":     result.ApprovedCount,
		"required_approvals": result.RequiredApprovals,
	})

	_ = app.writeJSON(w, http.StatusOK, envelope{"withdrawal": result}, nil)
}

func (app *application) rejectWithdrawal(w http.ResponseWriter, r *http.Request) {
	tenantID := app.tenantIDFromRequest(r)
	if tenantID == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "X-Tenant-ID header is required")
		return
	}

	withdrawalID := app.pathParam(r, "id")
	if withdrawalID == "" {
		app.notFoundErrorResponse(w, r)
		return
	}
	if !isValidUUID(withdrawalID) {
		app.errorResponse(w, r, http.StatusBadRequest, "withdrawal id must be a valid UUID")
		return
	}

	var input struct {
		Reason string `json:"reason"`
	}
	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	result, err := app.walletService.RejectWithdrawal(r.Context(), wallet.RejectWithdrawalInput{
		TenantID:     tenantID,
		WithdrawalID: withdrawalID,
		RejectedBy:   app.userIDFromContext(r.Context()),
		Reason:       input.Reason,
	})
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			app.notFoundErrorResponse(w, r)
			return
		}
		app.badRequestResponse(w, r, err)
		return
	}

	app.logAuditEvent(r, "withdrawal.reject", result.Status, map[string]interface{}{
		"tenant_id":     tenantID,
		"withdrawal_id": result.ID,
		"policy_reason": result.PolicyReason,
	})

	_ = app.writeJSON(w, http.StatusOK, envelope{"withdrawal": result}, nil)
}

func (app *application) tenantIDFromRequest(r *http.Request) string {
	return strings.TrimSpace(r.Header.Get("X-Tenant-ID"))
}

func (app *application) pathParam(r *http.Request, key string) string {
	params := httprouter.ParamsFromContext(r.Context())
	return strings.TrimSpace(params.ByName(key))
}

func isValidUUID(value string) bool {
	_, err := uuid.Parse(strings.TrimSpace(value))
	return err == nil
}

func (app *application) ingestAlchemyDepositWebhook(w http.ResponseWriter, r *http.Request) {
	walletID := app.pathParam(r, "id")
	if !isValidUUID(walletID) {
		app.errorResponse(w, r, http.StatusBadRequest, "wallet id must be a valid UUID")
		return
	}
	app.processProviderWebhook(w, r, "alchemy", walletID)
}
