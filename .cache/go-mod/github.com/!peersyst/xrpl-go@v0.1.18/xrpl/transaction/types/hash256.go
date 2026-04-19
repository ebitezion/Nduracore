//revive:disable:var-naming
package types

// Hash256 represents a 256-bit hash encoded as a string.
type Hash256 string

// String returns the string representation of the Hash256.
func (h *Hash256) String() string {
	return string(*h)
}
