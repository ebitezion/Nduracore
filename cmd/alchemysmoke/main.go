package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

func main() {
	cfg, err := loadSmokeConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	if len(cfg.rpcURLs) == 0 {
		fmt.Fprintln(os.Stderr, "no networks configured; set ALCHEMY_RPC_URLS_JSON / ALCHEMY_API_KEYS_JSON / ALCHEMY_RPC_URL / ALCHEMY_API_KEY")
		os.Exit(1)
	}

	networks := make([]string, 0, len(cfg.rpcURLs))
	for network := range cfg.rpcURLs {
		networks = append(networks, network)
	}
	sort.Strings(networks)

	client := &http.Client{Timeout: 12 * time.Second}
	failed := false

	fmt.Printf("Alchemy RPC smoke check: %d network(s)\n", len(networks))
	for _, network := range networks {
		chainID, err := probeNetwork(context.Background(), client, network, cfg.rpcURLs[network])
		if err != nil {
			failed = true
			fmt.Printf("FAIL  %-20s %v\n", network, err)
			continue
		}
		fmt.Printf("PASS  %-20s chainId=%s\n", network, chainID)
	}

	if failed {
		os.Exit(1)
	}
}

type smokeConfig struct {
	network string
	rpcURLs map[string]string
}

func loadSmokeConfig() (smokeConfig, error) {
	cfg := smokeConfig{
		network: strings.ToLower(strings.TrimSpace(getEnv("ALCHEMY_NETWORK", "eth-sepolia"))),
		rpcURLs: map[string]string{},
	}
	if cfg.network == "" {
		cfg.network = "eth-sepolia"
	}

	apiKey, err := getSecretEnv("ALCHEMY_API_KEY", "")
	if err != nil {
		return cfg, err
	}
	apiKeysJSON, err := getSecretEnv("ALCHEMY_API_KEYS_JSON", "")
	if err != nil {
		return cfg, err
	}
	apiKeys, err := parseStringMapJSON(apiKeysJSON)
	if err != nil {
		return cfg, fmt.Errorf("invalid ALCHEMY_API_KEYS_JSON: %w", err)
	}

	rpcURL := getEnv("ALCHEMY_RPC_URL", "")
	rpcURLsJSON, err := getSecretEnv("ALCHEMY_RPC_URLS_JSON", "")
	if err != nil {
		return cfg, err
	}
	rpcURLs, err := parseStringMapJSON(rpcURLsJSON)
	if err != nil {
		return cfg, fmt.Errorf("invalid ALCHEMY_RPC_URLS_JSON: %w", err)
	}
	for n, url := range rpcURLs {
		cfg.rpcURLs[n] = url
	}

	for n, key := range apiKeys {
		if _, exists := cfg.rpcURLs[n]; !exists {
			cfg.rpcURLs[n] = fmt.Sprintf("https://%s.g.alchemy.com/v2/%s", n, key)
		}
	}

	if strings.TrimSpace(rpcURL) != "" {
		cfg.rpcURLs[cfg.network] = strings.TrimSpace(rpcURL)
	} else if _, exists := cfg.rpcURLs[cfg.network]; !exists && strings.TrimSpace(apiKey) != "" {
		cfg.rpcURLs[cfg.network] = fmt.Sprintf("https://%s.g.alchemy.com/v2/%s", cfg.network, strings.TrimSpace(apiKey))
	}

	return cfg, nil
}

func probeNetwork(ctx context.Context, client *http.Client, network, rpcURL string) (string, error) {
	switch networkFamily(network) {
	case "solana":
		return fetchRPCResult(ctx, client, rpcURL, "getHealth", []any{})
	case "stellar":
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, rpcURL, nil)
		if err != nil {
			return "", err
		}
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			return "", fmt.Errorf("http status %d", resp.StatusCode)
		}
		return "ok", nil
	case "xrpl":
		return fetchRPCResult(ctx, client, rpcURL, "server_info", []any{map[string]any{}})
	case "aptos":
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimRight(rpcURL, "/")+"/v1", nil)
		if err != nil {
			return "", err
		}
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			return "", fmt.Errorf("http status %d", resp.StatusCode)
		}
		return "ok", nil
	case "sui":
		return fetchRPCResult(ctx, client, rpcURL, "suix_getLatestSuiSystemState", []any{})
	case "tron":
		return fetchRPCResult(ctx, client, strings.TrimRight(rpcURL, "/")+"/wallet/getnodeinfo", "", nil)
	default:
		return fetchRPCResult(ctx, client, rpcURL, "eth_chainId", []any{})
	}
}

