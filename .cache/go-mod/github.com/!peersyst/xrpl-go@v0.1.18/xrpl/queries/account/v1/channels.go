// Package v1 contains version 1 account queries for XRPL.
package v1

import (
	accounttypes "github.com/Peersyst/xrpl-go/xrpl/queries/account/types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// ############################################################################
// Request
// ############################################################################

// ChannelsRequest returns information about an account's Payment Channels where the specified account is the source.
// Only channels where the account is the channel's source are included.
// All information is relative to a specific ledger version.
type ChannelsRequest struct {
	common.BaseRequest
	Account            types.Address          `json:"account"`
	DestinationAccount types.Address          `json:"destination_account,omitempty"`
	LedgerIndex        common.LedgerSpecifier `json:"ledger_index,omitempty"`
	LedgerHash         common.LedgerHash      `json:"ledger_hash,omitempty"`
	Limit              int                    `json:"limit,omitempty"`
	Marker             any                    `json:"marker,omitempty"`
}

// Method returns the method name for the ChannelsRequest.
func (*ChannelsRequest) Method() string {
	return "account_channels"
}

// APIVersion returns the Rippled API version for ChannelsRequest.
func (*ChannelsRequest) APIVersion() int {
	return version.RippledAPIV1
}

// Validate method to be added to each request struct
func (r *ChannelsRequest) Validate() error {
	if r.Account == "" {
		return accounttypes.ErrNoAccountID
	}

	return nil
}

// ############################################################################
// Response
// ############################################################################

// ChannelsResponse is the response returned by the account_channels method.
type ChannelsResponse struct {
	Account     types.Address                `json:"account"`
	Channels    []accounttypes.ChannelResult `json:"channels"`
	LedgerIndex common.LedgerIndex           `json:"ledger_index,omitempty"`
	LedgerHash  common.LedgerHash            `json:"ledger_hash,omitempty"`
	Validated   bool                         `json:"validated,omitempty"`
	Limit       int                          `json:"limit,omitempty"`
	Marker      any                          `json:"marker,omitempty"`
}
