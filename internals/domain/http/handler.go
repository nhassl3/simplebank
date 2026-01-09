package domain

import (
	"log/slog"

	db "github.com/nhassl3/simplebank/internals/db/sqlc"
	"github.com/nhassl3/simplebank/internals/domain/http/handlers"
	"github.com/nhassl3/simplebank/internals/http/simplebank"
)

// Handler realize domain layer accessing the repository (database) layer
type Handler struct {
	*handlers.AccountHandler
	*handlers.TransferHandler
	*handlers.UserHandler
}

func NewHandler(log *slog.Logger, store db.Store) simplebank.Simplebank {
	return &Handler{
		handlers.NewAccountHandler(log, store),
		handlers.NewTransferHandler(log, store),
		handlers.NewUserHandler(log, store),
	}
}
