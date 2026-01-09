package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/nhassl3/simplebank/internals/db/sqlc"
	domain "github.com/nhassl3/simplebank/internals/domain/http"
	"github.com/nhassl3/simplebank/internals/http/simplebank"
)

type App struct {
	store   db.Store
	pool    *pgxpool.Pool
	handler simplebank.Simplebank
	Server  *simplebank.Server
	Address string
}

func NewApp(log *slog.Logger, connectionDBString, host string, port int) *App {
	pool, err := pgxpool.New(context.Background(), connectionDBString)
	if err != nil {
		panic(err)
	}
	store := db.NewStore(pool)
	return &App{
		pool:    pool,
		store:   store,
		handler: domain.NewHandler(log, store),
		Server:  simplebank.NewServer(),
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
