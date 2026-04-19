package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDestinationTag(t *testing.T) {
	tests := []struct {
		input    uint32
		expected uint32
	}{
		{input: 12345, expected: 12345},
		{input: 0, expected: 0},
		{input: 4294967295, expected: 4294967295},
	}

	for _, test := range tests {
		result := DestinationTag(test.input)
		require.Equal(t, test.expected, *result, "DestinationTag(%d) = %d; expected %d", test.input, *result, test.expected)
	}
}
