//revive:disable:var-naming
package types

import "math/big"

// IsFlagEnabled performs bitwise AND (&) to check if a flag is enabled within Flags (as a number).
func IsFlagEnabled(flags, checkFlag uint32) bool {
	flagsBigInt := new(big.Int).SetUint64(uint64(flags))
	checkFlagBigInt := new(big.Int).SetUint64(uint64(checkFlag))
	return new(big.Int).And(flagsBigInt, checkFlagBigInt).Cmp(checkFlagBigInt) == 0
}
