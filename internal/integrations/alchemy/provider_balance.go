package alchemy

import (
	"context"
	"fmt"
	"math/big"
	"net/url"
	"strconv"
	"strings"

	"github.com/ebitezion/Nduracore/internal/integrations"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gagliardetto/solana-go"
	"github.com/stellar/go/keypair"
)

func (p *Provider) GetBalance(ctx context.Context, req integrations.GetBalanceRequest) (integrations.GetBalanceResult, error) {
	network := strings.ToLower(strings.TrimSpace(req.Network))
	asset := strings.ToUpper(strings.TrimSpace(req.Asset))
	address := strings.TrimSpace(req.Address)

	if network == "" {
		return integrations.GetBalanceResult{}, fmt.Errorf("network is required")
	}
	if address == "" {
		return integrations.GetBalanceResult{}, fmt.Errorf("address is required")
	}

	switch networkFamily(network) {
	case "solana":
		return p.getBalanceSolana(ctx, network, address, asset)
	case "stellar":
		return p.getBalanceStellar(ctx, network, address, asset)
	case "xrpl":
		return p.getBalanceXRPL(ctx, network, address, asset)
	case "tron":
		return p.getBalanceTron(ctx, network, address, asset)
	case "aptos":
		return p.getBalanceAptos(ctx, network, address, asset)
	case "sui":
		return p.getBalanceSui(ctx, network, address, asset)
	default:
		return p.getBalanceEVM(ctx, network, address, asset)
	}
}

func (p *Provider) getBalanceEVM(ctx context.Context, network, address, asset string) (integrations.GetBalanceResult, error) {
	if !common.IsHexAddress(address) {
		return integrations.GetBalanceResult{}, fmt.Errorf("invalid wallet address")
	}

	normalizedAddress := common.HexToAddress(address).Hex()
	isNative, contractAddress, err := p.resolveEVMAsset(ctx, network, asset)
	if err != nil {
		return integrations.GetBalanceResult{}, err
	}

	if isNative {
		var resp struct {
			Result string `json:"result"`
		}
		if err := p.rpcCall(ctx, network, "eth_getBalance", []any{normalizedAddress, "latest"}, &resp); err != nil {
			return integrations.GetBalanceResult{}, err
		}
		balance, err := hexToBigInt(resp.Result)
		if err != nil {
			return integrations.GetBalanceResult{}, fmt.Errorf("invalid eth_getBalance response")
		}
		return integrations.GetBalanceResult{BalanceMinor: balance.String()}, nil
	}

	callData := evmBalanceOfCallData(normalizedAddress)
	var resp struct {
		Result string `json:"result"`
	}
	if err := p.rpcCall(ctx, network, "eth_call", []any{
		map[string]any{
			"to":   contractAddress.Hex(),
			"data": callData,
		},
		"latest",
	}, &resp); err != nil {
		return integrations.GetBalanceResult{}, err
	}

	balance, err := hexToBigInt(resp.Result)
	if err != nil {
		return integrations.GetBalanceResult{}, fmt.Errorf("invalid erc20 balance response")
	}
	return integrations.GetBalanceResult{BalanceMinor: balance.String()}, nil
}

func (p *Provider) getBalanceSolana(ctx context.Context, network, address, asset string) (integrations.GetBalanceResult, error) {
	native, err := p.supportsNativeAsset(ctx, "solana", network, asset)
	if err != nil {
		return integrations.GetBalanceResult{}, err
	}
	if !native {
		return integrations.GetBalanceResult{}, fmt.Errorf("solana balance currently supports native SOL only")
	}
	if _, err := solana.PublicKeyFromBase58(address); err != nil {
		return integrations.GetBalanceResult{}, fmt.Errorf("invalid wallet address")
	}

	var resp struct {
		Result struct {
			Value int64 `json:"value"`
		} `json:"result"`
	}
	if err := p.rpcCall(ctx, network, "getBalance", []any{
		address,
		map[string]any{"commitment": "confirmed"},
	}, &resp); err != nil {
		return integrations.GetBalanceResult{}, err
	}
	if resp.Result.Value < 0 {
		return integrations.GetBalanceResult{}, fmt.Errorf("invalid getBalance response")
	}
	return integrations.GetBalanceResult{BalanceMinor: strconv.FormatInt(resp.Result.Value, 10)}, nil
}

