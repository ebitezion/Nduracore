//revive:disable:var-naming
package types

// NFToken represents a non-fungible token with its ID and URI.
type NFToken struct {
	NFTokenID  NFTokenID
	NFTokenURI NFTokenURI `json:"URI"`
}
