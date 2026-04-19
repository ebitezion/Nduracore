package account

import (
	accounttypes "github.com/Peersyst/xrpl-go/xrpl/queries/account/types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// ############################################################################
// Request
// ############################################################################

// NFTsRequest retrieves all NFTs currently owned by the specified account for a given ledger version.
type NFTsRequest struct {
	common.BaseRequest
	Account     types.Address          `json:"account"`
	LedgerIndex common.LedgerSpecifier `json:"ledger_index,omitempty"`
	LedgerHash  common.LedgerHash      `json:"ledger_hash,omitempty"`
	Limit       int                    `json:"limit,omitempty"`
	Marker      any                    `json:"marker,omitempty"`
}

// Method returns the JSON-RPC method name for NFTsRequest.
func (*NFTsRequest) Method() string {
	return "account_nfts"
}

// APIVersion returns the API version supported by NFTsRequest.
func (*NFTsRequest) APIVersion() int {
	return version.RippledAPIV2
}

// Validate performs validation on NFTsRequest.
// TODO: implement V2.
func (*NFTsRequest) Validate() error {
	return nil
}

// ############################################################################
// Response
// ############################################################################

// NFTsResponse represents the response from account_nfts, including pagination and NFT list.
type NFTsResponse struct {
	Account            types.Address      `json:"account"`
	AccountNFTs        []accounttypes.NFT `json:"account_nfts"`
	LedgerIndex        common.LedgerIndex `json:"ledger_index,omitempty"`
	LedgerHash         common.LedgerHash  `json:"ledger_hash,omitempty"`
	LedgerCurrentIndex common.LedgerIndex `json:"ledger_current_index,omitempty"`
	Validated          bool               `json:"validated"`
	Marker             any                `json:"marker,omitempty"`
	Limit              int                `json:"limit,omitempty"`
}
