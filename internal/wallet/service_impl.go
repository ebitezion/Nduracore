package wallet

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ebitezion/Nduracore/internal/compliance"
	"github.com/ebitezion/Nduracore/internal/data"
	"github.com/ebitezion/Nduracore/internal/integrations"
	"github.com/ebitezion/Nduracore/internal/txstate"
)

type service struct {
	models   data.Models
	provider integrations.WalletProvider
}

func NewService(models data.Models, provider integrations.WalletProvider) Service {
	return &service{models: models, provider: provider}
}

func (s *service) CreateWallet(ctx context.Context, input CreateWalletInput) (data.Wallet, error) {
	input.TenantID = strings.TrimSpace(input.TenantID)
	input.VaultID = strings.TrimSpace(input.VaultID)
	input.Asset = strings.ToUpper(strings.TrimSpace(input.Asset))
	input.Network = strings.ToLower(strings.TrimSpace(input.Network))

	if input.TenantID == "" || input.VaultID == "" || input.Asset == "" || input.Network == "" {
		return data.Wallet{}, fmt.Errorf("tenant_id, vault_id, asset, and network are required")
	}
	if _, err := s.models.Treasury.GetVault(ctx, input.TenantID, input.VaultID); err != nil {
		if err == data.ErrRecordNotFound {
			return data.Wallet{}, fmt.Errorf("vault_id not found for tenant")
		}
		return data.Wallet{}, err
	}

	addressResult, err := s.provider.CreateAddress(ctx, integrations.CreateAddressRequest{
		TenantID: input.TenantID,
		VaultID:  input.VaultID,
		Asset:    input.Asset,
		Network:  input.Network,
	})
	if err != nil {
		return data.Wallet{}, err
	}

	wallet := data.Wallet{
		TenantID: input.TenantID,
		VaultID:  input.VaultID,
		Asset:    input.Asset,
		Network:  input.Network,
		Address:  addressResult.Address,
		Provider: s.provider.Name(),
		Status:   "active",
	}

	if err := s.models.Wallets.CreateWallet(ctx, &wallet); err != nil {
		return data.Wallet{}, err
	}

	_, err = s.models.Wallets.GetPolicyRule(ctx, input.TenantID)
	if err == data.ErrRecordNotFound {
		_ = s.models.Wallets.UpsertPolicyRule(ctx, defaultPolicy(input.TenantID))
	}

	return wallet, nil
}

func (s *service) GetWallet(ctx context.Context, tenantID, walletID string) (data.Wallet, error) {
	wallet, err := s.models.Wallets.GetWallet(ctx, strings.TrimSpace(tenantID), strings.TrimSpace(walletID))
	if err != nil {
		return data.Wallet{}, err
	}
	return *wallet, nil
}

func (s *service) GetWalletBalance(ctx context.Context, input GetWalletBalanceInput) (data.WalletBalance, error) {
	input.TenantID = strings.TrimSpace(input.TenantID)
	input.WalletID = strings.TrimSpace(input.WalletID)
	input.Asset = strings.ToUpper(strings.TrimSpace(input.Asset))

	if input.TenantID == "" || input.WalletID == "" {
		return data.WalletBalance{}, fmt.Errorf("tenant_id and wallet_id are required")
	}

	walletRecord, err := s.models.Wallets.GetWallet(ctx, input.TenantID, input.WalletID)
	if err != nil {
		return data.WalletBalance{}, err
	}

	asset := walletRecord.Asset
	if input.Asset != "" {
		asset = input.Asset
	}

	balanceResult, err := s.provider.GetBalance(ctx, integrations.GetBalanceRequest{
		TenantID: walletRecord.TenantID,
		VaultID:  walletRecord.VaultID,
		Address:  walletRecord.Address,
		Asset:    asset,
		Network:  walletRecord.Network,
	})
	if err != nil {
		return data.WalletBalance{}, err
	}

	return data.WalletBalance{
		WalletID:     walletRecord.ID,
		TenantID:     walletRecord.TenantID,
		VaultID:      walletRecord.VaultID,
		Address:      walletRecord.Address,
		Asset:        asset,
		Network:      walletRecord.Network,
		BalanceMinor: balanceResult.BalanceMinor,
	}, nil
}

