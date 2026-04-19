// Package types contains data structures for v1 subscription stream types.
//
//revive:disable:var-naming
package types

// ConsensusStream represents a consensus notification from a v1 subscription stream.
// The message contains the new phase of consensus the server is in.
type ConsensusStream struct {
	// The value `consensusPhase` indicates this is from the consensus stream
	Type Type `json:"type"`
	// The new consensus phase the server is in. Possible values are `open`, `establish`, and `accepted`.
	Consensus string `json:"consensus"`
}
