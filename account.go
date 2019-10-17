// Package wallet defines the domain model for the wallet system
package wallet

import (
	"context"
	"errors"

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
	// ErrInsufficientBalance is returned if an account's balance is less than
	// an amount requested to be transferred
	ErrInsufficientBalance = errors.New("Account has an insufficient balance")
	// ErrInvalidAmount is returned if the amount requested to transfer
	// is not greater than 0
	ErrInvalidAmount = errors.New("Transfer amount must be positive")
)

// Account represents a user account in the wallet system
type Account struct {
	ID       string          `db:"id"`
	Balance  decimal.Decimal `db:"balance"`
	Currency string          `db:"currency"`
}

// NewAccount creates an account
func NewAccount(id, balance, currency string) (*Account, error) {
	if id == "" {
		return nil, ErrEmptyAccountID
	}

	if !IsValidCurrency(currency) {
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

// TransferIn adds to the account's balance
func (a *Account) TransferIn(amount decimal.Decimal) error {
	if !amount.IsPositive() {
		return ErrInvalidAmount
	}
	a.Balance = a.Balance.Add(amount)
	return nil
}

// TransferOut removes from the account's balance
func (a *Account) TransferOut(amount decimal.Decimal) error {
	if !amount.IsPositive() {
		return ErrInvalidAmount
	}
	if amount.GreaterThan(a.Balance) {
		return ErrInsufficientBalance
	}
	a.Balance = a.Balance.Sub(amount)
	return nil
}

// AccountRepository is the storage interface for accounts
type AccountRepository interface {
	Store(ctx context.Context, account *Account) error
	Get(ctx context.Context, id string) (*Account, error)
	All(ctx context.Context) ([]Account, error)
}

// Payment represent a transfer from one account to another.
type Payment struct {
	ID     string          `db:"id"`
	To     string          `db:"to_account_id"`
	From   string          `db:"from_account_id"`
	Amount decimal.Decimal `db:"amount"`
}

// NewPayment creates a Payment
func NewPayment(to, from *Account, amount decimal.Decimal) *Payment {
	return &Payment{
		ID:     "foo", // TODO -- use uuid
		To:     to.ID,
		From:   from.ID,
		Amount: amount,
	}
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
	Transfer(ctx context.Context, to, from string, amount decimal.Decimal) (*Payment, error)
	// Payments returns all payments
	Payments(ctx context.Context) ([]Payment, error)
	// Accounts returns all accounts
	Accounts(ctx context.Context) ([]Account, error)
}
