package wallet

import "context"

// Service defines wallet-domain use cases and isolates custody providers from callers.
type Service interface {
	CreateCustodialWallet(ctx context.Context, tenantID, asset, network string) (Wallet, error)
	GetWallet(ctx context.Context, tenantID, walletID string) (Wallet, error)
}

type Wallet struct {
	ID       string
	TenantID string
	Asset    string
	Network  string
	Address  string
	Status   string
}
