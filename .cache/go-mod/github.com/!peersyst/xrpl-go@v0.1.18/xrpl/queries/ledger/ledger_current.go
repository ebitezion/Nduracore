package ledger

import (
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
)

// ############################################################################
// Request
// ############################################################################

// CurrentRequest returns the unique identifiers of the current in-progress ledger.
type CurrentRequest struct {
	common.BaseRequest
}

// Method returns the JSON-RPC method name for CurrentRequest.
func (*CurrentRequest) Method() string {
	return "ledger_current"
}

// APIVersion returns the Rippled API version for CurrentRequest.
func (*CurrentRequest) APIVersion() int {
	return version.RippledAPIV2
}

// Validate checks the CurrentRequest for valid parameters.
// TODO implement V2
func (*CurrentRequest) Validate() error {
	return nil
}

// ############################################################################
// Response
// ############################################################################

// CurrentResponse is the response returned by the ledger_current method.
type CurrentResponse struct {
	LedgerCurrentIndex common.LedgerIndex `json:"ledger_current_index"`
}
