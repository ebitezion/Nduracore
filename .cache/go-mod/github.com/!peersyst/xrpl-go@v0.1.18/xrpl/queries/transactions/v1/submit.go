// Package v1 contains version 1 transaction queries for XRPL.
package v1

import (
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
)

// ############################################################################
// Request
// ############################################################################

// SubmitRequest applies a transaction and sends it to the network for confirmation.
type SubmitRequest struct {
	common.BaseRequest
	TxBlob   string `json:"tx_blob"`
	FailHard bool   `json:"fail_hard,omitempty"`
}

// Method returns the JSON-RPC method name for SubmitRequest.
func (*SubmitRequest) Method() string {
	return "submit"
}

// APIVersion returns the API version supported by SubmitRequest.
func (*SubmitRequest) APIVersion() int {
	return version.RippledAPIV1
}

// Validate performs validation on SubmitRequest.
func (req *SubmitRequest) Validate() error {
	if req.TxBlob == "" {
		return ErrNoTxBlob
	}
	return nil
}

// ############################################################################
// Response
// ############################################################################

// SubmitResponse represents the response from the submit method, including engine results and transaction details.
type SubmitResponse struct {
	EngineResult             string                      `json:"engine_result"`
	EngineResultCode         int                         `json:"engine_result_code"`
	EngineResultMessage      string                      `json:"engine_result_message"`
	TxBlob                   string                      `json:"tx_blob"`
	Tx                       transaction.FlatTransaction `json:"tx_json"`
	Accepted                 bool                        `json:"accepted"`
	AccountSequenceAvailable uint                        `json:"account_sequence_available"`
	AccountSequenceNext      uint                        `json:"account_sequence_next"`
	Applied                  bool                        `json:"applied"`
	Broadcast                bool                        `json:"broadcast"`
	Kept                     bool                        `json:"kept"`
	Queued                   bool                        `json:"queued"`
	OpenLedgerCost           string                      `json:"open_ledger_cost"`
	ValidatedLedgerIndex     common.LedgerIndex          `json:"validated_ledger_index"`
}
