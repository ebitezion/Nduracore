package alchemy

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ebitezion/Nduracore/internal/integrations"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Config struct {
	Network               string
	APIKey                string
	APIKeys               map[string]string
	RPCURL                string
	RPCURLs               map[string]string
	PrivateKey            string
	PrivateKeys           map[string]string
	VaultPrivateKeys      map[string]string
	ERC20Contracts        map[string]string
	ConfirmationsRequired int
	WebhookSigningSecret  string
	EnableLiveDeposits    bool
	EnableLiveBroadcasts  bool
	AssetLookup           AssetLookupFunc
}

type AssetDefinition struct {
	Network         string
	AssetCode       string
	ChainFamily     string
	AssetType       string
	ContractAddress string
	Decimals        int
	IsActive        bool
}

type AssetLookupFunc func(ctx context.Context, network, asset string) (AssetDefinition, bool, error)

type Provider struct {
	network               string
	rpcURLs               map[string]string
	privateKeys           map[string]string
	vaultPrivateKeys      map[string]string
	erc20Contracts        map[string]string
	confirmationsRequired int
	webhookSigningSecret  string
	enableLiveDeposits    bool
	enableLiveBroadcasts  bool
	assetLookup           AssetLookupFunc
	httpClient            *http.Client

	mu              sync.Mutex
	depositsEmitted map[string]bool
}

func NewProvider(cfg Config) *Provider {
	network := strings.TrimSpace(cfg.Network)
	if network == "" {
		network = "eth-sepolia"
	}
	network = strings.ToLower(network)

	rpcURLs := make(map[string]string)
	for n, url := range cfg.RPCURLs {
		key := strings.ToLower(strings.TrimSpace(n))
		value := strings.TrimSpace(url)
		if key == "" || value == "" {
			continue
		}
		rpcURLs[key] = value
	}
	for n, apiKey := range cfg.APIKeys {
		key := strings.ToLower(strings.TrimSpace(n))
		value := strings.TrimSpace(apiKey)
		if key == "" || value == "" {
			continue
		}
		if _, exists := rpcURLs[key]; !exists {
			rpcURLs[key] = fmt.Sprintf("https://%s.g.alchemy.com/v2/%s", key, value)
		}
	}

	rpcURL := strings.TrimSpace(cfg.RPCURL)
	if rpcURL != "" {
		rpcURLs[network] = rpcURL
	} else if _, exists := rpcURLs[network]; !exists && strings.TrimSpace(cfg.APIKey) != "" {
		rpcURLs[network] = fmt.Sprintf("https://%s.g.alchemy.com/v2/%s", network, strings.TrimSpace(cfg.APIKey))
	}

	privateKeys := make(map[string]string)
	for n, key := range cfg.PrivateKeys {
		networkKey := strings.ToLower(strings.TrimSpace(n))
		privateKey := strings.TrimSpace(strings.TrimPrefix(key, "0x"))
		if networkKey == "" || privateKey == "" {
			continue
		}
		privateKeys[networkKey] = privateKey
	}
	singlePrivateKey := strings.TrimSpace(strings.TrimPrefix(cfg.PrivateKey, "0x"))
	if singlePrivateKey != "" {
		privateKeys[network] = singlePrivateKey
	}

	vaultPrivateKeys := make(map[string]string)
	for scopeKey, key := range cfg.VaultPrivateKeys {
		parts := strings.Split(strings.ToLower(strings.TrimSpace(scopeKey)), ":")
		if len(parts) != 2 {
			continue
		}
		vaultID := strings.TrimSpace(parts[0])
		networkKey := strings.TrimSpace(parts[1])
		privateKey := strings.TrimSpace(strings.TrimPrefix(key, "0x"))
		if vaultID == "" || networkKey == "" || privateKey == "" {
			continue
		}
		vaultPrivateKeys[vaultID+":"+networkKey] = privateKey
	}

	erc20Contracts := make(map[string]string)
	for k, v := range cfg.ERC20Contracts {
		parts := strings.Split(strings.TrimSpace(strings.ToLower(k)), ":")
		if len(parts) != 2 {
			continue
		}
		networkKey := strings.TrimSpace(parts[0])
		assetKey := strings.ToUpper(strings.TrimSpace(parts[1]))
		address := strings.TrimSpace(v)
		if networkKey == "" || assetKey == "" || !common.IsHexAddress(address) {
			continue
		}
		erc20Contracts[networkKey+":"+assetKey] = common.HexToAddress(address).Hex()
	}

	confirmations := cfg.ConfirmationsRequired
	if confirmations < 1 {
		confirmations = 12
	}

	enableLive := cfg.EnableLiveDeposits && len(rpcURLs) > 0

	return &Provider{
		network:               network,
		rpcURLs:               rpcURLs,
		privateKeys:           privateKeys,
		vaultPrivateKeys:      vaultPrivateKeys,
		erc20Contracts:        erc20Contracts,
		confirmationsRequired: confirmations,
		webhookSigningSecret:  strings.TrimSpace(cfg.WebhookSigningSecret),
		enableLiveDeposits:    enableLive,
		enableLiveBroadcasts:  cfg.EnableLiveBroadcasts,
		assetLookup:           cfg.AssetLookup,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		depositsEmitted: make(map[string]bool),
	}
}

