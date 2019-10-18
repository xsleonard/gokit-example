// Package postgres implements storage using postgres
package postgres

import (
	"context"
	"database/sql"

	"github.com/go-kit/kit/log"
	"github.com/jmoiron/sqlx"
	uuid "github.com/satori/go.uuid"

	wallet "github.com/xsleonard/gokit-example"
)

var nullUUID uuid.UUID

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
	if uuid.Equal(account.ID, nullUUID) {
		return wallet.ErrEmptyAccountID
	}
	if !wallet.IsValidCurrency(account.Currency) {
		return wallet.ErrInvalidCurrency
	}

	return insideTx(ctx, r.logger, r.db, func(ctx context.Context, tx *sqlx.Tx) error {
		q := `insert into account (id, currency) values ($1, $2)`
		_, err := tx.ExecContext(ctx, q, account.ID, account.Currency)
		return err
	})
}

func (r *accountRepository) Get(ctx context.Context, id string) (*wallet.Account, error) {
	row := r.db.QueryRowxContext(ctx, `select id, balance, currency from account_balance where id=$1`, id)

	var a wallet.Account
	if err := row.StructScan(&a); err != nil {
		if err == sql.ErrNoRows {
			return nil, wallet.ErrNoAccount
		}
		return nil, err
	}

	return &a, nil
}

func (r *accountRepository) All(ctx context.Context) ([]wallet.Account, error) {
	rows, err := r.db.QueryxContext(ctx, `select id, balance, currency from account_balance`)
	if err != nil {
		return nil, err
	}

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

func (r *paymentRepository) Store(ctx context.Context, p *wallet.Payment) error {
	if p.ID == uuid.Nil {
		return wallet.ErrEmptyPaymentID
	}

	r.logger.Log("id", p.ID, "from", p.From.UUID, "to", p.To, "amount", p.Amount)

	return insideTx(ctx, r.logger, r.db, func(ctx context.Context, tx *sqlx.Tx) error {
		q := `insert into payment (id, from_account_id, to_account_id, amount) values ($1, $2, $3, $4)`
		_, err := tx.ExecContext(ctx, q, p.ID, p.From, p.To, p.Amount)
		return err
	})
}

func (r *paymentRepository) All(ctx context.Context) ([]wallet.Payment, error) {
	rows, err := r.db.QueryxContext(ctx, `select id, from_account_id, to_account_id, amount from payment`)
	if err != nil {
		return nil, err
	}

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

func insideTx(ctx context.Context, logger log.Logger, db *sqlx.DB, f func(ctx context.Context, tx *sqlx.Tx) error) (err error) {
	tx, err := db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelDefault,
		ReadOnly:  false,
	})
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			// Rollback on panic
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.With(logger, "err", rollbackErr).Log("msg", "Postgres tx rollback failed")
			}
			panic(r)
		} else if err != nil {
			// Rollback if the function failed
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				log.With(logger, "err", rollbackErr).Log("msg", "Postgres tx rollback failed")
			}
			logger.Log("err", err)
		} else {
			// Commit, and return any error from that
			err = tx.Commit()
		}
	}()

	return f(ctx, tx)
}
