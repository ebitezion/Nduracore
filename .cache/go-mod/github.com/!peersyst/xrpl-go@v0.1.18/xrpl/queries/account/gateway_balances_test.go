package account

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
)

func TestGatewayBalancesRequest(t *testing.T) {
	s := GatewayBalancesRequest{
		Account:     "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
		Strict:      true,
		HotWallet:   []string{"rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu"},
		LedgerHash:  "ABC123",
		LedgerIndex: common.LedgerTitle("validated"),
	}

	j := `{
	"account": "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
	"strict": true,
	"hotwallet": [
		"rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu"
	],
	"ledger_hash": "ABC123",
	"ledger_index": "validated"
}`
	if err := testutil.Serialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestGatewayBalancesResponse(t *testing.T) {
	s := GatewayBalancesResponse{
		Account: "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
		Obligations: map[string]string{
			"USD": "100",
			"EUR": "200",
		},
		Balances: map[string][]GatewayBalance{
			"rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu": {
				{
					Currency: "USD",
					Value:    "50",
				},
				{
					Currency: "EUR",
					Value:    "100",
				},
			},
		},
		Assets: map[string][]GatewayBalance{
			"rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu": {
				{
					Currency: "USD",
					Value:    "25",
				},
			},
		},
		LedgerHash:         "ABC123",
		LedgerCurrentIndex: 54321,
		LedgerIndex:        12345,
	}

	j := `{
	"account": "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
	"obligations": {
		"EUR": "200",
		"USD": "100"
	},
	"balances": {
		"rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu": [
			{
				"currency": "USD",
				"value": "50"
			},
			{
				"currency": "EUR",
				"value": "100"
			}
		]
	},
	"assets": {
		"rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu": [
			{
				"currency": "USD",
				"value": "25"
			}
		]
	},
	"ledger_hash": "ABC123",
	"ledger_current_index": 54321,
	"ledger_index": 12345
}`
	if err := testutil.Serialize(t, s, j); err != nil {
		t.Error(err)
	}
}
