// Package server contains server-related queries for XRPL.
package server

import (
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	"github.com/Peersyst/xrpl-go/xrpl/queries/server/types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
)

// ############################################################################
// Feature All Request
// ############################################################################

// FeatureAllRequest is the request type for the feature command.
// It returns information about amendments this server knows about,
// including whether they are enabled.
type FeatureAllRequest struct {
	common.BaseRequest
}

// Method returns the JSON-RPC method name for the FeatureAllRequest.
func (*FeatureAllRequest) Method() string {
	return "feature"
}

// APIVersion returns the API version required by the FeatureAllRequest.
func (*FeatureAllRequest) APIVersion() int {
	return version.RippledAPIV2
}

// Validate verifies the FeatureAllRequest parameters.
// TODO: implement V2 validation logic.
func (*FeatureAllRequest) Validate() error {
	return nil
}

// ############################################################################
// Feature All Response
// ############################################################################

// FeatureAllResponse is the response type returned by the feature command.
// It contains a map of amendment names to their FeatureStatus.
type FeatureAllResponse struct {
	Features map[string]types.FeatureStatus `json:"features"`
}

// ############################################################################
// Feature One Request
// ############################################################################

// FeatureOneRequest is the request type for the feature command for a single amendment.
// It specifies which amendment to query.
type FeatureOneRequest struct {
	common.BaseRequest
	Feature string `json:"feature"`
}

// Method returns the JSON-RPC method name for the FeatureOneRequest.
func (*FeatureOneRequest) Method() string {
	return "feature"
}

// APIVersion returns the API version required by the FeatureOneRequest.
func (*FeatureOneRequest) APIVersion() int {
	return version.RippledAPIV2
}

// Validate verifies the FeatureOneRequest parameters.
// TODO: implement V2 validation logic.
func (*FeatureOneRequest) Validate() error {
	return nil
}

// ############################################################################
// Feature One Response
// ############################################################################

// FeatureResponse is the response type returned by the feature command for a single amendment.
// It maps the amendment name to its FeatureStatus.
type FeatureResponse map[string]types.FeatureStatus
