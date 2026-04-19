// Package v1 contains version 1 CLIO server queries for XRPL.
package v1

import (
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
)

// ############################################################################
// Request
// ############################################################################

// NFTHistoryRequest retrieves a list of transactions that involved the specified NFToken via the nft_history method.
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

// Method returns the RPC method name for the NFTHistoryRequest.
func (*NFTHistoryRequest) Method() string {
	return "nft_history"
}

// APIVersion returns the API version for the NFTHistoryRequest.
func (*NFTHistoryRequest) APIVersion() int {
	return version.RippledAPIV1
}

// Validate checks the NFTHistoryRequest for correctness.
// TODO: Implement V2
func (*NFTHistoryRequest) Validate() error {
	return nil
}

// ############################################################################
// Response
// ############################################################################

// NFTHistoryTransactions represents a single transaction entry in the NFT history response.
type NFTHistoryTransactions struct {
	LedgerIndex uint                        `json:"ledger_index"`
	Meta        transaction.TxObjMeta       `json:"meta"`
	Tx          transaction.FlatTransaction `json:"tx,omitempty"`
	TxBlob      string                      `json:"tx_blob,omitempty"`
	Validated   bool                        `json:"validated"`
}

// NFTHistoryResponse represents the response returned by the nft_history method.
type NFTHistoryResponse struct {
	NFTokenID      string                   `json:"nft_id"`
	LedgerIndexMin uint                     `json:"ledger_index_min,omitempty"`
	LedgerIndexMax uint                     `json:"ledger_index_max,omitempty"`
	Limit          uint                     `json:"limit,omitempty"`
	Marker         any                      `json:"marker,omitempty"`
	Transactions   []NFTHistoryTransactions `json:"transactions"`
	Validated      bool                     `json:"validated,omitempty"`
}
