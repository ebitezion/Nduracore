package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMessageKey(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"testKey1", "testKey1"},
		{"anotherKey", "anotherKey"},
		{"", ""},
	}

	for _, test := range tests {
		result := MessageKey(test.input)
		require.Equal(t, test.expected, *result)
	}
}
