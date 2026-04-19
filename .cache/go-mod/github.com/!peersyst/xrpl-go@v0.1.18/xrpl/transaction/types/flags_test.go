package types

import (
	"testing"
)

func TestIsFlagEnabled(t *testing.T) {
	var flags uint32
	const flag1 = 0x00010000
	const flag2 = 0x00020000

	setup := func() {
		flags = 0x00000000
	}

	t.Run("verifies a flag is enabled", func(t *testing.T) {
		setup()
		flags |= flag1 | flag2
		if !IsFlagEnabled(flags, flag1) {
			t.Errorf("expected flag1 to be enabled")
		}
	})

	t.Run("verifies a flag is not enabled", func(t *testing.T) {
		setup()
		flags |= flag2
		if IsFlagEnabled(flags, flag1) {
			t.Errorf("expected flag1 to be not enabled")
		}
	})
}
