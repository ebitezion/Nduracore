//revive:disable:var-naming
package types

import (
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// NFTokenOffer represents an offer for a non-fungible token, including amount,
// owner, destination, flags, and expiration details.
type NFTokenOffer struct {
	Amount            any           `json:"amount"`
	Flags             uint          `json:"flags"`
	NFTokenOfferIndex string        `json:"nft_offer_index"`
	Owner             types.Address `json:"owner"`
	Destination       types.Address `json:"destination,omitempty"`
	Expiration        int           `json:"expiration,omitempty"`
}
