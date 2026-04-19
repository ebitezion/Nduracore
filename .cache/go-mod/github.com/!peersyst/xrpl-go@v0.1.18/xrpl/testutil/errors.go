package testutil

import "errors"

var (
	// ErrMarshalingPaymentFlattened is returned when marshaling the flattened payment fails.
	ErrMarshalingPaymentFlattened = errors.New("error marshaling payment flattened")

	// ErrUnmarshalingExpected is returned when unmarshaling expected data fails.
	ErrUnmarshalingExpected = errors.New("error unmarshaling expected")

	// ErrMarshalingExpectedPaymentObject is returned when marshaling the expected payment object fails.
	ErrMarshalingExpectedPaymentObject = errors.New("error marshaling expected payment object")

	// ErrFlattenedAndExpectedJSONNotEqual is returned when flattened and expected JSON are not equal.
	ErrFlattenedAndExpectedJSONNotEqual = errors.New("the flattened and expected JSON are not equal")
)
