package db

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ctx           = context.Background()
	configConnect = os.Getenv("TEST_CONNECTION")
	faker         = gofakeit.New(0)
	pool          *pgxpool.Pool
	store         Store
)

func TestMain(m *testing.M) {
	var err error
	pool, err = pgxpool.New(ctx, configConnect)
	if err != nil {
		log.Fatal("cannot connect to database:", err)
	}
	defer pool.Close()

	store = NewStore(pool)

	os.Exit(m.Run())
}
