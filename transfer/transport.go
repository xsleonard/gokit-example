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

type decodeError struct {
	error
}

func decodeTransferRequest(_ context.Context, r *http.Request) (interface{}, error) {
	if r.Method != http.MethodPost {
		return nil, errMethodNotAllowed
	}

	var req transferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, decodeError{err}
	}

	return req, nil
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
	switch err.(type) {
	case decodeError, errInvalidAccountID:
		w.WriteHeader(http.StatusBadRequest)
	default:
		switch err {
		case errMethodNotAllowed:
			w.WriteHeader(http.StatusMethodNotAllowed)
		case wallet.ErrNoAccount:
			w.WriteHeader(http.StatusNotFound)
		case decimal.ErrInvalidPrecision,
			decimal.ErrNegative,
			decimal.ErrInvalid,
			decimal.ErrNotFinite,
			decimal.ErrAmountNotMoreThanZero,
			decimal.ErrAmountNil,
			errInsufficientBalance,
			errDifferentCurrency,
			errSameAccount,
			errToRequired,
			errFromRequired,
			errAmountRequired:
			w.WriteHeader(http.StatusBadRequest)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
