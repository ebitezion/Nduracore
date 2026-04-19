package alchemy

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/block-vision/sui-go-sdk/models"
	suisigner "github.com/block-vision/sui-go-sdk/signer"
	"github.com/block-vision/sui-go-sdk/sui"
	"github.com/ebitezion/Nduracore/internal/integrations"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/mr-tron/base58"
	"golang.org/x/crypto/sha3"
)

func (p *Provider) tronSignerForNetwork(vaultID, network string) (*ecdsa.PrivateKey, string, error) {
	networkKey := strings.ToLower(strings.TrimSpace(network))
	if networkKey == "" {
		networkKey = p.network
	}

	privateKeyHex := strings.TrimSpace(strings.TrimPrefix(p.scopedPrivateKey(vaultID, networkKey), "0x"))
	if privateKeyHex == "" {
		return nil, "", fmt.Errorf("private key not configured for network %q", networkKey)
	}

	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, "", fmt.Errorf("invalid tron private key for network %q", networkKey)
	}

	addr := tronAddressFromECDSA(privateKey)
	return privateKey, addr, nil
}

func tronAddressFromECDSA(privateKey *ecdsa.PrivateKey) string {
	evmAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
	tronPayload := append([]byte{0x41}, evmAddress.Bytes()...)
	checksum := doubleSHA256(tronPayload)[:4]
	return base58.Encode(append(tronPayload, checksum...))
}

func doubleSHA256(input []byte) []byte {
	first := sha256.Sum256(input)
	second := sha256.Sum256(first[:])
	return second[:]
}

func (p *Provider) tronAddressForNetwork(vaultID, network string) (string, error) {
	_, addr, err := p.tronSignerForNetwork(vaultID, network)
	if err != nil {
		return "", err
	}
	return addr, nil
}

