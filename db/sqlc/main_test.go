package db

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
)

var (
	testQueries   *Queries
	ctx           = context.Background()
	configConnect = os.Getenv("TEST_CONNECTION")
)

func TestMain(m *testing.M) {
	conn, err := pgx.Connect(ctx, configConnect)
	if err != nil {
		log.Fatal("cannot connect to database:", err)
	}
	defer conn.Close(ctx)

	testQueries = New(conn)

	os.Exit(m.Run())
}