func (p *Provider) Name() string {
	return "alchemy"
}

func (p *Provider) CreateAddress(ctx context.Context, req integrations.CreateAddressRequest) (integrations.CreateAddressResult, error) {
	if p.enableLiveBroadcasts {
		switch networkFamily(req.Network) {
		case "solana":
			_, pubKey, err := p.solanaSignerForNetwork(req.VaultID, req.Network)
			if err != nil {
				return integrations.CreateAddressResult{}, err
			}
			return integrations.CreateAddressResult{Address: pubKey.String()}, nil
		case "stellar":
			signerAddress, err := p.stellarAddressForNetwork(req.VaultID, req.Network)
			if err != nil {
				return integrations.CreateAddressResult{}, err
			}
			return integrations.CreateAddressResult{Address: signerAddress}, nil
		case "xrpl":
			signerAddress, err := p.xrplAddressForNetwork(req.VaultID, req.Network)
			if err != nil {
				return integrations.CreateAddressResult{}, err
			}
			return integrations.CreateAddressResult{Address: signerAddress}, nil
		case "tron":
			signerAddress, err := p.tronAddressForNetwork(req.VaultID, req.Network)
			if err != nil {
				return integrations.CreateAddressResult{}, err
			}
			return integrations.CreateAddressResult{Address: signerAddress}, nil
		case "aptos":
			signerAddress, err := p.aptosAddressForNetwork(req.VaultID, req.Network)
			if err != nil {
				return integrations.CreateAddressResult{}, err
			}
			return integrations.CreateAddressResult{Address: signerAddress}, nil
		case "sui":
			signerAddress, err := p.suiAddressForNetwork(req.VaultID, req.Network)
			if err != nil {
				return integrations.CreateAddressResult{}, err
			}
			return integrations.CreateAddressResult{Address: signerAddress}, nil
		default:
			_, signerAddress, err := p.signerForNetwork(req.VaultID, req.Network)
			if err != nil {
				return integrations.CreateAddressResult{}, err
			}
			return integrations.CreateAddressResult{Address: strings.ToLower(signerAddress.Hex())}, nil
		}
	}

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
	if !p.enableLiveBroadcasts {
		return integrations.BroadcastTransferResult{}, fmt.Errorf("live broadcasts are disabled")
	}

	switch networkFamily(req.Network) {
	case "solana":
		return p.broadcastTransferSolana(ctx, req)
	case "stellar":
		return p.broadcastTransferStellar(ctx, req)
	case "xrpl":
		return p.broadcastTransferXRPL(ctx, req)
	case "tron":
		return p.broadcastTransferTron(ctx, req)
	case "aptos":
		return p.broadcastTransferAptos(ctx, req)
	case "sui":
		return p.broadcastTransferSui(ctx, req)
	default:
		return p.broadcastTransferEVM(ctx, req)
	}
}

