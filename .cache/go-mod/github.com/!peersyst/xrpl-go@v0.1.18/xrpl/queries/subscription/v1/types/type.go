// Package types contains data structures for v1 subscription stream types.
// revive:disable:var-naming
package types

// Type represents a v1 subscription stream type.
type Type string

// Stream types used in v1 subscription requests.
const (
	LedgerStreamType      Type = "ledgerClosed"
	ValidationStreamType  Type = "validationReceived"
	TransactionStreamType Type = "transaction"
	PeerStatusStreamType  Type = "peerStatusChange"
	OrderBookStreamType   Type = TransactionStreamType
	ConsensusStreamType   Type = "consensusPhase"
)
