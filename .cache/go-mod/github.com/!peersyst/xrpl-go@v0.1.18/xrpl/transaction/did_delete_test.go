package transaction

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDIDDelete_TxType(t *testing.T) {
	tx := &DIDDelete{}
	require.Equal(t, DIDDeleteTx, tx.TxType())
}

func TestDIDDelete_Flatten(t *testing.T) {
	testcases := []struct {
		name     string
		tx       *DIDDelete
		expected FlatTransaction
	}{
		{
			name: "pass - base transaction",
			tx: &DIDDelete{
				BaseTx: BaseTx{
					Account: "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				},
			},
			expected: FlatTransaction{
				"Account":         "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
				"TransactionType": "DIDDelete",
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			require.Equal(t, testcase.expected, testcase.tx.Flatten())
		})
	}
}

func TestDIDDelete_Validate(t *testing.T) {
	testcases := []struct {
		name        string
		tx          *DIDDelete
		expected    bool
		expectedErr error
	}{
		{
			name: "fail - base transaction missing account",
			tx: &DIDDelete{
				BaseTx: BaseTx{},
			},
			expected:    false,
			expectedErr: ErrInvalidAccount,
		},
		{
			name: "pass - base transaction",
			tx: &DIDDelete{
				BaseTx: BaseTx{
					Account:         "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
					TransactionType: DIDDeleteTx,
				},
			},
			expected: true,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			valid, err := testcase.tx.Validate()
			if testcase.expectedErr != nil {
				require.ErrorIs(t, err, testcase.expectedErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, testcase.expected, valid)
			}
		})
	}
}
