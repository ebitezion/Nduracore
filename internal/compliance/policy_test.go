package compliance

import (
	"testing"
	"time"

	"github.com/ebitezion/Nduracore/internal/data"
)

func TestEvaluateWithdrawalPolicyRejectsNonWhitelistedDestination(t *testing.T) {
	decision := EvaluateWithdrawalPolicy(WithdrawalPolicyInput{
		Policy: data.PolicyRule{
			WhitelistDestinations: []string{"0xabc"},
			PerTxLimitMinor:       1000,
			DailyLimitMinor:       10000,
			VelocityCount:         5,
			VelocityWindowMinutes: 60,
			ApprovalTier:          "one",
		},
		AmountMinor: 500,
		Destination: "0xdef",
		Now:         time.Now().UTC(),
	})

	if decision.Allowed {
		t.Fatalf("expected withdrawal to be rejected")
	}
	if decision.Reason != "destination_not_whitelisted" {
		t.Fatalf("unexpected reason: %s", decision.Reason)
	}
}

func TestEvaluateWithdrawalPolicyElevatesMediumRiskToTwoApprovals(t *testing.T) {
	now := time.Now().UTC()
	decision := EvaluateWithdrawalPolicy(WithdrawalPolicyInput{
		Policy: data.PolicyRule{
			WhitelistDestinations: []string{"0xabc"},
			PerTxLimitMinor:       5_000_000,
			DailyLimitMinor:       20_000_000,
			VelocityCount:         10,
			VelocityWindowMinutes: 60,
			ApprovalTier:          "one",
		},
		AmountMinor: 2_000_000,
		Destination: "0xabc",
		Risk: WithdrawalRiskAssessment{
			Level: "medium",
		},
		Now: now,
	})

	if !decision.Allowed {
		t.Fatalf("expected withdrawal to be allowed")
	}
	if decision.RequiredApprovals != 2 {
		t.Fatalf("expected required approvals to be 2, got %d", decision.RequiredApprovals)
	}
}
