// Package decimal contains utilities for handling currency decimals
package decimal

import (
	"errors"

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
	// ErrAmountNotMoreThanZero is returned when an amount is not > 0
	ErrAmountNotMoreThanZero = errors.New("Amount must be greater than 0")
	// ErrAmountNil is returned if the amount is nil
	ErrAmountNil = errors.New("Amount must not be nil")
)

// ParseCurrency parses a string to a fixed-precision decimal and ensures that
// not more than 2 decimal precision is used by the string and that the value
// is not negative.
func ParseCurrency(amount string) (*apd.Decimal, error) {
	dec, condition, err := apd.NewFromString(amount)
	if err != nil {
		return nil, err
	}

	// Catch any possible errors with the decimal
	if condition.Any() {
		return nil, ErrInvalid
	}

	// Reject non-finite values (e.g. NaN, Inf)
	if dec.Form != apd.Finite {
		return nil, ErrNotFinite
	}

	// TODO -- add test case for -0
	if dec.Sign() == -1 {
		return nil, ErrNegative
	}

	// The decimal should not have more than 2 digits of precision
	if dec.Exponent < -2 {
		return nil, ErrInvalidPrecision
	}

	return dec, nil
}

// ValidateTransferAmount validates a decimal amount for transfers
func ValidateTransferAmount(amount *apd.Decimal) error {
	if amount == nil {
		return ErrAmountNil
	}

	if amount.Form != apd.Finite {
		return ErrNotFinite
	}

	if amount.Sign() != 1 {
		return ErrAmountNotMoreThanZero
	}

	if amount.Exponent < -2 {
		return ErrInvalidPrecision
	}

	return nil
}
