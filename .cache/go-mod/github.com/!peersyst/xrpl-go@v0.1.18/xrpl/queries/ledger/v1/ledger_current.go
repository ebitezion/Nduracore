package v1

import (
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
)

// ############################################################################
// Request
// ############################################################################

// CurrentRequest is the request type for the ledger_current method.
// It returns the unique identifiers of the current in-progress ledger.
type CurrentRequest struct {
	common.BaseRequest
}

// Method returns the JSON-RPC method name for the CurrentRequest.
func (*CurrentRequest) Method() string {
	return "ledger_current"
}

// APIVersion returns the API version for the CurrentRequest.
func (*CurrentRequest) APIVersion() int {
	return version.RippledAPIV1
}

// Validate checks that the CurrentRequest parameters are valid.
// TODO: implement V2 validation.
func (*CurrentRequest) Validate() error {
	return nil
}

// ############################################################################
// Response
// ############################################################################

// CurrentResponse is the response type for the ledger_current method.
type CurrentResponse struct {
	LedgerCurrentIndex common.LedgerIndex `json:"ledger_current_index"`
}
