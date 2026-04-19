package hash

import "errors"

var (
	// transaction signature

	// ErrNonSignedTransaction indicates that a transaction lacks the required signature fields.
	ErrNonSignedTransaction = errors.New("transaction must have at least one of TxnSignature, Signers, or SigningPubKey")
	// ErrMissingSignature is returned when a transaction lacks the required signature fields.
	// A transaction must have at least one of: TxnSignature, Signers, or SigningPubKey,
	// unless it's an inner batch transaction (has TfInnerBatchTxn flag set).
	ErrMissingSignature = errors.New("transaction must have at least one of TxnSignature, Signers, or SigningPubKey")
)
