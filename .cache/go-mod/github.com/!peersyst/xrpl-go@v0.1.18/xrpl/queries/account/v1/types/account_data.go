//revive:disable:var-naming
package types

import "github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"

// AccountData represents an account's root ledger fields along with any
// signer lists associated with that account.
type AccountData struct {
	ledger.AccountRoot
	SignerLists []ledger.SignerList `json:"signer_lists,omitempty"`
}
