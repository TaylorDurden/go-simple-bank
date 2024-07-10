package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/taylordurden/go-simple-bank/util"
)

func TestCreateEntry(t *testing.T) {
	account := createRandAccount(t)
	createRandEntry(t, account.ID)
}

func TestGetEntry(t *testing.T) {
	account := createRandAccount(t)
	entry := createRandEntry(t, account.ID)
	dbEntry, err := testQueries.GetEntry(context.Background(), entry.ID)
	assert.NoError(t, err)
	require.NotEmpty(t, dbEntry)

	require.Equal(t, dbEntry.ID, entry.ID)
	require.Equal(t, dbEntry.AccountID, entry.AccountID)
	require.Equal(t, dbEntry.Amount, entry.Amount)
	require.WithinDuration(t, dbEntry.CreatedAt, dbEntry.CreatedAt, time.Second)
}

func TestListEntries(t *testing.T) {
	account := createRandAccount(t)
	for i := 0; i < 10; i++ {
		createRandEntry(t, account.ID)
	}

	arg := ListEntriesParams{
		AccountID: account.ID,
		Limit:     5,
		Offset:    5,
	}

	entries, err := testQueries.ListEntries(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, entries, 5)

	for _, entry := range entries {
		require.NotEmpty(t, entry)
	}
}

func createRandEntry(t *testing.T, accountId int64) Entry {
	arg := CreateEntryParams{
		AccountID: accountId,
		Amount:    util.RandomBalance(),
	}

	entry, err := testQueries.CreateEntry(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, entry)

	require.Equal(t, arg.AccountID, entry.AccountID)
	require.Equal(t, arg.Amount, entry.Amount)

	require.NotZero(t, entry.ID)
	require.NotZero(t, entry.CreatedAt)
	return entry
}
