// Package transfer defines the service layer for payment transfers between accounts
package transfer

import (
	"context"
	"errors"

	"github.com/cockroachdb/apd"
	uuid "github.com/satori/go.uuid"

	wallet "github.com/xsleonard/gokit-example"
)

var (
	// ErrDifferentCurrency is returned if a transfer is requested between accounts
	// that have different currencies
	ErrDifferentCurrency = errors.New("Transfers must use the same currency")
	// ErrSameAccount is returned if a transfer's sender and receiver are the same account
	ErrSameAccount = errors.New("Transfers must be between different accounts")
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

func (s service) Transfer(ctx context.Context, to, from string, amount *apd.Decimal) (*wallet.Payment, error) {
	// TODO -- use db txs
	// TODO -- wrap errors?

	// TODO -- error checking in the PaymentRepository level?
	// The amount must be > 0
	if amount.Sign() != 1 {
		return nil, wallet.ErrInvalidAmount
	}

	if to == from {
		return nil, ErrSameAccount
	}

	// TODO -- create NOT FOUND errors here
	// Check that the accounts exist
	toAccount, err := s.accounts.Get(ctx, to)
	if err != nil {
		return nil, err
	}

	fromAccount, err := s.accounts.Get(ctx, from)
	if err != nil {
		return nil, err
	}

	// Transfers between accounts of different currencies is not allowed
	if toAccount.Currency != fromAccount.Currency {
		return nil, ErrDifferentCurrency
	}

	paymentID, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	p, err := wallet.NewPayment(paymentID, *toAccount, fromAccount, amount)
	if err != nil {
		return nil, err
	}

	if err := s.payments.Store(ctx, p); err != nil {
		return nil, err
	}

	if err := s.accounts.Store(ctx, fromAccount); err != nil {
		return nil, err
	}

	if err := s.accounts.Store(ctx, toAccount); err != nil {
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
