package alchemy

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ebitezion/Nduracore/internal/integrations"
)

type Provider struct {
	Network string

	mu              sync.Mutex
	depositsEmitted map[string]bool
}

func NewProvider(network string) *Provider {
	if strings.TrimSpace(network) == "" {
		network = "eth-sepolia"
	}

	return &Provider{
		Network:         network,
		depositsEmitted: make(map[string]bool),
	}
}

func (p *Provider) Name() string {
	return "alchemy"
}

func (p *Provider) CreateAddress(ctx context.Context, req integrations.CreateAddressRequest) (integrations.CreateAddressResult, error) {
	h := sha256.Sum256([]byte(req.TenantID + ":" + req.Asset + ":" + req.Network + ":" + time.Now().UTC().Format(time.RFC3339Nano)))
	address := "0x" + hex.EncodeToString(h[:20])
	return integrations.CreateAddressResult{Address: strings.ToLower(address)}, nil
}

func (p *Provider) SimulateTransfer(ctx context.Context, req integrations.SimulateTransferRequest) (integrations.SimulateTransferResult, error) {
	switch {
	case req.AmountMinor >= 10_000_000:
		return integrations.SimulateTransferResult{RiskLevel: "high", Reason: "amount_exceeds_high_risk_threshold"}, nil
	case req.AmountMinor >= 1_000_000:
		return integrations.SimulateTransferResult{RiskLevel: "medium", Reason: "amount_requires_extra_review"}, nil
	default:
		return integrations.SimulateTransferResult{RiskLevel: "low", Reason: "within_policy_threshold"}, nil
	}
}

func (p *Provider) BroadcastTransfer(ctx context.Context, req integrations.BroadcastTransferRequest) (integrations.BroadcastTransferResult, error) {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%s:%d:%s", req.TenantID, req.FromAddress, req.ToAddress, req.AmountMinor, time.Now().UTC().Format(time.RFC3339Nano))))
	return integrations.BroadcastTransferResult{TxHash: "0x" + hex.EncodeToString(h[:])}, nil
}

func (p *Provider) ListDeposits(ctx context.Context, req integrations.ListDepositsRequest) ([]integrations.DetectedDeposit, error) {
	key := strings.ToLower(strings.TrimSpace(req.Address))
	if key == "" {
		return nil, nil
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.depositsEmitted[key] {
		return nil, nil
	}

	h := sha256.Sum256([]byte(req.TenantID + ":" + key + ":deposit"))
	txHash := "0x" + hex.EncodeToString(h[:])
	p.depositsEmitted[key] = true

	return []integrations.DetectedDeposit{
		{
			TxHash:        txHash,
			AmountMinor:   250000,
			Confirmations: 12,
			Status:        "confirmed",
		},
	}, nil
}
