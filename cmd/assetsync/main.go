package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ebitezion/Nduracore/internal/data"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	_ = godotenv.Load()

	defaultCatalogPath := strings.TrimSpace(os.Getenv("ASSET_REGISTRY_CATALOG_FILE"))
	seedDefaultNatives := getEnvBool("ASSET_REGISTRY_SEED_DEFAULT_NATIVES", false)
	includeEnvERC20 := getEnvBool("ASSET_REGISTRY_INCLUDE_ENV_ERC20", true)

	catalogPathFlag := flag.String("catalog", defaultCatalogPath, "path to catalog JSON file")
	seedDefaultNativesFlag := flag.Bool("seed-default-natives", seedDefaultNatives, "seed built-in default native assets")
	includeEnvERC20Flag := flag.Bool("from-env-erc20", includeEnvERC20, "include ALCHEMY_ERC20_CONTRACTS_JSON entries")
	resetActiveFlag := flag.Bool("reset-active", false, "set all current assets inactive before upserting new catalog entries")
	flag.Parse()

	dsn := strings.TrimSpace(os.Getenv("DB_DSN"))
	if dsn == "" {
		log.Fatal("DB_DSN is required")
	}

	entries := make([]data.AssetDefinition, 0)
	if *seedDefaultNativesFlag {
		entries = append(entries, defaultNativeAssets()...)
	}

	catalogPath := strings.TrimSpace(*catalogPathFlag)
	if catalogPath != "" {
		catalogEntries, err := loadCatalogFile(catalogPath)
		if err != nil {
			log.Fatalf("load catalog file failed: %v", err)
		}
		entries = append(entries, catalogEntries...)
	}

	if *includeEnvERC20Flag {
		envEntries, err := loadERC20FromEnv()
		if err != nil {
			log.Fatalf("load env ERC20 catalog failed: %v", err)
		}
		entries = append(entries, envEntries...)
	}

	if len(entries) == 0 {
		log.Fatal("no asset entries to sync (provide --catalog and/or --seed-default-natives and/or --from-env-erc20)")
	}

	byKey := make(map[string]data.AssetDefinition, len(entries))
	for _, entry := range entries {
		key := normalizePairKey(entry.Network + ":" + entry.AssetCode)
		if key == ":" {
			continue
		}
		byKey[key] = entry
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.PingContext(context.Background()); err != nil {
		log.Fatal(err)
	}

	models := data.NewModels(db)
	if *resetActiveFlag {
		if _, err := db.ExecContext(context.Background(), "UPDATE asset_registry SET is_active = false WHERE is_active = true"); err != nil {
			log.Fatalf("reset active assets failed: %v", err)
		}
	}

	keys := make([]string, 0, len(byKey))
	for k := range byKey {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	inserted := 0
	for _, key := range keys {
		item := byKey[key]
		if err := models.Assets.UpsertAsset(context.Background(), &item); err != nil {
			log.Fatalf("upsert asset %s:%s failed: %v", item.Network, item.AssetCode, err)
		}
		inserted++
	}

	fmt.Printf("asset sync completed: %d upsert(s)\n", inserted)
}

func splitPairKey(key string) (string, string, bool) {
	parts := strings.Split(strings.TrimSpace(key), ":")
	if len(parts) != 2 {
		return "", "", false
	}
	network := strings.ToLower(strings.TrimSpace(parts[0]))
	asset := strings.ToUpper(strings.TrimSpace(parts[1]))
	if network == "" || asset == "" {
		return "", "", false
	}
	return network, asset, true
}

func normalizePairKey(key string) string {
	network, asset, ok := splitPairKey(key)
	if !ok {
		return strings.TrimSpace(strings.ToLower(key))
	}
	return network + ":" + asset
}

type catalogFile struct {
	Assets []catalogEntry `json:"assets"`
}

type catalogEntry struct {
	Network         string                 `json:"network"`
	AssetCode       string                 `json:"asset_code"`
	ChainFamily     string                 `json:"chain_family"`
	AssetType       string                 `json:"asset_type"`
	ContractAddress string                 `json:"contract_address"`
	Decimals        *int                   `json:"decimals"`
	IsActive        *bool                  `json:"is_active"`
	Metadata        map[string]interface{} `json:"metadata"`
}

func loadCatalogFile(path string) ([]data.AssetDefinition, error) {
	raw, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}

	var direct []catalogEntry
	if err := json.Unmarshal(raw, &direct); err == nil && len(direct) > 0 {
		return normalizeCatalogEntries(direct), nil
	}

	var wrapped catalogFile
	if err := json.Unmarshal(raw, &wrapped); err != nil {
		return nil, err
	}
	return normalizeCatalogEntries(wrapped.Assets), nil
}

func normalizeCatalogEntries(entries []catalogEntry) []data.AssetDefinition {
	out := make([]data.AssetDefinition, 0, len(entries))
	for _, item := range entries {
		network := strings.ToLower(strings.TrimSpace(item.Network))
		asset := strings.ToUpper(strings.TrimSpace(item.AssetCode))
		if network == "" || asset == "" {
			continue
		}

		assetType := strings.ToLower(strings.TrimSpace(item.AssetType))
		if assetType == "" {
			if strings.TrimSpace(item.ContractAddress) == "" {
				assetType = "native"
			} else {
				assetType = "token"
			}
		}

		chainFamily := strings.ToLower(strings.TrimSpace(item.ChainFamily))
		if chainFamily == "" {
			chainFamily = networkFamily(network)
		}

		isActive := true
		if item.IsActive != nil {
			isActive = *item.IsActive
		}

		decimals := 18
		if assetType == "token" {
			decimals = 6
		}
		if item.Decimals != nil && *item.Decimals >= 0 {
			decimals = *item.Decimals
		}

		metadata := item.Metadata
		if metadata == nil {
			metadata = map[string]interface{}{}
		}
		metadata["synced_via"] = "cmd/assetsync"
		if _, exists := metadata["source"]; !exists {
			metadata["source"] = "catalog_file"
		}

		out = append(out, data.AssetDefinition{
			Network:         network,
			AssetCode:       asset,
			ChainFamily:     chainFamily,
			AssetType:       assetType,
			ContractAddress: strings.TrimSpace(item.ContractAddress),
			Decimals:        decimals,
			IsActive:        isActive,
			Metadata:        metadata,
		})
	}

	return out
}

