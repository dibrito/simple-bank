package db

import (
	"context"
	"testing"
	"time"

	"github.com/dibrito/simple-bank/db/util"
	"github.com/stretchr/testify/require"
)

func createTestTransfer(t *testing.T, from, to int64) Transfer {
	arg := CreateTransferParams{
		FromAccountID: from,
		ToAccountID:   to,
		Amount:        util.RandonMoney(),
	}
	transfer, err := testQueries.CreateTransfer(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, transfer)
	require.Equal(t, arg.FromAccountID, transfer.FromAccountID)
	require.Equal(t, arg.ToAccountID, transfer.ToAccountID)
	require.Equal(t, arg.Amount, transfer.Amount)

	require.NotZero(t, transfer.ID)
	require.NotZero(t, transfer.CreatedAt)
	return transfer
}

func TestCreateTransfer(t *testing.T) {
	from := createTestAccount(t)
	to := createTestAccount(t)
	createTestTransfer(t, from.ID, to.ID)
}

func TestGetTransfe(t *testing.T) {
	from := createTestAccount(t)
	to := createTestAccount(t)
	transf := createTestTransfer(t, from.ID, to.ID)

	transf2, err := testQueries.GetTransfer(context.Background(), transf.ID)
	require.NoError(t, err)
	require.NotEmpty(t, transf2)

	require.Equal(t, transf.FromAccountID, transf2.FromAccountID)
	require.Equal(t, transf.ToAccountID, transf2.ToAccountID)
	require.Equal(t, transf.Amount, transf2.Amount)

	require.Equal(t, transf.ID, transf2.ID)
	require.WithinDuration(t, transf.CreatedAt, transf2.CreatedAt, time.Second)
}

func TestListTransfer(t *testing.T) {
	from := createTestAccount(t)
	to := createTestAccount(t)
	for i := 0; i < 10; i++ {
		createTestTransfer(t, from.ID, to.ID)
	}

	transfers, err := testQueries.ListTransfers(context.Background(), ListTransfersParams{
		FromAccountID: from.ID,
		ToAccountID:   to.ID,
		Limit:         5,
		Offset:        5,
	})
	require.NoError(t, err)
	require.Len(t, transfers, 5)

	for _, v := range transfers {
		require.NotEmpty(t, v)
	}
}
