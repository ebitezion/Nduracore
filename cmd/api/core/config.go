package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// config type allows for system configuration
type config struct {
	port    int
	env     string
	appName string
	version string
	db      struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	security struct {
		corsTrustedOrigins []string
		rateLimitRPS       float64
		rateLimitBurst     int
		authRateLimitRPS   float64
		authRateLimitBurst int
		tokenSecret        string
		tokenIssuer        string
		tokenAudience      string
		tokenTTL           time.Duration
		trustedProxies     []*net.IPNet
	}
	redis struct {
		addr     string
		password string
		db       int
		queueKey string
		enabled  bool
	}
	tracing struct {
		enabled      bool
		otlpEndpoint string
		sampleRatio  float64
	}
	alchemy struct {
		network               string
		apiKey                string
		apiKeys               map[string]string
		rpcURL                string
		rpcURLs               map[string]string
		privateKey            string
		privateKeys           map[string]string
		vaultPrivateKeys      map[string]string
		erc20Contracts        map[string]string
		confirmationsRequired int
		webhookSigningSecret  string
		enableLiveDeposits    bool
		enableLiveBroadcasts  bool
	}
}

func loadConfig() (config, error) {
	cfg := config{}

	cfg.appName = getEnv("APP_NAME", "nduracore")
	cfg.version = getEnv("APP_VERSION", "1.0.0")
	cfg.env = getEnv("MY_ENV", "development")

	port, err := strconv.Atoi(getEnv("PORT", "4000"))
	if err != nil {
		return cfg, fmt.Errorf("invalid PORT value: %w", err)
	}
	cfg.port = port

	cfg.db.dsn, err = getSecretEnv("DB_DSN", "")
	if err != nil {
		return cfg, err
	}
	cfg.db.dsn = strings.TrimSpace(cfg.db.dsn)
	cfg.db.maxOpenConns = getEnvInt("DB_MAX_OPEN_CONNS", 25)
	cfg.db.maxIdleConns = getEnvInt("DB_MAX_IDLE_CONNS", 25)
	cfg.db.maxIdleTime = getEnv("DB_MAX_IDLE_TIME", "15m")

	cfg.security.corsTrustedOrigins = parseCSV(getEnv("CORS_TRUSTED_ORIGINS", "http://localhost:3000,http://localhost:5173"))
	cfg.security.rateLimitRPS = getEnvFloat("RATE_LIMIT_RPS", 5)
	cfg.security.rateLimitBurst = getEnvInt("RATE_LIMIT_BURST", 10)
	cfg.security.authRateLimitRPS = getEnvFloat("AUTH_RATE_LIMIT_RPS", 1)
	cfg.security.authRateLimitBurst = getEnvInt("AUTH_RATE_LIMIT_BURST", 3)
	cfg.security.tokenSecret, err = getSecretEnv("TOKEN_SECRET", "replace-me-in-production")
	if err != nil {
		return cfg, err
	}
	cfg.security.tokenIssuer = getEnv("TOKEN_ISSUER", cfg.appName)
	cfg.security.tokenAudience = getEnv("TOKEN_AUDIENCE", "nduracore-clients")
	cfg.security.tokenTTL = getEnvDuration("TOKEN_TTL", 24*time.Hour)
	cfg.security.trustedProxies = parseTrustedProxies(getEnv("TRUSTED_PROXIES", ""))

	cfg.redis.addr = getEnv("REDIS_ADDR", "")
	cfg.redis.password, err = getSecretEnv("REDIS_PASSWORD", "")
	if err != nil {
		return cfg, err
	}
	cfg.redis.db = getEnvInt("REDIS_DB", 0)
	cfg.redis.queueKey = getEnv("REDIS_QUEUE_KEY", "nduracore:jobs")
	cfg.redis.enabled = cfg.redis.addr != ""

	cfg.tracing.enabled = getEnvBool("OTEL_ENABLED", false)
	cfg.tracing.otlpEndpoint = getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	cfg.tracing.sampleRatio = getEnvFloat("OTEL_SAMPLE_RATIO", 1.0)
	cfg.alchemy.network = getEnv("ALCHEMY_NETWORK", "eth-sepolia")
	cfg.alchemy.apiKey, err = getSecretEnv("ALCHEMY_API_KEY", "")
	if err != nil {
		return cfg, err
	}
	apiKeysJSON, err := getSecretEnv("ALCHEMY_API_KEYS_JSON", "")
	if err != nil {
		return cfg, err
	}
	cfg.alchemy.apiKeys, err = parseStringMapJSON(apiKeysJSON)
	if err != nil {
		return cfg, fmt.Errorf("invalid ALCHEMY_API_KEYS_JSON: %w", err)
	}
	cfg.alchemy.rpcURL = getEnv("ALCHEMY_RPC_URL", "")
	rpcURLsJSON, err := getSecretEnv("ALCHEMY_RPC_URLS_JSON", "")
	if err != nil {
		return cfg, err
	}
	cfg.alchemy.rpcURLs, err = parseStringMapJSON(rpcURLsJSON)
	if err != nil {
		return cfg, fmt.Errorf("invalid ALCHEMY_RPC_URLS_JSON: %w", err)
	}
	cfg.alchemy.privateKey, err = getSecretEnv("ALCHEMY_PRIVATE_KEY", "")
	if err != nil {
		return cfg, err
	}
	privateKeysJSON, err := getSecretEnv("ALCHEMY_PRIVATE_KEYS_JSON", "")
	if err != nil {
		return cfg, err
	}
	cfg.alchemy.privateKeys, err = parseStringMapJSON(privateKeysJSON)
	if err != nil {
		return cfg, fmt.Errorf("invalid ALCHEMY_PRIVATE_KEYS_JSON: %w", err)
	}
	vaultPrivateKeysJSON, err := getSecretEnv("ALCHEMY_VAULT_PRIVATE_KEYS_JSON", "")
	if err != nil {
		return cfg, err
	}
	cfg.alchemy.vaultPrivateKeys, err = parseStringMapJSON(vaultPrivateKeysJSON)
	if err != nil {
		return cfg, fmt.Errorf("invalid ALCHEMY_VAULT_PRIVATE_KEYS_JSON: %w", err)
	}
	erc20ContractsJSON := getEnv("ALCHEMY_ERC20_CONTRACTS_JSON", "")
	cfg.alchemy.erc20Contracts, err = parseStringMapJSON(erc20ContractsJSON)
	if err != nil {
		return cfg, fmt.Errorf("invalid ALCHEMY_ERC20_CONTRACTS_JSON: %w", err)
	}
	cfg.alchemy.confirmationsRequired = getEnvInt("ALCHEMY_CONFIRMATIONS_REQUIRED", 12)
	cfg.alchemy.webhookSigningSecret, err = getSecretEnv("ALCHEMY_WEBHOOK_SIGNING_SECRET", "")
	if err != nil {
		return cfg, err
	}
	cfg.alchemy.enableLiveDeposits = getEnvBool("ALCHEMY_ENABLE_LIVE_DEPOSITS", false)
	cfg.alchemy.enableLiveBroadcasts = getEnvBool("ALCHEMY_ENABLE_LIVE_BROADCASTS", false)

	if err := validateConfig(cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}

func validateConfig(cfg config) error {
	var errs []string

	if cfg.port < 1 || cfg.port > 65535 {
		errs = append(errs, "PORT must be between 1 and 65535")
	}

	if cfg.appName == "" {
		errs = append(errs, "APP_NAME must be provided")
	}

	if cfg.env != "development" && cfg.env != "staging" && cfg.env != "production" && cfg.env != "test" {
		errs = append(errs, "MY_ENV must be one of development, staging, production, test")
	}

	if cfg.db.maxOpenConns < 1 {
		errs = append(errs, "DB_MAX_OPEN_CONNS must be greater than 0")
	}

	if cfg.db.maxIdleConns < 0 {
		errs = append(errs, "DB_MAX_IDLE_CONNS must be greater than or equal to 0")
	}

	if _, err := time.ParseDuration(cfg.db.maxIdleTime); err != nil {
		errs = append(errs, "DB_MAX_IDLE_TIME must be a valid duration")
	}

	if cfg.security.rateLimitRPS <= 0 {
		errs = append(errs, "RATE_LIMIT_RPS must be greater than 0")
	}

	if cfg.security.rateLimitBurst < 1 {
		errs = append(errs, "RATE_LIMIT_BURST must be greater than 0")
	}
	if cfg.security.authRateLimitRPS <= 0 {
		errs = append(errs, "AUTH_RATE_LIMIT_RPS must be greater than 0")
	}
	if cfg.security.authRateLimitBurst < 1 {
		errs = append(errs, "AUTH_RATE_LIMIT_BURST must be greater than 0")
	}

	if cfg.security.tokenSecret == "" {
		errs = append(errs, "TOKEN_SECRET must be provided")
	}
	if cfg.env != "test" && cfg.security.tokenSecret == "replace-me-in-production" {
		errs = append(errs, "TOKEN_SECRET must be changed")
	}
	if cfg.env != "test" && !isStrongSecret(cfg.security.tokenSecret) {
		errs = append(errs, "TOKEN_SECRET is too weak; use at least 32 chars with mixed character classes")
	}
	if cfg.security.tokenTTL < 5*time.Minute {
		errs = append(errs, "TOKEN_TTL must be at least 5m")
	}
	if cfg.security.tokenTTL > 24*time.Hour {
		errs = append(errs, "TOKEN_TTL must be at most 24h")
	}
	if cfg.tracing.sampleRatio < 0 || cfg.tracing.sampleRatio > 1 {
		errs = append(errs, "OTEL_SAMPLE_RATIO must be between 0 and 1")
	}
	if cfg.alchemy.confirmationsRequired < 1 {
		errs = append(errs, "ALCHEMY_CONFIRMATIONS_REQUIRED must be at least 1")
	}
	if err := validateAlchemyBroadcastConfig(cfg); err != nil {
		errs = append(errs, err.Error())
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}

	return nil
}

func validateAlchemyBroadcastConfig(cfg config) error {
	if !cfg.alchemy.enableLiveBroadcasts {
		return nil
	}

	rpcURLs := effectiveAlchemyRPCURLs(cfg)
	if len(rpcURLs) == 0 {
		return fmt.Errorf("ALCHEMY_ENABLE_LIVE_BROADCASTS requires at least one configured RPC endpoint via ALCHEMY_RPC_URL, ALCHEMY_RPC_URLS_JSON, ALCHEMY_API_KEY, or ALCHEMY_API_KEYS_JSON")
	}

	privateKeys := effectiveAlchemyPrivateKeys(cfg)
	vaultPrivateKeys := effectiveAlchemyVaultPrivateKeys(cfg)
	missingSigner := make([]string, 0)
	for network := range rpcURLs {
		if !hasAlchemySignerForNetwork(network, privateKeys, vaultPrivateKeys) {
			missingSigner = append(missingSigner, network)
		}
	}
	if len(missingSigner) > 0 {
		sort.Strings(missingSigner)
		return fmt.Errorf("missing private key for live broadcast network(s): %s (configure ALCHEMY_VAULT_PRIVATE_KEYS_JSON, ALCHEMY_PRIVATE_KEYS_JSON, or ALCHEMY_PRIVATE_KEY for default network)", strings.Join(missingSigner, ", "))
	}

	invalidVaultSigner := make([]string, 0)
	for key, value := range cfg.alchemy.vaultPrivateKeys {
		parts := strings.Split(strings.TrimSpace(strings.ToLower(key)), ":")
		if len(parts) != 2 {
			invalidVaultSigner = append(invalidVaultSigner, fmt.Sprintf("%s (expected vault_id:network)", key))
			continue
		}
		vaultID := strings.TrimSpace(parts[0])
		network := strings.TrimSpace(parts[1])
		privateKey := strings.TrimSpace(value)
		if vaultID == "" || network == "" || privateKey == "" {
			invalidVaultSigner = append(invalidVaultSigner, fmt.Sprintf("%s (expected non-empty vault_id:network => private_key)", key))
		}
	}
	if len(invalidVaultSigner) > 0 {
		sort.Strings(invalidVaultSigner)
		return fmt.Errorf("invalid ALCHEMY_VAULT_PRIVATE_KEYS_JSON entries: %s", strings.Join(invalidVaultSigner, ", "))
	}

	invalidERC20 := make([]string, 0)
	for key, addr := range cfg.alchemy.erc20Contracts {
		parts := strings.Split(strings.TrimSpace(strings.ToLower(key)), ":")
		if len(parts) != 2 {
			invalidERC20 = append(invalidERC20, fmt.Sprintf("%s (expected network:ASSET)", key))
			continue
		}
		network := strings.TrimSpace(parts[0])
		asset := strings.TrimSpace(parts[1])
		if network == "" || asset == "" {
			invalidERC20 = append(invalidERC20, fmt.Sprintf("%s (expected network:ASSET)", key))
			continue
		}
		if !isHexAddress(strings.TrimSpace(addr)) {
			invalidERC20 = append(invalidERC20, fmt.Sprintf("%s (invalid contract address)", key))
			continue
		}
	}
	if len(invalidERC20) > 0 {
		sort.Strings(invalidERC20)
		return fmt.Errorf("invalid ALCHEMY_ERC20_CONTRACTS_JSON entries: %s", strings.Join(invalidERC20, ", "))
	}

	return nil
}

func effectiveAlchemyRPCURLs(cfg config) map[string]string {
	network := strings.ToLower(strings.TrimSpace(cfg.alchemy.network))
	if network == "" {
		network = "eth-sepolia"
	}

	rpcURLs := make(map[string]string)
	for n, url := range cfg.alchemy.rpcURLs {
		key := strings.ToLower(strings.TrimSpace(n))
		value := strings.TrimSpace(url)
		if key == "" || value == "" {
			continue
		}
		rpcURLs[key] = value
	}

	for n, apiKey := range cfg.alchemy.apiKeys {
		key := strings.ToLower(strings.TrimSpace(n))
		value := strings.TrimSpace(apiKey)
		if key == "" || value == "" {
			continue
		}
		if _, exists := rpcURLs[key]; !exists {
			rpcURLs[key] = fmt.Sprintf("https://%s.g.alchemy.com/v2/%s", key, value)
		}
	}

	rpcURL := strings.TrimSpace(cfg.alchemy.rpcURL)
	if rpcURL != "" {
		rpcURLs[network] = rpcURL
	} else if _, exists := rpcURLs[network]; !exists && strings.TrimSpace(cfg.alchemy.apiKey) != "" {
		rpcURLs[network] = fmt.Sprintf("https://%s.g.alchemy.com/v2/%s", network, strings.TrimSpace(cfg.alchemy.apiKey))
	}

	return rpcURLs
}

func effectiveAlchemyPrivateKeys(cfg config) map[string]string {
	network := strings.ToLower(strings.TrimSpace(cfg.alchemy.network))
	if network == "" {
		network = "eth-sepolia"
	}

	privateKeys := make(map[string]string)
	for n, privateKey := range cfg.alchemy.privateKeys {
		key := strings.ToLower(strings.TrimSpace(n))
		value := strings.TrimSpace(strings.TrimPrefix(privateKey, "0x"))
		if key == "" || value == "" {
			continue
		}
		privateKeys[key] = value
	}

	singlePrivateKey := strings.TrimSpace(strings.TrimPrefix(cfg.alchemy.privateKey, "0x"))
	if singlePrivateKey != "" {
		privateKeys[network] = singlePrivateKey
	}

	return privateKeys
}

func effectiveAlchemyVaultPrivateKeys(cfg config) map[string]string {
	result := make(map[string]string)
	for key, privateKey := range cfg.alchemy.vaultPrivateKeys {
		parts := strings.Split(strings.TrimSpace(strings.ToLower(key)), ":")
		if len(parts) != 2 {
			continue
		}
		vaultID := strings.TrimSpace(parts[0])
		network := strings.TrimSpace(parts[1])
		value := strings.TrimSpace(strings.TrimPrefix(privateKey, "0x"))
		if vaultID == "" || network == "" || value == "" {
			continue
		}
		result[vaultID+":"+network] = value
	}
	return result
}

func hasAlchemySignerForNetwork(network string, privateKeys, vaultPrivateKeys map[string]string) bool {
	networkKey := strings.ToLower(strings.TrimSpace(network))
	if networkKey == "" {
		return false
	}
	if strings.TrimSpace(privateKeys[networkKey]) != "" {
		return true
	}
	suffix := ":" + networkKey
	for key, privateKey := range vaultPrivateKeys {
		if strings.HasSuffix(strings.ToLower(strings.TrimSpace(key)), suffix) && strings.TrimSpace(privateKey) != "" {
			return true
		}
	}
	return false
}

func isHexAddress(value string) bool {
	if len(value) != 42 || !strings.HasPrefix(strings.ToLower(value), "0x") {
		return false
	}
	for _, r := range value[2:] {
		switch {
		case r >= '0' && r <= '9':
		case r >= 'a' && r <= 'f':
		case r >= 'A' && r <= 'F':
		default:
			return false
		}
	}
	return true
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func getEnvFloat(key string, fallback float64) float64 {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fallback
	}

	return parsed
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func getEnvBool(key string, fallback bool) bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if value == "" {
		return fallback
	}

	switch value {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return fallback
	}
}

func parseCSV(raw string) []string {
	items := strings.Split(raw, ",")
	parsed := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			parsed = append(parsed, trimmed)
		}
	}
	return parsed
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

