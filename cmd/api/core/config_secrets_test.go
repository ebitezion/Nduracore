package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetSecretEnvPrefersDirectEnv(t *testing.T) {
	t.Setenv("TOKEN_SECRET", "direct-value")
	t.Setenv("TOKEN_SECRET_FILE", "")
	t.Setenv("TOKEN_SECRET_REF", "")

	secret, err := getSecretEnv("TOKEN_SECRET", "fallback")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if secret != "direct-value" {
		t.Fatalf("expected direct-value, got %q", secret)
	}
}

func TestGetSecretEnvReadsFile(t *testing.T) {
	dir := t.TempDir()
	secretPath := filepath.Join(dir, "secret.txt")
	if err := os.WriteFile(secretPath, []byte("file-value\n"), 0o600); err != nil {
		t.Fatalf("write secret file: %v", err)
	}

	t.Setenv("DB_DSN", "")
	t.Setenv("DB_DSN_FILE", secretPath)
	t.Setenv("DB_DSN_REF", "")

	secret, err := getSecretEnv("DB_DSN", "fallback")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if secret != "file-value" {
		t.Fatalf("expected file-value, got %q", secret)
	}
}

func TestGetSecretEnvReadsEnvReference(t *testing.T) {
	t.Setenv("TOKEN_SECRET", "")
	t.Setenv("TOKEN_SECRET_FILE", "")
	t.Setenv("TOKEN_SECRET_REF", "env:UPSTREAM_SECRET")
	t.Setenv("UPSTREAM_SECRET", "ref-value")

	secret, err := getSecretEnv("TOKEN_SECRET", "fallback")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if secret != "ref-value" {
		t.Fatalf("expected ref-value, got %q", secret)
	}
}

func TestGetSecretEnvReturnsErrorForInvalidReference(t *testing.T) {
	t.Setenv("TOKEN_SECRET", "")
	t.Setenv("TOKEN_SECRET_FILE", "")
	t.Setenv("TOKEN_SECRET_REF", "vault:secret/path")

	_, err := getSecretEnv("TOKEN_SECRET", "fallback")
	if err == nil {
		t.Fatal("expected error for invalid secret reference")
	}
	if !strings.Contains(err.Error(), "unsupported reference scheme") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetSecretEnvReturnsFallback(t *testing.T) {
	t.Setenv("REDIS_PASSWORD", "")
	t.Setenv("REDIS_PASSWORD_FILE", "")
	t.Setenv("REDIS_PASSWORD_REF", "")

	secret, err := getSecretEnv("REDIS_PASSWORD", "fallback-value")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if secret != "fallback-value" {
		t.Fatalf("expected fallback-value, got %q", secret)
	}
}

func TestParseStringMapJSON(t *testing.T) {
	result, err := parseStringMapJSON(`{"eth-mainnet":"https://mainnet"," eth-sepolia ":"https://sepolia"}`)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result["eth-mainnet"] != "https://mainnet" {
		t.Fatalf("unexpected mainnet value: %q", result["eth-mainnet"])
	}
	if result["eth-sepolia"] != "https://sepolia" {
		t.Fatalf("unexpected sepolia value: %q", result["eth-sepolia"])
	}
}

func TestParseStringMapJSONInvalid(t *testing.T) {
	_, err := parseStringMapJSON(`{"eth-mainnet":}`)
	if err == nil {
		t.Fatal("expected error for invalid json")
	}
}
