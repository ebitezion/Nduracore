package wallet

import (
	"context"

	"github.com/ebitezion/Nduracore/internal/data"
)

type Service interface {
	CreateWallet(ctx context.Context, input CreateWalletInput) (data.Wallet, error)
	GetWallet(ctx context.Context, tenantID, walletID string) (data.Wallet, error)
	ListWallets(ctx context.Context, tenantID string, filters data.Filters) ([]data.Wallet, data.Metadata, error)
	SyncDepositsForWallet(ctx context.Context, tenantID, walletID string) (int, error)
	ListDeposits(ctx context.Context, tenantID, walletID string, filters data.Filters) ([]data.WalletDeposit, data.Metadata, error)
	RequestWithdrawal(ctx context.Context, input RequestWithdrawalInput) (data.Withdrawal, error)
	GetWithdrawal(ctx context.Context, tenantID, withdrawalID string) (data.Withdrawal, error)
	ApproveWithdrawal(ctx context.Context, input ApproveWithdrawalInput) (data.Withdrawal, error)
	RejectWithdrawal(ctx context.Context, input RejectWithdrawalInput) (data.Withdrawal, error)
}

type CreateWalletInput struct {
	TenantID string
	Asset    string
	Network  string
}

type RequestWithdrawalInput struct {
	TenantID    string
	WalletID    string
	Destination string
	AmountMinor int64
	RequestedBy string
}

type ApproveWithdrawalInput struct {
	TenantID     string
	WithdrawalID string
	ApprovedBy   string
	Reason       string
}

type RejectWithdrawalInput struct {
	TenantID     string
	WithdrawalID string
	RejectedBy   string
	Reason       string
}
