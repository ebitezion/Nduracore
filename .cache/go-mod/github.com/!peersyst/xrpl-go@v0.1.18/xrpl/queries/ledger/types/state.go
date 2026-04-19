//revive:disable:var-naming
package types

import "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"

// State represents a raw ledger state object with its data and index.
type State struct {
	Data            string                  `json:"data,omitempty"`
	LedgerEntryType ledger.EntryType        `json:",omitempty"`
	LedgerObject    ledger.FlatLedgerObject `json:"-"`
	Index           string                  `json:"index"`
}
