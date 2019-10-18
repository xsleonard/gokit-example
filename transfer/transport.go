package transfer

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/transport"
	kithttp "github.com/go-kit/kit/transport/http"

	wallet "github.com/xsleonard/gokit-example"
	"github.com/xsleonard/gokit-example/decimal"
)

var errMethodNotAllowed = errors.New(http.StatusText(http.StatusMethodNotAllowed))

// TODO -- setup cors

// MakeHandler returns a handler for the tracking service.
func MakeHandler(s wallet.Service, logger log.Logger) http.Handler {
	r := http.NewServeMux()

	opts := []kithttp.ServerOption{
		kithttp.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		kithttp.ServerErrorEncoder(encodeError),
	}

	transferHandler := kithttp.NewServer(
		makeTransferEndpoint(s),
		decodeTransferRequest,
		encodeResponse,
		opts...,
	)

	paymentsHandler := kithttp.NewServer(
		makePaymentsEndpoint(s),
		decodeEmptyRequest([]string{http.MethodGet}),
		encodeResponse,
		opts...,
	)

	accountsHandler := kithttp.NewServer(
		makeAccountsEndpoint(s),
		decodeEmptyRequest([]string{http.MethodGet}),
		encodeResponse,
		opts...,
	)

	r.Handle("/v1/transfer", transferHandler)
	r.Handle("/v1/payments", paymentsHandler)
	r.Handle("/v1/accounts", accountsHandler)

	return r
}

func decodeEmptyRequest(allowedMethods []string) kithttp.DecodeRequestFunc {
	return func(_ context.Context, r *http.Request) (interface{}, error) {
		allowed := false
		for _, m := range allowedMethods {
			if r.Method == m {
				allowed = true
				break
			}
		}
		if !allowed {
			return nil, errMethodNotAllowed
		}

		return struct{}{}, nil
	}
}

func decodeTransferRequest(_ context.Context, r *http.Request) (interface{}, error) {
	if r.Method != http.MethodPost {
		return nil, errMethodNotAllowed
	}

	// TODO -- accept json, not form values

	from := r.FormValue("from")
	if from == "" {
		return nil, errors.New("from is required")
	}

	to := r.FormValue("to")
	if to == "" {
		return nil, errors.New("to is required")
	}

	amount := r.FormValue("amount")
	if amount == "" {
		return nil, errors.New("amount is required")
	}

	decimalAmount, err := decimal.ParseCurrency(amount)
	if err != nil {
		return nil, err
	}

	return transferRequest{
		From:   from,
		To:     to,
		Amount: decimalAmount,
	}, nil
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		encodeError(ctx, e.error(), w)
		return nil
	}
	// Note: charset=utf-8 mitigates some old browser vulnerabilities
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

type errorer interface {
	error() error
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	// Note: charset=utf-8 mitigates some old browser vulnerabilities
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch err {
	case errMethodNotAllowed:
		w.WriteHeader(http.StatusMethodNotAllowed)
	case wallet.ErrNoAccount:
		w.WriteHeader(http.StatusNotFound)
	case decimal.ErrInvalidPrecision,
		decimal.ErrNegative,
		decimal.ErrInvalid,
		decimal.ErrNotFinite,
		wallet.ErrEmptyAccountID,
		wallet.ErrEmptyPaymentID,
		wallet.ErrInvalidCurrency,
		ErrInsufficientBalance,
		ErrDifferentCurrency,
		ErrSameAccount:
		w.WriteHeader(http.StatusBadRequest)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
