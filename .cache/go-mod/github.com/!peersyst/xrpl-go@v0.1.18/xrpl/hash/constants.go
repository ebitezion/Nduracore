// Package hash provides constants for prefixes used in hashing XRPL objects.
package hash

// Prefix for hashing functions.
//
// These prefixes are inserted before the source material used to
// generate various hashes. This is done to put each hash in its own
// "space." This way, two different types of objects with the
// same binary data will produce different hashes.
//
// Each prefix is a 4-byte value with the last byte set to zero
// and the first three bytes formed from the ASCII equivalent of
// some arbitrary string. For example "TXN".

const (
	// TransactionPrefix is the 4-byte prefix for hashing a transaction plus signature
	// to generate the transaction ID ('TXN').
	TransactionPrefix uint32 = 0x54584E00
)
