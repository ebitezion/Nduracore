// Package testutil provides utilities for mocking JSON-RPC HTTP clients in tests.
package testutil

import (
	"bytes"
	"io"
	"net/http"
)

// JSONRPCMockClient is a mock implementation of an HTTP client for JSON-RPC requests, capturing calls and returning predefined responses.
type JSONRPCMockClient struct {
	DoFunc       func(req *http.Request) (*http.Response, error)
	Spy          *http.Request
	RequestCount int
}

// Do sends the HTTP request using the mock client, invoking DoFunc if set or returning a default empty response.
func (m *JSONRPCMockClient) Do(req *http.Request) (*http.Response, error) {
	if m.DoFunc != nil {
		return m.DoFunc(req)
	}
	// just in case you want default correct return value
	return &http.Response{}, nil
}

// MockResponse returns a DoFunc that captures the request in the mock client and returns an HTTP response with the given body and status code.
func MockResponse(resString string, statusCode int, m *JSONRPCMockClient) func(req *http.Request) (*http.Response, error) {
	return func(req *http.Request) (*http.Response, error) {
		m.Spy = req
		return &http.Response{
			StatusCode: statusCode,
			Body:       io.NopCloser(bytes.NewReader([]byte(resString))),
		}, nil
	}
}
