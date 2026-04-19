// Package clio provides types and requests for CLIO-specific XRPL queries.
package clio

import (
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
)

// ############################################################################
// Request
// ############################################################################

// NFTHistoryRequest retrieves a list of transactions that involved the specified NFToken.
type NFTHistoryRequest struct {
	common.BaseRequest
	NFTokenID      string `json:"nft_id"`
	LedgerIndexMin uint   `json:"ledger_index_min,omitempty"`
	LedgerIndexMax uint   `json:"ledger_index_max,omitempty"`
	Binary         bool   `json:"binary,omitempty"`
	Forward        bool   `json:"forward,omitempty"`
	Limit          uint   `json:"limit,omitempty"`
	Marker         any    `json:"marker,omitempty"`
}

// Method returns the JSON-RPC method name for NFTHistoryRequest.
func (*NFTHistoryRequest) Method() string {
	return "nft_history"
}

// APIVersion returns the Rippled API version for NFTHistoryRequest.
func (*NFTHistoryRequest) APIVersion() int {
	return version.RippledAPIV2
}

// Validate checks the NFTHistoryRequest parameters for validity.
// TODO implement V2
func (*NFTHistoryRequest) Validate() error {
	return nil
}

// ############################################################################
// Response
// ############################################################################

// NFTHistoryTransactions represents a transaction record returned in NFTHistoryResponse.
type NFTHistoryTransactions struct {
	LedgerIndex uint                        `json:"ledger_index"`
	Meta        transaction.TxObjMeta       `json:"meta"`
	Tx          transaction.FlatTransaction `json:"tx,omitempty"`
	TxBlob      string                      `json:"tx_blob,omitempty"`
	Validated   bool                        `json:"validated"`
}

// NFTHistoryResponse is the response returned by the nft_history method, containing matching transactions.
type NFTHistoryResponse struct {
	NFTokenID      string                   `json:"nft_id"`
	LedgerIndexMin uint                     `json:"ledger_index_min,omitempty"`
	LedgerIndexMax uint                     `json:"ledger_index_max,omitempty"`
	Limit          uint                     `json:"limit,omitempty"`
	Marker         any                      `json:"marker,omitempty"`
	Transactions   []NFTHistoryTransactions `json:"transactions"`
	Validated      bool                     `json:"validated,omitempty"`
}
