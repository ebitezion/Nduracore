// Package testutil provides helper functions for binary codec tests.
package testutil

import (
	"testing"

	definitions "github.com/Peersyst/xrpl-go/binary-codec/definitions"
)

// GetFieldInstance retrieves a FieldInstance by field name, failing the test if not found.
func GetFieldInstance(t *testing.T, fieldName string) definitions.FieldInstance {
	t.Helper()
	fi, err := definitions.Get().GetFieldInstanceByFieldName(fieldName)
	if err != nil {
		t.Fatalf("FieldInstance with FieldName %v", fieldName)
	}
	return *fi
}
