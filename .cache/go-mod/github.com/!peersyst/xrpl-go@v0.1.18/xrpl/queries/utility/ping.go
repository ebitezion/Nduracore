// Package utility provides commands to query XRPL utility methods.
package utility

import (
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
)

// ############################################################################
// Request
// ############################################################################

// PingRequest returns an acknowledgement so that clients can test the connection
// status and latency.
type PingRequest struct {
	common.BaseRequest
}

// Method returns the XRPL JSON-RPC method name for PingRequest.
func (*PingRequest) Method() string {
	return "ping"
}

// APIVersion returns the XRPL API version for PingRequest.
func (*PingRequest) APIVersion() int {
	return version.RippledAPIV2
}

// Validate ensures the PingRequest is valid.
func (*PingRequest) Validate() error {
	return nil
}

// ############################################################################
// Response
// ############################################################################

// PingResponse is the expected response from the ping method.
type PingResponse struct {
	Role      string `json:"role,omitempty"`
	Unlimited bool   `json:"unlimited,omitempty"`
}