func (p *Provider) getBalanceStellar(ctx context.Context, network, address, asset string) (integrations.GetBalanceResult, error) {
	native, err := p.supportsNativeAsset(ctx, "stellar", network, asset)
	if err != nil {
		return integrations.GetBalanceResult{}, err
	}
	if !native {
		return integrations.GetBalanceResult{}, fmt.Errorf("stellar balance currently supports native XLM only")
	}
	if _, err := keypair.ParseAddress(address); err != nil {
		return integrations.GetBalanceResult{}, fmt.Errorf("invalid wallet address")
	}

	rpcURL, err := p.rpcURLForNetwork(network)
	if err != nil {
		return integrations.GetBalanceResult{}, err
	}

	var resp struct {
		Balances []struct {
			AssetType string `json:"asset_type"`
			Balance   string `json:"balance"`
		} `json:"balances"`
	}
	if err := p.httpGetJSON(ctx, strings.TrimRight(rpcURL, "/")+"/accounts/"+address, &resp); err != nil {
		return integrations.GetBalanceResult{}, err
	}
	for _, bal := range resp.Balances {
		if strings.EqualFold(strings.TrimSpace(bal.AssetType), "native") {
			minor, err := decimalToMinorString(bal.Balance, 7)
			if err != nil {
				return integrations.GetBalanceResult{}, fmt.Errorf("invalid stellar balance payload")
			}
			return integrations.GetBalanceResult{BalanceMinor: minor}, nil
		}
	}
	return integrations.GetBalanceResult{BalanceMinor: "0"}, nil
}

func (p *Provider) getBalanceXRPL(ctx context.Context, network, address, asset string) (integrations.GetBalanceResult, error) {
	native, err := p.supportsNativeAsset(ctx, "xrpl", network, asset)
	if err != nil {
		return integrations.GetBalanceResult{}, err
	}
	if !native {
		return integrations.GetBalanceResult{}, fmt.Errorf("xrpl balance currently supports native XRP only")
	}

	var resp struct {
		Result struct {
			AccountData struct {
				Balance string `json:"Balance"`
			} `json:"account_data"`
		} `json:"result"`
	}
	if err := p.rpcCall(ctx, network, "account_info", []any{
		map[string]any{
			"account":      address,
			"ledger_index": "validated",
		},
	}, &resp); err != nil {
		return integrations.GetBalanceResult{}, err
	}
	if strings.TrimSpace(resp.Result.AccountData.Balance) == "" {
		return integrations.GetBalanceResult{}, fmt.Errorf("invalid xrpl account_info response")
	}
	return integrations.GetBalanceResult{BalanceMinor: strings.TrimSpace(resp.Result.AccountData.Balance)}, nil
}

func (p *Provider) getBalanceTron(ctx context.Context, network, address, asset string) (integrations.GetBalanceResult, error) {
	native, err := p.supportsNativeAsset(ctx, "tron", network, asset)
	if err != nil {
		return integrations.GetBalanceResult{}, err
	}
	if !native {
		return integrations.GetBalanceResult{}, fmt.Errorf("tron balance currently supports native TRX only")
	}

	rpcURL, err := p.rpcURLForNetwork(network)
	if err != nil {
		return integrations.GetBalanceResult{}, err
	}
	baseURL := strings.TrimRight(rpcURL, "/")
	if strings.Contains(baseURL, "/wallet/") {
		baseURL = baseURL[:strings.Index(baseURL, "/wallet/")]
	}

	var resp struct {
		Balance int64 `json:"balance"`
	}
	if err := p.httpPostJSON(ctx, baseURL+"/wallet/getaccount", map[string]any{
		"address": address,
		"visible": true,
	}, &resp); err != nil {
		return integrations.GetBalanceResult{}, err
	}
	if resp.Balance < 0 {
		return integrations.GetBalanceResult{}, fmt.Errorf("invalid tron getaccount response")
	}
	return integrations.GetBalanceResult{BalanceMinor: strconv.FormatInt(resp.Balance, 10)}, nil
}