func getSecretEnv(key, fallback string) (string, error) {
	direct := strings.TrimSpace(os.Getenv(key))
	if direct != "" {
		return direct, nil
	}

	filePath := strings.TrimSpace(os.Getenv(key + "_FILE"))
	if filePath != "" {
		value, err := readSecretFile(filePath)
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
		targetKey := strings.TrimSpace(strings.TrimPrefix(ref, "env:"))
		if targetKey == "" {
			return "", fmt.Errorf("empty env key in reference")
		}
		value := strings.TrimSpace(os.Getenv(targetKey))
		if value == "" {
			return "", fmt.Errorf("referenced env key %q is empty", targetKey)
		}
		return value, nil

	case strings.HasPrefix(ref, "file:"):
		targetPath := strings.TrimSpace(strings.TrimPrefix(ref, "file:"))
		if targetPath == "" {
			return "", fmt.Errorf("empty file path in reference")
		}
		value, err := readSecretFile(targetPath)
		if err != nil {
			return "", err
		}
		secret := strings.TrimSpace(string(value))
		if secret == "" {
			return "", fmt.Errorf("referenced secret file is empty")
		}
		return secret, nil
	}

	return "", fmt.Errorf("unsupported reference scheme, expected env: or file:")
}

func parseTrustedProxies(raw string) []*net.IPNet {
	items := parseCSV(raw)
	trusted := make([]*net.IPNet, 0, len(items))
	for _, item := range items {
		if ip := net.ParseIP(item); ip != nil {
			bits := 128
			if ip.To4() != nil {
				bits = 32
			}
			trusted = append(trusted, &net.IPNet{
				IP:   ip,
				Mask: net.CIDRMask(bits, bits),
			})
			continue
		}

		_, network, err := net.ParseCIDR(item)
		if err == nil {
			trusted = append(trusted, network)
		}
	}
	return trusted
}

func isStrongSecret(secret string) bool {
	if len(secret) < 32 {
		return false
	}

	var hasUpper, hasLower, hasDigit, hasSymbol bool
	for _, r := range secret {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		default:
			hasSymbol = true
		}
	}

	classes := 0
	if hasUpper {
		classes++
	}
	if hasLower {
		classes++
	}
	if hasDigit {
		classes++
	}
	if hasSymbol {
		classes++
	}

	return classes >= 3
}

func readSecretFile(path string) ([]byte, error) {
	clean := filepath.Clean(strings.TrimSpace(path))
	if clean == "" {
		return nil, fmt.Errorf("empty secret file path")
	}

	dir := filepath.Dir(clean)
	base := filepath.Base(clean)
	if base == "." || base == string(filepath.Separator) {
		return nil, fmt.Errorf("invalid secret file path")
	}

	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, err
	}
	defer root.Close()

	file, err := root.Open(base)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return io.ReadAll(file)
}
