package v1

import (
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
)

// ############################################################################
// Request
// ############################################################################

// ClosedRequest is the request type for the ledger_closed method.
// It returns the unique identifiers of the most recently closed ledger.
type ClosedRequest struct {
	common.BaseRequest
}

// Method returns the JSON-RPC method name for the ClosedRequest.
func (*ClosedRequest) Method() string {
	return "ledger_closed"
}

// APIVersion returns the API version for the ClosedRequest.
func (*ClosedRequest) APIVersion() int {
	return version.RippledAPIV1
}

// Validate checks that the ClosedRequest parameters are valid. TODO: implement V2 validation.
func (*ClosedRequest) Validate() error {
	return nil
}

// ############################################################################
// Response
// ############################################################################

// ClosedResponse is the response type for the ledger_closed method.
type ClosedResponse struct {
	LedgerHash  string             `json:"ledger_hash"`
	LedgerIndex common.LedgerIndex `json:"ledger_index"`
}
