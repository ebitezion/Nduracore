package v1

import (
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	pathtypes "github.com/Peersyst/xrpl-go/xrpl/queries/path/types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// SubCommand represents the path_find subcommand type for session management.
type SubCommand string

// SubCommand values for opening, closing, and checking status of a path_find session.
const (
	// Create starts sending pathfinding information.
	Create SubCommand = "create"
	// Close stops sending pathfinding information.
	Close SubCommand = "close"
	// Status retrieves information on the currently-open pathfinding request.
	Status SubCommand = "status"
)

// ############################################################################
// Create Request
// ############################################################################

// FindCreateRequest is the request type for the path_find create subcommand.
// It starts sending pathfinding information.
type FindCreateRequest struct {
	common.BaseRequest
	Subcommand         SubCommand             `json:"subcommand"`
	SourceAccount      types.Address          `json:"source_account,omitempty"`
	DestinationAccount types.Address          `json:"destination_account,omitempty"`
	DestinationAmount  types.CurrencyAmount   `json:"destination_amount,omitempty"`
	SendMax            types.CurrencyAmount   `json:"send_max,omitempty"`
	Paths              []transaction.PathStep `json:"paths,omitempty"`
}

// Method returns the JSON-RPC method name for the FindCreateRequest.
func (*FindCreateRequest) Method() string {
	return "path_find"
}

// APIVersion returns the API version required by the FindCreateRequest.
func (*FindCreateRequest) APIVersion() int {
	return version.RippledAPIV1
}

// Validate verifies the FindCreateRequest parameters.
// TODO: implement V2.
func (*FindCreateRequest) Validate() error {
	return nil
}

// ############################################################################
// Close Request
// ############################################################################

// FindCloseRequest is the request type for the path_find close subcommand.
// It stops sending pathfinding information.
type FindCloseRequest struct {
	common.BaseRequest
	Subcommand SubCommand `json:"subcommand"`
}

// Method returns the JSON-RPC method name for the FindCloseRequest.
func (*FindCloseRequest) Method() string {
	return "path_find"
}

// Validate verifies the FindCloseRequest parameters.
// TODO: implement V2.
func (*FindCloseRequest) Validate() error {
	return nil
}

// ############################################################################
// Status Request
// ############################################################################

// FindStatusRequest is the request type for the path_find status subcommand.
// It retrieves information on the currently-open pathfinding request.
type FindStatusRequest struct {
	common.BaseRequest
	Subcommand SubCommand `json:"subcommand"`
}

// Method returns the JSON-RPC method name for the FindStatusRequest.
func (*FindStatusRequest) Method() string {
	return "path_find"
}

// Validate verifies the FindStatusRequest parameters.
// TODO: implement V2.
func (*FindStatusRequest) Validate() error {
	return nil
}

// ############################################################################
// Response
// ############################################################################

// TODO: Add ID handling (v2)

// FindResponse is the response type returned by the path_find command.
// It includes payment path alternatives and session status.
type FindResponse struct {
	Alternatives       []pathtypes.Alternative `json:"alternatives"`
	DestinationAccount types.Address           `json:"destination_account"`
	DestinationAmount  any                     `json:"destination_amount"`
	ID                 any                     `json:"id,omitempty"`
	SourceAccount      types.Address           `json:"source_account"`
	FullReply          bool                    `json:"full_reply"`
	Closed             bool                    `json:"closed,omitempty"`
	Status             bool                    `json:"status,omitempty"`
}