func (p *Provider) getBalanceAptos(ctx context.Context, network, address, asset string) (integrations.GetBalanceResult, error) {
	native, err := p.supportsNativeAsset(ctx, "aptos", network, asset)
	if err != nil {
		return integrations.GetBalanceResult{}, err
	}
	if !native {
		return integrations.GetBalanceResult{}, fmt.Errorf("aptos balance currently supports native APT only")
	}

	normalized := normalizeHexAddress(address)
	if normalized == "" {
		return integrations.GetBalanceResult{}, fmt.Errorf("invalid wallet address")
	}

	rpcURL, err := p.rpcURLForNetwork(network)
	if err != nil {
		return integrations.GetBalanceResult{}, err
	}
	resourceType := url.PathEscape("0x1::coin::CoinStore<0x1::aptos_coin::AptosCoin>")

	var resp struct {
		Data struct {
			Coin struct {
				Value string `json:"value"`
			} `json:"coin"`
		} `json:"data"`
	}
	if err := p.httpGetJSON(ctx, strings.TrimRight(rpcURL, "/")+"/v1/accounts/"+normalized+"/resource/"+resourceType, &resp); err != nil {
		return integrations.GetBalanceResult{}, err
	}
	if strings.TrimSpace(resp.Data.Coin.Value) == "" {
		return integrations.GetBalanceResult{}, fmt.Errorf("invalid aptos coin store response")
	}
	return integrations.GetBalanceResult{BalanceMinor: strings.TrimSpace(resp.Data.Coin.Value)}, nil
}

func (p *Provider) getBalanceSui(ctx context.Context, network, address, asset string) (integrations.GetBalanceResult, error) {
	native, err := p.supportsNativeAsset(ctx, "sui", network, asset)
	if err != nil {
		return integrations.GetBalanceResult{}, err
	}
	if !native {
		return integrations.GetBalanceResult{}, fmt.Errorf("sui balance currently supports native SUI only")
	}
	if normalizeHexAddress(address) == "" {
		return integrations.GetBalanceResult{}, fmt.Errorf("invalid wallet address")
	}

	var resp struct {
		Result struct {
			TotalBalance string `json:"totalBalance"`
		} `json:"result"`
	}
	if err := p.rpcCall(ctx, network, "suix_getBalance", []any{
		address,
		"0x2::sui::SUI",
	}, &resp); err != nil {
		return integrations.GetBalanceResult{}, err
	}
	if strings.TrimSpace(resp.Result.TotalBalance) == "" {
		return integrations.GetBalanceResult{}, fmt.Errorf("invalid sui balance response")
	}
	return integrations.GetBalanceResult{BalanceMinor: strings.TrimSpace(resp.Result.TotalBalance)}, nil
}

func evmBalanceOfCallData(address string) string {
	trimmed := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(address)), "0x")
	return "0x70a08231" + strings.Repeat("0", 24) + trimmed
}

func hexToBigInt(value string) (*big.Int, error) {
	trimmed := strings.TrimSpace(strings.TrimPrefix(strings.ToLower(value), "0x"))
	if trimmed == "" {
		return big.NewInt(0), nil
	}
	balance := new(big.Int)
	if _, ok := balance.SetString(trimmed, 16); !ok {
		return nil, fmt.Errorf("invalid hex value")
	}
	return balance, nil
}

func decimalToMinorString(value string, decimals int) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", fmt.Errorf("empty decimal")
	}
	if decimals < 0 {
		return "", fmt.Errorf("invalid decimals")
	}

	parts := strings.SplitN(trimmed, ".", 2)
	whole := strings.TrimSpace(parts[0])
	fraction := ""
	if len(parts) == 2 {
		fraction = strings.TrimSpace(parts[1])
	}
	if whole == "" {
		whole = "0"
	}
	whole = strings.TrimPrefix(whole, "+")
	if strings.HasPrefix(whole, "-") {
		return "", fmt.Errorf("negative values are unsupported")
	}
	for _, r := range whole {
		if r < '0' || r > '9' {
			return "", fmt.Errorf("invalid decimal whole")
		}
	}
	for _, r := range fraction {
		if r < '0' || r > '9' {
			return "", fmt.Errorf("invalid decimal fraction")
		}
	}
	if len(fraction) > decimals {
		fraction = fraction[:decimals]
	}
	for len(fraction) < decimals {
		fraction += "0"
	}

	result := strings.TrimLeft(whole+fraction, "0")
	if result == "" {
		return "0", nil
	}
	return result, nil
}