func loadERC20FromEnv() ([]data.AssetDefinition, error) {
	contractsRaw := strings.TrimSpace(os.Getenv("ALCHEMY_ERC20_CONTRACTS_JSON"))
	if contractsRaw == "" {
		return nil, nil
	}

	contracts := map[string]string{}
	if err := json.Unmarshal([]byte(contractsRaw), &contracts); err != nil {
		return nil, fmt.Errorf("invalid ALCHEMY_ERC20_CONTRACTS_JSON: %w", err)
	}

	decimalsByKey := map[string]int{}
	decimalsRaw := strings.TrimSpace(os.Getenv("ALCHEMY_ERC20_DECIMALS_JSON"))
	if decimalsRaw != "" {
		raw := map[string]int{}
		if err := json.Unmarshal([]byte(decimalsRaw), &raw); err != nil {
			return nil, fmt.Errorf("invalid ALCHEMY_ERC20_DECIMALS_JSON: %w", err)
		}
		for k, v := range raw {
			decimalsByKey[normalizePairKey(k)] = v
		}
	}

	keys := make([]string, 0, len(contracts))
	for k := range contracts {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	out := make([]data.AssetDefinition, 0, len(keys))
	for _, pairKey := range keys {
		network, asset, ok := splitPairKey(pairKey)
		if !ok {
			log.Printf("skip invalid key %q (expected network:ASSET)", pairKey)
			continue
		}

		contractAddress := strings.TrimSpace(contracts[pairKey])
		if contractAddress == "" {
			log.Printf("skip empty contract for %q", pairKey)
			continue
		}

		decimals := 6
		if v, exists := decimalsByKey[normalizePairKey(pairKey)]; exists && v >= 0 {
			decimals = v
		}

		out = append(out, data.AssetDefinition{
			Network:         network,
			AssetCode:       asset,
			ChainFamily:     networkFamily(network),
			AssetType:       "token",
			ContractAddress: contractAddress,
			Decimals:        decimals,
			IsActive:        true,
			Metadata: map[string]interface{}{
				"source":     "alchemy_env",
				"synced_via": "cmd/assetsync",
				"pair_key":   normalizePairKey(pairKey),
			},
		})
	}

	return out, nil
}

func defaultNativeAssets() []data.AssetDefinition {
	base := map[string]interface{}{
		"source":     "default_native_seed",
		"synced_via": "cmd/assetsync",
	}
	return []data.AssetDefinition{
		{Network: "eth-mainnet", AssetCode: "ETH", ChainFamily: "evm", AssetType: "native", ContractAddress: "", Decimals: 18, IsActive: true, Metadata: cloneMetadata(base)},
		{Network: "eth-sepolia", AssetCode: "ETH", ChainFamily: "evm", AssetType: "native", ContractAddress: "", Decimals: 18, IsActive: true, Metadata: cloneMetadata(base)},
		{Network: "polygon-mainnet", AssetCode: "MATIC", ChainFamily: "evm", AssetType: "native", ContractAddress: "", Decimals: 18, IsActive: true, Metadata: cloneMetadata(base)},
		{Network: "polygon-amoy", AssetCode: "MATIC", ChainFamily: "evm", AssetType: "native", ContractAddress: "", Decimals: 18, IsActive: true, Metadata: cloneMetadata(base)},
		{Network: "arbitrum-mainnet", AssetCode: "ETH", ChainFamily: "evm", AssetType: "native", ContractAddress: "", Decimals: 18, IsActive: true, Metadata: cloneMetadata(base)},
		{Network: "arbitrum-sepolia", AssetCode: "ETH", ChainFamily: "evm", AssetType: "native", ContractAddress: "", Decimals: 18, IsActive: true, Metadata: cloneMetadata(base)},
		{Network: "optimism-mainnet", AssetCode: "ETH", ChainFamily: "evm", AssetType: "native", ContractAddress: "", Decimals: 18, IsActive: true, Metadata: cloneMetadata(base)},
		{Network: "optimism-sepolia", AssetCode: "ETH", ChainFamily: "evm", AssetType: "native", ContractAddress: "", Decimals: 18, IsActive: true, Metadata: cloneMetadata(base)},
		{Network: "base-mainnet", AssetCode: "ETH", ChainFamily: "evm", AssetType: "native", ContractAddress: "", Decimals: 18, IsActive: true, Metadata: cloneMetadata(base)},
		{Network: "base-sepolia", AssetCode: "ETH", ChainFamily: "evm", AssetType: "native", ContractAddress: "", Decimals: 18, IsActive: true, Metadata: cloneMetadata(base)},
	}
}

func cloneMetadata(in map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func getEnvBool(key string, fallback bool) bool {
	raw := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	switch raw {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return fallback
	}
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
