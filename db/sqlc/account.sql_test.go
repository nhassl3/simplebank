package db

import (
	"database/sql"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
)

func createRandomAccount() (account Account, err error, args CreateAccountParams) {
	args = CreateAccountParams{
		faker.Name(),
		int64(faker.IntN(1000)),
		faker.Currency().Short,
	}

	account, err = testQueries.CreateAccount(ctx, args)

	return
}

func TestCreateAccounts(t *testing.T) {
	for i := 0; i < 100; i++ {
		account, err, args := createRandomAccount()
		require.NoError(t, err)
		require.NotEmpty(t, account)

		require.Equal(t, args.Owner, account.Owner)
		require.Equal(t, args.Balance, account.Balance)
		require.Equal(t, args.Currency, account.Currency)

		require.NotZero(t, account.ID)
		require.NotZero(t, account.CreatedAt)
	}
}

func TestDeleteAccount(t *testing.T) {
	account, err, _ := createRandomAccount()
	require.NoError(t, err)
	require.NotEmpty(t, account)

	err = testQueries.DeleteAccount(ctx, account.ID)
	require.NoError(t, err)

	accountRes, err := testQueries.GetAccount(ctx, account.ID)
	require.Error(t, err)
	require.ErrorContains(t, sql.ErrNoRows, err.Error())
	require.Empty(t, accountRes)
}

func TestGetRandomAccount(t *testing.T) {
	account, err := testQueries.GetAccount(ctx, int64(gofakeit.IntRange(0, 99)))
	require.NoError(t, err)
	require.NotEmpty(t, account)
}

func TestGetAccount(t *testing.T) {
	account, err, args := createRandomAccount()

	require.NoError(t, err)
	require.NotEmpty(t, account)
	require.Equal(t, args.Owner, account.Owner)
	require.Equal(t, args.Balance, account.Balance)
	require.Equal(t, args.Currency, account.Currency)
	require.NotZero(t, account.ID)
	require.NotZero(t, account.CreatedAt)

	testAccount, err := testQueries.GetAccount(ctx, account.ID)

	require.NoError(t, err)
	require.NotEmpty(t, account)
	require.Equal(t, args.Owner, testAccount.Owner)
	require.Equal(t, args.Balance, testAccount.Balance)
	require.Equal(t, args.Currency, testAccount.Currency)
	require.NotZero(t, testAccount.ID)
	require.NotZero(t, testAccount.CreatedAt)

	// Most needed test
	require.WithinDuration(t, account.CreatedAt.Time, testAccount.CreatedAt.Time, time.Second)
}

func TestLimitAccounts(t *testing.T) {
	var (
		limit          int32 = 10
		offset         int32 = 0
		currentStartId int32 = 1
	)

	args := ListAccountsParams{
		Limit:  limit,
		Offset: offset,
	}

	accounts, err := testQueries.ListAccounts(ctx, args)
	require.NoError(t, err)
	require.NotEmpty(t, accounts)
	require.Len(t, accounts, int(limit))
	require.Equal(t, int64(currentStartId+offset), accounts[0].ID)
	require.Equal(t, int64(currentStartId+offset+limit-1), accounts[limit-1].ID)
}

func TestUpdateAccount(t *testing.T) {
	faker := gofakeit.New(0)

	account, err := testQueries.GetAccount(ctx, int64(faker.IntRange(0, 99)))
	require.NoError(t, err)
	require.NotEmpty(t, account)
	require.NotZero(t, account.ID)
	require.NotZero(t, account.CreatedAt)

	newArgs := UpdateAccountBalanceParams{
		account.ID,
		int64(faker.IntRange(1, 1000)),
	}

	if newArgs.Balance == account.Balance {
		newArgs.Balance += 1
	}

	newAccountData, err := testQueries.UpdateAccountBalance(ctx, newArgs)
	require.NoError(t, err)
	require.NotEmpty(t, newAccountData)
	require.Equal(t, newArgs.ID, newAccountData.ID)
	require.Equal(t, account.ID, newAccountData.ID)
	require.Equal(t, account.Owner, newAccountData.Owner)
	require.Equal(t, newArgs.Balance, newAccountData.Balance)
	require.Equal(t, account.Currency, newAccountData.Currency)
	require.NotEqual(t, account.Balance, newAccountData.Balance)

	require.NotZero(t, newAccountData.ID)

	// Check if created_at time not changed when updating statements
	require.WithinDuration(t, account.CreatedAt.Time, newAccountData.CreatedAt.Time, time.Second)
}
