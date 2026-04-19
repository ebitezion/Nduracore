package alchemy

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestListDepositsLiveUsesRequestedNetworkRPCURL(t *testing.T) {
	mainnetCalled := false
	sepoliaCalled := false

	mainnet := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mainnetCalled = true
		_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": 1, "result": "0x1"})
	}))
	defer mainnet.Close()

	sepolia := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Method string `json:"method"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)
		sepoliaCalled = true
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
							"hash":     "0xdef",
							"value":    "1",
							"blockNum": "0x20",
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
	defer sepolia.Close()

	p := NewProvider(Config{
		Network:            "eth-mainnet",
		EnableLiveDeposits: true,
		RPCURLs: map[string]string{
			"eth-mainnet": mainnet.URL,
			"eth-sepolia": sepolia.URL,
		},
	})

	deposits, err := p.ListDeposits(context.Background(), integrations.ListDepositsRequest{
		Address: "0x123",
		Network: "eth-sepolia",
	})
	if err != nil {
		t.Fatalf("list deposits: %v", err)
	}
	if len(deposits) != 1 || deposits[0].TxHash != "0xdef" {
		t.Fatalf("unexpected deposits: %+v", deposits)
	}
	if !sepoliaCalled {
		t.Fatal("expected sepolia rpc server to be called")
	}
	if mainnetCalled {
		t.Fatal("did not expect mainnet rpc server to be called")
	}
}

func TestBroadcastTransferLiveSendsRawTransaction(t *testing.T) {
	methodCalls := make([]string, 0)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     any           `json:"id"`
			Method string        `json:"method"`
			Params []interface{} `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		methodCalls = append(methodCalls, req.Method)

		switch req.Method {
		case "eth_chainId":
			_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": "0xaa36a7"}) // sepolia
		case "eth_getTransactionCount":
			_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": "0x1"})
		case "eth_maxPriorityFeePerGas":
			_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": "0x77359400"}) // 2 gwei
		case "eth_getBlockByNumber":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"jsonrpc": "2.0",
				"id":      req.ID,
				"result": map[string]any{
					"number":           "0x1",
					"hash":             "0x1111111111111111111111111111111111111111111111111111111111111111",
					"parentHash":       "0x2222222222222222222222222222222222222222222222222222222222222222",
					"nonce":            "0x0000000000000000",
					"sha3Uncles":       "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
					"logsBloom":        "0x" + strings.Repeat("0", 512),
					"transactionsRoot": "0x3333333333333333333333333333333333333333333333333333333333333333",
					"stateRoot":        "0x4444444444444444444444444444444444444444444444444444444444444444",
					"receiptsRoot":     "0x5555555555555555555555555555555555555555555555555555555555555555",
					"miner":            "0x0000000000000000000000000000000000000000",
					"difficulty":       "0x0",
					"extraData":        "0x",
					"gasLimit":         "0x1c9c380",
					"gasUsed":          "0x0",
					"timestamp":        "0x1",
					"transactions":     []any{},
					"uncles":           []any{},
					"baseFeePerGas":    "0x77359400",
				},
			})
		case "eth_estimateGas":
			_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": "0x5208"})
		case "eth_sendRawTransaction":
			_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"})
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer srv.Close()

	p := NewProvider(Config{
		Network:              "eth-sepolia",
		RPCURLs:              map[string]string{"eth-sepolia": srv.URL},
		PrivateKeys:          map[string]string{"eth-sepolia": "4f3edf983ac636a65a842ce7c78d9aa706d3b113bce036f4f5f94e5d6f9d7f37"},
		EnableLiveBroadcasts: true,
	})

	result, err := p.BroadcastTransfer(context.Background(), integrations.BroadcastTransferRequest{
		TenantID:    "tenant-1",
		FromAddress: "0x0000000000000000000000000000000000000000",
		ToAddress:   "0x1111111111111111111111111111111111111111",
		Asset:       "ETH",
		Network:     "eth-sepolia",
		AmountMinor: 100000,
		ReferenceID: "wd-1",
	})
	if err != nil {
		t.Fatalf("broadcast transfer: %v", err)
	}
	if !strings.HasPrefix(result.TxHash, "0x") || len(result.TxHash) != 66 {
		t.Fatalf("unexpected tx hash: %s", result.TxHash)
	}

	seenSendRaw := false
	for _, method := range methodCalls {
		if method == "eth_sendRawTransaction" {
			seenSendRaw = true
			break
		}
	}
	if !seenSendRaw {
		t.Fatalf("expected eth_sendRawTransaction call, got methods=%v", methodCalls)
	}
}

