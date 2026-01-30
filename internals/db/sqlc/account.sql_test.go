package db

import (
	"database/sql"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/require"
)

func createRandomAccount() (account Account, err error, args CreateAccountParams) {
	user, _, err := createRandomUser(generateRandomPassword())
	if err != nil {
		return
	}

	args = CreateAccountParams{
		user.Username,
		int64(faker.IntN(1000)),
		faker.RandomString([]string{"USD", "EUR"}),
	}

	account, err = store.CreateAccount(ctx, args)

	return
}

func TestCreateAccounts(t *testing.T) {
	for i := 0; i < 5; i++ {
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

	err = store.DeleteAccount(ctx, account.ID)
	require.NoError(t, err)

	accountRes, err := store.GetAccount(ctx, account.ID)
	require.Error(t, err)
	require.ErrorContains(t, sql.ErrNoRows, err.Error())
	require.Empty(t, accountRes)
}

func TestGetRandomAccount(t *testing.T) {
	account, err := store.GetAccount(ctx, int64(gofakeit.IntRange(1, 5)))
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

	testAccount, err := store.GetAccount(ctx, account.ID)

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

func TestListAccounts(t *testing.T) {
	account, err := store.GetAccount(ctx, int64(gofakeit.IntRange(1, 5)))
	require.NoError(t, err)
	require.NotEmpty(t, account)

	listParams := ListAccountsParams{
		Owner:  account.Owner,
		Limit:  10,
		Offset: 0,
	}

	listParams2 := ListAccountsParams{
		Owner:  account.Owner,
		Limit:  -1,
		Offset: 0,
	}

	listParams3 := ListAccountsParams{
		Limit:  10,
		Offset: 0,
	}

	for _, tc := range []struct {
		name          string
		params        ListAccountsParams
		checkResponse func(t *testing.T, tc ListAccountsParams, result []Account, err error)
	}{
		{
			name:   "OK",
			params: listParams,
			checkResponse: func(t *testing.T, tc ListAccountsParams, result []Account, err error) {
				require.NoError(t, err)
				require.NotEmpty(t, result)

				require.LessOrEqual(t, len(result), 2)

			},
		},
		{
			name:   "Invalid limit",
			params: listParams2,
			checkResponse: func(t *testing.T, tc ListAccountsParams, result []Account, err error) {
				require.Error(t, err)
				require.Empty(t, result)
				require.EqualError(t, err, "ERROR: LIMIT must not be negative (SQLSTATE 2201W)")
			},
		},
		{
			name:   "Empty slice (without owner)",
			params: listParams3,
			checkResponse: func(t *testing.T, tc ListAccountsParams, result []Account, err error) {
				require.Empty(t, result)
				require.NoError(t, err)
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			accounts, err := store.ListAccounts(ctx, tc.params)
			tc.checkResponse(t, tc.params, accounts, err)
		})
	}
}

func TestUpdateAccount(t *testing.T) {
	account, err := store.GetAccount(ctx, int64(faker.IntRange(1, 5)))
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

	newAccountData, err := store.UpdateAccountBalance(ctx, newArgs)
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

func TestAddAccountBalance(t *testing.T) {
	account, err := store.GetAccount(ctx, int64(faker.IntRange(1, 5)))
	require.NoError(t, err)
	require.NotEmpty(t, account)
	require.NotZero(t, account.ID)
	require.NotZero(t, account.CreatedAt)

	argsPlus := AddAccountBalanceParams{
		ID:     account.ID,
		Amount: 100,
	}

	argsMinus := AddAccountBalanceParams{
		ID:     account.ID,
		Amount: -100,
	}

	for _, tc := range []struct {
		name          string
		args          AddAccountBalanceParams
		checkResponse func(t *testing.T, tc AddAccountBalanceParams, result Account, err error)
	}{
		{
			name: "OK/Plus",
			args: argsPlus,
			checkResponse: func(t *testing.T, tc AddAccountBalanceParams, result Account, err error) {
				require.NoError(t, err)
				require.NotEmpty(t, result)
				require.NotZero(t, result.ID)
				require.NotEmpty(t, tc)
				require.NotZero(t, tc.ID)

				require.Equal(t, tc.ID, result.ID)
				require.Equal(t, account.Balance+100, result.Balance)
			},
		},
		{
			name: "OK/Minus",
			args: argsMinus,
			checkResponse: func(t *testing.T, tc AddAccountBalanceParams, result Account, err error) {
				require.NoError(t, err)
				require.NotEmpty(t, result)
				require.NotZero(t, result.ID)
				require.NotEmpty(t, tc)
				require.NotZero(t, tc.ID)

				require.Equal(t, tc.ID, result.ID)
				require.Equal(t, account.Balance, result.Balance)
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			accounts, err := store.AddAccountBalance(ctx, tc.args)
			tc.checkResponse(t, tc.args, accounts, err)
		})
	}
}
