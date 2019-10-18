package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/jmoiron/sqlx"

	"github.com/xsleonard/gokit-example/postgres"
	"github.com/xsleonard/gokit-example/transfer"

	_ "github.com/lib/pq" // load postgres driver
)

const (
	// serverShutdownTimeout is the timeout the http server's Shutdown() call
	serverShutdownTimeout = time.Second * 5

	// https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/
	// The timeout configuration is necessary for public servers, or else
	// connections will be used up
	serverReadTimeout  = time.Second * 10
	serverWriteTimeout = time.Second * 60
	serverIdleTimeout  = time.Second * 120

	defaultDatabaseURL = "postgresql://postgres@localhost:54320/wallet?sslmode=disable"
)

func main() {
	// TODO -- use viper to parse flags?
	// TODO -- use env vars as fallback
	var httpAddr string
	var databaseURL string
	flag.StringVar(&httpAddr, "addr", "localhost:8888", "HTTP listen address")
	flag.StringVar(&databaseURL, "db", defaultDatabaseURL, "Postgres DB URL")
	flag.Parse()

	ctx := context.Background()

	// Setup logger
	logger := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)

	// Setup DB
	db, err := sqlx.ConnectContext(ctx, "postgres", databaseURL)
	if err != nil {
		log.With(logger, "err", err).Log("Unable to connect to DB")
		os.Exit(1)
	}
	defer db.Close()

	accountStorage := postgres.NewAccountRepository(db, log.With(logger, "pkg", "postgres"))
	paymentStorage := postgres.NewPaymentRepository(db, log.With(logger, "pkg", "postgres"))

	transferLogger := log.With(logger, "pkg", "transfer")
	service := transfer.NewService(accountStorage, paymentStorage)
	service = transfer.NewLoggingService(transferLogger, service)

	// Setup HTTP server
	handler := transfer.MakeHandler(service, log.With(transferLogger, "transport", "http"))

	httpServer := &http.Server{
		Addr:         httpAddr,
		Handler:      handler,
		ReadTimeout:  serverReadTimeout,
		WriteTimeout: serverWriteTimeout,
		IdleTimeout:  serverIdleTimeout,
	}

	errs := make(chan error, 2)
	go func() {
		logger.Log("transport", "http", "address", httpAddr, "msg", "listening")
		err := httpServer.ListenAndServe()
		if err != http.ErrServerClosed {
			errs <- err
		}
	}()

	// Handle Ctrl+C for shutdown
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		<-c

		// TODO -- handle a 2nd ctrl+c to dump stack and panic

		ctx, cancel := context.WithTimeout(ctx, serverShutdownTimeout)
		defer cancel()
		errs <- httpServer.Shutdown(ctx)
	}()

	logger.Log("terminated", <-errs)
}