func TestCreateAddressUsesVaultScopedPrivateKeyForEVM(t *testing.T) {
	p := NewProvider(Config{
		Network:              "eth-sepolia",
		PrivateKeys:          map[string]string{"eth-sepolia": "4f3edf983ac636a65a842ce7c78d9aa706d3b113bce036f4f5f94e5d6f9d7f37"},
		VaultPrivateKeys:     map[string]string{"vault-1:eth-sepolia": "6c3699283bda56ad74f6b855546325b68d482e983852a7b2d6c7f717f5f38f06"},
		EnableLiveBroadcasts: true,
	})

	vaultAddr, err := p.CreateAddress(context.Background(), integrations.CreateAddressRequest{
		TenantID: "tenant-1",
		VaultID:  "vault-1",
		Asset:    "ETH",
		Network:  "eth-sepolia",
	})
	if err != nil {
		t.Fatalf("create address with vault key: %v", err)
	}

	defaultAddr, err := p.CreateAddress(context.Background(), integrations.CreateAddressRequest{
		TenantID: "tenant-1",
		VaultID:  "vault-2",
		Asset:    "ETH",
		Network:  "eth-sepolia",
	})
	if err != nil {
		t.Fatalf("create address with default network key: %v", err)
	}

	if strings.EqualFold(vaultAddr.Address, defaultAddr.Address) {
		t.Fatalf("expected different signer addresses for vault-scoped vs default key, got %s", vaultAddr.Address)
	}
}

func TestCreateAddressReturnsErrorWhenLiveSignerInvalid(t *testing.T) {
	p := NewProvider(Config{
		Network:              "eth-sepolia",
		PrivateKeys:          map[string]string{"eth-sepolia": "not-a-valid-hex-private-key"},
		EnableLiveBroadcasts: true,
	})

	_, err := p.CreateAddress(context.Background(), integrations.CreateAddressRequest{
		TenantID: "tenant-1",
		VaultID:  "vault-1",
		Asset:    "ETH",
		Network:  "eth-sepolia",
	})
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "invalid private key") {
		t.Fatalf("expected invalid private key error, got %v", err)
	}
}

