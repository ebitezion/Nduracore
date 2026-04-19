package alchemy

import (
	"context"
	"testing"

	"github.com/ebitezion/Nduracore/internal/integrations"
)

func TestNetworkFamily(t *testing.T) {
	cases := map[string]string{
		"eth-sepolia":     "evm",
		"polygon-mainnet": "evm",
		"solana-devnet":   "solana",
		"stellar-testnet": "stellar",
		"xrpl-testnet":    "xrpl",
		"tron-mainnet":    "tron",
		"aptos-mainnet":   "aptos",
		"sui-mainnet":     "sui",
	}
	for network, expected := range cases {
		if got := networkFamily(network); got != expected {
			t.Fatalf("network=%s expected family=%s got=%s", network, expected, got)
		}
	}
}

func TestCreateAddressLiveNonEVM(t *testing.T) {
	p := NewProvider(Config{
		EnableLiveBroadcasts: true,
		PrivateKeys: map[string]string{
			"solana-devnet":   "66cDvko73yAf8LYvFMM3r8vF5vJtkk7JKMgEKwkmBC86oHdq41C7i1a2vS3zE1yCcdLLk6VUatUb32ZzVjSBXtRs",
			"stellar-testnet": "SDQQUZMIPUP5TSDWH3UJYAKUOP55IJ4KTBXTY7RCOMEFRQGYA6GIR3OD",
			"xrpl-testnet":    "sEdTtvLmJmrb7GaivhWoXRkvU4NDjVf",
			"tron-mainnet":    "4f3edf983ac636a65a842ce7c78d9aa706d3b113bce036f0a14fd79f7f8f2a01",
			"aptos-mainnet":   "0a3d0c01f86ef9fe778ef4f701f4577fd2986440f7d6be27dcf8ad7a6ec9ab80",
			"sui-mainnet":     "ce3cf90d10d94300617cd6f66cfa36f2f57bf6bf4f2f70cd533e461ce5f8342d",
		},
	})

	tests := []struct {
		name    string
		network string
	}{
		{name: "solana", network: "solana-devnet"},
		{name: "stellar", network: "stellar-testnet"},
		{name: "xrpl", network: "xrpl-testnet"},
		{name: "tron", network: "tron-mainnet"},
		{name: "aptos", network: "aptos-mainnet"},
		{name: "sui", network: "sui-mainnet"},
	}

	for _, tc := range tests {
		result, err := p.CreateAddress(context.Background(), integrations.CreateAddressRequest{
			TenantID: "tenant-1",
			Asset:    "NATIVE",
			Network:  tc.network,
		})
		if err != nil {
			t.Fatalf("%s create address error: %v", tc.name, err)
		}
		if result.Address == "" {
			t.Fatalf("%s expected non-empty address", tc.name)
		}
	}
}

func TestFormatFixedAmount(t *testing.T) {
	if got := formatFixedAmount(12345678, 7); got != "1.2345678" {
		t.Fatalf("unexpected amount: %s", got)
	}
	if got := formatFixedAmount(10000000, 7); got != "1" {
		t.Fatalf("unexpected amount: %s", got)
	}
}
