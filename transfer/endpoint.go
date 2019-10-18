package transfer

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-kit/kit/endpoint"
	uuid "github.com/satori/go.uuid"

	wallet "github.com/xsleonard/gokit-example"
	"github.com/xsleonard/gokit-example/decimal"
)

// Payment is a JSON-representable form of wallet.Payment
type Payment struct {
	ID     string `json:"id"`
	To     string `json:"to"`
	From   string `json:"from,omitempty"`
	Amount string `json:"amount"`
}

func newPayment(p wallet.Payment) Payment {
	return Payment{
		ID:     p.ID.String(),
		To:     p.To.String(),
		From:   p.FromUUIDString(),
		Amount: p.Amount.Text('f'),
	}
}

func newPayments(payments []wallet.Payment) []Payment {
	if len(payments) == 0 {
		return nil
	}

	out := make([]Payment, len(payments))
	for i, p := range payments {
		out[i] = newPayment(p)
	}
	return out
}

// Account is a JSON-representable form of wallet.Account
type Account struct {
	ID       string `json:"id"`
	Currency string `json:"currency"`
	Balance  string `json:"balance"`
}

func newAccount(a wallet.Account) Account {
	return Account{
		ID:       a.ID.String(),
		Currency: a.Currency,
		Balance:  a.Balance.Text('f'),
	}
}

func newAccounts(accounts []wallet.Account) []Account {
	if len(accounts) == 0 {
		return nil
	}

	out := make([]Account, len(accounts))
	for i, a := range accounts {
		out[i] = newAccount(a)
	}
	return out
}

type transferRequest struct {
	To     string `json:"to"`
	From   string `json:"from"`
	Amount string `json:"amount"`
}

type transferResponse struct {
	Payment *Payment `json:"payment,omitempty"`
	Err     error    `json:"error,omitempty"`
}

func (r transferResponse) error() error {
	return r.Err
}

var (
	errFromRequired   = errors.New("from is required")
	errToRequired     = errors.New("to is required")
	errAmountRequired = errors.New("amount is required")
)

type errInvalidAccountID struct {
	Err   error
	Field string
}

func (e errInvalidAccountID) Error() string {
	return fmt.Sprintf("Invalid account ID for field %q: %v", e.Field, e.Err)
}

func makeTransferEndpoint(s wallet.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(transferRequest)

		// Parse and validate request fields
		// Note: we could use the parsed types (uuid.UUID, apd.Decimal)
		// in the request struct, but then we would lose control over the
		// response error handling
		if req.From == "" {
			return nil, errFromRequired
		}
		if req.To == "" {
			return nil, errToRequired
		}
		if req.Amount == "" {
			return nil, errAmountRequired
		}

		from, err := uuid.FromString(req.From)
		if err != nil {
			return nil, errInvalidAccountID{
				Err:   err,
				Field: "from",
			}
		}

		to, err := uuid.FromString(req.To)
		if err != nil {
			return nil, errInvalidAccountID{
				Err:   err,
				Field: "to",
			}
		}

		amount, err := decimal.ParseCurrency(req.Amount)
		if err != nil {
			return nil, err
		}

		p, err := s.Transfer(ctx, to, from, amount)
		if err != nil {
			return transferResponse{
				Err: err,
			}, nil
		}

		pp := newPayment(*p)
		return transferResponse{
			Payment: &pp,
		}, nil
	}
}

type paymentsResponse struct {
	Payments []Payment `json:"payments,omitempty"`
	Err      error     `json:"error,omitempty"`
}

func (r paymentsResponse) error() error {
	return r.Err
}

func makePaymentsEndpoint(s wallet.Service) endpoint.Endpoint {
	return func(ctx context.Context, _ interface{}) (interface{}, error) {
		p, err := s.Payments(ctx)
		return paymentsResponse{
			Payments: newPayments(p),
			Err:      err,
		}, nil
	}
}

type accountsResponse struct {
	Accounts []Account `json:"accounts,omitempty"`
	Err      error     `json:"error,omitempty"`
}

func (r accountsResponse) error() error {
	return r.Err
}

func makeAccountsEndpoint(s wallet.Service) endpoint.Endpoint {
	return func(ctx context.Context, _ interface{}) (interface{}, error) {
		a, err := s.Accounts(ctx)
		return accountsResponse{
			Accounts: newAccounts(a),
			Err:      err,
		}, nil
	}
}
