package suit

import (
	"fmt"
	"log/slog"
	"net/http/httptest"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/golang/mock/gomock"
	mockdb "github.com/nhassl3/simplebank/internals/db/mock"
	db "github.com/nhassl3/simplebank/internals/db/sqlc"
	http2 "github.com/nhassl3/simplebank/internals/domain/http"
	"github.com/nhassl3/simplebank/internals/http/simplebank"
)

var (
	baseUrl = "/api/v1"
)

type TestCase[T interface{}] struct {
	Name          string
	Data          T
	BuildStubs    func(store *mockdb.MockStore)
	CheckResponse func(t *testing.T, resp *httptest.ResponseRecorder)
}

type Suite struct {
	controller *gomock.Controller
	Store      *mockdb.MockStore
	Server     *simplebank.Server
}

func NewSuite(t *testing.T) *Suite {
	controller := gomock.NewController(t)
	store := mockdb.NewMockStore(controller)

	server := simplebank.NewServer()
	handler := http2.NewHandler(slog.Default(), store)
	server.Register(handler)

	return &Suite{
		controller: controller,
		Store:      store,
		Server:     server,
	}
}

func (s *Suite) Finish() {
	s.controller.Finish()
}

func CreateAccountUrl() string {
	return fmt.Sprintf("%s/accounts/", baseUrl)
}

func GetAccountUrl(ID int64) string {
	return fmt.Sprintf("%s/accounts/%d", baseUrl, ID)
}

func RandomAccount() db.Account {
	return db.Account{
		ID:       int64(gofakeit.IntRange(1, 1000)),
		Owner:    gofakeit.Name(),
		Balance:  int64(gofakeit.IntRange(0, 1000)),
		Currency: gofakeit.RandomString([]string{"USD", "EUR"}),
	}
}

func CreateRandomParams() (db.CreateAccountParams, db.Account) {
	account := RandomAccount()

	return db.CreateAccountParams{
		Owner:    account.Owner,
		Balance:  0,
		Currency: account.Currency,
	}, account
}
