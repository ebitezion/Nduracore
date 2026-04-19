// Package types contains data structures for account query types.
//
//revive:disable:var-naming
package types

import (
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// ChannelResult represents a payment channel between two accounts on the XRPL.
// It contains channel configuration, state information, and account details
// for both the source and destination of the channel.
type ChannelResult struct {
	Account            types.Address `json:"account,omitempty"`
	Amount             string        `json:"amount,omitempty"`
	Balance            string        `json:"balance,omitempty"`
	ChannelID          string        `json:"channel_id,omitempty"`
	DestinationAccount types.Address `json:"destination_account,omitempty"`
	SettleDelay        uint          `json:"settle_delay,omitempty"`
	PublicKey          string        `json:"public_key,omitempty"`
	PublicKeyHex       string        `json:"public_key_hex,omitempty"`
	Expiration         uint          `json:"expiration,omitempty"`
	CancelAfter        uint          `json:"cancel_after,omitempty"`
	SourceTag          uint          `json:"source_tag,omitempty"`
	DestinationTag     uint          `json:"destination_tag,omitempty"`
}
