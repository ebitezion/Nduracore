package utility

import (
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// ############################################################################
// Request
// ############################################################################

// RandomRequest provides a random number to be used as a source of
// entropy for random number generation by clients.
type RandomRequest struct {
	common.BaseRequest
}

// Method returns the XRPL JSON-RPC method name for RandomRequest.
func (*RandomRequest) Method() string {
	return "random"
}

// APIVersion returns the XRPL API version for RandomRequest.
func (*RandomRequest) APIVersion() int {
	return version.RippledAPIV2
}

// Validate ensures the RandomRequest is valid.
func (*RandomRequest) Validate() error {
	return nil
}

// ############################################################################
// Response
// ############################################################################

// RandomResponse is the expected response from the random method.
type RandomResponse struct {
	Random types.Hash256 `json:"random"`
}
