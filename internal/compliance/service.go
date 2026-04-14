package compliance

import "context"

// Service defines compliance hooks for onboarding and transaction checks.
type Service interface {
	EvaluateTransaction(ctx context.Context, input TransactionScreeningInput) (ScreeningDecision, error)
}

type TransactionScreeningInput struct {
	TenantID    string
	ActorID     string
	AmountMinor int64
	Asset       string
	Country     string
}

type ScreeningDecision struct {
	Decision string
	Reasons  []string
}
