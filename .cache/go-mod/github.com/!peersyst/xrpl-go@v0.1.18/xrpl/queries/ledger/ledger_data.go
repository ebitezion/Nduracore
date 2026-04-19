package ledger

import (
	"github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	ledgertypes "github.com/Peersyst/xrpl-go/xrpl/queries/ledger/types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
)

// ############################################################################
// Request
// ############################################################################

// DataRequest retrieves contents of the specified ledger. Multiple calls can be used to iterate through the ledger version.
type DataRequest struct {
	common.BaseRequest
	LedgerHash  common.LedgerHash      `json:"ledger_hash,omitempty"`
	LedgerIndex common.LedgerSpecifier `json:"ledger_index,omitempty"`
	Binary      bool                   `json:"binary,omitempty"`
	Limit       int                    `json:"limit,omitempty"`
	Marker      any                    `json:"marker,omitempty"`
	Type        ledger.EntryType       `json:"type,omitempty"`
}

// Method returns the JSON-RPC method name for DataRequest.
func (*DataRequest) Method() string {
	return "ledger_data"
}

// APIVersion returns the Rippled API version for DataRequest.
func (*DataRequest) APIVersion() int {
	return version.RippledAPIV2
}

// Validate checks the DataRequest fields for validity.
// TODO implement V2
func (*DataRequest) Validate() error {
	return nil
}

// ############################################################################
// Response
// ############################################################################

// DataResponse is the response returned by the ledger_data method, containing ledger state entries.
type DataResponse struct {
	LedgerIndex string              `json:"ledger_index"`
	LedgerHash  common.LedgerHash   `json:"ledger_hash"`
	State       []ledgertypes.State `json:"state"`
	Marker      any                 `json:"marker"`
}
