package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTickSize(t *testing.T) {
	tests := []struct {
		name  string
		value uint8
		want  uint8
	}{
		{"Valid tick size 3", 3, 3},
		{"Valid tick size 15", 15, 15},
		{"Valid tick size 0", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TickSize(tt.value)
			require.Equal(t, tt.want, *got)
		})
	}
}
