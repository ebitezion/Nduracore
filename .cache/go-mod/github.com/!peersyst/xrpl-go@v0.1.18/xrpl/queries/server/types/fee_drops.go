//revive:disable:var-naming
package types

import "github.com/Peersyst/xrpl-go/xrpl/transaction/types"

// FeeDrops represents the fee amounts in drops returned by the server.
type FeeDrops struct {
	BaseFee       types.XRPCurrencyAmount `json:"base_fee"`
	MedianFee     types.XRPCurrencyAmount `json:"median_fee"`
	MinimumFee    types.XRPCurrencyAmount `json:"minimum_fee"`
	OpenLedgerFee types.XRPCurrencyAmount `json:"open_ledger_fee"`
}
