// Package v1 contains version 1 path finding queries for XRPL.
package v1

import (
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	pathtypes "github.com/Peersyst/xrpl-go/xrpl/queries/path/types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// ############################################################################
// Request
// ############################################################################

// BookOffersRequest is the request type for the book_offers command.
// It retrieves a list of order book offers between two currencies.
type BookOffersRequest struct {
	common.BaseRequest
	TakerGets   pathtypes.BookOfferCurrency `json:"taker_gets"`
	TakerPays   pathtypes.BookOfferCurrency `json:"taker_pays"`
	Taker       types.Address               `json:"taker,omitempty"`
	LedgerHash  common.LedgerHash           `json:"ledger_hash,omitempty"`
	LedgerIndex common.LedgerIndex          `json:"ledger_index,omitempty"`
	Limit       int                         `json:"limit,omitempty"`
}

// Method returns the JSON-RPC method name for the BookOffersRequest.
func (*BookOffersRequest) Method() string {
	return "book_offers"
}

// APIVersion returns the API version required by the BookOffersRequest.
func (*BookOffersRequest) APIVersion() int {
	return version.RippledAPIV1
}

// Validate verifies the BookOffersRequest parameters.
// TODO: implement V2.
func (*BookOffersRequest) Validate() error {
	return nil
}

// ############################################################################
// Response
// ############################################################################

// BookOffersResponse is the response type returned by the book_offers command.
// It contains ledger details and a list of offers matched.
type BookOffersResponse struct {
	LedgerCurrentIndex common.LedgerIndex    `json:"ledger_current_index,omitempty"`
	LedgerIndex        common.LedgerIndex    `json:"ledger_index,omitempty"`
	LedgerHash         common.LedgerHash     `json:"ledger_hash,omitempty"`
	Offers             []pathtypes.BookOffer `json:"offers"`
	Validated          bool                  `json:"validated,omitempty"`
}
