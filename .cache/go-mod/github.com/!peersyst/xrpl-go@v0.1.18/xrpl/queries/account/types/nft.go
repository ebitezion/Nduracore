//revive:disable:var-naming
package types

import (
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// NFTokenFlag bits control NFT behavior.
const (
	// Burnable indicates the NFT can be burned.
	Burnable NFTokenFlag = 0x0001
	// OnlyXRP restricts the NFT to XRP-only offers.
	OnlyXRP NFTokenFlag = 0x0002
	// Transferable allows the NFT to be transferred freely.
	Transferable NFTokenFlag = 0x0008
	// ReservedFlag is reserved for future use.
	ReservedFlag NFTokenFlag = 0x8000
)

// NFTokenFlag represents a set of bit flags controlling NFT behavior.
type NFTokenFlag uint32

// NFT represents a non-fungible token entry, including identifiers and flags.
type NFT struct {
	Flags        NFTokenFlag `json:",omitempty"`
	Issuer       types.Address
	NFTokenID    types.NFTokenID
	NFTokenTaxon uint
	URI          types.NFTokenURI `json:",omitempty"`
	NFTSerial    uint             `json:"nft_serial"`
}