func (s *service) ListWallets(ctx context.Context, input ListWalletsInput) ([]data.Wallet, data.Metadata, error) {
	input.TenantID = strings.TrimSpace(input.TenantID)
	input.VaultID = strings.TrimSpace(input.VaultID)
	input.Asset = strings.ToUpper(strings.TrimSpace(input.Asset))
	input.Network = strings.ToLower(strings.TrimSpace(input.Network))

	return s.models.Wallets.ListWallets(ctx, input.TenantID, input.VaultID, input.Asset, input.Network, input.Filters)
}

func (s *service) ListWithdrawals(ctx context.Context, input ListWithdrawalsInput) ([]data.Withdrawal, data.Metadata, error) {
	input.TenantID = strings.TrimSpace(input.TenantID)
	input.Status = strings.ToLower(strings.TrimSpace(input.Status))
	return s.models.Wallets.ListWithdrawals(ctx, input.TenantID, input.Status, input.Filters)
}

func (s *service) SyncDepositsForWallet(ctx context.Context, tenantID, walletID string) (int, error) {
	wallet, err := s.models.Wallets.GetWallet(ctx, tenantID, walletID)
	if err != nil {
		return 0, err
	}

	deposits, err := s.provider.ListDeposits(ctx, integrations.ListDepositsRequest{
		TenantID: wallet.TenantID,
		Address:  wallet.Address,
		Asset:    wallet.Asset,
		Network:  wallet.Network,
	})
	if err != nil {
		return 0, err
	}

	for _, dep := range deposits {
		deposit := data.WalletDeposit{
			WalletID:      wallet.ID,
			TenantID:      wallet.TenantID,
			Asset:         wallet.Asset,
			Network:       wallet.Network,
			TxHash:        dep.TxHash,
			AmountMinor:   dep.AmountMinor,
			Confirmations: dep.Confirmations,
			Status:        dep.Status,
		}
		if err := s.models.Wallets.UpsertDeposit(ctx, &deposit); err != nil {
			return 0, err
		}
	}

	return len(deposits), nil
}

func (s *service) ListDeposits(ctx context.Context, tenantID, walletID string, filters data.Filters) ([]data.WalletDeposit, data.Metadata, error) {
	if _, err := s.SyncDepositsForWallet(ctx, tenantID, walletID); err != nil && err != data.ErrRecordNotFound {
		return nil, data.Metadata{}, err
	}
	return s.models.Wallets.ListDeposits(ctx, strings.TrimSpace(tenantID), strings.TrimSpace(walletID), filters)
}

func (s *service) RecordDeposit(ctx context.Context, input RecordDepositInput) (data.WalletDeposit, error) {
	input.TenantID = strings.TrimSpace(input.TenantID)
	input.WalletID = strings.TrimSpace(input.WalletID)
	input.TxHash = strings.ToLower(strings.TrimSpace(input.TxHash))
	input.Status = strings.ToLower(strings.TrimSpace(input.Status))
	if input.Status == "" {
		input.Status = "pending"
	}
	if input.Confirmations < 0 {
		input.Confirmations = 0
	}
	if input.TenantID == "" || input.WalletID == "" || input.TxHash == "" {
		return data.WalletDeposit{}, fmt.Errorf("tenant_id, wallet_id and tx_hash are required")
	}
	if input.AmountMinor <= 0 {
		return data.WalletDeposit{}, fmt.Errorf("amount_minor must be greater than 0")
	}

	walletRecord, err := s.models.Wallets.GetWallet(ctx, input.TenantID, input.WalletID)
	if err != nil {
		return data.WalletDeposit{}, err
	}

	deposit := data.WalletDeposit{
		WalletID:      walletRecord.ID,
		TenantID:      walletRecord.TenantID,
		Asset:         walletRecord.Asset,
		Network:       walletRecord.Network,
		TxHash:        input.TxHash,
		AmountMinor:   input.AmountMinor,
		Confirmations: input.Confirmations,
		Status:        input.Status,
	}
	if err := s.models.Wallets.UpsertDeposit(ctx, &deposit); err != nil {
		return data.WalletDeposit{}, err
	}
	return deposit, nil
}