func (p *Provider) broadcastTransferEVM(ctx context.Context, req integrations.BroadcastTransferRequest) (integrations.BroadcastTransferResult, error) {
	network := strings.ToLower(strings.TrimSpace(req.Network))
	client, err := p.dialClient(ctx, network)
	if err != nil {
		return integrations.BroadcastTransferResult{}, err
	}
	defer client.Close()

	privateKey, signerAddress, err := p.signerForNetwork(req.VaultID, network)
	if err != nil {
		return integrations.BroadcastTransferResult{}, err
	}

	chainID, err := client.ChainID(ctx)
	if err != nil {
		return integrations.BroadcastTransferResult{}, err
	}

	nonce, err := client.PendingNonceAt(ctx, signerAddress)
	if err != nil {
		return integrations.BroadcastTransferResult{}, err
	}

	toAddress := common.HexToAddress(strings.TrimSpace(req.ToAddress))
	if !common.IsHexAddress(req.ToAddress) {
		return integrations.BroadcastTransferResult{}, fmt.Errorf("invalid destination address")
	}

	var (
		txData []byte
		txTo   *common.Address
		value  = big.NewInt(0)
	)

	asset := strings.ToUpper(strings.TrimSpace(req.Asset))
	amount := big.NewInt(req.AmountMinor)
	if amount.Sign() <= 0 {
		return integrations.BroadcastTransferResult{}, fmt.Errorf("amount_minor must be greater than 0")
	}

	isNative, contractAddress, err := p.resolveEVMAsset(ctx, network, asset)
	if err != nil {
		return integrations.BroadcastTransferResult{}, err
	}

	if isNative {
		value = amount
		txTo = &toAddress
	} else {
		erc20ABI, err := abi.JSON(strings.NewReader(`[{"type":"function","name":"transfer","inputs":[{"name":"to","type":"address"},{"name":"value","type":"uint256"}],"outputs":[{"name":"","type":"bool"}],"stateMutability":"nonpayable"}]`))
		if err != nil {
			return integrations.BroadcastTransferResult{}, err
		}
		txData, err = erc20ABI.Pack("transfer", toAddress, amount)
		if err != nil {
			return integrations.BroadcastTransferResult{}, err
		}
		txTo = &contractAddress
	}

	callMsg := ethereum.CallMsg{
		From:  signerAddress,
		To:    txTo,
		Value: value,
		Data:  txData,
	}
	gasLimit, err := client.EstimateGas(ctx, callMsg)
	if err != nil {
		return integrations.BroadcastTransferResult{}, err
	}
	if gasLimit < 21_000 {
		gasLimit = 21_000
	}

	feeCap, tipCap, err := dynamicFeeCaps(ctx, client)
	if err != nil {
		return integrations.BroadcastTransferResult{}, err
	}

	unsignedTx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		GasTipCap: tipCap,
		GasFeeCap: feeCap,
		Gas:       gasLimit,
		To:        txTo,
		Value:     value,
		Data:      txData,
	})

	signedTx, err := types.SignTx(unsignedTx, types.LatestSignerForChainID(chainID), privateKey)
	if err != nil {
		return integrations.BroadcastTransferResult{}, err
	}

	if err := client.SendTransaction(ctx, signedTx); err != nil {
		return integrations.BroadcastTransferResult{}, err
	}

	return integrations.BroadcastTransferResult{TxHash: signedTx.Hash().Hex()}, nil
}

func networkFamily(network string) string {
	key := strings.ToLower(strings.TrimSpace(network))
	switch {
	case strings.HasPrefix(key, "solana"):
		return "solana"
	case strings.HasPrefix(key, "stellar"):
		return "stellar"
	case strings.HasPrefix(key, "xrpl"), strings.HasPrefix(key, "xrp"):
		return "xrpl"
	case strings.HasPrefix(key, "tron"):
		return "tron"
	case strings.HasPrefix(key, "aptos"):
		return "aptos"
	case strings.HasPrefix(key, "sui"):
		return "sui"
	default:
		return "evm"
	}
}

func dynamicFeeCaps(ctx context.Context, client *ethclient.Client) (*big.Int, *big.Int, error) {
	tipCap, err := client.SuggestGasTipCap(ctx)
	if err != nil || tipCap == nil || tipCap.Sign() <= 0 {
		tipCap = big.NewInt(2_000_000_000) // 2 gwei fallback
	}

	header, err := client.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, nil, err
	}

	baseFee := header.BaseFee
	if baseFee == nil || baseFee.Sign() <= 0 {
		gasPrice, priceErr := client.SuggestGasPrice(ctx)
		if priceErr != nil {
			return nil, nil, priceErr
		}
		baseFee = gasPrice
	}

	feeCap := new(big.Int).Mul(baseFee, big.NewInt(2))
	feeCap.Add(feeCap, tipCap)
	return feeCap, tipCap, nil
}

func isNativeAsset(asset string) bool {
	switch strings.ToUpper(strings.TrimSpace(asset)) {
	case "", "ETH", "MATIC", "BNB", "AVAX":
		return true
	default:
		return false
	}
}

