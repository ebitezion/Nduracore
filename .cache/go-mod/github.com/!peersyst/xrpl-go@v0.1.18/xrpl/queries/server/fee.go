package server

import (
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	servertypes "github.com/Peersyst/xrpl-go/xrpl/queries/server/types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
)

// ############################################################################
// Request
// ############################################################################

// FeeRequest is the request type for the fee command.
// It reports the current state of the open-ledger requirements for the
// transaction cost. This requires the FeeEscalation amendment to be enabled.
type FeeRequest struct {
	common.BaseRequest
}

// Method returns the JSON-RPC method name for the FeeRequest.
func (*FeeRequest) Method() string {
	return "fee"
}

// APIVersion returns the API version required by the FeeRequest.
func (*FeeRequest) APIVersion() int {
	return version.RippledAPIV2
}

// Validate verifies the FeeRequest parameters.
// TODO: implement V2 validation logic.
func (*FeeRequest) Validate() error {
	return nil
}

// ############################################################################
// Response
// ############################################################################

// FeeResponse is the response type returned by the fee command.
type FeeResponse struct {
	CurrentLedgerSize  string                `json:"current_ledger_size"`
	CurrentQueueSize   string                `json:"current_queue_size"`
	Drops              servertypes.FeeDrops  `json:"drops"`
	ExpectedLedgerSize string                `json:"expected_ledger_size"`
	LedgerCurrentIndex common.LedgerIndex    `json:"ledger_current_index"`
	Levels             servertypes.FeeLevels `json:"levels"`
	MaxQueueSize       string                `json:"max_queue_size"`
}
