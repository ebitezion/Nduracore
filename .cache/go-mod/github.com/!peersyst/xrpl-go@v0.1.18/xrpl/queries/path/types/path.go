// Package types contains data structures for path finding query types.
//
//revive:disable:var-naming
package types

import (
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
)

// Alternative represents one of the possible payment paths, including computed path steps and source/destination amounts.
type Alternative struct {
	PathsComputed [][]transaction.PathStep `json:"paths_computed"`
	// SourceAmount      types.CurrencyAmount     `json:"source_amount"`
	SourceAmount any `json:"source_amount"`
	// DestinationAmount types.CurrencyAmount     `json:"destination_amount,omitempty"`
	DestinationAmount any `json:"destination_amount,omitempty"`
}

// RippleAlternative represents an alternative payment path specifically for XRP, including computed path steps and source amount.
type RippleAlternative struct {
	PathsComputed [][]transaction.PathStep `json:"paths_computed"`
	// SourceAmount  types.CurrencyAmount     `json:"source_amount"`
	SourceAmount any `json:"source_amount"`
}
