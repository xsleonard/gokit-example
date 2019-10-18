// Package postgres implements storage using postgres
package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/cockroachdb/apd"
	"github.com/go-kit/kit/log"
	"github.com/jmoiron/sqlx"
	uuid "github.com/satori/go.uuid"

	wallet "github.com/xsleonard/gokit-example"
)

var (
	// errEmptyAccountID is returned when creating an account without an ID
	errEmptyAccountID = errors.New("Account ID must not be empty")
	// errEmptyPaymentID is returned when creating an account without an ID
	errEmptyPaymentID = errors.New("Payment ID must not be empty")
	// errInvalidCurrency is returned for unrecognized currency codes
	errInvalidCurrency = errors.New("Invalid currency code")

	nullUUID uuid.UUID
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
	if uuid.Equal(account.ID, nullUUID) {
		return errEmptyAccountID
	}
	if !wallet.IsValidCurrency(account.Currency) {
		return errInvalidCurrency
	}

	return withTx(ctx, r.logger, r.db, func(ctx context.Context, tx *sqlx.Tx) error {
		q := `insert into account (id, currency) values ($1, $2)`
		_, err := tx.ExecContext(ctx, q, account.ID, account.Currency)
		return err
	})
}

type account struct {
	ID       uuid.UUID    `db:"id"`
	Balance  *apd.Decimal `db:"balance"`
	Currency string       `db:"currency"`
}

func newWalletAccount(a account) wallet.Account {
	return wallet.Account{
		ID:       a.ID,
		Balance:  a.Balance,
		Currency: a.Currency,
	}
}

func newWalletAccounts(accounts []account) []wallet.Account {
	aa := make([]wallet.Account, len(accounts))
	for i, a := range accounts {
		aa[i] = newWalletAccount(a)
	}
	return aa
}

func (r *accountRepository) GetTx(ctx context.Context, tx *sqlx.Tx, id uuid.UUID) (*wallet.Account, error) {
	row := tx.QueryRowxContext(ctx, `select id, balance, currency from account_balance where id=$1`, id)

	var a account
	if err := row.StructScan(&a); err != nil {
		if err == sql.ErrNoRows {
			return nil, wallet.ErrNoAccount
		}
		return nil, err
	}

	wa := newWalletAccount(a)
	return &wa, nil
}

func (r *accountRepository) All(ctx context.Context) ([]wallet.Account, error) {
	rows, err := r.db.QueryxContext(ctx, `select id, balance, currency from account_balance`)
	if err != nil {
		return nil, err
	}

	var accounts []wallet.Account
	defer rows.Close()
	for rows.Next() {
		var a account
		if err := rows.StructScan(&a); err != nil {
			return nil, err
		}
		accounts = append(accounts, newWalletAccount(a))
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

func (r *paymentRepository) WithTx(ctx context.Context, f func(ctx context.Context, tx *sqlx.Tx) error) error {
	return withTx(ctx, r.logger, r.db, f)
}

func (r *paymentRepository) Store(ctx context.Context, p *wallet.Payment) error {
	return withTx(ctx, r.logger, r.db, func(ctx context.Context, tx *sqlx.Tx) error {
		return r.StoreTx(ctx, tx, p)
	})
}

func (r *paymentRepository) StoreTx(ctx context.Context, tx *sqlx.Tx, p *wallet.Payment) error {
	if p.ID == uuid.Nil {
		return errEmptyPaymentID
	}

	q := `insert into payment (id, from_account_id, to_account_id, amount) values ($1, $2, $3, $4)`
	_, err := tx.ExecContext(ctx, q, p.ID, p.From, p.To, p.Amount)
	return err
}

type payment struct {
	ID     uuid.UUID     `db:"id"`
	To     uuid.UUID     `db:"to_account_id"`
	From   uuid.NullUUID `db:"from_account_id"`
	Amount *apd.Decimal  `db:"amount"`
}

func newWalletPayment(p payment) wallet.Payment {
	if p.Amount == nil {
		panic("amount is unexpectedly nil")
	}
	pp := wallet.Payment{
		ID:     p.ID,
		To:     p.To,
		Amount: p.Amount,
	}
	if p.From.Valid {
		fromID := p.From.UUID
		pp.From = &fromID
	}
	return pp
}

func (r *paymentRepository) All(ctx context.Context) ([]wallet.Payment, error) {
	rows, err := r.db.QueryxContext(ctx, `select id, from_account_id, to_account_id, amount from payment`)
	if err != nil {
		return nil, err
	}

	var payments []wallet.Payment
	defer rows.Close()
	for rows.Next() {
		var p payment
		if err := rows.StructScan(&p); err != nil {
			return nil, err
		}
		payments = append(payments, newWalletPayment(p))
	}

	return payments, nil
}

func withTx(ctx context.Context, logger log.Logger, db *sqlx.DB, f func(ctx context.Context, tx *sqlx.Tx) error) (err error) {
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
