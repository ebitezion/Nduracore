package faucet

import (
	"fmt"
)

// ErrMarshalPayload is returned when the faucet payload cannot be marshaled to JSON.
type ErrMarshalPayload struct {
	Err error
}

// Error implements the error interface for ErrMarshalPayload
func (e ErrMarshalPayload) Error() string {
	return fmt.Sprintf("failed to marshal faucet payload: %v", e.Err)
}

// ErrCreateRequest is returned when the faucet request cannot be created.
type ErrCreateRequest struct {
	Err error
}

// Error implements the error interface for ErrCreateRequest
func (e ErrCreateRequest) Error() string {
	return fmt.Sprintf("failed to create faucet request: %v", e.Err)
}

// ErrSendRequest is returned when the faucet request cannot be sent.
type ErrSendRequest struct {
	Err error
}

// Error implements the error interface for ErrSendRequest
func (e ErrSendRequest) Error() string {
	return fmt.Sprintf("failed to send POST request to faucet: %v", e.Err)
}

// ErrUnexpectedStatusCode is returned when the faucet responds with a non-200 status code.
type ErrUnexpectedStatusCode struct {
	Code int
}

// Error implements the error interface for ErrUnexpectedStatusCode
func (e ErrUnexpectedStatusCode) Error() string {
	return fmt.Sprintf("unexpected faucet response status code: %d", e.Code)
}
