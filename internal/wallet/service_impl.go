package wallet

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ebitezion/Nduracore/internal/compliance"
	"github.com/ebitezion/Nduracore/internal/data"
	"github.com/ebitezion/Nduracore/internal/integrations"
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
	input.Asset = strings.ToUpper(strings.TrimSpace(input.Asset))
	input.Network = strings.ToLower(strings.TrimSpace(input.Network))

	if input.TenantID == "" || input.Asset == "" || input.Network == "" {
		return data.Wallet{}, fmt.Errorf("tenant_id, asset, and network are required")
	}

	addressResult, err := s.provider.CreateAddress(ctx, integrations.CreateAddressRequest{
		TenantID: input.TenantID,
		Asset:    input.Asset,
		Network:  input.Network,
	})
	if err != nil {
		return data.Wallet{}, err
	}

	wallet := data.Wallet{
		TenantID: input.TenantID,
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

func (s *service) ListWallets(ctx context.Context, tenantID string, filters data.Filters) ([]data.Wallet, data.Metadata, error) {
	return s.models.Wallets.ListWallets(ctx, strings.TrimSpace(tenantID), filters)
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
	input.Destination = strings.ToLower(strings.TrimSpace(input.Destination))
	input.RequestedBy = strings.TrimSpace(input.RequestedBy)
	if input.RequestedBy == "" {
		input.RequestedBy = "system"
	}

	if input.TenantID == "" || input.WalletID == "" || input.Destination == "" {
		return data.Withdrawal{}, fmt.Errorf("tenant_id, wallet_id and destination are required")
	}
	if input.AmountMinor <= 0 {
		return data.Withdrawal{}, fmt.Errorf("amount_minor must be greater than 0")
	}

	wallet, err := s.models.Wallets.GetWallet(ctx, input.TenantID, input.WalletID)
	if err != nil {
		return data.Withdrawal{}, err
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
		ReferenceID: input.WalletID,
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

	withdrawal := data.Withdrawal{
		TenantID:          input.TenantID,
		WalletID:          input.WalletID,
		Destination:       input.Destination,
		Asset:             wallet.Asset,
		Network:           wallet.Network,
		AmountMinor:       input.AmountMinor,
		Status:            status,
		RequiredApprovals: decision.RequiredApprovals,
		ApprovedCount:     0,
		PolicyReason:      decision.Reason,
		RiskLevel:         simulation.RiskLevel,
		RequestedBy:       input.RequestedBy,
	}

	if err := s.models.Wallets.CreateWithdrawal(ctx, &withdrawal); err != nil {
		return data.Withdrawal{}, err
	}

	if withdrawal.Status == "approved" {
		broadcast, err := s.provider.BroadcastTransfer(ctx, integrations.BroadcastTransferRequest{
			TenantID:    withdrawal.TenantID,
			FromAddress: wallet.Address,
			ToAddress:   withdrawal.Destination,
			Asset:       withdrawal.Asset,
			Network:     withdrawal.Network,
			AmountMinor: withdrawal.AmountMinor,
			ReferenceID: withdrawal.ID,
		})
		if err != nil {
			_ = s.models.Wallets.UpdateWithdrawalStatus(ctx, withdrawal.TenantID, withdrawal.ID, "failed", "", "broadcast_failed", 0)
			withdrawal.Status = "failed"
			withdrawal.PolicyReason = "broadcast_failed"
			return withdrawal, nil
		}

		withdrawal.Status = "broadcasted"
		withdrawal.ProviderTxHash = broadcast.TxHash
		if err := s.models.Wallets.UpdateWithdrawalStatus(ctx, withdrawal.TenantID, withdrawal.ID, withdrawal.Status, broadcast.TxHash, "", 0); err != nil {
			return data.Withdrawal{}, err
		}
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
		if err := s.models.Wallets.UpdateWithdrawalStatus(ctx, withdrawal.TenantID, withdrawal.ID, "policy_pending", "", "", count); err != nil {
			return data.Withdrawal{}, err
		}
		withdrawal.Status = "policy_pending"
		return *withdrawal, nil
	}

	wallet, err := s.models.Wallets.GetWallet(ctx, withdrawal.TenantID, withdrawal.WalletID)
	if err != nil {
		return data.Withdrawal{}, err
	}

	broadcast, err := s.provider.BroadcastTransfer(ctx, integrations.BroadcastTransferRequest{
		TenantID:    withdrawal.TenantID,
		FromAddress: wallet.Address,
		ToAddress:   withdrawal.Destination,
		Asset:       withdrawal.Asset,
		Network:     withdrawal.Network,
		AmountMinor: withdrawal.AmountMinor,
		ReferenceID: withdrawal.ID,
	})
	if err != nil {
		_ = s.models.Wallets.UpdateWithdrawalStatus(ctx, withdrawal.TenantID, withdrawal.ID, "failed", "", "broadcast_failed", count)
		withdrawal.Status = "failed"
		withdrawal.PolicyReason = "broadcast_failed"
		withdrawal.ApprovedCount = count
		return *withdrawal, nil
	}

	withdrawal.Status = "broadcasted"
	withdrawal.ProviderTxHash = broadcast.TxHash
	if err := s.models.Wallets.UpdateWithdrawalStatus(ctx, withdrawal.TenantID, withdrawal.ID, withdrawal.Status, broadcast.TxHash, "", count); err != nil {
		return data.Withdrawal{}, err
	}

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

	if err := s.models.Wallets.UpdateWithdrawalStatus(ctx, withdrawal.TenantID, withdrawal.ID, "rejected", "", "manually_rejected", withdrawal.ApprovedCount); err != nil {
		return data.Withdrawal{}, err
	}
	withdrawal.Status = "rejected"
	withdrawal.PolicyReason = "manually_rejected"
	return *withdrawal, nil
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
