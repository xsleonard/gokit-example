package transfer

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/shopspring/decimal"

	wallet "github.com/xsleonard/gokit-example"
)

// Payment is a JSON-representable form of wallet.Payment
type Payment struct {
	To     string `json:"to"`
	From   string `json:"from"`
	Amount string `json:"amount"`
}

func newPayment(p wallet.Payment) Payment {
	return Payment{
		To:     p.To,
		From:   p.From,
		Amount: p.Amount.StringFixed(2),
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
		ID:       a.ID,
		Currency: a.Currency,
		Balance:  a.Balance.StringFixed(2),
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
	To     string          `json:"to"`
	From   string          `json:"from"`
	Amount decimal.Decimal `json:"amount"`
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
