//revive:disable:var-naming
package types

// NFTokenURI represents the URI associated with a non-fungible token.
type NFTokenURI string

// String returns the string representation of a NFTokenURI.
func (n *NFTokenURI) String() string {
	return string(*n)
}
