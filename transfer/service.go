// Package transfer defines the service layer for payment transfers between accounts
package transfer

import (
	"context"
	"errors"

	"github.com/cockroachdb/apd"
	"github.com/jmoiron/sqlx"
	uuid "github.com/satori/go.uuid"

	wallet "github.com/xsleonard/gokit-example"
	"github.com/xsleonard/gokit-example/decimal"
)

var (
	// errSameAccount is returned if a transfer's sender and receiver are the same account
	errSameAccount = errors.New("Transfers must be between different accounts")
	// errInsufficientBalance is returned if an account's balance is less than
	// an amount requested to be transferred
	errInsufficientBalance = errors.New("Account has an insufficient balance")
	// errDifferentCurrency is returned if a transfer is requested between accounts
	// that have different currencies
	errDifferentCurrency = errors.New("Transfers must use the same currency")
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

func (s service) Transfer(ctx context.Context, to, from uuid.UUID, amount *apd.Decimal) (*wallet.Payment, error) {
	if err := decimal.ValidateTransferAmount(amount); err != nil {
		return nil, err
	}

	if uuid.Equal(to, from) {
		return nil, errSameAccount
	}

	paymentID, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	p := &wallet.Payment{
		ID:     paymentID,
		To:     to,
		From:   &from,
		Amount: amount,
	}

	if err := s.payments.WithTx(ctx, func(ctx context.Context, tx *sqlx.Tx) error {
		return s.transferTx(ctx, tx, p)
	}); err != nil {
		return nil, err
	}

	return p, nil
}

func (s service) transferTx(ctx context.Context, tx *sqlx.Tx, p *wallet.Payment) error {
	// Fetch the accounts, checking that they exist
	toAccount, err := s.accounts.GetTx(ctx, tx, p.To)
	if err != nil {
		return err
	}

	fromAccount, err := s.accounts.GetTx(ctx, tx, *p.From)
	if err != nil {
		return err
	}

	// Transfers between accounts of different currencies is not allowed
	if toAccount.Currency != fromAccount.Currency {
		return errDifferentCurrency
	}

	// The account must have sufficient balance
	if fromAccount.Balance.Cmp(p.Amount) < 0 {
		return errInsufficientBalance
	}

	return s.payments.StoreTx(ctx, tx, p)
}

func (s service) Payments(ctx context.Context) ([]wallet.Payment, error) {
	return s.payments.All(ctx)
}

func (s service) Accounts(ctx context.Context) ([]wallet.Account, error) {
	return s.accounts.All(ctx)
}
