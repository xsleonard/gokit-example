// Package wallet defines the domain model for the wallet system
package wallet

import (
	"context"
	"errors"

	"github.com/cockroachdb/apd"
	"github.com/jmoiron/sqlx"
	uuid "github.com/satori/go.uuid"
)

var (
	// ErrNoAccount is returned when an account is not found in storage by ID
	ErrNoAccount = errors.New("Account does not exist")
)

const (
	// USD United States Dollar
	USD string = "USD"
	// EUR Euro
	EUR string = "EUR"
	// SGD Singapore Dollar
	SGD string = "SGD"
	// GBP British Pound
	GBP string = "GBP"
)

// IsValidCurrency returns true if the currency code is valid
func IsValidCurrency(currency string) bool {
	switch currency {
	case USD, EUR, SGD, GBP:
		return true
	}
	return false
}

// Account represents a user account in the wallet system
type Account struct {
	ID       uuid.UUID
	Balance  *apd.Decimal
	Currency string
}

// AccountRepository is the storage interface for accounts
type AccountRepository interface {
	Store(ctx context.Context, account *Account) error
	GetTx(ctx context.Context, tx *sqlx.Tx, id uuid.UUID) (*Account, error)
	All(ctx context.Context) ([]Account, error)
}

// Payment represent a transfer from one account to another.
// A payment with a null "From" field is a credit to the "To" account.
type Payment struct {
	ID     uuid.UUID
	To     uuid.UUID
	From   *uuid.UUID
	Amount *apd.Decimal
}

// FromUUIDString returns the From field's UUID string if set,
// otherwise returns the empty string
func (p Payment) FromUUIDString() string {
	if p.From == nil {
		return ""
	}
	return p.From.String()
}

// PaymentRepository is the storage interface for payments
type PaymentRepository interface {
	WithTx(ctx context.Context, f func(ctx context.Context, tx *sqlx.Tx) error) error
	StoreTx(ctx context.Context, tx *sqlx.Tx, payment *Payment) error
	Store(ctx context.Context, payment *Payment) error
	All(ctx context.Context) ([]Payment, error)
}

// Service defines the payment transfer service
type Service interface {
	// Transfer transfers an amount of money from one account to another
	Transfer(ctx context.Context, to, from uuid.UUID, amount *apd.Decimal) (*Payment, error)
	// Payments returns all payments
	Payments(ctx context.Context) ([]Payment, error)
	// Accounts returns all accounts
	Accounts(ctx context.Context) ([]Account, error)
}