func (p *Provider) broadcastTransferTron(ctx context.Context, req integrations.BroadcastTransferRequest) (integrations.BroadcastTransferResult, error) {
	networkName := strings.ToLower(strings.TrimSpace(req.Network))
	native, err := p.supportsNativeAsset(ctx, "tron", networkName, req.Asset)
	if err != nil {
		return integrations.BroadcastTransferResult{}, err
	}
	if !native {
		return integrations.BroadcastTransferResult{}, fmt.Errorf("tron broadcast currently supports native TRX only")
	}
	if req.AmountMinor <= 0 {
		return integrations.BroadcastTransferResult{}, fmt.Errorf("amount_minor must be greater than 0")
	}

	privateKey, signerAddress, err := p.tronSignerForNetwork(req.VaultID, networkName)
	if err != nil {
		return integrations.BroadcastTransferResult{}, err
	}

	rpcURL, err := p.rpcURLForNetwork(networkName)
	if err != nil {
		return integrations.BroadcastTransferResult{}, err
	}
	baseURL := strings.TrimRight(rpcURL, "/")
	if strings.Contains(baseURL, "/wallet/") {
		baseURL = baseURL[:strings.Index(baseURL, "/wallet/")]
	}

	createPayload := map[string]any{
		"owner_address": signerAddress,
		"to_address":    strings.TrimSpace(req.ToAddress),
		"amount":        req.AmountMinor,
		"visible":       true,
	}

	var createResp struct {
		TxID       string   `json:"txID"`
		Signature  []string `json:"signature"`
		RawDataHex string   `json:"raw_data_hex"`
		Code       string   `json:"code"`
		Message    string   `json:"message"`
	}
	if err := p.httpPostJSON(ctx, baseURL+"/wallet/createtransaction", createPayload, &createResp); err != nil {
		return integrations.BroadcastTransferResult{}, err
	}
	if strings.TrimSpace(createResp.Code) != "" {
		return integrations.BroadcastTransferResult{}, fmt.Errorf("tron create tx failed: %s", strings.TrimSpace(createResp.Message))
	}
	if strings.TrimSpace(createResp.TxID) == "" {
		return integrations.BroadcastTransferResult{}, fmt.Errorf("tron create tx missing txID")
	}

	txIDBytes, err := hex.DecodeString(strings.TrimSpace(createResp.TxID))
	if err != nil {
		return integrations.BroadcastTransferResult{}, fmt.Errorf("invalid tron txID from rpc")
	}
	sig, err := crypto.Sign(txIDBytes, privateKey)
	if err != nil {
		return integrations.BroadcastTransferResult{}, err
	}

	broadcastPayload := map[string]any{
		"txID":         createResp.TxID,
		"raw_data_hex": createResp.RawDataHex,
		"signature":    []string{hex.EncodeToString(sig)},
		"visible":      true,
	}
	var broadcastResp struct {
		Result  bool   `json:"result"`
		TxID    string `json:"txid"`
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	if err := p.httpPostJSON(ctx, baseURL+"/wallet/broadcasttransaction", broadcastPayload, &broadcastResp); err != nil {
		return integrations.BroadcastTransferResult{}, err
	}
	if !broadcastResp.Result {
		if strings.TrimSpace(broadcastResp.Message) != "" {
			return integrations.BroadcastTransferResult{}, fmt.Errorf("tron broadcast failed: %s", strings.TrimSpace(broadcastResp.Message))
		}
		if strings.TrimSpace(broadcastResp.Code) != "" {
			return integrations.BroadcastTransferResult{}, fmt.Errorf("tron broadcast failed: %s", strings.TrimSpace(broadcastResp.Code))
		}
		return integrations.BroadcastTransferResult{}, fmt.Errorf("tron broadcast failed")
	}

	txHash := strings.TrimSpace(broadcastResp.TxID)
	if txHash == "" {
		txHash = strings.TrimSpace(createResp.TxID)
	}
	return integrations.BroadcastTransferResult{TxHash: txHash}, nil
}

func (p *Provider) aptosAddressForNetwork(vaultID, network string) (string, error) {
	networkKey := strings.ToLower(strings.TrimSpace(network))
	if networkKey == "" {
		networkKey = p.network
	}

	privateKeyRaw := p.scopedPrivateKey(vaultID, networkKey)
	if privateKeyRaw == "" {
		return "", fmt.Errorf("private key not configured for network %q", networkKey)
	}

	privateKey, publicKey, err := parseAptosPrivateKey(privateKeyRaw)
	if err != nil {
		return "", fmt.Errorf("invalid aptos private key for network %q", networkKey)
	}
	_ = privateKey
	return aptosAddressFromPublicKey(publicKey), nil
}

func parseAptosPrivateKey(raw string) (ed25519.PrivateKey, ed25519.PublicKey, error) {
	trimmed := strings.TrimSpace(strings.TrimPrefix(raw, "0x"))
	if trimmed == "" {
		return nil, nil, fmt.Errorf("empty private key")
	}

	decoded, err := hex.DecodeString(trimmed)
	if err != nil {
		return nil, nil, err
	}

	switch len(decoded) {
	case ed25519.SeedSize:
		privateKey := ed25519.NewKeyFromSeed(decoded)
		return privateKey, privateKey.Public().(ed25519.PublicKey), nil
	case ed25519.PrivateKeySize:
		privateKey := ed25519.PrivateKey(decoded)
		return privateKey, privateKey.Public().(ed25519.PublicKey), nil
	default:
		return nil, nil, fmt.Errorf("unexpected aptos key length")
	}
}

func aptosAddressFromPublicKey(publicKey ed25519.PublicKey) string {
	input := make([]byte, 0, len(publicKey)+1)
	input = append(input, publicKey...)
	input = append(input, byte(0x00))
	hash := sha3.Sum256(input)
	return "0x" + hex.EncodeToString(hash[:])
}

func (p *Provider) broadcastTransferAptos(ctx context.Context, req integrations.BroadcastTransferRequest) (integrations.BroadcastTransferResult, error) {
	networkName := strings.ToLower(strings.TrimSpace(req.Network))
	native, err := p.supportsNativeAsset(ctx, "aptos", networkName, req.Asset)
	if err != nil {
		return integrations.BroadcastTransferResult{}, err
	}
	if !native {
		return integrations.BroadcastTransferResult{}, fmt.Errorf("aptos broadcast currently supports native APT only")
	}
	if req.AmountMinor <= 0 {
		return integrations.BroadcastTransferResult{}, fmt.Errorf("amount_minor must be greater than 0")
	}

	privateKeyRaw := p.scopedPrivateKey(req.VaultID, networkName)
	if privateKeyRaw == "" {
		return integrations.BroadcastTransferResult{}, fmt.Errorf("private key not configured for network %q", networkName)
	}
	privateKey, publicKey, err := parseAptosPrivateKey(privateKeyRaw)
	if err != nil {
		return integrations.BroadcastTransferResult{}, fmt.Errorf("invalid aptos private key for network %q", networkName)
	}

	sender := aptosAddressFromPublicKey(publicKey)
	receiver := normalizeHexAddress(req.ToAddress)
	if receiver == "" {
		return integrations.BroadcastTransferResult{}, fmt.Errorf("invalid destination address")
	}

	rpcURL, err := p.rpcURLForNetwork(networkName)
	if err != nil {
		return integrations.BroadcastTransferResult{}, err
	}
	baseURL := strings.TrimRight(rpcURL, "/")

	var accountResp struct {
		SequenceNumber string `json:"sequence_number"`
	}
	if err := p.httpGetJSON(ctx, baseURL+"/v1/accounts/"+sender, &accountResp); err != nil {
		return integrations.BroadcastTransferResult{}, err
	}
	if strings.TrimSpace(accountResp.SequenceNumber) == "" {
		return integrations.BroadcastTransferResult{}, fmt.Errorf("aptos sequence_number unavailable")
	}

	gasPrice := "100"
	var gasResp struct {
		GasEstimate string `json:"gas_estimate"`
	}
	if err := p.httpGetJSON(ctx, baseURL+"/v1/estimate_gas_price", &gasResp); err == nil && strings.TrimSpace(gasResp.GasEstimate) != "" {
		gasPrice = strings.TrimSpace(gasResp.GasEstimate)
	}

	expires := strconv.FormatInt(time.Now().UTC().Add(10*time.Minute).Unix(), 10)
	submission := map[string]any{
		"sender":                    sender,
		"sequence_number":           strings.TrimSpace(accountResp.SequenceNumber),
		"max_gas_amount":            "2000",
		"gas_unit_price":            gasPrice,
		"expiration_timestamp_secs": expires,
		"payload": map[string]any{
			"type":           "entry_function_payload",
			"function":       "0x1::aptos_account::transfer",
			"type_arguments": []any{},
			"arguments":      []any{receiver, strconv.FormatInt(req.AmountMinor, 10)},
		},
	}

	var signingHex string
	if err := p.httpPostJSON(ctx, baseURL+"/v1/transactions/encode_submission", submission, &signingHex); err != nil {
		return integrations.BroadcastTransferResult{}, err
	}
	signingBytes, err := decodeHexString(signingHex)
	if err != nil {
		return integrations.BroadcastTransferResult{}, fmt.Errorf("invalid aptos signing payload")
	}

	sig := ed25519.Sign(privateKey, aptosSigningMessage(signingBytes))
	signedTxn := map[string]any{}
	for k, v := range submission {
		signedTxn[k] = v
	}
	signedTxn["signature"] = map[string]any{
		"type":       "ed25519_signature",
		"public_key": "0x" + hex.EncodeToString(publicKey),
		"signature":  "0x" + hex.EncodeToString(sig),
	}

	var submitResp struct {
		Hash string `json:"hash"`
	}
	if err := p.httpPostJSON(ctx, baseURL+"/v1/transactions", signedTxn, &submitResp); err != nil {
		return integrations.BroadcastTransferResult{}, err
	}
	if strings.TrimSpace(submitResp.Hash) == "" {
		return integrations.BroadcastTransferResult{}, fmt.Errorf("aptos submit response missing hash")
	}

	return integrations.BroadcastTransferResult{TxHash: strings.TrimSpace(submitResp.Hash)}, nil
}

func aptosSigningMessage(rawTxnBytes []byte) []byte {
	salt := sha3.Sum256([]byte("APTOS::RawTransaction"))
	msg := make([]byte, 0, len(salt)+len(rawTxnBytes))
	msg = append(msg, salt[:]...)
	msg = append(msg, rawTxnBytes...)
	return msg
}

func decodeHexString(input string) ([]byte, error) {
	trimmed := strings.TrimSpace(strings.TrimPrefix(input, "0x"))
	if trimmed == "" {
		return nil, fmt.Errorf("empty hex")
	}
	return hex.DecodeString(trimmed)
}

func normalizeHexAddress(input string) string {
	trimmed := strings.ToLower(strings.TrimSpace(input))
	trimmed = strings.TrimPrefix(trimmed, "0x")
	if len(trimmed) == 0 || len(trimmed) > 64 {
		return ""
	}
	if _, err := hex.DecodeString(trimmed); err != nil {
		return ""
	}
	return "0x" + strings.Repeat("0", 64-len(trimmed)) + trimmed
}

func (p *Provider) suiAddressForNetwork(vaultID, network string) (string, error) {
	networkKey := strings.ToLower(strings.TrimSpace(network))
	if networkKey == "" {
		networkKey = p.network
	}

	privateKeyRaw := p.scopedPrivateKey(vaultID, networkKey)
	if privateKeyRaw == "" {
		return "", fmt.Errorf("private key not configured for network %q", networkKey)
	}

	signer, err := parseSuiSigner(privateKeyRaw)
	if err != nil {
		return "", fmt.Errorf("invalid sui private key for network %q", networkKey)
	}
	return strings.ToLower(strings.TrimSpace(signer.Address)), nil
}

func parseSuiSigner(raw string) (*suisigner.Signer, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, fmt.Errorf("empty key")
	}
	if strings.HasPrefix(trimmed, "suiprivkey") {
		return suisigner.NewSignerWithSecretKey(trimmed)
	}
	decoded, err := hex.DecodeString(strings.TrimPrefix(trimmed, "0x"))
	if err != nil {
		return nil, err
	}
	switch len(decoded) {
	case ed25519.SeedSize:
		return suisigner.NewSigner(decoded), nil
	case ed25519.PrivateKeySize:
		return suisigner.NewSigner(decoded[:ed25519.SeedSize]), nil
	default:
		return nil, fmt.Errorf("unexpected sui key length")
	}
}

