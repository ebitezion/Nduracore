//revive:disable:var-naming
package types

import "github.com/Peersyst/xrpl-go/xrpl/transaction/types"

// FeeLevels represents the network fee levels returned by the server.
type FeeLevels struct {
	MedianLevel     types.XRPCurrencyAmount `json:"median_level"`
	MinimumLevel    types.XRPCurrencyAmount `json:"minimum_level"`
	OpenLedgerLevel types.XRPCurrencyAmount `json:"open_ledger_level"`
	ReferenceLevel  types.XRPCurrencyAmount `json:"reference_level"`
}
