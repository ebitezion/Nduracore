package v1

import (
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// ############################################################################
// Request
// ############################################################################

// NoRippleCheckRequest provides a quick way to check the status of the
// default ripple field for an account and the No Ripple flag of its trust
// lines, compared with the recommended settings. Expects a response in the form
// of an NoRippleCheckResponse.
type NoRippleCheckRequest struct {
	common.BaseRequest
	Account      types.Address          `json:"account"`
	Role         string                 `json:"role"`
	Transactions bool                   `json:"transactions,omitempty"`
	Limit        int                    `json:"limit,omitempty"`
	LedgerHash   common.LedgerHash      `json:"ledger_hash,omitempty"`
	LedgerIndex  common.LedgerSpecifier `json:"ledger_index,omitempty"`
}

// Method returns the JSON-RPC method name for NoRippleCheckRequest.
func (*NoRippleCheckRequest) Method() string {
	return "noripple_check"
}

// APIVersion returns the Rippled API version for NoRippleCheckRequest.
func (*NoRippleCheckRequest) APIVersion() int {
	return version.RippledAPIV1
}

// Validate checks the NoRippleCheckRequest for valid parameters.
func (*NoRippleCheckRequest) Validate() error {
	return nil
}

// ############################################################################
// Response
// ############################################################################

// NoRippleCheckResponse is the response returned by NoRippleCheckRequest.
type NoRippleCheckResponse struct {
	LedgerCurrentIndex common.LedgerIndex            `json:"ledger_current_index"`
	Problems           []string                      `json:"problems"`
	Transactions       []transaction.FlatTransaction `json:"transactions"`
}
