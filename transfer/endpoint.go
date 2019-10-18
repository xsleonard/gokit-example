package transfer

import (
	"context"

	"github.com/cockroachdb/apd"
	"github.com/go-kit/kit/endpoint"

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
		Amount: decimal.FormatCurrency(p.Amount),
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
		Balance:  decimal.FormatCurrency(a.Balance),
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
	To     string       `json:"to"`
	From   string       `json:"from"`
	Amount *apd.Decimal `json:"amount"`
}

type transferResponse struct {
	Payment *Payment `json:"payment,omitempty"`
	Err     error    `json:"error,omitempty"`
}

func (r transferResponse) error() error {
	return r.Err
}

func makeTransferEndpoint(s wallet.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(transferRequest)

		p, err := s.Transfer(ctx, req.To, req.From, req.Amount)
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
