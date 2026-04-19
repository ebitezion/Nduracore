package wallet

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestRequestWithdrawalRejectsMixedSourceModes(t *testing.T) {
	svc := &service{}

	_, err := svc.RequestWithdrawal(context.Background(), RequestWithdrawalInput{
		TenantID:    "tenant-1",
		WalletID:    "9d3b90ae-1e80-4b4f-856f-8a6fabb91eba",
		VaultID:     "9d3b90ae-1e80-4b4f-856f-8a6fabb91eba",
		Asset:       "USDC",
		Network:     "eth-sepolia",
		Destination: "0x1111111111111111111111111111111111111111",
		AmountMinor: 1000,
	})
	if err == nil {
		t.Fatal("expected error for mixed source modes")
	}
	if !strings.Contains(err.Error(), "either wallet_id or vault_id, asset, and network") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRequestWithdrawalRequiresOneSourceMode(t *testing.T) {
	svc := &service{}

	_, err := svc.RequestWithdrawal(context.Background(), RequestWithdrawalInput{
		TenantID:    "tenant-1",
		Destination: "0x1111111111111111111111111111111111111111",
		AmountMinor: 1000,
	})
	if err == nil {
		t.Fatal("expected error when source mode is missing")
	}
	if !strings.Contains(err.Error(), "either wallet_id or vault_id, asset, and network are required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildBroadcastFailureReasonIncludesUnderlyingError(t *testing.T) {
	err := buildBroadcastFailureReason(errors.New("insufficient funds for gas * price + value"))
	if !strings.Contains(err, "broadcast_failed:") {
		t.Fatalf("expected broadcast_failed prefix, got %q", err)
	}
	if !strings.Contains(err, "insufficient funds") {
		t.Fatalf("expected underlying reason, got %q", err)
	}
}

func TestBuildBroadcastFailureReasonRedactsAlchemyCredential(t *testing.T) {
	raw := `Post "https://eth-sepolia.g.alchemy.com/v2/abc123SUPERSECRET": 401 Unauthorized`
	reason := buildBroadcastFailureReason(errors.New(raw))

	if strings.Contains(reason, "abc123SUPERSECRET") {
		t.Fatalf("expected credential redaction, got %q", reason)
	}
	if !strings.Contains(reason, "[REDACTED]") {
		t.Fatalf("expected redacted marker, got %q", reason)
	}
}
