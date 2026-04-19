package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWalletSize(t *testing.T) {
	tests := []struct {
		input    uint32
		expected uint32
	}{
		{input: 0, expected: 0},
		{input: 1, expected: 1},
		{input: 100, expected: 100},
		{input: 4294967295, expected: 4294967295}, // Max uint32 value
	}

	for _, test := range tests {
		result := WalletSize(test.input)
		require.Equal(t, test.expected, *result)
	}
}
