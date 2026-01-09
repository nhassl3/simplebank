package handlers

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	db "github.com/nhassl3/simplebank/internals/db/sqlc"
	"github.com/nhassl3/simplebank/internals/http/simplebank/requests"
	"github.com/nhassl3/simplebank/internals/lib/logger/sl"
)

const (
	opCreateTransfer         = "domain.CreateTransfer"
	opAccountCurrencyChecker = "domain.accountCurrencyChecker"
)

type TransferHandler struct {
	log   *slog.Logger
	store db.Store
}

func NewTransferHandler(log *slog.Logger, store db.Store) *TransferHandler {
	return &TransferHandler{
		log:   log,
		store: store,
	}
}

func (h *TransferHandler) CreateTransfer(ctx context.Context, in requests.TransferRequest) (*db.Transfer, error) {
	log := h.log.With("op", opCreateTransfer)

	if ok, err := h.accountCurrenciesChecker(
		ctx,
		in.FromAccountID, in.ToAccountID, in.Amount,
		in.Currency); ok == false && err != nil {
		return nil, err
	}

	result, err := h.store.TransferTx(ctx, db.TransferTxOptions{
		FromAccountID: in.FromAccountID,
		ToAccountID:   in.ToAccountID,
		Amount:        in.Amount,
	})
	if err != nil {
		log.Error("failed to create transfer", sl.Err(err))
		return nil, sl.ErrUpLevel(opCreateTransfer, err)
	}

	return &result.Transfer, nil
}

func (h *TransferHandler) accountCurrenciesChecker(
	ctx context.Context,
	fromAccountID, toAccountID, amount int64,
	currency string,
) (bool, error) {
	log := h.log.With("op", opAccountCurrencyChecker)

	fromAccount, errF := h.store.GetAccount(ctx, fromAccountID)
	if errF == nil && fromAccount.Balance < amount {
		return false, sl.ErrorNotEnoughMoney
	}

	toAccount, errT := h.store.GetAccount(ctx, toAccountID)

	if errF != nil || errT != nil {
		if errors.Is(errF, sql.ErrNoRows) && errors.Is(errT, sql.ErrNoRows) {
			return false, sl.ErrorBothAccountIDNotFound
		} else if errors.Is(errF, sql.ErrNoRows) {
			return false, sl.ErrorFromAccountIDNotFound
		} else if errors.Is(errT, sql.ErrNoRows) {
			return false, sl.ErrorToAccountIDNotFound
		}
		if errF != nil {
			log.Error("failed to get sender account", sl.Err(errF))
			return false, sl.ErrUpLevel(opAccountCurrencyChecker, errF)
		} else {
			log.Error("failed to get rec account", sl.Err(errT))
			return false, sl.ErrUpLevel(opAccountCurrencyChecker, errT)
		}
	}

	if fromAccount.Currency != currency || toAccount.Currency != currency {
		return false, sl.ErrorMismatchCurrencies
	}

	return true, nil
}
