// Package types contains data structures for subscription stream types.
// revive:disable:var-naming
package types

// ConsensusStream sends consensusPhase messages when the consensus process changes phase.
// It contains the new consensus phase the server is in.
type ConsensusStream struct {
	// The value `consensusPhase` indicates this is from the consensus stream
	Type Type `json:"type"`
	// The new consensus phase the server is in. Possible values are `open`, `establish`, and `accepted`.
	Consensus string `json:"consensus"`
}
