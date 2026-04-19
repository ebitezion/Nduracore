package wallet

import (
	"context"

	"github.com/ebitezion/Nduracore/internal/data"
)

type Service interface {
	CreateWallet(ctx context.Context, input CreateWalletInput) (data.Wallet, error)
	GetWallet(ctx context.Context, tenantID, walletID string) (data.Wallet, error)
	GetWalletBalance(ctx context.Context, input GetWalletBalanceInput) (data.WalletBalance, error)
	ListWallets(ctx context.Context, input ListWalletsInput) ([]data.Wallet, data.Metadata, error)
	ListWithdrawals(ctx context.Context, input ListWithdrawalsInput) ([]data.Withdrawal, data.Metadata, error)
	SyncDepositsForWallet(ctx context.Context, tenantID, walletID string) (int, error)
	ListDeposits(ctx context.Context, tenantID, walletID string, filters data.Filters) ([]data.WalletDeposit, data.Metadata, error)
	RecordDeposit(ctx context.Context, input RecordDepositInput) (data.WalletDeposit, error)
	RequestWithdrawal(ctx context.Context, input RequestWithdrawalInput) (data.Withdrawal, error)
	GetWithdrawal(ctx context.Context, tenantID, withdrawalID string) (data.Withdrawal, error)
	ApproveWithdrawal(ctx context.Context, input ApproveWithdrawalInput) (data.Withdrawal, error)
	RejectWithdrawal(ctx context.Context, input RejectWithdrawalInput) (data.Withdrawal, error)
}

type RecordDepositInput struct {
	TenantID      string
	WalletID      string
	TxHash        string
	AmountMinor   int64
	Confirmations int
	Status        string
}

type CreateWalletInput struct {
	TenantID string
	VaultID  string
	Asset    string
	Network  string
}

type ListWalletsInput struct {
	TenantID string
	VaultID  string
	Asset    string
	Network  string
	Filters  data.Filters
}

type GetWalletBalanceInput struct {
	TenantID string
	WalletID string
	Asset    string
}

type ListWithdrawalsInput struct {
	TenantID string
	Status   string
	Filters  data.Filters
}

type RequestWithdrawalInput struct {
	TenantID    string
	WalletID    string
	VaultID     string
	Asset       string
	Network     string
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