func (p *Provider) lookupAsset(ctx context.Context, network, asset string) (AssetDefinition, bool, error) {
	if p.assetLookup == nil || strings.TrimSpace(asset) == "" {
		return AssetDefinition{}, false, nil
	}
	return p.assetLookup(ctx, strings.ToLower(strings.TrimSpace(network)), strings.ToUpper(strings.TrimSpace(asset)))
}

func (p *Provider) resolveEVMAsset(ctx context.Context, network, asset string) (bool, common.Address, error) {
	def, found, err := p.lookupAsset(ctx, network, asset)
	if err != nil {
		return false, common.Address{}, err
	}
	if found {
		if !def.IsActive {
			return false, common.Address{}, fmt.Errorf("asset is disabled for network=%s asset=%s", network, asset)
		}
		if family := strings.ToLower(strings.TrimSpace(def.ChainFamily)); family != "" && family != "evm" {
			return false, common.Address{}, fmt.Errorf("asset network family mismatch for network=%s asset=%s", network, asset)
		}
		if strings.ToLower(strings.TrimSpace(def.AssetType)) == "native" {
			return true, common.Address{}, nil
		}

		contract := strings.TrimSpace(def.ContractAddress)
		if !common.IsHexAddress(contract) {
			return false, common.Address{}, fmt.Errorf("asset contract is not configured for network=%s asset=%s", network, asset)
		}
		return false, common.HexToAddress(contract), nil
	}

	if isNativeAsset(asset) {
		return true, common.Address{}, nil
	}
	contract, ok := p.erc20ContractFor(network, asset)
	if !ok {
		return false, common.Address{}, fmt.Errorf("erc20 contract is not configured for network=%s asset=%s", network, asset)
	}
	return false, contract, nil
}

func (p *Provider) supportsNativeAsset(ctx context.Context, family, network, asset string) (bool, error) {
	def, found, err := p.lookupAsset(ctx, network, asset)
	if err != nil {
		return false, err
	}
	if found {
		if !def.IsActive {
			return false, nil
		}
		if configuredFamily := strings.ToLower(strings.TrimSpace(def.ChainFamily)); configuredFamily != "" && configuredFamily != strings.ToLower(strings.TrimSpace(family)) {
			return false, nil
		}
		return strings.ToLower(strings.TrimSpace(def.AssetType)) == "native", nil
	}

	return isNativeAssetForFamily(family, asset), nil
}

func (p *Provider) signerForNetwork(vaultID, network string) (*ecdsa.PrivateKey, common.Address, error) {
	networkKey := strings.ToLower(strings.TrimSpace(network))
	if networkKey == "" {
		networkKey = p.network
	}

	privateKeyHex := p.scopedPrivateKey(vaultID, networkKey)
	if privateKeyHex == "" {
		return nil, common.Address{}, fmt.Errorf("private key not configured for network %q", networkKey)
	}

	privateKey, err := crypto.HexToECDSA(strings.TrimPrefix(privateKeyHex, "0x"))
	if err != nil {
		return nil, common.Address{}, fmt.Errorf("invalid private key for network %q", networkKey)
	}

	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	return privateKey, address, nil
}

func (p *Provider) scopedPrivateKey(vaultID, network string) string {
	networkKey := strings.ToLower(strings.TrimSpace(network))
	if networkKey == "" {
		networkKey = p.network
	}

	trimmedVaultID := strings.TrimSpace(vaultID)
	if trimmedVaultID != "" {
		if key := strings.TrimSpace(p.vaultPrivateKeys[trimmedVaultID+":"+networkKey]); key != "" {
			return key
		}
	}
	return strings.TrimSpace(p.privateKeys[networkKey])
}

func (p *Provider) erc20ContractFor(network, asset string) (common.Address, bool) {
	key := strings.ToLower(strings.TrimSpace(network)) + ":" + strings.ToUpper(strings.TrimSpace(asset))
	value := strings.TrimSpace(p.erc20Contracts[key])
	if !common.IsHexAddress(value) {
		return common.Address{}, false
	}
	return common.HexToAddress(value), true
}

