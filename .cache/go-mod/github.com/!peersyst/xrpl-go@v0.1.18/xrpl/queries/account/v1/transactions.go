package v1

import (
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// Transaction represents a transaction with metadata and blob returned by account_tx.
type Transaction struct {
	LedgerIndex uint64                      `json:"ledger_index"`
	Meta        transaction.TxObjMeta       `json:"meta"`
	Tx          transaction.FlatTransaction `json:"tx"`
	TxBlob      string                      `json:"tx_blob"`
	Validated   bool                        `json:"validated"`
}

// ############################################################################
// Request
// ############################################################################

// TransactionsRequest retrieves a list of transactions involving the specified account.
type TransactionsRequest struct {
	common.BaseRequest
	Account        types.Address          `json:"account"`
	LedgerIndexMin int                    `json:"ledger_index_min,omitempty"`
	LedgerIndexMax int                    `json:"ledger_index_max,omitempty"`
	LedgerHash     common.LedgerHash      `json:"ledger_hash,omitempty"`
	LedgerIndex    common.LedgerSpecifier `json:"ledger_index,omitempty"`
	Binary         bool                   `json:"binary,omitempty"`
	Forward        bool                   `json:"forward,omitempty"`
	Limit          int                    `json:"limit,omitempty"`
	Marker         any                    `json:"marker,omitempty"`
}

// Method returns the JSON-RPC method name for TransactionsRequest.
func (*TransactionsRequest) Method() string {
	return "account_tx"
}

// APIVersion returns the Rippled API version for TransactionsRequest.
func (*TransactionsRequest) APIVersion() int {
	return version.RippledAPIV1
}

// Validate checks the TransactionsRequest for valid parameters.
// TODO implement v2
func (*TransactionsRequest) Validate() error {
	return nil
}

// ############################################################################
// Response
// ############################################################################

// TransactionsResponse represents the expected response from the account_tx method.
type TransactionsResponse struct {
	Account        types.Address      `json:"account"`
	LedgerIndexMin common.LedgerIndex `json:"ledger_index_min"`
	LedgerIndexMax common.LedgerIndex `json:"ledger_index_max"`
	Limit          int                `json:"limit"`
	Marker         any                `json:"marker,omitempty"`
	Transactions   []Transaction      `json:"transactions"`
	Validated      bool               `json:"validated"`
}
