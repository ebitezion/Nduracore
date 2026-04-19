package v1

import (
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// ############################################################################
// Request
// ############################################################################

// DepositAuthorizedRequest is the request type for the deposit_authorized command.
// It indicates whether one account is authorized to send payments directly to another.
type DepositAuthorizedRequest struct {
	common.BaseRequest
	SourceAccount      types.Address          `json:"source_account"`
	DestinationAccount types.Address          `json:"destination_account"`
	LedgerHash         common.LedgerHash      `json:"ledger_hash,omitempty"`
	LedgerIndex        common.LedgerSpecifier `json:"ledger_index,omitempty"`
}

// Method returns the JSON-RPC method name for the DepositAuthorizedRequest.
func (*DepositAuthorizedRequest) Method() string {
	return "deposit_authorized"
}

// APIVersion returns the API version required by the DepositAuthorizedRequest.
func (*DepositAuthorizedRequest) APIVersion() int {
	return version.RippledAPIV1
}

// Validate verifies the DepositAuthorizedRequest parameters.
// TODO: implement V2.
func (*DepositAuthorizedRequest) Validate() error {
	return nil
}

// ############################################################################
// Response
// ############################################################################

// DepositAuthorizedResponse is the response type returned by the deposit_authorized command.
// It indicates whether deposit is authorized along with ledger and account details.
type DepositAuthorizedResponse struct {
	DepositAuthorized  bool               `json:"deposit_authorized"`
	DestinationAccount types.Address      `json:"destination_account"`
	LedgerHash         common.LedgerHash  `json:"ledger_hash,omitempty"`
	LedgerIndex        common.LedgerIndex `json:"ledger_index,omitempty"`
	LedgerCurrentIndex common.LedgerIndex `json:"ledger_current_index,omitempty"`
	SourceAccount      types.Address      `json:"source_account"`
	Validated          bool               `json:"validated,omitempty"`
}
