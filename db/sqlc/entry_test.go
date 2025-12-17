package db

import (
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
)

type EntryWithArgs struct {
	*Entry
	srcAmount int64
	srcID     int64
	err       error
}

func createRandomEntries(t *testing.T, n int) []EntryWithArgs {
	if n == 0 {
		n = 1
	}
	entries := make([]EntryWithArgs, 0)

	for i := 0; i <= n; i++ {
		account, err, _ := createRandomAccount()
		require.NoError(t, err)
		require.NotEmpty(t, account)

		amount := int64(gofakeit.IntRange(1, 1000))

		args := CreateEntryParams{
			AccountID: pgtype.Int8{Int64: account.ID, Valid: true},
			Amount:    amount,
		}

		entry, err := testQueries.CreateEntry(ctx, args)
		entries = append(entries, EntryWithArgs{
			Entry:     &entry,
			srcAmount: amount,
			srcID:     account.ID,
			err:       err,
		})
	}

	return entries
}

func TestCreateEntry(t *testing.T) {
	entries := createRandomEntries(t, 2)
	for _, entry := range entries {
		require.NoError(t, entry.err)
		require.NotZero(t, entry.ID)
		require.NotZero(t, entry.CreatedAt)

		require.Equal(t, entry.srcAmount, entry.Amount)
		require.Equal(t, entry.srcID, entry.AccountID.Int64)
	}
}

func TestGetEntry(t *testing.T) {
	entries := createRandomEntries(t, 1)

	entry := entries[0]
	require.NoError(t, entry.err)
	require.NotZero(t, entry.ID)
	require.NotZero(t, entry.CreatedAt)

	require.Equal(t, entry.srcAmount, entry.Amount)
	require.Equal(t, entry.srcID, entry.AccountID.Int64)

	entryGet, err := testQueries.GetEntry(ctx, entry.ID)
	require.NoError(t, err)
	require.NotEmpty(t, entryGet)
	require.Equal(t, entry.Amount, entryGet.Amount)

	require.Equal(t, *entry.Entry, entryGet)

	require.WithinDuration(t, entry.CreatedAt.Time, entryGet.CreatedAt.Time, time.Second)
}
