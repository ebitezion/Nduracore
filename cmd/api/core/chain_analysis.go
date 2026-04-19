package core

import (
	"errors"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/ebitezion/Nduracore/internal/data"
	"github.com/ebitezion/Nduracore/internal/integrations"
	"github.com/ebitezion/Nduracore/internal/validator"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gagliardetto/solana-go"
	"github.com/stellar/go/keypair"
)

type walletTxItem struct {
	Type        string    `json:"type"`
	ReferenceID string    `json:"reference_id"`
	TxHash      string    `json:"tx_hash"`
	Asset       string    `json:"asset"`
	Network     string    `json:"network"`
	AmountMinor int64     `json:"amount_minor"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

func (app *application) getWalletTransactions(w http.ResponseWriter, r *http.Request) {
	tenantID := app.tenantIDFromRequest(r)
	if tenantID == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "X-Tenant-ID header is required")
		return
	}

	walletID := app.pathParam(r, "id")
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

	deposits, _, err := app.walletService.ListDeposits(r.Context(), tenantID, walletID, filters)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			app.notFoundErrorResponse(w, r)
			return
		}
		app.serverErrorResponse(w, r)
		return
	}

	withdrawals, _, err := app.model.Wallets.ListWithdrawalsByWallet(r.Context(), tenantID, walletID, filters)
	if err != nil {
		app.serverErrorResponse(w, r)
		return
	}

	items := make([]walletTxItem, 0, len(deposits)+len(withdrawals))
	for _, dep := range deposits {
		items = append(items, walletTxItem{
			Type:        "deposit",
			ReferenceID: dep.ID,
			TxHash:      dep.TxHash,
			Asset:       dep.Asset,
			Network:     dep.Network,
			AmountMinor: dep.AmountMinor,
			Status:      dep.Status,
			CreatedAt:   dep.DetectedAt,
		})
	}
	for _, wd := range withdrawals {
		items = append(items, walletTxItem{
			Type:        "withdrawal",
			ReferenceID: wd.ID,
			TxHash:      wd.ProviderTxHash,
			Asset:       wd.Asset,
			Network:     wd.Network,
			AmountMinor: wd.AmountMinor,
			Status:      wd.Status,
			CreatedAt:   wd.CreatedAt,
		})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})

	_ = app.writeJSON(w, http.StatusOK, envelope{"transactions": items}, nil)
}

func (app *application) getWalletGasEstimate(w http.ResponseWriter, r *http.Request) {
	tenantID := app.tenantIDFromRequest(r)
	if tenantID == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "X-Tenant-ID header is required")
		return
	}
	if app.alchemy == nil {
		app.serverErrorResponse(w, r)
		return
	}

	walletID := app.pathParam(r, "id")
	if !isValidUUID(walletID) {
		app.errorResponse(w, r, http.StatusBadRequest, "wallet id must be a valid UUID")
		return
	}

	walletRecord, err := app.walletService.GetWallet(r.Context(), tenantID, walletID)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			app.notFoundErrorResponse(w, r)
			return
		}
		app.serverErrorResponse(w, r)
		return
	}

	qs := r.URL.Query()
	to := strings.TrimSpace(app.readString(qs, "destination", ""))
	if to == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "destination query param is required")
		return
	}
	amountMinor := app.readInt(qs, "amount_minor", 0, validator.New())
	if amountMinor <= 0 {
		app.errorResponse(w, r, http.StatusBadRequest, "amount_minor must be greater than 0")
		return
	}
	asset := strings.TrimSpace(app.readString(qs, "asset", walletRecord.Asset))

	estimate, err := app.alchemy.EstimateTransferGas(r.Context(), walletRecord.Network, asset, walletRecord.Address, to, int64(amountMinor))
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	_ = app.writeJSON(w, http.StatusOK, envelope{"gas_estimate": estimate}, nil)
}

func (app *application) simulateWalletWithdrawal(w http.ResponseWriter, r *http.Request) {
	tenantID := app.tenantIDFromRequest(r)
	if tenantID == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "X-Tenant-ID header is required")
		return
	}
	if app.alchemy == nil {
		app.serverErrorResponse(w, r)
		return
	}
	walletID := app.pathParam(r, "id")
	if !isValidUUID(walletID) {
		app.errorResponse(w, r, http.StatusBadRequest, "wallet id must be a valid UUID")
		return
	}

	var input struct {
		Destination string `json:"destination"`
		AmountMinor int64  `json:"amount_minor"`
		Asset       string `json:"asset"`
	}
	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	if strings.TrimSpace(input.Destination) == "" || input.AmountMinor <= 0 {
		app.errorResponse(w, r, http.StatusBadRequest, "destination and amount_minor are required")
		return
	}

	walletRecord, err := app.walletService.GetWallet(r.Context(), tenantID, walletID)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			app.notFoundErrorResponse(w, r)
			return
		}
		app.serverErrorResponse(w, r)
		return
	}

	asset := strings.TrimSpace(input.Asset)
	if asset == "" {
		asset = walletRecord.Asset
	}

	simulation, err := app.alchemy.SimulateTransfer(r.Context(), integrations.SimulateTransferRequest{
		TenantID:    tenantID,
		FromAddress: walletRecord.Address,
		ToAddress:   strings.TrimSpace(input.Destination),
		Asset:       asset,
		Network:     walletRecord.Network,
		AmountMinor: input.AmountMinor,
		ReferenceID: walletID,
	})
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	resp := envelope{
		"wallet_id":     walletRecord.ID,
		"network":       walletRecord.Network,
		"asset":         asset,
		"destination":   strings.TrimSpace(input.Destination),
		"amount_minor":  input.AmountMinor,
		"risk_level":    simulation.RiskLevel,
		"risk_reason":   simulation.Reason,
		"can_broadcast": simulation.RiskLevel != "high",
	}
	if feePolicy, err := app.model.Wallets.GetFeePolicy(r.Context(), tenantID, asset, walletRecord.Network); err == nil {
		feeAmount := (input.AmountMinor * int64(feePolicy.FeeBps)) / 10_000
		if feeAmount < feePolicy.MinFeeMinor {
			feeAmount = feePolicy.MinFeeMinor
		}
		if feePolicy.MaxFeeMinor > 0 && feeAmount > feePolicy.MaxFeeMinor {
			feeAmount = feePolicy.MaxFeeMinor
		}
		if feeAmount < 0 {
			feeAmount = 0
		}
		netAmount := input.AmountMinor - feeAmount
		if netAmount < 0 {
			netAmount = 0
		}

		resp["fee_preview"] = envelope{
			"fee_bps":           feePolicy.FeeBps,
			"fee_amount_minor":  feeAmount,
			"net_amount_minor":  netAmount,
			"fee_destination":   feePolicy.FeeDestination,
			"policy_updated_at": feePolicy.UpdatedAt,
		}
	}
	if chainFamily(walletRecord.Network) == "evm" {
		if estimate, err := app.alchemy.EstimateTransferGas(r.Context(), walletRecord.Network, asset, walletRecord.Address, strings.TrimSpace(input.Destination), input.AmountMinor); err == nil {
			resp["gas_estimate"] = estimate
		}
	}

	_ = app.writeJSON(w, http.StatusOK, envelope{"simulation": resp}, nil)
}

func (app *application) getWithdrawalTrace(w http.ResponseWriter, r *http.Request) {
	tenantID := app.tenantIDFromRequest(r)
	if tenantID == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "X-Tenant-ID header is required")
		return
	}
	withdrawalID := app.pathParam(r, "id")
	if !isValidUUID(withdrawalID) {
		app.errorResponse(w, r, http.StatusBadRequest, "withdrawal id must be a valid UUID")
		return
	}

	wd, err := app.walletService.GetWithdrawal(r.Context(), tenantID, withdrawalID)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			app.notFoundErrorResponse(w, r)
			return
		}
		app.serverErrorResponse(w, r)
		return
	}

	trace := envelope{
		"withdrawal_id":    wd.ID,
		"status":           wd.Status,
		"policy_reason":    wd.PolicyReason,
		"provider_tx_hash": wd.ProviderTxHash,
		"network":          wd.Network,
		"asset":            wd.Asset,
		"updated_at":       wd.UpdatedAt,
	}
	if strings.TrimSpace(wd.ProviderTxHash) != "" && app.alchemy != nil && chainFamily(wd.Network) == "evm" {
		receipt, err := app.alchemy.GetTxReceiptState(r.Context(), wd.Network, wd.ProviderTxHash)
		if err != nil {
			trace["receipt_error"] = err.Error()
		} else {
			trace["receipt"] = receipt
		}
	}

	_ = app.writeJSON(w, http.StatusOK, envelope{"trace": trace}, nil)
}

func (app *application) getWalletTokenAllowances(w http.ResponseWriter, r *http.Request) {
	tenantID := app.tenantIDFromRequest(r)
	if tenantID == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "X-Tenant-ID header is required")
		return
	}
	if app.alchemy == nil {
		app.serverErrorResponse(w, r)
		return
	}

	walletID := app.pathParam(r, "id")
	if !isValidUUID(walletID) {
		app.errorResponse(w, r, http.StatusBadRequest, "wallet id must be a valid UUID")
		return
	}
	walletRecord, err := app.walletService.GetWallet(r.Context(), tenantID, walletID)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			app.notFoundErrorResponse(w, r)
			return
		}
		app.serverErrorResponse(w, r)
		return
	}

	spender := strings.TrimSpace(app.readString(r.URL.Query(), "spender", ""))
	if spender == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "spender query param is required")
		return
	}
	asset := strings.TrimSpace(app.readString(r.URL.Query(), "asset", walletRecord.Asset))
	allowance, err := app.alchemy.GetTokenAllowance(r.Context(), walletRecord.Network, asset, walletRecord.Address, spender)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	_ = app.writeJSON(w, http.StatusOK, envelope{"allowance": envelope{
		"wallet_id":       walletRecord.ID,
		"network":         walletRecord.Network,
		"asset":           strings.ToUpper(asset),
		"owner":           walletRecord.Address,
		"spender":         strings.ToLower(spender),
		"allowance_minor": allowance,
	}}, nil)
}

func (app *application) getNetworkStatus(w http.ResponseWriter, r *http.Request) {
	network := strings.TrimSpace(strings.ToLower(app.pathParam(r, "network")))
	if network == "" {
		app.notFoundErrorResponse(w, r)
		return
	}
	if app.alchemy == nil {
		app.serverErrorResponse(w, r)
		return
	}
	state, err := app.alchemy.GetNetworkState(r.Context(), network)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	_ = app.writeJSON(w, http.StatusOK, envelope{"network_status": state}, nil)
}

func (app *application) getWalletRiskScore(w http.ResponseWriter, r *http.Request) {
	tenantID := app.tenantIDFromRequest(r)
	if tenantID == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "X-Tenant-ID header is required")
		return
	}
	walletID := app.pathParam(r, "id")
	if !isValidUUID(walletID) {
		app.errorResponse(w, r, http.StatusBadRequest, "wallet id must be a valid UUID")
		return
	}

	_, err := app.walletService.GetWallet(r.Context(), tenantID, walletID)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			app.notFoundErrorResponse(w, r)
			return
		}
		app.serverErrorResponse(w, r)
		return
	}

	history, _, err := app.model.Wallets.ListWithdrawalsByWallet(r.Context(), tenantID, walletID, data.Filters{
		Page: 1, PageSize: 100, Sort: "-created_at", SortSafelist: []string{"created_at", "-created_at"},
	})
	if err != nil {
		app.serverErrorResponse(w, r)
		return
	}

	score := 10
	factors := []string{"baseline_wallet_activity"}
	for _, item := range history {
		switch item.Status {
		case "failed":
			score += 15
			factors = append(factors, "failed_withdrawal_history")
		case "rejected":
			score += 8
			factors = append(factors, "policy_rejections")
		case "policy_pending":
			score += 4
		}
	}
	if score > 100 {
		score = 100
	}
	level := "low"
	if score >= 35 {
		level = "medium"
	}
	if score >= 70 {
		level = "high"
	}

	_ = app.writeJSON(w, http.StatusOK, envelope{"risk": envelope{
		"wallet_id":   walletID,
		"score":       score,
		"level":       level,
		"factors":     uniqueStrings(factors),
		"method":      "internal_rule_heuristics_v1",
		"assessed_at": time.Now().UTC(),
	}}, nil)
}

func (app *application) validateAddress(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Network string `json:"network"`
		Address string `json:"address"`
	}
	if err := app.readJSON(w, r, &input); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	network := strings.ToLower(strings.TrimSpace(input.Network))
	address := strings.TrimSpace(input.Address)
	if network == "" || address == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "network and address are required")
		return
	}

	valid, normalized, reason := validateAddressByNetwork(network, address)
	_ = app.writeJSON(w, http.StatusOK, envelope{"validation": envelope{
		"network":    network,
		"address":    address,
		"is_valid":   valid,
		"normalized": normalized,
		"reason":     reason,
	}}, nil)
}

func (app *application) getAssetMetadata(w http.ResponseWriter, r *http.Request) {
	network := strings.ToLower(strings.TrimSpace(app.pathParam(r, "network")))
	asset := strings.ToUpper(strings.TrimSpace(app.pathParam(r, "asset")))
	if network == "" || asset == "" {
		app.notFoundErrorResponse(w, r)
		return
	}

	def, err := app.model.Assets.GetActiveAsset(r.Context(), network, asset)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			app.notFoundErrorResponse(w, r)
			return
		}
		app.serverErrorResponse(w, r)
		return
	}
	_ = app.writeJSON(w, http.StatusOK, envelope{"asset": def}, nil)
}

func (app *application) getWalletNonces(w http.ResponseWriter, r *http.Request) {
	tenantID := app.tenantIDFromRequest(r)
	if tenantID == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "X-Tenant-ID header is required")
		return
	}
	if app.alchemy == nil {
		app.serverErrorResponse(w, r)
		return
	}
	walletID := app.pathParam(r, "id")
	if !isValidUUID(walletID) {
		app.errorResponse(w, r, http.StatusBadRequest, "wallet id must be a valid UUID")
		return
	}

	walletRecord, err := app.walletService.GetWallet(r.Context(), tenantID, walletID)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			app.notFoundErrorResponse(w, r)
			return
		}
		app.serverErrorResponse(w, r)
		return
	}

	nonces, err := app.alchemy.GetNonceState(r.Context(), walletRecord.Network, walletRecord.Address)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}
	_ = app.writeJSON(w, http.StatusOK, envelope{"nonces": nonces}, nil)
}

func validateAddressByNetwork(network, address string) (bool, string, string) {
	switch chainFamily(network) {
	case "evm":
		if !common.IsHexAddress(address) {
			return false, "", "invalid evm hex address"
		}
		return true, strings.ToLower(common.HexToAddress(address).Hex()), "valid"
	case "solana":
		pub, err := solana.PublicKeyFromBase58(address)
		if err != nil {
			return false, "", "invalid solana address"
		}
		return true, pub.String(), "valid"
	case "stellar":
		if _, err := keypair.ParseAddress(address); err != nil {
			return false, "", "invalid stellar address"
		}
		return true, strings.TrimSpace(address), "valid"
	case "xrpl":
		trimmed := strings.TrimSpace(address)
		if !strings.HasPrefix(trimmed, "r") || len(trimmed) < 25 {
			return false, "", "invalid xrpl address"
		}
		return true, trimmed, "valid"
	case "tron":
		trimmed := strings.TrimSpace(address)
		if !strings.HasPrefix(trimmed, "T") || len(trimmed) < 25 {
			return false, "", "invalid tron address"
		}
		return true, trimmed, "valid"
	case "aptos", "sui":
		trimmed := strings.ToLower(strings.TrimSpace(address))
		trimmed = strings.TrimPrefix(trimmed, "0x")
		if len(trimmed) == 0 || len(trimmed) > 64 {
			return false, "", "invalid hex address"
		}
		for _, r := range trimmed {
			if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')) {
				return false, "", "invalid hex address"
			}
		}
		return true, "0x" + trimmed, "valid"
	default:
		return false, "", "unsupported network family"
	}
}

func chainFamily(network string) string {
	key := strings.ToLower(strings.TrimSpace(network))
	switch {
	case strings.HasPrefix(key, "solana"):
		return "solana"
	case strings.HasPrefix(key, "stellar"):
		return "stellar"
	case strings.HasPrefix(key, "xrpl"), strings.HasPrefix(key, "xrp"):
		return "xrpl"
	case strings.HasPrefix(key, "tron"):
		return "tron"
	case strings.HasPrefix(key, "aptos"):
		return "aptos"
	case strings.HasPrefix(key, "sui"):
		return "sui"
	default:
		return "evm"
	}
}

func uniqueStrings(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	out := make([]string, 0, len(items))
	for _, item := range items {
		key := strings.TrimSpace(item)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, key)
	}
	return out
}