func (s *service) RequestWithdrawal(ctx context.Context, input RequestWithdrawalInput) (data.Withdrawal, error) {
	input.TenantID = strings.TrimSpace(input.TenantID)
	input.WalletID = strings.TrimSpace(input.WalletID)
	input.VaultID = strings.TrimSpace(input.VaultID)
	input.Asset = strings.ToUpper(strings.TrimSpace(input.Asset))
	input.Network = strings.ToLower(strings.TrimSpace(input.Network))
	input.Destination = strings.ToLower(strings.TrimSpace(input.Destination))
	input.RequestedBy = strings.TrimSpace(input.RequestedBy)
	if input.RequestedBy == "" {
		input.RequestedBy = "system"
	}

	if input.TenantID == "" || input.Destination == "" {
		return data.Withdrawal{}, fmt.Errorf("tenant_id and destination are required")
	}
	if input.AmountMinor <= 0 {
		return data.Withdrawal{}, fmt.Errorf("amount_minor must be greater than 0")
	}

	hasWalletMode := input.WalletID != ""
	hasVaultMode := input.VaultID != "" || input.Asset != "" || input.Network != ""
	if hasWalletMode && hasVaultMode {
		return data.Withdrawal{}, fmt.Errorf("provide either wallet_id or vault_id, asset, and network")
	}
	if !hasWalletMode && !hasVaultMode {
		return data.Withdrawal{}, fmt.Errorf("either wallet_id or vault_id, asset, and network are required")
	}

	var (
		wallet           *data.Wallet
		err              error
		resolvedWalletID string
	)
	if hasWalletMode {
		wallet, err = s.models.Wallets.GetWallet(ctx, input.TenantID, input.WalletID)
		if err != nil {
			if err == data.ErrRecordNotFound {
				return data.Withdrawal{}, fmt.Errorf("wallet_id not found for tenant")
			}
			return data.Withdrawal{}, err
		}
		if wallet.Status != "active" {
			return data.Withdrawal{}, fmt.Errorf("wallet_id is not active")
		}

		resolvedWalletID = wallet.ID
		input.VaultID = strings.TrimSpace(wallet.VaultID)
	} else {
		if input.VaultID == "" || input.Asset == "" || input.Network == "" {
			return data.Withdrawal{}, fmt.Errorf("vault_id, asset, and network are required in vault mode")
		}

		if _, err = s.models.Treasury.GetVault(ctx, input.TenantID, input.VaultID); err != nil {
			if err == data.ErrRecordNotFound {
				return data.Withdrawal{}, fmt.Errorf("vault_id not found for tenant")
			}
			return data.Withdrawal{}, err
		}
		wallet, err = s.models.Wallets.GetWalletByVaultAssetNetwork(ctx, input.TenantID, input.VaultID, input.Asset, input.Network)
		if err != nil {
			if err == data.ErrRecordNotFound {
				return data.Withdrawal{}, fmt.Errorf("no active wallet found for vault_id, asset, and network")
			}
			return data.Withdrawal{}, err
		}
		resolvedWalletID = wallet.ID
	}

	rule, err := s.models.Wallets.GetPolicyRule(ctx, input.TenantID)
	if err == data.ErrRecordNotFound {
		rule = defaultPolicy(input.TenantID)
		if upsertErr := s.models.Wallets.UpsertPolicyRule(ctx, rule); upsertErr != nil {
			return data.Withdrawal{}, upsertErr
		}
	} else if err != nil {
		return data.Withdrawal{}, err
	}

	recentWithdrawals, err := s.models.Wallets.ListRecentWithdrawals(ctx, input.TenantID, time.Now().UTC().Add(-24*time.Hour))
	if err != nil {
		return data.Withdrawal{}, err
	}

	simulation, err := s.provider.SimulateTransfer(ctx, integrations.SimulateTransferRequest{
		TenantID:    input.TenantID,
		FromAddress: wallet.Address,
		ToAddress:   input.Destination,
		Asset:       wallet.Asset,
		Network:     wallet.Network,
		AmountMinor: input.AmountMinor,
		ReferenceID: resolvedWalletID,
	})
	if err != nil {
		return data.Withdrawal{}, err
	}

	decision := compliance.EvaluateWithdrawalPolicy(compliance.WithdrawalPolicyInput{
		Policy:            rule,
		AmountMinor:       input.AmountMinor,
		Destination:       input.Destination,
		RecentWithdrawals: recentWithdrawals,
		Risk: compliance.WithdrawalRiskAssessment{
			Level:  simulation.RiskLevel,
			Reason: simulation.Reason,
		},
		Now: time.Now().UTC(),
	})

	status := "policy_pending"
	if !decision.Allowed {
		status = "rejected"
	}
	if decision.Allowed && decision.RequiredApprovals == 0 {
		status = "approved"
	}

	grossAmountMinor := input.AmountMinor
	feeAmountMinor, feePolicyReason := s.computeFee(ctx, input.TenantID, wallet.Asset, wallet.Network, grossAmountMinor)
	netAmountMinor := grossAmountMinor - feeAmountMinor
	if netAmountMinor <= 0 {
		return data.Withdrawal{}, fmt.Errorf("amount_minor is too small after fee deduction")
	}

	withdrawal := data.Withdrawal{
		TenantID:          input.TenantID,
		WalletID:          resolvedWalletID,
		VaultID:           input.VaultID,
		Destination:       input.Destination,
		Asset:             wallet.Asset,
		Network:           wallet.Network,
		AmountMinor:       netAmountMinor,
		Status:            status,
		RequiredApprovals: decision.RequiredApprovals,
		ApprovedCount:     0,
		PolicyReason:      decision.Reason,
		RiskLevel:         simulation.RiskLevel,
		RequestedBy:       input.RequestedBy,
	}
	if feePolicyReason != "" {
		withdrawal.PolicyReason = strings.TrimSpace(withdrawal.PolicyReason + "; " + feePolicyReason)
	}

	if err := s.models.Wallets.CreateWithdrawal(ctx, &withdrawal); err != nil {
		return data.Withdrawal{}, err
	}
	_ = s.models.Wallets.AddWithdrawalStateHistory(ctx, &data.WithdrawalStateHistory{
		WithdrawalID: withdrawal.ID,
		TenantID:     withdrawal.TenantID,
		FromState:    "",
		ToState:      status,
		Reason:       withdrawal.PolicyReason,
	})
	if feeAmountMinor > 0 {
		_ = s.models.Wallets.CreateFeeCharge(ctx, &data.FeeCharge{
			WithdrawalID:     withdrawal.ID,
			TenantID:         withdrawal.TenantID,
			GrossAmountMinor: grossAmountMinor,
			FeeAmountMinor:   feeAmountMinor,
			NetAmountMinor:   netAmountMinor,
			Status:           "planned",
		})
	}

	if withdrawal.Status == "approved" {
		broadcast, err := s.provider.BroadcastTransfer(ctx, integrations.BroadcastTransferRequest{
			TenantID:    withdrawal.TenantID,
			VaultID:     withdrawal.VaultID,
			FromAddress: wallet.Address,
			ToAddress:   withdrawal.Destination,
			Asset:       withdrawal.Asset,
			Network:     withdrawal.Network,
			AmountMinor: withdrawal.AmountMinor,
			ReferenceID: withdrawal.ID,
		})
		if err != nil {
			reason := buildBroadcastFailureReason(err)
			if txstate.CanTransition(withdrawal.Status, txstate.StateFailed) {
				_ = s.models.Wallets.UpdateWithdrawalStatus(ctx, withdrawal.TenantID, withdrawal.ID, txstate.StateFailed, "", reason, 0)
				_ = s.models.Wallets.AddWithdrawalStateHistory(ctx, &data.WithdrawalStateHistory{
					WithdrawalID: withdrawal.ID,
					TenantID:     withdrawal.TenantID,
					FromState:    txstate.StateApproved,
					ToState:      txstate.StateFailed,
					Reason:       reason,
				})
			}
			withdrawal.Status = "failed"
			withdrawal.PolicyReason = reason
			return withdrawal, nil
		}

		withdrawal.Status = txstate.StateBroadcasted
		withdrawal.ProviderTxHash = broadcast.TxHash
		if err := s.models.Wallets.UpdateWithdrawalStatus(ctx, withdrawal.TenantID, withdrawal.ID, withdrawal.Status, broadcast.TxHash, "", 0); err != nil {
			return data.Withdrawal{}, err
		}
		_ = s.models.Wallets.AddWithdrawalStateHistory(ctx, &data.WithdrawalStateHistory{
			WithdrawalID: withdrawal.ID,
			TenantID:     withdrawal.TenantID,
			FromState:    txstate.StateApproved,
			ToState:      txstate.StateBroadcasted,
			Reason:       "",
		})
	}

	return withdrawal, nil
}

