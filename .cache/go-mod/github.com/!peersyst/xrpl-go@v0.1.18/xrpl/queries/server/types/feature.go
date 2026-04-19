//revive:disable:var-naming
package types

// FeatureStatus represents the status of a server feature including its enabled state, name, and support status.
type FeatureStatus struct {
	Enabled   bool   `json:"enabled"`
	Name      string `json:"name"`
	Supported bool   `json:"supported"`
}
