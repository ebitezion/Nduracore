// Package common provides shared types and utilities for XRPL query requests.
//
//revive:disable:var-naming
package common

// BaseRequest holds the API version for JSON-RPC requests.
type BaseRequest struct {
	Version int `json:"api_version,omitempty"`
}

// APIVersion returns the API version set on the BaseRequest.
func (r *BaseRequest) APIVersion() int {
	return r.Version
}

// SetAPIVersion sets the API version on the BaseRequest.
func (r *BaseRequest) SetAPIVersion(apiVersion int) {
	r.Version = apiVersion
}
