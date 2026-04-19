package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoadSmokeConfigBuildsRPCURLsFromAPIKeys(t *testing.T) {
	t.Setenv("ALCHEMY_NETWORK", "eth-sepolia")
	t.Setenv("ALCHEMY_RPC_URL", "")
	t.Setenv("ALCHEMY_RPC_URLS_JSON", `{"eth-mainnet":"https://mainnet"}`)
	t.Setenv("ALCHEMY_API_KEY", "")
	t.Setenv("ALCHEMY_API_KEYS_JSON", `{"eth-sepolia":"sepolia-key"}`)

	cfg, err := loadSmokeConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.rpcURLs["eth-mainnet"] != "https://mainnet" {
		t.Fatalf("unexpected mainnet rpc url: %q", cfg.rpcURLs["eth-mainnet"])
	}
	if cfg.rpcURLs["eth-sepolia"] != "https://eth-sepolia.g.alchemy.com/v2/sepolia-key" {
		t.Fatalf("unexpected sepolia rpc url: %q", cfg.rpcURLs["eth-sepolia"])
	}
}

func TestFetchRPCResult(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":1,"result":"0xaa36a7"}`))
	}))
	defer srv.Close()

	chainID, err := fetchRPCResult(t.Context(), srv.Client(), srv.URL, "eth_chainId", []any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if chainID != "0xaa36a7" {
		t.Fatalf("unexpected chain id: %q", chainID)
	}
}
