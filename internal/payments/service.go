package payments

import "context"

// Service handles payment orchestration across fiat and crypto rails.
type Service interface {
	CreateTransfer(ctx context.Context, request TransferRequest) (Transfer, error)
	GetTransfer(ctx context.Context, transferID string) (Transfer, error)
}

type TransferRequest struct {
	TenantID        string
	SourceRail      string
	DestinationRail string
	Asset           string
	AmountMinor     int64
	Reference       string
}

type Transfer struct {
	ID          string
	Status      string
	Reference   string
	Asset       string
	AmountMinor int64
}
