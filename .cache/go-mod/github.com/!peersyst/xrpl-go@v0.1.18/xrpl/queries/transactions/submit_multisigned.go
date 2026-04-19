package transactions

import (
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
)

// ############################################################################
// Request
// ############################################################################

// SubmitMultisignedRequest is the request type for the submit_multisigned command.
// It applies a multi-signed transaction and sends it to the network for inclusion in future ledgers.
type SubmitMultisignedRequest struct {
	common.BaseRequest
	Tx       transaction.FlatTransaction `json:"tx_json"`
	FailHard bool                        `json:"fail_hard"`
}

// Method returns the JSON-RPC method name for the SubmitMultisignedRequest.
func (*SubmitMultisignedRequest) Method() string {
	return "submit_multisigned"
}

// APIVersion returns the API version required by the SubmitMultisignedRequest.
func (*SubmitMultisignedRequest) APIVersion() int {
	return version.RippledAPIV2
}

// Validate verifies the SubmitMultisignedRequest parameters.
func (*SubmitMultisignedRequest) Validate() error {
	return nil
}

// ############################################################################
// Response
// ############################################################################

// SubmitMultisignedResponse is the response type returned by the submit_multisigned command.
type SubmitMultisignedResponse struct {
	EngineResult        string                      `json:"engine_result"`
	EngineResultCode    int                         `json:"engine_result_code"`
	EngineResultMessage string                      `json:"engine_result_message"`
	TxBlob              string                      `json:"tx_blob"`
	Tx                  transaction.FlatTransaction `json:"tx_json"`
}
