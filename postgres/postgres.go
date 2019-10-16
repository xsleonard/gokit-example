// Package postgres implements storage using postgres
package postgres

import (
	"context"
	"database/sql"

	"github.com/go-kit/kit/log"
	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"

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
	return insideTx(ctx, r.logger, r.db, func(ctx context.Context, tx *sqlx.Tx) error {
		q := `insert into account (id, currency, balance) values (?, ?, ?)`
		// TODO -- let postgres assign the ID?
		_, err := tx.ExecContext(ctx, q, account.ID, account.Currency, account.Balance.StringFixed(2))
		return err
	})
}

func (r *accountRepository) Get(ctx context.Context, id string) (*wallet.Account, error) {
	var balance string
	var currency string
	if err := insideTx(ctx, r.logger, r.db, func(ctx context.Context, tx *sqlx.Tx) error {
		row := tx.QueryRowContext(ctx, `select balance, currency from account where id=$1`, id)
		return row.Scan(&balance, &currency)
	}); err != nil {
		return nil, err
	}

	return wallet.NewAccount(id, balance, currency)
}

func (r *accountRepository) All(ctx context.Context) ([]wallet.Account, error) {
	// TODO
	return nil, nil
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
	return nil
}

func (r *paymentRepository) All(ctx context.Context) ([]wallet.Payment, error) {
	return nil, nil
}

func (r *paymentRepository) Balance(ctx context.Context, accountID string) (decimal.Decimal, error) {
	return decimal.New(0, 0), nil
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