func TestBroadcastTransferUsesAssetLookupForERC20(t *testing.T) {
	methodCalls := make([]string, 0)
	contractAddress := "0xA0b86991c6218b36c1d19d4a2e9eb0ce3606eb48"
	estimateTo := ""

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     any           `json:"id"`
			Method string        `json:"method"`
			Params []interface{} `json:"params"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		methodCalls = append(methodCalls, req.Method)

		switch req.Method {
		case "eth_chainId":
			_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": "0xaa36a7"})
		case "eth_getTransactionCount":
			_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": "0x1"})
		case "eth_maxPriorityFeePerGas":
			_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": "0x77359400"})
		case "eth_getBlockByNumber":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"jsonrpc": "2.0",
				"id":      req.ID,
				"result": map[string]any{
					"number":           "0x1",
					"hash":             "0x1111111111111111111111111111111111111111111111111111111111111111",
					"parentHash":       "0x2222222222222222222222222222222222222222222222222222222222222222",
					"nonce":            "0x0000000000000000",
					"sha3Uncles":       "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
					"logsBloom":        "0x" + strings.Repeat("0", 512),
					"transactionsRoot": "0x3333333333333333333333333333333333333333333333333333333333333333",
					"stateRoot":        "0x4444444444444444444444444444444444444444444444444444444444444444",
					"receiptsRoot":     "0x5555555555555555555555555555555555555555555555555555555555555555",
					"miner":            "0x0000000000000000000000000000000000000000",
					"difficulty":       "0x0",
					"extraData":        "0x",
					"gasLimit":         "0x1c9c380",
					"gasUsed":          "0x0",
					"timestamp":        "0x1",
					"transactions":     []any{},
					"uncles":           []any{},
					"baseFeePerGas":    "0x77359400",
				},
			})
		case "eth_estimateGas":
			if len(req.Params) > 0 {
				if m, ok := req.Params[0].(map[string]interface{}); ok {
					if toRaw, ok := m["to"].(string); ok {
						estimateTo = strings.ToLower(strings.TrimSpace(toRaw))
					}
				}
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": "0xea60"})
		case "eth_sendRawTransaction":
			_ = json.NewEncoder(w).Encode(map[string]any{"jsonrpc": "2.0", "id": req.ID, "result": "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"})
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer srv.Close()

	p := NewProvider(Config{
		Network:              "eth-sepolia",
		RPCURLs:              map[string]string{"eth-sepolia": srv.URL},
		PrivateKeys:          map[string]string{"eth-sepolia": "4f3edf983ac636a65a842ce7c78d9aa706d3b113bce036f4f5f94e5d6f9d7f37"},
		EnableLiveBroadcasts: true,
		AssetLookup: func(ctx context.Context, network, asset string) (AssetDefinition, bool, error) {
			if network == "eth-sepolia" && asset == "USDC" {
				return AssetDefinition{
					Network:         network,
					AssetCode:       asset,
					ChainFamily:     "evm",
					AssetType:       "token",
					ContractAddress: contractAddress,
					Decimals:        6,
					IsActive:        true,
				}, true, nil
			}
			return AssetDefinition{}, false, nil
		},
	})

	result, err := p.BroadcastTransfer(context.Background(), integrations.BroadcastTransferRequest{
		TenantID:    "tenant-1",
		FromAddress: "0x0000000000000000000000000000000000000000",
		ToAddress:   "0x1111111111111111111111111111111111111111",
		Asset:       "USDC",
		Network:     "eth-sepolia",
		AmountMinor: 1000000,
		ReferenceID: "wd-2",
	})
	if err != nil {
		t.Fatalf("broadcast transfer: %v", err)
	}
	if !strings.HasPrefix(result.TxHash, "0x") || len(result.TxHash) != 66 {
		t.Fatalf("unexpected tx hash: %s", result.TxHash)
	}
	if estimateTo != strings.ToLower(contractAddress) {
		t.Fatalf("expected estimateGas to target contract %s, got %s", strings.ToLower(contractAddress), estimateTo)
	}

	seenSendRaw := false
	for _, method := range methodCalls {
		if method == "eth_sendRawTransaction" {
			seenSendRaw = true
			break
		}
	}
	if !seenSendRaw {
		t.Fatalf("expected eth_sendRawTransaction call, got methods=%v", methodCalls)
	}
}

func TestGetBalanceEVMNative(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     any    `json:"id"`
			Method string `json:"method"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Method != "eth_getBalance" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"jsonrpc": "2.0",
			"id":      req.ID,
			"result":  "0xde0b6b3a7640000", // 1 ETH in wei
		})
	}))
	defer srv.Close()

	p := NewProvider(Config{
		Network: "eth-sepolia",
		RPCURLs: map[string]string{"eth-sepolia": srv.URL},
	})

	result, err := p.GetBalance(context.Background(), integrations.GetBalanceRequest{
		Address: "0x1111111111111111111111111111111111111111",
		Asset:   "ETH",
		Network: "eth-sepolia",
	})
	if err != nil {
		t.Fatalf("get balance: %v", err)
	}
	if result.BalanceMinor != "1000000000000000000" {
		t.Fatalf("unexpected balance: %s", result.BalanceMinor)
	}
}

func TestGetBalanceEVMERC20(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ID     any    `json:"id"`
			Method string `json:"method"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Method != "eth_call" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"jsonrpc": "2.0",
			"id":      req.ID,
			"result":  "0xf4240", // 1_000_000
		})
	}))
	defer srv.Close()

	p := NewProvider(Config{
		Network: "eth-sepolia",
		RPCURLs: map[string]string{"eth-sepolia": srv.URL},
		AssetLookup: func(ctx context.Context, network, asset string) (AssetDefinition, bool, error) {
			return AssetDefinition{
				Network:         network,
				AssetCode:       asset,
				ChainFamily:     "evm",
				AssetType:       "token",
				ContractAddress: "0x1c7d4b196cb0c7b01d743fbc6116a902379c7238",
				IsActive:        true,
			}, true, nil
		},
	})

	result, err := p.GetBalance(context.Background(), integrations.GetBalanceRequest{
		Address: "0x1111111111111111111111111111111111111111",
		Asset:   "USDC",
		Network: "eth-sepolia",
	})
	if err != nil {
		t.Fatalf("get balance: %v", err)
	}
	if result.BalanceMinor != "1000000" {
		t.Fatalf("unexpected balance: %s", result.BalanceMinor)
	}
}
