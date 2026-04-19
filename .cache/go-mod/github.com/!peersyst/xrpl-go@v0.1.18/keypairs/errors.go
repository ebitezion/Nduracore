package keypairs

import "errors"

var (
	// keypairs

	// ErrInvalidSignature is returned when the derived keypair did not generate a verifiable signature.
	ErrInvalidSignature = errors.New("derived keypair did not generate verifiable signature")

	// crypto

	// ErrInvalidCryptoImplementation is returned when the key does not match any crypto implementation.
	ErrInvalidCryptoImplementation = errors.New("not a valid crypto implementation")
)
