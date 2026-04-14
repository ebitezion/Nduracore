package compliance

import (
	"strings"
	"time"

	"github.com/ebitezion/Nduracore/internal/data"
)

type WithdrawalPolicyDecision struct {
	Allowed           bool
	RequiredApprovals int
	Reason            string
}

type WithdrawalRiskAssessment struct {
	Level  string
	Reason string
}

type WithdrawalPolicyInput struct {
	Policy            data.PolicyRule
	AmountMinor       int64
	Destination       string
	RecentWithdrawals []data.Withdrawal
	Risk              WithdrawalRiskAssessment
	Now               time.Time
}

func EvaluateWithdrawalPolicy(input WithdrawalPolicyInput) WithdrawalPolicyDecision {
	now := input.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}

	if !isDestinationAllowed(input.Policy.WhitelistDestinations, input.Destination) {
		return WithdrawalPolicyDecision{Allowed: false, Reason: "destination_not_whitelisted"}
	}

	if input.Policy.PerTxLimitMinor > 0 && input.AmountMinor > input.Policy.PerTxLimitMinor {
		return WithdrawalPolicyDecision{Allowed: false, Reason: "amount_exceeds_per_tx_limit"}
	}

	dailyTotal := int64(0)
	windowCount := 0
	velocityWindow := time.Duration(input.Policy.VelocityWindowMinutes) * time.Minute
	windowCutoff := now.Add(-velocityWindow)
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	for _, withdrawal := range input.RecentWithdrawals {
		if withdrawal.Status == "rejected" || withdrawal.Status == "failed" {
			continue
		}
		if withdrawal.CreatedAt.After(todayStart) {
			dailyTotal += withdrawal.AmountMinor
		}
		if input.Policy.VelocityWindowMinutes > 0 && withdrawal.CreatedAt.After(windowCutoff) {
			windowCount++
		}
	}

	if input.Policy.DailyLimitMinor > 0 && (dailyTotal+input.AmountMinor) > input.Policy.DailyLimitMinor {
		return WithdrawalPolicyDecision{Allowed: false, Reason: "amount_exceeds_daily_limit"}
	}

	if input.Policy.VelocityCount > 0 && windowCount >= input.Policy.VelocityCount {
		return WithdrawalPolicyDecision{Allowed: false, Reason: "velocity_limit_exceeded"}
	}

	requiredApprovals := approvalTierToCount(input.Policy.ApprovalTier)
	switch strings.ToLower(strings.TrimSpace(input.Risk.Level)) {
	case "high":
		return WithdrawalPolicyDecision{Allowed: false, Reason: "risk_rejected_high"}
	case "medium":
		if requiredApprovals < 2 {
			requiredApprovals = 2
		}
	}

	return WithdrawalPolicyDecision{Allowed: true, RequiredApprovals: requiredApprovals, Reason: "policy_passed"}
}

func approvalTierToCount(tier string) int {
	switch strings.ToLower(strings.TrimSpace(tier)) {
	case "auto":
		return 0
	case "two":
		return 2
	default:
		return 1
	}
}

func isDestinationAllowed(whitelist []string, destination string) bool {
	destination = strings.ToLower(strings.TrimSpace(destination))
	if destination == "" {
		return false
	}
	for _, allowed := range whitelist {
		if strings.ToLower(strings.TrimSpace(allowed)) == destination {
			return true
		}
	}
	return false
}
