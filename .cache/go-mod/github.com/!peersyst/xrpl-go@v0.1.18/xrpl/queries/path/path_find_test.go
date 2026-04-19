package path

import (
	"testing"

	"github.com/Peersyst/xrpl-go/xrpl/testutil"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

func TestPathFindCloseRequest(t *testing.T) {
	s := FindCloseRequest{
		Subcommand: Close,
	}

	j := `{
	"subcommand": "close"
}`

	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestPathFindStatusRequest(t *testing.T) {
	s := FindStatusRequest{
		Subcommand: Status,
	}

	j := `{
	"subcommand": "status"
}`

	if err := testutil.SerializeAndDeserialize(t, s, j); err != nil {
		t.Error(err)
	}
}

func TestFindCreateRequestWithDomain(t *testing.T) {
	domain := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	s := FindCreateRequest{
		Subcommand:         Create,
		SourceAccount:      "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
		DestinationAccount: "rDCNEYQDfUYEgHjfLZ6CVHXNUCg6SdQgFN",
		DestinationAmount:  types.XRPCurrencyAmount(1000000),
		Domain:             &domain,
	}

	j := `{
	"subcommand": "create",
	"source_account": "r9cZA1mLK5R5Am25ArfXFmqgNwjZgnfk59",
	"destination_account": "rDCNEYQDfUYEgHjfLZ6CVHXNUCg6SdQgFN",
	"destination_amount": "1000000",
	"domain": "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
}`

	if err := testutil.Serialize(t, s, j); err != nil {
		t.Error(err)
	}
}
