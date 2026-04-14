package alchemy

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ebitezion/Nduracore/internal/integrations"
)

type Config struct {
	Network               string
	APIKey                string
	RPCURL                string
	ConfirmationsRequired int
	WebhookSigningSecret  string
	EnableLiveDeposits    bool
}

type Provider struct {
	network               string
	apiKey                string
	rpcURL                string
	confirmationsRequired int
	webhookSigningSecret  string
	enableLiveDeposits    bool
	httpClient            *http.Client

	mu              sync.Mutex
	depositsEmitted map[string]bool
}

func NewProvider(cfg Config) *Provider {
	network := strings.TrimSpace(cfg.Network)
	if network == "" {
		network = "eth-sepolia"
	}

	rpcURL := strings.TrimSpace(cfg.RPCURL)
	if rpcURL == "" && strings.TrimSpace(cfg.APIKey) != "" {
		rpcURL = fmt.Sprintf("https://%s.g.alchemy.com/v2/%s", network, strings.TrimSpace(cfg.APIKey))
	}

	confirmations := cfg.ConfirmationsRequired
	if confirmations < 1 {
		confirmations = 12
	}

	enableLive := cfg.EnableLiveDeposits && rpcURL != ""

	return &Provider{
		network:               network,
		apiKey:                strings.TrimSpace(cfg.APIKey),
		rpcURL:                rpcURL,
		confirmationsRequired: confirmations,
		webhookSigningSecret:  strings.TrimSpace(cfg.WebhookSigningSecret),
		enableLiveDeposits:    enableLive,
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
	h := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%s:%d:%s", req.TenantID, req.FromAddress, req.ToAddress, req.AmountMinor, time.Now().UTC().Format(time.RFC3339Nano))))
	return integrations.BroadcastTransferResult{TxHash: "0x" + hex.EncodeToString(h[:])}, nil
}

func (p *Provider) ListDeposits(ctx context.Context, req integrations.ListDepositsRequest) ([]integrations.DetectedDeposit, error) {
	address := strings.ToLower(strings.TrimSpace(req.Address))
	if address == "" {
		return nil, nil
	}

	if p.enableLiveDeposits {
		deposits, err := p.listDepositsLive(ctx, address)
		if err == nil {
			return deposits, nil
		}
	}

	return p.listDepositsFallback(req.TenantID, address), nil
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

func (p *Provider) listDepositsLive(ctx context.Context, toAddress string) ([]integrations.DetectedDeposit, error) {
	latestBlock, err := p.ethBlockNumber(ctx)
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
	if err := p.rpcCall(ctx, "alchemy_getAssetTransfers", params, &resp); err != nil {
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

func (p *Provider) listDepositsFallback(tenantID, address string) []integrations.DetectedDeposit {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.depositsEmitted[address] {
		return nil
	}

	h := sha256.Sum256([]byte(tenantID + ":" + address + ":deposit"))
	txHash := "0x" + hex.EncodeToString(h[:])
	p.depositsEmitted[address] = true

	return []integrations.DetectedDeposit{{
		TxHash:        txHash,
		AmountMinor:   250000,
		Confirmations: p.confirmationsRequired,
		Status:        "confirmed",
	}}
}

func (p *Provider) ethBlockNumber(ctx context.Context) (int64, error) {
	var resp struct {
		Result string `json:"result"`
	}
	if err := p.rpcCall(ctx, "eth_blockNumber", []any{}, &resp); err != nil {
		return 0, err
	}
	value := parseHexInt64(resp.Result)
	if value < 0 {
		return 0, fmt.Errorf("invalid block number response")
	}
	return value, nil
}

func (p *Provider) rpcCall(ctx context.Context, method string, params []any, out any) error {
	if strings.TrimSpace(p.rpcURL) == "" {
		return fmt.Errorf("alchemy rpc url not configured")
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.rpcURL, bytes.NewReader(body))
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
		return fmt.Errorf("alchemy rpc status %d", resp.StatusCode)
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
