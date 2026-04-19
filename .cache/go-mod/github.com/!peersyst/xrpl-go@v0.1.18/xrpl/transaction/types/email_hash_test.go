package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEmailHash(t *testing.T) {
	tests := []struct {
		name  string
		input Hash128
		want  Hash128
	}{
		{
			name:  "ValidHash",
			input: Hash128("validCustomHash"),
			want:  Hash128("validCustomHash"),
		},
		{
			name:  "EmptyHash",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EmailHash(tt.input)
			require.Equal(t, tt.want, *result)
		})
	}
}
