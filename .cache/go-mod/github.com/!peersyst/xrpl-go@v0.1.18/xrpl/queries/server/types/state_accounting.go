//revive:disable:var-naming
package types

// StateAccountingFinal holds accounting information for various server states.
type StateAccountingFinal struct {
	Disconnected InfoAccounting `json:"disconnected"`
	Connected    InfoAccounting `json:"connected"`
	Full         InfoAccounting `json:"full"`
	Syncing      InfoAccounting `json:"syncing"`
	Tracking     InfoAccounting `json:"tracking"`
}

// InfoAccounting represents duration and transition metrics for a server state.
type InfoAccounting struct {
	DurationUS  string `json:"duration_us"`
	Transitions string `json:"transitions"`
}