func fetchRPCResult(ctx context.Context, client *http.Client, rpcURL, method string, params []any) (string, error) {
	var reqBody io.Reader
	if strings.TrimSpace(method) != "" {
		payload := map[string]any{
			"jsonrpc": "2.0",
			"id":      1,
			"method":  method,
			"params":  params,
		}
		body, err := json.Marshal(payload)
		if err != nil {
			return "", err
		}
		reqBody = bytes.NewReader(body)
	} else {
		reqBody = bytes.NewReader([]byte(`{}`))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rpcURL, reqBody)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("rpc status %d", resp.StatusCode)
	}

	var envelope struct {
		Result any `json:"result"`
		Error  *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(respBody, &envelope); err != nil {
		return "", err
	}
	if envelope.Error != nil {
		return "", fmt.Errorf("rpc error %d: %s", envelope.Error.Code, envelope.Error.Message)
	}
	if envelope.Result == nil {
		// Tron wallet endpoints do not use jsonrpc envelopes.
		var raw map[string]any
		if err := json.Unmarshal(respBody, &raw); err == nil && len(raw) > 0 {
			return "ok", nil
		}
		return "", fmt.Errorf("empty result response")
	}

	switch v := envelope.Result.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return "", fmt.Errorf("empty result response")
		}
		return strings.TrimSpace(v), nil
	default:
		return "ok", nil
	}
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getSecretEnv(key, fallback string) (string, error) {
	direct := strings.TrimSpace(os.Getenv(key))
	if direct != "" {
		return direct, nil
	}

	filePath := strings.TrimSpace(os.Getenv(key + "_FILE"))
	if filePath != "" {
		// #nosec G304,G703 -- path is supplied by deployment-controlled environment configuration.
		value, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("%s_FILE: %w", key, err)
		}
		secret := strings.TrimSpace(string(value))
		if secret != "" {
			return secret, nil
		}
	}

	ref := strings.TrimSpace(os.Getenv(key + "_REF"))
	if ref != "" {
		secret, err := resolveSecretReference(ref)
		if err != nil {
			return "", fmt.Errorf("%s_REF: %w", key, err)
		}
		if strings.TrimSpace(secret) != "" {
			return strings.TrimSpace(secret), nil
		}
	}

	return fallback, nil
}

func resolveSecretReference(ref string) (string, error) {
	switch {
	case strings.HasPrefix(ref, "env:"):
		target := strings.TrimSpace(strings.TrimPrefix(ref, "env:"))
		if target == "" {
			return "", fmt.Errorf("empty env reference")
		}
		return os.Getenv(target), nil
	case strings.HasPrefix(ref, "file:"):
		target := strings.TrimSpace(strings.TrimPrefix(ref, "file:"))
		if target == "" {
			return "", fmt.Errorf("empty file reference")
		}
		// #nosec G304,G703 -- file reference is trusted operational input from env/config.
		content, err := os.ReadFile(target)
		if err != nil {
			return "", err
		}
		return string(content), nil
	default:
		return "", fmt.Errorf("unsupported reference scheme")
	}
}

func parseStringMapJSON(raw string) (map[string]string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return map[string]string{}, nil
	}

	parsed := make(map[string]string)
	if err := json.Unmarshal([]byte(trimmed), &parsed); err != nil {
		return nil, err
	}

	result := make(map[string]string, len(parsed))
	for key, value := range parsed {
		k := strings.TrimSpace(strings.ToLower(key))
		v := strings.TrimSpace(value)
		if k == "" || v == "" {
			continue
		}
		result[k] = v
	}
	return result, nil
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
