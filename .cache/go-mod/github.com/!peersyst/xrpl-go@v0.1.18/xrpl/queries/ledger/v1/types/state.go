//revive:disable:var-naming
package types

import "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"

// State represents a state object returned by the ledger v1 query.
type State struct {
	Data            string                  `json:"data,omitempty"`
	LedgerEntryType ledger.EntryType        `json:",omitempty"`
	LedgerObject    ledger.FlatLedgerObject `json:"-"`
	Index           string                  `json:"index"`
}
