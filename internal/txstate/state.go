package txstate

import "strings"

const (
	StateRequested     = "requested"
	StatePolicyPending = "policy_pending"
	StateApproved      = "approved"
	StateBroadcasted   = "broadcasted"
	StateConfirmed     = "confirmed"
	StateFailed        = "failed"
	StateRejected      = "rejected"
	StateCancelled     = "cancelled"
)

var allowedTransitions = map[string]map[string]struct{}{
	"":                 {StateRequested: {}, StatePolicyPending: {}, StateApproved: {}, StateRejected: {}},
	StateRequested:     {StatePolicyPending: {}, StateApproved: {}, StateRejected: {}, StateFailed: {}},
	StatePolicyPending: {StateApproved: {}, StateRejected: {}, StateFailed: {}},
	StateApproved:      {StateBroadcasted: {}, StateFailed: {}, StateCancelled: {}},
	StateBroadcasted:   {StateConfirmed: {}, StateFailed: {}},
	StateConfirmed:     {},
	StateFailed:        {StateApproved: {}, StateCancelled: {}},
	StateRejected:      {},
	StateCancelled:     {},
}

func Normalize(state string) string {
	return strings.ToLower(strings.TrimSpace(state))
}

func CanTransition(from, to string) bool {
	from = Normalize(from)
	to = Normalize(to)
	if to == "" {
		return false
	}
	options, ok := allowedTransitions[from]
	if !ok {
		return false
	}
	_, exists := options[to]
	return exists
}
