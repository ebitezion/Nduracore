package v1

import (
	"github.com/Peersyst/xrpl-go/xrpl/ledger-entry-types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	ledgertypes "github.com/Peersyst/xrpl-go/xrpl/queries/ledger/types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
)

// DataRequest is the request type for the ledger_data method.
// It retrieves the contents of the specified ledger and supports iterating
// through multiple calls to obtain the full contents of a single ledger version.
type DataRequest struct {
	common.BaseRequest
	LedgerHash  common.LedgerHash      `json:"ledger_hash,omitempty"`
	LedgerIndex common.LedgerSpecifier `json:"ledger_index,omitempty"`
	Binary      bool                   `json:"binary,omitempty"`
	Limit       int                    `json:"limit,omitempty"`
	Marker      any                    `json:"marker,omitempty"`
	Type        ledger.EntryType       `json:"type,omitempty"`
}

// Method returns the JSON-RPC method name for the DataRequest.
func (*DataRequest) Method() string {
	return "ledger_data"
}

// APIVersion returns the API version for the DataRequest.
func (*DataRequest) APIVersion() int {
	return version.RippledAPIV1
}

// Validate checks that the DataRequest parameters are valid.
// TODO: implement V2 validation.
func (*DataRequest) Validate() error {
	return nil
}

// DataResponse is the response type for the ledger_data method.
type DataResponse struct {
	LedgerIndex uint32              `json:"ledger_index"`
	LedgerHash  common.LedgerHash   `json:"ledger_hash"`
	State       []ledgertypes.State `json:"state"`
	Marker      any                 `json:"marker"`
}
