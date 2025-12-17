package db

import (
	"context"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
)

func TestTransferTx(t *testing.T) {
	ctxTx := context.Background()

	store := MustNewStore(ctxTx, configConnect)

	accountFrom, err, _ := createRandomAccount()
	require.NoError(t, err)
	require.NotEmpty(t, accountFrom)

	accountTo, err, _ := createRandomAccount()

	require.NoError(t, err)
	require.NotEmpty(t, accountTo)

	// run n concurrent transfer transactions
	n := 5
	amount := int64(gofakeit.Number(0, 1000))

	errs := make(chan error)
	responses := make(chan TransferTxResponse)

	for i := 0; i < n; i++ {
		go func() {
			response, err := store.TransferTx(ctxTx, TransferTxOptions{
				FromAccountID: accountFrom.ID,
				ToAccountID:   accountTo.ID,
				Amount:        amount,
			})

			errs <- err
			responses <- response
		}()
	}

	// Check responses
	existed := make(map[int]bool)
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)

		response := <-responses
		require.NotEmpty(t, response)

		// Check account
		require.NotEmpty(t, response.FromAccount)
		require.NotEmpty(t, response.ToAccount)

		require.NotZero(t, response.FromAccount.ID)
		require.NotZero(t, response.ToAccount.ID)

		require.Equal(t, accountFrom.ID, response.FromAccount.ID)
		require.Equal(t, accountTo.ID, response.ToAccount.ID)

		require.Equal(t, accountFrom.Balance-(amount*int64(i+1)), response.FromAccount.Balance)
		require.Equal(t, accountTo.Balance+(amount*int64(i+1)), response.ToAccount.Balance)

		require.WithinDuration(t, accountFrom.CreatedAt.Time, response.FromAccount.CreatedAt.Time, time.Second)
		require.WithinDuration(t, accountTo.CreatedAt.Time, response.ToAccount.CreatedAt.Time, time.Second)

		// Check transfer
		require.NotEmpty(t, response.Transfer)
		require.Equal(t, accountFrom.ID, response.Transfer.FromAccountID.Int64)
		require.Equal(t, accountTo.ID, response.Transfer.ToAccountID.Int64)
		require.Equal(t, amount, response.Transfer.Amount)
		require.NotZero(t, response.Transfer.CreatedAt)
		require.NotZero(t, response.Transfer.ID)

		_, err = store.GetTransfer(ctxTx, response.Transfer.ID)
		require.NoError(t, err)

		// Check entries
		require.NotEmpty(t, response.FromEntry)
		require.NotEmpty(t, response.ToEntry)

		require.NotZero(t, response.FromEntry.CreatedAt)
		require.NotZero(t, response.ToEntry.CreatedAt)
		require.NotZero(t, response.FromEntry.ID)
		require.NotZero(t, response.ToEntry.ID)

		require.Equal(t, accountFrom.ID, response.FromEntry.AccountID.Int64)
		require.Equal(t, accountTo.ID, response.ToEntry.AccountID.Int64)
		require.Equal(t, -amount, response.FromEntry.Amount)
		require.Equal(t, amount, response.ToEntry.Amount)

		_, err = store.GetEntry(ctxTx, response.FromEntry.ID)
		require.NoError(t, err)

		_, err = store.GetEntry(ctxTx, response.ToEntry.ID)
		require.NoError(t, err)

		// check accounts
		fromAccount := response.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, accountFrom.ID, fromAccount.ID)

		toAccount := response.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, accountTo.ID, toAccount.ID)

		// check accounts' balance
		diff1 := accountFrom.Balance - fromAccount.Balance
		diff2 := toAccount.Balance - accountTo.Balance
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		require.True(t, diff1%amount == 0) // amount, 2 * amount, 3 * amount, ... n * amount

		k := int(diff1 / amount)
		require.True(t, k >= 1 && k <= n)
		require.NotContains(t, existed, k)
		existed[k] = true
	}

	// check the final updated balances
	updateAccount1, err := testQueries.GetAccount(context.Background(), accountFrom.ID)
	require.NoError(t, err)

	updateAccount2, err := testQueries.GetAccount(context.Background(), accountTo.ID)
	require.NoError(t, err)

	require.Equal(t, accountFrom.Balance-int64(n)*amount, updateAccount1.Balance)
	require.Equal(t, accountTo.Balance+int64(n)*amount, updateAccount2.Balance)
}
