package integrations

import "context"

// Service abstracts external rails and providers (fiat, crypto infra, and risk providers).
type Service interface {
	QuoteFX(ctx context.Context, base, quote string, amountMinor int64) (FXQuote, error)
}

type FXQuote struct {
	Provider     string
	Base         string
	Quote        string
	Rate         float64
	ExpiresAtUTC string
}
