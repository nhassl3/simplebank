package suit

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/golang/mock/gomock"
	mockdb "github.com/nhassl3/simplebank/internals/db/mock"
	db "github.com/nhassl3/simplebank/internals/db/sqlc"
	http2 "github.com/nhassl3/simplebank/internals/domain/http"
	"github.com/nhassl3/simplebank/internals/http/simplebank"
	"github.com/nhassl3/simplebank/internals/lib/token"
	testsToken "github.com/nhassl3/simplebank/internals/lib/token/tests"
)

var (
	baseUrl = "/api/v1"
)

type TestCase[T interface{}] struct {
	Name          string
	Data          T
	SetupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
	BuildStubs    func(store *mockdb.MockStore)
	CheckResponse func(t *testing.T, resp *httptest.ResponseRecorder)
}

type Suite struct {
	controller *gomock.Controller
	Store      *mockdb.MockStore
	Server     *simplebank.Server
	TGPMaker   token.Maker
}

func NewSuite(t *testing.T, duration time.Duration) (*Suite, error) {
	controller := gomock.NewController(t)
	store := mockdb.NewMockStore(controller)

	secretKey, err := testsToken.GenerateRandomBytes(32)
	if err != nil {
		return nil, err
	}

	log := slog.Default()
	PASETOMaker, err := token.NewPASETOMaker(secretKey, duration)
	if err != nil {
		return nil, err
	}

	server := simplebank.MustNewServer(PASETOMaker, log)
	handler := http2.NewHandler(log, store, PASETOMaker)
	server.Register(handler)

	return &Suite{
		controller: controller,
		Store:      store,
		Server:     server,
		TGPMaker:   PASETOMaker,
	}, nil
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

func RandomUser() db.User {
	hash, _ := argon2id.CreateHash("oaiduagd", argon2id.DefaultParams)

	return db.User{
		Username:       gofakeit.Username(),
		HashedPassword: hash,
		FullName:       "Madad Add",
		Email:          gofakeit.Email(),
	}
}

func RandomAccount() db.Account {
	owner := RandomUser().Username
	return db.Account{
		ID:       int64(gofakeit.IntRange(1, 1000)),
		Owner:    owner,
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
