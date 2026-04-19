package transaction

import (
	addresscodec "github.com/Peersyst/xrpl-go/address-codec"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// XChainAccountCreateCommit creates a new account for a witness server to submit transactions on an issuing chain.
// This transaction can only be used for XRP-XRP bridges and requires the XChainBridge amendment.
//
// ```json
//
//	{
//	  "Account": "rwEqJ2UaQHe7jihxGqmx6J4xdbGiiyMaGa",
//	  "Destination": "rD323VyRjgzzhY4bFpo44rmyh2neB5d8Mo",
//	  "TransactionType": "XChainAccountCreateCommit",
//	  "Amount": "20000000",
//	  "SignatureReward": "100",
//	  "XChainBridge": {
//	    "LockingChainDoor": "rMAXACCrp3Y8PpswXcg3bKggHX76V3F8M4",
//	    "LockingChainIssue": {
//	      "currency": "XRP"
//	    },
//	    "IssuingChainDoor": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
//	    "IssuingChainIssue": {
//	      "currency": "XRP"
//	    }
//	  }
//	}
//
// ```
type XChainAccountCreateCommit struct {
	BaseTx

	// The amount, in XRP, to use for account creation. This must be greater than or equal to
	// the MinAccountCreateAmount specified in the Bridge ledger object.
	Amount types.CurrencyAmount
	// The destination account on the destination chain.
	Destination types.Address
	// The amount, in XRP, to be used to reward the witness servers for providing signatures.
	// This must match the amount on the Bridge ledger object.
	SignatureReward types.CurrencyAmount `json:",omitempty"`
	// The bridge to create accounts for.
	XChainBridge types.XChainBridge
}

// TxType returns the transaction type for XChainAccountCreateCommit.
func (x *XChainAccountCreateCommit) TxType() TxType {
	return XChainAccountCreateCommitTx
}

// Flatten returns a flattened representation of the XChainAccountCreateCommit transaction.
func (x *XChainAccountCreateCommit) Flatten() FlatTransaction {
	flatTx := x.BaseTx.Flatten()

	flatTx["TransactionType"] = x.TxType().String()

	if x.Amount != nil {
		flatTx["Amount"] = x.Amount.Flatten()
	}

	if x.Destination != "" {
		flatTx["Destination"] = x.Destination.String()
	}

	if x.SignatureReward != nil {
		flatTx["SignatureReward"] = x.SignatureReward.Flatten()
	}

	if x.XChainBridge != (types.XChainBridge{}) {
		flatTx["XChainBridge"] = x.XChainBridge.Flatten()
	}

	return flatTx
}

// Validate validates the XChainAccountCreateCommit transaction.
func (x *XChainAccountCreateCommit) Validate() (bool, error) {
	_, err := x.BaseTx.Validate()
	if err != nil {
		return false, err
	}

	if ok, err := IsAmount(x.Amount, "Amount", true); !ok {
		return false, err
	}

	if !addresscodec.IsValidAddress(x.Destination.String()) {
		return false, ErrInvalidAccount
	}

	if ok, err := IsAmount(x.SignatureReward, "SignatureReward", false); !ok {
		return false, err
	}

	return x.XChainBridge.Validate()
}
