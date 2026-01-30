package db

import (
	"context"
	"log"
	"math/rand/v2"
	"os"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nhassl3/simplebank/internals/config"
)

var (
	ctx      = context.Background()
	cfg, err = config.LoadConfigByString("../../../config/local.yaml", "../../../.env")
	faker    = gofakeit.NewFaker(rand.NewPCG(11, 11), true)
	pool     *pgxpool.Pool
	store    Store
)

func TestMain(m *testing.M) {
	pool, err = pgxpool.New(ctx, cfg.ConnectionDBString)
	if err != nil {
		log.Fatal("cannot connect to database:", err)
	}
	defer pool.Close()

	store = NewStore(pool)

	os.Exit(m.Run())
}
