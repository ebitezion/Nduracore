package core

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGetWalletReturnsBadRequestForInvalidWalletID(t *testing.T) {
	app := newTestApp()

	token, err := app.security.GenerateToken("super-admin-id", "super_admin", time.Hour)
	if err != nil {
		t.Fatalf("token generation failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/wallets/0x54d1c69b4bbde118be93ad0021163fcd0d172c28", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Tenant-ID", "tenant-1")
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "wallet id must be a valid UUID") {
		t.Fatalf("expected descriptive validation message, got body=%s", rr.Body.String())
	}
}

func TestGetWalletBalanceReturnsBadRequestForInvalidWalletID(t *testing.T) {
	app := newTestApp()

	token, err := app.security.GenerateToken("super-admin-id", "super_admin", time.Hour)
	if err != nil {
		t.Fatalf("token generation failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/wallets/not-a-uuid/balance", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Tenant-ID", "tenant-1")
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "wallet id must be a valid UUID") {
		t.Fatalf("expected descriptive validation message, got body=%s", rr.Body.String())
	}
}

func TestListWalletsReturnsBadRequestForInvalidVaultIDFilter(t *testing.T) {
	app := newTestApp()

	token, err := app.security.GenerateToken("super-admin-id", "super_admin", time.Hour)
	if err != nil {
		t.Fatalf("token generation failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/wallets?vault_id=not-a-uuid", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Tenant-ID", "tenant-1")
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "vault_id must be a valid UUID") {
		t.Fatalf("expected descriptive validation message, got body=%s", rr.Body.String())
	}
}

func TestListWalletsPassesVaultAssetNetworkFiltersToService(t *testing.T) {
	app := newTestApp()
	mockSvc := &mockWalletService{}
	app.walletService = mockSvc

	token, err := app.security.GenerateToken("super-admin-id", "super_admin", time.Hour)
	if err != nil {
		t.Fatalf("token generation failed: %v", err)
	}

	vaultID := "9d3b90ae-1e80-4b4f-856f-8a6fabb91eba"
	req := httptest.NewRequest(http.MethodGet, "/v1/wallets?vault_id="+vaultID+"&asset=usdc&network=ETH-SEPOLIA&page=2&page_size=10", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Tenant-ID", "tenant-1")
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	if mockSvc.lastListWalletsInput.TenantID != "tenant-1" {
		t.Fatalf("unexpected tenant_id: %q", mockSvc.lastListWalletsInput.TenantID)
	}
	if mockSvc.lastListWalletsInput.VaultID != vaultID {
		t.Fatalf("unexpected vault_id: %q", mockSvc.lastListWalletsInput.VaultID)
	}
	if mockSvc.lastListWalletsInput.Asset != "usdc" {
		t.Fatalf("expected raw asset to reach service input, got %q", mockSvc.lastListWalletsInput.Asset)
	}
	if mockSvc.lastListWalletsInput.Network != "ETH-SEPOLIA" {
		t.Fatalf("expected raw network to reach service input, got %q", mockSvc.lastListWalletsInput.Network)
	}
	if mockSvc.lastListWalletsInput.Filters.Page != 2 || mockSvc.lastListWalletsInput.Filters.PageSize != 10 {
		t.Fatalf("unexpected filters: %+v", mockSvc.lastListWalletsInput.Filters)
	}
}

func TestGetWalletBalancePassesAssetOverrideToService(t *testing.T) {
	app := newTestApp()
	mockSvc := &mockWalletService{}
	app.walletService = mockSvc

	token, err := app.security.GenerateToken("super-admin-id", "super_admin", time.Hour)
	if err != nil {
		t.Fatalf("token generation failed: %v", err)
	}

	walletID := "9d3b90ae-1e80-4b4f-856f-8a6fabb91eba"
	req := httptest.NewRequest(http.MethodGet, "/v1/wallets/"+walletID+"/balance?asset=eth", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Tenant-ID", "tenant-1")
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	if mockSvc.lastWalletBalanceInput.TenantID != "tenant-1" {
		t.Fatalf("unexpected tenant_id: %q", mockSvc.lastWalletBalanceInput.TenantID)
	}
	if mockSvc.lastWalletBalanceInput.WalletID != walletID {
		t.Fatalf("unexpected wallet_id: %q", mockSvc.lastWalletBalanceInput.WalletID)
	}
	if mockSvc.lastWalletBalanceInput.Asset != "eth" {
		t.Fatalf("unexpected asset override: %q", mockSvc.lastWalletBalanceInput.Asset)
	}
}

func TestListWithdrawalsPassesPendingStatusFilterToService(t *testing.T) {
	app := newTestApp()
	mockSvc := &mockWalletService{}
	app.walletService = mockSvc

	token, err := app.security.GenerateToken("super-admin-id", "super_admin", time.Hour)
	if err != nil {
		t.Fatalf("token generation failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/withdrawals?status=policy_pending&page=1&page_size=20", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Tenant-ID", "tenant-1")
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	if mockSvc.lastListWithdrawalsInput.TenantID != "tenant-1" {
		t.Fatalf("unexpected tenant_id: %q", mockSvc.lastListWithdrawalsInput.TenantID)
	}
	if mockSvc.lastListWithdrawalsInput.Status != "policy_pending" {
		t.Fatalf("unexpected status filter: %q", mockSvc.lastListWithdrawalsInput.Status)
	}
}

func TestListWithdrawalsRejectsInvalidStatusFilter(t *testing.T) {
	app := newTestApp()

	token, err := app.security.GenerateToken("super-admin-id", "super_admin", time.Hour)
	if err != nil {
		t.Fatalf("token generation failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/withdrawals?status=queued", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-Tenant-ID", "tenant-1")
	rr := httptest.NewRecorder()

	app.routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d body=%s", rr.Code, rr.Body.String())
	}
	if !strings.Contains(rr.Body.String(), "must be a valid withdrawal status") {
		t.Fatalf("expected validation message, got body=%s", rr.Body.String())
	}
}
