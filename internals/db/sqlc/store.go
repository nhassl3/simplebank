package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store provide all functions to execute db queries and transactions
type Store interface {
	Querier
	TransferTx(ctx context.Context, arg TransferTxOptions) (TransferTxResponse, error)
}

// SQLStore provide all functions to execute SQL queries and transactions
type SQLStore struct {
	pool *pgxpool.Pool
	*Queries
}

// NewStore create a new Store
func NewStore(pool *pgxpool.Pool) Store {
	return &SQLStore{
		pool:    pool,
		Queries: New(pool), // From /sqlc/db of db package
	}
}

func (s *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	// Create new connection from pool
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed when acquire db connection: %w", err)
	}
	defer conn.Release()

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return err
	}

	if err = fn(s.WithTx(tx)); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("tx err : %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit(ctx)
}

type TransferTxOptions struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

type TransferTxResponse struct {
	Transfer    Transfer `json:"transfer"`
	FromAccount Account  `json:"from_account"`
	ToAccount   Account  `json:"to_account"`
	FromEntry   Entry    `json:"from_entry"`
	ToEntry     Entry    `json:"to_entry"`
}

// TransferTx performs a money transfer from one account to the other
// It creates a transfer record, add account entries, and update accounts' balances within a single database transaction
func (s *SQLStore) TransferTx(ctx context.Context, arg TransferTxOptions) (TransferTxResponse, error) {
	var response TransferTxResponse

	err := s.execTx(ctx, func(q *Queries) error {
		var fnErr error

		response.Transfer, fnErr = q.CreateTransfer(ctx, CreateTransferParams{
			FromAccountID: pgtype.Int8{Int64: arg.FromAccountID, Valid: true},
			ToAccountID:   pgtype.Int8{Int64: arg.ToAccountID, Valid: true},
			Amount:        arg.Amount,
		})
		if fnErr != nil {
			return fnErr
		}

		response.FromEntry, fnErr = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: pgtype.Int8{Int64: arg.FromAccountID, Valid: true},
			Amount:    -arg.Amount,
		})
		if fnErr != nil {
			return fnErr
		}

		response.ToEntry, fnErr = q.CreateEntry(ctx, CreateEntryParams{
			AccountID: pgtype.Int8{Int64: arg.ToAccountID, Valid: true},
			Amount:    arg.Amount,
		})
		if fnErr != nil {
			return fnErr
		}

		if arg.FromAccountID < arg.ToAccountID {
			response.FromAccount, response.ToAccount,
				fnErr = addMoney(ctx, q, arg.FromAccountID, -arg.Amount, arg.ToAccountID, arg.Amount)
		} else {
			response.ToAccount, response.FromAccount,
				fnErr = addMoney(ctx, q, arg.ToAccountID, arg.Amount, arg.FromAccountID, -arg.Amount)
		}

		if fnErr != nil {
			return fnErr
		}

		return nil
	})

	return response, err
}

func addMoney(
	ctx context.Context,
	q *Queries,
	accountID1,
	amount1,
	accountID2,
	amount2 int64,
) (account1 Account, account2 Account, err error) {
	account1, err = q.AddAccountBalanceByID(ctx, AddAccountBalanceByIDParams{
		ID:     accountID1,
		Amount: amount1,
	})
	if err != nil {
		return
	}
	account2, err = q.AddAccountBalanceByID(ctx, AddAccountBalanceByIDParams{
		ID:     accountID2,
		Amount: amount2,
	})
	return
}
