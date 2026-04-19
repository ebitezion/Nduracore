package clio

import (
	cliotypes "github.com/Peersyst/xrpl-go/xrpl/queries/clio/types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// ############################################################################
// Request
// ############################################################################

// NFTsByIssuerRequest returns a list of NFToken objects issued by the specified account.
// The order of the NFTs is not associated with their mint date.
type NFTsByIssuerRequest struct {
	common.BaseRequest
	Issuer   types.Address `json:"issuer"`
	Marker   any           `json:"marker,omitempty"`
	Limit    int           `json:"limit,omitempty"`
	NftTaxon uint32        `json:"nft_taxon,omitempty"`
}

// Method returns the JSON-RPC method name for NFTsByIssuerRequest.
func (*NFTsByIssuerRequest) Method() string {
	return "nfts_by_issuer"
}

// APIVersion returns the Rippled API version for NFTsByIssuerRequest.
func (*NFTsByIssuerRequest) APIVersion() int {
	return version.RippledAPIV2
}

// Validate checks the NFTsByIssuerRequest for valid parameters.
// TODO implement V2
func (*NFTsByIssuerRequest) Validate() error {
	return nil
}

// ############################################################################
// Response
// ############################################################################

// NFTsByIssuerResponse is the response returned by the nfts_by_issuer method, containing issued NFToken data.
type NFTsByIssuerResponse struct {
	Issuer       types.Address       `json:"issuer"`
	NFTs         []cliotypes.NFToken `json:"nfts"`
	Marker       any                 `json:"marker,omitempty"`
	Limit        int                 `json:"limit,omitempty"`
	NFTokenTaxon uint32              `json:"nft_taxon,omitempty"`
}
