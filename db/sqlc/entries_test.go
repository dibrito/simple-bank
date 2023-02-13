package db

import (
	"context"
	"testing"
	"time"

	"github.com/dibrito/simple-bank/util"
	"github.com/stretchr/testify/require"
)

func createTestEntry(t *testing.T, accountId int64) Entry {
	arg := CreateEntryParams{
		AccountID: accountId,
		Amount:    util.RandonMoney(),
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

func TestCreateEntry(t *testing.T) {
	acc := createTestAccount(t)
	createTestEntry(t, acc.ID)
}

func TestGetEntry(t *testing.T) {
	acc := createTestAccount(t)
	entry := createTestEntry(t, acc.ID)
	e, err := testQueries.GetEntry(context.Background(), entry.ID)
	require.NoError(t, err)
	require.NotEmpty(t, e)

	require.Equal(t, entry.ID, e.ID)
	require.Equal(t, entry.Amount, e.Amount)
	require.Equal(t, entry.AccountID, e.AccountID)
	require.WithinDuration(t, entry.CreatedAt, e.CreatedAt, time.Second)
}

func TestListEntry(t *testing.T) {
	acc := createTestAccount(t)
	for i := 0; i < 10; i++ {
		createTestEntry(t, acc.ID)
	}

	entries, err := testQueries.ListEntries(context.Background(), ListEntriesParams{
		AccountID: acc.ID,
		Limit:     5,
		Offset:    5,
	})
	require.NoError(t, err)
	require.Len(t, entries, 5)

	for _, v := range entries {
		require.NotEmpty(t, v)
	}
}
