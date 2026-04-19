// Package channel provides commands to query XRPL payment channel methods.
package channel

import (
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// ############################################################################
// Request
// ############################################################################

// VerifyRequest checks the validity of a signature that can be used to redeem a specific amount of XRP from a payment channel.
type VerifyRequest struct {
	common.BaseRequest
	Amount    types.XRPCurrencyAmount `json:"amount"`
	ChannelID string                  `json:"channel_id"`
	PublicKey string                  `json:"public_key"`
	Signature string                  `json:"signature"`
}

// Method returns the XRPL JSON-RPC method name for VerifyRequest.
func (*VerifyRequest) Method() string {
	return "channel_verify"
}

// APIVersion returns the XRPL API version for VerifyRequest.
func (*VerifyRequest) APIVersion() int {
	return version.RippledAPIV2
}

// Validate ensures the VerifyRequest is valid.
// TODO implement v2
func (*VerifyRequest) Validate() error {
	return nil
}

// ############################################################################
// Response
// ############################################################################

// VerifyResponse is the expected response from the channel_verify method.
type VerifyResponse struct {
	SignatureVerified bool `json:"signature_verified"`
}
