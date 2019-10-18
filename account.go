// Package wallet defines the domain model for the wallet system
package wallet

import (
	"context"
	"errors"

	"github.com/cockroachdb/apd"
	uuid "github.com/satori/go.uuid"

	"github.com/xsleonard/gokit-example/decimal"
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

var (
	// ErrEmptyAccountID is returned when creating an account without an ID
	ErrEmptyAccountID = errors.New("Account ID must not be empty")
	// ErrEmptyPaymentID is returned when creating an account without an ID
	ErrEmptyPaymentID = errors.New("Payment ID must not be empty")
	// ErrInvalidCurrency is returned for unrecognized currency codes
	ErrInvalidCurrency = errors.New("Invalid currency code")
	// ErrInvalidAmount is returned if the amount requested to transfer
	// is not greater than 0
	ErrInvalidAmount = errors.New("Transfer amount must be positive")
)

// Account represents a user account in the wallet system
type Account struct {
	ID       uuid.UUID    `db:"id"`
	Balance  *apd.Decimal `db:"balance"`
	Currency string       `db:"currency"`
}

// NewAccount creates an account
func NewAccount(id, balance, currency string) (*Account, error) {
	if id == "" {
		return nil, ErrEmptyAccountID
	}

	uid, err := uuid.FromString(id)
	if err != nil {
		return nil, err
	}

	if !IsValidCurrency(currency) {
		return nil, ErrInvalidCurrency
	}

	decBal, err := decimal.ParseCurrency(balance)
	if err != nil {
		return nil, err
	}

	return &Account{
		ID:       uid,
		Balance:  decBal,
		Currency: currency,
	}, nil
}

// AccountRepository is the storage interface for accounts
type AccountRepository interface {
	Store(ctx context.Context, account *Account) error
	Get(ctx context.Context, id string) (*Account, error)
	All(ctx context.Context) ([]Account, error)
}

// Payment represent a transfer from one account to another.
// A payment with a null "From" field is a credit to the "To" account.
type Payment struct {
	ID     uuid.UUID     `db:"id"`
	To     uuid.UUID     `db:"to_account_id"`
	From   uuid.NullUUID `db:"from_account_id"`
	Amount *apd.Decimal  `db:"amount"`
}

// FromUUIDString returns the From field's UUID string if set,
// otherwise returns the empty string
func (p Payment) FromUUIDString() string {
	if p.From.Valid {
		return p.From.UUID.String()
	}
	return ""
}

// NewPayment creates a Payment
func NewPayment(id uuid.UUID, to Account, from *Account, amount *apd.Decimal) (*Payment, error) {
	if amount == nil {
		return nil, errors.New("amount must not be nil")
	}

	p := &Payment{
		ID:     id,
		To:     to.ID,
		Amount: amount,
	}

	if from != nil {
		p.From.UUID = from.ID
		p.From.Valid = true
	}

	return p, nil
}

// PaymentRepository is the storage interface for payments
type PaymentRepository interface {
	Store(ctx context.Context, payment *Payment) error
	All(ctx context.Context) ([]Payment, error)
}

// TODO -- move back to Transfer?
// Service defines the payment transfer service
type Service interface {
	// Transfer transfers an amount of money from one account to another
	Transfer(ctx context.Context, to, from string, amount *apd.Decimal) (*Payment, error)
	// Payments returns all payments
	Payments(ctx context.Context) ([]Payment, error)
	// Accounts returns all accounts
	Accounts(ctx context.Context) ([]Account, error)
}
