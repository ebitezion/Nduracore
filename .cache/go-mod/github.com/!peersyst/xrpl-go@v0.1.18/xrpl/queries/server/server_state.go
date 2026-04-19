package server

import (
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	servertypes "github.com/Peersyst/xrpl-go/xrpl/queries/server/types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
)

// ############################################################################
// Request
// ############################################################################

// StateRequest is the request type for the server_state command.
// It asks the server for various machine-readable information about the
// rippled server's current state. The response is almost the same as the
// server_info method, but uses units that are easier to process instead of
// easier to read.
type StateRequest struct {
	common.BaseRequest
}

// Method returns the JSON-RPC method name for the StateRequest.
func (*StateRequest) Method() string {
	return "server_state"
}

// APIVersion returns the API version required by the StateRequest.
func (*StateRequest) APIVersion() int {
	return version.RippledAPIV2
}

// Validate verifies the StateRequest parameters.
// TODO: implement V2 validation logic.
func (*StateRequest) Validate() error {
	return nil
}

// ############################################################################
// Response
// ############################################################################

// StateResponse is the response type returned by the server_state command.
type StateResponse struct {
	State servertypes.State `json:"state"`
}
