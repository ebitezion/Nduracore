package clio

import (
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// ############################################################################
// Request
// ############################################################################

// NFTInfoRequest retrieves information about an NFToken.
type NFTInfoRequest struct {
	common.BaseRequest
	NFTokenID   types.NFTokenID        `json:"nft_id"`
	LedgerHash  common.LedgerHash      `json:"ledger_hash,omitempty"`
	LedgerIndex common.LedgerSpecifier `json:"ledger_index,omitempty"`
}

// Method returns the JSON-RPC method name for NFTInfoRequest.
func (*NFTInfoRequest) Method() string {
	return "nft_info"
}

// APIVersion returns the Rippled API version for NFTInfoRequest.
func (*NFTInfoRequest) APIVersion() int {
	return version.RippledAPIV2
}

// Validate checks the NFTInfoRequest parameters for validity.
func (*NFTInfoRequest) Validate() error {
	return nil
}

// ############################################################################
// Response
// ############################################################################

// NFTInfoResponse is the response returned by the nft_info method, containing NFToken details.
type NFTInfoResponse struct {
	NFTokenID       types.NFTokenID    `json:"nft_id"`
	LedgerIndex     common.LedgerIndex `json:"ledger_index"`
	Owner           types.Address      `json:"owner"`
	IsBurned        bool               `json:"is_burned"`
	Flags           uint               `json:"flags"`
	TransferFee     uint               `json:"transfer_fee"`
	Issuer          types.Address      `json:"issuer"`
	NFTokenTaxon    uint               `json:"nft_taxon"`
	NFTokenSequence uint               `json:"nft_sequence"`
	URI             types.NFTokenURI   `json:"uri"`
}