func (s *service) GetWithdrawal(ctx context.Context, tenantID, withdrawalID string) (data.Withdrawal, error) {
	withdrawal, err := s.models.Wallets.GetWithdrawal(ctx, strings.TrimSpace(tenantID), strings.TrimSpace(withdrawalID))
	if err != nil {
		return data.Withdrawal{}, err
	}
	return *withdrawal, nil
}

func (s *service) ApproveWithdrawal(ctx context.Context, input ApproveWithdrawalInput) (data.Withdrawal, error) {
	if strings.TrimSpace(input.ApprovedBy) == "" {
		input.ApprovedBy = "approver"
	}

	withdrawal, err := s.models.Wallets.GetWithdrawal(ctx, input.TenantID, input.WithdrawalID)
	if err != nil {
		return data.Withdrawal{}, err
	}
	if withdrawal.Status != "policy_pending" {
		return data.Withdrawal{}, fmt.Errorf("withdrawal is not awaiting approval")
	}

	approval := data.WithdrawalApproval{
		WithdrawalID: withdrawal.ID,
		ApprovedBy:   input.ApprovedBy,
		Decision:     "approved",
		Reason:       strings.TrimSpace(input.Reason),
	}
	if err := s.models.Wallets.AddApproval(ctx, &approval); err != nil {
		return data.Withdrawal{}, err
	}

	count, err := s.models.Wallets.CountApprovals(ctx, withdrawal.ID)
	if err != nil {
		return data.Withdrawal{}, err
	}
	withdrawal.ApprovedCount = count

	if count < withdrawal.RequiredApprovals {
		if err := s.models.Wallets.UpdateWithdrawalStatus(ctx, withdrawal.TenantID, withdrawal.ID, txstate.StatePolicyPending, "", "", count); err != nil {
			return data.Withdrawal{}, err
		}
		withdrawal.Status = txstate.StatePolicyPending
		return *withdrawal, nil
	}

	wallet, err := s.models.Wallets.GetWallet(ctx, withdrawal.TenantID, withdrawal.WalletID)
	if err != nil {
		return data.Withdrawal{}, err
	}

	broadcast, err := s.provider.BroadcastTransfer(ctx, integrations.BroadcastTransferRequest{
		TenantID:    withdrawal.TenantID,
		VaultID:     withdrawal.VaultID,
		FromAddress: wallet.Address,
		ToAddress:   withdrawal.Destination,
		Asset:       withdrawal.Asset,
		Network:     withdrawal.Network,
		AmountMinor: withdrawal.AmountMinor,
		ReferenceID: withdrawal.ID,
	})
	if err != nil {
		reason := buildBroadcastFailureReason(err)
		_ = s.models.Wallets.UpdateWithdrawalStatus(ctx, withdrawal.TenantID, withdrawal.ID, txstate.StateFailed, "", reason, count)
		_ = s.models.Wallets.AddWithdrawalStateHistory(ctx, &data.WithdrawalStateHistory{
			WithdrawalID: withdrawal.ID,
			TenantID:     withdrawal.TenantID,
			FromState:    txstate.StatePolicyPending,
			ToState:      txstate.StateFailed,
			Reason:       reason,
		})
		withdrawal.Status = txstate.StateFailed
		withdrawal.PolicyReason = reason
		withdrawal.ApprovedCount = count
		return *withdrawal, nil
	}

	withdrawal.Status = txstate.StateBroadcasted
	withdrawal.ProviderTxHash = broadcast.TxHash
	if err := s.models.Wallets.UpdateWithdrawalStatus(ctx, withdrawal.TenantID, withdrawal.ID, withdrawal.Status, broadcast.TxHash, "", count); err != nil {
		return data.Withdrawal{}, err
	}
	_ = s.models.Wallets.AddWithdrawalStateHistory(ctx, &data.WithdrawalStateHistory{
		WithdrawalID: withdrawal.ID,
		TenantID:     withdrawal.TenantID,
		FromState:    txstate.StatePolicyPending,
		ToState:      txstate.StateBroadcasted,
		Reason:       "",
	})

	return *withdrawal, nil
}

