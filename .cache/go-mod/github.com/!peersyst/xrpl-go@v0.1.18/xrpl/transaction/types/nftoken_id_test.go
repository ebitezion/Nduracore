package types

import "testing"

func TestNFTokenIDString(t *testing.T) {
	tests := []struct {
		name      string
		nfTokenID NFTokenID
		want      string
	}{
		{
			name:      "Empty NFTokenID",
			nfTokenID: NFTokenID(""),
			want:      "",
		},
		{
			name:      "Non-empty NFTokenID",
			nfTokenID: NFTokenID("1234567890abcdef"),
			want:      "1234567890abcdef",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.nfTokenID.String(); got != tt.want {
				t.Errorf("NFTokenID.String(), got: %v but we want %v", got, tt.want)
			}
		})
	}
}
