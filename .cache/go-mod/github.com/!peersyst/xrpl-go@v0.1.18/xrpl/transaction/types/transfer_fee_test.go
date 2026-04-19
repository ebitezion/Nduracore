package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTransferFee(t *testing.T) {
	tests := []struct {
		input    uint16
		expected uint16
	}{
		{input: 0, expected: 0},
		{input: 314, expected: 314},
		{input: 50000, expected: 50000}, // Max transfer fee
		{input: 65535, expected: 65535}, // Max uint16 value
	}

	for _, test := range tests {
		result := TransferFee(test.input)
		require.Equal(t, test.expected, *result)
	}
}
