// Package path contains path finding and order book queries for XRPL.
package path

import (
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	pathtypes "github.com/Peersyst/xrpl-go/xrpl/queries/path/types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// ############################################################################
// Request
// ############################################################################

// BookOffersRequest retrieves a list of offers (order book) between two currencies.
type BookOffersRequest struct {
	common.BaseRequest
	TakerGets   pathtypes.BookOfferCurrency `json:"taker_gets"`
	TakerPays   pathtypes.BookOfferCurrency `json:"taker_pays"`
	Taker       types.Address               `json:"taker,omitempty"`
	LedgerHash  common.LedgerHash           `json:"ledger_hash,omitempty"`
	LedgerIndex common.LedgerIndex          `json:"ledger_index,omitempty"`
	Limit       int                         `json:"limit,omitempty"`
	Domain      *string                     `json:"domain,omitempty"`
}

// Method returns the JSON-RPC method name for the BookOffersRequest.
func (*BookOffersRequest) Method() string {
	return "book_offers"
}

// APIVersion returns the supported API version for the BookOffersRequest.
func (*BookOffersRequest) APIVersion() int {
	return version.RippledAPIV2
}

// Validate checks that the BookOffersRequest is correctly formed.
// TODO implement V2
func (*BookOffersRequest) Validate() error {
	return nil
}

// ############################################################################
// Response
// ############################################################################

// BookOffersResponse contains the data returned by a BookOffersRequest, including the list of offers.
type BookOffersResponse struct {
	LedgerCurrentIndex common.LedgerIndex    `json:"ledger_current_index,omitempty"`
	LedgerIndex        common.LedgerIndex    `json:"ledger_index,omitempty"`
	LedgerHash         common.LedgerHash     `json:"ledger_hash,omitempty"`
	Offers             []pathtypes.BookOffer `json:"offers"`
	Validated          bool                  `json:"validated,omitempty"`
}
