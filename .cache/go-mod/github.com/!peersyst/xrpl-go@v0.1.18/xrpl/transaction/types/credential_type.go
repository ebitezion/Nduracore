package types

import "github.com/Peersyst/xrpl-go/pkg/typecheck"

const (
	// MinCredentialTypeLength is the minimum length of a credential type in hex characters (1 byte = 2 hex characters).
	MinCredentialTypeLength = 2

	// MaxCredentialTypeLength is the maximum length of a credential type in hex characters (64 bytes = 128 hex characters).
	MaxCredentialTypeLength = 128
)

// CredentialType represents a credential type encoded as a hexadecimal string.
type CredentialType string

// String returns the string representation of a CredentialType.
func (c *CredentialType) String() string {
	return string(*c)
}

// IsValid checks if the credential type meets all requirements: not empty, valid hex string, and length between MinCredentialTypeLength and MaxCredentialTypeLength.
func (c *CredentialType) IsValid() bool {
	if c.String() == "" {
		return false
	}

	credTypeStr := c.String()
	if !typecheck.IsHex(credTypeStr) {
		return false
	}

	length := len(credTypeStr)
	return length >= MinCredentialTypeLength && length <= MaxCredentialTypeLength
}
