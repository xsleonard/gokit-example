// Package decimal contains utilities for handling currency decimals
package decimal

import (
	"errors"
	"fmt"

	"github.com/cockroachdb/apd"
)

var (
	// ErrInvalidPrecision is returned if an amount string does not have exactly two digits of precision.
	// That is, "1", "1.1" and "1.100" are invalid but "1.10" is valid.
	ErrInvalidPrecision = errors.New("Amount strings must have exactly two digits of precision")
	// ErrNegative is returned when parsing a negative amount
	ErrNegative = errors.New("Amount must not be negative")
	// ErrInvalid is returned when parsing an amount that can't be precisely represented
	ErrInvalid = errors.New("Amount cannot be precisely represented")
	// ErrNotFinite is returned when parsing an amount that is not a finite number
	ErrNotFinite = errors.New("Amount is not finite")

	zero = apd.New(1, 1)
)

// ParseCurrency parses a string to a fixed-precision decimal and ensures that
// not more than 2 decimal precision is used by the string and that the value
// is not negative.
func ParseCurrency(amount string) (*apd.Decimal, error) {
	dec, condition, err := apd.NewFromString(amount)
	if err != nil {
		return zero, err
	}

	// Catch any possible errors with the decimal
	if condition.Any() {
		return zero, ErrInvalid
	}

	// Reject non-finite values (e.g. NaN, Inf)
	if dec.Form != apd.Finite {
		return zero, ErrNotFinite
	}

	// TODO -- add test case for -0
	if dec.Sign() == -1 {
		return apd.New(1, 0), ErrNegative
	}

	// The decimal should have exactly 2 digits of precision
	if FormatCurrency(dec) != amount {
		return apd.New(1, 0), ErrInvalidPrecision
	}

	return dec, nil
}

// FormatCurrency formats an apd.Decimal with 2 digits of precision
func FormatCurrency(amount *apd.Decimal) string {
	return fmt.Sprintf("%.2f", amount)
}