func (p *Provider) broadcastTransferSui(ctx context.Context, req integrations.BroadcastTransferRequest) (integrations.BroadcastTransferResult, error) {
	networkName := strings.ToLower(strings.TrimSpace(req.Network))
	native, err := p.supportsNativeAsset(ctx, "sui", networkName, req.Asset)
	if err != nil {
		return integrations.BroadcastTransferResult{}, err
	}
	if !native {
		return integrations.BroadcastTransferResult{}, fmt.Errorf("sui broadcast currently supports native SUI only")
	}
	if req.AmountMinor <= 0 {
		return integrations.BroadcastTransferResult{}, fmt.Errorf("amount_minor must be greater than 0")
	}

	privateKeyRaw := p.scopedPrivateKey(req.VaultID, networkName)
	if privateKeyRaw == "" {
		return integrations.BroadcastTransferResult{}, fmt.Errorf("private key not configured for network %q", networkName)
	}
	signer, err := parseSuiSigner(privateKeyRaw)
	if err != nil {
		return integrations.BroadcastTransferResult{}, fmt.Errorf("invalid sui private key for network %q", networkName)
	}

	rpcURL, err := p.rpcURLForNetwork(networkName)
	if err != nil {
		return integrations.BroadcastTransferResult{}, err
	}
	client := sui.NewSuiClient(rpcURL)

	coins, err := client.SuiXGetCoins(ctx, models.SuiXGetCoinsRequest{
		Owner:    signer.Address,
		CoinType: "0x2::sui::SUI",
		Limit:    10,
	})
	if err != nil {
		return integrations.BroadcastTransferResult{}, err
	}
	if len(coins.Data) == 0 {
		return integrations.BroadcastTransferResult{}, fmt.Errorf("no sui coin objects available for signer")
	}

	coinObjectIDs := make([]string, 0, len(coins.Data))
	for _, coin := range coins.Data {
		if strings.TrimSpace(coin.CoinObjectId) != "" {
			coinObjectIDs = append(coinObjectIDs, coin.CoinObjectId)
		}
	}
	if len(coinObjectIDs) == 0 {
		return integrations.BroadcastTransferResult{}, fmt.Errorf("no sui coin object ids available for signer")
	}

	txnMeta, err := client.PaySui(ctx, models.PaySuiRequest{
		Signer:      signer.Address,
		SuiObjectId: coinObjectIDs,
		Recipient:   []string{strings.TrimSpace(req.ToAddress)},
		Amount:      []string{strconv.FormatInt(req.AmountMinor, 10)},
		GasBudget:   "5000000",
	})
	if err != nil {
		return integrations.BroadcastTransferResult{}, err
	}

	resp, err := client.SignAndExecuteTransactionBlock(ctx, models.SignAndExecuteTransactionBlockRequest{
		TxnMetaData: txnMeta,
		PriKey:      signer.PriKey,
		Options:     models.SuiTransactionBlockOptions{ShowEffects: true},
		RequestType: "WaitForLocalExecution",
	})
	if err != nil {
		return integrations.BroadcastTransferResult{}, err
	}
	if strings.TrimSpace(resp.Digest) == "" {
		return integrations.BroadcastTransferResult{}, fmt.Errorf("sui submit response missing digest")
	}

	return integrations.BroadcastTransferResult{TxHash: strings.TrimSpace(resp.Digest)}, nil
}

func (p *Provider) httpPostJSON(ctx context.Context, url string, payload any, out any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
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
		return fmt.Errorf("rpc status %d", resp.StatusCode)
	}

	if err := json.Unmarshal(respBody, out); err != nil {
		return err
	}
	return nil
}

func (p *Provider) httpGetJSON(ctx context.Context, url string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

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
		return fmt.Errorf("rpc status %d", resp.StatusCode)
	}

	if err := json.Unmarshal(respBody, out); err != nil {
		return err
	}
	return nil
}
