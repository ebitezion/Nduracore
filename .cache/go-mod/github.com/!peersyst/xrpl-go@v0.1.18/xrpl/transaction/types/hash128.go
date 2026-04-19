//revive:disable:var-naming
package types

// Hash128 represents a 128-bit hash encoded as a string.
type Hash128 string

// String returns the string representation of the Hash128.
func (h *Hash128) String() string {
	return string(*h)
}
