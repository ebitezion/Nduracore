package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/ebitezion/Nduracore/internal/data"
	"github.com/ebitezion/Nduracore/internal/validator"
	"github.com/ebitezion/Nduracore/internal/wallet"
	"github.com/julienschmidt/httprouter"
)

func (app *application) createWallet(w http.ResponseWriter, r *http.Request) {
	tenantID := app.tenantIDFromRequest(r)
	if tenantID == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "X-Tenant-ID header is required")
		return
	}

	var input struct {
		Asset   string `json:"asset"`
		Network string `json:"network"`
	}

	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	createdWallet, err := app.walletService.CreateWallet(r.Context(), wallet.CreateWalletInput{
		TenantID: tenantID,
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

	wallets, metadata, err := app.walletService.ListWallets(r.Context(), tenantID, filters)
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
		"amount_minor":       withdrawalResult.AmountMinor,
		"required_approvals": withdrawalResult.RequiredApprovals,
		"risk_level":         withdrawalResult.RiskLevel,
		"policy_reason":      withdrawalResult.PolicyReason,
	})

	_ = app.writeJSON(w, http.StatusAccepted, envelope{"withdrawal": withdrawalResult}, nil)
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

func (app *application) ingestAlchemyDepositWebhook(w http.ResponseWriter, r *http.Request) {
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

	r.Body = http.MaxBytesReader(w, r.Body, 1_048_576)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	signature := strings.TrimSpace(r.Header.Get("X-Alchemy-Signature"))
	if app.alchemy != nil && !app.alchemy.VerifyWebhookSignature(body, signature) {
		app.errorResponse(w, r, http.StatusUnauthorized, "invalid webhook signature")
		return
	}

	parsed, err := decodeAlchemyWebhookDeposits(body)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	recorded := make([]data.WalletDeposit, 0, len(parsed))
	for _, dep := range parsed {
		deposit, err := app.walletService.RecordDeposit(r.Context(), wallet.RecordDepositInput{
			TenantID:      tenantID,
			WalletID:      walletID,
			TxHash:        dep.TxHash,
			AmountMinor:   dep.AmountMinor,
			Confirmations: dep.Confirmations,
			Status:        dep.Status,
		})
		if err != nil {
			if errors.Is(err, data.ErrRecordNotFound) {
				app.notFoundErrorResponse(w, r)
				return
			}
			app.badRequestResponse(w, r, err)
			return
		}
		recorded = append(recorded, deposit)
	}

	app.logAuditEvent(r, "wallet.deposit.webhook", "accepted", map[string]interface{}{
		"tenant_id":       tenantID,
		"wallet_id":       walletID,
		"deposit_count":   len(recorded),
		"primary_tx_hash": recorded[0].TxHash,
	})

	_ = app.writeJSON(w, http.StatusAccepted, envelope{"deposits": recorded}, nil)
}

type parsedWebhookDeposit struct {
	TxHash        string
	AmountMinor   int64
	Confirmations int
	Status        string
}

func decodeAlchemyWebhookDeposits(body []byte) ([]parsedWebhookDeposit, error) {
	var simple struct {
		TxHash        string `json:"tx_hash"`
		AmountMinor   int64  `json:"amount_minor"`
		Confirmations int    `json:"confirmations"`
		Status        string `json:"status"`
	}
	if err := json.Unmarshal(body, &simple); err != nil {
		return nil, err
	}
	if strings.TrimSpace(simple.TxHash) != "" && simple.AmountMinor > 0 {
		return []parsedWebhookDeposit{{
			TxHash:        strings.ToLower(strings.TrimSpace(simple.TxHash)),
			AmountMinor:   simple.AmountMinor,
			Confirmations: maxInt(simple.Confirmations, 0),
			Status:        defaultDepositStatus(simple.Status, simple.Confirmations),
		}}, nil
	}

	var alchemy struct {
		Event struct {
			Activity []struct {
				Hash             string `json:"hash"`
				Value            any    `json:"value"`
				NumConfirmations int    `json:"numConfirmations"`
				Confirmations    int    `json:"confirmations"`
				Status           string `json:"status"`
				RawContract      struct {
					Value string `json:"value"`
				} `json:"rawContract"`
			} `json:"activity"`
		} `json:"event"`
	}
	if err := json.Unmarshal(body, &alchemy); err != nil {
		return nil, err
	}

	deposits := make([]parsedWebhookDeposit, 0, len(alchemy.Event.Activity))
	for _, activity := range alchemy.Event.Activity {
		txHash := strings.ToLower(strings.TrimSpace(activity.Hash))
		if txHash == "" {
			continue
		}

		amountMinor := parseWebhookAmountMinor(activity.Value, activity.RawContract.Value)
		if amountMinor <= 0 {
			continue
		}

		confirmations := activity.NumConfirmations
		if confirmations == 0 {
			confirmations = activity.Confirmations
		}
		if confirmations < 0 {
			confirmations = 0
		}

		deposits = append(deposits, parsedWebhookDeposit{
			TxHash:        txHash,
			AmountMinor:   amountMinor,
			Confirmations: confirmations,
			Status:        defaultDepositStatus(activity.Status, confirmations),
		})
	}

	if len(deposits) == 0 {
		return nil, fmt.Errorf("no deposit activity found in webhook payload")
	}
	return deposits, nil
}

func parseWebhookAmountMinor(value any, rawHexValue string) int64 {
	if strings.TrimSpace(rawHexValue) != "" {
		hexValue := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(rawHexValue)), "0x")
		if hexValue != "" {
			parsed, err := strconv.ParseInt(hexValue, 16, 64)
			if err == nil && parsed > 0 {
				return parsed
			}
		}
	}

	switch v := value.(type) {
	case float64:
		if v <= 0 {
			return 0
		}
		return int64(math.Round(v * 1_000_000))
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
		if err != nil || parsed <= 0 {
			return 0
		}
		return int64(math.Round(parsed * 1_000_000))
	default:
		return 0
	}
}

func defaultDepositStatus(raw string, confirmations int) string {
	status := strings.ToLower(strings.TrimSpace(raw))
	if status != "" {
		return status
	}
	if confirmations > 0 {
		return "confirmed"
	}
	return "pending"
}

func maxInt(value, floor int) int {
	if value < floor {
		return floor
	}
	return value
}
