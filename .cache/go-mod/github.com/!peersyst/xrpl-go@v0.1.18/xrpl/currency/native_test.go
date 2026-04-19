package currency

import (
	"errors"
	"testing"
)

func TestXrpToDrops(t *testing.T) {
	tt := []struct {
		name        string
		xrp         string
		drops       string
		expectedErr error
	}{
		{
			name:        "XRP to Drops (no decimals)",
			xrp:         "1",
			drops:       "1000000",
			expectedErr: nil,
		},
		{
			name:        "XRP to Drops (decimals)(1)",
			xrp:         "3.456789",
			drops:       "3456789",
			expectedErr: nil,
		},
		{
			name:        "XRP to Drops (decimals)(2)",
			xrp:         "0.000001",
			drops:       "1",
			expectedErr: nil,
		},
		{
			name:        "XRP to Drops (decimals)(3)",
			xrp:         "0.000000",
			drops:       "0",
			expectedErr: nil,
		},
		{
			name:        "XRP to Drops (decimals)(4)",
			xrp:         "3.400000",
			drops:       "3400000",
			expectedErr: nil,
		},
		{
			name:        "XRP to Drops (scientific notation)",
			xrp:         "1e-6",
			drops:       "1",
			expectedErr: nil,
		},
		{
			name:        "XRP to Drops (too many decimals)",
			xrp:         "0.0000001",
			drops:       "1",
			expectedErr: errors.New("xrp to drops: value has too many decimals"),
		},
		{
			name:        "XRP to Drops (invalid input)",
			xrp:         "abc",
			drops:       "",
			expectedErr: errors.New("strconv.ParseFloat: parsing \"abc\": invalid syntax"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			drops, err := XrpToDrops(tc.xrp)
			if tc.expectedErr != nil {
				if err == nil || err.Error() != tc.expectedErr.Error() {
					t.Errorf("Expected error: %v, got: %v", tc.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if drops != tc.drops {
					t.Errorf("Expected %s drops, got %s", tc.drops, drops)
				}
			}
		})
	}
}

func TestDropsToXrp(t *testing.T) {
	tt := []struct {
		name        string
		drops       string
		xrp         string
		expectedErr error
	}{
		{
			name:        "Drops to XRP (whole number)",
			drops:       "1000000",
			xrp:         "1",
			expectedErr: nil,
		},
		{
			name:        "Drops to XRP (decimal)",
			drops:       "1234567",
			xrp:         "1.234567",
			expectedErr: nil,
		},
		{
			name:        "Drops to XRP (zero)",
			drops:       "0",
			xrp:         "0",
			expectedErr: nil,
		},
		{
			name:        "Drops to XRP (small amount)",
			drops:       "1",
			xrp:         "0.000001",
			expectedErr: nil,
		},
		{
			name:        "Drops to XRP (large amount)",
			drops:       "123456789000000",
			xrp:         "123456789",
			expectedErr: nil,
		},
		{
			name:        "Drops to XRP (invalid input)",
			drops:       "abc",
			xrp:         "",
			expectedErr: errors.New("strconv.ParseUint: parsing \"abc\": invalid syntax"),
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			xrp, err := DropsToXrp(tc.drops)
			if tc.expectedErr != nil {
				if err == nil || err.Error() != tc.expectedErr.Error() {
					t.Errorf("Expected error: %v, got: %v", tc.expectedErr, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if xrp != tc.xrp {
					t.Errorf("Expected %s XRP, got %s", tc.xrp, xrp)
				}
			}
		})
	}
}
