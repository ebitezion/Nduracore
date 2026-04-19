//revive:disable:var-naming
package types

// Hash192 represents a 192-bit hash encoded as a string.
type Hash192 string

// String returns the string representation of the Hash192.
func (h *Hash192) String() string {
	return string(*h)
}
