package server

import (
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	servertypes "github.com/Peersyst/xrpl-go/xrpl/queries/server/types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
)

// ############################################################################
// Request
// ############################################################################

// InfoRequest represents a server_info request for human-readable server information.
type InfoRequest struct {
	common.BaseRequest
}

// Method returns the JSON-RPC method name for the InfoRequest.
func (*InfoRequest) Method() string {
	return "server_info"
}

// APIVersion returns the supported API version for the InfoRequest.
func (*InfoRequest) APIVersion() int {
	return version.RippledAPIV2
}

// Validate checks that the InfoRequest is correctly formed.
// // TODO implement V2
func (*InfoRequest) Validate() error {
	return nil
}

// ############################################################################
// Response
// ############################################################################

// InfoResponse represents the expected response from the server_info method.
type InfoResponse struct {
	Info servertypes.Info `json:"info"`
}
