package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5/pgconn"
	mockdb "github.com/nhassl3/simplebank/internals/db/mock"
	db "github.com/nhassl3/simplebank/internals/db/sqlc"
	"github.com/nhassl3/simplebank/tests/suit"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

func TestCreateAccount(t *testing.T) {
	t.Helper()

	s := suit.NewSuite(t)
	defer s.Finish()

	testParam, account := suit.CreateRandomParams()

	testParam2, _ := suit.CreateRandomParams()
	testParam2.Currency = strconv.FormatBool(false)

	testCases := []suit.TestCase[db.CreateAccountParams]{
		{
			Name: "OK",
			Data: testParam,
			BuildStubs: func(store *mockdb.MockStore) {
				s.Store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(testParam)).
					Times(1).
					Return(account, nil)
			},
			CheckResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, resp.Code)
			},
		},
		{
			Name: "AccountAlreadyExists",
			Data: testParam,
			BuildStubs: func(store *mockdb.MockStore) {
				s.Store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(testParam)).
					Times(1).
					Return(db.Account{}, &pgconn.PgError{
						Code:           "23505", // UniqueViolation
						Message:        "duplicate key value violates unique constraint",
						Detail:         "Key (owner, currency)=(test_user, USD) already exists.",
						ConstraintName: "accounts_owner_currency_key",
					})
			},
			CheckResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusConflict, resp.Code)
			},
		},
		{
			Name: "InvalidCurrency",
			Data: testParam2,
			BuildStubs: func(store *mockdb.MockStore) {
				s.Store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(testParam2)).
					Times(0)
			},
			CheckResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, resp.Code)
			},
		},
		{
			Name: "InternalError",
			Data: testParam,
			BuildStubs: func(store *mockdb.MockStore) {
				s.Store.EXPECT().
					CreateAccount(gomock.Any(), gomock.Eq(testParam)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			CheckResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, resp.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			tc.BuildStubs(s.Store)

			recorder := httptest.NewRecorder()

			// Здесь нужно создать JSON тело запроса для создания аккаунта
			body := map[string]interface{}{
				"owner":    tc.Data.Owner,
				"balance":  0,
				"currency": tc.Data.Currency,
			}

			jsonBody, err := json.Marshal(body)
			require.NoError(t, err)

			request, err := http.NewRequest(http.MethodPost, suit.CreateAccountUrl(), bytes.NewBuffer(jsonBody))
			require.NoError(t, err)
			request.Header.Set("Content-Type", "application/json")

			s.Server.Router.ServeHTTP(recorder, request)
			tc.CheckResponse(t, recorder)
		})
	}
}

func TestGetAccount(t *testing.T) {
	t.Helper()

	account := suit.RandomAccount()

	s := suit.NewSuite(t)
	defer s.Finish()

	testCases := []suit.TestCase[int64]{
		{
			Name: "OK",
			Data: account.ID,
			BuildStubs: func(store *mockdb.MockStore) {
				s.Store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(account, nil)
			},
			CheckResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, resp.Code)
				data, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				var gotAccount db.Account
				err = json.Unmarshal(data, &gotAccount)
				require.NoError(t, err)
				require.Equal(t, account, gotAccount)
			},
		},
		{
			Name: "AccountNotFound",
			Data: account.ID,
			BuildStubs: func(store *mockdb.MockStore) {
				s.Store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrNoRows)
			},
			CheckResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, resp.Code)
			},
		},
		{
			Name: "InvalidAccountID",
			Data: 0,
			BuildStubs: func(store *mockdb.MockStore) {
				s.Store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(0)
			},
			CheckResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, resp.Code)
			},
		},
		{
			Name: "InternalError",
			Data: account.ID,
			BuildStubs: func(store *mockdb.MockStore) {
				s.Store.EXPECT().
					GetAccount(gomock.Any(), gomock.Eq(account.ID)).
					Times(1).
					Return(db.Account{}, sql.ErrConnDone)
			},
			CheckResponse: func(t *testing.T, resp *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, resp.Code)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Building stubs for testing
			tc.BuildStubs(s.Store)

			recorder := httptest.NewRecorder()
			request, err := http.NewRequest(http.MethodGet, suit.GetAccountUrl(tc.Data), nil)
			require.NoError(t, err)

			// Run serving
			s.Server.Router.ServeHTTP(recorder, request)

			// Check response
			tc.CheckResponse(t, recorder)
		})
	}
}

func TestListAccounts(t *testing.T) {

}

func TestUpdateAccountBalance(t *testing.T) {

}

func TestAddAccountBalance(t *testing.T) {

}

func TestDeleteAccount(t *testing.T) {

}
