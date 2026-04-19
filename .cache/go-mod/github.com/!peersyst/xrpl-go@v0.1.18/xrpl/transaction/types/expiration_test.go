package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExpiration(t *testing.T) {
	tests := []struct {
		input    uint32
		expected uint32
	}{
		{input: 0, expected: 0},
		{input: 1234567890, expected: 1234567890},
		{input: 946684800, expected: 946684800},   // Year 2000 timestamp
		{input: 4294967295, expected: 4294967295}, // Max uint32 value
	}

	for _, test := range tests {
		result := Expiration(test.input)
		require.Equal(t, test.expected, *result)
	}
}
