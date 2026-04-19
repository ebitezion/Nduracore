package oracle

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/queries/oracle/types"
	"github.com/Peersyst/xrpl-go/xrpl/testutil"
)

func TestGetAggregatePriceRequest(t *testing.T) {
	s := GetAggregatePriceRequest{
		BaseAsset:  "XRP",
		QuoteAsset: "USD",
		Oracles: []types.Oracle{
			{
				Account:          "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
				OracleDocumentID: "123",
			},
		},
		Trim:          10,
		TrimThreshold: 3600,
	}

	j := `{
	"base_asset": "XRP",
	"quote_asset": "USD",
	"oracles": [
		{
			"account": "rLHmBn4fT92w4F6ViyYbjoizLTo83tHTHu",
			"oracle_document_id": "123"
		}
	],
	"trim": 10,
	"trim_threshold": 3600
}`
	if err := testutil.Serialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestGetAggregatePriceResponse(t *testing.T) {
	s := GetAggregatePriceResponse{
		EntireSet: types.Set{
			Mean:              "0.5123",
			Size:              10,
			StandardDeviation: "0.0123",
		},
		TrimmedSet: types.Set{
			Mean:              "0.5100",
			Size:              8,
			StandardDeviation: "0.0100",
		},
		Median:             "0.5110",
		Time:               1609459200,
		LedgerCurrentIndex: 54321,
		Validated:          true,
	}

	j := `{
	"entire_set": {
		"mean": "0.5123",
		"size": 10,
		"standard_deviation": "0.0123"
	},
	"trimmed_set": {
		"mean": "0.5100",
		"size": 8,
		"standard_deviation": "0.0100"
	},
	"median": "0.5110",
	"time": 1609459200,
	"ledger_current_index": 54321,
	"validated": true
}`
	if err := testutil.Serialize(t, s, j); err != nil {
		t.Error(err)
	}
}
