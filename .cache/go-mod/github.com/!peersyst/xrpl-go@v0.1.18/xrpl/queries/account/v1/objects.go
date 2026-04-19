package v1

import (
	"github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// ObjectType defines the type of ledger object for the account_objects request.
type ObjectType string

const (
	// CheckObject represents a ledger object of type 'check'.
	CheckObject ObjectType = "check"
	// DepositPreauthObject represents a ledger object of type 'deposit_preauth'.
	DepositPreauthObject ObjectType = "deposit_preauth"
	// EscrowObject represents a ledger object of type 'escrow'.
	EscrowObject ObjectType = "escrow"
	// NFTOfferObject represents a ledger object of type 'nft_offer'.
	NFTOfferObject ObjectType = "nft_offer"
	// OfferObject represents a ledger object of type 'offer'.
	OfferObject ObjectType = "offer"
	// PaymentChannelObject represents a ledger object of type 'payment_channel'.
	PaymentChannelObject ObjectType = "payment_channel"
	// SignerListObject represents a ledger object of type 'signer_list'.
	SignerListObject ObjectType = "signer_list"
	// StateObject represents a ledger object of type 'state'.
	StateObject ObjectType = "state"
	// TicketObject represents a ledger object of type 'ticket'.
	TicketObject ObjectType = "ticket"
)

// ############################################################################
// Request
// ############################################################################

// ObjectsRequest returns raw ledger objects owned by an account.
type ObjectsRequest struct {
	common.BaseRequest
	Account              types.Address          `json:"account"`
	Type                 ObjectType             `json:"type,omitempty"`
	DeletionBlockersOnly bool                   `json:"deletion_blockers_only,omitempty"`
	LedgerHash           common.LedgerHash      `json:"ledger_hash,omitempty"`
	LedgerIndex          common.LedgerSpecifier `json:"ledger_index,omitempty"`
	Limit                int                    `json:"limit,omitempty"`
	Marker               any                    `json:"marker,omitempty"`
}

// Method returns the JSON-RPC method name for ObjectsRequest.
func (*ObjectsRequest) Method() string {
	return "account_objects"
}

// APIVersion returns the Rippled API version for ObjectsRequest.
func (*ObjectsRequest) APIVersion() int {
	return version.RippledAPIV1
}

// Validate checks the ObjectsRequest parameters for validity.
// TODO implement v2
func (*ObjectsRequest) Validate() error {
	return nil
}

// ############################################################################
// Response
// ############################################################################

// ObjectsResponse is the response returned by the account_objects method.
type ObjectsResponse struct {
	Account            types.Address             `json:"account"`
	AccountObjects     []ledger.FlatLedgerObject `json:"account_objects"`
	LedgerHash         common.LedgerHash         `json:"ledger_hash,omitempty"`
	LedgerIndex        common.LedgerIndex        `json:"ledger_index,omitempty"`
	LedgerCurrentIndex common.LedgerIndex        `json:"ledger_current_index,omitempty"`
	Limit              int                       `json:"limit,omitempty"`
	Marker             any                       `json:"marker,omitempty"`
	Validated          bool                      `json:"validated,omitempty"`
}