func (s *service) RejectWithdrawal(ctx context.Context, input RejectWithdrawalInput) (data.Withdrawal, error) {
	if strings.TrimSpace(input.RejectedBy) == "" {
		input.RejectedBy = "approver"
	}

	withdrawal, err := s.models.Wallets.GetWithdrawal(ctx, input.TenantID, input.WithdrawalID)
	if err != nil {
		return data.Withdrawal{}, err
	}
	if withdrawal.Status != "policy_pending" {
		return data.Withdrawal{}, fmt.Errorf("withdrawal is not awaiting approval")
	}

	approval := data.WithdrawalApproval{
		WithdrawalID: withdrawal.ID,
		ApprovedBy:   input.RejectedBy,
		Decision:     "rejected",
		Reason:       strings.TrimSpace(input.Reason),
	}
	if err := s.models.Wallets.AddApproval(ctx, &approval); err != nil {
		return data.Withdrawal{}, err
	}

	if err := s.models.Wallets.UpdateWithdrawalStatus(ctx, withdrawal.TenantID, withdrawal.ID, txstate.StateRejected, "", "manually_rejected", withdrawal.ApprovedCount); err != nil {
		return data.Withdrawal{}, err
	}
	_ = s.models.Wallets.AddWithdrawalStateHistory(ctx, &data.WithdrawalStateHistory{
		WithdrawalID: withdrawal.ID,
		TenantID:     withdrawal.TenantID,
		FromState:    txstate.StatePolicyPending,
		ToState:      txstate.StateRejected,
		Reason:       "manually_rejected",
	})
	withdrawal.Status = txstate.StateRejected
	withdrawal.PolicyReason = "manually_rejected"
	return *withdrawal, nil
}

