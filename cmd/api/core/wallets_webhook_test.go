package core

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ebitezion/Nduracore/internal/data"
	"github.com/ebitezion/Nduracore/internal/integrations/alchemy"
	"github.com/ebitezion/Nduracore/internal/wallet"
)

type mockWalletService struct {
	recorded []wallet.RecordDepositInput
}

func (m *mockWalletService) CreateWallet(ctx context.Context, input wallet.CreateWalletInput) (data.Wallet, error) {
	return data.Wallet{}, nil
}
func (m *mockWalletService) GetWallet(ctx context.Context, tenantID, walletID string) (data.Wallet, error) {
	return data.Wallet{}, nil
}
func (m *mockWalletService) ListWallets(ctx context.Context, tenantID string, filters data.Filters) ([]data.Wallet, data.Metadata, error) {
	return nil, data.Metadata{}, nil
}
func (m *mockWalletService) SyncDepositsForWallet(ctx context.Context, tenantID, walletID string) (int, error) {
	return 0, nil
}
func (m *mockWalletService) ListDeposits(ctx context.Context, tenantID, walletID string, filters data.Filters) ([]data.WalletDeposit, data.Metadata, error) {
	return nil, data.Metadata{}, nil
}
func (m *mockWalletService) RecordDeposit(ctx context.Context, input wallet.RecordDepositInput) (data.WalletDeposit, error) {
	m.recorded = append(m.recorded, input)
	return data.WalletDeposit{
		TxHash:        input.TxHash,
		AmountMinor:   input.AmountMinor,
		Confirmations: input.Confirmations,
		Status:        input.Status,
	}, nil
}
func (m *mockWalletService) RequestWithdrawal(ctx context.Context, input wallet.RequestWithdrawalInput) (data.Withdrawal, error) {
	return data.Withdrawal{}, nil
}
func (m *mockWalletService) GetWithdrawal(ctx context.Context, tenantID, withdrawalID string) (data.Withdrawal, error) {
	return data.Withdrawal{}, nil
}
func (m *mockWalletService) ApproveWithdrawal(ctx context.Context, input wallet.ApproveWithdrawalInput) (data.Withdrawal, error) {
	return data.Withdrawal{}, nil
}
func (m *mockWalletService) RejectWithdrawal(ctx context.Context, input wallet.RejectWithdrawalInput) (data.Withdrawal, error) {
	return data.Withdrawal{}, nil
}

func TestParseWebhookDepositsFromActivityPayload(t *testing.T) {
	payload := []byte(`{
		"event": {
			"activity": [
				{
					"hash": "0xabc",
					"value": "0.25",
					"numConfirmations": 12,
					"rawContract": {"value": ""}
				}
			]
		}
	}`)

	provider := alchemy.NewProvider(alchemy.Config{})
	deposits, err := provider.ParseWebhookDeposits(payload)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(deposits) != 1 {
		t.Fatalf("expected 1 deposit, got %d", len(deposits))
	}
	if deposits[0].TxHash != "0xabc" {
		t.Fatalf("unexpected tx hash: %s", deposits[0].TxHash)
	}
	if deposits[0].AmountMinor <= 0 {
		t.Fatalf("expected positive amount_minor")
	}
	if deposits[0].Status != "confirmed" {
		t.Fatalf("expected confirmed status, got %s", deposits[0].Status)
	}
}

func TestIngestAlchemyDepositWebhookAcceptsActivityPayload(t *testing.T) {
	app := newTestApp()
	mockSvc := &mockWalletService{}
	app.walletService = mockSvc
	secret := "whsec_test"
	app.alchemy = alchemy.NewProvider(alchemy.Config{WebhookSigningSecret: secret})

	payload := []byte(`{
		"event": {
			"activity": [
				{
					"hash": "0xabc",
					"value": "0.25",
					"numConfirmations": 9,
					"rawContract": {"value": ""}
				},
				{
					"hash": "0xdef",
					"rawContract": {"value": "0x3e8"},
					"confirmations": 3
				}
			]
		}
	}`)

	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(payload)
	signature := hex.EncodeToString(mac.Sum(nil))

	req := httptest.NewRequest(http.MethodPost, "/v1/wallets/w1/deposits/webhook", bytes.NewReader(payload))
	req.Header.Set("X-Tenant-ID", "tenant-1")
	req.Header.Set("X-Alchemy-Signature", signature)
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d body=%s", rr.Code, rr.Body.String())
	}
	if len(mockSvc.recorded) != 2 {
		t.Fatalf("expected 2 deposits recorded, got %d", len(mockSvc.recorded))
	}
	if mockSvc.recorded[0].TenantID != "tenant-1" || mockSvc.recorded[0].WalletID != "w1" {
		t.Fatalf("unexpected routing values: tenant=%s wallet=%s", mockSvc.recorded[0].TenantID, mockSvc.recorded[0].WalletID)
	}
	if mockSvc.recorded[1].AmountMinor != 1000 {
		t.Fatalf("expected hex amount to parse to 1000, got %d", mockSvc.recorded[1].AmountMinor)
	}
}
