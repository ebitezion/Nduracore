//revive:disable:var-naming
package types

// OfferResultFlags defines the bit flags associated with an offer result.
type OfferResultFlags uint

// OfferResult represents the result of an offer in the XRP Ledger order book.
type OfferResult struct {
	Flags    OfferResultFlags `json:"flags"`
	Sequence uint             `json:"seq"`
	// TakerGets  types.CurrencyAmount `json:"taker_gets"`
	TakerGets any `json:"taker_gets"`
	// TakerPays  types.CurrencyAmount `json:"taker_pays"`
	TakerPays  any    `json:"taker_pays"`
	Quality    string `json:"quality"`
	Expiration uint   `json:"expiration,omitempty"`
}
