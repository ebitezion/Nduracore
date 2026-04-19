package path

import (
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// ############################################################################
// Request
// ############################################################################

// DepositAuthorizedRequest indicates whether one account is authorized to send payments directly to another.
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

// APIVersion returns the supported API version for the DepositAuthorizedRequest.
func (*DepositAuthorizedRequest) APIVersion() int {
	return version.RippledAPIV2
}

// Validate checks that the DepositAuthorizedRequest is correctly formed.
// TODO implement V2
func (*DepositAuthorizedRequest) Validate() error {
	return nil
}

// ############################################################################
// Response
// ############################################################################

// DepositAuthorizedResponse represents the expected response from the deposit_authorized method.
type DepositAuthorizedResponse struct {
	DepositAuthorized  bool               `json:"deposit_authorized"`
	DestinationAccount types.Address      `json:"destination_account"`
	LedgerHash         common.LedgerHash  `json:"ledger_hash,omitempty"`
	LedgerIndex        common.LedgerIndex `json:"ledger_index,omitempty"`
	LedgerCurrentIndex common.LedgerIndex `json:"ledger_current_index,omitempty"`
	SourceAccount      types.Address      `json:"source_account"`
	Validated          bool               `json:"validated,omitempty"`
}
