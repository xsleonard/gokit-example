// Package transfer defines the service layer for payment transfers between accounts
package transfer

import (
	"context"
	"errors"

	"github.com/shopspring/decimal"

	wallet "github.com/xsleonard/gokit-example"
)

var (
	// ErrInvalidAmount is returned if the amount requested to transfer
	// is not greater than 0
	ErrInvalidAmount = errors.New("Transfer amount must be positive")
	// ErrDifferentCurrency is returned if a transfer is requested between accounts
	// that have different currencies
	ErrDifferentCurrency = errors.New("Transfers must use the same currency")
	// ErrInsufficientBalance is returned if an account's balance is less than
	// an amount requested to be transferred
	ErrInsufficientBalance = errors.New("Account has an insufficient balance")
)

type service struct {
	accounts wallet.AccountRepository
	payments wallet.PaymentRepository
}

// NewService creates a wallet.Service
func NewService(accounts wallet.AccountRepository, payments wallet.PaymentRepository) wallet.Service {
	return service{
		accounts: accounts,
		payments: payments,
	}
}

func (s service) Transfer(ctx context.Context, to, from string, amount decimal.Decimal) (*wallet.Payment, error) {
	// TODO -- use db txs
	// TODO -- wrap errors?

	// TODO -- error checking in the PaymentRepository level?
	// The amount must be > 0
	if !amount.IsPositive() {
		return nil, ErrInvalidAmount
	}

	// Check that the accounts exist
	toAccount, err := s.accounts.Get(ctx, to)
	if err != nil {
		return nil, err
	}

	fromAccount, err := s.accounts.Get(ctx, from)
	if err != nil {
		return nil, err
	}

	// Check that the account has enough money to send
	fromBalance, err := s.payments.Balance(ctx, from)
	if err != nil {
		return nil, err
	}

	if fromBalance.LessThan(amount) {
		return nil, ErrInsufficientBalance
	}

	// Transfers between accounts of different currencies is not allowed
	if toAccount.Currency != fromAccount.Currency {
		return nil, ErrDifferentCurrency
	}

	p := wallet.NewPayment(toAccount, fromAccount, amount)

	if err := s.payments.Store(ctx, p); err != nil {
		return nil, err
	}

	return p, nil
}

func (s service) Payments(ctx context.Context) ([]wallet.Payment, error) {
	return s.payments.All(ctx)
}

func (s service) Accounts(ctx context.Context) ([]wallet.Account, error) {
	return s.accounts.All(ctx)
}
