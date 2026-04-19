package transaction

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDIDSet_TxType(t *testing.T) {
	tx := &DIDSet{}
	require.Equal(t, DIDSetTx, tx.TxType())
}

func TestDIDSet_Flatten(t *testing.T) {
	testcases := []struct {
		name     string
		tx       *DIDSet
		expected FlatTransaction
	}{
		{
			name: "pass - base transaction",
			tx: &DIDSet{
				BaseTx: BaseTx{
					Account: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				},
			},
			expected: FlatTransaction{
				"Account":         "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				"TransactionType": "DIDSet",
			},
		},
		{
			name: "pass - with data",
			tx: &DIDSet{
				BaseTx: BaseTx{
					Account: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				},
				Data: "data",
			},
			expected: FlatTransaction{
				"Account":         "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				"TransactionType": "DIDSet",
				"Data":            "data",
			},
		},
		{
			name: "pass - with did document",
			tx: &DIDSet{
				BaseTx: BaseTx{
					Account: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				},
				DIDDocument: "did document",
			},
			expected: FlatTransaction{
				"Account":         "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				"TransactionType": "DIDSet",
				"DIDDocument":     "did document",
			},
		},
		{
			name: "pass - with uri",
			tx: &DIDSet{
				BaseTx: BaseTx{
					Account: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				},
				URI: "uri",
			},
			expected: FlatTransaction{
				"Account":         "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
				"TransactionType": "DIDSet",
				"URI":             "uri",
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			flattened := tc.tx.Flatten()
			require.Equal(t, tc.expected, flattened)
		})
	}
}

func TestDIDSet_Validate(t *testing.T) {
	testcases := []struct {
		name        string
		tx          *DIDSet
		expected    bool
		expectedErr error
	}{
		{
			name:        "fail - missing account",
			tx:          &DIDSet{},
			expected:    false,
			expectedErr: ErrInvalidAccount,
		},
		{
			name: "fail - no field set",
			tx: &DIDSet{
				BaseTx: BaseTx{
					Account:         "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					TransactionType: DIDSetTx,
				},
			},
			expected:    false,
			expectedErr: ErrDIDSetMustSetEitherDataOrDIDDocumentOrURI,
		},
		{
			name: "pass - set data",
			tx: &DIDSet{
				BaseTx: BaseTx{
					Account:         "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					TransactionType: DIDSetTx,
				},
				Data: "data",
			},
			expected: true,
		},
		{
			name: "pass - set did document",
			tx: &DIDSet{
				BaseTx: BaseTx{
					Account:         "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					TransactionType: DIDSetTx,
				},
				DIDDocument: "did document",
			},
			expected: true,
		},
		{
			name: "pass - set uri",
			tx: &DIDSet{
				BaseTx: BaseTx{
					Account:         "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					TransactionType: DIDSetTx,
				},
				URI: "uri",
			},
			expected: true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			valid, err := tc.tx.Validate()
			if tc.expectedErr != nil {
				require.Equal(t, tc.expectedErr, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expected, valid)
			}
		})
	}
}
