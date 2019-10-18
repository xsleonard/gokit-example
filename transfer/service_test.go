package transfer

import (
	"context"
	// "errors"
	"fmt"
	"testing"

	"github.com/cockroachdb/apd"
	"github.com/go-kit/kit/log"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"

	wallet "github.com/xsleonard/gokit-example"
	"github.com/xsleonard/gokit-example/decimal"
	"github.com/xsleonard/gokit-example/postgres"

	_ "github.com/lib/pq" // load postgres driver
)

var (
	databaseName = "wallet_test"
	databaseURL  = fmt.Sprintf("postgresql://postgres@localhost:54320/%s?sslmode=disable", databaseName)
)

// setupDB sets up a clean test database and returns a teardown function
func setupDB(t *testing.T) (*sqlx.DB, func()) {
	m, err := migrate.New("file://../migrations", databaseURL)
	require.NoError(t, err)

	err = m.Down()
	if err != migrate.ErrNoChange {
		require.NoError(t, err)
	}

	err = m.Up()
	require.NoError(t, err)

	db, err := sqlx.Connect("postgres", databaseURL)
	require.NoError(t, err)

	return db, func() {
		err := m.Down()
		require.NoError(t, err)

		err = db.Close()
		require.NoError(t, err)
	}
}

func TestServiceTransfer(t *testing.T) {
	toID := uuid.Must(uuid.NewV4())
	fromID := uuid.Must(uuid.NewV4())

	cases := []struct {
		name   string
		to     uuid.UUID
		from   uuid.UUID
		amount *apd.Decimal
		setup  func(*testing.T, context.Context, service)
		err    error
	}{
		{
			name:   "nil amount",
			to:     toID,
			from:   fromID,
			amount: nil,
			err:    decimal.ErrAmountNil,
		},

		{
			name:   "negative amount",
			to:     toID,
			from:   fromID,
			amount: apd.New(-123, -2),
			err:    decimal.ErrAmountNotMoreThanZero,
		},

		{
			name:   "to and from are the same id",
			to:     toID,
			from:   toID,
			amount: apd.New(123, -2),
			err:    errSameAccount,
		},

		{
			name:   "To account does not exist",
			to:     toID,
			from:   fromID,
			amount: apd.New(123, -2),
			err:    wallet.ErrNoAccount,
		},

		{
			name:   "From account does not exist",
			to:     toID,
			from:   fromID,
			amount: apd.New(123, -2),
			err:    wallet.ErrNoAccount,
			setup: func(t *testing.T, ctx context.Context, s service) {
				err := s.accounts.Store(ctx, &wallet.Account{
					ID:       toID,
					Currency: wallet.USD,
				})
				require.NoError(t, err)
			},
		},

		{
			name:   "account currencies do not match",
			to:     toID,
			from:   fromID,
			amount: apd.New(123, -2),
			err:    errDifferentCurrency,
			setup: func(t *testing.T, ctx context.Context, s service) {
				err := s.accounts.Store(ctx, &wallet.Account{
					ID:       toID,
					Currency: wallet.USD,
				})
				require.NoError(t, err)

				err = s.accounts.Store(ctx, &wallet.Account{
					ID:       fromID,
					Currency: wallet.SGD,
				})
				require.NoError(t, err)
			},
		},

		{
			name:   "insufficient balance",
			to:     toID,
			from:   fromID,
			amount: apd.New(9999999, -2),
			err:    errInsufficientBalance,
			setup: func(t *testing.T, ctx context.Context, s service) {
				err := s.accounts.Store(ctx, &wallet.Account{
					ID:       toID,
					Currency: wallet.USD,
				})
				require.NoError(t, err)

				err = s.accounts.Store(ctx, &wallet.Account{
					ID:       fromID,
					Currency: wallet.USD,
				})
				require.NoError(t, err)

				err = s.payments.Store(ctx, &wallet.Payment{
					ID:     uuid.Must(uuid.NewV4()),
					To:     fromID,
					From:   nil,
					Amount: apd.New(100, 0),
				})
				require.NoError(t, err)
			},
		},

		{
			name:   "valid, full balance",
			to:     toID,
			from:   fromID,
			amount: apd.New(100, 0),
			setup: func(t *testing.T, ctx context.Context, s service) {
				err := s.accounts.Store(ctx, &wallet.Account{
					ID:       toID,
					Currency: wallet.USD,
				})
				require.NoError(t, err)

				err = s.accounts.Store(ctx, &wallet.Account{
					ID:       fromID,
					Currency: wallet.USD,
				})
				require.NoError(t, err)

				err = s.payments.Store(ctx, &wallet.Payment{
					ID:     uuid.Must(uuid.NewV4()),
					To:     fromID,
					From:   nil,
					Amount: apd.New(100, 0),
				})
				require.NoError(t, err)
			},
		},

		{
			name:   "valid, partial balance",
			to:     toID,
			from:   fromID,
			amount: apd.New(1, -2),
			setup: func(t *testing.T, ctx context.Context, s service) {
				err := s.accounts.Store(ctx, &wallet.Account{
					ID:       toID,
					Currency: wallet.USD,
				})
				require.NoError(t, err)

				err = s.accounts.Store(ctx, &wallet.Account{
					ID:       fromID,
					Currency: wallet.USD,
				})
				require.NoError(t, err)

				err = s.payments.Store(ctx, &wallet.Payment{
					ID:     uuid.Must(uuid.NewV4()),
					To:     fromID,
					From:   nil,
					Amount: apd.New(100, 0),
				})
				require.NoError(t, err)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db, shutdown := setupDB(t)
			defer shutdown()

			logger := log.NewNopLogger()
			accountsRepo := postgres.NewAccountRepository(db, logger)
			paymentsRepo := postgres.NewPaymentRepository(db, logger)

			s := NewService(accountsRepo, paymentsRepo)

			ctx := context.Background()

			if tc.setup != nil {
				tc.setup(t, ctx, s.(service))
			}

			p, err := s.Transfer(ctx, tc.to, tc.from, tc.amount)
			if tc.err != nil {
				require.Error(t, err)
				require.Equal(t, tc.err, err, "%v != %v", tc.err, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, p)
			require.NotEqual(t, uuid.Nil, p.ID)
			require.True(t, uuid.Equal(tc.to, p.To))
			require.True(t, uuid.Equal(tc.from, *p.From))
			require.Equal(t, 0, tc.amount.Cmp(p.Amount))
		})
	}
}
