package types

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
)

func TestBatchSigner_Flatten(t *testing.T) {
	tests := []struct {
		name     string
		input    BatchSigner
		expected string
	}{
		{
			name: "pass - basic batch signer",
			input: BatchSigner{
				BatchSigner: BatchSignerData{
					Account:       "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					SigningPubKey: "ED5F5AC8B98974A3CA843326D9B88CEBD0560177B973EE0B149F782CFAA06DC66A",
					TxnSignature:  "C4E2834B9C0E7519DC47E4C48F19B4B2C5C92FB4F8C5C8F8C8C8C8C8C8C8",
				},
			},
			expected: `{
				"BatchSigner": {
					"Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					"SigningPubKey": "ED5F5AC8B98974A3CA843326D9B88CEBD0560177B973EE0B149F782CFAA06DC66A",
					"TxnSignature": "C4E2834B9C0E7519DC47E4C48F19B4B2C5C92FB4F8C5C8F8C8C8C8C8C8C8"
				}
			}`,
		},
		{
			name: "pass - batch signer with inner signers",
			input: BatchSigner{
				BatchSigner: BatchSignerData{
					Account: "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					Signers: []Signer{
						{
							SignerData: SignerData{
								Account:       "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
								SigningPubKey: "ED5F5AC8B98974A3CA843326D9B88CEBD0560177B973EE0B149F782CFAA06DC66A",
								TxnSignature:  "C4E2834B9C0E7519DC47E4C48F19B4B2C5C92FB4F8C5C8F8C8C8C8C8C8C8",
							},
						},
					},
				},
			},
			expected: `{
				"BatchSigner": {
					"Account": "rHb9CJAWyB4rj91VRWn96DkukG4bwdtyTh",
					"Signers": [
						{
							"Signer": {
								"Account": "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH",
								"SigningPubKey": "ED5F5AC8B98974A3CA843326D9B88CEBD0560177B973EE0B149F782CFAA06DC66A",
								"TxnSignature": "C4E2834B9C0E7519DC47E4C48F19B4B2C5C92FB4F8C5C8F8C8C8C8C8C8C8"
							}
						}
					]
				}
			}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.Flatten()
			err := testutil.CompareFlattenAndExpected(result, []byte(tt.expected))
			if err != nil {
				t.Error(err)
			}
		})
	}
}
