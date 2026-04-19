package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type sourceConfig struct {
	URL         string
	NetworkHint string
}

type sourceTokenList struct {
	Name      string        `json:"name"`
	Timestamp string        `json:"timestamp"`
	Tokens    []sourceToken `json:"tokens"`
}

type sourceToken struct {
	ChainID  int    `json:"chainId"`
	Address  string `json:"address"`
	Symbol   string `json:"symbol"`
	Decimals int    `json:"decimals"`
}

type catalog struct {
	Assets []catalogAsset `json:"assets"`
}

type catalogAsset struct {
	Network         string                 `json:"network"`
	AssetCode       string                 `json:"asset_code"`
	ChainFamily     string                 `json:"chain_family"`
	AssetType       string                 `json:"asset_type"`
	ContractAddress string                 `json:"contract_address,omitempty"`
	Decimals        int                    `json:"decimals"`
	IsActive        bool                   `json:"is_active"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

func main() {
	var (
		outPath = flag.String("out", "./documentation/asset_catalog.top1000.json", "output catalog path")
		limit   = flag.Int("limit", 1000, "max assets in generated catalog")
		timeout = flag.Duration("timeout", 45*time.Second, "HTTP timeout")
	)
	flag.Parse()

	if *limit < 1 {
		fmt.Fprintln(os.Stderr, "limit must be >= 1")
		os.Exit(1)
	}

	sources := defaultSources()
	fetched := make([]fetchedSource, 0, len(sources))
	for _, src := range sources {
		list, err := fetchSource(src.URL, *timeout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warn: fetch %s failed: %v\n", src.URL, err)
			continue
		}
		fetched = append(fetched, fetchedSource{Config: src, List: list})
	}
	if len(fetched) == 0 {
		fmt.Fprintln(os.Stderr, "no source token lists fetched")
		os.Exit(1)
	}

	assets := buildCatalogAssets(fetched, *limit)
	if len(assets) == 0 {
		fmt.Fprintln(os.Stderr, "no assets generated")
		os.Exit(1)
	}

	payload, err := json.MarshalIndent(catalog{Assets: assets}, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "marshal catalog failed: %v\n", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(filepath.Dir(*outPath), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "create output dir failed: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(*outPath, payload, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write catalog failed: %v\n", err)
		os.Exit(1)
	}

	byTier := map[string]int{"mainnet": 0, "testnet": 0}
	for _, a := range assets {
		tier := tierFromNetwork(a.Network)
		byTier[tier]++
	}
	fmt.Printf("generated catalog: %s (%d assets; mainnet=%d testnet=%d)\n", *outPath, len(assets), byTier["mainnet"], byTier["testnet"])
}

type fetchedSource struct {
	Config sourceConfig
	List   sourceTokenList
}

func defaultSources() []sourceConfig {
	return []sourceConfig{
		{URL: "https://tokens.coingecko.com/uniswap/all.json", NetworkHint: ""},
		{URL: "https://tokens.coingecko.com/arbitrum-one/all.json", NetworkHint: "arbitrum-mainnet"},
		{URL: "https://tokens.coingecko.com/optimistic-ethereum/all.json", NetworkHint: "optimism-mainnet"},
		{URL: "https://tokens.coingecko.com/polygon-pos/all.json", NetworkHint: "polygon-mainnet"},
		{URL: "https://tokens.coingecko.com/base/all.json", NetworkHint: "base-mainnet"},
		{URL: "https://tokens.coingecko.com/binance-smart-chain/all.json", NetworkHint: "bnb-mainnet"},
		{URL: "https://tokens.coingecko.com/avalanche/all.json", NetworkHint: "avalanche-mainnet"},
		{URL: "https://tokens.coingecko.com/linea/all.json", NetworkHint: "linea-mainnet"},
		{URL: "https://tokens.coingecko.com/scroll/all.json", NetworkHint: "scroll-mainnet"},
	}
}

func fetchSource(sourceURL string, timeout time.Duration) (sourceTokenList, error) {
	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequest(http.MethodGet, strings.TrimSpace(sourceURL), nil)
	if err != nil {
		return sourceTokenList{}, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return sourceTokenList{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return sourceTokenList{}, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var list sourceTokenList
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return sourceTokenList{}, err
	}
	return list, nil
}

func buildCatalogAssets(sources []fetchedSource, limit int) []catalogAsset {
	assets := make([]catalogAsset, 0, limit)
	seen := make(map[string]struct{}, limit)

	latestTimestamp := ""
	for _, src := range sources {
		if strings.TrimSpace(src.List.Timestamp) > latestTimestamp {
			latestTimestamp = strings.TrimSpace(src.List.Timestamp)
		}
	}

	for _, item := range defaultNativeAssets(latestTimestamp) {
		key := item.Network + ":" + item.AssetCode
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		assets = append(assets, item)
		if len(assets) >= limit {
			return assets
		}
	}

	networkCandidates := make(map[string][]catalogAsset)
	for _, src := range sources {
		for _, tok := range src.List.Tokens {
			network, ok := networkFromSource(src.Config, tok.ChainID)
			if !ok {
				continue
			}

			symbol := strings.ToUpper(strings.TrimSpace(tok.Symbol))
			address := strings.TrimSpace(tok.Address)
			if symbol == "" || !isHexAddress(address) {
				continue
			}
			if tok.Decimals < 0 || tok.Decimals > 36 {
				continue
			}

			asset := catalogAsset{
				Network:         network,
				AssetCode:       symbol,
				ChainFamily:     "evm",
				AssetType:       "token",
				ContractAddress: address,
				Decimals:        tok.Decimals,
				IsActive:        true,
				Metadata: map[string]interface{}{
					"source":           "coingecko_token_list",
					"source_list":      src.List.Name,
					"source_url":       src.Config.URL,
					"source_timestamp": strings.TrimSpace(src.List.Timestamp),
					"network_tier":     tierFromNetwork(network),
					"generated_by":     "cmd/assetcataloggen",
				},
			}
			networkCandidates[network] = append(networkCandidates[network], asset)
		}
	}

	networks := make([]string, 0, len(networkCandidates))
	for n := range networkCandidates {
		networks = append(networks, n)
	}
	sortNetworksForSelection(networks)

	// De-dupe per network by asset_code to satisfy DB uniqueness.
	for _, n := range networks {
		sort.Slice(networkCandidates[n], func(i, j int) bool {
			ai := networkCandidates[n][i]
			aj := networkCandidates[n][j]
			if ai.AssetCode == aj.AssetCode {
				return strings.ToLower(ai.ContractAddress) < strings.ToLower(aj.ContractAddress)
			}
			return ai.AssetCode < aj.AssetCode
		})
		uniq := make([]catalogAsset, 0, len(networkCandidates[n]))
		seenSymbol := map[string]struct{}{}
		for _, c := range networkCandidates[n] {
			if _, exists := seenSymbol[c.AssetCode]; exists {
				continue
			}
			seenSymbol[c.AssetCode] = struct{}{}
			uniq = append(uniq, c)
		}
		networkCandidates[n] = uniq
	}

	indices := make(map[string]int, len(networks))
	for len(assets) < limit {
		progress := false
		for _, n := range networks {
			i := indices[n]
			cands := networkCandidates[n]
			if i >= len(cands) {
				continue
			}
			item := cands[i]
			indices[n] = i + 1

			key := item.Network + ":" + item.AssetCode
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			assets = append(assets, item)
			progress = true
			if len(assets) >= limit {
				break
			}
		}
		if !progress {
			break
		}
	}

	return assets
}

func networkFromSource(src sourceConfig, chainID int) (string, bool) {
	if strings.TrimSpace(src.NetworkHint) != "" {
		return strings.TrimSpace(src.NetworkHint), true
	}
	switch chainID {
	case 1:
		return "eth-mainnet", true
	case 10:
		return "optimism-mainnet", true
	case 56:
		return "bnb-mainnet", true
	case 137:
		return "polygon-mainnet", true
	case 480:
		return "worldchain-mainnet", true
	case 8453:
		return "base-mainnet", true
	case 42161:
		return "arbitrum-mainnet", true
	case 43114:
		return "avalanche-mainnet", true
	case 59144:
		return "linea-mainnet", true
	case 534352:
		return "scroll-mainnet", true
	default:
		return "", false
	}
}

func defaultNativeAssets(sourceTimestamp string) []catalogAsset {
	baseMeta := map[string]interface{}{
		"source":           "default_native_seed",
		"source_timestamp": sourceTimestamp,
		"generated_by":     "cmd/assetcataloggen",
	}
	natives := []catalogAsset{
		{Network: "eth-mainnet", AssetCode: "ETH", ChainFamily: "evm", AssetType: "native", Decimals: 18, IsActive: true},
		{Network: "eth-sepolia", AssetCode: "ETH", ChainFamily: "evm", AssetType: "native", Decimals: 18, IsActive: true},
		{Network: "polygon-mainnet", AssetCode: "MATIC", ChainFamily: "evm", AssetType: "native", Decimals: 18, IsActive: true},
		{Network: "polygon-amoy", AssetCode: "MATIC", ChainFamily: "evm", AssetType: "native", Decimals: 18, IsActive: true},
		{Network: "arbitrum-mainnet", AssetCode: "ETH", ChainFamily: "evm", AssetType: "native", Decimals: 18, IsActive: true},
		{Network: "arbitrum-sepolia", AssetCode: "ETH", ChainFamily: "evm", AssetType: "native", Decimals: 18, IsActive: true},
		{Network: "optimism-mainnet", AssetCode: "ETH", ChainFamily: "evm", AssetType: "native", Decimals: 18, IsActive: true},
		{Network: "optimism-sepolia", AssetCode: "ETH", ChainFamily: "evm", AssetType: "native", Decimals: 18, IsActive: true},
		{Network: "base-mainnet", AssetCode: "ETH", ChainFamily: "evm", AssetType: "native", Decimals: 18, IsActive: true},
		{Network: "base-sepolia", AssetCode: "ETH", ChainFamily: "evm", AssetType: "native", Decimals: 18, IsActive: true},
		{Network: "bnb-mainnet", AssetCode: "BNB", ChainFamily: "evm", AssetType: "native", Decimals: 18, IsActive: true},
		{Network: "bnb-testnet", AssetCode: "BNB", ChainFamily: "evm", AssetType: "native", Decimals: 18, IsActive: true},
		{Network: "avalanche-mainnet", AssetCode: "AVAX", ChainFamily: "evm", AssetType: "native", Decimals: 18, IsActive: true},
		{Network: "avalanche-fuji", AssetCode: "AVAX", ChainFamily: "evm", AssetType: "native", Decimals: 18, IsActive: true},
		{Network: "linea-mainnet", AssetCode: "ETH", ChainFamily: "evm", AssetType: "native", Decimals: 18, IsActive: true},
		{Network: "linea-sepolia", AssetCode: "ETH", ChainFamily: "evm", AssetType: "native", Decimals: 18, IsActive: true},
		{Network: "scroll-mainnet", AssetCode: "ETH", ChainFamily: "evm", AssetType: "native", Decimals: 18, IsActive: true},
		{Network: "scroll-sepolia", AssetCode: "ETH", ChainFamily: "evm", AssetType: "native", Decimals: 18, IsActive: true},
		{Network: "worldchain-mainnet", AssetCode: "ETH", ChainFamily: "evm", AssetType: "native", Decimals: 18, IsActive: true},
		{Network: "worldchain-sepolia", AssetCode: "ETH", ChainFamily: "evm", AssetType: "native", Decimals: 18, IsActive: true},
	}

	for i := range natives {
		natives[i].Metadata = cloneMap(baseMeta)
		natives[i].Metadata["network_tier"] = tierFromNetwork(natives[i].Network)
	}
	return natives
}

func tierFromNetwork(network string) string {
	n := strings.ToLower(strings.TrimSpace(network))
	testnetMarkers := []string{"sepolia", "amoy", "fuji", "testnet", "devnet", "goerli", "mumbai", "alfajores"}
	for _, marker := range testnetMarkers {
		if strings.Contains(n, marker) {
			return "testnet"
		}
	}
	return "mainnet"
}

func sortNetworksForSelection(networks []string) {
	sort.Slice(networks, func(i, j int) bool {
		ni := networks[i]
		nj := networks[j]
		ti := tierFromNetwork(ni)
		tj := tierFromNetwork(nj)
		if ti != tj {
			return ti == "mainnet"
		}
		return ni < nj
	})
}

func cloneMap(in map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func isHexAddress(value string) bool {
	v := strings.TrimSpace(value)
	if len(v) != 42 || !strings.HasPrefix(v, "0x") {
		return false
	}
	for _, r := range v[2:] {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') && (r < 'A' || r > 'F') {
			return false
		}
	}
	return true
}
