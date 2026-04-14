package alchemy

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ebitezion/Nduracore/internal/integrations"
)

func TestListDepositsLiveParsesTransfers(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Method string `json:"method"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}

		switch req.Method {
		case "eth_blockNumber":
			_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": 1, "result": "0x20"})
		case "alchemy_getAssetTransfers":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"jsonrpc": "2.0",
				"id":      1,
				"result": map[string]any{
					"transfers": []map[string]any{
						{
							"hash":     "0xabc",
							"value":    "0.5",
							"blockNum": "0x1f",
							"rawContract": map[string]any{
								"value": "",
							},
						},
					},
				},
			})
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer srv.Close()

	p := NewProvider(Config{
		Network:               "eth-sepolia",
		RPCURL:                srv.URL,
		EnableLiveDeposits:    true,
		ConfirmationsRequired: 2,
	})

	deposits, err := p.ListDeposits(context.Background(), integrations.ListDepositsRequest{Address: "0x123"})
	if err != nil {
		t.Fatalf("list deposits: %v", err)
	}
	if len(deposits) != 1 {
		t.Fatalf("expected 1 deposit, got %d", len(deposits))
	}
	if deposits[0].TxHash != "0xabc" {
		t.Fatalf("unexpected tx hash: %s", deposits[0].TxHash)
	}
	if deposits[0].AmountMinor <= 0 {
		t.Fatalf("expected positive amount_minor, got %d", deposits[0].AmountMinor)
	}
	if deposits[0].Status != "confirmed" {
		t.Fatalf("expected confirmed status, got %s", deposits[0].Status)
	}
}

func TestVerifyWebhookSignature(t *testing.T) {
	secret := "super-secret"
	body := []byte(`{"tx_hash":"0xabc"}`)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	sig := hex.EncodeToString(mac.Sum(nil))

	p := NewProvider(Config{WebhookSigningSecret: secret})
	if !p.VerifyWebhookSignature(body, sig) {
		t.Fatal("expected valid signature")
	}
	if p.VerifyWebhookSignature(body, "deadbeef") {
		t.Fatal("expected invalid signature")
	}
}
