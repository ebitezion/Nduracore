//revive:disable:var-naming
package types

import "github.com/Peersyst/xrpl-go/pkg/typecheck"

// CredentialIDs represents a list of credential ID strings encoded as hexadecimal values.
type CredentialIDs []string

// IsValid checks that the CredentialIDs slice is non-empty and contains only valid hex strings.
func (c CredentialIDs) IsValid() bool {
	if len(c) == 0 {
		return false
	}

	for _, id := range c {
		if !typecheck.IsHex(id) {
			return false
		}
	}

	return true
}

// Flatten returns the underlying slice of credential ID strings.
func (c CredentialIDs) Flatten() []string {
	return c
}
