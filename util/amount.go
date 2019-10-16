// Package util contains common shared utility methods
package util

import (
	"errors"

	"github.com/shopspring/decimal"
)

var (
	// ErrInvalidAmountPrecision is returned if an amount string does not have exactly two digits of precision.
	// That is, "1", "1.1" and "1.100" are invalid but "1.10" is valid.
	ErrInvalidAmountPrecision = errors.New("Amount strings must have exactly two digits of precision")
	// ErrNegativeAmount is returned when parsing a negative amount
	ErrNegativeAmount = errors.New("Amount must not be negative")
)

// ParseAmount parses a string to a fixed-precision decimal and ensures that
// not more than 2 decimal precision is used by the string and that the value
// is not negative.
func ParseAmount(amount string) (decimal.Decimal, error) {
	dec, err := decimal.NewFromString(amount)
	if err != nil {
		return decimal.Zero, err
	}
	if dec.IsNegative() {
		return decimal.Zero, ErrNegativeAmount
	}
	if dec.StringFixed(2) != amount {
		return decimal.Zero, ErrInvalidAmountPrecision
	}

	return dec, nil
}
