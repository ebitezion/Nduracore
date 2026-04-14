package ledger

import "context"

// Service defines double-entry ledger responsibilities for all value movement.
type Service interface {
	Post(ctx context.Context, entry Entry) (PostingResult, error)
	Balance(ctx context.Context, accountID string) (Balance, error)
}

type Entry struct {
	Reference string
	Debit     Leg
	Credit    Leg
	Metadata  map[string]string
}

type Leg struct {
	AccountID string
	Asset     string
	Amount    int64
}

type PostingResult struct {
	EntryID string
	Status  string
}

type Balance struct {
	AccountID string
	Asset     string
	Available int64
	Pending   int64
}
