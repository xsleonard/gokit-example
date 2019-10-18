package main

import (
	"context"
	"errors"
	"flag"
	"os"

	"github.com/cockroachdb/apd"
	"github.com/go-kit/kit/log"
	"github.com/jmoiron/sqlx"
	uuid "github.com/satori/go.uuid"

	wallet "github.com/xsleonard/gokit-example"
	"github.com/xsleonard/gokit-example/postgres"

	_ "github.com/lib/pq" // load postgres driver
)

const (
	defaultDatabaseURL = "postgresql://postgres@localhost:54320/wallet?sslmode=disable"
)

func main() {
	// TODO -- use viper to parse flags?
	// TODO -- use env vars as fallback
	var databaseURL string
	flag.StringVar(&databaseURL, "db", defaultDatabaseURL, "Postgres DB URL")
	flag.Parse()

	ctx := context.Background()

	// Setup logger
	var logger log.Logger
	logger = log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)

	db, err := sqlx.ConnectContext(ctx, "postgres", databaseURL)
	if err != nil {
		log.With(logger, "err", err).Log("Unable to connect to DB")
		os.Exit(1)
	}
	defer db.Close()

	accountStorage := postgres.NewAccountRepository(db, log.With(logger, "pkg", "postgres"))
	paymentStorage := postgres.NewPaymentRepository(db, log.With(logger, "pkg", "postgres"))

	accounts := makeTestAccounts(logger)
	for _, a := range accounts {
		err := accountStorage.Store(ctx, &a)
		exitOnErr(logger, err)
	}

	payments := makeTestPayments(logger, accounts)
	for _, p := range payments {
		err := paymentStorage.Store(ctx, &p)
		exitOnErr(logger, err)
	}
}

func makeUUIDs(logger log.Logger, uuidStrings []string) []uuid.UUID {
	uuids := make([]uuid.UUID, len(uuidStrings))
	for i := range uuidStrings {
		id, err := uuid.FromString(uuidStrings)
		exitOnErr(logger, err)
		uuids[i] = id
	}
	return uuids
}

func makeTestAccounts(logger log.Logger) []wallet.Account {
	uuids := makeUUIDs(logger, []string{
		"d3f05a8d-1708-47de-8e1c-304e7fb5a93f",
		"5e0281df-cb1e-4b2f-bf61-0286295d07c9",
		"46e0b1dd-5cb2-4b40-b4d9-06b5e3d51059",
		"92820a1f-4249-44fd-a152-b956fb001274",
		"a88d1536-73c0-4aef-bf1c-a89e355a00fe",
		"ab5977f7-cb1a-4619-b76c-25a437d07ea7",
	})

	accounts := []wallet.Account{
		{
			ID:       uuids[0],
			Currency: wallet.USD,
		},
		{
			ID:       uuids[1],
			Currency: wallet.EUR,
		},
		{
			ID:       uuids[2],
			Currency: wallet.SGD,
		},
		{
			ID:       uuids[3],
			Currency: wallet.USD,
		},
		{
			ID:       uuids[4],
			Currency: wallet.EUR,
		},
		{
			ID:       uuids[5],
			Currency: wallet.SGD,
		},
	}

	return accounts
}

func makeTestPayments(logger log.Logger, accounts []wallet.Account) []wallet.Payment {
	uuids := makeUUIDs(logger, []string{
		"8c7ecafb-df60-400a-a985-8f260c2fbb2a",
		"18da7d72-c33a-410b-ae6a-c3bd027082fd",
		"8d84d67e-2cf6-43fa-a2a1-e2e121db8ee3",
		"ef1ac34e-e7e7-4946-9ae4-fa1da6dccca7",
		"0ed53dc7-946b-45c4-a717-9946aab1ac3f",
		"38f4b350-c848-400d-bc91-a112fb4f58df",
	})

	// Credits to account (no "From" field)
	payments := make([]wallet.Payment, len(accounts))
	for i, a := range accounts {
		amount, cond, err := apd.NewFromString("100.00")
		exitOnErr(logger, err)
		if cond.Any() {
			exitOnErr(logger, errors.New("default amount has a condition"))
		}

		payments[i] = wallet.Payment{
			ID:     uuids[i],
			To:     a.ID,
			From:   nil,
			Amount: amount,
		}
	}

	return payments
}

func exitOnErr(logger log.Logger, err error) {
	if err != nil {
		logger.Log("err", err)
		os.Exit(1)
	}
}
