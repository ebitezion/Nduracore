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

// OffersRequest retrieves a list of outstanding offers made by a given account at a specific ledger version.
type OffersRequest struct {
	common.BaseRequest
	Account     types.Address          `json:"account"`
	LedgerHash  common.LedgerHash      `json:"ledger_hash,omitempty"`
	LedgerIndex common.LedgerSpecifier `json:"ledger_index,omitempty"`
	Limit       int                    `json:"limit,omitempty"`
	Marker      any                    `json:"marker,omitempty"`
	Strict      bool                   `json:"strict,omitempty"`
}

// Method returns the RPC method name for the OffersRequest.
func (*OffersRequest) Method() string {
	return "account_offers"
}

// APIVersion returns the API version for the OffersRequest.
func (*OffersRequest) APIVersion() int {
	return version.RippledAPIV2
}

// Validate checks the OffersRequest for correctness.
// TODO: Implement (V2)
func (*OffersRequest) Validate() error {
	return nil
}

// OffersResponse represents the response returned by the account_offers method.
type OffersResponse struct {
	Account            types.Address              `json:"account"`
	Offers             []accounttypes.OfferResult `json:"offers"`
	LedgerCurrentIndex common.LedgerIndex         `json:"ledger_current_index,omitempty"`
	LedgerIndex        common.LedgerIndex         `json:"ledger_index,omitempty"`
	LedgerHash         common.LedgerHash          `json:"ledger_hash,omitempty"`
	Marker             any                        `json:"marker,omitempty"`
}
