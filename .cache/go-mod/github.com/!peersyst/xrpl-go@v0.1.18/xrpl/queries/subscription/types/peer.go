// Package types contains data structures for subscription stream types.
//
//revive:disable:var-naming
package types

import "github.com/Peersyst/xrpl-go/xrpl/queries/common"

// PeerStatusEvents represents event types for peer status changes in the subscription stream.
type PeerStatusEvents string

const (
	// PeerStatusClosingLedger indicates the peer closed a ledger version with this LedgerIndex,
	// usually meaning it is about to start consensus.
	PeerStatusClosingLedger PeerStatusEvents = "CLOSING_LEDGER"
	// PeerStatusAcceptedLedger indicates the peer built this ledger version as the result of a consensus round.
	PeerStatusAcceptedLedger PeerStatusEvents = "ACCEPTED_LEDGER"
	// PeerStatusSwitchedLedger indicates the peer switched to a different ledger version,
	// concluding it was not following the rest of the network.
	PeerStatusSwitchedLedger PeerStatusEvents = "SWITCHED_LEDGER"
	// PeerStatusLostSync indicates the peer fell behind the rest of the network in tracking ledger validation and consensus.
	PeerStatusLostSync PeerStatusEvents = "LOST_SYNC"
)

// PeerStatusStream represents a message received from the peer status subscription stream,
// containing the event type, timestamp, and optional ledger details.
type PeerStatusStream struct {
	// `peerStatusChange` indicates this comes from the Peer Status stream.
	Type Type `json:"type"`
	// The type of event that prompted this message. See Peer Status Events for possible values.
	Action PeerStatusEvents `json:"action"`
	// The time this event occurred, in seconds since the Ripple Epoch.
	Date uint64 `json:"date"`
	// (May be omitted) The identifying Hash of a ledger version to which this message pertains.
	LedgerHash common.LedgerHash `json:"ledger_hash,omitempty"`
	// (May be omitted) The Ledger Index of a ledger version to which this message pertains.
	LedgerIndex common.LedgerIndex `json:"ledger_index,omitempty"`
	// (May be omitted) The largest Ledger Index the peer has currently available.
	LedgerIndexMax common.LedgerIndex `json:"ledger_index_max,omitempty"`
	// (May be omitted) The smallest Ledger Index the peer has currently available.
	LedgerIndexMin common.LedgerIndex `json:"ledger_index_min,omitempty"`
}
