package transfer

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cockroachdb/apd"
	"github.com/go-kit/kit/log"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"

	wallet "github.com/xsleonard/gokit-example"
	"github.com/xsleonard/gokit-example/postgres"
)

func TestEndpoints(t *testing.T) {
	toID := uuid.Must(uuid.FromString("b0505aa0-b927-4667-a484-906b4e2a410b"))
	fromID := uuid.Must(uuid.FromString("5136843a-0948-432d-8ce6-060362edb538"))

	paymentIDs := make([]uuid.UUID, 3)
	paymentIDs[0] = uuid.Must(uuid.FromString("7e09ef65-1203-4d10-849a-9e56b9368166"))
	paymentIDs[1] = uuid.Must(uuid.FromString("1024abad-6de0-466f-9022-4499a97c3f87"))
	paymentIDs[2] = uuid.Must(uuid.FromString("e76c0e9d-499f-4759-ae40-895fec818035"))

	cases := []struct {
		name          string
		url           string
		method        string
		body          string
		statusCode    int
		checkResponse func(*testing.T, string)
		response      string
		setup         func(*testing.T, context.Context, service)
		checkDB       func(*testing.T, context.Context, service)
	}{
		{
			name:       "list all accounts, empty",
			url:        "/v1/accounts",
			method:     http.MethodGet,
			statusCode: http.StatusOK,
			response:   "{}",
		},

		{
			name:       "list all accounts with balances",
			url:        "/v1/accounts",
			method:     http.MethodGet,
			statusCode: http.StatusOK,
			response:   `{"accounts":[{"id":"5136843a-0948-432d-8ce6-060362edb538","currency":"USD","balance":"69.67"},{"id":"b0505aa0-b927-4667-a484-906b4e2a410b","currency":"USD","balance":"30.33"}]}`,
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
					ID:     paymentIDs[0],
					To:     fromID,
					From:   nil,
					Amount: apd.New(100, 0),
				})
				require.NoError(t, err)

				err = s.payments.Store(ctx, &wallet.Payment{
					ID:     paymentIDs[1],
					To:     toID,
					From:   &fromID,
					Amount: apd.New(3033, -2),
				})
				require.NoError(t, err)
			},
		},

		{
			name:       "list all accounts, bad method",
			url:        "/v1/accounts",
			method:     http.MethodPost,
			statusCode: http.StatusMethodNotAllowed,
			response:   `{"error":"Method Not Allowed"}`,
		},

		{
			name:       "list all payments, empty",
			url:        "/v1/payments",
			method:     http.MethodGet,
			statusCode: http.StatusOK,
			response:   "{}",
		},

		{
			name:       "list all payments",
			url:        "/v1/payments",
			method:     http.MethodGet,
			statusCode: http.StatusOK,
			response:   `{"payments":[{"id":"1024abad-6de0-466f-9022-4499a97c3f87","to":"b0505aa0-b927-4667-a484-906b4e2a410b","from":"5136843a-0948-432d-8ce6-060362edb538","amount":"30.33"},{"id":"7e09ef65-1203-4d10-849a-9e56b9368166","to":"5136843a-0948-432d-8ce6-060362edb538","amount":"100.00"},{"id":"e76c0e9d-499f-4759-ae40-895fec818035","to":"5136843a-0948-432d-8ce6-060362edb538","from":"b0505aa0-b927-4667-a484-906b4e2a410b","amount":"10.11"}]}`,
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
					ID:     paymentIDs[0],
					To:     fromID,
					From:   nil,
					Amount: apd.New(100, 0),
				})
				require.NoError(t, err)

				err = s.payments.Store(ctx, &wallet.Payment{
					ID:     paymentIDs[1],
					To:     toID,
					From:   &fromID,
					Amount: apd.New(3033, -2),
				})
				require.NoError(t, err)

				err = s.payments.Store(ctx, &wallet.Payment{
					ID:     paymentIDs[2],
					To:     fromID,
					From:   &toID,
					Amount: apd.New(1011, -2),
				})
				require.NoError(t, err)
			},
		},

		{
			name:       "list all payments, bad method",
			url:        "/v1/accounts",
			method:     http.MethodPost,
			statusCode: http.StatusMethodNotAllowed,
			response:   `{"error":"Method Not Allowed"}`,
		},

		{
			name:       "transfer, bad method",
			url:        "/v1/transfer",
			method:     http.MethodGet,
			statusCode: http.StatusMethodNotAllowed,
			response:   `{"error":"Method Not Allowed"}`,
		},

		{
			name:       "transfer, bad json body",
			url:        "/v1/transfer",
			method:     http.MethodPost,
			body:       `{5`,
			statusCode: http.StatusBadRequest,
			response:   `{"error":"invalid character '5' looking for beginning of object key string"}`,
		},

		{
			name:       "transfer, invalid amount too many decimals",
			url:        "/v1/transfer",
			method:     http.MethodPost,
			body:       fmt.Sprintf(`{"to":%q,"from":%q,"amount":"123.456"}`, toID, fromID),
			statusCode: http.StatusBadRequest,
			response:   `{"error":"Amount strings must have exactly two digits of precision"}`,
		},

		{
			name:       "transfer, invalid amount negative",
			url:        "/v1/transfer",
			method:     http.MethodPost,
			body:       fmt.Sprintf(`{"to":%q,"from":%q,"amount":"-123.45"}`, toID, fromID),
			statusCode: http.StatusBadRequest,
			response:   `{"error":"Amount must not be negative"}`,
		},

		{
			name:       "transfer, missing amount",
			url:        "/v1/transfer",
			method:     http.MethodPost,
			body:       fmt.Sprintf(`{"to":%q,"from":%q}`, toID, fromID),
			statusCode: http.StatusBadRequest,
			response:   `{"error":"amount is required"}`,
		},

		{
			name:       "transfer, missing to",
			url:        "/v1/transfer",
			method:     http.MethodPost,
			body:       fmt.Sprintf(`{"from":%q,"amount":"1.23"}`, fromID),
			statusCode: http.StatusBadRequest,
			response:   `{"error":"to is required"}`,
		},

		{
			name:       "transfer, missing from",
			url:        "/v1/transfer",
			method:     http.MethodPost,
			body:       fmt.Sprintf(`{"to":%q,"amount":"1.23"}`, toID),
			statusCode: http.StatusBadRequest,
			response:   `{"error":"from is required"}`,
		},

		{
			name:       "transfer, invalid to",
			url:        "/v1/transfer",
			method:     http.MethodPost,
			body:       fmt.Sprintf(`{"to":"abc","from":%q,"amount":"1.23"}`, fromID),
			statusCode: http.StatusBadRequest,
			response:   `{"error":"Invalid account ID for field \"to\": uuid: incorrect UUID length: abc"}`,
		},

		{
			name:       "transfer, invalid from",
			url:        "/v1/transfer",
			method:     http.MethodPost,
			body:       fmt.Sprintf(`{"to":%q,"from":"abc","amount":"1.23"}`, toID),
			statusCode: http.StatusBadRequest,
			response:   `{"error":"Invalid account ID for field \"from\": uuid: incorrect UUID length: abc"}`,
		},

		{
			name:       "transfer, to does not exist",
			url:        "/v1/transfer",
			method:     http.MethodPost,
			body:       fmt.Sprintf(`{"to":%q,"from":%q,"amount":"1.23"}`, toID, fromID),
			statusCode: http.StatusNotFound,
			response:   `{"error":"Account does not exist"}`,
		},

		{
			name:       "transfer, from does not exist",
			url:        "/v1/transfer",
			method:     http.MethodPost,
			body:       fmt.Sprintf(`{"to":%q,"from":%q,"amount":"1.23"}`, toID, fromID),
			statusCode: http.StatusNotFound,
			response:   `{"error":"Account does not exist"}`,
		},

		{
			name:       "transfer, different currencies",
			url:        "/v1/transfer",
			method:     http.MethodPost,
			body:       fmt.Sprintf(`{"to":%q,"from":%q,"amount":"1.23"}`, toID, fromID),
			statusCode: http.StatusBadRequest,
			response:   `{"error":"Transfers must use the same currency"}`,
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

				err = s.payments.Store(ctx, &wallet.Payment{
					ID:     paymentIDs[0],
					To:     fromID,
					From:   nil,
					Amount: apd.New(100, 0),
				})
				require.NoError(t, err)
			},
		},

		{
			name:       "transfer, valid",
			url:        "/v1/transfer",
			method:     http.MethodPost,
			body:       fmt.Sprintf(`{"to":%q,"from":%q,"amount":"1.23"}`, toID, fromID),
			statusCode: http.StatusOK,
			checkResponse: func(t *testing.T, resp string) {
				// The payment ID will be random each time, so a custom compare is used
				expTo := toID.String()
				expFrom := fromID.String()
				expAmount := "1.23"

				var r transferResponse
				err := json.Unmarshal([]byte(resp), &r)
				require.NoError(t, err)

				require.NoError(t, r.Err)
				require.NotNil(t, r.Payment)
				require.Equal(t, expTo, r.Payment.To)
				require.Equal(t, expFrom, r.Payment.From)
				require.Equal(t, expAmount, r.Payment.Amount)

				require.NotEmpty(t, r.Payment.ID)
				_, err = uuid.FromString(r.Payment.ID)
				require.NoError(t, err)
			},
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
					ID:     paymentIDs[0],
					To:     fromID,
					From:   nil,
					Amount: apd.New(100, 0),
				})
				require.NoError(t, err)
			},
			checkDB: func(t *testing.T, ctx context.Context, s service) {
				// Check that a payment was added to the DB
				payments, err := s.payments.All(ctx)
				require.NoError(t, err)
				require.Equal(t, 2, len(payments))

				// Check the new balances of accounts
				accounts, err := s.accounts.All(ctx)
				require.NoError(t, err)

				for _, a := range accounts {
					if a.ID == toID {
						require.Equal(t, apd.New(123, -2), a.Balance)
					} else if a.ID == fromID {
						require.Equal(t, apd.New(9877, -2), a.Balance)
					} else {
						t.Fatal("found an unexpected account")
					}
				}
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

			handler := MakeHandler(s, logger)
			w := httptest.NewRecorder()

			req, err := http.NewRequest(tc.method, tc.url, strings.NewReader(tc.body))
			require.NoError(t, err)

			handler.ServeHTTP(w, req)

			resp := w.Result()

			body, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			if tc.checkResponse != nil {
				require.Empty(t, tc.response, "response should not be set when using checkResponse")
				tc.checkResponse(t, string(body))
			} else {
				require.Equal(t, tc.response+"\n", string(body))
			}

			require.Equal(t, tc.statusCode, resp.StatusCode)
			require.Equal(t, "application/json; charset=utf-8", resp.Header.Get("Content-Type"))

			if tc.checkDB != nil {
				tc.checkDB(t, ctx, s.(service))
			}
		})
	}
}
