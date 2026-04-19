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
	"time"

	"github.com/ebitezion/Nduracore/internal/data"
	"github.com/ebitezion/Nduracore/internal/integrations/alchemy"
	"github.com/ebitezion/Nduracore/internal/wallet"
)

type mockWalletService struct {
	recorded                  []wallet.RecordDepositInput
	lastWithdrawalRequestBody wallet.RequestWithdrawalInput
	lastListWalletsInput      wallet.ListWalletsInput
	lastListWithdrawalsInput  wallet.ListWithdrawalsInput
	lastWalletBalanceInput    wallet.GetWalletBalanceInput
}

func (m *mockWalletService) CreateWallet(ctx context.Context, input wallet.CreateWalletInput) (data.Wallet, error) {
	return data.Wallet{}, nil
}
func (m *mockWalletService) GetWallet(ctx context.Context, tenantID, walletID string) (data.Wallet, error) {
	return data.Wallet{}, nil
}
func (m *mockWalletService) GetWalletBalance(ctx context.Context, input wallet.GetWalletBalanceInput) (data.WalletBalance, error) {
	m.lastWalletBalanceInput = input
	return data.WalletBalance{
		WalletID:     input.WalletID,
		TenantID:     input.TenantID,
		Asset:        "USDC",
		Network:      "eth-sepolia",
		BalanceMinor: "1000",
	}, nil
}
func (m *mockWalletService) ListWallets(ctx context.Context, input wallet.ListWalletsInput) ([]data.Wallet, data.Metadata, error) {
	m.lastListWalletsInput = input
	return nil, data.Metadata{}, nil
}
func (m *mockWalletService) ListWithdrawals(ctx context.Context, input wallet.ListWithdrawalsInput) ([]data.Withdrawal, data.Metadata, error) {
	m.lastListWithdrawalsInput = input
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
	m.lastWithdrawalRequestBody = input
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
	walletID := "11111111-1111-1111-1111-111111111111"

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

	req := httptest.NewRequest(http.MethodPost, "/v1/wallets/"+walletID+"/deposits/webhook", bytes.NewReader(payload))
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
	if mockSvc.recorded[0].TenantID != "tenant-1" || mockSvc.recorded[0].WalletID != walletID {
		t.Fatalf("unexpected routing values: tenant=%s wallet=%s", mockSvc.recorded[0].TenantID, mockSvc.recorded[0].WalletID)
	}
	if mockSvc.recorded[1].AmountMinor != 1000 {
		t.Fatalf("expected hex amount to parse to 1000, got %d", mockSvc.recorded[1].AmountMinor)
	}
}

func TestCreateWithdrawalPassesVaultAwareFieldsToService(t *testing.T) {
	app := newTestApp()
	mockSvc := &mockWalletService{}
	app.walletService = mockSvc

	token, err := app.security.GenerateToken("super-admin-id", "super_admin", time.Hour)
	if err != nil {
		t.Fatalf("token generation failed: %v", err)
	}

	payload := []byte(`{
		"vault_id":"v-1",
		"asset":"USDC",
		"network":"eth-sepolia",
		"destination":"0x1111111111111111111111111111111111111111",
		"amount_minor":25000
	}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/withdrawals", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Tenant-ID", "tenant-1")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)
	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d body=%s", rr.Code, rr.Body.String())
	}

	if mockSvc.lastWithdrawalRequestBody.VaultID != "v-1" {
		t.Fatalf("expected vault_id to pass through, got %q", mockSvc.lastWithdrawalRequestBody.VaultID)
	}
	if mockSvc.lastWithdrawalRequestBody.WalletID != "" {
		t.Fatalf("expected wallet_id to be empty in vault mode, got %q", mockSvc.lastWithdrawalRequestBody.WalletID)
	}
	if mockSvc.lastWithdrawalRequestBody.Asset != "USDC" {
		t.Fatalf("expected asset to pass through, got %q", mockSvc.lastWithdrawalRequestBody.Asset)
	}
	if mockSvc.lastWithdrawalRequestBody.Network != "eth-sepolia" {
		t.Fatalf("expected network to pass through, got %q", mockSvc.lastWithdrawalRequestBody.Network)
	}
}

func TestCreateWithdrawalPassesWalletModeFieldsToService(t *testing.T) {
	app := newTestApp()
	mockSvc := &mockWalletService{}
	app.walletService = mockSvc

	token, err := app.security.GenerateToken("super-admin-id", "super_admin", time.Hour)
	if err != nil {
		t.Fatalf("token generation failed: %v", err)
	}

	payload := []byte(`{
		"wallet_id":"9d3b90ae-1e80-4b4f-856f-8a6fabb91eba",
		"destination":"0x1111111111111111111111111111111111111111",
		"amount_minor":25000
	}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/withdrawals", bytes.NewReader(payload))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Tenant-ID", "tenant-1")
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)
	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d body=%s", rr.Code, rr.Body.String())
	}

	if mockSvc.lastWithdrawalRequestBody.WalletID != "9d3b90ae-1e80-4b4f-856f-8a6fabb91eba" {
		t.Fatalf("expected wallet_id to pass through, got %q", mockSvc.lastWithdrawalRequestBody.WalletID)
	}
	if mockSvc.lastWithdrawalRequestBody.VaultID != "" || mockSvc.lastWithdrawalRequestBody.Asset != "" || mockSvc.lastWithdrawalRequestBody.Network != "" {
		t.Fatalf("expected vault fields to be empty in wallet mode, got vault_id=%q asset=%q network=%q",
			mockSvc.lastWithdrawalRequestBody.VaultID,
			mockSvc.lastWithdrawalRequestBody.Asset,
			mockSvc.lastWithdrawalRequestBody.Network,
		)
	}
}
