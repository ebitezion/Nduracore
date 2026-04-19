package core

import (
	"strings"
	"testing"
)

func TestValidateConfigRejectsLiveBroadcastWithoutRPC(t *testing.T) {
	cfg := validSecurityConfigForTest()
	cfg.alchemy.enableLiveBroadcasts = true
	cfg.alchemy.privateKey = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

	err := validateConfig(cfg)
	if err == nil || !strings.Contains(err.Error(), "requires at least one configured RPC endpoint") {
		t.Fatalf("expected rpc endpoint validation error, got %v", err)
	}
}

func TestValidateConfigRejectsLiveBroadcastMissingSignerPerNetwork(t *testing.T) {
	cfg := validSecurityConfigForTest()
	cfg.alchemy.enableLiveBroadcasts = true
	cfg.alchemy.network = "eth-sepolia"
	cfg.alchemy.rpcURLs = map[string]string{
		"eth-mainnet": "https://eth-mainnet.g.alchemy.com/v2/key",
		"eth-sepolia": "https://eth-sepolia.g.alchemy.com/v2/key",
	}
	cfg.alchemy.privateKeys = map[string]string{
		"eth-sepolia": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}

	err := validateConfig(cfg)
	if err == nil || !strings.Contains(err.Error(), "missing private key for live broadcast network(s): eth-mainnet") {
		t.Fatalf("expected missing signer validation error, got %v", err)
	}
}

func TestValidateConfigRejectsInvalidERC20ContractMapping(t *testing.T) {
	cfg := validSecurityConfigForTest()
	cfg.alchemy.enableLiveBroadcasts = true
	cfg.alchemy.network = "eth-sepolia"
	cfg.alchemy.rpcURLs = map[string]string{
		"eth-sepolia": "https://eth-sepolia.g.alchemy.com/v2/key",
	}
	cfg.alchemy.privateKeys = map[string]string{
		"eth-sepolia": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
	cfg.alchemy.erc20Contracts = map[string]string{
		"eth-sepolia":      "0x1111111111111111111111111111111111111111",
		"eth-sepolia:usdc": "not-an-address",
	}

	err := validateConfig(cfg)
	if err == nil || !strings.Contains(err.Error(), "invalid ALCHEMY_ERC20_CONTRACTS_JSON entries") {
		t.Fatalf("expected invalid erc20 map error, got %v", err)
	}
}

func TestValidateConfigAcceptsValidLiveBroadcastConfig(t *testing.T) {
	cfg := validSecurityConfigForTest()
	cfg.alchemy.enableLiveBroadcasts = true
	cfg.alchemy.network = "eth-sepolia"
	cfg.alchemy.rpcURLs = map[string]string{
		"eth-mainnet": "https://eth-mainnet.g.alchemy.com/v2/key",
		"eth-sepolia": "https://eth-sepolia.g.alchemy.com/v2/key",
	}
	cfg.alchemy.privateKeys = map[string]string{
		"eth-mainnet": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"eth-sepolia": "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
	}
	cfg.alchemy.erc20Contracts = map[string]string{
		"eth-mainnet:usdc": "0xA0b86991c6218b36c1d19d4a2e9eb0ce3606eb48",
		"eth-sepolia:usdc": "0x1c7d4b196cb0c7b01d743fbc6116a902379c7238",
	}

	if err := validateConfig(cfg); err != nil {
		t.Fatalf("expected valid live broadcast config, got %v", err)
	}
}

func TestValidateConfigAcceptsVaultScopedSignerForNetwork(t *testing.T) {
	cfg := validSecurityConfigForTest()
	cfg.alchemy.enableLiveBroadcasts = true
	cfg.alchemy.network = "eth-sepolia"
	cfg.alchemy.rpcURLs = map[string]string{
		"eth-sepolia": "https://eth-sepolia.g.alchemy.com/v2/key",
	}
	cfg.alchemy.vaultPrivateKeys = map[string]string{
		"vault-1:eth-sepolia": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}

	if err := validateConfig(cfg); err != nil {
		t.Fatalf("expected vault-scoped signer config to be valid, got %v", err)
	}
}

func TestValidateConfigRejectsInvalidVaultSignerMapEntry(t *testing.T) {
	cfg := validSecurityConfigForTest()
	cfg.alchemy.enableLiveBroadcasts = true
	cfg.alchemy.network = "eth-sepolia"
	cfg.alchemy.rpcURLs = map[string]string{
		"eth-sepolia": "https://eth-sepolia.g.alchemy.com/v2/key",
	}
	cfg.alchemy.privateKeys = map[string]string{
		"eth-sepolia": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
	cfg.alchemy.vaultPrivateKeys = map[string]string{
		"bad-format": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}

	err := validateConfig(cfg)
	if err == nil || !strings.Contains(err.Error(), "invalid ALCHEMY_VAULT_PRIVATE_KEYS_JSON entries") {
		t.Fatalf("expected invalid vault key map validation error, got %v", err)
	}
}
