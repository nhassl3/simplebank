package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/nhassl3/simplebank/internals/db/sqlc"
	domain "github.com/nhassl3/simplebank/internals/domain/http"
	"github.com/nhassl3/simplebank/internals/http/simplebank"
	"github.com/nhassl3/simplebank/internals/lib/token"
)

type App struct {
	store   db.Store
	pool    *pgxpool.Pool
	handler simplebank.Simplebank
	Server  *simplebank.Server
	Address string
}

// MustNewApp creates new application if not - panic
func MustNewApp(log *slog.Logger, secretKey, connectionDBString, host string, port int, duration time.Duration) *App {
	pool, err := pgxpool.New(context.Background(), connectionDBString)
	if err != nil {
		panic(err)
	}

	PASETOMaker, err := token.NewPASETOMaker([]byte(secretKey), duration)
	if err != nil {
		panic(err)
	}

	store := db.NewStore(pool)

	return &App{
		pool:    pool,
		store:   store,
		handler: domain.NewHandler(log, store, PASETOMaker),
		Server:  simplebank.MustNewServer(PASETOMaker, log),
		Address: fmt.Sprintf("%s:%d", host, port),
	}
}

func (app *App) MustStart() {
	app.Server.Register(app.handler)
	if err := app.Server.Router.Run(app.Address); err != nil {
		panic(err)
	}
}

// Stop provide functionality of "Graceful Shutdown"
func (app *App) Stop() {
	// Stop close a pool of connection to the database.
	// To prevent data leaks
	app.pool.Close()
	// TODO: Stop HTTP server
}
