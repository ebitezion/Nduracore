package v1

import (
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// ############################################################################
// Request
// ############################################################################

// NFTInfoRequest retrieves information about an NFToken via the nft_info method.
type NFTInfoRequest struct {
	common.BaseRequest
	NFTokenID   types.NFTokenID        `json:"nft_id"`
	LedgerHash  common.LedgerHash      `json:"ledger_hash,omitempty"`
	LedgerIndex common.LedgerSpecifier `json:"ledger_index,omitempty"`
}

// Method returns the RPC method name for the NFTInfoRequest.
func (*NFTInfoRequest) Method() string {
	return "nft_info"
}

// APIVersion returns the API version for the NFTInfoRequest.
func (*NFTInfoRequest) APIVersion() int {
	return version.RippledAPIV1
}

// Validate checks the NFTInfoRequest for correctness.
// TODO: Implement V2
func (*NFTInfoRequest) Validate() error {
	return nil
}

// ############################################################################
// Response
// ############################################################################

// NFTInfoResponse represents the response returned by the nft_info method.
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