func (s *service) computeFee(ctx context.Context, tenantID, asset, network string, grossMinor int64) (int64, string) {
	policy, err := s.models.Wallets.GetFeePolicy(ctx, tenantID, asset, network)
	if err != nil {
		return 0, ""
	}
	if policy.FeeBps <= 0 || grossMinor <= 0 {
		return 0, ""
	}

	fee := (grossMinor * int64(policy.FeeBps)) / 10_000
	if fee < policy.MinFeeMinor {
		fee = policy.MinFeeMinor
	}
	if policy.MaxFeeMinor > 0 && fee > policy.MaxFeeMinor {
		fee = policy.MaxFeeMinor
	}
	if fee < 0 {
		fee = 0
	}
	if fee >= grossMinor {
		fee = grossMinor - 1
	}
	if fee < 0 {
		fee = 0
	}

	return fee, fmt.Sprintf("fee_applied_bps:%d fee_minor:%d", policy.FeeBps, fee)
}

func defaultPolicy(tenantID string) data.PolicyRule {
	return data.PolicyRule{
		TenantID:              tenantID,
		WhitelistDestinations: []string{"0x1111111111111111111111111111111111111111"},
		PerTxLimitMinor:       2_500_000,
		DailyLimitMinor:       10_000_000,
		VelocityCount:         5,
		VelocityWindowMinutes: 60,
		ApprovalTier:          "one",
	}
}

func buildBroadcastFailureReason(err error) string {
	if err == nil {
		return "broadcast_failed"
	}

	// Never return raw transport/provider errors with embedded credentials.
	reason := strings.TrimSpace(err.Error())
	reason = redactAlchemyCredential(reason)
	if reason == "" {
		return "broadcast_failed"
	}

	const prefix = "broadcast_failed: "
	const maxLen = 240
	available := maxLen - len(prefix)
	if available < 1 {
		return "broadcast_failed"
	}
	if len(reason) > available {
		reason = reason[:available]
	}
	return prefix + reason
}

func redactAlchemyCredential(message string) string {
	const marker = ".g.alchemy.com/v2/"
	idx := strings.Index(message, marker)
	if idx == -1 {
		return message
	}

	start := idx + len(marker)
	if start >= len(message) {
		return message
	}

	end := start
	for end < len(message) {
		switch message[end] {
		case ' ', '"', '\'', '\n', '\t', '\r', ',', ')':
			return message[:start] + "[REDACTED]" + message[end:]
		default:
			end++
		}
	}
	return message[:start] + "[REDACTED]"
}
