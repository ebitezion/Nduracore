package integrations

import "context"

type WalletProvider interface {
	Name() string
	CreateAddress(ctx context.Context, req CreateAddressRequest) (CreateAddressResult, error)
	SimulateTransfer(ctx context.Context, req SimulateTransferRequest) (SimulateTransferResult, error)
	BroadcastTransfer(ctx context.Context, req BroadcastTransferRequest) (BroadcastTransferResult, error)
	ListDeposits(ctx context.Context, req ListDepositsRequest) ([]DetectedDeposit, error)
}

type CreateAddressRequest struct {
	TenantID string
	Asset    string
	Network  string
}

type CreateAddressResult struct {
	Address string
}

type SimulateTransferRequest struct {
	TenantID    string
	FromAddress string
	ToAddress   string
	Asset       string
	Network     string
	AmountMinor int64
	ReferenceID string
}

type SimulateTransferResult struct {
	RiskLevel string
	Reason    string
}

type BroadcastTransferRequest struct {
	TenantID    string
	FromAddress string
	ToAddress   string
	Asset       string
	Network     string
	AmountMinor int64
	ReferenceID string
}

type BroadcastTransferResult struct {
	TxHash string
}

type ListDepositsRequest struct {
	TenantID string
	Address  string
	Asset    string
	Network  string
}

type DetectedDeposit struct {
	TxHash        string
	AmountMinor   int64
	Confirmations int
	Status        string
}
