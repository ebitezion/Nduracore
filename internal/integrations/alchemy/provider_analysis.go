package alchemy

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

type GasEstimate struct {
	Network         string `json:"network"`
	Asset           string `json:"asset"`
	ToAddress       string `json:"to_address"`
	AmountMinor     int64  `json:"amount_minor"`
	GasLimit        uint64 `json:"gas_limit"`
	GasTipCapWei    string `json:"gas_tip_cap_wei"`
	GasFeeCapWei    string `json:"gas_fee_cap_wei"`
	EstimatedFeeWei string `json:"estimated_fee_wei"`
}

type NonceState struct {
	Address      string `json:"address"`
	Network      string `json:"network"`
	LatestNonce  uint64 `json:"latest_nonce"`
	PendingNonce uint64 `json:"pending_nonce"`
}

type NetworkState struct {
	Network     string `json:"network"`
	ChainFamily string `json:"chain_family"`
	Healthy     bool   `json:"healthy"`
	LatestBlock string `json:"latest_block,omitempty"`
	Detail      string `json:"detail,omitempty"`
}

type TxReceiptState struct {
	TxHash       string `json:"tx_hash"`
	Found        bool   `json:"found"`
	Confirmed    bool   `json:"confirmed"`
	Success      bool   `json:"success"`
	BlockNumber  string `json:"block_number,omitempty"`
	BlockHash    string `json:"block_hash,omitempty"`
	Confirmations int64 `json:"confirmations"`
}

func (p *Provider) EstimateTransferGas(ctx context.Context, network, asset, fromAddress, toAddress string, amountMinor int64) (GasEstimate, error) {
	if amountMinor <= 0 {
		return GasEstimate{}, fmt.Errorf("amount_minor must be greater than 0")
	}
	if networkFamily(network) != "evm" {
		return GasEstimate{}, fmt.Errorf("gas estimate currently supports evm networks only")
	}

	client, err := p.dialClient(ctx, network)
	if err != nil {
		return GasEstimate{}, err
	}
	defer client.Close()

	var (
		txData []byte
		txTo   *common.Address
		value  = big.NewInt(0)
	)

	to := common.HexToAddress(strings.TrimSpace(toAddress))
	if !common.IsHexAddress(toAddress) {
		return GasEstimate{}, fmt.Errorf("invalid destination address")
	}

	amount := big.NewInt(amountMinor)
	isNative, contractAddress, err := p.resolveEVMAsset(ctx, strings.ToLower(strings.TrimSpace(network)), strings.ToUpper(strings.TrimSpace(asset)))
	if err != nil {
		return GasEstimate{}, err
	}

	if isNative {
		value = amount
		txTo = &to
	} else {
		erc20ABI, err := abi.JSON(strings.NewReader(`[{"type":"function","name":"transfer","inputs":[{"name":"to","type":"address"},{"name":"value","type":"uint256"}],"outputs":[{"name":"","type":"bool"}],"stateMutability":"nonpayable"}]`))
		if err != nil {
			return GasEstimate{}, err
		}
		txData, err = erc20ABI.Pack("transfer", to, amount)
		if err != nil {
			return GasEstimate{}, err
		}
		txTo = &contractAddress
	}

	callMsg := ethereum.CallMsg{
		From:  common.HexToAddress(strings.TrimSpace(fromAddress)),
		To:    txTo,
		Value: value,
		Data:  txData,
	}
	gasLimit, err := client.EstimateGas(ctx, callMsg)
	if err != nil {
		return GasEstimate{}, err
	}

	feeCap, tipCap, err := dynamicFeeCaps(ctx, client)
	if err != nil {
		return GasEstimate{}, err
	}
	estimatedFee := new(big.Int).Mul(new(big.Int).SetUint64(gasLimit), feeCap)

	return GasEstimate{
		Network:         strings.ToLower(strings.TrimSpace(network)),
		Asset:           strings.ToUpper(strings.TrimSpace(asset)),
		ToAddress:       strings.ToLower(strings.TrimSpace(toAddress)),
		AmountMinor:     amountMinor,
		GasLimit:        gasLimit,
		GasTipCapWei:    tipCap.String(),
		GasFeeCapWei:    feeCap.String(),
		EstimatedFeeWei: estimatedFee.String(),
	}, nil
}

func (p *Provider) GetTokenAllowance(ctx context.Context, network, asset, owner, spender string) (string, error) {
	if networkFamily(network) != "evm" {
		return "", fmt.Errorf("token allowance currently supports evm networks only")
	}
	if !common.IsHexAddress(owner) || !common.IsHexAddress(spender) {
		return "", fmt.Errorf("owner and spender must be valid hex addresses")
	}
	isNative, contractAddress, err := p.resolveEVMAsset(ctx, network, asset)
	if err != nil {
		return "", err
	}
	if isNative {
		return "", fmt.Errorf("allowance is only applicable to token assets")
	}

	data := "0xdd62ed3e" + strings.Repeat("0", 24) + strings.TrimPrefix(strings.ToLower(owner), "0x") +
		strings.Repeat("0", 24) + strings.TrimPrefix(strings.ToLower(spender), "0x")
	var resp struct {
		Result string `json:"result"`
	}
	if err := p.rpcCall(ctx, network, "eth_call", []any{
		map[string]any{
			"to":   contractAddress.Hex(),
			"data": data,
		},
		"latest",
	}, &resp); err != nil {
		return "", err
	}
	value, err := hexToBigInt(resp.Result)
	if err != nil {
		return "", fmt.Errorf("invalid allowance response")
	}
	return value.String(), nil
}

