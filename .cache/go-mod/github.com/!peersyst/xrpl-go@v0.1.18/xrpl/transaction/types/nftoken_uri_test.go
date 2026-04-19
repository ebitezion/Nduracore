package types

import (
	"testing"
)

func TestNFTokenURI_String(t *testing.T) {
	tests := []struct {
		name     string
		input    NFTokenURI
		expected string
	}{
		{
			name:     "valid URI",
			input:    NFTokenURI("https://example.com/nft/1"),
			expected: "https://example.com/nft/1",
		},
		{
			name:     "empty URI",
			input:    NFTokenURI(""),
			expected: "",
		},
		{
			name:     "URI with special characters",
			input:    NFTokenURI("https://example.com/nft/1?query=param&another=param"),
			expected: "https://example.com/nft/1?query=param&another=param",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.input.String(); got != tt.expected {
				t.Errorf("NFTokenURI.String(), got: %v but we want: %v", got, tt.expected)
			}
		})
	}
}
