// Package types provides data structures for server query responses.
//
//revive:disable:var-naming
package types

// ServerPort represents a network port and its supported protocols in the server response.
type ServerPort struct {
	Port     string   `json:"port"`
	Protocol []string `json:"protocol"`
}