func (p *Provider) dialClient(ctx context.Context, network string) (*ethclient.Client, error) {
	rpcURL, err := p.rpcURLForNetwork(network)
	if err != nil {
		return nil, err
	}
	client, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (p *Provider) ListDeposits(ctx context.Context, req integrations.ListDepositsRequest) ([]integrations.DetectedDeposit, error) {
	address := strings.ToLower(strings.TrimSpace(req.Address))
	if address == "" {
		return nil, nil
	}
	network := strings.ToLower(strings.TrimSpace(req.Network))

	if p.enableLiveDeposits {
		deposits, err := p.listDepositsLive(ctx, network, address)
		if err == nil {
			return deposits, nil
		}
	}

	return p.listDepositsFallback(req.TenantID, network, address), nil
}

func (p *Provider) VerifyWebhookSignature(body []byte, signature string) bool {
	secret := strings.TrimSpace(p.webhookSigningSecret)
	if secret == "" {
		return true
	}

	signature = strings.TrimSpace(signature)
	if signature == "" {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	provided := strings.TrimPrefix(strings.ToLower(signature), "sha256=")
	return hmac.Equal([]byte(expected), []byte(provided))
}

func (p *Provider) ParseWebhookDeposits(body []byte) ([]integrations.DetectedDeposit, error) {
	var simple struct {
		TxHash        string `json:"tx_hash"`
		AmountMinor   int64  `json:"amount_minor"`
		Confirmations int    `json:"confirmations"`
		Status        string `json:"status"`
	}
	if err := json.Unmarshal(body, &simple); err != nil {
		return nil, err
	}
	if strings.TrimSpace(simple.TxHash) != "" && simple.AmountMinor > 0 {
		return []integrations.DetectedDeposit{{
			TxHash:        strings.ToLower(strings.TrimSpace(simple.TxHash)),
			AmountMinor:   simple.AmountMinor,
			Confirmations: clampInt(simple.Confirmations, 0),
			Status:        deriveDepositStatus(simple.Status, simple.Confirmations),
		}}, nil
	}

	var payload struct {
		Event struct {
			Activity []struct {
				Hash             string `json:"hash"`
				Value            any    `json:"value"`
				NumConfirmations int    `json:"numConfirmations"`
				Confirmations    int    `json:"confirmations"`
				Status           string `json:"status"`
				RawContract      struct {
					Value string `json:"value"`
				} `json:"rawContract"`
			} `json:"activity"`
		} `json:"event"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	deposits := make([]integrations.DetectedDeposit, 0, len(payload.Event.Activity))
	for _, activity := range payload.Event.Activity {
		txHash := strings.ToLower(strings.TrimSpace(activity.Hash))
		if txHash == "" {
			continue
		}

		amountMinor := parseAmountMinor(activity.Value, activity.RawContract.Value)
		if amountMinor <= 0 {
			continue
		}

		confirmations := activity.NumConfirmations
		if confirmations == 0 {
			confirmations = activity.Confirmations
		}
		if confirmations < 0 {
			confirmations = 0
		}

		deposits = append(deposits, integrations.DetectedDeposit{
			TxHash:        txHash,
			AmountMinor:   amountMinor,
			Confirmations: confirmations,
			Status:        deriveDepositStatus(activity.Status, confirmations),
		})
	}

	if len(deposits) == 0 {
		return nil, fmt.Errorf("no deposit activity found in webhook payload")
	}
	return deposits, nil
}

func (p *Provider) listDepositsLive(ctx context.Context, network, toAddress string) ([]integrations.DetectedDeposit, error) {
	latestBlock, err := p.ethBlockNumber(ctx, network)
	if err != nil {
		return nil, err
	}

	type transferResponse struct {
		Result struct {
			Transfers []struct {
				Hash        string `json:"hash"`
				Value       any    `json:"value"`
				BlockNum    string `json:"blockNum"`
				RawContract struct {
					Value string `json:"value"`
				} `json:"rawContract"`
			} `json:"transfers"`
		} `json:"result"`
	}

	var resp transferResponse
	params := []any{map[string]any{
		"fromBlock":        "0x0",
		"toAddress":        toAddress,
		"category":         []string{"external", "internal", "erc20", "erc721", "erc1155"},
		"excludeZeroValue": true,
		"withMetadata":     false,
		"maxCount":         "0x64",
	}}
	if err := p.rpcCall(ctx, network, "alchemy_getAssetTransfers", params, &resp); err != nil {
		return nil, err
	}

	deposits := make([]integrations.DetectedDeposit, 0, len(resp.Result.Transfers))
	for _, transfer := range resp.Result.Transfers {
		amountMinor := parseAmountMinor(transfer.Value, transfer.RawContract.Value)
		if amountMinor <= 0 {
			continue
		}

		blockNumber := parseHexInt64(transfer.BlockNum)
		confirmations := int(latestBlock - blockNumber + 1)
		if blockNumber <= 0 {
			confirmations = 0
		}
		status := "pending"
		if confirmations >= p.confirmationsRequired {
			status = "confirmed"
		}

		deposits = append(deposits, integrations.DetectedDeposit{
			TxHash:        strings.ToLower(strings.TrimSpace(transfer.Hash)),
			AmountMinor:   amountMinor,
			Confirmations: confirmations,
			Status:        status,
		})
	}

	return deposits, nil
}

func (p *Provider) listDepositsFallback(tenantID, network, address string) []integrations.DetectedDeposit {
	p.mu.Lock()
	defer p.mu.Unlock()

	key := network + ":" + address
	if p.depositsEmitted[key] {
		return nil
	}

	h := sha256.Sum256([]byte(tenantID + ":" + network + ":" + address + ":deposit"))
	txHash := "0x" + hex.EncodeToString(h[:])
	p.depositsEmitted[key] = true

	return []integrations.DetectedDeposit{{
		TxHash:        txHash,
		AmountMinor:   250000,
		Confirmations: p.confirmationsRequired,
		Status:        "confirmed",
	}}
}

func (p *Provider) ethBlockNumber(ctx context.Context, network string) (int64, error) {
	var resp struct {
		Result string `json:"result"`
	}
	if err := p.rpcCall(ctx, network, "eth_blockNumber", []any{}, &resp); err != nil {
		return 0, err
	}
	value := parseHexInt64(resp.Result)
	if value < 0 {
		return 0, fmt.Errorf("invalid block number response")
	}
	return value, nil
}

func (p *Provider) rpcCall(ctx context.Context, network, method string, params []any, out any) error {
	rpcURL, err := p.rpcURLForNetwork(network)
	if err != nil {
		return err
	}

	payload := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  method,
		"params":  params,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rpcURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		bodySnippet := strings.TrimSpace(string(respBody))
		bodySnippet = strings.ReplaceAll(bodySnippet, "\n", " ")
		bodySnippet = strings.ReplaceAll(bodySnippet, "\t", " ")
		if len(bodySnippet) > 180 {
			bodySnippet = bodySnippet[:180]
		}
		if bodySnippet == "" {
			return fmt.Errorf("alchemy rpc status %d", resp.StatusCode)
		}
		return fmt.Errorf("alchemy rpc status %d: %s", resp.StatusCode, bodySnippet)
	}

	var envelope struct {
		Error *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(respBody, &envelope); err == nil && envelope.Error != nil {
		return fmt.Errorf("alchemy rpc error %d: %s", envelope.Error.Code, envelope.Error.Message)
	}

	if err := json.Unmarshal(respBody, out); err != nil {
		return err
	}
	return nil
}

func (p *Provider) rpcURLForNetwork(network string) (string, error) {
	requested := strings.ToLower(strings.TrimSpace(network))
	if requested == "" {
		requested = p.network
	}
	if requested == "" {
		return "", fmt.Errorf("alchemy network not configured")
	}

	rpcURL := strings.TrimSpace(p.rpcURLs[requested])
	if rpcURL == "" {
		return "", fmt.Errorf("alchemy rpc url not configured for network %q", requested)
	}
	return rpcURL, nil
}

func parseHexInt64(raw string) int64 {
	raw = strings.TrimSpace(strings.ToLower(raw))
	raw = strings.TrimPrefix(raw, "0x")
	if raw == "" {
		return 0
	}
	value, err := strconv.ParseInt(raw, 16, 64)
	if err != nil {
		return -1
	}
	return value
}

func parseAmountMinor(value any, rawHexValue string) int64 {
	if strings.TrimSpace(rawHexValue) != "" {
		raw := parseHexInt64(rawHexValue)
		if raw > 0 {
			if raw > math.MaxInt64 {
				return math.MaxInt64
			}
			return raw
		}
	}

	switch v := value.(type) {
	case float64:
		if v <= 0 {
			return 0
		}
		return int64(math.Round(v * 1_000_000))
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
		if err != nil || f <= 0 {
			return 0
		}
		return int64(math.Round(f * 1_000_000))
	default:
		return 0
	}
}

func deriveDepositStatus(raw string, confirmations int) string {
	status := strings.ToLower(strings.TrimSpace(raw))
	if status != "" {
		return status
	}
	if confirmations > 0 {
		return "confirmed"
	}
	return "pending"
}

func clampInt(value, floor int) int {
	if value < floor {
		return floor
	}
	return value
}
