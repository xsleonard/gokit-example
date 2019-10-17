// Package postgres implements storage using postgres
package postgres

import (
	"context"
	"database/sql"

	"github.com/go-kit/kit/log"
	"github.com/jmoiron/sqlx"

	wallet "github.com/xsleonard/gokit-example"
)

type accountRepository struct {
	db     *sqlx.DB
	logger log.Logger
}

// NewAccountRepository creates a wallet.AccountRepository that uses postgres for storage
func NewAccountRepository(db *sqlx.DB, logger log.Logger) wallet.AccountRepository {
	return &accountRepository{
		db:     db,
		logger: logger,
	}
}

func (r *accountRepository) Store(ctx context.Context, account *wallet.Account) error {
	if account.ID == "" {
		return wallet.ErrEmptyAccountID
	}
	if !wallet.IsValidCurrency(account.Currency) {
		return wallet.ErrInvalidCurrency
	}

	return insideTx(ctx, r.logger, r.db, func(ctx context.Context, tx *sqlx.Tx) error {
		q := `insert into account (id, currency) values (?, ?)`
		_, err := tx.ExecContext(ctx, q, account.ID, account.Currency)
		return err
	})
}

func (r *accountRepository) Get(ctx context.Context, id string) (*wallet.Account, error) {
	var a wallet.Account
	if err := insideTx(ctx, r.logger, r.db, func(ctx context.Context, tx *sqlx.Tx) error {
		row := tx.QueryRowxContext(ctx, `select balance, currency from account_balance where id=$1`, id)
		return row.StructScan(&a)
	}); err != nil {
		return nil, err
	}

	return &a, nil
}

func (r *accountRepository) All(ctx context.Context) ([]wallet.Account, error) {
	var rows *sqlx.Rows
	if err := insideTx(ctx, r.logger, r.db, func(ctx context.Context, tx *sqlx.Tx) error {
		var err error
		rows, err = tx.QueryxContext(ctx, `select id, balance, currency from account_balance`)
		return err
	}); err != nil {
		return nil, err
	}

	// TODO -- are rows valid outside of the tx?
	var accounts []wallet.Account
	defer rows.Close()
	for rows.Next() {
		var a wallet.Account
		if err := rows.StructScan(&a); err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}

	return accounts, nil
}

type paymentRepository struct {
	db     *sqlx.DB
	logger log.Logger
}

// NewPaymentRepository creates a wallet.PaymentRepository that uses postgres for storage
func NewPaymentRepository(db *sqlx.DB, logger log.Logger) wallet.PaymentRepository {
	return &paymentRepository{
		db:     db,
		logger: logger,
	}
}

func (r *paymentRepository) Store(ctx context.Context, payment *wallet.Payment) error {
	if payment.ID == "" {
		return wallet.ErrEmptyPaymentID
	}

	return insideTx(ctx, r.logger, r.db, func(ctx context.Context, tx *sqlx.Tx) error {
		q := `insert into payment (id, from_account_id, to_account_id, amount) values (?, ?, ?)`
		_, err := tx.ExecContext(ctx, q, payment.ID, payment.From, payment.To, payment.Amount)
		return err
	})
}

func (r *paymentRepository) All(ctx context.Context) ([]wallet.Payment, error) {
	var rows *sqlx.Rows
	if err := insideTx(ctx, r.logger, r.db, func(ctx context.Context, tx *sqlx.Tx) error {
		var err error
		rows, err = tx.QueryxContext(ctx, `select id, from_account_id, to_account_id, amount, created_at from payment`)
		return err
	}); err != nil {
		return nil, err
	}

	// TODO -- are rows valid outside of the tx?
	var payments []wallet.Payment
	defer rows.Close()
	for rows.Next() {
		var p wallet.Payment
		if err := rows.StructScan(&p); err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}

	return payments, nil
}

func insideTx(ctx context.Context, logger log.Logger, db *sqlx.DB, f func(ctx context.Context, tx *sqlx.Tx) error) error {
	tx, err := db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelDefault,
		ReadOnly:  false,
	})
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil {
			log.With(logger, "err", err).Log("Postgres tx rollback failed")
		}
	}()

	if err := f(ctx, tx); err != nil {
		return err
	}

	return tx.Commit()
}
