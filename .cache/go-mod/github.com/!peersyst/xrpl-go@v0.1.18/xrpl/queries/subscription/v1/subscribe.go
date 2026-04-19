// Package v1 contains version 1 subscription functionality for XRPL streams.
package v1

import (
	"github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	streamtypes "github.com/Peersyst/xrpl-go/xrpl/queries/subscription/v1/types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// ############################################################################
// Request
// ############################################################################

// Request subscribes to specified streams, accounts, or order books for periodic notifications.
type Request struct {
	Streams          []string                `json:"streams,omitempty"`
	Accounts         []types.Address         `json:"accounts,omitempty"`
	AccountsProposed []types.Address         `json:"accounts_proposed,omitempty"`
	Books            []streamtypes.OrderBook `json:"books,omitempty"`
	URL              string                  `json:"url,omitempty"`
	URLUsername      string                  `json:"url_username,omitempty"`
	URLPassword      string                  `json:"url_password,omitempty"`
}

// Method returns the JSON-RPC method name for Request.
func (*Request) Method() string {
	return "subscribe"
}

// Validate performs validation on Request.
// TODO: implement V2.
func (*Request) Validate() error {
	return nil
}

// APIVersion returns the API version supported by Request.
func (*Request) APIVersion() int {
	return version.RippledAPIV1
}

// ############################################################################
// Response
// ############################################################################

// Response represents the response from the subscribe method, including server status and ledger information.
type Response struct {
	LoadBase         uint               `json:"load_base,omitempty"`
	LoadFactor       uint               `json:"load_factor,omitempty"`
	Random           string             `json:"random,omitempty"`
	ServerStatus     string             `json:"server_status,omitempty"`
	FeeBase          uint               `json:"fee_base,omitempty"`
	FeeRef           uint               `json:"fee_ref,omitempty"`
	LedgerHash       common.LedgerHash  `json:"ledger_hash,omitempty"`
	LedgerIndex      common.LedgerIndex `json:"ledger_index,omitempty"`
	LedgerTime       uint64             `json:"ledger_time,omitempty"`
	ReserveBase      uint               `json:"reserve_base,omitempty"`
	ReserveInc       uint               `json:"reserve_inc,omitempty"`
	ValidatedLedgers string             `json:"validated_ledgers,omitempty"`
	Offers           []ledger.Offer     `json:"offers,omitempty"`
}
