package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDomain(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"example.com", "example.com"},
		{"testdomain.org", "testdomain.org"},
		{"", ""},
	}

	for _, test := range tests {
		result := Domain(test.input)
		require.Equal(t, test.expected, *result, "Domain(%s) = %s; expected %s", test.input, *result, test.expected)
	}
}
