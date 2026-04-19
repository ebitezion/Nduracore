package txstate

import "strings"

var providerStatusMap = map[string]map[string]string{
	"fireblocks": {
		"pending_authorization": StatePolicyPending,
		"submitted":             StateApproved,
		"broadcasting":          StateBroadcasted,
		"completed":             StateConfirmed,
		"confirmed":             StateConfirmed,
		"failed":                StateFailed,
		"rejected":              StateRejected,
		"cancelled":             StateCancelled,
	},
	"alchemy": {
		"pending":     StateBroadcasted,
		"broadcasted": StateBroadcasted,
		"mined":       StateConfirmed,
		"confirmed":   StateConfirmed,
		"failed":      StateFailed,
	},
}

func CanonicalFromProvider(provider, status string) string {
	provider = strings.ToLower(strings.TrimSpace(provider))
	status = strings.ToLower(strings.TrimSpace(status))
	if provider == "" || status == "" {
		return ""
	}
	lookup, ok := providerStatusMap[provider]
	if !ok {
		return ""
	}
	return lookup[status]
}
