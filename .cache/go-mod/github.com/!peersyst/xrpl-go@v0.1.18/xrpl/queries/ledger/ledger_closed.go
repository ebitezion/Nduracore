package ledger

import (
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
)

// ############################################################################
// Request
// ############################################################################

// ClosedRequest returns the unique identifiers of the most recently closed ledger.
type ClosedRequest struct {
	common.BaseRequest
}

// Method returns the JSON-RPC method name for ClosedRequest.
func (*ClosedRequest) Method() string {
	return "ledger_closed"
}

// APIVersion returns the Rippled API version for ClosedRequest.
func (*ClosedRequest) APIVersion() int {
	return version.RippledAPIV2
}

// Validate checks the ClosedRequest for valid parameters.
// TODO implement V2
func (*ClosedRequest) Validate() error {
	return nil
}

// ############################################################################
// Response
// ############################################################################

// ClosedResponse is the response returned by the ledger_closed method.
type ClosedResponse struct {
	LedgerHash  string             `json:"ledger_hash"`
	LedgerIndex common.LedgerIndex `json:"ledger_index"`
}
