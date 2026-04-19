// Package types contains data structures for oracle query types.
// revive:disable:var-naming
package types

// Set represents statistical metrics returned by the oracle query, including mean, size, and standard deviation.
type Set struct {
	Mean              string `json:"mean"`
	Size              uint32 `json:"size"`
	StandardDeviation string `json:"standard_deviation"`
}
