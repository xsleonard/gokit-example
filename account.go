// Package wallet defines the domain model for the wallet system
package wallet

import (
	"context"
	"errors"
	"time"

	"github.com/shopspring/decimal"

	"github.com/xsleonard/gokit-example/util"
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

func isValidCurrency(string string) bool {
	switch string {
	case USD, EUR, SGD, GBP:
		return true
	}
	return false
}

var (
	// ErrEmptyAccountID is returned when creating an account without an ID
	ErrEmptyAccountID = errors.New("Account ID must not be empty")
	// ErrInvalidCurrency is returned for unrecognized currency codes
	ErrInvalidCurrency = errors.New("Invalid currency code")
)

// Account represents a user account in the wallet system
type Account struct {
	ID       string
	Balance  decimal.Decimal
	Currency string
}

// NewAccount creates an account
func NewAccount(id, balance, currency string) (*Account, error) {
	if id == "" {
		return nil, ErrEmptyAccountID
	}

	if !isValidCurrency(currency) {
		return nil, ErrInvalidCurrency
	}

	decBal, err := util.ParseAmount(balance)
	if err != nil {
		return nil, err
	}

	return &Account{
		ID:       id,
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

// type Direction string

// const (
// 	Incoming Direction = "incoming"
// 	Outgoing Direction = "outgoing"
// )

// Payment represent a transfer from one account to another.
// A payment can either be "incoming" or "outgoing".
// When a user makes a payment action, two Payment entries are created:
// one "incoming" and one "outgoing".
type Payment struct {
	To     string
	From   string
	Amount decimal.Decimal
	// Direction Direction
	Timestamp time.Time
}

// NewPayment creates a Payment
func NewPayment(to, from *Account, amount decimal.Decimal) *Payment {
	return &Payment{
		To:        to.ID,
		From:      from.ID,
		Amount:    amount,
		Timestamp: time.Now().UTC(), // TODO -- get it from postgres?
	}
}

// PaymentRepository is the storage interface for payments
type PaymentRepository interface {
	Store(ctx context.Context, payment *Payment) error
	All(ctx context.Context) ([]Payment, error)
	Balance(ctx context.Context, accountID string) (decimal.Decimal, error)
}

// Service defines the payment transfer service
type Service interface {
	// Transfer transfers an amount of money from one account to another
	Transfer(ctx context.Context, to, from string, amount decimal.Decimal) (*Payment, error)
	// Payments returns all payments
	Payments(ctx context.Context) ([]Payment, error)
	// Accounts returns all accounts
	Accounts(ctx context.Context) ([]Account, error)
}
