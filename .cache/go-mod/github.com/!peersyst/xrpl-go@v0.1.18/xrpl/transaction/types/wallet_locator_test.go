package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWalletLocator(t *testing.T) {
	tests := []struct {
		name     string
		value    Hash256
		expected Hash256
	}{
		{
			name:     "Test with non-zero value",
			value:    Hash256("locator1"),
			expected: Hash256("locator1"),
		},
		{
			name:     "Test with empty value",
			value:    Hash256(""),
			expected: Hash256(""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WalletLocator(tt.value)
			require.Equal(t, tt.expected, *got)
		})
	}
}
