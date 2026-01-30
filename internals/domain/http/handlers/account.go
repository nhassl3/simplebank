package handlers

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/jackc/pgx/v5/pgconn"
	db "github.com/nhassl3/simplebank/internals/db/sqlc"
	"github.com/nhassl3/simplebank/internals/http/simplebank/session"
	"github.com/nhassl3/simplebank/internals/lib/logger/sl"
)

const (
	opCreateAccount        = "domain.CreateAccount"
	opGetAccount           = "domain.GetAccount"
	opListAccounts         = "domain.ListAccounts"
	opUpdateAccountBalance = "domain.UpdateAccountBalance"
	opAddAccountBalance    = "domain.AddAccountBalance"
	opDeleteAccount        = "domain.DeleteAccount"
)

type AccountHandler struct {
	log   *slog.Logger
	store db.Store
}

func NewAccountHandler(log *slog.Logger, store db.Store) *AccountHandler {
	return &AccountHandler{
		log:   log,
		store: store,
	}
}

// CreateAccount creates new account with given parameters
func (h *AccountHandler) CreateAccount(ctx context.Context, in session.CreateAccountRequest) (*db.Account, error) {
	log := h.log.With(slog.String("op", opCreateAccount))

	account, err := h.store.CreateAccount(ctx, db.CreateAccountParams{
		Owner:    in.Owner,
		Balance:  0,
		Currency: in.Currency,
	})

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23503":
				return nil, sl.ErrorNoUsers
			case "23505":
				return nil, sl.ErrorAccountAlreadyExists
			default:
				return nil, pgErr
			}
		}
		log.Error("failed to create account", sl.Err(err))
		return nil, sl.ErrUpLevel(opCreateAccount, err)
	}

	return &account, nil
}

// GetAccount returns account by identifier (int64)
func (h *AccountHandler) GetAccount(ctx context.Context, id int64) (*db.Account, error) {
	log := h.log.With("op", opGetAccount)

	account, err := h.store.GetAccount(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sl.ErrorNoAccounts
		}
		log.Error("failed to get account", sl.Err(err))
		return nil, sl.ErrUpLevel(opGetAccount, err)
	}

	return &account, nil
}

// ListAccounts lists accounts with given offset and limit
func (h *AccountHandler) ListAccounts(ctx context.Context, in session.ListAccountsRequest) (*[]db.Account, error) {
	log := h.log.With("op", opListAccounts)

	account, err := h.store.ListAccounts(ctx, db.ListAccountsParams{
		Owner:  in.Owner,
		Offset: (in.Page - 1) * in.Limit,
		Limit:  in.Limit,
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sl.ErrorNoAccounts
		}
		log.Error("failed to list accounts", sl.Err(err))
		return nil, sl.ErrUpLevel(opListAccounts, err)
	}

	return &account, nil
}

// UpdateAccountBalance updates balance of account by given balance
func (h *AccountHandler) UpdateAccountBalance(ctx context.Context, in session.UpdateAccountRequest) (*db.Account, error) {
	log := h.log.With("op", opUpdateAccountBalance)

	account, err := h.store.UpdateAccountBalance(ctx, db.UpdateAccountBalanceParams{
		ID:      in.ID,
		Balance: in.Balance,
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sl.ErrorNoAccounts
		}
		log.Error("failed to update account balance", sl.Err(err))
		return nil, sl.ErrUpLevel(opUpdateAccountBalance, err)
	}

	return &account, nil
}

// AddAccountBalance adds balance for account on was given amount
func (h *AccountHandler) AddAccountBalance(ctx context.Context, in session.AddAccountBalanceRequest) (*db.Account, error) {
	log := h.log.With("op", opAddAccountBalance)

	account, err := h.store.AddAccountBalance(ctx, db.AddAccountBalanceParams{
		ID:     in.ID,
		Amount: in.Amount,
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sl.ErrorNoAccounts
		}
		log.Error("failed to add account balance", sl.Err(err))
		return nil, sl.ErrUpLevel(opAddAccountBalance, err)
	}

	return &account, nil
}

// DeleteAccount deletes account from the system
func (h *AccountHandler) DeleteAccount(ctx context.Context, id int64) error {
	log := h.log.With("op", opDeleteAccount)

	if err := h.store.DeleteAccount(ctx, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sl.ErrorNoAccounts
		}
		log.Error("failed to delete account", sl.Err(err))
		return sl.ErrUpLevel(opDeleteAccount, err)
	}

	return nil
}