func (p *Provider) GetNonceState(ctx context.Context, network, address string) (NonceState, error) {
	if networkFamily(network) != "evm" {
		return NonceState{}, fmt.Errorf("nonce state currently supports evm networks only")
	}
	if !common.IsHexAddress(address) {
		return NonceState{}, fmt.Errorf("invalid wallet address")
	}
	var latest struct {
		Result string `json:"result"`
	}
	if err := p.rpcCall(ctx, network, "eth_getTransactionCount", []any{common.HexToAddress(address).Hex(), "latest"}, &latest); err != nil {
		return NonceState{}, err
	}
	var pending struct {
		Result string `json:"result"`
	}
	if err := p.rpcCall(ctx, network, "eth_getTransactionCount", []any{common.HexToAddress(address).Hex(), "pending"}, &pending); err != nil {
		return NonceState{}, err
	}

	latestValue := parseHexInt64(latest.Result)
	pendingValue := parseHexInt64(pending.Result)
	if latestValue < 0 || pendingValue < 0 {
		return NonceState{}, fmt.Errorf("invalid nonce response")
	}

	return NonceState{
		Address:      strings.ToLower(common.HexToAddress(address).Hex()),
		Network:      strings.ToLower(strings.TrimSpace(network)),
		LatestNonce:  uint64(latestValue),
		PendingNonce: uint64(pendingValue),
	}, nil
}

func (p *Provider) GetNetworkState(ctx context.Context, network string) (NetworkState, error) {
	network = strings.ToLower(strings.TrimSpace(network))
	switch networkFamily(network) {
	case "evm":
		var resp struct {
			Result string `json:"result"`
		}
		if err := p.rpcCall(ctx, network, "eth_blockNumber", []any{}, &resp); err != nil {
			return NetworkState{Network: network, ChainFamily: "evm", Healthy: false, Detail: err.Error()}, nil
		}
		return NetworkState{Network: network, ChainFamily: "evm", Healthy: true, LatestBlock: resp.Result}, nil
	case "solana":
		var resp struct {
			Result string `json:"result"`
		}
		if err := p.rpcCall(ctx, network, "getHealth", []any{}, &resp); err != nil {
			return NetworkState{Network: network, ChainFamily: "solana", Healthy: false, Detail: err.Error()}, nil
		}
		return NetworkState{Network: network, ChainFamily: "solana", Healthy: true, Detail: resp.Result}, nil
	case "xrpl":
		var resp struct {
			Result map[string]any `json:"result"`
		}
		if err := p.rpcCall(ctx, network, "server_info", []any{map[string]any{}}, &resp); err != nil {
			return NetworkState{Network: network, ChainFamily: "xrpl", Healthy: false, Detail: err.Error()}, nil
		}
		return NetworkState{Network: network, ChainFamily: "xrpl", Healthy: true}, nil
	case "sui":
		var resp struct {
			Result map[string]any `json:"result"`
		}
		if err := p.rpcCall(ctx, network, "suix_getLatestSuiSystemState", []any{}, &resp); err != nil {
			return NetworkState{Network: network, ChainFamily: "sui", Healthy: false, Detail: err.Error()}, nil
		}
		return NetworkState{Network: network, ChainFamily: "sui", Healthy: true}, nil
	default:
		return NetworkState{Network: network, ChainFamily: networkFamily(network), Healthy: false, Detail: "network status not yet implemented for this chain family"}, nil
	}
}

func (p *Provider) GetTxReceiptState(ctx context.Context, network, txHash string) (TxReceiptState, error) {
	if networkFamily(network) != "evm" {
		return TxReceiptState{}, fmt.Errorf("transaction trace currently supports evm networks only")
	}
	var receipt struct {
		Result *struct {
			BlockHash   string `json:"blockHash"`
			BlockNumber string `json:"blockNumber"`
			Status      string `json:"status"`
		} `json:"result"`
	}
	if err := p.rpcCall(ctx, network, "eth_getTransactionReceipt", []any{strings.TrimSpace(txHash)}, &receipt); err != nil {
		return TxReceiptState{}, err
	}
	if receipt.Result == nil {
		return TxReceiptState{TxHash: strings.TrimSpace(txHash), Found: false, Confirmed: false, Success: false}, nil
	}

	success := strings.TrimSpace(receipt.Result.Status) == "0x1"
	blockNumber := parseHexInt64(receipt.Result.BlockNumber)
	confirmations := int64(0)
	if blockNumber > 0 {
		var latest struct {
			Result string `json:"result"`
		}
		if err := p.rpcCall(ctx, network, "eth_blockNumber", []any{}, &latest); err == nil {
			latestBlock := parseHexInt64(latest.Result)
			if latestBlock >= blockNumber {
				confirmations = latestBlock - blockNumber + 1
			}
		}
	}

	return TxReceiptState{
		TxHash:        strings.TrimSpace(txHash),
		Found:         true,
		Confirmed:     true,
		Success:       success,
		BlockNumber:   receipt.Result.BlockNumber,
		BlockHash:     receipt.Result.BlockHash,
		Confirmations: confirmations,
	}, nil
}
