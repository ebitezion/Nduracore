//revive:disable:var-naming
package types

import "github.com/Peersyst/xrpl-go/xrpl/transaction/types"

// RipplePathFindCurrency represents a currency and optional issuer used in a path find request.
type RipplePathFindCurrency struct {
	Currency string        `json:"currency"`
	Issuer   types.Address `json:"issuer,omitempty"`
}
