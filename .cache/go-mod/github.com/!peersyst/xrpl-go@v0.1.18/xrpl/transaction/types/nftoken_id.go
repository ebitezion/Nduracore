//revive:disable:var-naming
package types

// NFTokenID represents the identifier of a non-fungible token.
type NFTokenID Hash256

// String returns the string representation of a NFTokenID.
func (n *NFTokenID) String() string {
	return string(*n)
}
