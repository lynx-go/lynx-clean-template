package vo

import (
	"github.com/lynx-go/lynx-clean-template/pkg/errors"
)

// Money represents a monetary amount in a specific currency.
// Amount is stored in the smallest currency unit (e.g., cents for USD).
type Money struct {
	amount   int64
	currency string
}

// NewMoney creates a Money value object.
// currency should be an ISO-4217 code (e.g. "USD", "CNY").
func NewMoney(amount int64, currency string) (Money, error) {
	if currency == "" {
		return Money{}, errors.Cause("currency must not be empty")
	}
	return Money{amount: amount, currency: currency}, nil
}

// Amount returns the raw amount in the smallest currency unit.
func (m Money) Amount() int64 { return m.amount }

// Currency returns the ISO-4217 currency code.
func (m Money) Currency() string { return m.currency }

// Equals returns true when both amount and currency are equal.
func (m Money) Equals(other Money) bool {
	return m.amount == other.amount && m.currency == other.currency
}

// Add returns a new Money that is the sum of both values.
// Returns an error when currencies differ.
func (m Money) Add(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, errors.Cause("cannot add money with different currencies")
	}
	return Money{amount: m.amount + other.amount, currency: m.currency}, nil
}

// IsZero returns true when the amount is zero.
func (m Money) IsZero() bool { return m.amount == 0 }
