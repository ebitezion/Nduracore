package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNFTokenMinter(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"rExampleAddress1", "rExampleAddress1"},
		{"rExampleAddress2", "rExampleAddress2"},
		{"", ""},
	}

	for _, test := range tests {
		result := NFTokenMinter(test.input)
		require.Equal(t, test.expected, *result)
	}
}
