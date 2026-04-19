// Package v1 contains version 1 payment channel queries for XRPL.
package v1

import (
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// ############################################################################
// Request
// ############################################################################

// VerifyRequest represents a request to check a signature used to redeem XRP from a payment channel.
type VerifyRequest struct {
	common.BaseRequest
	Amount    types.XRPCurrencyAmount `json:"amount"`
	ChannelID string                  `json:"channel_id"`
	PublicKey string                  `json:"public_key"`
	Signature string                  `json:"signature"`
}

// Method returns the JSON-RPC method name for the VerifyRequest.
func (*VerifyRequest) Method() string {
	return "channel_verify"
}

// APIVersion returns the supported API version for the VerifyRequest.
func (*VerifyRequest) APIVersion() int {
	return version.RippledAPIV1
}

// Validate checks that the VerifyRequest is correctly formed.
// TODO implement V2
func (*VerifyRequest) Validate() error {
	return nil
}

// ############################################################################
// Response
// ############################################################################

// VerifyResponse contains the result of a channel_verify request indicating if the signature was verified.
type VerifyResponse struct {
	SignatureVerified bool `json:"signature_verified"`
}
