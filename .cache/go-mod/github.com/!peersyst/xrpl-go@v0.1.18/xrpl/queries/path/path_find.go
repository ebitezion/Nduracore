package path

import (
	"github.com/Peersyst/xrpl-go/xrpl/queries/common"
	pathtypes "github.com/Peersyst/xrpl-go/xrpl/queries/path/types"
	"github.com/Peersyst/xrpl-go/xrpl/queries/version"
	"github.com/Peersyst/xrpl-go/xrpl/transaction"
	"github.com/Peersyst/xrpl-go/xrpl/transaction/types"
)

// SubCommand represents the type of path_find subcommand (create, close, status).
type SubCommand string

const (
	// Create starts sending pathfinding information.
	Create SubCommand = "create"
	// Close stops sending pathfinding information.
	Close SubCommand = "close"
	// Status retrieves the status of a pathfinding request.
	Status SubCommand = "status"
)

// ############################################################################
// Create Request
// ############################################################################

// FindCreateRequest starts sending pathfinding information for a path_find request.
type FindCreateRequest struct {
	common.BaseRequest
	Subcommand         SubCommand             `json:"subcommand"`
	SourceAccount      types.Address          `json:"source_account,omitempty"`
	DestinationAccount types.Address          `json:"destination_account,omitempty"`
	DestinationAmount  types.CurrencyAmount   `json:"destination_amount,omitempty"`
	SendMax            types.CurrencyAmount   `json:"send_max,omitempty"`
	Paths              []transaction.PathStep `json:"paths,omitempty"`
	Domain             *string                `json:"domain,omitempty"`
}

// Method returns the JSON-RPC method name for FindCreateRequest.
func (*FindCreateRequest) Method() string {
	return "path_find"
}

// APIVersion returns the supported API version for FindCreateRequest.
func (*FindCreateRequest) APIVersion() int {
	return version.RippledAPIV2
}

// Validate checks that FindCreateRequest is correctly formed.
// TODO implement V2
func (*FindCreateRequest) Validate() error {
	return nil
}

// ############################################################################
// Close Request
// ############################################################################

// FindCloseRequest stops sending pathfinding information for a path_find request.
type FindCloseRequest struct {
	common.BaseRequest
	Subcommand SubCommand `json:"subcommand"`
}

// Method returns the JSON-RPC method name for FindCloseRequest.
func (*FindCloseRequest) Method() string {
	return "path_find"
}

// APIVersion returns the supported API version for FindCloseRequest.
func (*FindCloseRequest) APIVersion() int {
	return version.RippledAPIV2
}

// Validate checks that FindCloseRequest is correctly formed.
// TODO implement V2
func (*FindCloseRequest) Validate() error {
	return nil
}

// ############################################################################
// Status Request
// ############################################################################

// FindStatusRequest retrieves information of the currently-open pathfinding request.
type FindStatusRequest struct {
	common.BaseRequest
	Subcommand SubCommand `json:"subcommand"`
}

// Method returns the JSON-RPC method name for FindStatusRequest.
func (*FindStatusRequest) Method() string {
	return "path_find"
}

// APIVersion returns the supported API version for FindStatusRequest.
func (*FindStatusRequest) APIVersion() int {
	return version.RippledAPIV2
}

// Validate checks that FindStatusRequest is correctly formed.
// TODO implement V2
func (*FindStatusRequest) Validate() error {
	return nil
}

// ############################################################################
// Response
// ############################################################################

// TODO: Add ID handling (v2)

// FindResponse is returned by a path_find request, including alternatives, destination, and status.
type FindResponse struct {
	Alternatives       []pathtypes.Alternative `json:"alternatives"`
	DestinationAccount types.Address           `json:"destination_account"`
	// DestinationAmount  types.CurrencyAmount    `json:"destination_amount"`
	DestinationAmount any           `json:"destination_amount"`
	SourceAccount     types.Address `json:"source_account"`
	FullReply         bool          `json:"full_reply"`
	Closed            bool          `json:"closed,omitempty"`
	Status            bool          `json:"status,omitempty"`
	Domain            string        `json:"domain,omitempty"`
}
